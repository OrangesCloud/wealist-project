-- ============================================
-- Project Board Management System
-- Migration 005: Add Custom Fields to Boards
-- ============================================
-- This migration adds a JSONB custom_fields column to boards table
-- and migrates existing stage, importance, and role data into it

-- ============================================
-- Step 1: Add custom_fields JSONB column
-- ============================================
ALTER TABLE boards ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

-- ============================================
-- Step 2: Migrate existing data to custom_fields
-- ============================================
-- Copy stage, importance, and role values into custom_fields JSONB
UPDATE boards 
SET custom_fields = jsonb_build_object(
    'stage', stage,
    'role', role,
    'importance', importance
)
WHERE custom_fields = '{}'::jsonb OR custom_fields IS NULL;

-- ============================================
-- Step 3: Create GIN index for JSONB queries
-- ============================================
-- This index optimizes queries filtering by custom_fields
CREATE INDEX IF NOT EXISTS idx_boards_custom_fields ON boards USING GIN (custom_fields);

-- ============================================
-- Step 4: Drop old indexes (will be removed in future migration)
-- ============================================
-- Note: We keep the old columns for now to allow rollback
-- They will be removed in a separate migration after verification

-- ============================================
-- Comments for documentation
-- ============================================
COMMENT ON COLUMN boards.custom_fields IS 'JSONB field storing custom board attributes (stage, role, importance, etc.)';
