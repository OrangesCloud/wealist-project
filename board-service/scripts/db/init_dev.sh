#!/bin/bash
set -e

# ============================================
# Initialize Development Database
# Usage: ./init_dev.sh
# ============================================

echo "🔄 Initializing development database..."

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found. Please create one from .env.example"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' .env | xargs)

# Extract database connection details from DATABASE_URL
# Format: postgresql://user:password@host:port/dbname
DB_USER=$(echo $DATABASE_URL | sed -n 's/.*:\/\/\([^:]*\):.*/\1/p')
DB_PASSWORD=$(echo $DATABASE_URL | sed -n 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/p')
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
DB_NAME=$(echo $DATABASE_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')

echo "📊 Database details:"
echo "  Host: $DB_HOST:$DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""

# Check PostgreSQL connection
if ! command -v psql &> /dev/null; then
    echo "❌ psql command not found. Please install PostgreSQL client tools."
    exit 1
fi

# Recreate database
echo "🗑️  Dropping existing database (if exists)..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U postgres -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || true

echo "🆕 Creating database..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || true

echo "✅ Database recreated successfully"
echo ""

# Ask user which initialization method to use
echo "📋 Choose initialization method:"
echo "  1) GORM AutoMigrate (run Go server)"
echo "  2) Manual migration (apply SQL files)"
read -p "Enter choice (1 or 2): " -r CHOICE
echo ""

if [ "$CHOICE" = "1" ]; then
    echo "🚀 Starting server with AutoMigrate..."
    echo "📝 Make sure USE_AUTO_MIGRATE=true in your .env file"
    echo ""
    go run cmd/api/main.go
elif [ "$CHOICE" = "2" ]; then
    echo "🚀 Applying manual migrations..."
    ./scripts/db/apply_migrations.sh dev
else
    echo "❌ Invalid choice. Exiting."
    exit 1
fi
