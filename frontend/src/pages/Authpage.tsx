import React, { useState } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import { AuthResponse } from '../api/userService';
// 💡 실제 API 호출(login, signup) 대신 Google 연동 플로우를 위해 AuthResponse 타입만 사용합니다.

interface AuthPageProps {
  // 로그인 성공 시 인증 데이터를 전달하여 다음 단계(조직 선택)로 이동
  onLogin: (authData: AuthResponse) => void;
}

const AuthPage: React.FC<AuthPageProps> = ({ onLogin }) => {
  const { theme } = useTheme();

  // 상태: 로딩, 에러
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // 1. Google 로그인 핸들러 (OAuth2 시작점)
  const handleGoogleLogin = () => {
    setError(null);
    setIsLoading(true);

    try {
      // 🚧 [Mock] 백엔드에서 인증에 성공했다고 가정하고, 더미 데이터를 생성합니다.
      // 이 데이터는 다음 단계(조직 선택)에 필요한 최소한의 정보입니다.
      const dummyAuthData: AuthResponse = {
        accessToken: 'MOCK_KANBAN_ACCESS_TOKEN_FOR_API_CALLS',
        refreshToken: 'MOCK_KANBAN_REFRESH_TOKEN',
        userId: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', // 조직 선택에 필요한 사용자 ID (UUID 형식)
        name: 'Mock User (Google)',
        email: 'mock.user@wealist.com',
        tokenType: 'Bearer',
      };

      // 1초 딜레이 후 성공 처리 (실제 네트워크 지연 효과)
      setTimeout(() => {
        setIsLoading(false);
        alert('Mock 인증 성공! 다음 단계로 이동합니다.');
        // 🚀 onLogin 호출 -> App.tsx에서 SELECT_GROUP 상태로 전환
        onLogin(dummyAuthData);
      }, 1000);
    } catch (e) {
      setIsLoading(false);
      setError('인증 처리 중 오류가 발생했습니다.');
    }
  };

  return (
    <div
      className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4 relative overflow-hidden`}
    >
      <div
        className="absolute inset-0 opacity-5"
        style={{
          backgroundImage:
            'linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
      ></div>

      <div
        className={`${theme.colors.primary} ${theme.effects.borderRadius} p-2 w-full max-w-md relative z-10 shadow-2xl ${theme.effects.borderWidth} ${theme.colors.border}`}
      >
        <div
          className={`${theme.colors.secondary} ${theme.effects.cardBorderWidth} ${theme.colors.border} p-4 sm:p-6 ${theme.effects.borderRadius}`}
        >
          <img src="/logo.png" alt="logo"></img>
          <p className={`${theme.font.size.xs} ${theme.colors.text} mb-4 sm:mb-6 text-center`}>
            Google 계정으로 로그인하고 워크스페이스를 시작하세요.
          </p>

          {/* 에러 메시지 표시 */}
          {error && <p className="text-red-500 text-center mb-4 text-sm font-medium">{error}</p>}

          {/* 💡 기존 이메일/비밀번호 입력 필드와 버튼 영역은 제거되었습니다. */}

          <div className="relative mb-4 sm:mb-6">
            <div className="absolute inset-0 flex items-center">
              <div className={`w-full border-t-2 sm:border-t-4 border-gray-200`}></div>
            </div>
            <div className={`relative flex justify-center ${theme.font.size.xs}`}>
              <span
                className={`px-2 ${theme.colors.secondary} ${theme.colors.text} ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.effects.borderRadius}`}
              >
                START WITH
              </span>
            </div>
          </div>

          <div className="space-y-2 sm:space-y-3">
            <div className="relative">
              <button
                onClick={handleGoogleLogin}
                disabled={isLoading}
                className={`relative w-full flex items-center justify-center gap-2 sm:gap-3 px-3 sm:px-4 py-2 sm:py-3 ${
                  theme.effects.cardBorderWidth
                } ${theme.colors.border} ${theme.colors.secondary} hover:bg-gray-50 transition ${
                  theme.effects.borderRadius
                } ${isLoading ? 'opacity-50 cursor-not-allowed' : ''}`}
              >
                <div
                  className={`w-4 h-4 sm:w-5 sm:h-5 bg-red-500 border-2 ${theme.colors.border} flex-shrink-0`}
                ></div>
                <span className={`font-bold ${theme.font.size.xs}`}>
                  {isLoading ? '인증 처리 중...' : 'GOOGLE 계정으로 계속하기 (Mock)'}
                </span>
              </button>
            </div>
            {/* 💡 GITHUB, KAKAO 버튼은 Google 전용 플로우를 위해 제거되었습니다. */}
          </div>

          {/* 💡 하단 전환 버튼 영역도 제거되었습니다. */}
        </div>
      </div>
    </div>
  );
};

export default AuthPage;
