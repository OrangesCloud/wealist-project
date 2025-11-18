# Parameter Error Handling Tests

## 개요

이 문서는 CD 워크플로우의 파라미터 에러 처리 및 트러블슈팅 가이드 속성 테스트에 대해 설명합니다.

## 테스트 파일

### 1. test-parameter-error-handling.sh

**목적**: Parameter Store 접근 실패 시 에러 메시지가 올바르게 생성되는지 검증

**검증하는 속성**:
- **Property 2**: Missing parameter error handling
  - 파라미터가 누락되거나 접근 불가능할 때 파라미터 이름이 에러 메시지에 포함되는지 확인
- **Property 4**: Error messages include troubleshooting guidance
  - 모든 파라미터 관련 에러에 트러블슈팅 가이드가 포함되는지 확인

**테스트 케이스**:
1. 누락된 파라미터 에러에 파라미터 이름 포함
2. 누락된 시크릿 파라미터 에러에 파라미터 이름 포함
3. 접근 거부 에러에 파라미터 이름 포함
4. 에러 메시지에 트러블슈팅 가이드 포함
5. 시크릿 파라미터 에러에 KMS 가이드 포함
6. 에러 메시지에 전체 파라미터 경로 포함
7. 랜덤 파라미터 이름으로 50회 반복 테스트
8. 랜덤 파라미터 에러에 트러블슈팅 가이드 포함 확인
9. 빈 파라미터 이름 처리
10. 특수 문자가 포함된 파라미터 이름 처리

**실행 방법**:
```bash
bash scripts/test-parameter-error-handling.sh
```

**예상 출력**:
```
=========================================
Parameter Error Handling Property Tests
=========================================

[TEST] Property 2.1: Missing parameter error includes parameter name
[PASS] Missing parameter error includes parameter name
[TEST] Property 2.2: Missing secret parameter error includes parameter name
[PASS] Missing secret parameter error includes parameter name
...
Tests Run:    10
Tests Passed: 10
Tests Failed: 0

✅ All tests passed!
```

### 2. test-infrastructure-error-handling.sh

**목적**: 인프라 서비스 시작 실패 및 컨테이너 재생성 실패 시 에러 메시지 검증

**검증하는 속성**:
- **Property 4**: Error messages include troubleshooting guidance
  - 인프라 관련 모든 에러에 구체적인 트러블슈팅 가이드가 포함되는지 확인

**테스트 케이스**:
1. 인프라 서비스 시작 실패 시 트러블슈팅 가이드 포함
2. 헬스체크 실패 시 트러블슈팅 가이드 포함
3. 서비스 재생성 실패 시 트러블슈팅 가이드 포함
4. 에러 메시지에 진단 명령어 포함
5. 랜덤 서비스 이름으로 20회 반복 테스트
6. 연속된 여러 실패에 모두 가이드 포함
7. 실패 유형에 따라 구체적인 가이드 제공

**실행 방법**:
```bash
bash scripts/test-infrastructure-error-handling.sh
```

**예상 출력**:
```
=========================================
Infrastructure Error Handling Tests
=========================================

[TEST] Property 4.4: Infrastructure start failure includes troubleshooting
[PASS] Infrastructure start failure includes comprehensive troubleshooting
...
Tests Run:    7
Tests Passed: 7
Tests Failed: 0

✅ All tests passed!
```

## 검증되는 요구사항

### Requirements 2.5
> WHEN a parameter is missing or inaccessible THEN the system SHALL fail with a clear error message indicating which parameter is missing

**검증 방법**:
- `load_param()` 및 `load_secret()` 함수가 파라미터 이름을 에러 메시지에 포함
- 전체 파라미터 경로 표시
- 명확한 실패 메시지 출력

### Requirements 4.5
> WHEN parameters are misconfigured THEN the system SHALL provide troubleshooting guidance in error messages

**검증 방법**:
- 모든 에러 메시지에 "📋 Troubleshooting:" 섹션 포함
- 구체적인 해결 방법 제시:
  - Parameter Store 존재 여부 확인
  - IAM 권한 확인
  - AWS 리전 확인
  - KMS 키 권한 확인 (SecureString의 경우)
  - Docker 데몬 상태 확인 (인프라 에러의 경우)
  - 디스크 공간 확인
  - 로그 확인 명령어 제공

## 에러 메시지 예시

### Parameter Store 접근 실패
```
❌ Failed to load parameter: db/postgres_password
   Full path: /wealist/dev/db/postgres_password
📋 Troubleshooting:
   - Check if parameter exists in Parameter Store
   - Verify IAM role has ssm:GetParameter permission
   - Confirm AWS region is correct: ap-northeast-2
```

### SecureString 파라미터 접근 실패
```
❌ Failed to load secret parameter: jwt/jwt_secret
   Full path: /wealist/dev/jwt/jwt_secret
📋 Troubleshooting:
   - Check if parameter exists in Parameter Store
   - Verify IAM role has ssm:GetParameter permission
   - Confirm parameter type is SecureString
   - Verify KMS key permissions for decryption
   - Confirm AWS region is correct: ap-northeast-2
```

### 인프라 서비스 시작 실패
```
❌ Failed to start infrastructure services
📋 Error details:
Error response from daemon: Cannot start container postgres: ...
📋 Docker Compose status:
NAME                STATUS              PORTS
wealist-postgres    Exited (1)          
📋 Recent logs:
postgres  | Error: Database initialization failed
📋 Troubleshooting:
   - Check if Docker daemon is running
   - Verify docker-compose.yml file is valid
   - Check disk space: df -h
   - Check Docker logs: docker logs <container_name>
   - Verify port availability: netstat -tuln | grep <port>
```

### PostgreSQL 헬스체크 실패
```
❌ PostgreSQL failed to become ready
📋 Container status:
NAME                STATUS              PORTS
wealist-postgres    Exited (1)          
📋 Recent logs:
postgres  | Error: Database initialization failed
📋 Troubleshooting:
   - Check PostgreSQL container logs: docker logs wealist-postgres
   - Verify PostgreSQL configuration
   - Check if data directory is corrupted
   - Verify environment variables are set correctly
   - Check disk space and permissions
```

## Property-Based Testing 접근

### 랜덤 입력 생성
- 랜덤 파라미터 이름 생성 (50회 반복)
- 랜덤 서비스 이름 생성 (20회 반복)
- 특수 문자가 포함된 파라미터 이름 테스트

### 불변 속성 검증
- **Property 2**: 모든 파라미터 에러는 파라미터 이름을 포함해야 함
- **Property 4**: 모든 에러는 트러블슈팅 가이드를 포함해야 함

### 엣지 케이스
- 빈 파라미터 이름
- 특수 문자가 포함된 파라미터 이름
- 연속된 여러 실패
- 다양한 실패 유형 (Parameter Store, Docker, 헬스체크)

## 실제 배포 스크립트와의 관계

이 테스트들은 `.github/workflows/cd-dev-board-service.yml`에 포함된 배포 스크립트의 에러 처리 로직을 검증합니다.

**배포 스크립트의 주요 에러 처리 함수**:
- `load_param()`: 일반 파라미터 로드
- `load_secret()`: SecureString 파라미터 로드
- `start_infrastructure_services()`: PostgreSQL, Redis 시작
- `check_postgres_health()`: PostgreSQL 헬스체크
- `force_recreate_service()`: 서비스 컨테이너 재생성

## CI/CD 통합

이 테스트들은 다음 상황에서 실행되어야 합니다:
1. 배포 스크립트 수정 시
2. 워크플로우 파일 수정 시
3. Pull Request 생성 시
4. 정기적인 회귀 테스트

## 문제 해결

### 테스트 실패 시
1. 에러 메시지를 확인하여 어떤 속성이 위반되었는지 파악
2. 배포 스크립트의 해당 함수 확인
3. 에러 메시지 형식이 요구사항을 충족하는지 검증
4. 필요시 에러 메시지에 누락된 정보 추가

### 새로운 에러 케이스 추가
1. 새로운 테스트 함수 작성
2. Property 2 또는 Property 4 검증
3. 랜덤 입력으로 반복 테스트
4. 엣지 케이스 고려

## 참고 자료

- [Design Document](.kiro/specs/cd-workflow-improvement/design.md)
- [Requirements Document](.kiro/specs/cd-workflow-improvement/requirements.md)
- [CD Workflow](.github/workflows/cd-dev-board-service.yml)
