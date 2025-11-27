#!/bin/bash
# =============================================================================
# weAlist - Development Environment Startup Script
# =============================================================================
# 개발 환경을 시작하는 스크립트입니다.
#
# 사용법:
#   ./docker/scripts/dev.sh [command]
#
# Commands:
#   up         - 개발 환경 시작 (기본값)
#   down       - 개발 환경 중지
#   restart    - 개발 환경 재시작
#   logs       - 로그 확인
#   build      - 이미지 다시 빌드
#   clean      - 볼륨 포함 모두 삭제
# =============================================================================

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 프로젝트 루트 디렉토리로 이동 (ex. wealist-backend-repo/)
cd "$(dirname "$0")/../.."

# 환경변수 파일 확인
ENV_FILE="docker/env/.env.dev"
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}⚠️  환경변수 파일이 없습니다. 템플릿에서 생성합니다...${NC}"
    cp docker/env/.env.dev.example "$ENV_FILE"
    echo -e "${GREEN}✅ $ENV_FILE 파일이 생성되었습니다.${NC}"
    echo -e "${YELLOW}   필요한 값들을 수정한 후 다시 실행하세요.${NC}"
    exit 1
fi

# Docker Compose 파일 경로
COMPOSE_FILES="-f docker/compose/docker-compose.yml -f docker/compose/docker-compose.dev.yml"

# 환경변수 파일을 명시적으로 지정 (compose 파일 내 변수 치환용)
ENV_FILE_OPTION="--env-file $ENV_FILE"

# =============================================================================
# [⭐️ 핵심 수정] 공통 네트워크 검사 및 생성 (프론트엔드 연결용)
#   - 레이블 검사를 제거하고 존재 여부만 확인하여 프로세스 단순화 및 안정화
# =============================================================================
NETWORK_NAME="wealist-net"

# 'up' 명령어 계열이거나 명령어가 생략되었을 때만 네트워크를 확인하고 생성합니다.
if [ "$1" == "up" ] || [ "$1" == "up-fg" ] || [ -z "$1" ]; then
    echo -e "${BLUE}🔗 공통 네트워크 ${NETWORK_NAME} 검사 중...${NC}"
    
    # 네트워크 존재 여부 확인 및 생성 (if not exists)
    if docker network ls --filter name=^${NETWORK_NAME}$ --format "{{.Name}}" | grep -q ${NETWORK_NAME}; then
        echo -e "${GREEN}✅ 공통 네트워크 ${NETWORK_NAME}이(가) 이미 존재합니다. 재사용합니다.${NC}"
    else
        echo -e "${YELLOW}🚨 공통 네트워크 ${NETWORK_NAME}이(가) 존재하지 않습니다. 새로 생성합니다.${NC}"
        # --attachable: 다른 Docker Compose 파일의 서비스가 이 네트워크에 쉽게 연결될 수 있도록 허용
        docker network create --driver bridge --attachable ${NETWORK_NAME}
        echo -e "${GREEN}✅ 공통 네트워크 ${NETWORK_NAME} 생성 완료.${NC}"
    fi
    echo ""
fi
# =============================================================================


# 커맨드 처리
COMMAND=${1:-up}

case $COMMAND in
    up)
        echo -e "${BLUE}🚀 1/2 단계: 백엔드 개발 환경을 백그라운드로 시작합니다...${NC}"
        # --build 옵션 제거 (up 시 자동으로 build 필요한지 검사함)
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES up -d 
        echo -e "${GREEN}✅ 백엔드 서비스(DB, MinIO, API Gateway 등) 시작 완료.${NC}"

        # ---------------------------------------------------------------------
        # 프론트엔드 실행 안내 (프론트엔드 리포가 분리되어 있으므로 수동 실행이 필요함)
        # ---------------------------------------------------------------------
        echo -e "\n${BLUE}🚀 2/2 단계: 프론트엔드 개발 서버를 시작해주세요!${NC}"
        echo -e "${YELLOW}💡 프론트엔드 리포지토리로 이동하여 다음 명령을 실행하세요:${NC}"
        echo -e "   ${YELLOW}cd ../wealist-frontend-repo (경로 확인)${NC}"
        echo -e "   ${YELLOW}docker compose up -d${NC}"
        echo -e "   (또는 로컬에서 pnpm run dev)${NC}"
        # ---------------------------------------------------------------------

        echo -e "\n${BLUE}📊 서비스 접속 정보:${NC}"
        echo "   - NGINX/Frontend API Gateway: http://localhost:80 (API 호출 기본 주소)"
        echo "   - User API:    http://localhost:8080"
        echo "   - Board API:   http://localhost:8000"
        echo "   - MinIO Console: http://localhost:9001"
        echo -e ""
        echo -e "${BLUE}💡 로그 확인: ./docker/scripts/dev.sh logs${NC}"
        ;;

    up-fg)
        echo -e "${BLUE}🚀 개발 환경을 포그라운드로 시작합니다...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES up
        ;;

    down)
        echo -e "${YELLOW}⏹️  개발 환경을 중지합니다...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES down
        echo -e "${GREEN}✅ 개발 환경이 중지되었습니다.${NC}"
        ;;

    restart)
        echo -e "${YELLOW}🔄 개발 환경을 재시작합니다...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES restart
        echo -e "${GREEN}✅ 개발 환경이 재시작되었습니다.${NC}"
        ;;

    logs)
        SERVICE=${2:-}
        if [ -z "$SERVICE" ]; then
            docker compose $ENV_FILE_OPTION $COMPOSE_FILES logs -f
        else
            docker compose $ENV_FILE_OPTION $COMPOSE_FILES logs -f "$SERVICE"
        fi
        ;;

    build)
        echo -e "${BLUE}🔨 이미지를 다시 빌드합니다...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES build --no-cache
        echo -e "${GREEN}✅ 빌드가 완료되었습니다.${NC}"
        ;;

    rebuild)
        echo -e "${BLUE}🔨 이미지를 다시 빌드하고 시작합니다...${NC}"
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES up -d --build
        echo -e "${GREEN}✅ 빌드 및 시작이 완료되었습니다.${NC}"
        ;;

    clean)
        echo -e "${RED}⚠️  모든 백엔드 컨테이너, 볼륨, 이미지를 삭제합니다.${NC}"
        read -p "계속하시겠습니까? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker compose $ENV_FILE_OPTION $COMPOSE_FILES down -v --remove-orphans
            
            echo -e "${YELLOW}🔗 공통 네트워크 ${NETWORK_NAME}를 삭제합니다. (프론트엔드 컨테이너가 중지되었는지 확인하세요.)${NC}"
            docker network rm ${NETWORK_NAME} || true
            
            echo -e "${GREEN}✅ 정리가 완료되었습니다.${NC}"
        else
            echo -e "${YELLOW}취소되었습니다.${NC}"
        fi
        ;;

    ps)
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES ps
        ;;

    exec)
        SERVICE=${2:-user-service}
        SHELL=${3:-bash}
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES exec "$SERVICE" "$SHELL"
        ;;

    *)
        echo -e "${RED}❌ 알 수 없는 명령어: $COMMAND${NC}"
        echo ""
        echo "사용 가능한 명령어:"
        echo "  up         - 개발 환경 시작 (백그라운드)"
        echo "  up-fg      - 개발 환경 시작 (포그라운드)"
        echo "  down       - 개발 환경 중지"
        echo "  restart    - 개발 환경 재시작"
        echo "  logs       - 로그 확인 (logs [service])"
        echo "  build      - 이미지 다시 빌드"
        echo "  rebuild    - 빌드 후 시작"
        echo "  clean      - 모두 삭제 (볼륨 포함)"
        echo "  ps         - 실행 중인 서비스 확인"
        echo "  exec       - 컨테이너 접속 (exec [service] [shell])"
        exit 1
        ;;
esac