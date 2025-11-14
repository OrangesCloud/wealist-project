-- ============================================
-- Add participant_ids to boards table
-- Created: 2025-11-14
-- Description: Add participant_ids column to support multiple assignees (participants)
-- ============================================

-- Add participant_ids column (UUID array)
ALTER TABLE boards ADD COLUMN IF NOT EXISTS participant_ids UUID[] DEFAULT '{}';

-- Create GIN index for efficient participant lookups
CREATE INDEX IF NOT EXISTS idx_boards_participant_ids ON boards USING GIN (participant_ids);

COMMENT ON COLUMN boards.participant_ids IS 'Array of user IDs who are participants (multiple assignees) of this board';

-- Insert migration version
INSERT INTO schema_versions (version, description)
VALUES ('20251114120000', 'Add participant_ids column to boards table')
ON CONFLICT (version) DO NOTHING;
