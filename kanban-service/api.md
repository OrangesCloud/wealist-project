# Simple API Guide

This guide provides simple `curl` examples for key APIs.

---

## 1. User Service (Port: 8080)

### 1.1. User Signup

Creates a new user account.

**Endpoint:** `POST http://localhost:8080/api/auth/signup`

**Request Body:**
```json
{
  "name": "your_username",
  "email": "user@example.com",
  "password": "your_password"
}
```

**Example `curl` command:**
```bash
curl -X POST -H "Content-Type: application/json" \
-d '{
  "name": "testuser",
  "email": "testuser123@example.com",
  "password": "password123"
}' \
http://localhost:8080/api/auth/signup
```

### 1.2. User Login

Authenticates a user and returns JWT tokens.

**Endpoint:** `POST http://localhost:8080/api/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "your_password"
}
```

**Example `curl` command:**
```bash
curl -X POST -H "Content-Type: application/json" \
-d '{
  "email": "testuser123@example.com",
  "password": "password123"
}' \
http://localhost:8080/api/auth/login
```
**Successful Response:**
```json
{
    "accessToken": "ey...",
    "refreshToken": "ey...",
    "userId": "...",
    "name": "testuser",
    "email": "testuser123@example.com",
    "tokenType": "Bearer"
}
```

---

## 2. Board (Kanban) Service (Port: 8000)

**Note:** All requests to the board service require an `Authorization` header with the `accessToken` obtained from the login response.

**Header Format:** `Authorization: Bearer <your_access_token>`

### 2.1. List Workspaces

Retrieves a list of all workspaces.

**Endpoint:** `GET http://localhost:8000/api/workspaces/`

**Example `curl` command:**
```bash
# Replace <your_access_token> with the actual token
ACCESS_TOKEN="<your_access_token>"

curl -H "Authorization: Bearer $ACCESS_TOKEN" http://localhost:8000/api/workspaces/
```

### 2.2. Create a Workspace

Creates a new workspace.

**Endpoint:** `POST http://localhost:8000/api/workspaces/`

**Request Body:**
```json
{
  "name": "My New Workspace",
  "description": "A description for the new workspace."
}
```

**Example `curl` command:**
```bash
# Replace <your_access_token> with the actual token
ACCESS_TOKEN="<your_access_token>"

curl -X POST -H "Content-Type: application/json" \
-H "Authorization: Bearer $ACCESS_TOKEN" \
-d '{
  "name": "My New Workspace",
  "description": "A description for the new workspace."
}' \
http://localhost:8000/api/workspaces/
```

### 2.3. List Projects

Retrieves a list of projects, optionally filtered by workspace.

**Endpoint:** `GET http://localhost:8000/api/projects/`

**Example `curl` command (all projects):**
```bash
# Replace <your_access_token> with the actual token
ACCESS_TOKEN="<your_access_token>"

curl -H "Authorization: Bearer $ACCESS_TOKEN" http://localhost:8000/api/projects/
```

**Example `curl` command (filter by workspace):**
```bash
# Replace <your_access_token> and <workspace_id>
ACCESS_TOKEN="<your_access_token>"
WORKSPACE_ID="<workspace_id_from_list_workspaces>"

curl -H "Authorization: Bearer $ACCESS_TOKEN" "http://localhost:8000/api/projects/?workspace_id=$WORKSPACE_ID"
```
