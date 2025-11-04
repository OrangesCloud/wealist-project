#!/bin/bash

# docker_reset_and_check.sh
# Docker Compose 완전 재시작 및 DB 구조 확인 스크립트

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Configuration
COMPOSE_FILE="docker-compose.yml"
MAX_WAIT_TIME=120  # Maximum wait time in seconds
CHECK_INTERVAL=5   # Check interval in seconds

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE} Docker Compose 완전 재시작 & DB 구조 확인${NC}"
echo -e "${BLUE}============================================${NC}"

# Function to print step
print_step() {
    echo -e "\n${YELLOW}📋 STEP $1: $2${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Function to print info
print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Function to wait for service to be ready
wait_for_service() {
    local service_name=$1
    local health_check_cmd=$2
    local max_attempts=$((MAX_WAIT_TIME / CHECK_INTERVAL))
    local attempt=1

    print_info "$service_name 서비스 준비 대기 중..."

    while [ $attempt -le $max_attempts ]; do
        if eval "$health_check_cmd" > /dev/null 2>&1; then
            print_success "$service_name 서비스 준비 완료 (${attempt}/${max_attempts})"
            return 0
        fi

        echo -ne "\r${CYAN}대기 중... (${attempt}/${max_attempts}) - $((attempt * CHECK_INTERVAL))초 경과${NC}"
        sleep $CHECK_INTERVAL
        ((attempt++))
    done

    echo # New line
    print_error "$service_name 서비스가 준비되지 않았습니다 (${MAX_WAIT_TIME}초 초과)"
}

# Step 1: Stop and remove all containers, networks, and volumes
print_step "1" "기존 Docker 환경 완전 정리"
print_info "실행 중인 컨테이너 중지 중..."
docker-compose down --remove-orphans > /dev/null 2>&1 || true

print_info "볼륨 및 네트워크 정리 중..."
docker-compose down -v --remove-orphans > /dev/null 2>&1 || true

print_info "dangling 이미지 정리 중..."
docker system prune -f > /dev/null 2>&1 || true

print_info "사용하지 않는 볼륨 정리 중..."
docker volume prune -f > /dev/null 2>&1 || true

print_success "Docker 환경 완전 정리 완료"

# Step 2: Rebuild images
print_step "2" "Docker 이미지 재빌드"
print_info "캐시 없이 이미지 재빌드 중... (시간이 소요될 수 있습니다)"
docker-compose build --no-cache --parallel

print_success "Docker 이미지 재빌드 완료"

# Step 3: Start services
print_step "3" "서비스 순차적 시작"
print_info "데이터베이스 및 캐시 서비스 시작 중..."
docker-compose up -d postgres redis

# Wait for database
wait_for_service "PostgreSQL" "docker exec wealist-postgres pg_isready -U postgres"
wait_for_service "Redis" "docker exec wealist-redis redis-cli ping"

print_info "애플리케이션 서비스 시작 중..."
docker-compose up -d user-service board-service

# Wait for application services
wait_for_service "User Service" "curl -f http://localhost:8080/actuator/health"
wait_for_service "Board Service" "curl -f http://localhost:8000/health"

print_success "모든 서비스 시작 완료"

# Step 4: Show running services
print_step "4" "실행 중인 서비스 확인"
echo -e "${CYAN}"
docker-compose ps
echo -e "${NC}"

# Step 5: Check service logs for any errors
print_step "5" "서비스 상태 검사"
print_info "User Service 상태 확인..."
user_health=$(curl -s http://localhost:8080/actuator/health | grep -o '"status":"[^"]*"' | cut -d'"' -f4 || echo "UNKNOWN")
if [ "$user_health" = "UP" ]; then
    print_success "User Service: $user_health"
else
    echo -e "${YELLOW}⚠️  User Service: $user_health${NC}"
fi

print_info "Board Service 상태 확인..."
board_health=$(curl -s http://localhost:8000/health | grep -o '"status":"[^"]*"' | cut -d'"' -f4 || echo "UNKNOWN")
if [ "$board_health" = "ok" ]; then
    print_success "Board Service: $board_health"
else
    echo -e "${YELLOW}⚠️  Board Service: $board_health${NC}"
fi

# Step 6: Database structure inspection
print_step "6" "데이터베이스 구조 확인"

echo -e "\n${MAGENTA}==================== 마이그레이션 히스토리 ====================${NC}"
docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
SELECT
    version,
    dirty,
    CASE
        WHEN dirty THEN '❌ 실패'
        ELSE '✅ 성공'
    END as status
FROM schema_migrations
ORDER BY version;" 2>/dev/null || echo "마이그레이션 테이블을 찾을 수 없습니다."

echo -e "\n${MAGENTA}==================== 전체 테이블 목록 ====================${NC}"
docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
SELECT
    schemaname as 스키마,
    tablename as 테이블명,
    tableowner as 소유자
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;" 2>/dev/null

echo -e "\n${MAGENTA}==================== 테이블별 레코드 수 ====================${NC}"
docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
SELECT
    schemaname,
    tablename,
    n_tup_ins as 삽입된_레코드,
    n_tup_upd as 업데이트된_레코드,
    n_tup_del as 삭제된_레코드,
    n_live_tup as 현재_레코드_수
FROM pg_stat_user_tables
ORDER BY tablename;" 2>/dev/null || echo "통계 정보를 가져올 수 없습니다."

# Step 7: Detailed table structures
print_step "7" "주요 테이블 구조 상세 정보"

# Function to show table structure
show_table_structure() {
    local table_name=$1
    local description=$2

    echo -e "\n${CYAN}📋 $table_name ($description)${NC}"
    docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
    SELECT
        column_name as 컬럼명,
        data_type as 데이터타입,
        CASE
            WHEN character_maximum_length IS NOT NULL
            THEN data_type || '(' || character_maximum_length || ')'
            ELSE data_type
        END as 상세타입,
        CASE
            WHEN is_nullable = 'YES' THEN '✓'
            ELSE '✗'
        END as NULL허용,
        COALESCE(column_default, '') as 기본값
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = '$table_name'
    ORDER BY ordinal_position;" 2>/dev/null || echo "테이블 '$table_name'을 찾을 수 없습니다."

    # Show indexes
    echo -e "${CYAN}🔍 $table_name 인덱스:${NC}"
    docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
    SELECT
        indexname as 인덱스명,
        indexdef as 정의
    FROM pg_indexes
    WHERE schemaname = 'public' AND tablename = '$table_name'
    ORDER BY indexname;" 2>/dev/null
}

# Show core tables
show_table_structure "roles" "기본 역할"
show_table_structure "workspaces" "워크스페이스"
show_table_structure "workspace_members" "워크스페이스 멤버"
show_table_structure "workspace_join_requests" "워크스페이스 참여 신청"
show_table_structure "projects" "프로젝트"
show_table_structure "project_members" "프로젝트 멤버"
show_table_structure "custom_roles" "커스텀 역할"
show_table_structure "custom_stages" "커스텀 진행단계"
show_table_structure "custom_importance" "커스텀 중요도"
show_table_structure "kanbans" "칸반"
show_table_structure "kanban_roles" "칸반-역할 매핑"

# Show order management tables
echo -e "\n${MAGENTA}==================== 순서 관리 테이블 ====================${NC}"
show_table_structure "user_role_column_order" "사용자별 역할 컬럼 순서"
show_table_structure "user_stage_column_order" "사용자별 진행단계 컬럼 순서"
show_table_structure "user_kanban_order_in_role" "사용자별 역할내 칸반 순서"
show_table_structure "user_kanban_order_in_stage" "사용자별 진행단계내 칸반 순서"

# Step 8: Sample data check
print_step "8" "샘플 데이터 확인"

echo -e "\n${CYAN}📋 기본 역할 데이터:${NC}"
docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
SELECT
    name as 역할명,
    level as 권한레벨,
    description as 설명,
    created_at as 생성일시
FROM roles
ORDER BY level DESC;" 2>/dev/null

echo -e "\n${CYAN}📋 데이터베이스 연결 정보:${NC}"
docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
SELECT
    current_database() as 현재_데이터베이스,
    current_user as 현재_사용자,
    inet_server_addr() as 서버_주소,
    inet_server_port() as 서버_포트,
    version() as PostgreSQL_버전;" 2>/dev/null

# Step 9: Performance and size info
print_step "9" "데이터베이스 성능 및 크기 정보"

echo -e "\n${CYAN}📊 데이터베이스 크기:${NC}"
docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
SELECT
    pg_database.datname as 데이터베이스,
    pg_size_pretty(pg_database_size(pg_database.datname)) as 크기
FROM pg_database
WHERE pg_database.datname = 'wealist_kanban_db';" 2>/dev/null

echo -e "\n${CYAN}📊 테이블별 크기:${NC}"
docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -c "
SELECT
    schemaname,
    tablename as 테이블명,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as 전체크기,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as 테이블크기
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;" 2>/dev/null

# Step 10: Connection and configuration info
print_step "10" "연결 및 설정 정보"

echo -e "\n${CYAN}🔗 현재 데이터베이스 연결:${NC}"
docker exec wealist-postgres psql -U postgres -d postgres -c "
SELECT
    datname as 데이터베이스,
    usename as 사용자,
    client_addr as 클라이언트_IP,
    state as 상태,
    backend_start as 연결_시작시간
FROM pg_stat_activity
WHERE datname IS NOT NULL
ORDER BY backend_start DESC
LIMIT 10;" 2>/dev/null

echo -e "\n${CYAN}⚙️ Redis 정보:${NC}"
echo -e "Redis 상태: $(docker exec wealist-redis redis-cli ping 2>/dev/null || echo 'DISCONNECTED')"
echo -e "사용 중인 데이터베이스: $(docker exec wealist-redis redis-cli config get databases 2>/dev/null | tail -1 || echo 'N/A')"
echo -e "메모리 사용량: $(docker exec wealist-redis redis-cli info memory | grep used_memory_human | cut -d: -f2 | tr -d '\r' 2>/dev/null || echo 'N/A')"

# Final summary
echo -e "\n${BLUE}============================================${NC}"
echo -e "${BLUE} 환경 구축 완료 요약${NC}"
echo -e "${BLUE}============================================${NC}"

echo -e "${GREEN}✅ 서비스 상태:${NC}"
echo -e "   🐘 PostgreSQL: $(docker exec wealist-postgres pg_isready -U postgres 2>/dev/null | grep -o 'accepting connections' | head -1 || echo 'NOT READY')"
echo -e "   🔴 Redis: $(docker exec wealist-redis redis-cli ping 2>/dev/null || echo 'NOT READY')"
echo -e "   👤 User Service: $user_health (http://localhost:8080)"
echo -e "   📋 Board Service: $board_health (http://localhost:8000)"

echo -e "\n${GREEN}📊 데이터베이스 요약:${NC}"
table_count=$(docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs || echo "N/A")
migration_version=$(docker exec wealist-postgres psql -U kanban_service -d wealist_kanban_db -t -c "SELECT MAX(version) FROM schema_migrations;" 2>/dev/null | xargs || echo "N/A")

echo -e "   📋 총 테이블 수: $table_count개"
echo -e "   🔄 마이그레이션 버전: $migration_version"
echo -e "   🎯 기본 역할: 3개 (OWNER, ADMIN, MEMBER)"

echo -e "\n${YELLOW}🚀 다음 단계:${NC}"
echo -e "   테스트 실행: ${CYAN}./test_board_service.sh${NC}"
echo -e "   성능 테스트: ${CYAN}./performance_test_auto.sh${NC}"
echo -e "   로그 확인: ${CYAN}docker-compose logs -f board-service${NC}"
echo -e "   DB 접속: ${CYAN}docker exec -it wealist-postgres psql -U kanban_service -d wealist_kanban_db${NC}"

print_success "환경 구축이 성공적으로 완료되었습니다! 🎉"
