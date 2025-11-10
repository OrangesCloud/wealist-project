import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios';

// 1. User/Workspace 서비스 (Java 백엔드) 기본 URL
export const USER_REPO_API_URL = 'http://localhost:8080';

// 2. Board/Project 서비스 (Go 백엔드) 기본 URL
export const BOARD_SERVICE_API_URL = 'http://localhost:8000';

// 💡 재시도 설정 상수
const MAX_RETRIES = 5; // 최대 5회 재시도 (총 6회 요청)
const RETRY_DELAY_MS = 1000; // 재시도 간격 (1초)

/**
 * 재시도 로직을 구현하는 인터셉터 함수
 * @param client Axios 인스턴스
 */
const setupRetryInterceptor = (client: AxiosInstance) => {
  client.interceptors.response.use(
    (response) => response,
    async (error) => {
      const { config, response } = error;

      // 💡 4xx 에러나 500이 아닌 경우에만 재시도 로직을 실행합니다.
      // (주로 네트워크 단절, DNS 오류 등 서버가 꺼졌을 때 발생하는 에러에 초점)
      // 400~500번대 HTTP 에러는 백엔드 로직 오류이므로 재시도하지 않습니다.
      if (response && response.status >= 400 && response.status < 599) {
        return Promise.reject(error);
      }

      // 재시도 관련 설정이 없다면, 현재 요청을 위한 임시 설정 추가
      const currentConfig: InternalAxiosRequestConfig & { retryCount?: number } = config;
      currentConfig.retryCount = currentConfig.retryCount || 0;

      // 1. 재시도 횟수가 최대 횟수를 초과했는지 확인
      if (currentConfig.retryCount >= MAX_RETRIES) {
        console.error(
          `[Axios Interceptor] 최대 재시도 횟수(${MAX_RETRIES}회) 초과. 요청 중단: ${config.url}`,
        );
        return Promise.reject(error); // 최종적으로 에러를 반환하고 호출자에게 전달
      }

      // 2. 재시도 횟수 증가
      currentConfig.retryCount += 1;

      // 3. 지수 백오프 대신 단순 지연 시간 적용
      const delay = new Promise((resolve) => {
        setTimeout(resolve, RETRY_DELAY_MS);
      });

      console.warn(
        `[Axios Interceptor] 요청 실패(${currentConfig.retryCount}회 재시도 중): ${config.url}`,
      );

      // 4. 지연 후, 원래 설정으로 요청 재시도
      await delay;
      return client(currentConfig);
    },
  );
};

/**
 * User Repo API (Java)를 위한 Axios 인스턴스
 */
export const userRepoClient = axios.create({
  baseURL: USER_REPO_API_URL,
  headers: { 'Content-Type': 'application/json' },
});

/**
 * Board Service API (Go)를 위한 Axios 인스턴스
 */
export const boardServiceClient = axios.create({
  baseURL: BOARD_SERVICE_API_URL,
  headers: { 'Content-Type': 'application/json' },
});

// 💡 두 클라이언트 인스턴스에 인터셉터 적용
setupRetryInterceptor(userRepoClient);
setupRetryInterceptor(boardServiceClient);

/**
 * JWT 토큰을 포함하는 인증 헤더를 반환합니다.
 */
export const getAuthHeaders = (token: string) => ({
  Authorization: `Bearer ${token}`,
  Accept: 'application/json',
});
