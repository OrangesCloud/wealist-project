-- Many-to-many relationship between kanbans and custom_roles
CREATE TABLE IF NOT EXISTS kanban_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kanban_id UUID NOT NULL,
    custom_role_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_kanban_custom_role UNIQUE(kanban_id, custom_role_id)
);

CREATE INDEX idx_kanban_roles_kanban_id ON kanban_roles(kanban_id);
CREATE INDEX idx_kanban_roles_custom_role_id ON kanban_roles(custom_role_id);

COMMENT ON TABLE kanban_roles IS 'Many-to-many relationship between kanbans and custom roles';
COMMENT ON COLUMN kanban_roles.kanban_id IS 'References kanbans.id (no FK for sharding compatibility)';
COMMENT ON COLUMN kanban_roles.custom_role_id IS 'References custom_roles.id (no FK for sharding compatibility)';
