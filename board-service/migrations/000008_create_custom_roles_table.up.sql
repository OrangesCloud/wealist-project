CREATE TABLE IF NOT EXISTS custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7),
    is_system_default BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_project_role_name UNIQUE(project_id, name)
);

CREATE INDEX idx_custom_roles_project_id ON custom_roles(project_id);
CREATE INDEX idx_custom_roles_is_system_default ON custom_roles(is_system_default);

COMMENT ON TABLE custom_roles IS 'Project-specific custom roles for kanbans';
COMMENT ON COLUMN custom_roles.project_id IS 'References projects.id (no FK for sharding compatibility)';
COMMENT ON COLUMN custom_roles.is_system_default IS 'System default values cannot be deleted';
COMMENT ON COLUMN custom_roles.display_order IS 'Display order for UI sorting';
