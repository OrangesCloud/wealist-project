import React, { useState } from 'react';
// Styles & Contexts
import { ThemeProvider } from './contexts/ThemeContext';

// API Services & Types (Mock 함수가 여기서 import 됩니다.)
import { AuthResponse } from './api/userService';
import { createWorkspace, WorkspaceCreate } from './api/KanbanService';

// Pages & Components (실제 경로에 맞게 수정했습니다.)
import AuthPage from './pages/Authpage';
import SelectGroupPage from './components/SelectGroupPage';
import MainDashboard from './pages/Dashboard';

// 애플리케이션의 주요 흐름 상태 정의 (상태 머신)
type AppState = 'AUTH' | 'SELECT_GROUP' | 'CREATE_WORKSPACE' | 'KANBAN';

const App: React.FC = () => {
  // 1. 상태 관리
  const [appState, setAppState] = useState<AppState>('AUTH');
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [userId, setUserId] = useState<string | null>(null);
  const [currentGroupId, setCurrentGroupId] = useState<string | null>(null);
  const [loadingMessage, setLoadingMessage] = useState<string | null>(null);

  // 2. 인증 성공 핸들러 (AUTH -> SELECT_GROUP)
  const handleAuthSuccess = (authData: AuthResponse) => {
    setAccessToken(authData.accessToken);
    setUserId(authData.userId);

    localStorage.setItem('access_token', authData.accessToken);
    localStorage.setItem('user_id', authData.userId);

    setAppState('SELECT_GROUP');
  };

  // 3. 로그아웃 핸들러 (KANBAN -> AUTH)
  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user_id');
    setAccessToken(null);
    setUserId(null);
    setCurrentGroupId(null);
    setAppState('AUTH');
  };

  // 4. 그룹 선택 성공 핸들러 (SELECT_GROUP -> CREATE_WORKSPACE -> KANBAN)
  const handleGroupSelectionSuccess = async (groupId: string) => {
    // 💡 오류 1 해결: 이 시점에서는 accessToken과 userId가 string임을 확신합니다.
    if (!accessToken || !userId) {
      alert('인증 정보가 유효하지 않습니다. 다시 로그인해주세요.');
      handleLogout();
      return;
    }

    setCurrentGroupId(groupId);
    setLoadingMessage('워크스페이스를 생성하고 초기 설정을 진행합니다...');
    setAppState('CREATE_WORKSPACE');

    try {
      // 4-1. Workspace 생성 요청 데이터 준비
      const workspaceData: WorkspaceCreate = {
        name: 'My Kanban Workspace - ' + groupId.substring(0, 8),
        description: `Group ID ${groupId}를 위한 기본 공간 (그룹 선택 완료)`,
      };

      // 4-2. 💡 Mock API 호출: accessToken! 으로 non-null assertion 사용
      // Mock API가 성공적으로 호출됩니다.
      const newWorkspace = await createWorkspace(workspaceData, accessToken!);

      console.log('✅ Workspace 생성 성공 (Mock):', newWorkspace);

      setLoadingMessage(null);
      setAppState('KANBAN');
    } catch (error: any) {
      const errorMessage = error.message || '알 수 없는 오류';
      console.error('❌ Workspace 생성 실패 (Mock/API):', error);
      alert(`Workspace 생성 중 오류가 발생했습니다. 메시지: ${errorMessage}`);
      setLoadingMessage(null);
      setAppState('SELECT_GROUP'); // 실패 시 그룹 선택 페이지로 복귀
    }
  };

  // 5. 렌더링 로직 (상태 기반 분기)
  const renderContent = () => {
    // 💡 AUTH 단계
    if (appState === 'AUTH') {
      return <AuthPage onLogin={handleAuthSuccess} />;
    }

    // 💡 SELECT_GROUP 단계
    if (appState === 'SELECT_GROUP' && userId && accessToken) {
      return (
        <SelectGroupPage
          userId={userId}
          accessToken={accessToken}
          onGroupSelected={handleGroupSelectionSuccess}
        />
      );
    }

    // 💡 CREATE_WORKSPACE 단계
    if (appState === 'CREATE_WORKSPACE') {
      return (
        <div className="text-center min-h-screen flex items-center justify-center bg-gray-50">
          <div className="p-8 bg-white rounded-xl shadow-lg">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
            <h1 className="text-xl font-medium text-gray-800">{loadingMessage}</h1>
          </div>
        </div>
      );
    }

    // 💡 KANBAN 단계
    // 오류 2 해결: 명확한 상태 분기를 통해 타입 비교 오류 방지
    if (appState === 'KANBAN' && currentGroupId && accessToken) {
      return (
        // 💡 오류 3 해결: MainDashboard가 currentGroupId와 accessToken을 props로 받습니다.
        <MainDashboard
          onLogout={handleLogout}
          currentGroupId={currentGroupId}
          accessToken={accessToken}
        />
      );
    }

    // 알 수 없는 상태로 빠졌을 경우 (예외 처리)
    return (
      <div className="p-4 text-center">
        <p className="text-red-500 mb-4">세션 오류 발생. 재로그인합니다.</p>
        <AuthPage onLogin={handleAuthSuccess} />
      </div>
    );
  };

  return <ThemeProvider>{renderContent()}</ThemeProvider>;
};

export default App;
