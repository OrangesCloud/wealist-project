CREATE TABLE IF NOT EXISTS kanbans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    stage_id UUID NOT NULL,
    importance_id UUID,
    assignee_id UUID,
    author_id UUID NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_kanbans_project_id ON kanbans(project_id);
CREATE INDEX idx_kanbans_stage_id ON kanbans(stage_id);
CREATE INDEX idx_kanbans_importance_id ON kanbans(importance_id);
CREATE INDEX idx_kanbans_assignee_id ON kanbans(assignee_id);
CREATE INDEX idx_kanbans_author_id ON kanbans(author_id);
CREATE INDEX idx_kanbans_deleted_at ON kanbans(deleted_at);
CREATE INDEX idx_kanbans_due_date ON kanbans(due_date);

COMMENT ON TABLE kanbans IS 'Kanban cards (tasks/issues) for projects';
COMMENT ON COLUMN kanbans.project_id IS 'References projects.id (no FK for sharding compatibility)';
COMMENT ON COLUMN kanbans.stage_id IS 'References custom_stages.id (no FK, required field)';
COMMENT ON COLUMN kanbans.importance_id IS 'References custom_importance.id (no FK, optional)';
COMMENT ON COLUMN kanbans.assignee_id IS 'References users.id in User Service (no FK, optional)';
COMMENT ON COLUMN kanbans.author_id IS 'References users.id in User Service (no FK, required)';
COMMENT ON COLUMN kanbans.deleted_at IS 'Soft delete timestamp';
