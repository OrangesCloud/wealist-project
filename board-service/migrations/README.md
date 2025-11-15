# Database Migrations

This directory contains SQL migration files for the Project Board Management System.

## Migration Files

- `001_init_schema.sql` - Initial schema creation (up migration)
- `001_init_schema_down.sql` - Schema rollback (down migration)

## Schema Overview

The migration creates the following tables:

1. **projects** - Projects within workspaces
2. **boards** - Work items with stage, importance, and role attributes
3. **participants** - Users participating in boards
4. **comments** - Discussion comments on boards

## Running Migrations

### Using psql

```bash
# Apply migration
psql -U postgres -d project_board -f migrations/001_init_schema.sql

# Rollback migration
psql -U postgres -d project_board -f migrations/001_init_schema_down.sql
```

### Using Docker

```bash
# Apply migration
docker exec -i postgres_container psql -U postgres -d project_board < migrations/001_init_schema.sql

# Rollback migration
docker exec -i postgres_container psql -U postgres -d project_board < migrations/001_init_schema_down.sql
```

### Using Makefile

Add these targets to your Makefile:

```makefile
migrate-up:
	psql $(DATABASE_URL) -f migrations/001_init_schema.sql

migrate-down:
	psql $(DATABASE_URL) -f migrations/001_init_schema_down.sql
```

## Features

### Automatic Timestamps

All tables include automatic `updated_at` timestamp updates via database triggers.

### Soft Deletes

All tables support soft deletes through the `deleted_at` column. Indexes are optimized for queries that filter out soft-deleted records.

### Constraints

- **Foreign Keys**: Cascade deletes to maintain referential integrity
- **Unique Constraints**: Prevent duplicate participants per board
- **Check Constraints**: Validate enum values for stage, importance, and role

### Indexes

Optimized indexes for:
- Foreign key lookups
- Soft delete filtering
- Workspace and project queries
- Board filtering by stage, importance, and role
- Comment ordering by creation time

## Schema Diagram

```
projects (1) ──< (N) boards (1) ──< (N) participants
                         │
                         └──< (N) comments
```

## Notes

- All IDs use UUID type with automatic generation
- The `pgcrypto` extension is required for UUID generation
- Timestamps use PostgreSQL's `TIMESTAMP` type (without timezone)
- All tables follow the soft delete pattern with `deleted_at` column
