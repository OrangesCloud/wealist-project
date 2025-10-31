# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**weAlist** is a Jira-style project management platform built with a microservices architecture in a **monorepo**. The system consists of three main services:

1. **User Service** (`user/`) - Spring Boot/Java - User authentication and management
2. **Kanban Service** (`kanban-service/`) - FastAPI/Python - Project and task management
3. **Frontend** (`frontend/`) - React/TypeScript - Frontend application

All services are managed in a single Git repository, orchestrated by a unified `docker-compose.yaml`, and share JWT authentication with a common `JWT_SECRET` across a Docker bridge network (`wealist-net`).

## Architecture

### Service Communication
- **JWT Token Flow**: User Service issues tokens → Kanban Service validates tokens
- **No Direct Service-to-Service Calls**: Services only validate JWT, no HTTP calls between backends
- **Shared Database Strategy**: Each service has its own PostgreSQL database (no shared DB)
- **User References**: Kanban service stores `user_id` (UUID) but doesn't have FK to User service (microservice isolation)

### Monorepo Structure
- **Single Git Repository**: All services (User, Kanban, Frontend) are managed in one repository
- **Unified Orchestration**: Single `docker-compose.yaml` at root level manages all services
- **Shared Configuration**: Common `.env` file for all service configurations
- **Independent Services**: Despite being in monorepo, services remain loosely coupled

### Database Design Philosophy
- **No Foreign Keys**: Application-level CASCADE for sharding readiness
- **UUID Primary Keys**: All tables use `gen_random_uuid()` for distributed system compatibility
- **Soft Deletes**: `is_deleted` boolean flag instead of hard deletes
- **Audit Fields**: Every entity tracks `created_by`, `updated_by`, `created_at`, `updated_at`

### Kanban Service Entity Hierarchy
```
Workspace
  └── Project
      ├── Ticket (has ticket_type_id, assignee_id)
      │   ├── Task (has assignee_id)
      │   └── Comment
      ├── ProjectRole
      ├── ProjectMember (links user to role)
      ├── TicketType (project-specific categories)
      └── Attachment

Notification (global, per user)
```

## Development Commands

### All Services (Docker Compose)

**Start all services:**
```bash
docker-compose up -d
```

**View logs:**
```bash
docker-compose logs -f kanban-service
docker-compose logs -f user-service
docker-compose logs -f frontend
```

**Restart a service:**
```bash
docker restart wealist-kanban-service
docker restart wealist-user-service
```

**Stop all services:**
```bash
docker-compose down
```

### Kanban Service (FastAPI/Python)

**Database migrations:**
```bash
# Create new migration
docker exec wealist-kanban-service alembic revision --autogenerate -m "Description"

# Apply migrations
docker exec wealist-kanban-service alembic upgrade head

# Check current version
docker exec wealist-kanban-service alembic current

# Rollback one version
docker exec wealist-kanban-service alembic downgrade -1
```

**Generate test JWT token:**
```bash
docker exec wealist-kanban-service python scripts/generate_test_token.py
docker exec wealist-kanban-service python scripts/generate_test_token.py --user-id UUID --expire-days 30
```

**Run tests:**
```bash
cd kanban-service
pytest
pytest --cov=app tests/
```

**Access API docs:**
- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc

**Database access:**
```bash
docker exec -it wealist-postgres psql -U kanban_service -d wealist_kanban_db
```

### User Service (Spring Boot/Java)

**Build:**
```bash
cd user
./gradlew build
```

**Run tests:**
```bash
cd user
./gradlew test
```

**Database access:**
```bash
docker exec -it wealist-postgres psql -U user_service -d wealist_user_db
```

### Frontend (React/TypeScript)

**Start development server:**
```bash
cd frontend
npm start
# Note: BROWSER=None prevents auto-opening browser
```

**Build for production:**
```bash
cd frontend
npm run build
```

**Run tests:**
```bash
cd frontend
npm test
```

## Critical Conventions

### UUID Usage
- **Always use UUID type** for all ID fields (database, schemas, API parameters)
- Import: `from uuid import UUID` (Python)
- Database: `Column(PG_UUID, server_default=sa.text('gen_random_uuid()'))`
- Pydantic: `id: UUID`, `user_id: UUID`
- FastAPI: `user_id: UUID = Path(...)`

### Migration Workflow (Kanban Service)
1. Modify models in `kanban-service/app/models/`
2. Generate migration: `docker exec wealist-kanban-service alembic revision --autogenerate -m "message"`
3. Review generated file in `kanban-service/alembic/versions/`
4. Test locally: `alembic upgrade head` → `alembic downgrade -1` → `alembic upgrade head`
5. Commit migration file with code changes

### Git Branch Strategy
**Monorepo**: All services are in a single repository.

#### Branch Workflow
```bash
git branch -a                    # Check all branches
git checkout -b feature/...      # Create new feature branch
# Work on any service (user/, kanban-service/, frontend/)
git add .
git commit -m "feat: ..."
git push origin feature/...
```

#### Branch Rules
- Check existing branches before starting work: `git branch -a`
- Review `.claude/branches.md` to avoid duplicate work
- Feature branches: `feature/descriptive-name`
- Service-specific branches: `feature/user-something`, `feature/kanban-something`, `feature/frontend-something`
- Always check if similar work exists: `git log feature/branch-name --oneline`
- After merge: delete local and remote branches

### Authentication Testing
1. Generate token: `docker exec wealist-kanban-service python scripts/generate_test_token.py`
2. Use Swagger UI at http://localhost:8000/docs
3. Click "Authorize" → Paste token (without "Bearer") → Authorize
4. All subsequent API calls include token automatically

## Important Files

- **`.claude/token-optimization.md`** - 🔥 **토큰 절약 가이드** - Claude Code 사용 시 효율적인 토큰 사용 전략 (반드시 참조!)
- **`.claude/context.md`** - Current project state, completed features, database schema
- **`.claude/branches.md`** - Branch status, cleanup targets, git workflows
- **`.claude/api-user-documentation.md`** - User Service API 문서 (최신화 필요)
- **`.claude/api-kanban-documentation.md`** - Kanban Service API 문서 (최신화 필요)
- **`.claude/api-kanban-test-guide.md`** - Kanban Service 테스트 가이드 (최신화 필요)
- **`docker-compose.yaml`** - Service orchestration (root level)
- **`.env`** - Environment variables for all services (root level, git-ignored)
- **`README.md`** - 프로젝트 실행 가이드 (root level)

## Project Structure

```
wealist/                                  # Monorepo root (Single Git repository)
├── user/                                 # User Service (Spring Boot/Java)
│   ├── src/main/java/OrangeCloud/UserRepo/
│   │   ├── controller/                  # REST API controllers
│   │   ├── service/                     # Business logic
│   │   ├── repository/                  # JPA repositories
│   │   ├── entity/                      # JPA entities
│   │   ├── dto/                         # Data transfer objects
│   │   ├── config/                      # Configuration classes
│   │   └── util/                        # Utilities (JWT)
│   ├── build.gradle                     # Gradle build config
│   └── Dockerfile
├── kanban-service/                       # Kanban Service (FastAPI/Python)
│   ├── app/
│   │   ├── api/                         # Route handlers (workspaces, projects, tickets, tasks)
│   │   ├── models/                      # SQLAlchemy models (NO foreign keys)
│   │   ├── schemas/                     # Pydantic schemas for validation
│   │   ├── auth.py                      # JWT validation
│   │   ├── database.py                  # DB session management
│   │   └── main.py                      # FastAPI app entry point
│   ├── alembic/                         # Database migrations
│   ├── scripts/                         # Utilities (generate_test_token.py)
│   ├── tests/                           # Pytest tests
│   ├── requirements.txt
│   └── Dockerfile
├── frontend/                             # Frontend (React/TypeScript)
│   ├── src/
│   │   ├── components/                  # React components
│   │   ├── pages/                       # Page components
│   │   ├── services/                    # API services
│   │   ├── contexts/                    # React contexts
│   │   ├── types/                       # TypeScript types
│   │   └── App.tsx
│   ├── package.json
│   └── tailwind.config.js
├── docker-compose.yaml                   # Orchestrates all services
├── .env                                  # Shared environment variables (git-ignored)
├── .env.example                          # Environment template
├── init-db.sh                            # PostgreSQL init script
├── README.md                             # Setup and execution guide
├── CLAUDE.md                             # This file - Claude Code guide
└── .claude/                              # Claude Code context files (git-ignored)
    ├── context.md                        # Project status
    ├── branches.md                       # Branch management
    ├── api-user-documentation.md         # User Service API docs
    ├── api-kanban-documentation.md       # Kanban Service API docs
    └── api-kanban-test-guide.md          # Kanban Service testing guide
```

## Kanban Service API Structure

All endpoints require JWT authentication via `Authorization: Bearer <token>` header.

**상세 문서**: `.claude/api-kanban-documentation.md` (최신화 필요)
**테스트 가이드**: `.claude/api-kanban-test-guide.md` (최신화 필요)

**Workspaces:** `/api/workspaces/`
- POST, GET (list), GET /{id}, PATCH /{id}, DELETE /{id}

**Projects:** `/api/projects/`
- POST, GET (list with filters), GET /{id}, PATCH /{id}, DELETE /{id}
- Filters: workspace_id, status, priority

**Tickets:** `/api/tickets/`
- POST, GET (list with filters), GET /{id}, PATCH /{id}, DELETE /{id}
- Filters: project_id, status, priority

**Tasks:** `/api/tasks/`
- POST, GET (list with filters), GET /{id}, PATCH /{id}, PATCH /{id}/complete, DELETE /{id}
- Filters: ticket_id, status

**Ticket Types:** `/api/projects/{project_id}/ticket-types/`
- POST, GET (list), GET /{type_id}, PATCH /{type_id}, DELETE /{type_id} (soft delete)

**Notifications:** `/api/notifications/`
- POST, GET (list), GET /unread-count, GET /{id}, PATCH /{id}/read, POST /mark-all-read, DELETE /{id}
- Filters: is_read, notification_type

## Common Pitfalls

1. **Don't use `int` for IDs** - Everything is UUID now
2. **Don't add Foreign Keys** - Use comments like `comment='References users.id (no FK for sharding)'`
3. **Don't skip migration testing** - Always test upgrade/downgrade before committing
4. **Check for existing branches** - Read `.claude/branches.md` before creating new feature branches
5. **Update context files** - When making significant changes, update `.claude/context.md`
6. **JWT_SECRET must match** - User service and Kanban service must share the same secret
7. **Don't query User service from Kanban** - Only store user_id, no cross-service DB calls
8. **Monorepo awareness** - All services are in one repo, so coordinate changes carefully

## Service Ports

- **User Service**: http://localhost:8080
- **Kanban Service**: http://localhost:8000
- **Frontend**: http://localhost:3000 (dev server, not in docker-compose)
- **PostgreSQL**: localhost:5432 (shared instance with multiple databases)
  - `wealist_user_db` (User Service)
  - `wealist_kanban_db` (Kanban Service)
- **Redis**: localhost:6379 (shared instance with database separation)
  - DB 0: User Service
  - DB 1: Kanban Service

## Testing Strategy

### Kanban Service
- Use `pytest` with fixtures for DB session and test client
- Generate test tokens with `generate_test_token.py --user-id <uuid>`
- Tests use in-memory SQLite or test PostgreSQL database
- Mock User Service responses (no actual HTTP calls)
- Run tests: `cd kanban-service && pytest --cov=app tests/`

### User Service
- Use JUnit/Mockito for unit tests
- Use `@SpringBootTest` for integration tests
- Test JWT token generation and validation separately
- Run tests: `cd user && ./gradlew test`

### Frontend
- Use React Testing Library and Jest
- Run tests: `cd frontend && npm test`

## Project Status Tracking

Before starting new work:
1. Read `.claude/context.md` for current project state
2. Check `.claude/branches.md` for existing/completed branches
3. Run `git log main..feature/branch-name --oneline` to check if feature already exists

After completing work:
1. Update `.claude/context.md` with new features
2. Mark completed branches in `.claude/branches.md`
3. Clean up merged branches
