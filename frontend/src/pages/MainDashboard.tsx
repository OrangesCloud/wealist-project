// src/pages/Dashboard.tsx (MainDashboard.tsx)

import { useParams } from 'react-router-dom';
import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Briefcase } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';

// 💡 [분리된 컴포넌트]
import MainLayout from '../components/layout/MainLayout';
import { ProjectHeader } from '../components/layout/ProjectHeader';
import { ProjectContent } from '../components/layout/ProjectContent';

import UserProfileModal from '../components/modals/user/UserProfileModal';
import { LoadingSpinner } from '../components/common/LoadingSpinner';

// 💡 [수정] Init Data API 임포트
import { getProjects, getProjectInitData } from '../api/board/boardService';
import { getWorkspaceMembers } from '../api/user/userService';

import {
  ProjectResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
  FieldWithOptionsResponse,
  FieldOptionsLookup, // 💡 필드와 옵션 정보를 담는 통합 DTO
} from '../types/board';
import { WorkspaceMemberResponse } from '../types/user';
import { CustomFieldAddModal } from '../components/modals/board/customFields/CustomFieldAddModal';
import { BoardManageModal } from '../components/modals/board/BoardManageModal';
import { ProjectModal } from '../components/modals/board/ProjectModal';

interface MainDashboardProps {
  onLogout: () => void;
}

// 💡 [추가] UI/모달 상태를 통합하는 인터페이스
interface UIState {
  showProjectSelector?: boolean;
  showUserProfile?: boolean;
  showCreateProject?: boolean;
  showManageModal?: boolean;
  showProjectSettings?: boolean;
  showCreateBoard?: boolean;
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

  const [uiState, setUiState] = useState<UIState>();
  const [editBoardData, setEditBoardData] = useState<any>(null);

  // 💡 [추가] 초기 옵션 데이터를 저장할 상태 (ProjectContent로 전달)
  const [fieldOptionsLookup, setFieldOptionsLookup] = useState<FieldOptionsLookup>({
    roles: [],
    importances: [],
    stages: [],
  });

  const toggleUiState = useCallback((key: keyof UIState, show?: boolean) => {
    setUiState((prev) => ({
      ...prev,
      [key]: show !== undefined ? show : !prev?.[key],
    }));
  }, []);

  // 💡 [추가] Helper: FieldWithOptionsResponse -> Custom DTO 변환
  const mapFieldOptions = (fields: FieldWithOptionsResponse[]) => {
    const roles: CustomRoleResponse[] = [];
    const importances: CustomImportanceResponse[] = [];

    fields.forEach((field) => {
      if (field.fieldType === 'single_select' || field.fieldType === 'multi_select') {
        field.options.forEach((opt) => {
          const base = {
            label: opt.label,
            color: opt.color,
            displayOrder: opt.displayOrder,
            fieldId: opt.fieldId,
            isSystemDefault: field.isSystemDefault,
            description: opt.description || '',
          };

          if (field.name === '역할') {
            // 💡 필드 이름으로 구분
            roles.push({ ...base, roleId: opt.optionId });
          } else if (field.name === '중요도') {
            importances.push({ ...base, importanceId: opt.optionId, level: 1 }); // Level은 Config에서 추출해야 하지만 Mock 처리
          }
        });
      }
    });

    return { roles, importances };
  };

  // 1. 프로젝트 목록 조회 함수 (Header Dropdown용)
  const fetchProjects = useCallback(async () => {
    if (!currentWorkspaceId) return;

    setIsLoadingProjects(true);
    setError(null);
    try {
      const fetchedProjects = await getProjects(currentWorkspaceId);
      setProjects(fetchedProjects);

      if (fetchedProjects.length > 0 && !selectedProject) {
        setSelectedProject(fetchedProjects[0]);
      }
    } catch (err: any) {
      const error = err as Error;
      setError(`프로젝트 목록 로드 실패: ${error.message}`);
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

  // 💡 [핵심 구현] 프로젝트 선택 시 모든 데이터 로드
  const fetchProjectContentInitData = useCallback(async () => {
    if (!selectedProject) return;

    // 로딩은 ProjectContent 내부에서 다시 시작되므로, 여기서는 에러만 초기화
    setError(null);

    try {
      const initData = await getProjectInitData(selectedProject.projectId);

      // 1. Project Header의 selectedProject를 InitData의 project DTO로 업데이트 (더 풍부한 정보 포함)
      // setSelectedProject(initData.project);

      // 2. 멤버 목록 업데이트 (InitData에서 제공된 멤버 사용)
      // setWorkspaceMembers(initData.members);

      // 3. 필드 옵션 룩업 테이블 생성
      const fieldLookup = mapFieldOptions(initData.fields);
      setFieldOptionsLookup(fieldLookup);

      console.log('✅ Project Init Data (Fields/Boards) Loaded.');
    } catch (err: any) {
      setError(`초기 컨텐츠 로드 실패: ${err.message}`);
    }
  }, [selectedProject]);

  // 3. 초기 로드 및 트리거
  useEffect(() => {
    fetchProjects();
    fetchWorkspaceMembers();
  }, [fetchProjects, fetchWorkspaceMembers]);

  useEffect(() => {
    if (selectedProject) {
      fetchProjectContentInitData();
    }
  }, [selectedProject, fetchProjectContentInitData]);

  // 💡 ProjectContent에서 보드/필드 업데이트 시 호출될 함수
  const handleBoardContentUpdate = useCallback(() => {
    console.log('[Dashboard] Board content updated in ProjectContent. Reloading Field Data.');
    // 💡 보드/필드 데이터가 변경되면, InitData를 다시 로드하여 ProjectContent에 전달
    fetchProjectContentInitData();
  }, [fetchProjectContentInitData]);

  // 💡 필드가 생성된 후 호출될 핸들러
  const handleFieldCreated = useCallback(
    (newField: any) => {
      toggleUiState('showManageModal', false);
      handleBoardContentUpdate(); // 필드 데이터 및 보드 데이터 갱신
      console.log(`✅ New field created and propagated: ${newField?.name}`);
    },
    [handleBoardContentUpdate, toggleUiState],
  );

  return (
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
        setShowCreateProject={() => toggleUiState('showCreateProject', true)}
        setShowProjectSettings={() => toggleUiState('showProjectSettings', true)}
        showProjectSelector={uiState?.showProjectSelector || false}
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
          <LoadingSpinner message="프로젝트 목록 로드 중..." />
        ) : selectedProject ? (
          <ProjectContent
            selectedProject={selectedProject}
            workspaceId={currentWorkspaceId}
            onProjectContentUpdate={handleBoardContentUpdate}
            onManageModalOpen={() => toggleUiState('showManageModal', true)}
            onEditBoard={setEditBoardData}
            showCreateBoard={uiState?.showCreateBoard || false}
            setShowCreateBoard={(show) => toggleUiState('showCreateBoard', show)}
            // 💡 [전달] 룩업 데이터를 ProjectContent로 전달 (Mock 대체)
            fieldOptionsLookup={fieldOptionsLookup}
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

      {/* UserProfile Modal */}
      {uiState?.showUserProfile && (
        <UserProfileModal onClose={() => toggleUiState('showUserProfile', false)} />
      )}

      {/* Create Project Modal */}
      {uiState?.showCreateProject && (
        <ProjectModal
          workspaceId={currentWorkspaceId}
          onClose={() => toggleUiState('showCreateProject', false)}
          onProjectSaved={fetchProjects}
        />
      )}

      {/* Project Settings Modal */}
      {uiState?.showProjectSettings && selectedProject && (
        <ProjectModal
          workspaceId={currentWorkspaceId}
          project={selectedProject}
          onClose={() => toggleUiState('showProjectSettings', false)}
          onProjectSaved={fetchProjects}
        />
      )}
      {/* 💡 Custom Field Add Modal (필드 추가/정의) */}
      {uiState?.showManageModal && selectedProject && (
        <CustomFieldAddModal
          projectId={selectedProject.projectId}
          onClose={() => toggleUiState('showManageModal', false)}
          onFieldCreated={handleFieldCreated} // 필드 생성 후 갱신 트리거
        />
      )}

      {/* Create/Edit Board Modal */}
      {(editBoardData || uiState?.showCreateBoard) && selectedProject && (
        <BoardManageModal
          projectId={selectedProject.projectId}
          stageId={editBoardData?.stageId}
          editData={editBoardData}
          workspaceId={currentWorkspaceId}
          onClose={() => {
            setEditBoardData(null);
            toggleUiState('showCreateBoard', false);
          }}
          onBoardCreated={handleBoardContentUpdate}
          onAddFieldsClick={() => {
            toggleUiState('showManageModal', true);
          }}
          // 💡 [추가] 필드 옵션 룩업 데이터 전달
          fieldOptionsLookup={fieldOptionsLookup}
        />
      )}
    </MainLayout>
  );
};

export default MainDashboard;
