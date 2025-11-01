#!/bin/sh
set -e

echo "🔄 Starting Kanban Service..."

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL..."
until PGPASSWORD=$KANBAN_DB_PASSWORD psql -h "postgres" -U "$KANBAN_DB_USER" -d "$KANBAN_DB_NAME" -c '\q' 2>/dev/null; do
  echo "   PostgreSQL is unavailable - sleeping"
  sleep 2
done
echo "✅ PostgreSQL is ready!"

# Run Alembic migrations
echo "🔄 Running database migrations..."

# Check current migration state
CURRENT_VERSION=$(alembic current 2>/dev/null | grep -v "INFO" | grep -v "Context" | grep -v "Will assume" | head -1 || echo "none")
echo "📊 Current migration version: $CURRENT_VERSION"

# Apply migrations to head
# Alembic automatically skips already-applied migrations
echo "⬆️  Upgrading to latest version..."
alembic upgrade head 2>&1 || {
  echo "❌ Migration failed - check logs above"
  exit 1
}

# Verify final state
FINAL_VERSION=$(alembic current 2>/dev/null | grep -v "INFO" | grep -v "Context" | grep -v "Will assume" | head -1 || echo "unknown")
echo "✅ Migrations completed! Final version: $FINAL_VERSION"

# Start the application
echo "🚀 Starting Uvicorn server..."
exec "$@"
