# Board Service - 클라우드 네이티브 아키텍처

## 목차

- [개요](#개요)
- [클라우드 네이티브 설계 원칙](#클라우드-네이티브-설계-원칙)
- [마이크로서비스 아키텍처](#마이크로서비스-아키텍처)
- [Clean Architecture 구현](#clean-architecture-구현)
- [데이터베이스 설계](#데이터베이스-설계)
- [API 설계](#api-설계)
- [배포 아키텍처](#배포-아키텍처)
- [확장성 및 성능](#확장성-및-성능)

## 개요

Board Service는 **클라우드 네이티브 원칙**을 따라 설계된 마이크로서비스로, 프로젝트 관리 플랫폼 weAlist의 핵심 보드 관리 기능을 담당합니다.

### 핵심 설계 목표

- **독립적 배포**: User Service와 완전히 독립적으로 배포 및 확장 가능
- **무상태(Stateless)**: 모든 요청은 독립적으로 처리되어 수평 확장 가능
- **복원력(Resilience)**: 장애 격리 및 graceful degradation 지원
- **관찰 가능성(Observability)**: 구조화된 로깅, 메트릭, 헬스 체크 제공

## 클라우드 네이티브 설계 원칙

### 1. 12-Factor App 준수

| 요소 | 구현 방식 |
|------|----------|
| **코드베이스** | 단일 Git 저장소에서 모든 환경 배포 |
| **의존성** | Go modules로 명시적 선언 및 격리 |
| **설정** | 환경 변수로 분리 (.env, config.yaml) |
| **백엔드 서비스** | PostgreSQL, Redis를 첨부된 리소스로 취급 |
| **빌드/배포** | Docker 멀티스테이지 빌드로 완전 분리 |
| **프로세스** | 무상태 실행, 모든 상태는 DB/Redis에 저장 |
| **포트 바인딩** | 자체 HTTP 서버 (Gin) 내장 |
| **동시성** | 수평 확장으로 처리량 증가 |
| **폐기 용이성** | Graceful shutdown, 빠른 시작 시간 |
| **개발/운영 환경 일치** | Docker Compose로 동일 환경 구성 |
| **로그** | 이벤트 스트림으로 stdout 출력 (Zap) |
| **관리 프로세스** | 마이그레이션 스크립트, 일회성 작업 |

### 2. 마이크로서비스 패턴

#### 서비스 분리 원칙
```
User Service (Spring Boot)          Board Service (Go)
├─ 인증/인가                         ├─ 프로젝트 관리
├─ 워크스페이스 관리                 ├─ 보드 CRUD
├─ 사용자 프로필                     ├─ 커스텀 필드
└─ OAuth2 통합                       └─ 댓글/첨부파일

         ↓ RESTful API ↓
         JWT 토큰 기반 통신
```

#### 통신 방식
- **동기 통신**: RESTful HTTP API (현재)
- **인증**: JWT 토큰 검증 (공유 Secret Key)
- **서비스 디스커버리**: Docker Compose DNS / ALB 라우팅

### 3. 데이터 독립성

```
┌──────────────────┐         ┌──────────────────┐
│  User Service    │         │  Board Service   │
├──────────────────┤         ├──────────────────┤
│ wealist_user_db  │         │ wealist_board_db │
│ - users          │         │ - projects       │
│ - workspaces     │         │ - boards         │
│ - members        │         │ - comments       │
└──────────────────┘         └──────────────────┘
        ↑                             ↑
        └─────────────────────────────┘
          서로의 DB 직접 접근 금지
          API 호출로만 데이터 교환
```

**핵심 원칙**:
- ❌ **No Foreign Keys**: 서비스 간 DB 외래 키 없음
- ✅ **Application-level Join**: 필요 시 API 호출로 데이터 결합
- ✅ **Eventually Consistent**: 최종 일관성 허용
- ✅ **Autonomous**: 각 서비스는 독립적으로 동작

## Clean Architecture 구현

### 계층 구조

```
┌─────────────────────────────────────────────────────┐
│                   Presentation                       │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐│
│  │   Handler    │  │  Middleware  │  │   Router   ││
│  │  (Gin HTTP)  │  │   (Auth)     │  │  (Routes)  ││
│  └──────────────┘  └──────────────┘  └────────────┘│
└───────────────────────┬─────────────────────────────┘
                        │ DTO
┌───────────────────────▼─────────────────────────────┐
│                 Business Logic                       │
│  ┌──────────────────────────────────────────────┐  │
│  │              Service Layer                   │  │
│  │  - 비즈니스 규칙 검증                         │  │
│  │  - 트랜잭션 관리 (Unit of Work)              │  │
│  │  - 외부 서비스 호출 (User Service)           │  │
│  └──────────────────────────────────────────────┘  │
└───────────────────────┬─────────────────────────────┘
                        │ Interface
┌───────────────────────▼─────────────────────────────┐
│                  Data Access                         │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐│
│  │  Repository  │  │   Database   │  │   Cache    ││
│  │    (GORM)    │  │ (PostgreSQL) │  │  (Redis)   ││
│  └──────────────┘  └──────────────┘  └────────────┘│
└───────────────────────┬─────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────┐
│                    Domain                            │
│  ┌──────────────────────────────────────────────┐  │
│  │           Domain Models (Entities)           │  │
│  │  - Rich Domain Models (26 methods)          │  │
│  │  - 비즈니스 로직 캡슐화                       │  │
│  │  - 불변성 보장                                │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### 의존성 규칙

```
Handler ─depends on→ Service Interface
Service ─depends on→ Repository Interface
Repository ─depends on→ Domain Model
Domain ─depends on→ Nothing (순수 비즈니스 로직)
```

**핵심 원칙**:
- 의존성은 항상 **안쪽 방향**으로만 흐름
- Domain 계층은 **어떤 외부 프레임워크도 모름**
- Interface를 통한 **의존성 역전 (DIP)** 적용

### 주요 패턴

#### 1. Rich Domain Model
```go
// 빈약한 모델 (Anemic Model) ❌
type Board struct {
    ID    string
    Title string
}

// 풍부한 도메인 모델 (Rich Domain Model) ✅
type Board struct {
    id          string
    title       string
    content     string
    attachments []Attachment
}

// 비즈니스 로직을 도메인에 캡슐화
func (b *Board) AddAttachment(att Attachment) error {
    if b.IsDeleted() {
        return ErrBoardDeleted
    }
    if len(b.attachments) >= MaxAttachments {
        return ErrTooManyAttachments
    }
    b.attachments = append(b.attachments, att)
    return nil
}
```

#### 2. Repository Pattern with Generics
```go
// 공통 CRUD 연산을 제네릭으로 추상화
type BaseRepository[T domain.Entity] interface {
    Create(ctx context.Context, entity *T) error
    FindByID(ctx context.Context, id string) (*T, error)
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id string) error
}

// 도메인별 특화 메서드 추가
type BoardRepository interface {
    BaseRepository[domain.Board]
    FindByProjectID(ctx context.Context, projectID string) ([]*domain.Board, error)
}
```

#### 3. Unit of Work Pattern
```go
// 여러 리포지토리 작업을 하나의 트랜잭션으로 묶음
type UnitOfWork interface {
    BeginTransaction(ctx context.Context) (Transaction, error)
}

type Transaction interface {
    Boards() repository.BoardRepository
    Comments() repository.CommentRepository
    Commit() error
    Rollback() error
}

// 사용 예시
func (s *boardService) CreateBoardWithComments(ctx context.Context, req CreateRequest) error {
    tx, _ := s.uow.BeginTransaction(ctx)
    defer tx.Rollback()

    board, _ := tx.Boards().Create(ctx, newBoard)
    comment, _ := tx.Comments().Create(ctx, newComment)

    return tx.Commit()
}
```

#### 4. Centralized Authorization
```go
// 권한 검증 로직을 한 곳에서 관리
type ProjectAuthorizer interface {
    CheckProjectAccess(ctx context.Context, userID, projectID string) error
    CheckBoardAccess(ctx context.Context, userID, boardID string) error
}

// 모든 서비스에서 재사용
func (s *boardService) GetBoard(ctx context.Context, userID, boardID string) error {
    if err := s.authorizer.CheckBoardAccess(ctx, userID, boardID); err != nil {
        return err
    }
    // ...
}
```

## 데이터베이스 설계

### 스키마 구조

```sql
-- UUID 기반 독립적 식별자
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 프로젝트 (워크스페이스 소속)
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL,  -- User Service 워크스페이스 참조 (FK 없음)
    name VARCHAR(255) NOT NULL,
    is_default BOOLEAN DEFAULT false,
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 보드 (프로젝트 소속)
CREATE TABLE boards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    position VARCHAR(50) NOT NULL,  -- Fractional indexing
    stage VARCHAR(50),
    importance VARCHAR(50),
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 보드 참여자
CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    board_id UUID NOT NULL,
    user_id UUID NOT NULL,  -- User Service 사용자 참조 (FK 없음)
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(board_id, user_id)  -- 중복 참여 방지
);

-- 댓글
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    board_id UUID NOT NULL,
    user_id UUID NOT NULL,  -- User Service 사용자 참조 (FK 없음)
    content TEXT NOT NULL,
    deleted_at TIMESTAMP,  -- GORM soft delete
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 데이터베이스 전략

#### 1. Soft Delete
```go
// 모든 엔티티는 is_deleted 플래그 사용 (Comment 제외)
type BaseModel struct {
    ID        string    `gorm:"type:uuid;primary_key"`
    IsDeleted bool      `gorm:"default:false"`
    CreatedAt time.Time `gorm:"not null"`
    UpdatedAt time.Time `gorm:"not null"`
}

// Comment만 GORM의 DeletedAt 사용
type Comment struct {
    ID        string         `gorm:"type:uuid;primary_key"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 장점:
// ✅ 데이터 복구 가능
// ✅ 감사 추적 (Audit Trail)
// ✅ 참조 무결성 유지 (삭제된 데이터도 조회 가능)
```

#### 2. No Foreign Keys (샤딩 준비)
```
┌─────────────────────────────────────────────────┐
│ Why No Foreign Keys?                            │
├─────────────────────────────────────────────────┤
│ ✅ 데이터베이스 샤딩 가능 (수평 확장)            │
│ ✅ 마이크로서비스 독립성 보장                     │
│ ✅ 서비스 간 순환 종속성 방지                     │
│ ✅ 스키마 변경 유연성                            │
│                                                 │
│ Trade-offs:                                     │
│ ⚠️  참조 무결성은 애플리케이션 레벨에서 관리      │
│ ⚠️  Orphaned records 방지 로직 필요              │
└─────────────────────────────────────────────────┘
```

#### 3. Fractional Indexing (정렬 최적화)
```
기존 방식 (Integer Position):
Board A: position=1
Board B: position=2  → A와 B 사이에 삽입 시 모든 레코드 UPDATE
Board C: position=3
...
Board Z: position=26

Fractional Indexing:
Board A: position="a0"
Board B: position="a0V"  → 삽입 시 해당 레코드만 UPDATE
Board C: position="a1"
...
Board Z: position="z"

장점:
✅ O(1) 시간 복잡도로 위치 변경
✅ 데이터베이스 락 최소화
✅ 동시성 성능 향상
```

#### 4. 마이그레이션 관리
```bash
migrations/
├── 001_init_schema.sql              # 초기 스키마
├── 002_add_project_members.sql      # 멤버 테이블 추가
├── 003_migrate_owners.sql           # 데이터 마이그레이션
└── *_down.sql                       # 롤백 스크립트

# 개발 환경: GORM AutoMigrate
ENV=dev USE_AUTO_MIGRATE=true

# 프로덕션: 수동 마이그레이션
./scripts/db/apply_migrations.sh prod
```

## API 설계

### RESTful 원칙

```
리소스 중심 설계:
GET    /api/boards              # 보드 목록 조회
POST   /api/boards              # 보드 생성
GET    /api/boards/{id}         # 보드 상세 조회
PUT    /api/boards/{id}         # 보드 수정
DELETE /api/boards/{id}         # 보드 삭제 (soft delete)

하위 리소스:
GET    /api/boards/{id}/comments       # 보드의 댓글 목록
POST   /api/boards/{id}/comments       # 댓글 작성
GET    /api/boards/{id}/attachments    # 보드의 첨부파일 목록

동사 기반 (특수 액션):
PUT    /api/boards/{id}/move    # 보드 위치 이동 (Fractional Indexing)
```

### 응답 형식 표준화

```json
// 성공 응답
{
  "data": {
    "board_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Implement Authentication"
  },
  "message": "Board created successfully"
}

// 에러 응답 (RFC 7807 Problem Details 기반)
{
  "error": {
    "code": "BOARD_NOT_FOUND",
    "message": "Board with ID 550e8400-... not found",
    "details": "The board may have been deleted"
  },
  "message": "Failed to retrieve board"
}
```

### API 버전 관리

```
현재: /api/boards (버전 없음)
향후: /api/v2/boards (호환성 깨지는 변경 시)

Breaking Changes:
- 필드명 변경
- 필수 파라미터 추가
- 응답 구조 변경

Non-Breaking Changes (버전 불필요):
- 선택적 파라미터 추가
- 새 엔드포인트 추가
- 응답 필드 추가
```

### Swagger/OpenAPI 문서화

```go
// @Summary      보드 생성
// @Description  새로운 보드를 생성합니다
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBoardRequest true "보드 생성 요청"
// @Success      201 {object} response.SuccessResponse{data=dto.BoardResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /api/boards [post]
func (h *BoardHandler) CreateBoard(c *gin.Context) {
    // ...
}

// 자동 생성: http://localhost:8000/swagger/index.html
```

## 배포 아키텍처

weAlist는 **3가지 배포 환경**을 지원합니다:

### 환경 비교

| 환경 | 용도 | 인프라 | 데이터베이스 | 접근 방식 |
|------|------|--------|-------------|-----------|
| **Local** | 개발/디버깅 | Docker Compose | 로컬 PostgreSQL | `localhost:8000` |
| **EC2 Dev** | 통합 테스트 | EC2 단일 인스턴스 | EC2 내 PostgreSQL | `http://<EC2-IP>:8000` |
| **Production** | 운영 서비스 | ECS/EC2 + ALB | RDS PostgreSQL | `https://api.wealist.co.kr/api/boards` |

---

### 1. Local 개발 환경

**목적**: 로컬 머신에서 빠른 개발 및 디버깅

**특징**:
- ✅ 모든 포트 노출 (디버깅 용이)
- ✅ Hot reload 지원 (코드 변경 시 자동 재시작)
- ✅ 개발자별 독립 환경
- ⚠️ Docker Desktop 필요

**구성**:
```yaml
# docker/compose/docker-compose.yml + docker-compose.dev.yml
services:
  postgres:
    image: postgres:17-alpine
    ports:
      - "5432:5432"    # 호스트 접근 가능
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"    # 호스트 접근 가능

  user-service:
    build: ./user-service
    ports:
      - "8080:8080"
    environment:
      - SPRING_PROFILES_ACTIVE=local
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/wealist_user_db

  board-service:
    build: ./board-service
    ports:
      - "8000:8000"
    environment:
      - ENV=dev
      - SERVER_BASE_PATH=""                    # ← ALB 없음
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/wealist_board_db
      - USER_SERVICE_URL=http://user-service:8080
```

**실행 방법**:
```bash
# 프로젝트 루트에서
./docker/scripts/dev.sh up

# 접속
curl http://localhost:8000/health
curl http://localhost:8000/swagger/index.html
```

**환경 변수**:
```bash
ENV=dev
SERVER_BASE_PATH=""
DATABASE_URL=postgresql://postgres:password@postgres:5432/wealist_board_db
USER_SERVICE_URL=http://user-service:8080
USE_AUTO_MIGRATE=true    # 자동 마이그레이션
```

---

### 2. EC2 Dev 환경 (단일 인스턴스)

**목적**: 팀 통합 테스트 및 QA 환경

**특징**:
- ✅ All-in-one 구성 (서비스 + DB + 모니터링)
- ✅ CI/CD 자동 배포
- ✅ 낮은 비용 (~$15-20/월)
- ⚠️ 단일 장애점 (개발용)
- ⚠️ Production 사용 비권장

**구성**:
```
t3.small EC2 (2 vCPU, 2GB RAM)
├── Docker Compose
│   ├── user-service (Spring Boot)
│   ├── board-service (Go)
│   ├── postgres (PostgreSQL 17)     ← EC2 내부
│   └── redis (Redis 7)              ← EC2 내부
├── Prometheus (모니터링)
└── Grafana (대시보드)
```

**Docker Compose 파일**:
```yaml
# docker/compose/docker-compose.ec2-dev.yml
services:
  postgres:
    image: postgres:17-alpine
    # 포트 노출 안함 (내부 통신만)
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  board-service:
    image: ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-2.amazonaws.com/wealist-dev-board-service:latest
    ports:
      - "8000:8000"    # EC2 public IP로 접근
    environment:
      - ENV=dev
      - SERVER_BASE_PATH=""              # ← ALB 없음
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/wealist_board_db
    restart: unless-stopped
```

**CI/CD 자동 배포**:
```
GitHub Push (main)
    ↓
CI: Build & Test
    ↓
Docker Image → ECR
    ↓
CD: Deploy to EC2
    ├─ Parameter Store에서 환경 변수 로드
    ├─ ECR에서 이미지 풀
    ├─ 데이터베이스 마이그레이션 실행
    ├─ board-service 재시작
    └─ Health check (60초 타임아웃)
```

**접근 방법**:
```bash
# EC2 Public IP로 직접 접근
curl http://<EC2-PUBLIC-IP>:8000/health
curl http://<EC2-PUBLIC-IP>:8000/api/boards/...
```

**환경 변수**:
```bash
ENV=dev
SERVER_BASE_PATH=""                   # ALB 없음
DATABASE_URL=postgresql://postgres:password@postgres:5432/wealist_board_db
USER_SERVICE_URL=http://user-service:8080
USE_AUTO_MIGRATE=false               # 수동 마이그레이션
```

---

### 3. Production 환경 (AWS with ALB)

**목적**: 실제 운영 서비스

**특징**:
- ✅ 고가용성 (Multi-AZ)
- ✅ Auto Scaling
- ✅ 관리형 데이터베이스 (RDS)
- ✅ 관리형 캐시 (ElastiCache)
- ✅ ALB를 통한 통합 API 엔드포인트
- ✅ HTTPS/SSL 지원

**AWS 인프라 아키텍처**:

```
                    Internet
                       │
                       ▼
        ┌──────────────────────────────┐
        │   Route 53 (DNS)             │
        │   api.wealist.co.kr          │
        └──────────────┬───────────────┘
                       │
                       ▼
        ┌──────────────────────────────┐
        │   Application Load Balancer  │
        │   - HTTPS/SSL Termination    │
        │   - Path-based Routing       │
        └──────────┬───────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
        ▼                     ▼
┌──────────────┐      ┌──────────────┐
│ Target Group │      │ Target Group │
│ /api/users/* │      │/api/boards/* │
└──────┬───────┘      └──────┬───────┘
       │                     │
       ▼                     ▼
┌──────────────┐      ┌──────────────┐
│ User Service │      │Board Service │
│ ECS/EC2      │◄────►│ ECS/EC2      │
│ Auto Scaling │      │ Auto Scaling │
└──────┬───────┘      └──────┬───────┘
       │                     │
       └──────────┬──────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
        ▼                   ▼
┌──────────────┐    ┌──────────────┐
│ Amazon RDS   │    │ElastiCache   │
│ PostgreSQL   │    │   Redis      │
│ Multi-AZ     │    │   Cluster    │
└──────────────┘    └──────────────┘
```

**Docker Compose 파일**:
```yaml
# docker/compose/docker-compose.yml + docker-compose.prod.yml
services:
  board-service:
    image: ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-2.amazonaws.com/wealist-prod-board-service:latest
    # 포트는 ALB가 관리 (내부 통신만)
    environment:
      - ENV=prod
      - SERVER_BASE_PATH="/api/boards"     # ← ALB 경로 재작성
      - DATABASE_URL=${RDS_DATABASE_URL}   # ← RDS 엔드포인트
      - REDIS_URL=${ELASTICACHE_URL}       # ← ElastiCache 엔드포인트
      - USER_SERVICE_URL=http://user-service-internal:8080
    restart: always
    logging:
      driver: awslogs
      options:
        awslogs-group: /ecs/board-service
        awslogs-region: ap-northeast-2
```

**ALB Path-Based Routing**:

```
클라이언트 요청:
https://api.wealist.co.kr/api/boards/550e8400-e29b-41d4-a716-446655440000

ALB 처리:
┌────────────────────────────────────────────────────────┐
│ Listener Rule: Path = /api/boards/*                    │
│ Action:                                                │
│   1. Forward to board-service target group             │
│   2. Path rewrite: /api/boards/* → /*                  │
└────────────────────────────────────────────────────────┘

Board Service 수신:
GET /550e8400-e29b-41d4-a716-446655440000
```

**환경별 설정 비교**:

| 설정 항목 | Local | EC2 Dev | Production |
|----------|-------|---------|------------|
| `ENV` | `dev` | `dev` | `prod` |
| `SERVER_BASE_PATH` | `""` | `""` | `"/api/boards"` |
| `DATABASE_URL` | 로컬 PostgreSQL | EC2 내 PostgreSQL | RDS 엔드포인트 |
| `REDIS_URL` | 로컬 Redis | EC2 내 Redis | ElastiCache 엔드포인트 |
| `USE_AUTO_MIGRATE` | `true` | `false` | `false` |
| Swagger UI | ✅ 활성화 | ✅ 활성화 | ❌ 비활성화 |
| 로그 레벨 | `debug` | `info` | `warn` |

**접근 방법**:
```bash
# HTTPS로 접근 (ALB SSL Termination)
curl https://api.wealist.co.kr/api/boards/health

# ALB가 /api/boards 경로를 제거하고 board-service로 전달
# board-service는 /health로 요청 수신
```

**환경 변수**:
```bash
ENV=prod
SERVER_BASE_PATH="/api/boards"       # ALB 경로 재작성 대응
DATABASE_URL=postgresql://user:pass@rds-endpoint.ap-northeast-2.rds.amazonaws.com:5432/wealist_board_db
REDIS_URL=redis://elasticache-endpoint.cache.amazonaws.com:6379
USER_SERVICE_URL=http://user-service-internal:8080
USE_AUTO_MIGRATE=false              # 프로덕션은 수동 마이그레이션만
LOG_LEVEL=warn                      # 프로덕션 로그 최소화
```

---

### 환경별 배포 전략

#### Local → EC2 Dev → Production 흐름

```
개발자 로컬 (Local)
    ├─ Feature 개발
    ├─ 단위 테스트
    └─ Git Push
         ↓
EC2 Dev (자동 배포)
    ├─ CI/CD 파이프라인
    ├─ 통합 테스트
    ├─ QA 검증
    └─ 승인 완료
         ↓
Production (수동 배포)
    ├─ Release Tag 생성
    ├─ Production 배포
    └─ 모니터링
```

#### CI/CD 파이프라인 (환경별)

**EC2 Dev 자동 배포**:
```
GitHub Push (main branch)
    ↓
CI Workflow
    ├─ Go test 실행
    ├─ Docker 빌드
    └─ ECR Push (wealist-dev-board-service:latest)
    ↓
CD Workflow (자동 트리거)
    ├─ AWS SSM Parameter Store에서 환경 변수 로드
    ├─ ECR에서 이미지 풀
    ├─ 데이터베이스 마이그레이션 실행
    ├─ Docker Compose로 board-service 재시작
    ├─ Health check (60초 타임아웃)
    └─ Slack 알림
```

**Production 수동 배포**:
```
GitHub Release Tag (v1.0.0)
    ↓
Manual Approval
    ↓
Production Deploy Workflow
    ├─ ECR Push (wealist-prod-board-service:v1.0.0)
    ├─ ECS Task Definition 업데이트
    ├─ Blue/Green Deployment
    ├─ Health check
    ├─ CloudWatch 알람 확인
    └─ Rollback 준비 (이전 버전 유지)
```

## 확장성 및 성능

### 수평 확장 (Horizontal Scaling)

```
┌─────────────────────────────────────────────────┐
│              Load Balancer (ALB)                │
└────┬────────┬────────┬────────┬────────┬────────┘
     │        │        │        │        │
     ▼        ▼        ▼        ▼        ▼
  ┌────┐  ┌────┐  ┌────┐  ┌────┐  ┌────┐
  │BS-1│  │BS-2│  │BS-3│  │BS-4│  │BS-5│ Board Service 인스턴스
  └────┘  └────┘  └────┘  └────┘  └────┘
     │        │        │        │        │
     └────────┴────────┴────────┴────────┘
                      ▼
              ┌──────────────┐
              │  PostgreSQL  │
              │  (RDS)       │
              └──────────────┘

무상태 설계:
✅ 세션 데이터 없음 (JWT 토큰 기반)
✅ 인스턴스 간 데이터 공유 불필요
✅ 자동 스케일링 가능 (CPU 기반)
```

### 캐싱 전략 (Redis)

```go
// 1. 프로젝트 메타데이터 캐싱
Key: "project:{projectId}"
TTL: 1시간
사용: 프로젝트 정보 조회 시 DB 부하 감소

// 2. 사용자 권한 캐싱
Key: "project_member:{projectId}:{userId}"
TTL: 10분
사용: 권한 검증 시 User Service 호출 최소화

// 3. 보드 목록 캐싱
Key: "boards:{projectId}"
TTL: 5분
사용: 보드 목록 조회 성능 향상

캐시 무효화:
- Write-Through: 생성/수정 시 즉시 캐시 갱신
- Cache Aside: 읽기 시 캐시 미스 → DB 조회 → 캐시 저장
```

### 데이터베이스 최적화

```sql
-- 인덱스 전략
CREATE INDEX idx_boards_project_id ON boards(project_id) WHERE is_deleted = false;
CREATE INDEX idx_boards_position ON boards(project_id, position) WHERE is_deleted = false;
CREATE INDEX idx_comments_board_id ON comments(board_id) WHERE deleted_at IS NULL;

-- 파티셔닝 (미래 확장)
-- workspace_id 기반 샤딩 가능
CREATE TABLE boards_partition_ws1 PARTITION OF boards
    FOR VALUES WITH (workspace_id = 'ws-1');

-- 연결 풀링
DB_MAX_OPEN_CONNS=25  # 최대 연결 수
DB_MAX_IDLE_CONNS=5   # 유휴 연결 수
```

### 성능 모니터링

```
메트릭 수집:
┌──────────────────────────────────────────────┐
│  Prometheus Exporter                         │
│  - HTTP 요청 수 (by endpoint)                │
│  - 응답 시간 분포 (P50, P95, P99)            │
│  - 에러율 (by status code)                   │
│  - DB 연결 풀 상태                           │
│  - Goroutine 수                              │
└──────────────┬───────────────────────────────┘
               │
               ▼
      ┌─────────────────┐
      │   Prometheus    │
      │   (Time Series) │
      └────────┬────────┘
               │
               ▼
      ┌─────────────────┐
      │     Grafana     │
      │   (Dashboard)   │
      └─────────────────┘

주요 지표:
- Request Rate: req/sec
- Error Rate: errors/total requests
- Duration: P95 < 200ms
- DB Connections: < 80% of max
```

### 장애 대응

```go
// 1. Circuit Breaker (User Service 호출)
// 연속 실패 시 빠른 실패로 전환
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    state       State // Open/Closed/HalfOpen
}

// 2. Graceful Degradation
// User Service 장애 시에도 핵심 기능은 동작
func (s *boardService) GetBoard(ctx context.Context, boardID string) (*Board, error) {
    board, err := s.repo.FindByID(ctx, boardID)
    if err != nil {
        return nil, err
    }

    // User Service 호출 실패 시에도 보드는 반환
    userProfile, err := s.userClient.GetUserProfile(ctx, board.CreatorID)
    if err != nil {
        logger.Warn("Failed to fetch user profile, using default", zap.Error(err))
        userProfile = &DefaultUserProfile
    }

    return board, nil
}

// 3. Graceful Shutdown
// 진행 중인 요청 완료 후 종료
func (s *Server) Shutdown(ctx context.Context) error {
    // 1. 새 요청 거부
    // 2. 진행 중인 요청 완료 대기 (최대 30초)
    // 3. DB 연결 정리
    // 4. 종료
}
```

## 보안

### 인증 및 인가

```
1. JWT 토큰 검증
   ┌──────────────────────────────────────────┐
   │ Client → ALB → Board Service             │
   │ Authorization: Bearer <JWT>              │
   └──────────────────────────────────────────┘

   Board Service:
   - JWT 서명 검증 (HS512, shared secret)
   - 만료 시간 확인
   - 사용자 ID 추출 (claims)

2. 권한 검증
   func (a *Authorizer) CheckBoardAccess(userID, boardID string) error {
       // 1. 보드 조회
       board := repo.FindByID(boardID)

       // 2. User Service에 워크스페이스 멤버십 확인
       isMember := userClient.ValidateWorkspaceMember(board.WorkspaceID, userID)

       // 3. 권한 캐싱 (10분)
       redis.Set("auth:{userID}:{boardID}", true, 10*time.Minute)

       return nil
   }
```

### 입력 검증

```go
// 1. DTO 레벨 검증 (go-playground/validator)
type CreateBoardRequest struct {
    ProjectID  string `json:"project_id" binding:"required,uuid"`
    Title      string `json:"title" binding:"required,min=1,max=255"`
    Content    string `json:"content" binding:"max=5000"`
    Stage      string `json:"stage" binding:"omitempty,oneof=in_progress pending approved"`
}

// 2. 도메인 레벨 검증
func NewBoard(title, content string) (*Board, error) {
    if strings.TrimSpace(title) == "" {
        return nil, ErrInvalidTitle
    }
    if len(content) > MaxContentLength {
        return nil, ErrContentTooLong
    }
    return &Board{...}, nil
}

// 3. SQL Injection 방지 (GORM Parameterized Queries)
db.Where("project_id = ? AND is_deleted = ?", projectID, false).Find(&boards)
```

### CORS 설정

```go
// 환경별 CORS 정책
config := cors.Config{
    AllowOrigins: []string{
        "http://localhost:3000",           // 로컬 개발
        "https://app.wealist.co.kr",       // 프로덕션
    },
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Authorization", "Content-Type"},
    MaxAge:       12 * time.Hour,
}
```

## 관찰 가능성 (Observability)

### 구조화된 로깅 (Zap)

```go
logger.Info("Board created",
    zap.String("board_id", board.ID),
    zap.String("project_id", board.ProjectID),
    zap.String("user_id", userID),
    zap.Duration("duration", time.Since(start)),
)

// JSON 출력
{
  "level": "info",
  "ts": 1704067200.123,
  "msg": "Board created",
  "board_id": "550e8400-e29b-41d4-a716-446655440000",
  "project_id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "770e8400-e29b-41d4-a716-446655440000",
  "duration": 0.025
}
```

### 헬스 체크

```
GET /health

응답:
{
  "status": "healthy",
  "version": "1.0.0",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy",
    "user_service": "healthy"
  },
  "uptime": "48h30m15s"
}

사용:
- ALB Target Group Health Check
- Kubernetes Liveness/Readiness Probe
- 모니터링 알림
```

## 결론

Board Service는 다음 클라우드 네이티브 원칙을 구현합니다:

✅ **마이크로서비스 아키텍처**: User Service와 완전히 독립적
✅ **무상태 설계**: 수평 확장 가능
✅ **컨테이너화**: Docker 멀티스테이지 빌드
✅ **선언적 API**: RESTful 설계 원칙 준수
✅ **관찰 가능성**: 구조화된 로깅, 메트릭, 헬스 체크
✅ **복원력**: Graceful degradation, Circuit breaker
✅ **자동화**: CI/CD 파이프라인, 자동 마이그레이션

이러한 설계는 시스템의 **확장성, 유지보수성, 안정성**을 보장합니다.
