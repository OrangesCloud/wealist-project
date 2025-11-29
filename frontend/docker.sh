#!/bin/bash
# 파일명: frontend/docker.sh

# ==========================================
# 0. 변수 설정
# ==========================================
COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="wealist-frontend-dev"

# 현재 디렉토리(.env 파일이 있는)로 이동하여 스크립트를 실행하도록 설정
# (./frontend/docker-compose.yml이 기준이 됨)
cd "$(dirname "$0")"

# ==========================================
# 1. 메인 함수
# ==========================================
main() {
    COMMAND=$1
    shift # 첫 번째 인자(command)를 제외하고 나머지 인자를 shift

    case "$COMMAND" in
        up)
            echo "🚀 WeAlist Frontend 개발 환경을 백그라운드로 시작합니다..."
            # -d: 백그라운드 실행
            # --wait: 서비스가 Healthy/Exited 상태가 될 때까지 기다림
            docker compose -f ${COMPOSE_FILE} -p ${PROJECT_NAME} up -d --wait
            
            if [ $? -eq 0 ]; then
                echo "✅ 프론트엔드 컨테이너가 성공적으로 실행되었습니다."
                echo ""
                echo "📊 접속 정보: http://localhost:3000"
                echo "💡 로그 확인: ./docker.sh logs"
            else
                echo "🚨 Error: 프론트엔드 컨테이너 시작에 실패했습니다."
                exit 1
            fi
            ;;

        down)
            echo "🛑 WeAlist Frontend 개발 환경을 종료합니다..."
            # -v: 볼륨도 함께 삭제 (선택 사항)
            docker compose -f ${COMPOSE_FILE} -p ${PROJECT_NAME} down
            echo "✅ 환경 종료 완료."
            ;;

        logs)
            echo "📖 프론트엔드 컨테이너 로그 출력 (Ctrl+C로 종료)..."
            docker compose -f ${COMPOSE_FILE} -p ${PROJECT_NAME} logs -f
            ;;

        *)
            echo "Usage: $0 {up|down|logs}"
            exit 1
            ;;
    esac
}

main "$@"