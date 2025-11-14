-- ============================================
-- Rollback: Remove participant_ids from boards table
-- Created: 2025-11-14
-- ============================================

-- Drop index
DROP INDEX IF EXISTS idx_boards_participant_ids;

-- Drop column
ALTER TABLE boards DROP COLUMN IF EXISTS participant_ids;

-- Remove migration version
DELETE FROM schema_versions WHERE version = '20251114120000';
