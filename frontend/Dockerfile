# 1. 베이스 이미지 선택
FROM node:18-alpine

# 2. 작업 디렉토리 설정
WORKDIR /app

# 3. 의존성 파일 복사
COPY package.json .
# package-lock.json이 있다면 함께 복사
# COPY package-lock.json .

# 4. 의존성 설치
RUN npm install

# 5. 소스코드 복사
COPY . .

# 6. React 개발 서버 실행
CMD ["npm", "start"]