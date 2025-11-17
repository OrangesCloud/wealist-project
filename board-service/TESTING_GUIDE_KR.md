# Board Service 테스트 가이드

board-service와 user-service 간의 통합을 테스트하기 위한 가이드입니다.

## 🎯 목적

이 테스트 스크립트들은 다음을 확인합니다:
- board-service가 user-service를 정상적으로 호출하는지
- 인증 토큰이 올바르게 전달되는지
- API 엔드포인트가 예상대로 동작하는지
- 서비스 간 네트워크 연결이 정상인지

## 🚀 빠른 시작

### 1. 전체 통합 테스트 실행

```bash
./board-service/scripts/integration-test.sh
```

이 스크립트는:
- user-service와 board-service의 health check를 수행합니다
- 실제 user-service 로그인 API를 호출하여 JWT 토큰을 발급받습니다
- 발급받은 토큰으로 board-service의 주요 API를 테스트합니다
- 서비스 간 통신 및 인증 플로우를 검증합니다

> **참고**: 이전 버전의 mock 기반 테스트 스크립트(`test-with-mock-user.sh`)는 참고용으로 `.deprecated` 확장자로 보관되어 있습니다.

### 2. 특정 엔드포인트만 빠르게 테스트

```bash
# 프로젝트 목록 조회
./board-service/scripts/quick-test.sh projects

# 프로젝트 생성
./board-service/scripts/quick-test.sh create-project

# 보드 목록 조회
./board-service/scripts/quick-test.sh boards

# user-service 연결 확인
./board-service/scripts/quick-test.sh user-check
```

## 📊 테스트 결과 해석

### 성공적인 user-service 호출

로그에서 다음과 같은 메시지를 확인할 수 있습니다:

```json
{
  "level":"info",
  "message":"Making request to User Service",
  "url":"http://user-service:8080/api/workspaces/.../validate-member/...",
  "has_token":true
}
```

```json
{
  "level":"info",
  "message":"Received response from User Service",
  "status_code":200,
  "response_preview":"{\"workspaceId\":\"...\",\"userId\":\"...\",\"valid\":true}"
}
```

### 500 에러가 발생하는 경우

user-service 호출은 성공했지만 500 에러가 발생한다면:

1. **데이터베이스 문제 확인**
   ```bash
   docker logs wealist-board-service | grep -i "database\|sql\|postgres"
   ```

2. **상세 에러 로그 확인**
   ```bash
   docker logs wealist-board-service --tail 100
   ```

3. **board-service 데이터베이스 연결 확인**
   ```bash
   docker exec wealist-board-service ping wealist-postgres
   ```

## 🔍 문제 해결

### user-service 호출 실패

```bash
# 네트워크 연결 확인
docker network inspect wealist-network

# board-service에서 user-service로 ping
docker exec wealist-board-service ping user-service

# user-service 상태 확인
docker ps | grep user-service
docker logs wealist-user-service --tail 50
```

### 인증 토큰 문제

```bash
# 테스트 토큰 생성 확인
USER_ID="<your-user-id>"
curl http://localhost:8080/api/users/test/${USER_ID}

# 토큰으로 API 호출 테스트
TOKEN="<generated-token>"
curl -H "Authorization: Bearer ${TOKEN}" \
  http://localhost:8000/api/projects?workspaceId=<workspace-id>
```

### 테스트 데이터가 없는 경우

```bash
# user-service 로그에서 데이터 초기화 확인
docker logs wealist-user-service | grep "Dummy data"

# 데이터가 없다면 user-service 재시작 (dev/local 프로파일에서 자동 초기화)
docker-compose restart user-service

# 데이터베이스 직접 확인
docker exec wealist-postgres psql -U wealist_user -d wealist_user_db \
  -c "SELECT user_id, email FROM users LIMIT 5;"
```

## 📝 테스트 유저 정보

user-service의 `DataInitializer`가 자동으로 생성하는 테스트 데이터:

- **유저**: user1@example.com ~ user50@example.com (총 50명)
- **워크스페이스**: 10개 (각 5명씩 배정)
- **기본 테스트 유저**: user1@example.com

### 데이터 구조

```
워크스페이스 1 (테스트 워크스페이스 1)
├── user1@example.com (OWNER, default workspace)
├── user2@example.com (MEMBER)
├── user3@example.com (MEMBER)
├── user4@example.com (MEMBER)
└── user5@example.com (MEMBER)

워크스페이스 2 (테스트 워크스페이스 2)
├── user6@example.com (OWNER, default workspace)
├── user7@example.com (MEMBER)
...
```

## 🛠️ 수동 테스트 예제

### 1. 테스트 유저 ID 가져오기

```bash
docker exec wealist-postgres psql -U wealist_user -d wealist_user_db \
  -c "SELECT user_id, email FROM users WHERE email = 'user1@example.com';"
```

### 2. 워크스페이스 ID 가져오기

```bash
USER_ID="<user-id-from-step-1>"
docker exec wealist-postgres psql -U wealist_user -d wealist_user_db \
  -c "SELECT workspace_id FROM workspace_members WHERE user_id = '${USER_ID}' AND is_default = true;"
```

### 3. 테스트 토큰 생성

```bash
USER_ID="<user-id>"
curl http://localhost:8080/api/users/test/${USER_ID}
```

### 4. board-service API 호출

```bash
TOKEN="<generated-token>"
WORKSPACE_ID="<workspace-id>"

# 프로젝트 목록 조회
curl -H "Authorization: Bearer ${TOKEN}" \
  "http://localhost:8000/api/projects?workspaceId=${WORKSPACE_ID}"

# 프로젝트 생성
curl -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "workspaceId": "'${WORKSPACE_ID}'",
    "projectName": "Test Project",
    "projectDescription": "Test Description",
    "projectColor": "#3B82F6"
  }' \
  http://localhost:8000/api/projects
```

## 📚 관련 스크립트

- `integration-test.sh` - 완전한 통합 테스트 (실제 로그인 API 사용)
- `quick-test.sh` - 빠른 단일 엔드포인트 테스트
- `test-500-error-fix.sh` - 특정 버그 수정 검증
- `test-network-connectivity.sh` - 네트워크 연결 테스트
- `test-with-mock-user.sh.deprecated` - 이전 버전 mock 기반 테스트 (참고용)

자세한 내용은 [scripts/README.md](./scripts/README.md)를 참조하세요.

## 🎓 학습 포인트

이 테스트를 통해 확인할 수 있는 것들:

1. **마이크로서비스 간 통신**
   - board-service → user-service HTTP 호출
   - Docker 네트워크를 통한 서비스 디스커버리
   - 서비스 이름 기반 DNS 해석 (user-service:8080)

2. **인증 흐름**
   - JWT 토큰 생성 및 검증
   - Authorization 헤더를 통한 토큰 전달
   - 서비스 간 토큰 전파

3. **에러 디버깅**
   - 로그 분석을 통한 문제 파악
   - 네트워크 연결 문제 vs 애플리케이션 로직 문제 구분
   - 데이터베이스 쿼리 문제 식별

## 💡 팁

- 테스트 실행 전에 모든 서비스가 healthy 상태인지 확인하세요
- 로그를 실시간으로 보려면: `docker logs -f wealist-board-service`
- 네트워크 문제가 의심되면 `test-network-connectivity.sh`를 먼저 실행하세요
- 500 에러가 발생하면 user-service 호출이 성공했는지 먼저 확인하세요
