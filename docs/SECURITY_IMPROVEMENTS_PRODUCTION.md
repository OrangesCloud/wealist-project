# 🔐 실무 수준 보안 개선 가이드

현재 구현된 보안 수준에서 **프로덕션 환경**으로 가기 위한 추가 보안 개선 사항

---

## 📊 현재 보안 수준 vs 프로덕션 권장 수준

| 보안 항목 | 현재 (Dev) | 프로덕션 권장 | 우선순위 |
|----------|-----------|-------------|---------|
| AWS 인증 방식 | Access Key (GitHub Secrets) | **OIDC** (임시 자격증명) | 🔴 High |
| 비밀 저장소 | Parameter Store | **Secrets Manager** (자동 로테이션) | 🟡 Medium |
| KMS 암호화 | AWS 관리형 키 | **고객 관리형 키 (CMK)** | 🟡 Medium |
| Terraform State | 로컬 파일 | **S3 Backend + 암호화** | 🔴 High |
| 배포 승인 | 자동 배포 | **Environment Protection Rules** | 🟡 Medium |
| 이미지 보안 | 스캔 없음 | **ECR 이미지 스캔** | 🟢 Low |
| 네트워크 | Public Subnet | **Private Subnet + VPC Endpoint** | 🟡 Medium |
| 감사 로그 | CloudTrail (기본) | **CloudWatch Alarms + 알림** | 🟢 Low |
| IAM 권한 | 기본 정책 | **최소 권한 원칙 강화** | 🔴 High |

---

## 🔴 High Priority (즉시 적용 권장)

### 1. OIDC 기반 GitHub Actions 인증 (★★★★★)

**현재 문제**:
```yaml
# ❌ 장기 자격증명을 GitHub Secrets에 저장
secrets:
  WEALIST_DEV_AWS_ACCESS_KEY_ID
  WEALIST_DEV_AWS_SECRET_ACCESS_KEY
```
- Access Key가 유출되면 모든 권한 탈취 가능
- 정기적 로테이션 어려움
- GitHub 해킹 시 위험

**개선 방법**:
```yaml
# ✅ OIDC로 임시 자격증명 자동 발급
permissions:
  id-token: write  # OIDC 토큰 생성 권한
  contents: read

- name: Configure AWS Credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: arn:aws:iam::290008131187:role/GitHubActionsOIDCRole
    aws-region: ap-northeast-2
    # Access Key 불필요!
```

**장점**:
- ✅ 장기 자격증명 불필요 (GitHub Secrets에 저장 안함)
- ✅ 매 배포마다 새로운 임시 토큰 (1시간 유효)
- ✅ 특정 Repository + Branch만 허용 가능
- ✅ 자동 만료로 보안 위험 최소화

**설정 방법**: `docs/terraform/oidc-github-actions.tf` 참고

---

### 2. Terraform State를 S3 Backend로 관리 (★★★★★)

**현재 문제**:
```bash
# ❌ 로컬에 terraform.tfstate 저장
docs/terraform/terraform.tfstate  # 실제 비밀번호가 평문으로 저장됨!
```
- State 파일에 모든 Parameter 값이 평문으로 저장
- 실수로 Git에 커밋될 위험
- 팀원 간 State 공유 어려움
- 동시 수정 시 충돌

**개선 방법**:
```hcl
# ✅ S3에 암호화 저장 + DynamoDB로 잠금
terraform {
  backend "s3" {
    bucket         = "wealist-terraform-state"
    key            = "parameter-store/dev/terraform.tfstate"
    region         = "ap-northeast-2"
    encrypt        = true               # S3 암호화
    kms_key_id     = "arn:aws:kms:..."  # KMS 암호화
    dynamodb_table = "terraform-locks"  # 동시 수정 방지
  }
}
```

**장점**:
- ✅ State 파일 암호화 저장
- ✅ 버저닝으로 롤백 가능
- ✅ 팀원과 안전하게 공유
- ✅ State locking으로 충돌 방지

**설정 방법**: `docs/terraform/s3-backend.tf` 참고

---

### 3. IAM 최소 권한 원칙 강화 (★★★★☆)

**현재 IAM Policy**:
```json
// ⚠️ 모든 Parameter 읽기 가능
"Resource": "arn:aws:ssm:ap-northeast-2:*:parameter/wealist/dev/*"
```

**개선된 IAM Policy**:
```json
{
  "Statement": [
    {
      "Sid": "ParameterStoreReadSpecific",
      "Effect": "Allow",
      "Action": ["ssm:GetParameter"],
      "Resource": [
        "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/postgres/superuser-password",
        "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/db/user-password",
        "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/db/board-password",
        "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/redis/password",
        "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/jwt/secret",
        "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/oauth/google-client-secret",
        "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/grafana/admin-password"
      ]
    },
    {
      "Sid": "KMSDecryptSpecific",
      "Effect": "Allow",
      "Action": ["kms:Decrypt"],
      "Resource": "arn:aws:kms:ap-northeast-2:290008131187:key/YOUR_KMS_KEY_ID",
      "Condition": {
        "StringEquals": {
          "kms:ViaService": "ssm.ap-northeast-2.amazonaws.com"
        }
      }
    },
    {
      "Sid": "ECRReadOnlySpecific",
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchGetImage",
        "ecr:GetDownloadUrlForLayer"
      ],
      "Resource": [
        "arn:aws:ecr:ap-northeast-2:290008131187:repository/wealist-dev-user-service",
        "arn:aws:ecr:ap-northeast-2:290008131187:repository/wealist-dev-board-service"
      ]
    }
  ]
}
```

**장점**:
- ✅ 정확히 필요한 Parameter만 접근 가능
- ✅ 특정 KMS 키만 사용
- ✅ 특정 ECR Repository만 접근
- ✅ Wildcard (*) 최소화

---

## 🟡 Medium Priority (단계적 적용)

### 4. AWS Secrets Manager + 자동 로테이션 (★★★★☆)

**Parameter Store vs Secrets Manager**:

| 기능 | Parameter Store | Secrets Manager |
|-----|----------------|-----------------|
| 비용 | 무료 (표준), $0.05/파라미터 (고급) | $0.40/비밀/월 + $0.05/10,000 API 호출 |
| 자동 로테이션 | ❌ 없음 | ✅ Lambda 통합 자동 로테이션 |
| RDS 통합 | ❌ 없음 | ✅ RDS 비밀번호 자동 관리 |
| 버전 관리 | ⚠️ 제한적 | ✅ 완전 지원 |
| 교차 리전 복제 | ❌ 없음 | ✅ 지원 |

**언제 Secrets Manager를 써야 하나?**
- ✅ DB 비밀번호 자동 로테이션 필요
- ✅ 프로덕션 환경
- ✅ 컴플라이언스 요구사항 (예: 90일마다 비밀번호 변경)
- ✅ 교차 리전 복제 필요

**Terraform 예제**:
```hcl
resource "aws_secretsmanager_secret" "db_password" {
  name = "wealist/prod/db/user-password"

  # 30일마다 자동 로테이션
  rotation_rules {
    automatically_after_days = 30
  }
}

resource "aws_secretsmanager_secret_rotation" "db_password" {
  secret_id           = aws_secretsmanager_secret.db_password.id
  rotation_lambda_arn = aws_lambda_function.rotate_secret.arn

  rotation_rules {
    automatically_after_days = 30
  }
}
```

**비용 비교 (8개 비밀)**:
- Parameter Store: **무료** (표준 tier)
- Secrets Manager: **$3.20/월** + API 호출 비용

**권장**: Dev/Staging은 Parameter Store, **Prod는 Secrets Manager**

---

### 5. KMS 고객 관리형 키 (CMK) (★★★☆☆)

**현재**:
```hcl
# AWS 관리형 KMS 키 사용 (기본)
resource "aws_ssm_parameter" "password" {
  type = "SecureString"  # AWS 관리형 키로 자동 암호화
}
```

**개선**:
```hcl
# 고객 관리형 KMS 키
resource "aws_kms_key" "wealist_secrets" {
  description             = "KMS key for wealist secrets"
  deletion_window_in_days = 30
  enable_key_rotation     = true  # 자동 키 로테이션 (1년)

  tags = {
    Project = "wealist"
    Environment = "prod"
  }
}

resource "aws_kms_alias" "wealist_secrets" {
  name          = "alias/wealist-secrets-prod"
  target_key_id = aws_kms_key.wealist_secrets.key_id
}

resource "aws_ssm_parameter" "password" {
  type   = "SecureString"
  kms_key_id = aws_kms_key.wealist_secrets.arn  # CMK 사용
}
```

**장점**:
- ✅ 키 사용 감사 로그 (CloudTrail)
- ✅ 세밀한 접근 제어 (누가 복호화 가능한지)
- ✅ 자동 키 로테이션
- ✅ 키 비활성화/삭제 제어

**비용**:
- KMS 키: **$1/월**
- API 호출: $0.03/10,000 요청

---

### 6. GitHub Environment Protection Rules (★★★★☆)

**현재**:
```yaml
# ⚠️ deploy-dev 브랜치에 push하면 자동 배포
on:
  push:
    branches: [deploy-dev]
```

**개선**:
```yaml
# ✅ 수동 승인 필요
jobs:
  deploy:
    environment:
      name: production
      url: https://wealist.com
    # 승인자: @team-lead, @devops-admin
```

**GitHub 설정** (Repository → Settings → Environments):
1. Environment 생성: `production`
2. **Required reviewers**: 팀장, DevOps 담당자
3. **Deployment branches**: `main` 브랜치만 허용
4. **Wait timer**: 5분 대기 후 배포 (긴급 중단 가능)
5. **Environment secrets**: Prod 전용 secrets 분리

**효과**:
- ✅ 실수로 배포 방지
- ✅ 배포 전 코드 리뷰 강제
- ✅ 배포 이력 추적
- ✅ 환경별 secrets 분리

---

### 7. VPC Private Subnet + VPC Endpoint (★★★☆☆)

**현재 네트워크 구조**:
```
EC2 (Public Subnet)
  ↓ Internet Gateway
  ↓ 인터넷 경유
AWS Services (ECR, SSM, Secrets Manager)
```
- ⚠️ 트래픽이 인터넷을 거침
- ⚠️ 22번 포트 열림 (SSH)

**개선된 구조**:
```
EC2 (Private Subnet)
  ↓ VPC Endpoint (PrivateLink)
  ↓ AWS 내부 네트워크
AWS Services (ECR, SSM, Secrets Manager)
```

**Terraform 예제**:
```hcl
# VPC Endpoint for SSM (Session Manager)
resource "aws_vpc_endpoint" "ssm" {
  vpc_id            = aws_vpc.main.id
  service_name      = "com.amazonaws.ap-northeast-2.ssm"
  vpc_endpoint_type = "Interface"

  subnet_ids         = [aws_subnet.private.id]
  security_group_ids = [aws_security_group.vpc_endpoint.id]

  private_dns_enabled = true
}

# VPC Endpoint for ECR
resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id            = aws_vpc.main.id
  service_name      = "com.amazonaws.ap-northeast-2.ecr.api"
  vpc_endpoint_type = "Interface"

  subnet_ids         = [aws_subnet.private.id]
  security_group_ids = [aws_security_group.vpc_endpoint.id]

  private_dns_enabled = true
}

resource "aws_vpc_endpoint" "ecr_dkr" {
  vpc_id            = aws_vpc.main.id
  service_name      = "com.amazonaws.ap-northeast-2.ecr.dkr"
  vpc_endpoint_type = "Interface"

  subnet_ids         = [aws_subnet.private.id]
  security_group_ids = [aws_security_group.vpc_endpoint.id]

  private_dns_enabled = true
}

# VPC Endpoint for S3 (Gateway type)
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.main.id
  service_name = "com.amazonaws.ap-northeast-2.s3"

  route_table_ids = [aws_route_table.private.id]
}
```

**장점**:
- ✅ 인터넷 게이트웨이 불필요
- ✅ 22번 포트 닫기 (SSH 불필요)
- ✅ 트래픽이 AWS 내부 네트워크만 사용
- ✅ 데이터 전송 비용 절감

**비용**:
- Interface Endpoint: **$0.01/시간** × 3개 = $21.6/월
- Gateway Endpoint (S3): **무료**

---

## 🟢 Low Priority (선택 사항)

### 8. ECR 이미지 스캔 (★★★☆☆)

**Terraform 설정**:
```hcl
resource "aws_ecr_repository" "user_service" {
  name                 = "wealist-dev-user-service"
  image_tag_mutability = "IMMUTABLE"  # 태그 덮어쓰기 방지

  # 이미지 스캔 활성화
  image_scanning_configuration {
    scan_on_push = true  # Push 시 자동 스캔
  }

  # 이미지 수명 주기
  lifecycle_policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 10 images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 10
      }
      action = {
        type = "expire"
      }
    }]
  })
}
```

**CI 워크플로우에서 스캔 결과 확인**:
```yaml
- name: Check ECR Scan Results
  run: |
    SCAN_STATUS=$(aws ecr describe-image-scan-findings \
      --repository-name wealist-dev-user-service \
      --image-id imageTag=latest \
      --query 'imageScanFindings.findingSeverityCounts' \
      --region ap-northeast-2)

    if echo "$SCAN_STATUS" | grep -q "CRITICAL"; then
      echo "❌ Critical vulnerabilities found!"
      exit 1
    fi
```

**장점**:
- ✅ CVE 취약점 자동 검사
- ✅ 심각도별 분류 (CRITICAL, HIGH, MEDIUM, LOW)
- ✅ 배포 전 차단 가능

---

### 9. CloudWatch Alarms + SNS 알림 (★★☆☆☆)

**Parameter Store 무단 접근 감지**:
```hcl
resource "aws_cloudwatch_log_metric_filter" "parameter_store_access" {
  name           = "wealist-parameter-store-access"
  log_group_name = "/aws/cloudtrail/wealist"

  pattern = <<PATTERN
{
  ($.eventName = GetParameter || $.eventName = GetParameters) &&
  $.requestParameters.name = "/wealist/prod/*" &&
  $.userIdentity.principalId != "EXPECTED_ROLE_ID"
}
PATTERN

  metric_transformation {
    name      = "UnauthorizedParameterAccess"
    namespace = "Wealist/Security"
    value     = "1"
  }
}

resource "aws_cloudwatch_metric_alarm" "unauthorized_access" {
  alarm_name          = "wealist-unauthorized-parameter-access"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "UnauthorizedParameterAccess"
  namespace           = "Wealist/Security"
  period              = "300"
  statistic           = "Sum"
  threshold           = "0"
  alarm_description   = "Unauthorized access to Parameter Store detected"

  alarm_actions = [aws_sns_topic.security_alerts.arn]
}

resource "aws_sns_topic" "security_alerts" {
  name = "wealist-security-alerts"
}

resource "aws_sns_topic_subscription" "security_email" {
  topic_arn = aws_sns_topic.security_alerts.arn
  protocol  = "email"
  endpoint  = "security-team@wealist.com"
}
```

**감지 가능한 이벤트**:
- ✅ 비정상 Parameter 접근
- ✅ 실패한 로그인 시도
- ✅ IAM 정책 변경
- ✅ ECR 이미지 삭제

---

## 📋 우선순위별 적용 로드맵

### Phase 1: 즉시 적용 (1-2주)
1. ✅ **OIDC 기반 인증** 전환
2. ✅ **S3 Backend** 설정 (Terraform State)
3. ✅ **IAM 최소 권한** 강화

→ **효과**: GitHub Secrets에서 장기 자격증명 제거, State 파일 보안 강화

---

### Phase 2: 단계적 적용 (1개월)
4. ✅ **Environment Protection Rules** 설정
5. ✅ **KMS CMK** 도입
6. ✅ **ECR 이미지 스캔** 활성화

→ **효과**: 배포 승인 프로세스, 암호화 키 제어, 취약점 검사

---

### Phase 3: 프로덕션 준비 (2-3개월)
7. ✅ **Secrets Manager + 자동 로테이션** (Prod만)
8. ✅ **VPC Private Subnet + VPC Endpoint**
9. ✅ **CloudWatch Alarms + SNS 알림**

→ **효과**: 완전한 프로덕션 보안 체계 구축

---

## 💰 비용 추정 (월별)

| 항목 | Dev | Prod |
|-----|-----|------|
| Parameter Store | 무료 | 무료 |
| Secrets Manager (8개) | - | $3.20 |
| KMS CMK | - | $1.00 |
| VPC Endpoint (4개) | - | $28.80 |
| CloudWatch Alarms | - | $0.50 |
| **합계** | **무료** | **$33.50/월** |

**참고**: Dev 환경은 현재대로 유지, Prod만 추가 보안 적용 권장

---

## 🎯 실무에서 가장 중요한 3가지

1. **OIDC 인증** - 장기 자격증명 제거 (보안 사고의 90%가 키 유출)
2. **최소 권한 원칙** - 필요한 만큼만 권한 부여
3. **Environment Protection Rules** - 실수 방지 + 승인 프로세스

이 3가지만 적용해도 **80%의 보안 위험**을 줄일 수 있습니다.

---

## 📚 참고 자료

- [AWS Security Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [GitHub Actions OIDC with AWS](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [AWS Secrets Manager vs Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-vs-secrets-manager.html)
- [Terraform AWS Best Practices](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
