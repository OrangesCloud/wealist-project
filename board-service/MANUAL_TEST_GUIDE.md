# Board Service 인증 미들웨어 수동 테스트 가이드

## 현재 상태
✅ **테스트 1 완료**: 인증되지 않은 요청이 401을 반환하는 것 확인됨
```bash
curl http://localhost:8000/api/projects
# 결과: {"error":{"code":"UNAUTHORIZED","message":"Authorization header is required"},"message":"인증이 필요합니다"}
# HTTP Status: 401
```

## 프론트엔드에서 확인하는 방법

### 1. 프론트엔드 실행
```bash
cd frontend
npm run dev
```

### 2. 브라우저에서 테스트

#### 테스트 2: 로그인 후 인증된 요청이 성공하는지 확인
1. 브라우저에서 `http://localhost:3000` 접속
2. 로그인 페이지에서 로그인
3. 프로젝트 목록이나 보드 목록 페이지로 이동
4. **예상 결과**: 데이터가 정상적으로 로드됨 (401 에러 없음)

#### 테스트 3: Health Check 엔드포인트 확인
브라우저에서 직접 접속:
```
http://localhost:8000/health
```
**예상 결과**: 인증 없이 접근 가능, 상태 정보 표시

#### 테스트 4: Swagger 문서 엔드포인트 확인
브라우저에서 직접 접속:
```
http://localhost:8000/swagger/index.html
```
**예상 결과**: 인증 없이 접근 가능, Swagger UI 표시

## 브라우저 개발자 도구로 확인

### Chrome/Firefox 개발자 도구 사용
1. `F12` 또는 `Cmd+Option+I` (Mac)로 개발자 도구 열기
2. **Network** 탭 선택
3. 프론트엔드에서 API 호출 (프로젝트 목록 조회 등)
4. 요청 확인:
   - Request Headers에 `Authorization: Bearer <token>` 있는지 확인
   - Response Status가 `200 OK`인지 확인 (401이 아님)

### 로그아웃 후 테스트
1. 로그아웃
2. 브라우저 콘솔에서 직접 API 호출:
```javascript
fetch('http://localhost:8000/api/projects')
  .then(r => r.json())
  .then(console.log)
```
**예상 결과**: 401 에러와 "Authorization header is required" 메시지

## 터미널에서 간단 테스트 (선택사항)

### 인증 없이 요청 (401 예상)
```bash
curl -v http://localhost:8000/api/projects
```

### Health Check (200 예상)
```bash
curl http://localhost:8000/health
```

### Swagger (200 예상)
```bash
curl http://localhost:8000/swagger/index.html
```

## 테스트 체크리스트

- [x] 테스트 1: 인증되지 않은 요청이 401 반환 ✅
- [ ] 테스트 2: 로그인 후 인증된 요청 성공
- [ ] 테스트 3: Health check 엔드포인트 인증 없이 접근 가능
- [ ] 테스트 4: Swagger 문서 엔드포인트 인증 없이 접근 가능

## 문제 발생 시

### 401 에러가 계속 발생하는 경우
- 프론트엔드가 Authorization 헤더를 제대로 보내는지 확인
- 토큰이 만료되지 않았는지 확인
- 다시 로그인 시도

### Health check나 Swagger가 401을 반환하는 경우
- 라우터 설정 확인 필요 (이 엔드포인트들은 인증 미들웨어 밖에 있어야 함)
