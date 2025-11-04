-- User Role Column Order (역할 컬럼 순서)
CREATE TABLE IF NOT EXISTS user_role_column_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_role_id UUID NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, project_id, custom_role_id)
);

CREATE INDEX idx_user_role_order ON user_role_column_order(user_id, project_id);

COMMENT ON TABLE user_role_column_order IS 'Stores user-specific display order for role columns in project board view';
COMMENT ON COLUMN user_role_column_order.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_role_column_order.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_role_column_order.custom_role_id IS 'References custom_roles.id (no FK for sharding)';

-- User Stage Column Order (진행단계 컬럼 순서)
CREATE TABLE IF NOT EXISTS user_stage_column_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_stage_id UUID NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, project_id, custom_stage_id)
);

CREATE INDEX idx_user_stage_order ON user_stage_column_order(user_id, project_id);

COMMENT ON TABLE user_stage_column_order IS 'Stores user-specific display order for stage columns in project board view';
COMMENT ON COLUMN user_stage_column_order.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_stage_column_order.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_stage_column_order.custom_stage_id IS 'References custom_stages.id (no FK for sharding)';

-- User Kanban Order in Role (역할별 칸반 순서)
CREATE TABLE IF NOT EXISTS user_kanban_order_in_role (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_role_id UUID NOT NULL,
    kanban_id UUID NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, project_id, custom_role_id, kanban_id)
);

CREATE INDEX idx_user_kanban_role_order ON user_kanban_order_in_role(user_id, project_id, custom_role_id);

COMMENT ON TABLE user_kanban_order_in_role IS 'Stores user-specific display order for kanbans within each role column';
COMMENT ON COLUMN user_kanban_order_in_role.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_kanban_order_in_role.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_kanban_order_in_role.custom_role_id IS 'References custom_roles.id (no FK for sharding)';
COMMENT ON COLUMN user_kanban_order_in_role.kanban_id IS 'References kanbans.id (no FK for sharding)';

-- User Kanban Order in Stage (진행단계별 칸반 순서)
CREATE TABLE IF NOT EXISTS user_kanban_order_in_stage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID NOT NULL,
    custom_stage_id UUID NOT NULL,
    kanban_id UUID NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, project_id, custom_stage_id, kanban_id)
);

CREATE INDEX idx_user_kanban_stage_order ON user_kanban_order_in_stage(user_id, project_id, custom_stage_id);

COMMENT ON TABLE user_kanban_order_in_stage IS 'Stores user-specific display order for kanbans within each stage column';
COMMENT ON COLUMN user_kanban_order_in_stage.user_id IS 'References users.id (no FK for microservice isolation)';
COMMENT ON COLUMN user_kanban_order_in_stage.project_id IS 'References projects.id (no FK for sharding)';
COMMENT ON COLUMN user_kanban_order_in_stage.custom_stage_id IS 'References custom_stages.id (no FK for sharding)';
COMMENT ON COLUMN user_kanban_order_in_stage.kanban_id IS 'References kanbans.id (no FK for sharding)';
