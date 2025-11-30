import React, { Suspense, lazy } from 'react';
// AuthProvider와 ThemeProvider가 src/contexts 아래에 있으므로 상대 경로를 사용합니다.
import { ThemeProvider } from './contexts/ThemeContext';
// 💡 [수정] AuthProvider와 useAuth 훅을 가져옵니다.
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { Loader2 } from 'lucide-react'; // 로딩 스피너 아이콘

// Lazy load 페이지들 (이름 일관성 유지)
const AuthPage = lazy(() => import('./pages/Authpage'));
const SelectWorkspacePage = lazy(() => import('./pages/SelectWorkspacePage'));
const MainDashboard = lazy(() => import('./pages/MainDashboard'));
const OAuthRedirectPage = lazy(() => import('./pages/OAuthRedirectPage'));

// -----------------------------------------------------------------------------
// 로딩 스크린 컴포넌트 (Loader2 아이콘 사용)
// -----------------------------------------------------------------------------
const LoadingScreen = ({ msg = '로딩 중...' }) => (
  <div className="text-center min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
    <div className="p-8 bg-white dark:bg-gray-800 rounded-xl shadow-lg">
      <Loader2 className="h-12 w-12 animate-spin text-blue-500 mx-auto mb-4" />
      <h1 className="text-xl font-medium text-gray-800 dark:text-gray-100">{msg}</h1>
    </div>
  </div>
);

// -----------------------------------------------------------------------------
// 2. 인증이 필요한 페이지를 감싸는 '보호 라우트' 컴포넌트
// -----------------------------------------------------------------------------
const ProtectedRoute = () => {
  // 💡 [수정] useAuth 훅을 사용하여 상태를 가져옵니다.
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    // 인증 상태를 확인하는 동안 로딩 스크린 표시
    return <LoadingScreen msg="인증 확인 중..." />;
  }

  // 토큰이 없거나 유효하지 않으면 로그인 페이지로 리다이렉트
  if (!isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  // 인증되면 자식 컴포넌트 렌더링
  return <Outlet />;
};

const App: React.FC = () => {
  // 4. [제거] handleLogout 함수와 useNavigate 선언을 제거합니다.
  // MainDashboard에서는 useAuth().logout()을 직접 호출해야 합니다.

  // 5. renderContent 함수 대신 Routes를 사용합니다.
  return (
    // AuthProvider가 ThemeProvider를 감싸도록 순서를 조정했습니다.
    <ThemeProvider>
      <Suspense fallback={<LoadingScreen />}>
        <AuthProvider>
          <Routes>
            {/* 1. 로그인 페이지 */}
            <Route path="/" element={<AuthPage />} />

            {/* 2. OAuth 콜백 페이지 */}
            <Route path="/oauth/callback" element={<OAuthRedirectPage />} />

            {/* 3. 보호되는 라우트 (인증 필요) */}
            <Route element={<ProtectedRoute />}>
              {/* SelectWorkspacePage는 이제 props가 필요 없습니다. */}
              <Route path="/workspaces" element={<SelectWorkspacePage />} />

              {/* [수정] MainDashboard는 onLogout prop이 필요 없습니다. useAuth().logout을 사용하세요. */}
              <Route path="/workspace/:workspaceId" element={<MainDashboard />} />
            </Route>

            {/* 4. 일치하는 라우트가 없으면 로그인 페이지로 */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </AuthProvider>
      </Suspense>
    </ThemeProvider>
  );
};

export default App;
