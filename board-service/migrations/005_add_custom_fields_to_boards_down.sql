-- ============================================
-- Project Board Management System
-- Migration 005 Rollback: Remove Custom Fields from Boards
-- ============================================
-- This rollback script restores the original board structure
-- by removing the custom_fields column

-- ============================================
-- Step 1: Verify data integrity (optional check)
-- ============================================
-- Before rollback, ensure stage, importance, and role columns still exist
-- If they were dropped, this rollback cannot restore data

-- ============================================
-- Step 2: Restore data from custom_fields to original columns (if needed)
-- ============================================
-- If the original columns were dropped, this would need to restore them first
-- For now, we assume the columns still exist (they're dropped in a later migration)

-- Update original columns from custom_fields if they exist
DO $
BEGIN
    -- Check if we need to restore data
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'boards' AND column_name = 'custom_fields'
    ) THEN
        -- Restore stage, importance, role from custom_fields if columns exist
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'boards' AND column_name = 'stage'
        ) THEN
            UPDATE boards 
            SET stage = COALESCE(custom_fields->>'stage', stage)
            WHERE custom_fields IS NOT NULL AND custom_fields->>'stage' IS NOT NULL;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'boards' AND column_name = 'importance'
        ) THEN
            UPDATE boards 
            SET importance = COALESCE(custom_fields->>'importance', importance)
            WHERE custom_fields IS NOT NULL AND custom_fields->>'importance' IS NOT NULL;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'boards' AND column_name = 'role'
        ) THEN
            UPDATE boards 
            SET role = COALESCE(custom_fields->>'role', role)
            WHERE custom_fields IS NOT NULL AND custom_fields->>'role' IS NOT NULL;
        END IF;
    END IF;
END $;

-- ============================================
-- Step 3: Drop GIN index
-- ============================================
DROP INDEX IF EXISTS idx_boards_custom_fields;

-- ============================================
-- Step 4: Drop custom_fields column
-- ============================================
ALTER TABLE boards DROP COLUMN IF EXISTS custom_fields;

-- ============================================
-- Comments
-- ============================================
-- Note: This rollback assumes the original stage, importance, and role columns
-- still exist. If they were dropped in a later migration, those columns would
-- need to be recreated first before running this rollback.
