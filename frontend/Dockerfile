# 1단계: 빌드 스테이지 (Build Stage)
# pnpm 환경이 필요하므로 Node.js 이미지 사용
FROM node:20-alpine AS build

# 작업 디렉토리 설정
WORKDIR /app

# 종속성 파일 먼저 복사 (캐시 활용 극대화)
COPY package.json pnpm-lock.yaml ./

# pnpm 설치
RUN npm install -g pnpm

# 종속성 설치
# --frozen-lockfile을 사용해 안정성 확보
RUN pnpm install --frozen-lockfile

# 소스 코드 복사
COPY . .

# Vite 빌드 (VITE_API_BASE_URL은 CI/CD에서 주입되거나, 여기서는 기본값 사용)
# 로컬 테스트 시에는 하드코딩된 API 주소 또는 빌드 인자를 사용할 수 있음
RUN pnpm run build

# ----------------------------------------------------
# 2단계: 서빙 스테이지 (Serve Stage) - 최종 이미지
# Nginx는 정적 파일 서빙에 가장 가볍고 효율적
FROM nginx:alpine AS final

# Nginx 기본 설정 파일 삭제
RUN rm /etc/nginx/conf.d/default.conf

# Nginx 설정 파일 복사
# (정적 파일 서빙 및 SPA 라우팅을 위한 설정 파일)
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf

# 1단계에서 빌드된 결과물 복사
# /app/dist가 빌드 결과물이 있는 경로라고 가정합니다.
COPY --from=build /app/dist /usr/share/nginx/html

# 컨테이너 실행 시 80번 포트 노출
EXPOSE 80

# Nginx 웹 서버 실행 (기본 CMD 사용)
CMD ["nginx", "-g", "daemon off;"]