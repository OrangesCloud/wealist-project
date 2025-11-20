import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';

// 환경 변수 가져오기
const INJECTED_API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

// ============================================================================
// 💡 [핵심 수정]: Context Path를 환경에 따라 조건부로 붙입니다.
// ============================================================================

const getApiBaseUrl = (path: string): string => {
  // 1. 환경 변수 주입 확인
  if (INJECTED_API_BASE_URL) {
    // 쉘 스크립트에서 VITE_API_BASE_URL='http://localhost'가 주입된 경우
    const isLocalDevelopment = INJECTED_API_BASE_URL.includes('localhost');

    if (isLocalDevelopment) {
      // 🔥 로컬 개발: 각 서비스별 포트 직접 지정
      if (path === '/api/users') return `${INJECTED_API_BASE_URL}:8080`;
      if (path === '/api/boards/api') return `${INJECTED_API_BASE_URL}:8000/api`;
      if (path === '/api/chats') return `${INJECTED_API_BASE_URL}:8001/api`; // 🔥 추가
    }

    return `${INJECTED_API_BASE_URL}${path}`;
  }

  // 환경 변수가 없을 경우 (Fallback, CI/CD 실패 대비)
  return `https://api.wealist.co.kr${path}`;
};

export const USER_REPO_API_URL = getApiBaseUrl('/api/users');
export const BOARD_SERVICE_API_URL = getApiBaseUrl('/api/boards/api');
export const CHAT_SERVICE_API_URL = getApiBaseUrl('/api/chats'); // 🔥 추가

// ============================================================================
// Axios 인스턴스 생성
// ============================================================================

/**
 * User Repo API (Java)를 위한 Axios 인스턴스
 */
export const userRepoClient = axios.create({
  baseURL: USER_REPO_API_URL,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
});

/**
 * Board Service API (Go)를 위한 Axios 인스턴스
 */
export const boardServiceClient = axios.create({
  baseURL: BOARD_SERVICE_API_URL,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
});

/**
 * Chat Service API (Go)를 위한 Axios 인스턴스
 */
export const chatServiceClient = axios.create({
  baseURL: CHAT_SERVICE_API_URL,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
});

// ============================================================================
// 인증 갱신 헬퍼 함수 (기존 코드 유지)
// ============================================================================

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: Error | null, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

const performLogout = () => {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
  localStorage.removeItem('nickName');
  localStorage.removeItem('userEmail');
  window.location.href = '/';
};

const refreshAccessToken = async (): Promise<string> => {
  const refreshToken = localStorage.getItem('refreshToken');
  if (!refreshToken) {
    console.warn('⚠️ Refresh token not found. Logging out...');
    performLogout();
    throw new Error('No refresh token available');
  }

  try {
    const response = await axios.post(`${USER_REPO_API_URL}/api/auth/refresh`, {
      refreshToken,
    });

    const { accessToken, refreshToken: newRefreshToken } = response.data;

    localStorage.setItem('accessToken', accessToken);
    if (newRefreshToken) {
      localStorage.setItem('refreshToken', newRefreshToken);
    }

    return accessToken;
  } catch (error) {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('nickName');
    localStorage.removeItem('userEmail');
    window.location.href = '/';
    throw error;
  }
};

// ============================================================================
// 인터셉터 설정
// ============================================================================

const setupRequestInterceptor = (client: AxiosInstance) => {
  client.interceptors.request.use(
    (config) => {
      const accessToken = localStorage.getItem('accessToken');
      if (accessToken && !config.headers.Authorization) {
        config.headers.Authorization = `Bearer ${accessToken}`;
      }
      return config;
    },
    (error) => {
      return Promise.reject(error);
    },
  );
};

const setupUnifiedResponseInterceptor = (client: AxiosInstance) => {
  client.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
      const originalRequest = error.config as InternalAxiosRequestConfig & {
        _retry?: boolean;
        retryCount?: number;
      };
      const status = error.response?.status;

      if (status === 401 && !originalRequest._retry) {
        if (isRefreshing) {
          return new Promise((resolve, reject) => {
            failedQueue.push({ resolve, reject });
          })
            .then((token) => {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              return client(originalRequest);
            })
            .catch((err) => {
              return Promise.reject(err);
            });
        }

        originalRequest._retry = true;
        isRefreshing = true;

        try {
          const newAccessToken = await refreshAccessToken();
          processQueue(null, newAccessToken);
          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
          return client(originalRequest);
        } catch (refreshError) {
          processQueue(refreshError as Error, null);
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      }

      if (status && status >= 400 && status < 599) {
        return Promise.reject(error);
      }

      if (!status && error.code !== 'ERR_CANCELED') {
        originalRequest.retryCount = originalRequest.retryCount || 0;

        if (originalRequest.retryCount >= 5) {
          console.error(`[Axios Interceptor] 최대 재시도 횟수 초과: ${originalRequest.url}`);
          return Promise.reject(error);
        }

        originalRequest.retryCount += 1;
        const delay = new Promise((resolve) => setTimeout(resolve, 1000));
        console.warn(
          `[Axios Interceptor] 재시도 중 (${originalRequest.retryCount}회): ${originalRequest.url}`,
        );
        await delay;
        return client(originalRequest);
      }
    },
  );
};

// 인터셉터 적용
setupRequestInterceptor(userRepoClient);
setupRequestInterceptor(boardServiceClient);
setupRequestInterceptor(chatServiceClient); // 🔥 추가

setupUnifiedResponseInterceptor(userRepoClient);
setupUnifiedResponseInterceptor(boardServiceClient);
setupUnifiedResponseInterceptor(chatServiceClient); // 🔥 추가

export const getAuthHeaders = (token: string) => ({
  Authorization: `Bearer ${token}`,
  Accept: 'application/json',
});
