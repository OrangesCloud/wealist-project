# User Service

사용자 관리 및 인증을 담당하는 Spring Boot 기반 마이크로서비스입니다.

## 주요 기능

- 사용자 인증 및 권한 관리
- 프로필 관리 (워크스페이스별)
- 워크스페이스 관리
- OAuth 2.0 소셜 로그인
- **Presigned URL 기반 프로필 이미지 업로드**

## API 문서

### Swagger UI

서버 실행 후 다음 URL에서 API 문서를 확인할 수 있습니다:

```
http://localhost:8081/swagger-ui.html
```

### Presigned URL 기반 프로필 이미지 업로드 API

사용자 프로필 이미지를 S3에 직접 업로드할 수 있는 Presigned URL 기반 API를 제공합니다.

**주요 특징:**
- 클라이언트가 S3에 직접 업로드하여 서버 부하 최소화
- 이미지 파일만 지원 (최대 20MB)
- 워크스페이스별 프로필 이미지 관리

**지원 파일 형식:**
- 이미지: jpg, jpeg, png, gif, webp

**API 엔드포인트:**
- `POST /api/profiles/me/image/presigned-url` - Presigned URL 생성
- `PUT /api/profiles/me/image` - 프로필 이미지 업데이트

**상세 가이드:** [docs/PRESIGNED_URL_PROFILE_IMAGE_GUIDE.md](docs/PRESIGNED_URL_PROFILE_IMAGE_GUIDE.md)를 참조하세요.

## 빌드 및 실행

### 로컬 개발

```bash
# 빌드
./gradlew clean build

# 실행
./gradlew bootRun
```

### Docker 환경

#### 1. 컨테이너 중지 및 제거

```bash
docker stop wealist-user-service
docker rm wealist-user-service
```

#### 2. 이미지 재빌드 (Gradle build 후 jar 파일 생성 포함)

```bash
./gradlew clean build
```

#### 3. Docker Compose (또는 run)으로 다시 실행

```bash
docker-compose up -d --build
```

## 환경 설정

주요 환경 변수:

```bash
# Server
SERVER_PORT=8081

# Database
SPRING_DATASOURCE_URL=jdbc:postgresql://localhost:5432/user_db
SPRING_DATASOURCE_USERNAME=postgres
SPRING_DATASOURCE_PASSWORD=password

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=86400000

# S3 (Presigned URL 업로드용)
AWS_S3_BUCKET=wealist-dev-files
AWS_S3_REGION=ap-northeast-2
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
```

## 테스트

```bash
# 모든 테스트 실행
./gradlew test

# 특정 테스트 실행
./gradlew test --tests ProfileImageControllerTest
```

