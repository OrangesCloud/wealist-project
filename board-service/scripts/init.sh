#!/bin/sh
set -e

echo "==================================="
echo "Starting Board Service Initialization"
echo "==================================="

# Wait for database (BOARD_DB_* 환경변수 사용)
echo "Waiting for PostgreSQL..."
until PGPASSWORD=$BOARD_DB_PASSWORD psql -h postgres -U $BOARD_DB_USER -d $BOARD_DB_NAME -c '\q' 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up!"

# Run migrations
echo "Running migrations..."
migrate -path /app/migrations -database "$DATABASE_URL" up

if [ $? -eq 0 ]; then
    echo "✅ Migrations completed successfully!"
else
    echo "❌ Migration failed!"
    exit 1
fi

# Start application
echo "Starting board-api..."
exec /app/board-api
