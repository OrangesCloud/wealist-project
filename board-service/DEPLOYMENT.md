# Board Service 배포 가이드

## CORS 설정 업데이트 완료

Board Service의 CORS 설정이 다음과 같이 업데이트되었습니다:

### 변경 사항
- CloudFront 도메인 패턴 매칭 추가 (`*.cloudfront.net`)
- 로컬 개발 환경 추가 (`http://localhost:5173`)
- 프로덕션 도메인 추가 (`https://wealist.co.kr`)
- Origin 헤더 기반 동적 CORS 설정

### 허용된 Origin 목록
1. `https://wealist.co.kr` - 프로덕션 도메인
2. `http://localhost:5173` - Vite 개발 서버
3. `http://localhost:3000` - 기존 개발 환경
4. `https://*.cloudfront.net` - CloudFront 배포 도메인 (패턴 매칭)

## 배포 방법

### 방법 1: GitHub Actions를 통한 자동 배포 (권장)

1. **코드를 deploy-dev 브랜치에 푸시**
   ```bash
   git add board-service/internal/middleware/cors.go
   git commit -m "feat: Update Board Service CORS configuration for ALB integration"
   git push origin deploy-dev
   ```

2. **CI 워크플로우가 자동으로 실행됩니다**
   - 코드 빌드 및 테스트
   - Docker 이미지 빌드
   - ECR에 이미지 푸시

3. **CD 워크플로우가 자동으로 실행됩니다**
   - EC2에서 최신 이미지 Pull
   - 컨테이너 재시작
   - Health Check 수행

### 방법 2: 수동 배포

#### 사전 요구사항
- AWS CLI 설치 및 구성
- Docker 설치
- ECR 접근 권한

#### 배포 단계

1. **AWS 자격 증명 설정**
   ```bash
   export AWS_REGION=ap-northeast-2
   export AWS_ACCOUNT_ID=<your-account-id>
   ```

2. **ECR 로그인**
   ```bash
   aws ecr get-login-password --region ${AWS_REGION} | \
     docker login --username AWS --password-stdin \
     ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com
   ```

3. **Docker 이미지 빌드**
   ```bash
   cd board-service
   docker build -f docker/Dockerfile -t wealist-board-service:latest .
   ```

4. **이미지 태그 지정**
   ```bash
   docker tag wealist-board-service:latest \
     ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/wealist-dev-board-service:latest
   ```

5. **ECR에 푸시**
   ```bash
   docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/wealist-dev-board-service:latest
   ```

6. **EC2에서 컨테이너 재시작**
   ```bash
   # EC2에 SSH 접속
   ssh ubuntu@<ec2-ip>
   
   # 최신 이미지 Pull
   aws ecr get-login-password --region ap-northeast-2 | \
     docker login --username AWS --password-stdin \
     ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-2.amazonaws.com
   
   docker pull ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-2.amazonaws.com/wealist-dev-board-service:latest
   
   # Docker Compose로 재시작
   cd /home/ubuntu/wealist
   docker-compose -f docker/compose/docker-compose.ec2-dev.yml up -d --force-recreate board-service
   ```

7. **Health Check 확인**
   ```bash
   curl http://localhost:8000/health
   ```

## 검증

### CORS 설정 테스트

1. **Preflight 요청 테스트**
   ```bash
   curl -X OPTIONS https://api.wealist.co.kr/api/boards \
     -H "Origin: http://localhost:5173" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Authorization" \
     -v
   ```

   예상 응답 헤더:
   ```
   Access-Control-Allow-Origin: http://localhost:5173
   Access-Control-Allow-Credentials: true
   Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS
   Access-Control-Max-Age: 43200
   ```

2. **실제 API 요청 테스트**
   ```bash
   curl -X GET https://api.wealist.co.kr/api/boards \
     -H "Origin: http://localhost:5173" \
     -H "Authorization: Bearer <token>" \
     -v
   ```

3. **CloudFront 도메인 테스트**
   ```bash
   curl -X OPTIONS https://api.wealist.co.kr/api/boards \
     -H "Origin: https://d1234567890.cloudfront.net" \
     -H "Access-Control-Request-Method: GET" \
     -v
   ```

### 로컬 개발 환경 테스트

1. **프론트엔드 실행**
   ```bash
   cd frontend
   npm run dev
   ```

2. **브라우저에서 확인**
   - 개발자 도구 > 네트워크 탭 열기
   - API 호출 시 CORS 오류가 없는지 확인
   - Response Headers에 `Access-Control-Allow-Origin: http://localhost:5173` 확인

## 트러블슈팅

### CORS 오류가 계속 발생하는 경우

1. **Origin 헤더 확인**
   ```bash
   # 브라우저 개발자 도구 > 네트워크 탭에서 Request Headers 확인
   Origin: http://localhost:5173
   ```

2. **서버 로그 확인**
   ```bash
   docker logs wealist-board-service
   ```

3. **컨테이너 재시작**
   ```bash
   docker-compose -f docker/compose/docker-compose.ec2-dev.yml restart board-service
   ```

### Health Check 실패

1. **컨테이너 상태 확인**
   ```bash
   docker ps -a | grep board-service
   ```

2. **로그 확인**
   ```bash
   docker logs --tail=100 wealist-board-service
   ```

3. **포트 확인**
   ```bash
   netstat -tlnp | grep 8000
   ```

## 다음 단계

CORS 설정 업데이트가 완료되었습니다. 다음 작업을 진행하세요:

1. ✅ Task 6.1: cors.go 미들웨어 수정 (완료)
2. ✅ Task 6.2: Docker 이미지 빌드 (완료)
3. ⏳ Task 6.2: ECR 푸시 및 배포 (GitHub Actions 또는 수동 배포 필요)

배포 후 Task 7 "SSM Parameter Store 설정"으로 진행하세요.
