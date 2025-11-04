CREATE TABLE IF NOT EXISTS project_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, user_id)
);

CREATE INDEX idx_project_join_requests_project_id ON project_join_requests(project_id);
CREATE INDEX idx_project_join_requests_user_id ON project_join_requests(user_id);
CREATE INDEX idx_project_join_requests_status ON project_join_requests(status);

COMMENT ON COLUMN project_join_requests.project_id IS 'References projects.id (no FK for microservice isolation)';
COMMENT ON COLUMN project_join_requests.user_id IS 'References users.id from User Service (no FK for microservice isolation)';
