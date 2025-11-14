# GitHub Secrets 설정 가이드

이 문서는 wealist 프로젝트의 CI/CD 파이프라인에 필요한 GitHub Secrets 설정 방법을 안내합니다.

## 목차
1. [GitHub Secrets란?](#github-secrets란)
2. [설정 방법](#설정-방법)
3. [필수 Secrets 목록](#필수-secrets-목록)
4. [Secrets 값 확인 방법](#secrets-값-확인-방법)

---

## GitHub Secrets란?

GitHub Secrets는 GitHub Actions 워크플로우에서 사용할 수 있는 암호화된 환경 변수입니다.
- AWS 인증 정보, 데이터베이스 비밀번호 등 민감한 정보를 안전하게 저장
- 코드에 하드코딩하지 않고 CI/CD 파이프라인에서 사용 가능
- Repository, Environment 단위로 관리 가능

---

## 설정 방법

### 1. GitHub Repository Settings 접속
```
GitHub Repository → Settings → Secrets and variables → Actions
```

### 2. Environment 생성 (권장)
```
Settings → Environments → New environment
- Environment name: development
```

### 3. Secrets 추가
```
New repository secret 또는 Environment secrets 에서 추가
- Name: SECRET 이름 (대문자, 언더스코어 사용)
- Value: 실제 값 입력
- Add secret 클릭
```

---

## 필수 Secrets 목록

### 📦 AWS 관련 Secrets

| Secret Name | 설명 | 예시 | 확인 방법 |
|------------|------|------|----------|
| `AWS_ACCOUNT_ID` | AWS 계정 ID (12자리) | `290008131187` | AWS Console 우측 상단 계정 메뉴 |
| `AWS_ACCESS_KEY_ID` | AWS IAM Access Key | `AKIAIOSFODNN7EXAMPLE` | IAM > Users > Security credentials |
| `AWS_SECRET_ACCESS_KEY` | AWS IAM Secret Access Key | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` | IAM > Users > Security credentials |

**필요한 IAM 권한:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": "*"
    }
  ]
}
```

---

### 🖥️ EC2 배포 관련 Secrets

| Secret Name | 설명 | 예시 | 확인 방법 |
|------------|------|------|----------|
| `EC2_HOST` | EC2 Public IP 또는 도메인 | `13.125.XXX.XXX` | AWS Console > EC2 > Instances |
| `EC2_SSH_PRIVATE_KEY` | EC2 SSH 접속용 Private Key | `-----BEGIN RSA PRIVATE KEY-----\n...` | Terraform 출력 또는 .pem 파일 내용 |

**EC2_SSH_PRIVATE_KEY 값 복사:**
```bash
# macOS/Linux
cat ~/.ssh/wealist-dev-ec2.pem | pbcopy

# 또는 파일 내용 직접 복사
cat ~/.ssh/wealist-dev-ec2.pem
```

**중요:** Private Key는 `-----BEGIN RSA PRIVATE KEY-----` 부터 `-----END RSA PRIVATE KEY-----` 까지 전체를 복사해야 합니다.

---

### 🗄️ Database 관련 Secrets

#### User Service Database

| Secret Name | 설명 | 예시 |
|------------|------|------|
| `USER_DB_NAME` | User Service DB 이름 | `wealist_user_db` |
| `USER_DB_USER` | User Service DB 사용자 | `wealist_user` |
| `USER_DB_PASSWORD` | User Service DB 비밀번호 | `your_secure_password_123` |

#### Board Service Database

| Secret Name | 설명 | 예시 |
|------------|------|------|
| `BOARD_DB_NAME` | Board Service DB 이름 | `wealist_board_db` |
| `BOARD_DB_USER` | Board Service DB 사용자 | `wealist_board` |
| `BOARD_DB_PASSWORD` | Board Service DB 비밀번호 | `your_secure_password_456` |

#### PostgreSQL Superuser

| Secret Name | 설명 | 예시 |
|------------|------|------|
| `POSTGRES_SUPERUSER` | PostgreSQL 관리자 계정 | `postgres` |
| `POSTGRES_SUPERUSER_PASSWORD` | PostgreSQL 관리자 비밀번호 | `your_postgres_password` |

---

### 🔐 Redis & JWT Secrets

| Secret Name | 설명 | 예시 |
|------------|------|------|
| `REDIS_PASSWORD` | Redis 비밀번호 | `your_redis_password` |
| `JWT_SECRET` | JWT 서명 비밀키 (64+ bytes) | `your_super_secret_jwt_key_at_least_64_bytes_for_hs512_algorithm` |

**JWT_SECRET 생성 방법:**
```bash
# OpenSSL 사용
openssl rand -base64 64

# 또는 온라인 도구 사용
# https://www.grc.com/passwords.htm
```

---

### 🔑 OAuth Secrets

| Secret Name | 설명 | 확인 방법 |
|------------|------|----------|
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID | [Google Cloud Console](https://console.cloud.google.com/apis/credentials) |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Client Secret | Google Cloud Console > Credentials |

**Google OAuth 설정:**
1. Google Cloud Console > APIs & Services > Credentials
2. Create Credentials > OAuth 2.0 Client ID
3. Application type: Web application
4. Authorized redirect URIs 추가:
   - `http://localhost:8080/login/oauth2/code/google` (로컬 개발용)
   - `http://<EC2_HOST>:8080/login/oauth2/code/google` (EC2 Dev용)

---

### 📊 Monitoring Secrets

| Secret Name | 설명 | 예시 |
|------------|------|------|
| `GRAFANA_ADMIN_PASSWORD` | Grafana 관리자 비밀번호 | `your_grafana_password` |

---

## Secrets 값 확인 방법

### AWS 계정 ID 확인
```bash
# AWS CLI 사용
aws sts get-caller-identity --query Account --output text

# 또는 AWS Console 우측 상단 계정 메뉴에서 확인
```

### ECR 주소 확인
```bash
# Terraform 출력에서 확인
cd infrastructure/terraform/dev
terraform output ecr_user_service_url
terraform output ecr_board_service_url

# 형식: {AWS_ACCOUNT_ID}.dkr.ecr.{REGION}.amazonaws.com/{REPOSITORY_NAME}
```

### EC2 Public IP 확인
```bash
# Terraform 출력
terraform output ec2_public_ip

# 또는 AWS Console
aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=wealist-dev-ec2" \
  --query 'Reservations[0].Instances[0].PublicIpAddress' \
  --output text
```

### SSH Private Key 확인
```bash
# Terraform으로 생성한 경우
terraform output -raw ec2_private_key > wealist-dev-ec2.pem

# 전체 내용 확인
cat wealist-dev-ec2.pem
```

---

## 설정 검증

### 1. Secrets 설정 완료 체크리스트

```
✅ AWS 관련 (3개)
  - AWS_ACCOUNT_ID
  - AWS_ACCESS_KEY_ID
  - AWS_SECRET_ACCESS_KEY

✅ EC2 배포 (2개)
  - EC2_HOST
  - EC2_SSH_PRIVATE_KEY

✅ Database (9개)
  - USER_DB_NAME
  - USER_DB_USER
  - USER_DB_PASSWORD
  - BOARD_DB_NAME
  - BOARD_DB_USER
  - BOARD_DB_PASSWORD
  - POSTGRES_SUPERUSER
  - POSTGRES_SUPERUSER_PASSWORD
  - REDIS_PASSWORD

✅ JWT & OAuth (3개)
  - JWT_SECRET
  - GOOGLE_CLIENT_ID
  - GOOGLE_CLIENT_SECRET

✅ Monitoring (1개)
  - GRAFANA_ADMIN_PASSWORD

총 18개 Secrets
```

### 2. CI/CD 테스트

Secrets 설정 후 CI/CD가 정상 동작하는지 확인:

```bash
# 1. User Service 변경 후 push
cd user-service
# 소스 수정
git add .
git commit -m "test: CI/CD test"
git push origin feature/cicd-dev-ec2-deploy

# 2. GitHub Actions 탭에서 워크플로우 확인
# - User Service CI - ECR 성공
# - Backend EC2 CD - ECR 성공

# 3. EC2에서 배포 확인
ssh ec2-user@<EC2_HOST>
docker ps  # 컨테이너 실행 확인
curl http://localhost:8080/actuator/health  # User Service health check
curl http://localhost:8000/health  # Board Service health check
```

---

## 보안 권장 사항

### 1. Secrets 관리
- ✅ 절대 코드에 하드코딩하지 말 것
- ✅ .env 파일은 절대 Git에 커밋하지 말 것 (`.gitignore`에 포함)
- ✅ 주기적으로 비밀번호 변경 (3-6개월)
- ✅ Production 환경은 별도의 Secrets 사용

### 2. AWS IAM 사용자
- ✅ 최소 권한 원칙 적용 (ECR, EC2 필요한 권한만)
- ✅ MFA(Multi-Factor Authentication) 활성화
- ✅ Access Key 주기적 로테이션

### 3. SSH Key
- ✅ Private Key는 절대 공개 저장소에 업로드하지 말 것
- ✅ Key 권한 설정: `chmod 400 wealist-dev-ec2.pem`
- ✅ EC2 Security Group에서 SSH 접근 IP 제한

### 4. JWT Secret
- ✅ 64 bytes 이상의 강력한 랜덤 문자열 사용
- ✅ User Service와 Board Service에 동일한 값 사용
- ✅ Production 환경은 다른 Secret 사용

---

## 문제 해결

### Q1. "Error: ECR login failed"
**원인:** AWS Credentials가 잘못되었거나 IAM 권한 부족

**해결:**
```bash
# AWS Credentials 확인
aws sts get-caller-identity

# ECR 로그인 테스트
aws ecr get-login-password --region ap-northeast-2 | \
  docker login --username AWS --password-stdin \
  {AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-2.amazonaws.com
```

### Q2. "Error: Permission denied (publickey)"
**원인:** SSH Private Key가 잘못되었거나 EC2 Security Group 설정 오류

**해결:**
```bash
# SSH Key 권한 확인
chmod 400 wealist-dev-ec2.pem

# SSH 연결 테스트
ssh -i wealist-dev-ec2.pem ec2-user@{EC2_HOST}

# Security Group에서 SSH(22) 포트 허용 확인
```

### Q3. "Error: Database connection failed"
**원인:** Database Secrets가 잘못되었거나 PostgreSQL 컨테이너 미실행

**해결:**
```bash
# EC2에서 PostgreSQL 컨테이너 확인
ssh ec2-user@{EC2_HOST}
docker ps | grep postgres

# 로그 확인
docker logs wealist-postgres

# .env 파일 확인
cat ~/wealist-deploy/.env
```

---

## 참고 자료

- [GitHub Actions Secrets 공식 문서](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [AWS ECR 인증 가이드](https://docs.aws.amazon.com/AmazonECR/latest/userguide/Registries.html)
- [SSH Key 관리 가이드](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)

---

## 관련 문서

- [EC2 Dev 배포 가이드](./EC2-DEV-DEPLOYMENT.md)
- [CI/CD 워크플로우 구조](../.github/workflows/README.md)
- [Docker Compose 가이드](../docker/README.md)
