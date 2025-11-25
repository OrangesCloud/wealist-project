# syntax=docker/dockerfile:1
FROM gradle:8-jdk21-alpine AS builder

WORKDIR /app

# 의존성 캐싱을 위한 파일들
COPY build.gradle settings.gradle gradle.properties* ./
COPY gradle ./gradle

# GitHub Actions 캐시 디렉토리 활용
RUN --mount=type=cache,target=/home/gradle/.gradle,sharing=locked \
    mkdir -p src/main/java && \
    echo "class Temp{}" > src/main/java/Temp.java && \
    gradle build --no-daemon -x test || true && \
    rm -rf src/

# 실제 소스 빌드
COPY src ./src
RUN --mount=type=cache,target=/home/gradle/.gradle,sharing=locked \
    gradle bootJar -x test --no-daemon --build-cache

FROM eclipse-temurin:21-jre-alpine AS final

RUN apk add --no-cache tzdata && \
    addgroup -S spring && adduser -S spring -G spring

WORKDIR /app
USER spring

COPY --from=builder /app/build/libs/*.jar app.jar

EXPOSE 8080
ENTRYPOINT ["java", "-XX:MaxRAMPercentage=75.0", "-jar", "app.jar"]