#!/bin/bash
# =============================================================================
# weAlist - Monitoring Stack Management Script
# =============================================================================
# 모니터링 스택(Prometheus, Grafana, Exporters)을 관리하는 스크립트입니다.
#
# 사용법:
#   ./docker/scripts/monitoring.sh [command] [env]
#
# Commands:
#   up         - 모니터링 스택 시작
#   down       - 모니터링 스택 중지
#   restart    - 모니터링 스택 재시작
#   logs       - 로그 확인
#
# Environment:
#   dev        - 개발 환경과 함께 시작 (기본값)
#   prod       - 프로덕션 환경과 함께 시작
# =============================================================================

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 프로젝트 루트 디렉토리로 이동
cd "$(dirname "$0")/../.."

# 환경 결정
ENV=${2:-dev}
if [ "$ENV" = "prod" ]; then
    ENV_FILE="docker/env/.env.prod"
    COMPOSE_FILES="-f docker/compose/docker-compose.yml -f docker/compose/docker-compose.prod.yml -f docker/compose/docker-compose.monitoring.yml"
else
    ENV_FILE="docker/env/.env.dev"
    COMPOSE_FILES="-f docker/compose/docker-compose.yml -f docker/compose/docker-compose.dev.yml -f docker/compose/docker-compose.monitoring.yml"
fi

# 환경변수 파일 확인
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ 환경변수 파일이 없습니다: $ENV_FILE${NC}"
    exit 1
fi

# 환경변수 파일을 명시적으로 지정 (compose 파일 내 변수 치환용)
ENV_FILE_OPTION="--env-file $ENV_FILE"

# 커맨드 처리
COMMAND=${1:-up}

case $COMMAND in
    up)
        echo -e "${BLUE}📊 모니터링 스택을 시작합니다 (환경: $ENV)...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES up -d prometheus grafana loki promtail redis-exporter postgres-exporter node-exporter
        echo -e "${GREEN}✅ 모니터링 스택이 시작되었습니다.${NC}"
        echo ""
        echo -e "${BLUE}📊 모니터링 서비스 접속 정보:${NC}"
        echo -e "   - Prometheus:  http://localhost:9090"
        echo -e "   - Grafana:     http://localhost:3001"
        echo -e "   - Loki:        http://localhost:3100"
        echo -e "   - Grafana 기본 계정: admin / admin"
        echo ""
        echo -e "${YELLOW}💡 Grafana에서 Prometheus, Loki 데이터소스를 추가하세요:${NC}"
        echo -e "   - Prometheus URL: http://prometheus:9090"
        echo -e "   - Loki URL: http://loki:3100"
        ;;

    down)
        echo -e "${YELLOW}⏹️  모니터링 스택을 중지합니다...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES stop prometheus grafana loki promtail redis-exporter postgres-exporter node-exporter
        echo -e "${GREEN}✅ 모니터링 스택이 중지되었습니다.${NC}"
        ;;

    restart)
        echo -e "${YELLOW}🔄 모니터링 스택을 재시작합니다...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES restart prometheus grafana loki promtail redis-exporter postgres-exporter node-exporter
        echo -e "${GREEN}✅ 모니터링 스택이 재시작되었습니다.${NC}"
        ;;

    logs)
        SERVICE=${3:-prometheus}
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES logs -f "$SERVICE"
        ;;

    status)
        echo -e "${BLUE}📊 모니터링 서비스 상태:${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES ps prometheus grafana loki promtail redis-exporter postgres-exporter node-exporter
        ;;

    *)
        echo -e "${RED}❌ 알 수 없는 명령어: $COMMAND${NC}"
        echo ""
        echo "사용 가능한 명령어:"
        echo "  up [dev|prod]     - 모니터링 스택 시작"
        echo "  down [dev|prod]   - 모니터링 스택 중지"
        echo "  restart [dev|prod] - 모니터링 스택 재시작"
        echo "  logs [dev|prod] [service] - 로그 확인"
        echo "  status [dev|prod] - 서비스 상태 확인"
        exit 1
        ;;
esac
