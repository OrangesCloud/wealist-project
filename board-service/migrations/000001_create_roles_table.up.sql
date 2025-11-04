-- 권한 테이블
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    level INT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE roles IS '권한 테이블 (OWNER=100, ADMIN=50, MEMBER=10)';
COMMENT ON COLUMN roles.name IS 'OWNER, ADMIN, MEMBER, PROJECT_LEADER 등';
COMMENT ON COLUMN roles.level IS '권한 레벨 (숫자가 클수록 높은 권한)';

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_level ON roles(level);
