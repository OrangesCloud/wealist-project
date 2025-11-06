#!/bin/bash
set -e

# ============================================
# Apply Database Migrations
# Usage: ./apply_migrations.sh [env]
# Example: ./apply_migrations.sh prod
# ============================================

# Environment setup
ENV=${1:-prod}
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

echo "🚀 Applying migrations to $ENV environment..."
echo "📍 Database: $DATABASE_URL"

# Check if migrations directory exists
if [ ! -d "migrations" ]; then
    echo "❌ migrations/ directory not found"
    exit 1
fi

# Apply migrations in order
for file in migrations/*up.sql; do
    if [ -f "$file" ]; then
        VERSION=$(basename "$file" | cut -d'_' -f1)
        DESCRIPTION=$(basename "$file" .up.sql | cut -d'_' -f2- | sed 's/_/ /g')

        echo ""
        echo "📝 Checking migration: $VERSION"

        # Check if already applied
        APPLIED=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM schema_versions WHERE version='$VERSION'" 2>/dev/null || echo "0")
        APPLIED=$(echo "$APPLIED" | xargs)  # Trim whitespace

        if [ "$APPLIED" -gt 0 ]; then
            echo "⏭️  Already applied: $VERSION - $DESCRIPTION"
            continue
        fi

        echo "▶️  Applying: $file"

        # Apply migration
        if psql "$DATABASE_URL" -f "$file"; then
            echo "✅ Applied: $VERSION - $DESCRIPTION"
        else
            echo "❌ Failed to apply migration: $VERSION"
            exit 1
        fi
    fi
done

echo ""
echo "🎉 All migrations applied successfully!"
echo ""
echo "📊 Current schema version:"
psql "$DATABASE_URL" -c "SELECT version, description, applied_at FROM schema_versions ORDER BY applied_at DESC LIMIT 5;"
