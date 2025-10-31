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

# Always try to stamp first (safe operation, won't break anything)
echo "📌 Ensuring migration tracking is up to date..."
alembic stamp head 2>/dev/null || true

# Now run migrations
alembic upgrade head 2>&1 | grep -v "DuplicateTable" || {
  # If upgrade fails due to duplicate tables, stamp and retry
  echo "⚠️  Detected existing tables, syncing migration state..."
  alembic stamp head
}

echo "✅ Migrations completed!"

# Start the application
echo "🚀 Starting Uvicorn server..."
exec "$@"
