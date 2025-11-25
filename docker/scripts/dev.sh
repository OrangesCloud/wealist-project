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

# 프로젝트 루트 디렉토리로 이동
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
# [⭐️ 핵심 추가] 공통 네트워크 검사 및 생성 (프론트엔드 연결용)
# =============================================================================
NETWORK_NAME="wealist-net"
EXPECTED_LABEL="com.docker.compose.network=${NETWORK_NAME}"

# 'up' 명령어 계열이거나 명령어가 생략되었을 때만 네트워크를 확인하고 생성합니다.
if [ "$1" == "up" ] || [ "$1" == "up-fg" ] || [ -z "$1" ]; then
    echo -e "${BLUE}🔗 공통 네트워크 ${NETWORK_NAME} 검사 중...${NC}"
    
    NETWORK_ID=$(docker network ls -q -f name=^${NETWORK_NAME}$)
    
    if [ -n "$NETWORK_ID" ]; then
        NETWORK_LABELS=$(docker network inspect $NETWORK_ID --format '{{json .Labels}}')
        
        # Docker Compose에서 기대하는 레이블을 포함하고 있는지 확인
        if echo "$NETWORK_LABELS" | grep -q "\"${EXPECTED_LABEL}\""; then
            echo -e "${BLUE}✅ 공통 네트워크 ${NETWORK_NAME} 이미 존재하고 레이블이 올바름.${NC}"
        else
            # 💡 [핵심 수정] 레이블이 잘못된 경우, 먼저 모든 컨테이너를 강제 중지/제거하고 네트워크를 삭제합니다.
            echo -e "${RED}❌ 공통 네트워크 ${NETWORK_NAME}의 레이블이 올바르지 않습니다.${NC}"
            
            # 네트워크에 연결된 모든 컨테이너 목록 조회
            CONTAINERS=$(docker network inspect ${NETWORK_NAME} --format '{{range .Containers}}{{.Name}} {{end}}')
            
            if [ -n "$CONTAINERS" ]; then
                echo -e "${YELLOW}   ⚠️ 네트워크에 연결된 컨테이너 (${CONTAINERS})를 강제 중지 및 제거합니다.${NC}"
                # 컨테이너 강제 중지 및 제거 (다른 프로젝트 컨테이너도 포함될 수 있으므로 주의)
                docker rm -f $CONTAINERS || true 
            fi
            
            echo -e "${YELLOW}   잘못된 레이블의 네트워크를 삭제 후 재생성합니다.${NC}"
            docker network rm ${NETWORK_NAME} || true # 삭제 실패해도 계속 진행 (이 단계에서는 대부분 성공해야 함)
            NETWORK_ID="" # 네트워크 ID 초기화하여 아래 로직으로 이동
        fi
    fi
    
    # 네트워크 ID가 비어있으면 (존재하지 않거나 방금 삭제된 경우) 생성
    if [ -z "$NETWORK_ID" ]; then
        echo -e "${GREEN}✅ 공통 네트워크 ${NETWORK_NAME} 생성.${NC}"
        
        # 💡 [핵심] 레이블을 명시적으로 부여하여 프론트엔드 Docker Compose가 인식하도록 합니다.
        docker network create \
            --driver bridge \
            --label ${EXPECTED_LABEL} \
            ${NETWORK_NAME} 
            
    fi
    echo ""
fi
# =============================================================================


# =============================================================================
# 로컬 환경 API Base URL 강제 오버라이드 (프론트엔드 컨테이너에서 필요)
# =============================================================================
# 이 설정은 프론트엔드 레포지토리의 .env 파일에만 필요하므로, 이 백엔드 스크립트에서는 제거합니다.
# 대신 프론트엔드 컨테이너가 http://nginx 로 통신하도록 합니다.
# =============================================================================

# 커맨드 처리
COMMAND=${1:-up}

case $COMMAND in
    up)
        echo -e "${BLUE}🚀 개발 환경을 백그라운드로 시작합니다...${NC}"
        # --build 옵션 제거 (up 시 자동으로 build 필요한지 검사함)
        docker compose $ENV_FILE_OPTION $COMPOSE_FILES up -d 
        echo -e "${GREEN}✅ 개발 환경이 시작되었습니다.${NC}"
        echo -e "${BLUE}📊 서비스 접속 정보:${NC}"
        echo "   - User API:    http://localhost:8080"
        echo "   - Board API:   http://localhost:8000"
        echo "   - NGINX/Frontend API Gateway: http://localhost:80 (프론트엔드 접속 주소)"
        echo "   - PostgreSQL:  localhost:5432"
        echo "   - Redis:       localhost:6379"
        echo "   - User API swagger:    http://localhost:8080/swagger-ui/index.html"
        echo "   - Board API swagger:   http://localhost:8000/swagger/index.html"
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
        echo -e "${RED}⚠️  모든 컨테이너, 볼륨, 이미지를 삭제합니다.${NC}"
        read -p "계속하시겠습니까? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker compose $ENV_FILE_OPTION $COMPOSE_FILES down -v --remove-orphans
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