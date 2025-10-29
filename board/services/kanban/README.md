# Kanban Service

칸반 보드 기반 프로젝트 관리 서비스입니다. 마이크로서비스 아키텍처와 클라우드 네이티브 환경을 고려하여 설계되었습니다.

## 주요 특징

- **마이크로서비스 아키텍처**: Member 서비스와 독립적으로 동작
- **샤딩 대비 설계**: Foreign Key 없이 애플리케이션 레벨에서 데이터 관계 관리
- **JWT 인증**: Member 서비스에서 발급한 JWT 토큰 검증
- **감사 추적**: 모든 엔티티의 생성자/수정자 기록
- **RESTful API**: FastAPI 기반의 표준 REST API
- **자동 문서화**: Swagger UI (OpenAPI)

---

## 기술 스택

- **Language**: Python 3.11+
- **Framework**: FastAPI
- **Database**: PostgreSQL
- **ORM**: SQLAlchemy
- **Authentication**: JWT (python-jose)
- **Container**: Docker, Docker Compose

---

## 데이터 모델

### 계층 구조

```
Workspace (워크스페이스)
  └── Project (프로젝트)
      └── Ticket (티켓)
          └── Task (작업)
```

### 주요 필드

**공통 감사 필드** (모든 엔티티):
- `created_at`: 생성 시각
- `updated_at`: 수정 시각
- `created_by`: 생성자 user_id (Member 서비스 참조)
- `updated_by`: 수정자 user_id (Member 서비스 참조)

**Ticket & Task**:
- `assignee_id`: 담당자 user_id (Member 서비스 참조)

> **Note**: Member 서비스와의 관계는 FK 없이 user_id만 저장합니다. (샤딩 고려)

---

## 시작하기

### 1. 환경 설정

```bash
# .env 파일 생성
cp .env.example .env
```

`.env` 파일 내용:
```env
# Database
DATABASE_URL=postgresql://kanban:kanban123@postgres:5432/kanban_db

# JWT 설정 (Member 서비스와 동일한 SECRET_KEY 사용)
SECRET_KEY=dev-secret-key-change-in-production
ALGORITHM=HS256
```

### 2. Docker Compose로 실행

```bash
# 서비스 시작
docker-compose up -d

# 로그 확인
docker-compose logs -f kanban
```

### 3. API 문서 확인

브라우저에서 http://localhost:8000/docs 접속

---

## JWT 인증

### 개요

Kanban 서비스는 **JWT 검증만** 수행하며, 토큰 발급은 Member 서비스에서 담당합니다.

- **토큰 발급**: Member 서비스 (미구현 시 테스트 스크립트 사용)
- **토큰 검증**: Kanban 서비스
- **공유 SECRET_KEY**: 두 서비스가 동일한 SECRET_KEY 사용 필요

### 테스트용 토큰 생성

Member 서비스가 아직 없을 때 테스트용 JWT 토큰 생성:

```bash
# Docker 컨테이너 내부에서 실행
docker-compose exec kanban python scripts/generate_test_token.py

# 다른 user_id로 생성
docker-compose exec kanban python scripts/generate_test_token.py --user-id 2

# 만료 기간 지정 (기본 7일)
docker-compose exec kanban python scripts/generate_test_token.py --expire-days 30
```

### API 인증 방법

#### Swagger UI (추천)

1. http://localhost:8000/docs 접속
2. 우측 상단 **"Authorize"** 버튼 클릭
3. 토큰 입력 (Bearer 없이)
4. **"Authorize"** 클릭
5. 모든 API에 자동으로 토큰 포함

#### cURL

```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X POST http://localhost:8000/api/workspaces/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "백엔드팀",
    "description": "백엔드 개발팀"
  }'
```

자세한 내용은 [JWT_TEST.md](./JWT_TEST.md) 참조

---

## API 사용 예시

### 1. Workspace 생성

```bash
POST /api/workspaces/
Authorization: Bearer {token}

{
  "name": "백엔드팀",
  "description": "백엔드 개발팀"
}
```

### 2. Project 생성

```bash
POST /api/projects/
Authorization: Bearer {token}

{
  "name": "회원 시스템",
  "workspace_id": 1,
  "status": "ACTIVE",
  "priority": "HIGH"
}
```

### 3. Ticket 생성 (담당자 지정)

```bash
POST /api/tickets/
Authorization: Bearer {token}

{
  "title": "로그인 API 개발",
  "project_id": 1,
  "status": "IN_PROGRESS",
  "priority": "HIGH",
  "assignee_id": 2
}
```

### 4. Task 생성

```bash
POST /api/tasks/
Authorization: Bearer {token}

{
  "title": "JWT 발급 로직",
  "ticket_id": 1,
  "status": "TODO",
  "assignee_id": 2
}
```

---

## 주요 설계 결정

### 1. Foreign Key 제거

**이유**: 샤딩 및 마이크로서비스 확장성 고려

- DB 레벨 FK 대신 애플리케이션 레벨에서 관계 관리
- Member 서비스와의 관계는 user_id만 저장
- CASCADE 삭제도 애플리케이션 레벨에서 처리

**장점**:
- 샤딩 시 데이터 분산 용이
- 마이크로서비스 간 독립성 보장
- DB 분리 시 마이그레이션 간소화

### 2. JWT 인증

**역할 분리**:
- **Member 서비스**: 사용자 관리, 로그인, JWT 발급
- **Kanban 서비스**: JWT 검증, 칸반 보드 관리

**공유 SECRET_KEY**:
- 두 서비스가 같은 SECRET_KEY로 토큰 암호화/검증
- 환경변수로 관리 (프로덕션에서는 Secret Manager 사용 권장)

### 3. 감사 필드 (Audit Trail)

모든 엔티티에 자동 기록:
- `created_by`: JWT에서 추출한 user_id
- `updated_by`: 수정 시 JWT에서 추출한 user_id
- `created_at`, `updated_at`: 자동 타임스탬프

---

## 개발

### 로컬 개발 환경

```bash
# 가상환경 생성
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 의존성 설치
pip install -r requirements.txt

# DB 마이그레이션
alembic upgrade head

# 개발 서버 실행
uvicorn app.main:app --reload
```

### 테스트

```bash
# 테스트 실행
pytest

# 커버리지 확인
pytest --cov=app tests/
```

### DB 마이그레이션

```bash
# 새로운 마이그레이션 생성
alembic revision --autogenerate -m "마이그레이션 메시지"

# 마이그레이션 적용
alembic upgrade head

# 롤백
alembic downgrade -1
```

---

## 프로젝트 구조

```
kanban/
├── app/
│   ├── api/           # API 라우터
│   │   ├── workspaces.py
│   │   ├── projects.py
│   │   ├── tickets.py
│   │   └── tasks.py
│   ├── models/        # SQLAlchemy 모델
│   │   ├── base.py
│   │   ├── workspace.py
│   │   ├── project.py
│   │   ├── ticket.py
│   │   └── task.py
│   ├── schemas/       # Pydantic 스키마
│   ├── services/      # 비즈니스 로직
│   ├── repositories/  # 데이터 액세스
│   ├── auth.py        # JWT 인증
│   ├── config.py      # 설정
│   ├── database.py    # DB 연결
│   └── main.py        # FastAPI 앱
├── alembic/           # DB 마이그레이션
├── scripts/           # 유틸리티 스크립트
│   └── generate_test_token.py
├── tests/             # 테스트
├── docker-compose.yml
├── Dockerfile
├── requirements.txt
├── JWT_TEST.md        # JWT 테스트 가이드
└── README.md
```

---

## 향후 계획

### 클라우드 네이티브 마이그레이션

1. **Kubernetes 배포**
   - Deployment, Service, ConfigMap, Secret 구성
   - Ingress 설정

2. **CI/CD**
   - GitHub Actions / GitLab CI
   - 자동 빌드 및 배포

3. **모니터링**
   - Prometheus + Grafana
   - ELK Stack (로그 관리)

4. **Helm 차트**
   - 재사용 가능한 배포 패키지

5. **Auto Scaling**
   - HPA (Horizontal Pod Autoscaler)
   - VPA (Vertical Pod Autoscaler)

6. **보안**
   - Network Policy
   - RBAC
   - Secret 암호화
   - Service Mesh (Istio)

---

## 라이선스

MIT License

---

## 문의

프로젝트 관련 문의사항은 Issue를 통해 남겨주세요.
