#!/bin/bash
set -e

# ============================================
# Dump Database Schema
# Usage: ./dump_schema.sh [env]
# Example: ./dump_schema.sh dev
# ============================================

ENV=${1:-dev}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

CONFIG_FILE=".env"
if [ "$ENV" = "prod" ]; then
    CONFIG_FILE=".env.production"
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Config file not found: $CONFIG_FILE"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' "$CONFIG_FILE" | xargs)

echo "📦 Dumping schema from $ENV environment..."
echo "📍 Database: $DATABASE_URL"

# Create directories if they don't exist
mkdir -p docs/schema/backups

# Dump schema only (no data)
OUTPUT_FILE="docs/schema/backups/schema_${TIMESTAMP}.sql"

pg_dump "$DATABASE_URL" \
    --schema-only \
    --no-owner \
    --no-privileges \
    > "$OUTPUT_FILE"

echo "✅ Schema dumped: $OUTPUT_FILE"

# Update latest.sql symlink (use relative path)
cd docs/schema
rm -f latest.sql
ln -s "backups/schema_${TIMESTAMP}.sql" latest.sql
cd ../..

echo "📌 Latest link updated: docs/schema/latest.sql"

# Show file size
SIZE=$(ls -lh "$OUTPUT_FILE" | awk '{print $5}')
echo "📏 File size: $SIZE"

# Show current schema version from database
echo ""
echo "📊 Current schema version:"
psql "$DATABASE_URL" -c "SELECT version, description, applied_at FROM schema_versions ORDER BY applied_at DESC LIMIT 5;"
