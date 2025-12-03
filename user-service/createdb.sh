#!/bin/bash
set -e

echo "========================================="
echo "Starting Database Initialization Script"
echo "========================================="
echo "POSTGRES_USER: $POSTGRES_USER"
echo "USER_DB_NAME: $USER_DB_NAME"
echo "USER_DB_USER: $USER_DB_USER"
echo "BOARD_DB_NAME: $BOARD_DB_NAME" 
echo "BOARD_DB_USER: $BOARD_DB_USER"
echo "========================================="

# 1단계: 사용자 및 데이터베이스 생성
echo "Step 1: Creating users and databases..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
    -- wealist_user 생성
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$USER_DB_USER') THEN
            CREATE USER $USER_DB_USER WITH PASSWORD '$USER_DB_PASSWORD';
            RAISE NOTICE 'User $USER_DB_USER created successfully';
        ELSE
            RAISE NOTICE 'User $USER_DB_USER already exists';
        END IF;
    END \$\$;

    -- board_service 생성  
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$BOARD_DB_USER') THEN
            CREATE USER $BOARD_DB_USER WITH PASSWORD '$BOARD_DB_PASSWORD';
            RAISE NOTICE 'User $BOARD_DB_USER created successfully';
        ELSE
            RAISE NOTICE 'User $BOARD_DB_USER already exists';
        END IF;
    END \$\$;

    -- 데이터베이스 생성
    DROP DATABASE IF EXISTS $USER_DB_NAME;
    CREATE DATABASE $USER_DB_NAME OWNER $USER_DB_USER;
    GRANT ALL PRIVILEGES ON DATABASE $USER_DB_NAME TO $USER_DB_USER;
    
    DROP DATABASE IF EXISTS $BOARD_DB_NAME;
    CREATE DATABASE $BOARD_DB_NAME OWNER $BOARD_DB_USER;
    GRANT ALL PRIVILEGES ON DATABASE $BOARD_DB_NAME TO $BOARD_DB_USER;

    SELECT 'Databases created: $USER_DB_NAME, $BOARD_DB_NAME' as result;
EOSQL

# 2단계: wealist_user_db에 테이블 생성
echo "Step 2: Creating tables in $USER_DB_NAME..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$USER_DB_NAME" <<-EOSQL
    -- UUID 확장 활성화
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    
    -- 1. users 테이블
    CREATE TABLE IF NOT EXISTS users (
        "userId" UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
        email VARCHAR(255) NOT NULL UNIQUE,
        provider VARCHAR(255) DEFAULT 'google',
        "googleId" VARCHAR(255) UNIQUE,
        "createdAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "updatedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "isActive" BOOLEAN DEFAULT true NOT NULL,
        "deletedAt" TIMESTAMP NULL
    );
    
    CREATE UNIQUE INDEX IF NOT EXISTS uc_user_email ON users (email);
    CREATE INDEX IF NOT EXISTS idx_user_email ON users (email);
    
    -- 2. workspaces 테이블
    CREATE TABLE IF NOT EXISTS workspaces (
        "workspaceId" UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
        "ownerId" UUID NOT NULL,
        "workspaceName" VARCHAR(255) NOT NULL,
        "workspaceDescription" TEXT NOT NULL,
        "isPublic" BOOLEAN DEFAULT true NOT NULL,
        "needApproved" BOOLEAN DEFAULT true NOT NULL,
        "createdAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "deletedAt" TIMESTAMP NULL,
        "isActive" BOOLEAN DEFAULT true NOT NULL,
        FOREIGN KEY ("ownerId") REFERENCES users("userId")
    );
    
    CREATE UNIQUE INDEX IF NOT EXISTS uc_workspace_member_unique ON workspaces ("workspaceId", "ownerId");
    
    -- 3. userProfile 테이블
    CREATE TABLE IF NOT EXISTS "userProfile" (
        "profileId" UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
        "workspaceId" UUID NOT NULL,
        "userId" UUID NOT NULL,
        "nickName" VARCHAR(50),
        email VARCHAR(100),
        "profileImageUrl" TEXT,
        "createdAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "updatedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY ("userId") REFERENCES users("userId"),
        FOREIGN KEY ("workspaceId") REFERENCES workspaces("workspaceId")
    );
    
    CREATE UNIQUE INDEX IF NOT EXISTS userprofile_userid_workspaceid_unique ON "userProfile" ("userId", "workspaceId");
    
    -- 4. workspaceJoinRequests 테이블
    CREATE TABLE IF NOT EXISTS "workspaceJoinRequests" (
        "joinRequestId" UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
        "workspaceId" UUID NOT NULL,
        "userId" UUID NOT NULL,
        status VARCHAR(50) DEFAULT 'PENDING' NOT NULL,
        "requestedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "updatedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY ("workspaceId") REFERENCES workspaces("workspaceId"),
        FOREIGN KEY ("userId") REFERENCES users("userId"),
        CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
    );
    
    -- 5. workspaceMembers 테이블
    CREATE TABLE IF NOT EXISTS "workspaceMembers" (
        "workspaceMemberId" UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
        "workspaceId" UUID NOT NULL,
        "userId" UUID NOT NULL,
        "roleName" VARCHAR(50) NOT NULL,
        "isDefault" BOOLEAN DEFAULT false NOT NULL,
        "joinedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "updatedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "isActive" BOOLEAN DEFAULT true NOT NULL,
        FOREIGN KEY ("workspaceId") REFERENCES workspaces("workspaceId"),
        FOREIGN KEY ("userId") REFERENCES users("userId"),
        CHECK ("roleName" IN ('OWNER', 'ADMIN', 'MEMBER'))
    );
    
    -- 6. attachments 테이블
    CREATE TABLE IF NOT EXISTS attachments (
        id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
        "entityType" VARCHAR(50) NOT NULL,
        "entityId" UUID NULL,
        status VARCHAR(20) DEFAULT 'TEMP' NOT NULL,
        "fileName" VARCHAR(255) NOT NULL,
        "fileUrl" TEXT NOT NULL,
        "fileSize" BIGINT NOT NULL,
        "contentType" VARCHAR(100) NOT NULL,
        "uploadedBy" UUID NOT NULL,
        "expiresAt" TIMESTAMP NULL,
        "createdAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "updatedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "deletedAt" TIMESTAMP NULL,
        FOREIGN KEY ("uploadedBy") REFERENCES users("userId"),
        CHECK ("entityType" IN ('USER_PROFILE')),
        CHECK (status IN ('TEMP', 'CONFIRMED'))
    );
    
    -- attachments 인덱스
    CREATE INDEX IF NOT EXISTS idx_attachments_entity ON attachments ("entityType", "entityId");
    CREATE INDEX IF NOT EXISTS idx_attachments_entity_id ON attachments ("entityId");
    CREATE INDEX IF NOT EXISTS idx_attachments_status ON attachments (status);
    CREATE INDEX IF NOT EXISTS idx_attachments_uploaded_by ON attachments ("uploadedBy");
    CREATE INDEX IF NOT EXISTS idx_attachments_expires_at ON attachments ("expiresAt");
    
    -- 트리거 함수 생성
    CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS \$trigger\$
    BEGIN
        NEW."updatedAt" = CURRENT_TIMESTAMP;
        RETURN NEW;
    END;
    \$trigger\$ language 'plpgsql';
    
    -- 트리거 생성
    DROP TRIGGER IF EXISTS update_users_updated_at ON users;
    CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        
    DROP TRIGGER IF EXISTS update_userprofile_updated_at ON "userProfile";
    CREATE TRIGGER update_userprofile_updated_at BEFORE UPDATE ON "userProfile" 
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        
    DROP TRIGGER IF EXISTS update_workspacejoinrequests_updated_at ON "workspaceJoinRequests";
    CREATE TRIGGER update_workspacejoinrequests_updated_at BEFORE UPDATE ON "workspaceJoinRequests" 
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        
    DROP TRIGGER IF EXISTS update_workspacemembers_updated_at ON "workspaceMembers";
    CREATE TRIGGER update_workspacemembers_updated_at BEFORE UPDATE ON "workspaceMembers" 
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        
    DROP TRIGGER IF EXISTS update_attachments_updated_at ON attachments;
    CREATE TRIGGER update_attachments_updated_at BEFORE UPDATE ON attachments 
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
    -- 권한 부여
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $USER_DB_USER;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $USER_DB_USER;
    GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO $USER_DB_USER;
    GRANT USAGE ON SCHEMA public TO $USER_DB_USER;
    
    SELECT 'Tables created successfully in $USER_DB_NAME!' as result;
EOSQL

echo "========================================="
echo "Database initialization completed!"
echo "========================================="
EOSQL