# weAlist User Service 개선점 및 진행 상황

## 📝 향후 개선점 (할 일 목록)

### 1. 개발 환경 개선 (Development Environment Improvement) - **높은 우선순위 (사용자 요청)**

*   **IDE 기반 로컬 개발 환경 구축**:
    *   **내용**: `gradlew` `JAVA_HOME` 문제로 인한 빌드 실패를 우회하고, 로컬 개발 시 IDE(IntelliJ, VS Code 등)를 활용하여 Spring Boot 애플리케이션을 직접 실행하는 환경을 구축합니다. `spring-boot-devtools`를 활용한 핫 리로딩을 통해 빠른 개발 피드백을 얻습니다.
    *   **방법**:
        1.  프로젝트를 IDE로 임포트합니다.
        2.  IDE에 올바른 JDK(`echo $JAVA_HOME`으로 확인된 경로)를 설정합니다.
        3.  `weAlist-deploy`에서 `docker compose up -d postgres redis`로 종속 서비스만 실행합니다.
        4.  IDE에서 `OrangeCloud.UserRepo.UserRepoApplication` 메인 클래스를 실행합니다.
    *   **이점**: `gradlew` 문제 우회, 빠른 개발 주기, IDE의 디버깅 기능 활용.

### 2. API 테스트 및 검증 (API Testing & Verification) - **높은 우선순위 (현재 진행 중)**

*   **Redis 기반 기능 최종 테스트**:
    *   **내용**: 현재까지 구현된 Redis 기반 기능(리프레시 토큰 블랙리스트, 속도 제한, 캐싱)에 대한 통합 테스트를 수행하여 모든 기능이 예상대로 작동하는지 확인합니다.
    *   **방법**:
        1.  `docker push ressbe/wealist-user:latest` (최신 이미지 푸시)
        2.  `cd weAlist-deploy && ./deploy.sh` (서비스 재배포)
        3.  다음 시나리오별 API 테스트 및 로그 확인:
            *   **로그아웃**: 액세스 토큰 블랙리스트 확인.
            *   **리프레시 토큰**: 이전 리프레시 토큰 블랙리스트 및 새 토큰 발급 확인.
            *   **속도 제한**: `signup` 엔드포인트에 여러 번 요청하여 제한 초과 시 `405 METHOD_NOT_ALLOWED` 응답 확인.
            *   **캐싱**: `getCurrentUser`를 여러 번 호출하여 캐시 동작 확인.

### 3. 기타 개선점 (Other Improvements)

*   **포괄적인 입력 유효성 검사 강화**: 모든 DTO에 JSR-303 어노테이션 및 커스텀 유효성 검사기 적용.
*   **API 문서화 상세화**: Swagger/OpenAPI 문서에 요청/응답, 파라미터, 오류 응답 예시 추가.
*   **테스트 - 포괄적인 테스트 스위트 구축**: 단위, 통합, DB 통합 테스트 커버리지 확보.
*   **보안 - 보안 헤더 및 CORS 설정 정교화**: 프로덕션 환경에 맞는 보안 헤더 및 CORS Origin 제한.
*   **관찰 가능성 - 메트릭 및 분산 추적**: Micrometer, Prometheus, OpenTelemetry/Zipkin 통합.
*   **코드 품질 및 유지보수성 향상**: 정기 코드 리뷰, 정적 분석 도구 활용, 리팩토링.
*   **국제화 (i18n)**: 다국어 지원을 위한 메시지 외부화.
