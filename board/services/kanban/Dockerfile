FROM python:3.11-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev postgresql-dev

COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

FROM python:3.11-alpine

WORKDIR /app

RUN apk add --no-cache libpq curl

# 비루트 사용자 생성
RUN adduser -D appuser

# appuser의 홈 디렉토리로 패키지 복사 (핵심 수정)
COPY --from=builder /root/.local /home/appuser/.local

# appuser 소유권 설정
RUN chown -R appuser:appuser /home/appuser/.local

# PATH를 appuser의 로컬 bin으로 설정
ENV PATH=/home/appuser/.local/bin:$PATH

# 사용자 전환
USER appuser

# 애플리케이션 코드 복사
COPY . .

# 환경변수 기본값
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8000/health || exit 1

EXPOSE 8000

# Production mode
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
