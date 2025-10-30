# OO Project

프로젝트 관리 플랫폼 (User Service + Kanban Service + Frontend)

## 실행 방법

### 1. 환경 변수 설정

`.env.example` 파일을 참고하여 `.env` 파일을 생성하고 필요한 값을 입력합니다.

```bash
cp .env.example .env
# .env 파일을 열어 필요한 값 수정
```

### 2. Docker Compose 실행

```bash
# 일반 실행
docker compose up -d

# 빌드 캐시 없이 실행 (업데이트 후 문제 발생 시)
docker compose build --no-cache
docker compose up -d
```

### 3. 서비스 확인

- **User Service**: http://localhost:8080
- **User Service Swagger**: http://localhost:8080/swagger-ui/index.html
- **Kanban Service**: http://localhost:8000
- **Kanban Service Swagger**: http://localhost:8000/docs

## 중지 및 삭제

```bash
# 서비스 중지
docker compose down

# 서비스 중지 및 볼륨 삭제
docker compose down -v
```
