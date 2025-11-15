# 🔐 보안 가이드 문서 모음

wealist 프로젝트의 보안 관련 문서 및 설정 가이드

---

## 📚 문서 구조

### 1. 현재 구현된 보안 (Dev 환경)

| 문서 | 설명 | 용도 |
|-----|------|------|
| **[SECURE_DEPLOYMENT_SETUP.md](./SECURE_DEPLOYMENT_SETUP.md)** | Parameter Store 기반 배포 (AWS CLI) | 수동 설정 가이드 |
| **[TERRAFORM_PARAMETER_STORE_SETUP.md](./TERRAFORM_PARAMETER_STORE_SETUP.md)** | Parameter Store 관리 (Terraform) | **권장** IaC 방식 |
| **[EC2_IAM_POLICY.json](./EC2_IAM_POLICY.json)** | EC2 IAM 정책 (기본) | 현재 사용 중 |

**현재 보안 수준**:
- ✅ Parameter Store에 비밀 저장
- ✅ SSM을 통한 EC2 접속 (SSH 불필요)
- ✅ GitHub Secrets 최소화
- ⚠️ 장기 AWS 자격증명 사용 (Access Key)

---

### 2. 프로덕션 권장 보안

| 문서 | 설명 | 우선순위 |
|-----|------|---------|
| **[SECURITY_IMPROVEMENTS_PRODUCTION.md](./SECURITY_IMPROVEMENTS_PRODUCTION.md)** | 실무 수준 보안 개선 가이드 | ⭐⭐⭐⭐⭐ |
| **[terraform/oidc-github-actions.tf](./terraform/oidc-github-actions.tf)** | OIDC 기반 인증 설정 | 🔴 High |
| **[terraform/s3-backend.tf](./terraform/s3-backend.tf)** | Terraform State 보안 관리 | 🔴 High |
| **[workflows-examples/dev-backend-deploy-oidc.yml](./workflows-examples/dev-backend-deploy-oidc.yml)** | OIDC 워크플로우 예제 | 🔴 High |
| **[EC2_IAM_POLICY_STRICT.json](./EC2_IAM_POLICY_STRICT.json)** | 최소 권한 IAM 정책 | 🔴 High |

**프로덕션 보안 수준**:
- ✅ OIDC 임시 자격증명 (Access Key 제거)
- ✅ Terraform State S3 암호화 저장
- ✅ 최소 권한 원칙 강화
- ✅ KMS CMK 암호화
- ✅ Environment Protection Rules
- ✅ VPC Private Subnet + VPC Endpoint
- ✅ ECR 이미지 스캔
- ✅ CloudWatch Alarms

---

## 🚀 빠른 시작

### Dev 환경 (현재)

**Terraform으로 Parameter Store 설정** (권장):
```bash
cd docs/terraform

# 1. 변수 파일 생성
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars  # 실제 비밀번호 입력

# 2. Terraform 초기화 및 배포
terraform init
terraform plan
terraform apply
```

**또는 AWS CLI로 수동 설정**:
```bash
# docs/SECURE_DEPLOYMENT_SETUP.md 참고
aws ssm put-parameter --name "/wealist/dev/postgres/superuser-password" ...
```

---

### Prod 환경으로 전환 (권장)

**Phase 1: OIDC 인증 전환** (🔴 High Priority)
```bash
# 1. OIDC Provider 생성
cd docs/terraform
terraform apply -target=aws_iam_openid_connect_provider.github_actions

# 2. IAM Role 생성
terraform apply -target=aws_iam_role.github_actions_deploy

# 3. Role ARN 확인
terraform output github_actions_role_arn

# 4. GitHub Actions 워크플로우 업데이트
# .github/workflows/dev-backend-deploy-secure.yml을
# docs/workflows-examples/dev-backend-deploy-oidc.yml 참고하여 수정

# 5. GitHub Secrets에서 Access Key 삭제
# WEALIST_DEV_AWS_ACCESS_KEY_ID (삭제)
# WEALIST_DEV_AWS_SECRET_ACCESS_KEY (삭제)
```

**Phase 2: Terraform State 보안 강화** (🔴 High Priority)
```bash
# 1. S3 Backend 리소스 생성
terraform apply -target=aws_s3_bucket.terraform_state -target=aws_dynamodb_table.terraform_locks

# 2. parameter-store.tf에 backend 블록 추가
# terraform {
#   backend "s3" {
#     bucket = "wealist-terraform-state"
#     key    = "parameter-store/dev/terraform.tfstate"
#     region = "ap-northeast-2"
#     encrypt = true
#     dynamodb_table = "wealist-terraform-locks"
#   }
# }

# 3. State 마이그레이션
terraform init -migrate-state
```

**Phase 3: IAM 최소 권한 강화** (🔴 High Priority)
```bash
# EC2 IAM Role 업데이트
aws iam put-role-policy \
  --role-name WealistEC2Role \
  --policy-name WealistEC2StrictPolicy \
  --policy-document file://docs/EC2_IAM_POLICY_STRICT.json
```

---

## 📊 보안 수준 비교

### 현재 (Dev)
```
보안 점수: ⭐⭐⭐⭐ (4/5)

장점:
✅ Parameter Store 사용
✅ SSM 접속
✅ GitHub Secrets 최소화

개선 필요:
⚠️ 장기 AWS 자격증명
⚠️ Terraform State 로컬 저장
⚠️ Wildcard IAM 권한
```

### 프로덕션 권장
```
보안 점수: ⭐⭐⭐⭐⭐ (5/5)

추가 장점:
✅ OIDC 임시 자격증명
✅ S3 암호화 State 저장
✅ 최소 권한 IAM 정책
✅ KMS CMK 암호화
✅ 배포 승인 프로세스
✅ VPC 네트워크 격리
✅ 이미지 취약점 스캔
```

---

## 🎯 우선순위별 적용 가이드

### 즉시 적용 (1-2주) - 보안 위험 80% 감소

1. **OIDC 인증 전환**
   - 문서: `terraform/oidc-github-actions.tf`
   - 효과: 장기 자격증명 제거

2. **S3 Backend**
   - 문서: `terraform/s3-backend.tf`
   - 효과: State 파일 보안 강화

3. **IAM 최소 권한**
   - 문서: `EC2_IAM_POLICY_STRICT.json`
   - 효과: 권한 남용 방지

### 단계적 적용 (1개월)

4. **Environment Protection Rules**
   - GitHub Repository Settings
   - 효과: 배포 승인 프로세스

5. **KMS CMK**
   - Terraform 설정
   - 효과: 암호화 키 제어

6. **ECR 이미지 스캔**
   - Terraform 설정
   - 효과: 취약점 자동 검사

### 프로덕션 준비 (2-3개월)

7. **Secrets Manager** (Prod 전용)
   - 자동 비밀번호 로테이션
   - 월 $3.20 추가 비용

8. **VPC Private Subnet**
   - 네트워크 격리
   - 월 $28.80 추가 비용 (VPC Endpoint)

9. **CloudWatch Alarms**
   - 보안 이벤트 모니터링

---

## 💰 비용 영향

| 구성 | Dev | Prod |
|-----|-----|------|
| Parameter Store | 무료 | 무료 |
| **OIDC** | **무료** | **무료** |
| **S3 Backend** | **~$0.05/월** | **~$0.05/월** |
| Secrets Manager | - | $3.20/월 |
| KMS CMK | - | $1.00/월 |
| VPC Endpoint | - | $28.80/월 |
| CloudWatch Alarms | - | $0.50/월 |
| **합계** | **~$0.05/월** | **~$33.55/월** |

**Phase 1 (OIDC + S3 Backend)만 적용해도**:
- 비용: 거의 무료 (~$0.05/월)
- 보안 개선: 80%
- 구현 시간: 1-2주

---

## 🔍 자세한 내용

각 보안 개선 사항의 자세한 설명, 코드 예제, 적용 방법은 다음 문서를 참고하세요:

**[SECURITY_IMPROVEMENTS_PRODUCTION.md](./SECURITY_IMPROVEMENTS_PRODUCTION.md)**

이 문서에는 다음이 포함됩니다:
- 각 개선 사항의 기술적 설명
- Before/After 코드 비교
- 단계별 적용 방법
- 비용 분석
- 실무 Best Practices
- 문제 해결 가이드

---

## 📞 추가 질문

- **"어디서부터 시작해야 하나요?"** → OIDC 인증 전환 (가장 효과적)
- **"비용이 얼마나 드나요?"** → Phase 1은 거의 무료, Prod 전체는 ~$33/월
- **"언제 Secrets Manager를 써야 하나요?"** → Prod 환경 + 자동 로테이션 필요 시
- **"VPC Endpoint가 필수인가요?"** → 네트워크 격리가 필요한 Prod 환경에서 권장
- **"Dev 환경도 전부 적용해야 하나요?"** → Phase 1만 적용 권장, 나머지는 Prod만

---

## 📝 체크리스트

### Dev 환경 보안 체크리스트
- [ ] Parameter Store 설정 완료 (Terraform)
- [ ] EC2 IAM Role 설정
- [ ] SSM 접속 가능
- [ ] GitHub Secrets 최소화 (AWS credentials만)
- [ ] `.gitignore`에 terraform.tfvars 추가

### Prod 환경 보안 체크리스트
- [ ] OIDC 인증 전환
- [ ] S3 Backend 설정
- [ ] IAM 최소 권한 정책 적용
- [ ] Environment Protection Rules 설정
- [ ] KMS CMK 사용
- [ ] ECR 이미지 스캔 활성화
- [ ] Secrets Manager 설정 (선택)
- [ ] VPC Private Subnet 배포 (선택)
- [ ] CloudWatch Alarms 설정 (선택)

---

## 🚨 보안 사고 발생 시

1. **AWS Access Key 유출 의심**
   - IAM Console에서 즉시 키 비활성화
   - CloudTrail 로그 확인
   - OIDC로 전환되어 있다면 영향 없음

2. **Parameter Store 무단 접근**
   - CloudTrail에서 접근 로그 확인
   - IAM 정책 재검토
   - 비밀번호 즉시 로테이션

3. **GitHub Actions 워크플로우 변경 감지**
   - PR 리뷰 프로세스 강화
   - CODEOWNERS 파일 추가
   - Branch Protection Rules 설정

---

**문서 버전**: 1.0
**마지막 업데이트**: 2025-01-15
**작성자**: Claude Code
