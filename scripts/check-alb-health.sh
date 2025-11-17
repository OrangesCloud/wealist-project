#!/bin/bash

# ALB Target Group Health 확인 스크립트
# AWS CLI를 사용하여 Target Group의 health 상태를 확인합니다.

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "ALB Target Group Health 확인"
echo "=========================================="
echo ""

# AWS CLI 설치 확인
if ! command -v aws &> /dev/null; then
    echo -e "${RED}Error: AWS CLI가 설치되어 있지 않습니다.${NC}"
    echo "AWS CLI 설치: https://aws.amazon.com/cli/"
    exit 1
fi

# AWS 자격 증명 확인
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}Error: AWS 자격 증명이 설정되어 있지 않습니다.${NC}"
    echo "AWS 자격 증명 설정: aws configure"
    exit 1
fi

echo -e "${GREEN}✓ AWS CLI 설정 확인 완료${NC}"
echo ""

# Target Group ARN 찾기
echo "Target Group 검색 중..."
echo ""

# User Service Target Group
USER_TG_ARN=$(aws elbv2 describe-target-groups \
    --query "TargetGroups[?contains(TargetGroupName, 'user-service')].TargetGroupArn" \
    --output text 2>/dev/null || echo "")

# Board Service Target Group
BOARD_TG_ARN=$(aws elbv2 describe-target-groups \
    --query "TargetGroups[?contains(TargetGroupName, 'board-service')].TargetGroupArn" \
    --output text 2>/dev/null || echo "")

# 함수: Target Group Health 확인
check_target_group_health() {
    local tg_arn=$1
    local service_name=$2
    
    if [ -z "$tg_arn" ]; then
        echo -e "${YELLOW}⚠ $service_name Target Group을 찾을 수 없습니다${NC}"
        echo "  Target Group 이름에 '$service_name'가 포함되어 있는지 확인하세요."
        return 1
    fi
    
    echo "=========================================="
    echo "$service_name Target Group"
    echo "=========================================="
    echo "ARN: $tg_arn"
    echo ""
    
    # Target Group 정보 가져오기
    tg_info=$(aws elbv2 describe-target-groups --target-group-arns "$tg_arn" --output json)
    
    # Health Check 설정 출력
    echo -e "${BLUE}Health Check 설정:${NC}"
    echo "$tg_info" | jq -r '.TargetGroups[0] | 
        "  Protocol: \(.HealthCheckProtocol)",
        "  Path: \(.HealthCheckPath)",
        "  Interval: \(.HealthCheckIntervalSeconds)s",
        "  Timeout: \(.HealthCheckTimeoutSeconds)s",
        "  Healthy Threshold: \(.HealthyThresholdCount)",
        "  Unhealthy Threshold: \(.UnhealthyThresholdCount)",
        "  Matcher: \(.Matcher.HttpCode)"'
    echo ""
    
    # Target Health 확인
    echo -e "${BLUE}Target Health 상태:${NC}"
    health_status=$(aws elbv2 describe-target-health --target-group-arn "$tg_arn" --output json)
    
    # 각 타겟의 상태 출력
    echo "$health_status" | jq -r '.TargetHealthDescriptions[] | 
        "  Target: \(.Target.Id):\(.Target.Port)",
        "  State: \(.TargetHealth.State)",
        "  Reason: \(.TargetHealth.Reason // "N/A")",
        "  Description: \(.TargetHealth.Description // "N/A")",
        ""'
    
    # Healthy 타겟 수 계산
    healthy_count=$(echo "$health_status" | jq '[.TargetHealthDescriptions[] | select(.TargetHealth.State == "healthy")] | length')
    total_count=$(echo "$health_status" | jq '.TargetHealthDescriptions | length')
    
    echo -e "  ${GREEN}Healthy: $healthy_count / $total_count${NC}"
    echo ""
    
    if [ "$healthy_count" -eq "$total_count" ] && [ "$total_count" -gt 0 ]; then
        echo -e "${GREEN}✓ 모든 타겟이 healthy 상태입니다${NC}"
        return 0
    elif [ "$healthy_count" -gt 0 ]; then
        echo -e "${YELLOW}⚠ 일부 타겟이 unhealthy 상태입니다${NC}"
        return 1
    else
        echo -e "${RED}✗ 모든 타겟이 unhealthy 상태입니다${NC}"
        return 1
    fi
}

# User Service Target Group 확인
user_service_status=0
if ! check_target_group_health "$USER_TG_ARN" "user-service"; then
    user_service_status=1
fi

echo ""

# Board Service Target Group 확인
board_service_status=0
if ! check_target_group_health "$BOARD_TG_ARN" "board-service"; then
    board_service_status=1
fi

echo ""

# ALB Listener Rules 확인 (선택사항)
echo "=========================================="
echo "ALB Listener Rules 확인"
echo "=========================================="
echo ""

# ALB ARN 찾기
ALB_ARN=$(aws elbv2 describe-load-balancers \
    --query "LoadBalancers[?contains(LoadBalancerName, 'wealist')].LoadBalancerArn" \
    --output text 2>/dev/null || echo "")

if [ -z "$ALB_ARN" ]; then
    echo -e "${YELLOW}⚠ ALB를 찾을 수 없습니다${NC}"
    echo "  LoadBalancer 이름에 'wealist'가 포함되어 있는지 확인하세요."
else
    echo "ALB ARN: $ALB_ARN"
    echo ""
    
    # Listener 찾기 (HTTPS:443)
    LISTENER_ARN=$(aws elbv2 describe-listeners \
        --load-balancer-arn "$ALB_ARN" \
        --query "Listeners[?Port==\`443\`].ListenerArn" \
        --output text 2>/dev/null || echo "")
    
    if [ -z "$LISTENER_ARN" ]; then
        echo -e "${YELLOW}⚠ HTTPS Listener를 찾을 수 없습니다${NC}"
    else
        echo "Listener ARN: $LISTENER_ARN"
        echo ""
        
        # Rules 가져오기
        echo -e "${BLUE}Listener Rules:${NC}"
        aws elbv2 describe-rules --listener-arn "$LISTENER_ARN" --output json | \
            jq -r '.Rules | sort_by(.Priority) | .[] | 
                "Priority: \(.Priority)",
                "  Conditions: \(.Conditions | map(
                    if .Field == "path-pattern" then
                        "Path: \(.Values | join(", "))"
                    else
                        "\(.Field): \(.Values | join(", "))"
                    end
                ) | join(", "))",
                "  Actions: \(.Actions | map(
                    if .Type == "forward" then
                        "Forward to \(.TargetGroupArn | split("/") | .[1])"
                    else
                        .Type
                    end
                ) | join(", "))",
                ""'
    fi
fi

echo ""

# 최종 결과
echo "=========================================="
echo "검증 결과 요약"
echo "=========================================="

if [ $user_service_status -eq 0 ] && [ $board_service_status -eq 0 ]; then
    echo -e "${GREEN}✓ 모든 Target Group이 정상 상태입니다${NC}"
    echo ""
    echo "다음 단계:"
    echo "1. 실제 API 호출 테스트: ./scripts/verify-alb-setup.sh"
    echo "2. 자세한 검증 가이드: docs/ALB_VERIFICATION_GUIDE.md"
    exit 0
else
    echo -e "${RED}✗ 일부 Target Group에 문제가 있습니다${NC}"
    echo ""
    echo "문제 해결:"
    echo "1. 서비스 로그 확인: docker logs user-service / docker logs board-service"
    echo "2. Health check 경로 직접 테스트"
    echo "3. 환경 변수 확인"
    echo "4. 자세한 가이드: docs/ALB_VERIFICATION_GUIDE.md"
    exit 1
fi
