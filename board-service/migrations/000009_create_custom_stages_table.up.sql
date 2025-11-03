CREATE TABLE IF NOT EXISTS custom_stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7),
    is_system_default BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_project_stage_name UNIQUE(project_id, name)
);

CREATE INDEX idx_custom_stages_project_id ON custom_stages(project_id);
CREATE INDEX idx_custom_stages_is_system_default ON custom_stages(is_system_default);

COMMENT ON TABLE custom_stages IS 'Project-specific custom stages for kanbans';
COMMENT ON COLUMN custom_stages.project_id IS 'References projects.id (no FK for sharding compatibility)';
COMMENT ON COLUMN custom_stages.is_system_default IS 'System default values cannot be deleted';
COMMENT ON COLUMN custom_stages.display_order IS 'Display order for UI sorting';
