CREATE TABLE IF NOT EXISTS custom_importance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7),
    is_system_default BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_project_importance_name UNIQUE(project_id, name)
);

CREATE INDEX idx_custom_importance_project_id ON custom_importance(project_id);
CREATE INDEX idx_custom_importance_is_system_default ON custom_importance(is_system_default);
CREATE INDEX idx_custom_importance_is_deleted ON custom_importance(is_deleted);

COMMENT ON TABLE custom_importance IS 'Project-specific custom importance levels for kanbans';
COMMENT ON COLUMN custom_importance.project_id IS 'References projects.id (no FK for sharding compatibility)';
COMMENT ON COLUMN custom_importance.is_system_default IS 'System default values cannot be deleted';
COMMENT ON COLUMN custom_importance.display_order IS 'Display order for UI sorting';
