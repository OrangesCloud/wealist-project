import { useParams } from 'react-router-dom';
import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Briefcase } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
// 💡 [분리된 컴포넌트]
import MainLayout from '../components/layout/MainLayout';
import { ProjectHeader } from '../components/layout/ProjectHeader';
import { ProjectContent } from '../components/layout/ProjectContent';

import UserProfileModal from '../components/modals/UserProfileModal';
import { ProjectModal } from '../components/modals/ProjectModal';
import { LoadingSpinner } from '../components/common/LoadingSpinner';

import { getProjects } from '../api/board/boardService';
import { getWorkspaceMembers } from '../api/user/userService';
import { ProjectResponse } from '../types/board';
import { WorkspaceMemberResponse } from '../types/user';
import { CreateBoardModal } from '../components/modals/CreateBoardModal';
import { CustomFieldAddModal } from '../components/modals/customFields/CustomFieldAddModal';

interface MainDashboardProps {
  onLogout: () => void;
}
// 💡 [추가] UI/모달 상태를 통합하는 인터페이스
interface UIState {
  showProjectSelector: boolean; // 프로젝트 드롭다운
  showUserProfile: boolean; // 사용자 프로필 모달
  showCreateProject: boolean; // 프로젝트 생성 모달
  showManageModal: boolean; // 커스텀 필드 관리 모달
  showProjectSettings: boolean; // 프로젝트 설정 모달
  showCreateBoard: boolean;
}
// =============================================================================
// MainDashboard (컨테이너 역할)
// =============================================================================
const MainDashboard: React.FC<MainDashboardProps> = ({ onLogout }) => {
  const { workspaceId } = useParams<{ workspaceId: string }>();
  const currentWorkspaceId = workspaceId || '';

  const { theme } = useTheme();
  const currentRole = useRef<'OWNER' | 'ORGANIZER' | 'MEMBER'>('ORGANIZER');
  const canAccessSettings = currentRole.current === 'OWNER' || currentRole.current === 'ORGANIZER';

  // [핵심 상태]
  const [projects, setProjects] = useState<ProjectResponse[]>([]);
  const [selectedProject, setSelectedProject] = useState<ProjectResponse | null>(null);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);
  const [isLoadingProjects, setIsLoadingProjects] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 💡 [통합] UI/모달 상태를 하나의 객체로 관리
  const [uiState, setUiState] = useState<UIState>({
    showProjectSelector: false,
    showUserProfile: false,
    showCreateProject: false,
    showManageModal: false,
    showProjectSettings: false,
    showCreateBoard: false,
  });

  // 💡 ProjectContent로 넘길 상태: 보드 생성/수정 데이터
  const [editBoardData, setEditBoardData] = useState<any>(null);

  // 💡 [추가] 모달 상태 토글 헬퍼 함수
  const toggleUiState = useCallback((key: keyof UIState, show?: boolean) => {
    setUiState((prev) => ({
      ...prev,
      [key]: show !== undefined ? show : !prev[key],
    }));
  }, []);

  // 1. 프로젝트 목록 조회 함수
  const fetchProjects = useCallback(async () => {
    if (!currentWorkspaceId) return;

    setIsLoadingProjects(true);
    setError(null);
    try {
      const fetchedProjects = await getProjects(currentWorkspaceId);
      setProjects(fetchedProjects);

      if (fetchedProjects.length > 0 && !selectedProject) {
        // 프로젝트가 로드되었지만 아직 선택된 프로젝트가 없으면 첫 번째 프로젝트를 선택
        setSelectedProject(fetchedProjects[0]);
      }
    } catch (err) {
      const error = err as Error;
      // 에러 메시지 처리를 위해 객체 형태 대신 문자열로 변환하여 저장
      setError(`프로젝트 로드 실패: ${error.message}`);
    } finally {
      setIsLoadingProjects(false);
    }
  }, [currentWorkspaceId, selectedProject]);

  // 2. 워크스페이스 회원 조회 함수 (유지)
  const fetchWorkspaceMembers = useCallback(async () => {
    if (!currentWorkspaceId) return;
    try {
      const members = await getWorkspaceMembers(currentWorkspaceId);
      setWorkspaceMembers(members);
    } catch (err) {
      setWorkspaceMembers([]);
    }
  }, [currentWorkspaceId]);

  // 3. 초기 로드 (유지)
  useEffect(() => {
    fetchProjects();
    fetchWorkspaceMembers();
  }, [fetchProjects, fetchWorkspaceMembers]);

  // 💡 ProjectContent에서 보드가 업데이트되면 호출될 함수 (이 함수는 fetchBoards와 동일한 역할)
  const handleBoardContentUpdate = useCallback(() => {
    console.log('[Dashboard] Board content updated in ProjectContent.');
  }, []);

  // 💡 [추가] 필드가 생성된 후 호출될 핸들러
  const handleFieldCreated = (newField: any) => {
    // 1. 필드 생성 모달 닫기
    toggleUiState('showManageModal', false);

    // 2. MainDashboard가 필드 변경 사항을 ProjectContent에게 알려주기 위해 업데이트 트리거
    handleBoardContentUpdate();
  };

  return (
    // 💡 [수정] MainLayout에 UserProfile 토글 함수를 전달
    <MainLayout
      onLogout={onLogout}
      workspaceId={currentWorkspaceId}
      onProfileModalOpen={() => toggleUiState('showUserProfile', true)}
    >
      {/* 1. 헤더 영역 */}
      <ProjectHeader
        projects={projects}
        selectedProject={selectedProject}
        workspaceMembers={workspaceMembers}
        setSelectedProject={setSelectedProject}
        // 💡 [수정] 통합된 Setter 사용
        setShowCreateProject={() => toggleUiState('showCreateProject', true)}
        setShowProjectSettings={() => toggleUiState('showProjectSettings', true)}
        showProjectSelector={uiState.showProjectSelector}
        setShowProjectSelector={(show) => toggleUiState('showProjectSelector', show)}
        canAccessSettings={canAccessSettings}
      />

      {/* 2. 메인 콘텐츠 영역 */}
      <div className="flex-grow flex flex-col p-3 sm:p-6 overflow-auto mt-16 ml-20">
        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-300 rounded-lg text-red-700">
            {error}
          </div>
        )}

        {isLoadingProjects ? (
          <LoadingSpinner message="프로젝트를 로드 중..." />
        ) : selectedProject ? (
          <ProjectContent
            selectedProject={selectedProject}
            workspaceId={currentWorkspaceId}
            onProjectContentUpdate={handleBoardContentUpdate}
            // 💡 [수정] 통합된 Setter 사용
            onManageModalOpen={() => toggleUiState('showManageModal', true)}
            onEditBoard={setEditBoardData}
            // 💡 [추가] ProjectContent 내부에 필요한 모달 상태를 전달
            showCreateBoard={uiState.showCreateBoard}
            setShowCreateBoard={(show) => toggleUiState('showCreateBoard', show)}
          />
        ) : (
          <div className="flex flex-col items-center justify-center h-full text-center p-8">
            <Briefcase className="w-16 h-16 mb-4 text-gray-400" />
            <h2 className={`${theme.font.size.xl} ${theme.colors.text} mb-2`}>
              프로젝트를 선택하세요
            </h2>
            <p className={`${theme.colors.subText}`}>프로젝트 목록을 불러오고 선택하세요.</p>
          </div>
        )}
      </div>

      {/* 3. 모달 영역 */}

      {/* UserProfile Modal 💡 [복구] */}
      {uiState.showUserProfile && (
        <UserProfileModal onClose={() => toggleUiState('showUserProfile', false)} />
      )}

      {/* Create Project Modal */}
      {uiState.showCreateProject && (
        <ProjectModal
          workspaceId={currentWorkspaceId}
          onClose={() => toggleUiState('showCreateProject', false)}
          onProjectSaved={fetchProjects}
        />
      )}

      {/* Project Settings Modal */}
      {uiState.showProjectSettings && selectedProject && (
        <ProjectModal
          workspaceId={currentWorkspaceId}
          project={selectedProject}
          onClose={() => toggleUiState('showProjectSettings', false)}
          onProjectSaved={fetchProjects}
        />
      )}
      {/* 💡 [수정] Custom Field Add Modal */}
      {uiState.showManageModal && selectedProject && (
        <CustomFieldAddModal
          projectId={selectedProject.projectId}
          onClose={() => toggleUiState('showManageModal', false)}
          // 💡 [핵심] 필드 생성 후 실행될 핸들러 연결
          onFieldCreated={handleFieldCreated}
        />
      )}

      {/* Create/Edit Board Modal (editBoardData 상태 기반) */}
      {(editBoardData || uiState.showCreateBoard) && selectedProject && (
        <CreateBoardModal
          projectId={selectedProject.projectId}
          stageId={editBoardData?.stageId}
          editData={editBoardData}
          workspaceId={currentWorkspaceId}
          onClose={() => {
            setEditBoardData(null); // 편집 데이터 초기화
            toggleUiState('showCreateBoard', false); // 생성 모달 닫기
          }}
          onBoardCreated={handleBoardContentUpdate}
          // 💡 [추가] 필드 추가 모달을 열기 위한 핸들러
          onAddFieldsClick={() => {
            setEditBoardData(null); // 혹시 모를 편집 상태 제거
            toggleUiState('showCreateBoard', false); // CreateBoardModal 닫기
            toggleUiState('showManageModal', true); // CustomFieldAddModal 열기
          }}
        />
      )}
    </MainLayout>
  );
};

export default MainDashboard;
