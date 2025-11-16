# Database Migrations

This directory contains SQL migration files for the Project Board Management System.

## Migration Files

- `001_init_schema.sql` - Initial schema creation (up migration)
- `001_init_schema_down.sql` - Schema rollback (down migration)
- `002_add_project_members_and_board_fields.sql` - Add project members, join requests, and board fields (up migration)
- `002_add_project_members_and_board_fields_down.sql` - Rollback project members and board fields (down migration)
- `003_migrate_existing_project_owners.sql` - Create OWNER members for existing projects (up migration)
- `003_migrate_existing_project_owners_down.sql` - Remove migrated OWNER members (down migration)

## Schema Overview

The migrations create the following tables:

1. **projects** - Projects within workspaces (with owner_id and is_public fields)
2. **boards** - Work items with stage, importance, and role attributes (with author_id, assignee_id, and due_date fields)
3. **participants** - Users participating in boards
4. **comments** - Discussion comments on boards
5. **project_members** - Members of projects with roles (OWNER, ADMIN, MEMBER)
6. **project_join_requests** - Requests to join projects with approval workflow

## Running Migrations

### Using psql

```bash
# Apply migrations
psql -U postgres -d project_board -f migrations/001_init_schema.sql
psql -U postgres -d project_board -f migrations/002_add_project_members_and_board_fields.sql
psql -U postgres -d project_board -f migrations/003_migrate_existing_project_owners.sql

# Rollback migrations (in reverse order)
psql -U postgres -d project_board -f migrations/003_migrate_existing_project_owners_down.sql
psql -U postgres -d project_board -f migrations/002_add_project_members_and_board_fields_down.sql
psql -U postgres -d project_board -f migrations/001_init_schema_down.sql
```

### Using Docker

```bash
# Apply migrations
docker exec -i postgres_container psql -U postgres -d project_board < migrations/001_init_schema.sql
docker exec -i postgres_container psql -U postgres -d project_board < migrations/002_add_project_members_and_board_fields.sql
docker exec -i postgres_container psql -U postgres -d project_board < migrations/003_migrate_existing_project_owners.sql

# Rollback migrations (in reverse order)
docker exec -i postgres_container psql -U postgres -d project_board < migrations/003_migrate_existing_project_owners_down.sql
docker exec -i postgres_container psql -U postgres -d project_board < migrations/002_add_project_members_and_board_fields_down.sql
docker exec -i postgres_container psql -U postgres -d project_board < migrations/001_init_schema_down.sql
```

### Using Makefile

```bash
# Apply all migrations
make migrate-up

# Rollback all migrations
make migrate-down
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
    │                    │
    │                    └──< (N) comments
    │
    ├──< (N) project_members
    │
    └──< (N) project_join_requests
```

## Notes

- All IDs use UUID type with automatic generation
- The `pgcrypto` extension is required for UUID generation
- Timestamps use PostgreSQL's `TIMESTAMP` type (without timezone)
- All tables follow the soft delete pattern with `deleted_at` column
