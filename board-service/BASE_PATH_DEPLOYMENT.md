# Board Service Base Path Deployment Guide

## Overview

This document describes the base path configuration for Board Service to support ALB (Application Load Balancer) path-based routing in AWS environments.

## Configuration

### Environment Variable

The base path is configured via the `SERVER_BASE_PATH` environment variable:

```bash
# AWS Environment (with ALB)
SERVER_BASE_PATH=/api/boards

# Local Development (no ALB)
SERVER_BASE_PATH=
```

### Docker Compose

The environment variable is set in `docker/compose/docker-compose.ec2-dev.yml`:

```yaml
board-service:
  environment:
    - SERVER_BASE_PATH=/api/boards
```

## How It Works

### Request Flow

```
Client Request: GET /api/boards/health
    ↓
ALB (Path Pattern: /api/boards/*)
    ↓
Board Service (Base Path: /api/boards)
    ↓
Handler receives: GET /health
```

### Code Implementation

1. **Config Structure** (`internal/config/config.go`):
   - Added `BasePath` field to `ServerConfig`
   - Loads from `SERVER_BASE_PATH` environment variable

2. **Router Setup** (`internal/router/router.go`):
   - Creates base path group if configured
   - All routes are registered under the base path group

3. **Main Application** (`cmd/api/main.go`):
   - Passes `BasePath` from config to router

## Deployment Steps

### 1. Build and Push Docker Image

```bash
# Build the image
docker build -t board-service:latest -f docker/Dockerfile .

# Tag for ECR
docker tag board-service:latest ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/wealist-dev-board-service:latest

# Push to ECR
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/wealist-dev-board-service:latest
```

### 2. Deploy via CI/CD

The GitHub Actions workflow will automatically:
- Build the Docker image
- Push to ECR
- Deploy to EC2 via Docker Compose

### 3. Verify Deployment

Run the verification script:

```bash
# Local verification (no base path)
./scripts/verify-base-path.sh

# AWS verification (with base path)
BASE_URL=https://api.wealist.co.kr BASE_PATH=/api/boards ./scripts/verify-base-path.sh
```

### 4. Check Logs

```bash
# View service logs
docker compose -f docker/compose/docker-compose.ec2-dev.yml logs -f board-service

# Look for base path configuration message
# Expected: "Base path configured for ALB routing" with base_path="/api/boards"
```

## Testing

### Health Check

```bash
# With base path (AWS)
curl https://api.wealist.co.kr/api/boards/health

# Without base path (Local)
curl http://localhost:8000/health
```

### API Endpoints

```bash
# With base path (AWS)
curl -H "Authorization: Bearer $TOKEN" https://api.wealist.co.kr/api/boards/api/projects

# Without base path (Local)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/projects
```

## Rollback

If issues occur, rollback by:

1. Remove or empty the `SERVER_BASE_PATH` environment variable
2. Redeploy the service
3. Update ALB configuration to route directly

## Troubleshooting

### Issue: 404 Not Found

**Cause**: Base path mismatch between ALB and service configuration

**Solution**:
- Verify `SERVER_BASE_PATH` is set to `/api/boards`
- Check ALB listener rule path pattern is `/api/boards/*`
- Review service logs for base path configuration

### Issue: Health Check Failing

**Cause**: ALB health check path doesn't include base path

**Solution**:
- Update Target Group health check path to `/api/boards/health`
- Verify service is responding at the correct path

### Issue: Routes Not Working

**Cause**: Router not applying base path correctly

**Solution**:
- Check logs for "Base path configured" message
- Verify all routes are registered under base path group
- Test with curl to confirm actual endpoint paths
