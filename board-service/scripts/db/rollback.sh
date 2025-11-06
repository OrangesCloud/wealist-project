#!/bin/bash
set -e

# ============================================
# Rollback Database Migration
# Usage: ./rollback.sh [env] [version]
# Example: ./rollback.sh prod 20250106120000
# ============================================

ENV=${1:-prod}
TARGET_VERSION=${2:-""}

if [ -z "$TARGET_VERSION" ]; then
    echo "❌ Usage: ./rollback.sh [env] [version]"
    echo "Example: ./rollback.sh prod 20250106120000"
    exit 1
fi

CONFIG_FILE=".env.${ENV}"
if [ "$ENV" = "dev" ]; then
    CONFIG_FILE=".env"
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Config file not found: $CONFIG_FILE"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' "$CONFIG_FILE" | xargs)

echo "🔙 Rolling back to version: $TARGET_VERSION"
echo "📍 Database: $DATABASE_URL"

# Find the down migration file
DOWN_FILE=$(ls migrations/${TARGET_VERSION}_*.down.sql 2>/dev/null | head -n 1)

if [ -z "$DOWN_FILE" ] || [ ! -f "$DOWN_FILE" ]; then
    echo "❌ Down migration file not found for version: $TARGET_VERSION"
    exit 1
fi

# Check if migration was applied
APPLIED=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM schema_versions WHERE version='$TARGET_VERSION'" 2>/dev/null || echo "0")
APPLIED=$(echo "$APPLIED" | xargs)

if [ "$APPLIED" -eq 0 ]; then
    echo "⚠️  Migration $TARGET_VERSION was not applied, nothing to rollback"
    exit 0
fi

echo ""
echo "⚠️  WARNING: This will rollback migration $TARGET_VERSION"
echo "File: $DOWN_FILE"
echo ""
read -p "Are you sure? (yes/no): " -r
echo

if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "❌ Rollback cancelled"
    exit 0
fi

echo "▶️  Rolling back: $DOWN_FILE"

# Apply rollback
if psql "$DATABASE_URL" -f "$DOWN_FILE"; then
    echo "✅ Rollback completed: $TARGET_VERSION"
else
    echo "❌ Failed to rollback migration: $TARGET_VERSION"
    exit 1
fi

echo ""
echo "📊 Current schema version:"
psql "$DATABASE_URL" -c "SELECT version, description, applied_at FROM schema_versions ORDER BY applied_at DESC LIMIT 5;"
