# weAlist 공통 인프라

weAlist 프로젝트의 공통 인프라 (PostgreSQL, Redis)

## 사용 팀

- Kanban 팀
- Member 팀 (회원 시스템)
- 기타 weAlist 서비스 팀

## 설정

### 1. 환경변수 파일 생성

```bash
cp .env.example .env
```

### 2. `.env` 파일 수정

**중요: 비밀번호를 반드시 변경하세요!**

```bash
POSTGRES_PASSWORD=강력한_비밀번호로_변경
REDIS_PASSWORD=강력한_비밀번호로_변경
```

## 실행

```bash
# 인프라 시작
docker-compose up -d

# 로그 확인
docker-compose logs -f

# 인프라 중지
docker-compose down
```

## 데이터베이스 접속 정보

- **Host**: localhost (또는 wealist-postgres)
- **Port**: 5432
- **Database**: wealist_db
- **User**: wealist_user
- **Password**: `.env` 파일의 `POSTGRES_PASSWORD`

## Redis 접속 정보

- **Host**: localhost (또는 wealist-redis)
- **Port**: 6379
- **Password**: `.env` 파일의 `REDIS_PASSWORD`

## 네트워크

- **네트워크 이름**: `wealist-network`
- 모든 서비스는 이 네트워크를 통해 통신합니다

## 주의사항

⚠️ `.env` 파일은 **절대 Git에 커밋하지 마세요!**
⚠️ 운영 환경에서는 **반드시 강력한 비밀번호**를 사용하세요!
