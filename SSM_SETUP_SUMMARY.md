# SSM 및 CD 워크플로우 설정 완료 요약

## 환경 정보

### AWS 계정 정보
- **AWS Account ID**: `290008131187`
- **AWS Region**: `ap-northeast-2`
- **VPC ID**: `vpc-0ebd6645ca4194a59`
- **Subnet ID**: `subnet-03999d112592928c0`
- **Route Table ID**: `rtb-02fa893135e7708f6`
- **EC2 Instance ID**: `i-000f35cf67afa3c33`
- **EC2 Private IP**: `10.0.2.11`
- **EC2 Public IP**: `43.201.180.177`

### IAM 역할
- **IAM Role Name**: `wealist-dev-ec2-role`
- **Instance Profile Name**: `wealist-dev-ec2-profile`
- **Security Group ID**: `sg-02e9b3b9c23ae4573`

---

## 1. VPC 엔드포인트 생성 완료

### SSM 관련 엔드포인트 (3개)
```
vpce-00ffde510046f00fe - com.amazonaws.ap-northeast-2.ssm (available)
vpce-01166091b253c6375 - com.amazonaws.ap-northeast-2.ssmmessages (available)
vpce-04d8031dd56e7076b - com.amazonaws.ap-northeast-2.ec2messages (available)
```

### ECR 관련 엔드포인트 (2개)
```
com.amazonaws.ap-northeast-2.ecr.api (available, PrivateDns: true)
com.amazonaws.ap-northeast-2.ecr.dkr (available, PrivateDns: true)
```

### S3 Gateway 엔드포인트 (1개)
```
vpce-03c6f576a310113c5 - com.amazonaws.ap-northeast-2.s3 (available)
Route Table: rtb-02fa893135e7708f6
```

---

## 2. IAM 권한 설정

### EC2 역할에 연결된 정책

#### Managed Policies
1. **AmazonSSMManagedInstanceCore**
   - SSM Session Manager, Run Command 권한
   
2. **AmazonEC2ContainerRegistryReadOnly**
   - ECR 이미지 읽기 권한

#### Inline Policies

**custom-ssm-ec2-role**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:UpdateInstanceInformation",
        "ssmmessages:CreateControlChannel",
        "ssmmessages:CreateDataChannel",
        "ssmmessages:OpenControlChannel",
        "ssmmessages:OpenDataChannel"
      ],
      "Resource": "*",
      "Sid": "SSMAgentPermissions"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2messages:AcknowledgeMessage",
        "ec2messages:DeleteMessage",
        "ec2messages:FailMessage",
        "ec2messages:GetEndpoint",
        "ec2messages:GetMessages",
        "ec2messages:SendReply"
      ],
      "Resource": "*",
      "Sid": "EC2MessagesPermissions"
    }
  ]
}
```

**wealist-dev-ecr-access**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage"
      ],
      "Effect": "Allow",
      "Resource": "*",
      "Sid": "ECRAccess"
    }
  ]
}
```

**wealist-dev-parameter-store-read**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters",
        "ssm:GetParametersByPath"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:ssm:ap-northeast-2:290008131187:parameter/wealist/dev/*",
      "Sid": "ReadParameterStore"
    },
    {
      "Action": "kms:Decrypt",
      "Effect": "Allow",
      "Resource": "*",
      "Sid": "DecryptSecureStrings"
    }
  ]
}
```

---

## 3. Security Group 규칙

### Security Group: sg-02e9b3b9c23ae4573

#### Inbound Rules
```
Rule ID: sgr-07bfb2fb52074c3fb
- Type: Custom TCP
- Protocol: TCP
- Port: 8000
- Source: sg-099df9cf966a0dc34 (wealist-dev-alb-sg)

Rule ID: sgr-0ee3db9365e0882e8
- Type: Custom TCP
- Protocol: TCP
- Port: 8080
- Source: sg-099df9cf966a0dc34 (wealist-dev-alb-sg)

Rule ID: sgr-0d10cbe409de68459
- Type: HTTPS
- Protocol: TCP
- Port: 443
- Source: 10.0.2.0/24
```

#### Outbound Rules
```
Rule ID: sgr-0fef782a10a00fee1
- Type: All traffic
- Protocol: All
- Port: All
- Destination: 0.0.0.0/0
```

---

## 4. NACL 규칙 추가

### Network ACL for subnet-03999d112592928c0

#### Outbound Rules
```
Rule 90: HTTPS (443) - Allow - 0.0.0.0/0
Rule 100: TCP 1024-65535 - Allow - 10.0.3.0/24
Rule 32767: All - Deny - 0.0.0.0/0
```

#### Inbound Rules
```
Rule 90: HTTPS (443) - Allow - 0.0.0.0/0
Rule 100: TCP 8000 - Allow - 10.0.3.0/24
Rule 110: TCP 8080 - Allow - 10.0.3.0/24
Rule 32767: All - Deny - 0.0.0.0/0
```

---

## 5. Parameter Store 설정

### Parameter Store 경로: `/wealist/dev/`

#### 추가된 파라미터
```
/wealist/dev/aws_account_id = "290008131187" (String)
```

#### 기존 파라미터 (사용 중)
```
/wealist/dev/db/postgres_superuser (String)
/wealist/dev/db/postgres_superuser_password (SecureString)
/wealist/dev/db/user_db_name (String)
/wealist/dev/db/user_db_user (String)
/wealist/dev/db/user_db_password (SecureString)
/wealist/dev/db/board_db_name (String)
/wealist/dev/db/board_db_user (String)
/wealist/dev/db/board_db_password (SecureString)
/wealist/dev/cache/redis_password (SecureString)
/wealist/dev/jwt/jwt_secret (SecureString)
/wealist/dev/oauth/google_client_id (SecureString)
/wealist/dev/oauth/google-client-secret (SecureString)
/wealist/dev/monitor/grafana_admin_user (String)
/wealist/dev/monitor/grafana_admin_password (SecureString)
```

---

## 6. CD 워크플로우 수정 사항

### 파일: `.github/workflows/cd-dev-board-service.yml`

#### 수정된 파라미터 경로
```bash
# 변경 전 → 변경 후
board-db-name → db/board_db_name
board-db-user → db/board_db_user
board-db-password → db/board_db_password
postgres-superuser → db/postgres_superuser
postgres-superuser-password → db/postgres_superuser_password
user-db-name → db/user_db_name
user-db-user → db/user_db_user
user-db-password → db/user_db_password
redis-password → cache/redis_password
jwt-secret → jwt/jwt_secret
google-client-id → oauth/google_client_id
google-client-secret → oauth/google-client-secret
grafana-admin-user → monitor/grafana_admin_user
grafana-admin-password → monitor/grafana_admin_password
```

#### AWS Account ID 로드 방식
```bash
# 변경 후
export AWS_ACCOUNT_ID=$(load_param "aws_account_id")
```

---

## 7. SSM 접속 방법

### AWS CLI로 접속
```bash
aws ssm start-session \
  --target i-000f35cf67afa3c33 \
  --region ap-northeast-2
```

### AWS Console로 접속
1. EC2 Console → Instances
2. Instance ID: `i-000f35cf67afa3c33` 선택
3. **Connect** 버튼 클릭
4. **Session Manager** 탭 선택
5. **Connect** 클릭

---

## 8. 확인 명령어

### SSM 등록 확인
```bash
aws ssm describe-instance-information \
  --filters "Key=InstanceIds,Values=i-000f35cf67afa3c33" \
  --region ap-northeast-2
```

### VPC 엔드포인트 확인
```bash
aws ec2 describe-vpc-endpoints \
  --filters "Name=vpc-id,Values=vpc-0ebd6645ca4194a59" \
  --region ap-northeast-2 \
  --query 'VpcEndpoints[*].[ServiceName,State]' \
  --output table
```

### Docker 컨테이너 상태 확인 (SSM 세션 내부)
```bash
sudo docker ps
```

### 환경 변수 확인 (SSM 세션 내부)
```bash
sudo docker exec wealist-board-service env | grep -E "DB_|POSTGRES|AWS"
```

---

## 9. 남은 문제

### ECR 이미지 Pull 타임아웃
- **증상**: `dial tcp 52.219.148.94:443: i/o timeout`
- **원인**: S3에서 이미지 레이어 다운로드 시 타임아웃
- **임시 해결**: Docker 재시작 후 재시도
- **근본 해결**: S3 Gateway 엔드포인트 라우팅 확인 필요

### Board Service 환경 변수 문제
- **증상**: 컨테이너에 환경 변수가 전달되지 않음
- **원인**: Docker Compose 실행 시 환경 변수 미설정
- **해결 방법**: Parameter Store에서 환경 변수 로드 후 `sudo -E docker compose` 실행

---

## 10. 비용 절감 효과

### VPC 엔드포인트 사용으로 인한 절감
- **S3 Gateway 엔드포인트**: 무료
- **ECR 엔드포인트**: NAT Gateway 데이터 전송 비용 절감
- **SSM 엔드포인트**: NAT Gateway 데이터 전송 비용 절감

### 예상 절감액
- NAT Gateway 데이터 전송: Docker 이미지 pull, SSM 통신 등이 NAT를 거치지 않음
- 월 예상 절감: 수십 달러 (트래픽량에 따라 다름)

---

## 참고 문서

- [AWS Systems Manager Session Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager.html)
- [VPC Endpoints for ECR](https://docs.aws.amazon.com/AmazonECR/latest/userguide/vpc-endpoints.html)
- [S3 Gateway Endpoints](https://docs.aws.amazon.com/vpc/latest/privatelink/vpc-endpoints-s3.html)
