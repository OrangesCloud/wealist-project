# weAlist 아키텍처 문서

이 문서는 weAlist 프로젝트의 전체 아키텍처와 설계 원칙을 상세히 설명합니다.

## 목차

- [개요](#개요)
- [시스템 아키텍처](#시스템-아키텍처)
- [마이크로서비스 설계](#마이크로서비스-설계)
- [데이터베이스 설계](#데이터베이스-설계)
- [인증 및 인가](#인증-및-인가)
- [배포 아키텍처](#배포-아키텍처)
- [모니터링 및 로깅](#모니터링-및-로깅)
- [확장성 전략](#확장성-전략)

## 개요

weAlist는 **마이크로서비스 아키텍처**를 기반으로 한 프로젝트 관리 플랫폼입니다. 각 서비스는 독립적으로 배포 가능하며, 확장성과 유지보수성을 고려하여 설계되었습니다.

### 핵심 설계 목표

- **독립성**: 서비스 간 느슨한 결합, 독립적인 배포 및 확장
- **확장성**: 수평 확장 가능한 무상태(Stateless) 서비스
- **복원력**: 장애 격리 및 Graceful Degradation
- **관찰성**: 구조화된 로깅, 메트릭, 추적

## 시스템 아키텍처

### 전체 구조도

```
                          ┌─────────────────────┐
                          │   Client / User     │
                          └──────────┬──────────┘
                                     │
                          ┌──────────▼──────────┐
                          │  ALB (Production)   │
                          │ /api/users/*        │
                          │ /api/boards/*       │
                          └──────────┬──────────┘
                                     │
                 ┌───────────────────┴───────────────────┐
                 │                                       │
        ┌────────▼────────┐                   ┌─────────▼────────┐
        │  User Service   │                   │  Board Service   │
        │  (Spring Boot)  │◄──────JWT─────────│     (Go/Gin)     │
        │   Port 8080     │    Validation     │   Port 8000      │
        └────────┬────────┘                   └─────────┬────────┘
                 │                                       │
                 │                                       │
        ┌────────▼────────┐                   ┌─────────▼────────┐
        │  wealist_user_db│                   │ wealist_board_db │
        │   (PostgreSQL)  │                   │   (PostgreSQL)   │
        └─────────────────┘                   └──────────────────┘
                 │                                       │
                 └───────────────┬───────────────────────┘
                                 │
                          ┌──────▼──────┐
                          │    Redis    │
                          │   (Cache)   │
                          └─────────────┘
```

### 서비스 분리 원칙

| 서비스 | 책임 | 데이터 |
|--------|------|--------|
| **User Service** | - 사용자 인증/인가<br>- 워크스페이스 관리<br>- OAuth2 통합<br>- 멤버십 관리 | - 사용자 정보<br>- 워크스페이스<br>- 멤버십 |
| **Board Service** | - 프로젝트 관리<br>- 보드 CRUD<br>- 커스텀 필드<br>- 댓글/첨부파일 | - 프로젝트<br>- 보드<br>- 커스텀 필드<br>- 댓글 |

## 마이크로서비스 설계

### 1. User Service (Spring Boot)

**아키텍처 패턴**: Layered Architecture

```
Controller Layer
    ↓ (DTO)
Service Layer
    ↓ (Entity)
Repository Layer
    ↓
Database (PostgreSQL)
```

**핵심 기술**:
- Spring Boot 3.x (Java 21)
- Spring Security + JWT
- Spring Data JPA
- OAuth2 Client (Google)

**주요 엔티티**:
- `User`: 사용자 인증 정보
- `UserProfile`: 사용자 프로필 (분리된 테이블)
- `Workspace`: 워크스페이스
- `WorkspaceMember`: 워크스페이스 멤버십

**코딩 컨벤션**:
- **camelCase** 필수 (underscore_case 금지)
- 엔티티 prefix 필수: `id` → `workspaceId`, `name` → `workspaceName`
- UserProfile은 반드시 `nickName` 사용 (`name` 금지)

**참고**: [_README/CODING_CONVENTIONS.md](./_README/CODING_CONVENTIONS.md)

### 2. Board Service (Go)

**아키텍처 패턴**: Clean Architecture + DDD

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │
│  (Handler, Middleware, Router)          │
└──────────────┬──────────────────────────┘
               │ DTO
┌──────────────▼──────────────────────────┐
│         Business Logic Layer            │
│  (Service, Domain Models)               │
│  - Rich Domain Models (26 methods)     │
│  - Business Rules                       │
└──────────────┬──────────────────────────┘
               │ Domain Entity
┌──────────────▼──────────────────────────┐
│         Data Access Layer               │
│  (Repository, Unit of Work)             │
│  - Generic Repository Pattern           │
│  - Transaction Management               │
└──────────────┬──────────────────────────┘
               │
        ┌──────▼──────┐
        │  PostgreSQL │
        │   + Redis   │
        └─────────────┘
```

**핵심 기술**:
- Go 1.21+
- Gin (HTTP Framework)
- GORM (ORM)
- Zap (Structured Logging)
- Viper (Configuration)

**주요 패턴**:
- **Rich Domain Models**: 비즈니스 로직이 도메인 엔티티에 포함
- **Generic Repository**: Go 제네릭을 활용한 베이스 레포지토리
- **Unit of Work**: 트랜잭션 관리 패턴
- **Centralized Authorization**: `ProjectAuthorizer`를 통한 일관된 권한 검사

**참고**: [board-service/ARCHITECTURE.md](./board-service/ARCHITECTURE.md)

### 서비스 간 통신

**통신 방식**: RESTful HTTP API (동기 통신)

**Board Service → User Service 호출**:
- `GET /api/workspaces/:workspace_id/members/:user_id` - 멤버십 검증
- `GET /api/users/:user_id` - 사용자 정보 조회

**인증 방식**: JWT 토큰 (공유 SECRET_KEY)

## 데이터베이스 설계

### 핵심 원칙

1. **서비스별 독립 데이터베이스**
   - User Service: `wealist_user_db`
   - Board Service: `wealist_board_db`
   - 서비스 간 직접 DB 접근 금지

2. **No Foreign Keys (Cross-Service)**
   - 서비스 간 외래 키 제약 조건 없음
   - 애플리케이션 레벨에서 관계 관리
   - 샤딩 대비 설계

3. **UUID Primary Keys**
   - 모든 엔티티는 UUID 타입 PK 사용
   - 글로벌 고유성 보장
   - 분산 환경에서 충돌 방지

4. **Soft Delete**
   - `is_deleted` 플래그 사용 (User, Workspace, Project, Board)
   - `deleted_at` 타임스탬프 기록
   - Comment는 GORM의 `DeletedAt` 사용

### 데이터 일관성 전략

**Eventually Consistent**:
- 서비스 간 데이터 일관성은 최종 일관성(Eventually Consistent) 허용
- 예: Workspace 삭제 시 Board Service의 프로젝트는 별도 처리

**API 호출로 검증**:
- Board Service는 User Service API로 워크스페이스 멤버십 검증
- 실시간 권한 체크

### 트랜잭션 범위

- **User Service**: JPA 트랜잭션 (`@Transactional`)
- **Board Service**: Unit of Work 패턴으로 트랜잭션 관리
- **Cross-Service**: 분산 트랜잭션 없음 (Saga 패턴 고려 중)

## 인증 및 인가

### JWT 기반 인증

**알고리즘**: HS512 (HMAC with SHA-512)

**토큰 타입**:
- Access Token: 30분 (1800000ms)
- Refresh Token: 7일 (604800000ms)

**공유 Secret Key**:
```bash
# User Service
JWT_SECRET=your_secret_key_at_least_64_bytes

# Board Service
SECRET_KEY=your_secret_key_at_least_64_bytes  # 동일한 값

# 환경 변수 매핑 (docker-compose.yml)
- JWT_SECRET → User Service
- SECRET_KEY → Board Service (동일한 값)
```

### 인증 흐름

```
1. 클라이언트 로그인
   ↓
2. User Service: JWT 토큰 발급 (Access + Refresh)
   ↓
3. 클라이언트 → Board Service: Authorization: Bearer {token}
   ↓
4. Board Service: JWT 검증 (공유 SECRET_KEY 사용)
   ↓
5. Board Service → User Service: 워크스페이스 멤버십 검증 API 호출
   ↓
6. 권한 확인 후 요청 처리
```

### 역할 기반 접근 제어 (RBAC)

**역할 레벨**:
- `OWNER` (100): 모든 권한 + 워크스페이스 삭제
- `ADMIN` (50): 프로젝트/멤버 관리
- `MEMBER` (10): 읽기 + 보드 편집

**권한 체크**:
- Board Service: `ProjectAuthorizer`에서 중앙 집중식 권한 검사
- User Service: Spring Security + 커스텀 권한 체크

## 배포 아키텍처

### 환경별 구성

| 환경 | 배포 방식 | 인프라 |
|------|----------|--------|
| **Local Dev** | Docker Compose (dev.yml) | 로컬 머신 |
| **EC2 Dev** | Docker Compose (ec2-dev.yml) | EC2 단일 인스턴스 (t3.small) |
| **Production** | Docker Compose (prod.yml) + CodeDeploy | EC2 (Auto Scaling 고려) |

### CI/CD Pipeline

**Development**:
```
GitHub Push → CI (ECR) → CD (SSH) → EC2 Dev
```

**Production**:
```
GitHub Push → CI (ECR) → CD (CodeDeploy) → EC2 Prod
```

**워크플로우**:
- CI: `.github/workflows/ci-{env}-{service}.yml`
- CD: `.github/workflows/cd-{env}-{service}.yml`

**참고**: [.github/workflows/README.md](./.github/workflows/README.md)

### AWS 리소스

**현재 사용**:
- **EC2**: 애플리케이션 서버
- **ECR**: Docker 이미지 레지스트리
- **ALB**: 경로 기반 라우팅 (Production)
- **CodeDeploy**: 프로덕션 배포 자동화
- **SSM Parameter Store**: 환경 변수 관리

**향후 고려**:
- **RDS**: 관리형 PostgreSQL (현재 EC2 내 컨테이너)
- **ElastiCache**: 관리형 Redis (현재 EC2 내 컨테이너)
- **S3**: 파일 스토리지
- **CloudWatch**: 로그 및 메트릭 통합

### ALB 라우팅 (Production)

```
Client → ALB (https://api.wealist.co.kr)
    ├─ /api/users/*  → User Service (EC2:8080)
    └─ /api/boards/* → Board Service (EC2:8000)
```

**Path Stripping**:
- User Service: ALB가 `/api/users` prefix 제거 후 전달
- Board Service: ALB가 `/api/boards` prefix 제거 후 전달

**Service Configuration**:
- User Service: `server.servlet.context-path=/api/users` (AWS profile)
- Board Service: `SERVER_BASE_PATH=/api/boards` (ENV=prod)

**참고**: [.kiro/docs/DEPLOYMENT_GUIDES.md](./.kiro/docs/DEPLOYMENT_GUIDES.md)

## 모니터링 및 로깅

### 모니터링 스택

**Metrics**:
- **Prometheus**: 메트릭 수집 및 저장
- **Grafana**: 시각화 대시보드
- **Node Exporter**: 시스템 메트릭 (CPU, Memory, Disk)
- **Cadvisor**: 컨테이너 메트릭
- **RDS Exporter**: PostgreSQL 메트릭 (Production)

**Logging**:
- **Loki**: 로그 집계
- **Promtail**: 로그 수집 에이전트
- **Structured Logging**: JSON 형식 로그 (Zap, Logback)

### 주요 메트릭

**Application Metrics**:
- HTTP 응답 시간 (p50, p95, p99)
- 에러율 (4xx, 5xx)
- 요청 처리량 (req/sec)

**Infrastructure Metrics**:
- CPU/Memory 사용률 (컨테이너별)
- Disk I/O
- Network Traffic

**Database Metrics**:
- 커넥션 풀 사용률
- 쿼리 실행 시간
- Active 커넥션 수

### 배포

```bash
# 모니터링 스택 시작
./docker/scripts/monitoring.sh up

# 접속
# - Grafana: http://localhost:3001
# - Prometheus: http://localhost:9090
```

**참고**: [CLAUDE.md - Monitoring & Observability](./CLAUDE.md#monitoring--observability)

## 확장성 전략

### 수평 확장 (Horizontal Scaling)

**무상태 서비스**:
- 모든 서비스는 상태를 유지하지 않음
- 세션 데이터는 Redis에 저장
- 다중 인스턴스 배포 가능

**로드 밸런싱**:
- ALB를 통한 트래픽 분산
- Health Check 기반 인스턴스 관리

### 데이터베이스 확장

**수직 확장 (Vertical Scaling)**:
- 현재 전략: EC2/RDS 인스턴스 크기 증가

**수평 확장 준비**:
- No Foreign Keys: 샤딩 준비 완료
- UUID PK: 분산 환경에서 충돌 없음
- 서비스별 독립 DB: 마이크로서비스 확장 가능

**향후 고려**:
- Read Replica (읽기 부하 분산)
- Sharding (워크스페이스 단위)
- CQRS 패턴 (Command/Query 분리)

### 캐싱 전략

**Redis 캐싱**:
- User Service: 세션, 토큰
- Board Service: 프로젝트 정보, 멤버십 정보

**Cache Invalidation**:
- TTL 기반 만료
- 이벤트 기반 무효화 (업데이트/삭제 시)

**참고**: [board-service/CACHE_STRATEGY.md](./board-service/CACHE_STRATEGY.md)

### 성능 최적화

**Database**:
- 인덱스 최적화 (복합 인덱스 활용)
- N+1 쿼리 방지 (Eager Loading)
- Connection Pool 튜닝

**Application**:
- Fractional Indexing: O(1) 보드 위치 업데이트
- Lazy Loading: 필요한 데이터만 로드
- Pagination: 대용량 데이터 처리

## 설계 결정 (ADR - Architecture Decision Records)

### ADR-001: 마이크로서비스 아키텍처 선택

**결정**: User Service와 Board Service를 분리

**이유**:
- 독립적인 배포 및 확장
- 기술 스택 다양성 (Spring Boot vs Go)
- 팀 분할 가능
- 장애 격리

**Trade-off**:
- 복잡성 증가
- 분산 트랜잭션 어려움
- 네트워크 오버헤드

### ADR-002: No Foreign Keys (Cross-Service)

**결정**: 서비스 간 외래 키 제약 조건 없음

**이유**:
- 샤딩 대비 (향후 수평 확장)
- 서비스 독립성 보장
- 스키마 변경 유연성

**Trade-off**:
- 참조 무결성은 애플리케이션에서 관리
- Eventually Consistent 허용

### ADR-003: JWT 기반 인증

**결정**: JWT (HS512) 사용

**이유**:
- 무상태 인증
- 서비스 간 토큰 공유 가능
- 확장성 우수

**Trade-off**:
- 토큰 크기가 큼
- 즉시 무효화 어려움 (Refresh Token으로 보완)

### ADR-004: Fractional Indexing

**결정**: 보드 정렬에 Fractional Indexing 사용

**이유**:
- O(1) 위치 업데이트 (캐스케이드 업데이트 없음)
- 빈번한 정렬 변경에 효율적
- 충돌 없는 동시 편집 가능

**Trade-off**:
- 구현 복잡도 증가
- 프론트엔드가 문자열 위치 처리 필요

## 향후 개선 사항

### 단기 (1-3개월)

- [ ] RDS/ElastiCache로 마이그레이션 (관리형 서비스)
- [ ] Auto Scaling 구성 (CPU/Memory 기반)
- [ ] Blue-Green 배포 구현
- [ ] 분산 추적 (OpenTelemetry)

### 중기 (3-6개월)

- [ ] Event-Driven Architecture (이벤트 기반 통신)
- [ ] CQRS 패턴 적용 (읽기/쓰기 분리)
- [ ] API Gateway 도입 (통합 엔드포인트)
- [ ] Rate Limiting (요청 제한)

### 장기 (6개월+)

- [ ] Kubernetes 마이그레이션 (컨테이너 오케스트레이션)
- [ ] Database Sharding (워크스페이스 단위)
- [ ] Multi-Region 배포 (글로벌 서비스)
- [ ] GraphQL API (유연한 쿼리)

## 참고 문서

- **전체 개발 가이드**: [CLAUDE.md](./CLAUDE.md)
- **User Service 컨벤션**: [_README/CODING_CONVENTIONS.md](./_README/CODING_CONVENTIONS.md)
- **Board Service 아키텍처**: [board-service/ARCHITECTURE.md](./board-service/ARCHITECTURE.md)
- **Board Service 캐싱**: [board-service/CACHE_STRATEGY.md](./board-service/CACHE_STRATEGY.md)
- **CI/CD 워크플로우**: [.github/workflows/README.md](./.github/workflows/README.md)
- **배포 가이드**: [.kiro/docs/DEPLOYMENT_GUIDES.md](./.kiro/docs/DEPLOYMENT_GUIDES.md)
