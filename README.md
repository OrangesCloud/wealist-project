# weAlist BASIC Project

프로젝트 관리 플랫폼 - 마이크로서비스 아키텍처


## 🏗️ 서비스 구조

| 서비스 | 기술 스택 | 포트 | 상태 | 설명 |
|--------|----------|------|------|------|
| **User Service** | Spring Boot (Java) | 8080 | ✅ Active | 사용자 인증 및 관리 |
| **Board Service** | Gin (Go) | 8000 | ✅ Active | 보드/칸반 관리, 커스텀 필드 |
| **Frontend** | React (TypeScript) | 3000 | 🚧 Dev | 프론트엔드 애플리케이션 |

## 🌐 ALB Path-Based Routing

AWS 환경에서는 Application Load Balancer(ALB)를 통해 서비스별 경로 기반 라우팅을 제공합니다.

### 라우팅 구조

```
Client Request
    ↓
ALB (https://api.wealist.co.kr)
    ├─ /api/users/*   → User Service (port 8080)
    └─ /api/boards/*  → Board Service (port 8000)
```

### 환경별 설정

#### 로컬 개발 환경
- ALB 없이 직접 서비스 접근
- Context/Base Path 설정 없음

```bash
# User Service
curl http://localhost:8080/api/workspaces/all

# Board Service
curl http://localhost:8000/health
```

#### AWS 환경 (Production)
- ALB를 통한 통합 엔드포인트
- 서비스별 prefix 자동 제거

```bash
# User Service (ALB → /api/users 제거 → User Service)
curl https://api.wealist.co.kr/api/users/api/workspaces/all

# Board Service (ALB → /api/boards 제거 → Board Service)
curl https://api.wealist.co.kr/api/boards/health
```

### API 엔드포인트 예시

| 환경 | User Service | Board Service |
|------|-------------|---------------|
| **로컬** | `http://localhost:8080/api/workspaces/all` | `http://localhost:8000/health` |
| **AWS** | `https://api.wealist.co.kr/api/users/api/workspaces/all` | `https://api.wealist.co.kr/api/boards/health` |

### 서비스별 Path 설정

**User Service (Spring Boot)**
- AWS 환경: `server.servlet.context-path=/api/users` (application-aws.yml)
- 로컬 환경: Context path 없음 (application-local.yml)
- Profile 전환: `SPRING_PROFILES_ACTIVE` 환경 변수

**Board Service (Go)**
- AWS 환경: `SERVER_BASE_PATH=/api/boards` 환경 변수
- 로컬 환경: `SERVER_BASE_PATH=""` (빈 문자열)
- 환경 전환: `ENV` 환경 변수

자세한 배포 가이드는 [docs/ALB_ROUTING_DEPLOYMENT.md](docs/ALB_ROUTING_DEPLOYMENT.md)를 참조하세요.

## 🚀 주요 기능

- ✅ 워크스페이스 & 프로젝트 관리
- ✅ 커스텀 보드 (역할, 진행단계, 중요도 기반)
- ✅ 드래그 앤 드롭 기능 (사용자별 순서 저장)
- ✅ 멤버 관리 및 역할 기반 접근 제어
- ✅ JWT 기반 인증
- ✅ 소프트 삭제 (복구 가능)
- ✅ RESTful API with Swagger

## 📋 실행 방법

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

- **User Service**: http://localhost:8080/health
- **User Service Swagger**: http://localhost:8080/swagger-ui/index.html
- **Board Service**: http://localhost:8000/health
- **Board Service Swagger** (dev 모드만): http://localhost:8000/swagger/index.html

### 4. 테스트

Board Service 통합 테스트:
```bash
cd scripts/board_test_script
./test_board_service.sh
```

### 5. 추가 정보

- **API 테스트**: Board Service Swagger UI에서 직접 테스트 가능

## ⚙️ 중지 및 삭제

```bash
# 서비스 중지
docker compose down

# 서비스 중지 및 볼륨 삭제 (데이터베이스 초기화)
docker compose down -v
```

## 🛠️ 개발 가이드

### 디렉토리 구조

```
wealist/
├── user/               # User Service (Spring Boot)
├── board-service/      # Board Service (Go)
├── frontend/           # Frontend (React)
├── scripts/            # 테스트 스크립트
├── docker-compose.yaml # 서비스 오케스트레이션
├── CLAUDE.md          # 프로젝트 전체 가이드
└── README.md          # 이 파일
```

### 개발 시 주의사항

- **Board Service (Go)** 사용 권장 -
- JWT 토큰은 User Service와 Board Service 간 공유 (`SECRET_KEY` 일치 필요)
- 모든 ID는 UUID 타입 사용
- Foreign Key 없음 (샤딩 대비, 애플리케이션 레벨에서 관계 관리)
- Soft Delete 방식 (`is_deleted` 플래그)

- **User Service API**: [.claude/api-user-documentation.md](./.claude/api-user-documentation.md)

## 📦 기술 스택

### Backend
- **User Service**: Spring Boot 3.x, Java 17, Spring Security, JWT
- **Board Service**: Go 1.21+, Gin, GORM, Viper, Zap Logger

### Database & Cache
- **PostgreSQL 17**: 각 서비스별 독립 DB
- **Redis 7**: 캐싱 및 세션 관리

### Frontend
- **React 18**: TypeScript, Tailwind CSS

### DevOps
- **Docker & Docker Compose**: 컨테이너 오케스트레이션
- **Git**: 모노레포 구조

