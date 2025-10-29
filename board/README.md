# weAlist Kanban Service

FastAPI 기반 칸반 보드 관리 서비스

## 주요 기능
test context - githubaction

칸반 보드 계층 구조:
```
Workspace (워크스페이스)
  └── Project (프로젝트)
       ├── Ticket (티켓)
       │    └── Task (작업)
       ├── Ticket Type (티켓 타입)
       └── Project Member (프로젝트 멤버)

Notification (알림) - 사용자별
```

- ✅ **Workspace**: 팀/조직 단위 관리
- ✅ **Project**: 프로젝트별 상태/우선순위 관리
- ✅ **Ticket**: 이슈/기능 단위 추적
- ✅ **Task**: 세부 작업 관리 및 완료 처리
- ✅ **Notification**: 사용자별 알림 시스템
- ✅ **Ticket Type**: 프로젝트별 커스텀 타입

## 기술 스택

- **Python 3.11** + **FastAPI**
- **SQLAlchemy 2.0** (Async ORM)
- **Alembic** (DB Migration)
- **Pydantic v2** (데이터 검증)
- **PostgreSQL 16**
- **Redis 7**
- **Docker & Docker Compose**

## 빠른 시작

### 전제 조건
- Docker & Docker Compose

### 실행 방법

```bash
# 루트 디렉토리에서 전체 환경 시작
docker-compose up -d

# Kanban Service 로그 확인
docker logs -f wealist-kanban-service

# 접속 확인
curl http://localhost:8000/health
```

## API 문서

- **Swagger UI**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc
- **상세 API 문서**: [services/kanban/API_DOCUMENTATION.md](services/kanban/API_DOCUMENTATION.md)
- **테스트 가이드**: [services/kanban/API_TEST_GUIDE.md](services/kanban/API_TEST_GUIDE.md)

### 헬스체크
- **Health Check**: `GET /health`
- **Liveness**: `GET /health/live`
- **Readiness**: `GET /health/ready`

### 주요 엔드포인트

| 엔드포인트 | 설명 | 인증 |
|-----------|------|------|
| `GET /api/workspaces/` | 워크스페이스 목록 | Required |
| `POST /api/workspaces/` | 워크스페이스 생성 | Required |
| `GET /api/projects/` | 프로젝트 목록 | Required |
| `GET /api/tickets/` | 티켓 목록 | Required |
| `GET /api/tasks/` | 작업 목록 | Required |
| `GET /api/notifications/` | 알림 목록 | Required |
| `POST /api/projects/{id}/ticket-types/` | 티켓 타입 생성 | Required |

**주의**: 모든 엔드포인트 URL은 **마지막 슬래시(/) 필수**

## 인증

JWT 토큰이 필요합니다. 테스트 토큰 생성:

```bash
# 기본 토큰 생성
docker exec wealist-kanban-service python scripts/generate_test_token.py

# 특정 사용자 ID로 생성
docker exec wealist-kanban-service python scripts/generate_test_token.py --user-id {UUID}

# 만료 기간 설정
docker exec wealist-kanban-service python scripts/generate_test_token.py --expire-days 30
```

## 로컬 개발

### 가상환경 설정
```bash
cd services/kanban

# 가상환경 생성 및 활성화
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 의존성 설치
pip install -r requirements.txt

# 개발 서버 실행
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

### DB 마이그레이션
```bash
# 마이그레이션 생성
docker exec wealist-kanban-service alembic revision --autogenerate -m "설명"

# 마이그레이션 적용
docker exec wealist-kanban-service alembic upgrade head

# 현재 버전 확인
docker exec wealist-kanban-service alembic current

# 롤백
docker exec wealist-kanban-service alembic downgrade -1
```

### 테스트 실행
```bash
# 전체 테스트 + 커버리지
docker exec wealist-kanban-service pytest --cov=app tests/

# 상세 출력
docker exec wealist-kanban-service pytest -v

# 특정 파일만
docker exec wealist-kanban-service pytest tests/test_workspaces.py
```

**테스트 통계**: 총 36개 테스트, 커버리지 97%+

## 프로젝트 구조

```
services/kanban/
├── app/
│   ├── api/                 # API 라우터
│   │   ├── workspaces.py
│   │   ├── projects.py
│   │   ├── tickets.py
│   │   ├── tasks.py
│   │   ├── notifications.py
│   │   └── ticket_types.py
│   ├── models/              # SQLAlchemy 모델 (FK 없음)
│   ├── schemas/             # Pydantic 스키마
│   ├── auth.py              # JWT 검증
│   ├── database.py          # DB 세션
│   └── main.py              # FastAPI 앱
├── alembic/                 # DB 마이그레이션
├── scripts/                 # 유틸리티 스크립트
│   └── generate_test_token.py
├── tests/                   # 테스트 코드
├── API_DOCUMENTATION.md     # 상세 API 문서
├── API_TEST_GUIDE.md        # 테스트 가이드
└── requirements.txt
```

## 환경 변수

주요 환경 변수는 루트의 `.env` 파일에서 관리됩니다:

```env
KANBAN_SERVICE_PORT=8000
KANBAN_DB_HOST=kanban-db
KANBAN_REDIS_HOST=kanban-redis
JWT_SECRET=your-secret-key  # User Service와 동일
```

## 아키텍처 특징

### 확장성 대비 설계
- ❌ **DB 레벨 Foreign Key 사용 안함**
- ✅ **애플리케이션 레벨 CASCADE 처리**
- ✅ **UUID Primary Key 사용**
- ✅ **Soft Delete 지원** (Ticket Type)


## 트러블슈팅

### 포트 충돌
```bash
# 포트 사용 확인
lsof -i :8000

# 포트 변경 (.env 파일)
KANBAN_SERVICE_PORT=8001
```

### 데이터베이스 연결 실패
```bash
# PostgreSQL 상태 확인
docker logs wealist-kanban-db

# 연결 테스트
docker exec -it wealist-kanban-db psql -U wealist_kanban_user -d wealist_kanban_db
```

### 307 Redirect 에러
URL 끝에 슬래시(/) 추가 필요:
- ❌ `/api/workspaces` → ✅ `/api/workspaces/`

## 관련 문서

- **API 상세 문서**: [services/kanban/API_DOCUMENTATION.md](services/kanban/API_DOCUMENTATION.md)
- **테스트 가이드**: [services/kanban/API_TEST_GUIDE.md](services/kanban/API_TEST_GUIDE.md)
- **Infrastructure 가이드**: [infrastructure/README.md](infrastructure/README.md)

## 라이선스

이 프로젝트는 학습 목적으로 개발되었습니다.
