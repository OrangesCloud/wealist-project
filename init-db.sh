#!/bin/bash
set -e

# PostgreSQL 초기화 스크립트
# 두 개의 독립된 데이터베이스와 사용자를 생성합니다

echo "🚀 weAlist 데이터베이스 초기화 시작..."

# User Service Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${USER_DB_NAME};
    CREATE USER ${USER_DB_USER} WITH PASSWORD '${USER_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${USER_DB_NAME} TO ${USER_DB_USER};
    \c ${USER_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${USER_DB_USER};
EOSQL

echo "✅ User 서비스 데이터베이스 생성 완료: ${USER_DB_NAME}"

# Kanban Service Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${KANBAN_DB_NAME};
    CREATE USER ${KANBAN_DB_USER} WITH PASSWORD '${KANBAN_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${KANBAN_DB_NAME} TO ${KANBAN_DB_USER};
    \c ${KANBAN_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${KANBAN_DB_USER};
EOSQL

echo "✅ Kanban 서비스 데이터베이스 생성 완료: ${KANBAN_DB_NAME}"
echo "🎉 데이터베이스 초기화 완료!"
