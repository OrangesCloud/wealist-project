import React, { createContext, useContext, useState, useEffect, useMemo, ReactNode } from 'react';

// 사용자 정보 전체 구조 정의
interface UserInfo {
  userId: string;
  nickName?: string;
  userEmail?: string;
}

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: UserInfo | null; // UserInfo 타입 사용
  // 💡 [수정] userId, nickName, userEmail을 Context의 최상위 속성으로 노출
  userId: string | undefined;
  nickName: string | undefined;
  userEmail: string | undefined;
  // ⭐️ [복원] token 속성을 추가하여 기존 컴포넌트의 하위 호환성을 유지합니다.
  token: string | undefined;
  login: (
    accessToken: string,
    refreshToken: string,
    userId: string,
    nickName: string,
    userEmail: string,
  ) => void; // 인자 확장
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// -----------------------------------------------------------------------------
// Auth Provider
// -----------------------------------------------------------------------------
export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState<UserInfo | null>(null);

  // 💡 [추가] user 상태에서 모든 필드를 파생시켜 Context에 직접 노출합니다.
  const currentUserId = user?.userId;
  const currentNickName = user?.nickName;
  const currentUserEmail = user?.userEmail;

  // ⭐️ [추가] accessToken을 Context에서 바로 사용할 수 있도록 파생시킵니다.
  const currentToken = localStorage.getItem('accessToken') || undefined;

  useEffect(() => {
    // 💡 초기 로드 시 localStorage에서 토큰 및 모든 사용자 정보 확인
    const checkAuthStatus = () => {
      const accessToken = localStorage.getItem('accessToken');
      const userId = localStorage.getItem('userId');
      const nickName = localStorage.getItem('nickName');
      const userEmail = localStorage.getItem('userEmail');

      // 토큰과 userId가 존재하면 인증된 것으로 간주
      if (accessToken && userId) {
        // 실제로는 여기서 백엔드에 토큰 유효성 검증 API 호출을 해야 함
        setIsAuthenticated(true);
        setUser({
          userId: userId,
          nickName: nickName || undefined,
          userEmail: userEmail || undefined,
        });
      } else {
        setIsAuthenticated(false);
        setUser(null);
      }
      setIsLoading(false);
    };

    checkAuthStatus();
  }, []);

  const login = (
    accessToken: string,
    refreshToken: string,
    userId: string,
    nickName: string,
    userEmail: string,
  ) => {
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', refreshToken);
    localStorage.setItem('userId', userId);
    localStorage.setItem('nickName', nickName);
    localStorage.setItem('userEmail', userEmail);

    setIsAuthenticated(true);
    setUser({ userId, nickName, userEmail });
  };

  const logout = () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('userId');
    localStorage.removeItem('nickName');
    localStorage.removeItem('userEmail');

    setIsAuthenticated(false);
    setUser(null);
  };

  const value = useMemo(
    // 💡 [수정] Context 값에 모든 파생 변수를 포함하여 반환
    () => ({
      isAuthenticated,
      isLoading,
      user,
      userId: currentUserId,
      nickName: currentNickName,
      userEmail: currentUserEmail,
      token: currentToken, // ⭐️ [추가] token 속성 노출
      login,
      logout,
    }),
    // 💡 [수정] 의존성 배열에 모든 파생 변수를 추가합니다.
    [
      isAuthenticated,
      isLoading,
      user,
      currentUserId,
      currentNickName,
      currentUserEmail,
      currentToken,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// -----------------------------------------------------------------------------
// Hook
// -----------------------------------------------------------------------------
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
