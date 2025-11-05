-- 워크스페이스 참여 신청 테이블
CREATE TABLE IF NOT EXISTS workspace_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    processed_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_join_request_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
);

COMMENT ON TABLE workspace_join_requests IS '워크스페이스 참여 신청';
COMMENT ON COLUMN workspace_join_requests.status IS 'PENDING, APPROVED, REJECTED';
COMMENT ON COLUMN workspace_join_requests.user_id IS '신청한 사용자';
COMMENT ON COLUMN workspace_join_requests.processed_by IS '승인/거절 처리한 관리자';

CREATE INDEX idx_workspace_join_requests_workspace ON workspace_join_requests(workspace_id);
CREATE INDEX idx_workspace_join_requests_user ON workspace_join_requests(user_id);
CREATE INDEX idx_workspace_join_requests_status ON workspace_join_requests(status);

-- 워크스페이스 멤버 테이블
CREATE TABLE IF NOT EXISTS workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, user_id)
);

COMMENT ON TABLE workspace_members IS '워크스페이스 멤버';
COMMENT ON COLUMN workspace_members.role_id IS 'roles 테이블의 id';
COMMENT ON COLUMN workspace_members.is_default IS '대표 워크스페이스 여부';
COMMENT ON COLUMN workspace_members.left_at IS '탈퇴 시간 (NULL이면 현재 멤버)';

CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_role ON workspace_members(role_id);
CREATE INDEX idx_workspace_members_is_default ON workspace_members(user_id, is_default) WHERE is_default = TRUE;
