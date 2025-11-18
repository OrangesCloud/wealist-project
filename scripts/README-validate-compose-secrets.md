# Docker Compose 보안 검증 스크립트

## 개요

`validate-compose-secrets.sh`는 docker-compose 파일에서 하드코딩된 시크릿을 자동으로 감지하는 보안 검증 스크립트입니다.

## 기능

다음과 같은 시크릿 패턴을 감지합니다:

1. **하드코딩된 비밀번호**: `POSTGRES_PASSWORD`, `REDIS_PASSWORD`, `DB_PASSWORD`, `password` 등
2. **AWS Access Key**: `AKIA`로 시작하는 20자 문자열
3. **JWT Secret**: `jwt_secret`, `secret_jwt` 등의 패턴
4. **API Key**: `api_key`, `key_api` 등의 패턴

## 사용법

### 기본 사용

```bash
./scripts/validate-compose-secrets.sh <docker-compose-file>
```

### 예시

```bash
# 단일 파일 검증
./scripts/validate-compose-secrets.sh docker-compose.yml

# 특정 환경 파일 검증
./scripts/validate-compose-secrets.sh docker/compose/docker-compose.ec2-dev.yml
```

### 종료 코드

- `0`: 검증 성공 (시크릿 없음)
- `1`: 검증 실패 (하드코딩된 시크릿 발견)

## 허용되는 패턴

환경변수 치환을 사용하는 경우 검증을 통과합니다:

### ✅ 올바른 사용 (통과)

```yaml
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}  # ✅ 환경변수 치환
      REDIS_PASSWORD: $REDIS_PASSWORD          # ✅ 환경변수 치환
```

### ❌ 잘못된 사용 (실패)

```yaml
version: '3.8'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: mysecretpassword123   # ❌ 하드코딩된 비밀번호
      AWS_ACCESS_KEY_ID: AKIAIOSFODNN7EXAMPLE  # ❌ 하드코딩된 AWS 키
```

## GitHub Actions 통합

CD 워크플로우에서 사용하는 예시:

```yaml
- name: Validate docker-compose for secrets
  run: |
    chmod +x scripts/validate-compose-secrets.sh
    ./scripts/validate-compose-secrets.sh docker/compose/docker-compose.ec2-dev.yml
```

## 테스트

속성 기반 테스트를 실행하여 검증 스크립트의 정확성을 확인할 수 있습니다:

```bash
./scripts/test-validate-compose-secrets.sh
```

테스트는 다음을 검증합니다:

- **Property 7**: 하드코딩된 시크릿 패턴 감지
- **Property 8**: 환경변수 치환 허용

테스트는 20개의 케이스를 실행하며, 랜덤 값을 생성하여 속성 기반 테스트를 수행합니다.

## 출력 예시

### 성공 케이스

```
🔍 Scanning docker-compose file for hardcoded secrets...
📄 File: docker-compose.yml

🔎 Checking for hardcoded passwords...
✅ No hardcoded passwords found

🔎 Checking for AWS Access Keys...
✅ No AWS Access Keys found

🔎 Checking for hardcoded JWT secrets...
✅ No hardcoded JWT secrets found

🔎 Checking for hardcoded API keys...
✅ No hardcoded API keys found

✅ Security validation passed!
   No hardcoded secrets found in docker-compose file.
```

### 실패 케이스

```
🔍 Scanning docker-compose file for hardcoded secrets...
📄 File: docker-compose.yml

🔎 Checking for hardcoded passwords...
6:      POSTGRES_PASSWORD: mysecretpassword123
❌ Found hardcoded password(s)
💡 Use environment variable substitution: ${PASSWORD}

❌ Security validation failed!
   Remove hardcoded secrets and use environment variable substitution.
```

## 관련 문서

- Requirements: 7.1, 7.2, 7.3, 7.5
- Design: Property 7, Property 8
- Feature: cd-workflow-improvement

## 문제 해결

### 오탐 (False Positive)

주석이나 예시 코드에서 오탐이 발생하는 경우, 해당 라인을 `#`로 주석 처리하면 검증을 통과합니다.

### 환경변수 치환이 감지되는 경우

다음 형식을 사용하세요:
- `${VARIABLE_NAME}` (권장)
- `$VARIABLE_NAME`

다음 형식은 감지됩니다:
- `password: mysecret` (하드코딩)
- `password: "mysecret"` (하드코딩)
