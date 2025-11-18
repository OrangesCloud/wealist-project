# 배포 스크립트 로그 마스킹

## 개요

배포 스크립트는 민감한 정보(비밀번호, API 키, JWT 시크릿 등)가 로그에 노출되지 않도록 자동으로 마스킹합니다.

## 마스킹되는 정보

다음 유형의 민감한 정보가 자동으로 `[REDACTED]`로 마스킹됩니다:

### Parameter Store 로드 시
- SecureString 타입의 모든 파라미터 값
- 파라미터 이름은 표시되지만, 실제 값은 표시되지 않음

### 환경변수 요약 출력 시
- `POSTGRES_SUPERUSER_PASSWORD`
- `USER_DB_PASSWORD`
- `BOARD_DB_PASSWORD`
- `REDIS_PASSWORD`
- `JWT_SECRET`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GRAFANA_ADMIN_PASSWORD`

### 에러 메시지
- Parameter Store 접근 실패 시 실제 파라미터 값은 표시되지 않음
- 트러블슈팅 가이드만 표시됨

## 로그 출력 예시

### Parameter 로드 (정상)
```
  ⏳ Loading parameter: aws_account_id
  ✅ Loaded: aws_account_id
```

### Secret 로드 (정상)
```
  ⏳ Loading secret parameter: db/postgres_superuser_password
  ✅ Loaded: db/postgres_superuser_password (SecureString)
```

### 환경변수 요약
```
📋 Loaded variables summary:
   - AWS_ACCOUNT_ID: 123456789012
   - POSTGRES_SUPERUSER: postgres
   - POSTGRES_SUPERUSER_PASSWORD: [REDACTED]
   - BOARD_DB_NAME: board_db
   - BOARD_DB_PASSWORD: [REDACTED]
   - JWT_SECRET: [REDACTED]
```

### 에러 발생 시
```
  ❌ Failed to load secret parameter: db/postgres_password
     Full path: /wealist/dev/db/postgres_password
  📋 Troubleshooting:
     - Check if parameter exists in Parameter Store
     - Verify IAM role has ssm:GetParameter permission
     - Confirm parameter type is SecureString
     - Verify KMS key permissions for decryption
     - Confirm AWS region is correct: ap-northeast-2
```

## 환경변수 자동 정리

배포 스크립트는 종료 시 (성공/실패 여부와 관계없이) 다음 환경변수를 자동으로 정리합니다:

```bash
cleanup_environment() {
  # 민감한 환경변수 unset
  unset POSTGRES_SUPERUSER_PASSWORD
  unset USER_DB_PASSWORD
  unset BOARD_DB_PASSWORD
  unset REDIS_PASSWORD
  unset JWT_SECRET
  unset GOOGLE_CLIENT_ID
  unset GOOGLE_CLIENT_SECRET
  unset GRAFANA_ADMIN_PASSWORD
}
```

이는 `trap EXIT` 명령어를 통해 스크립트 종료 시 자동으로 실행됩니다.

## 테스트

로그 마스킹 기능은 다음 테스트로 검증됩니다:

### 1. Property-Based Test
```bash
bash scripts/test-log-masking.sh
```

**테스트 내용:**
- 100개의 랜덤 시크릿 값이 로그에 노출되지 않는지 확인
- 특수문자가 포함된 시크릿도 올바르게 마스킹되는지 확인
- 환경변수 요약에서 시크릿이 `[REDACTED]`로 표시되는지 확인
- 에러 메시지에서 시크릿이 제거되는지 확인

### 2. 배포 스크립트 통합 테스트
```bash
bash scripts/test-deployment-log-masking.sh
```

**테스트 내용:**
- 실제 배포 스크립트의 `load_param()` 및 `load_secret()` 함수 테스트
- Parameter Store 로드 시 시크릿 값이 로그에 나타나지 않는지 확인
- 파라미터 이름과 성공 메시지는 정상적으로 표시되는지 확인

### 3. 환경변수 요약 테스트
```bash
bash scripts/test-env-summary-masking.sh
```

**테스트 내용:**
- 환경변수 요약 출력에서 모든 시크릿이 `[REDACTED]`로 표시되는지 확인
- 비밀이 아닌 값(계정 ID, 리전, DB 이름 등)은 정상적으로 표시되는지 확인
- 최소 8개의 `[REDACTED]` 마커가 있는지 확인

## 보안 고려사항

### GitHub Actions 자동 마스킹
GitHub Actions는 Secrets로 등록된 값을 자동으로 마스킹합니다. 하지만 다음 경우에는 추가 주의가 필요합니다:

1. **Base64 인코딩된 시크릿**: GitHub는 인코딩된 형태를 자동으로 마스킹하지 않을 수 있음
2. **동적으로 생성된 시크릿**: Parameter Store에서 로드한 값은 GitHub Secrets가 아니므로 수동 마스킹 필요
3. **에러 메시지**: AWS CLI 에러 출력에 시크릿이 포함될 수 있으므로 에러 출력을 표시하지 않음

### CloudWatch Logs
EC2에서 실행되는 배포 스크립트의 로그는 SSM Run Command를 통해 CloudWatch Logs에 저장될 수 있습니다. 로그 마스킹은 이러한 로그에도 적용됩니다.

## 문제 해결

### 시크릿이 로그에 노출되는 경우

1. **즉시 시크릿 교체**: Parameter Store에서 해당 시크릿을 새 값으로 업데이트
2. **로그 확인**: GitHub Actions 로그 및 CloudWatch Logs에서 노출된 로그 삭제 요청
3. **원인 분석**: 어떤 코드 경로에서 시크릿이 노출되었는지 확인
4. **테스트 실행**: `test-log-masking.sh`를 실행하여 마스킹이 올바르게 작동하는지 확인

### 디버깅이 어려운 경우

로그 마스킹으로 인해 디버깅이 어려운 경우:

1. **파라미터 이름 확인**: 파라미터 이름은 마스킹되지 않으므로 어떤 파라미터가 문제인지 확인 가능
2. **AWS Console 확인**: Parameter Store에서 직접 파라미터 값 확인
3. **로컬 테스트**: 로컬 환경에서 동일한 파라미터로 테스트
4. **IAM 권한 확인**: CloudTrail에서 Parameter Store 접근 시도 확인

## 관련 파일

- `.github/workflows/cd-dev-board-service.yml`: 배포 워크플로우 (로그 마스킹 구현)
- `scripts/test-log-masking.sh`: Property-based 테스트
- `scripts/test-deployment-log-masking.sh`: 배포 스크립트 통합 테스트
- `scripts/test-env-summary-masking.sh`: 환경변수 요약 테스트

## 요구사항 검증

이 구현은 다음 요구사항을 충족합니다:

- **Requirement 6.1**: 환경변수 요약 출력 시 모든 시크릿 값을 `[REDACTED]`로 표시
- **Requirement 6.2**: Parameter Store 로드 시 파라미터 이름과 로드 상태만 로깅, 실제 값은 로깅하지 않음
- **Requirement 6.3**: 에러 발생 시 파라미터 값을 에러 메시지에 포함하지 않음
- **Requirement 5.6**: 배포 완료 후 모든 민감한 환경변수를 자동으로 unset

## Property 6 검증

**Property 6: Sensitive information masking in logs**

*For any* secret value loaded from Parameter Store (SecureString type), all log outputs should display `[REDACTED]` instead of the actual value, including in success messages, error messages, and debug output.

이 속성은 다음 테스트로 검증됩니다:
- 100회 반복 테스트로 랜덤 시크릿 값이 로그에 노출되지 않음을 확인
- 특수문자가 포함된 시크릿도 올바르게 마스킹됨을 확인
- 실제 배포 스크립트 함수가 시크릿을 마스킹함을 확인
