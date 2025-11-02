import React, { useEffect, useState, useCallback } from 'react';
import { Menu, User, ChevronDown, Plus, MoreVertical, X, Search } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import UserProfileModal from '../components/modals/UserProfileModal';
import TaskDetailModal from '../components/modals/TaskDetailModal';
import ColumnDetailModal from '../components/modals/ColumnDetailModal'; // 💡 ColumnDetailModal import
import { UserProfile } from '../types';

// --- 1. API 스펙에 맞춘 Mock 데이터 타입 정의 ---
interface WorkspaceResponse {
  id: string;
  name: string;
  created_by: string;
}
interface ProjectResponse {
  id: string;
  name: string;
  workspace_id: string;
}
interface Task {
  id: string;
  title: string;
  assignee_id: string | null;
  status: string;
  assignee: string;
}
interface Column {
  id: string;
  title: string;
  tasks: Task[];
}
// -------------------------------------------------

interface MainDashboardProps {
  onLogout: () => void;
  currentGroupId: string;
  accessToken: string;
}

// --- 2. Mock API 함수 정의 (백엔드 대체) ---
const mockFetchWorkspaces = async (_token: string): Promise<{ items: WorkspaceResponse[] }> => {
  console.log('[Mock] API: 조직(Workspace) 목록 조회');
  await new Promise((resolve) => setTimeout(resolve, 300));
  return {
    items: [
      { id: 'ws-mock-111', name: 'Wealist 개발팀 (Mock)', created_by: 'user-1' },
      { id: 'ws-mock-222', name: 'Orange Cloud 디자인팀 (Mock)', created_by: 'user-2' },
      { id: 'ws-mock-333', name: '개인 스터디 (Mock)', created_by: 'user-1' },
    ],
  };
};

const mockFetchProjects = async (
  workspaceId: string,
  _token: string,
): Promise<{ items: ProjectResponse[] }> => {
  console.log(`[Mock] API: 프로젝트 목록 조회 (Workspace: ${workspaceId})`);
  await new Promise((resolve) => setTimeout(resolve, 200));
  if (workspaceId === 'ws-mock-222') {
    return {
      items: [
        { id: 'prj-mock-design-A', name: '랜딩페이지 디자인', workspace_id: workspaceId },
        { id: 'prj-mock-design-B', name: 'BI/CI 리뉴얼', workspace_id: workspaceId },
      ],
    };
  }
  return {
    items: [
      { id: 'prj-mock-samsung', name: '삼성물산 백오피스 (Mock)', workspace_id: workspaceId },
      { id: 'prj-mock-cj', name: 'CJ 어드민 페이지 (Mock)', workspace_id: workspaceId },
      { id: 'prj-mock-internal', name: '자체 서비스 (Wealist)', workspace_id: workspaceId },
    ],
  };
};

const mockFetchKanbanBoard = async (projectId: string, _token: string): Promise<Column[]> => {
  console.log(`[Mock] API: 칸반 보드 로드 (Project: ${projectId})`);
  await new Promise((resolve) => setTimeout(resolve, 400));

  const baseTasks: Task[] = [
    {
      id: 't-1',
      title: `[${projectId.slice(0, 5)}] 인증 API 개발`,
      assignee_id: 'user-1',
      status: 'BACKEND',
      assignee: '김개발',
    },
    {
      id: 't-2',
      title: `[${projectId.slice(0, 5)}] JWT 시큐리티 적용`,
      assignee_id: 'user-2',
      status: 'BACKEND',
      assignee: '박보안',
    },
    {
      id: 't-3',
      title: `[${projectId.slice(0, 5)}] 로그인 페이지 UI`,
      assignee_id: 'user-3',
      status: 'FRONTEND',
      assignee: '이디자인',
    },
    {
      id: 't-4',
      title: `[${projectId.slice(0, 5)}] EKS 클러스터 구성`,
      assignee_id: 'user-4',
      status: 'DEVOPS',
      assignee: '최데브옵스',
    },
    {
      id: 't-5',
      title: `[${projectId.slice(0, 5)}] API 배포 완료`,
      assignee_id: 'user-1',
      status: 'DONE',
      assignee: '김개발',
    },
    {
      id: 't-6',
      title: `[${projectId.slice(0, 5)}] UI QA 피드백`,
      assignee_id: 'user-3',
      status: 'FRONTEND',
      assignee: '이디자인',
    },
  ];

  return [
    {
      id: 'BACKEND',
      title: '백엔드 (Backend)',
      tasks: baseTasks.filter((t) => t.status === 'BACKEND'),
    },
    {
      id: 'FRONTEND',
      title: '프론트엔드 (Frontend)',
      tasks: baseTasks.filter((t) => t.status === 'FRONTEND'),
    },
    {
      id: 'DEVOPS',
      title: '인프라 (DevOps)',
      tasks: baseTasks.filter((t) => t.status === 'DEVOPS'),
    },
    { id: 'DONE', title: '완료 (Done)', tasks: baseTasks.filter((t) => t.status === 'DONE') },
  ];
};
// ----------------------------------------------------

const MainDashboard: React.FC<MainDashboardProps> = ({
  onLogout,
  // currentGroupId,
  accessToken,
}) => {
  const { theme } = useTheme();

  // --- 3. 상태 관리 (API 연동) ---
  const [workspaces, setWorkspaces] = useState<WorkspaceResponse[]>([]);
  const [projects, setProjects] = useState<ProjectResponse[]>([]);
  const [columns, setColumns] = useState<Column[]>([]);

  const [selectedWorkspace, setSelectedWorkspace] = useState<WorkspaceResponse | null>(null);
  const [selectedProject, setSelectedProject] = useState<ProjectResponse | null>(null);

  const [userProfile, setUserProfile] = useState<UserProfile>({
    name: 'Mock User',
    email: 'mock@wealist.com',
    avatar: 'P',
  });

  // UI 상태
  const [isLoadingData, setIsLoadingData] = useState<boolean>(true);
  const [dataError, setDataError] = useState<string | null>(null);

  const [showUserMenu, setShowUserMenu] = useState<boolean>(false);
  const [showWorkspaceMenu, setShowWorkspaceMenu] = useState<boolean>(false);
  const [showMobileMenu, setShowMobileMenu] = useState<boolean>(false);
  const [showUserProfile, setShowUserProfile] = useState<boolean>(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [workspaceSearchQuery, setWorkspaceSearchQuery] = useState('');

  // 💡 새 상태: 선택된 컬럼 모달
  const [selectedColumn, setSelectedColumn] = useState<Column | null>(null);

  // --- 4. 데이터 연동 (useEffect 연쇄) ---

  const fetchProjectData = useCallback(
    async (workspaceId: string) => {
      setIsLoadingData(true);
      try {
        console.log(`[Phase 2] 🚀 프로젝트 로드 시작 (Workspace: ${workspaceId})`);
        const projectListResponse = await mockFetchProjects(workspaceId, accessToken);
        const projectList = projectListResponse.items || [];

        setProjects(projectList);
        if (projectList.length > 0) {
          setSelectedProject(projectList[0]);
        } else {
          setSelectedProject(null);
        }
      } catch (err) {
        console.error('Project Load Failed:', err);
        setDataError('프로젝트 목록을 불러오지 못했습니다.');
      } finally {
        setIsLoadingData(false);
      }
    },
    [accessToken],
  );

  const initDataFetch = useCallback(async () => {
    setIsLoadingData(true);
    setDataError(null);
    try {
      setUserProfile({
        name: 'Mock User',
        email: 'mock@wealist.com',
        avatar: 'P',
      });

      const workspaceListResponse = await mockFetchWorkspaces(accessToken);
      const loadedWorkspaces = workspaceListResponse.items || [];
      setWorkspaces(loadedWorkspaces);

      if (loadedWorkspaces.length > 0) {
        const defaultWorkspace = loadedWorkspaces[0];
        setSelectedWorkspace(defaultWorkspace);
      }
    } catch (err) {
      console.error('❌ API Data Fetch failed:', err);
      setDataError('초기 데이터 로드에 실패했습니다. (Kanban API Mock)');
    } finally {
      setIsLoadingData(false);
    }
  }, [accessToken]);

  useEffect(() => {
    initDataFetch();
  }, [initDataFetch]);

  // Workspace 선택 변경 시 Project 데이터 다시 로드
  useEffect(() => {
    if (selectedWorkspace) {
      // 현재 프로젝트가 선택된 Workspace의 ID와 다를 때만 fetchProjectData를 실행
      if (!selectedProject || selectedProject.workspace_id !== selectedWorkspace.id) {
        fetchProjectData(selectedWorkspace.id);
      }
    }
  }, [selectedWorkspace, fetchProjectData, selectedProject]);

  // 프로젝트 변경 시 칸반 보드 리로드
  useEffect(() => {
    if (!selectedProject) {
      setColumns([]);
      return;
    }

    console.log(`[Phase 3] 🔄 프로젝트 변경 감지: ${selectedProject.name}. 칸반 보드 로드 시작.`);

    setIsLoadingData(true);
    mockFetchKanbanBoard(selectedProject.id, accessToken)
      .then((data) => {
        setColumns(data);
      })
      .catch((err) => console.error('칸반 보드 로드 실패', err))
      .finally(() => setIsLoadingData(false));
  }, [selectedProject, accessToken]);

  // 💡 5. 컬럼 업데이트 핸들러 구현 (모달에서 호출됨)
  const handleColumnUpdate = (updatedColumn: Column) => {
    // 💡 Mock: 서버에 저장하는 로직 없이, 로컬 상태만 즉시 갱신합니다.
    setColumns((prev) => prev.map((col) => (col.id === updatedColumn.id ? updatedColumn : col)));
    // 💡 TODO: 실제 API 연동 시, 여기서 컬럼 업데이트 API (PATCH /api/projects/{id}/columns/{id}) 호출 필요
    console.log(`[Mock] 컬럼 제목 업데이트 완료: ${updatedColumn.title}`);
  };

  // --- 6. 드래그 앤 드롭 로직 (Mock 데이터 기준) ---
  const [draggedTask, setDraggedTask] = useState<Task | null>(null);
  const [draggedFromColumn, setDraggedFromColumn] = useState<string | null>(null);

  const handleDragStart = (task: Task, columnId: string): void => {
    setDraggedTask(task);
    setDraggedFromColumn(columnId);
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
  };

  const handleDrop = (targetColumnId: string): void => {
    if (!draggedTask || !draggedFromColumn || draggedFromColumn === targetColumnId) return;

    const updatedTask: Task = {
      ...draggedTask,
      status: targetColumnId,
      assignee: draggedTask.assignee,
    };

    const newColumns = columns.map((col) => {
      if (col.id === draggedFromColumn) {
        return { ...col, tasks: col.tasks.filter((t) => t.id !== draggedTask.id) };
      }
      if (col.id === targetColumnId) {
        return { ...col, tasks: [...col.tasks, updatedTask] };
      }
      return col;
    });

    setColumns(newColumns);
    setDraggedTask(null);
    setDraggedFromColumn(null);

    console.log(`[Mock] API: Task ${draggedTask.id} 상태를 ${targetColumnId}(으)로 변경 요청`);
  };

  const columnColors = ['bg-blue-500', 'bg-yellow-500', 'bg-purple-500', 'bg-green-500'];

  // --- 7. Workspace 검색 필터링 로직 ---
  const filteredWorkspaces = workspaces.filter((ws) => {
    if (!workspaceSearchQuery.trim()) {
      return true;
    }
    const query = workspaceSearchQuery.toLowerCase();
    return ws.name.toLowerCase().includes(query) || ws.id.toLowerCase().includes(query);
  });

  // --- 로딩/에러 화면 ---
  if (isLoadingData && columns.length === 0) {
    return (
      <div className={`min-h-screen ${theme.colors.background} flex items-center justify-center`}>
        <div className="p-8">
          <p className={`${theme.font.size.xl} ${theme.colors.text}`}>
            데이터를 로드하는 중입니다... 🚀
          </p>
        </div>
      </div>
    );
  }

  if (dataError) {
    return (
      <div
        className={`min-h-screen ${theme.colors.background} flex items-center justify-center text-center p-8`}
      >
        <div className="p-8 rounded-lg shadow-lg border">
          <h1 className={`${theme.font.size.xl} ${theme.colors.danger} mb-4`}>데이터 로드 실패</h1>
          <p className={`${theme.colors.subText} mb-6`}>{dataError}</p>
          <button
            onClick={initDataFetch}
            className={`bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 transition`}
          >
            다시 시도
          </button>
          <button
            onClick={onLogout}
            className={`bg-gray-400 text-white px-4 py-2 rounded hover:bg-gray-500 transition ml-2`}
          >
            로그아웃
          </button>
        </div>
      </div>
    );
  }

  // --- 8. UI 렌더링 ---
  return (
    <div className={`min-h-screen ${theme.colors.background}`}>
      {/* 백그라운드 패턴 */}
      <div
        className="fixed inset-0 opacity-5"
        style={{
          backgroundImage:
            'linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
      ></div>

      {/* 헤더 */}
      <header
        className={`${theme.colors.primary} ${theme.effects.borderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0 px-3 sm:px-6 py-2 sm:py-4 relative z-20`}
        style={{ boxShadow: theme.effects.headerShadow }}
      >
        <div className="flex items-center justify-between relative z-10">
          <div className="flex items-center gap-2 sm:gap-4">
            {/* 조직(Workspace) 선택 드롭다운 */}
            <div className="relative hidden md:block">
              <button
                onClick={() => setShowWorkspaceMenu(!showWorkspaceMenu)}
                className={`relative flex items-center gap-2 px-3 sm:px-4 py-1 sm:py-2 ${theme.colors.secondary} ${theme.effects.cardBorderWidth} ${theme.colors.border} hover:bg-gray-100 transition ${theme.font.size.xs} ${theme.effects.borderRadius}`}
              >
                <Menu className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
                <span className="hidden lg:inline font-bold">
                  {selectedWorkspace?.name || '조직 선택'}
                </span>
                <ChevronDown className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
              </button>
              {showWorkspaceMenu && (
                <div
                  className={`absolute top-full left-0 mt-2 w-64 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius}`}
                  style={{ boxShadow: theme.effects.shadow }}
                >
                  <div className="p-2 border-b">
                    <div className="relative">
                      <input
                        type="text"
                        placeholder="조직(Workspace) 검색..."
                        value={workspaceSearchQuery}
                        onChange={(e) => setWorkspaceSearchQuery(e.target.value)}
                        className={`w-full px-3 py-2 pl-8 ${theme.colors.secondary} ${theme.font.size.sm} rounded-md border`}
                      />
                      <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                    </div>
                  </div>
                  <div className="max-h-60 overflow-y-auto">
                    {filteredWorkspaces.length > 0 ? (
                      filteredWorkspaces.map((workspace) => (
                        <button
                          key={workspace.id}
                          onClick={() => {
                            setSelectedWorkspace(workspace); // 💡 조직(Workspace) 변경
                            setShowWorkspaceMenu(false);
                            setWorkspaceSearchQuery('');
                          }}
                          className={`w-full px-3 sm:px-4 py-2 sm:py-3 text-left hover:bg-blue-50 transition last:border-b-0 ${
                            theme.font.size.xs
                          } ${
                            selectedWorkspace?.id === workspace.id ? 'bg-blue-100 font-bold' : ''
                          }`}
                        >
                          {workspace.name}
                        </button>
                      ))
                    ) : (
                      <p className={`p-3 text-center ${theme.colors.subText}`}>
                        {workspaces.length > 0
                          ? '검색 결과가 없습니다.'
                          : '로드된 조직이 없습니다.'}
                      </p>
                    )}
                  </div>
                  <div
                    className={`${theme.effects.cardBorderWidth} ${theme.colors.border} border-b-0 border-l-0 border-r-0`}
                  ></div>
                  <button
                    className={`w-full px-3 sm:px-4 py-2 sm:py-3 text-left ${theme.colors.success} text-white ${theme.colors.successHover} transition flex items-center gap-2 ${theme.font.size.xs} ${theme.effects.borderRadius} rounded-t-none`}
                  >
                    <Plus className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
                    새로운 조직
                  </button>
                </div>
              )}
            </div>

            <button
              onClick={() => setShowMobileMenu(!showMobileMenu)}
              className={`md:hidden relative ${theme.colors.secondary} ${theme.effects.cardBorderWidth} ${theme.colors.border} p-2 ${theme.effects.borderRadius}`}
            >
              <Menu className="w-5 h-5" style={{ strokeWidth: 3 }} />
            </button>
          </div>

          {/* 사용자 메뉴 */}
          <div className="relative">
            <button
              onClick={() => setShowUserMenu(!showUserMenu)}
              className={`relative flex items-center gap-1 sm:gap-2 px-2 sm:px-4 py-1 sm:py-2 ${theme.colors.secondary} ${theme.effects.cardBorderWidth} ${theme.colors.border} hover:bg-gray-100 transition ${theme.font.size.xs} ${theme.effects.borderRadius}`}
            >
              <User className="w-4 h-4 sm:w-5 sm:h-5" style={{ strokeWidth: 3 }} />
              <span className="hidden sm:inline">{userProfile.name}</span>
              <ChevronDown className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
            </button>
            {showUserMenu && (
              <div
                className={`absolute top-full right-0 mt-2 w-48 sm:w-56 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius}`}
                style={{ boxShadow: theme.effects.shadow }}
              >
                <div
                  className={`px-3 sm:px-4 py-2 sm:py-3 ${theme.effects.cardBorderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0 ${theme.colors.primary} text-white`}
                >
                  <p className={`font-bold ${theme.font.size.xs}`}>{userProfile.email}</p>
                </div>
                <button
                  onClick={() => {
                    setShowUserProfile(true);
                    setShowUserMenu(false);
                  }}
                  className={`w-full px-3 sm:px-4 py-2 sm:py-3 text-left hover:bg-blue-50 transition ${theme.effects.cardBorderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0 ${theme.font.size.xs}`}
                >
                  프로필
                </button>
                <div
                  className={`${theme.effects.cardBorderWidth} ${theme.colors.border} border-b-0 border-l-0 border-r-0`}
                ></div>
                <button
                  className={`w-full px-3 sm:px-4 py-2 sm:py-3 text-left bg-red-500 hover:bg-red-600 transition text-white ${theme.font.size.xs} ${theme.effects.borderRadius} rounded-t-none`}
                  onClick={onLogout}
                >
                  로그아웃
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* 모바일 메뉴 */}
      {showMobileMenu && (
        <div
          className="md:hidden fixed inset-0 bg-black bg-opacity-50 z-50"
          onClick={() => setShowMobileMenu(false)}
        >
          <div
            className={`${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} w-64 h-full p-4`}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-4">
              <h2 className={`${theme.font.size.xs} font-bold`}>메뉴</h2>
              <button
                onClick={() => setShowMobileMenu(false)}
                className={`bg-red-500 ${theme.effects.cardBorderWidth} ${theme.colors.border} p-1`}
              >
                <X className="w-4 h-4 text-white" style={{ strokeWidth: 3 }} />
              </button>
            </div>
            <div className="space-y-2">
              <p className={`text-[8px] ${theme.colors.subText} mb-2`}>조직(Workspaces):</p>
              {workspaces.map((workspace) => (
                <button
                  key={workspace.id}
                  onClick={() => {
                    setSelectedWorkspace(workspace); // 💡 조직(Workspace) 변경
                    setShowMobileMenu(false);
                  }}
                  className={`w-full px-3 py-2 text-left ${theme.effects.cardBorderWidth} ${
                    theme.colors.border
                  } text-[8px] ${theme.effects.borderRadius} ${
                    selectedWorkspace?.id === workspace.id
                      ? `${theme.colors.primary} text-white`
                      : `${theme.colors.secondary} hover:bg-gray-100`
                  }`}
                >
                  {workspace.name}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* 프로젝트 탭 바 */}
      <div
        className={`${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0 px-3 sm:px-6 py-2 sm:py-3 overflow-x-auto`}
      >
        <div className="flex items-center gap-2 sm:gap-4 min-w-max">
          <div className="flex gap-2 flex-nowrap">
            {projects.map((project) => (
              <div key={project.id} className="relative flex-shrink-0">
                <button
                  onClick={() => setSelectedProject(project)} // 💡 프로젝트 변경
                  className={`relative px-2 sm:px-4 py-1 sm:py-2 ${theme.effects.cardBorderWidth} ${
                    theme.colors.border
                  } transition ${theme.font.size.xs} ${
                    theme.effects.borderRadius
                  } whitespace-nowrap ${
                    selectedProject?.id === project.id
                      ? `${theme.colors.primary} text-white`
                      : `${theme.colors.secondary} ${theme.colors.text} hover:bg-gray-100`
                  }`}
                >
                  {project.name}
                </button>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* 칸반 보드 */}
      <div className="p-3 sm:p-6 relative z-10">
        <div className="flex flex-col lg:flex-row gap-3 sm:gap-4 lg:overflow-x-auto pb-4">
          {columns.map((column, idx) => (
            <div
              key={column.id}
              onDragOver={handleDragOver}
              onDrop={() => handleDrop(column.id)}
              className="w-full lg:w-80 lg:flex-shrink-0 relative"
            >
              <div
                className={`relative ${theme.effects.cardBorderWidth} ${theme.colors.border} p-3 sm:p-4 ${theme.colors.card} ${theme.effects.borderRadius}`}
              >
                <div
                  className={`flex items-center justify-between mb-3 sm:mb-4 pb-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0`}
                >
                  {/* 💡 컬럼 제목 클릭 시 모달 열기 */}
                  <h3
                    onClick={() => setSelectedColumn(column)}
                    className={`font-bold ${theme.colors.text} flex items-center gap-2 ${theme.font.size.xs} cursor-pointer hover:underline`}
                  >
                    <span
                      className={`w-3 h-3 sm:w-4 sm:h-4 ${
                        columnColors[idx % columnColors.length]
                      } ${theme.effects.cardBorderWidth} ${theme.colors.border}`}
                    ></span>
                    {column.title}
                    <span
                      className={`bg-black text-white px-1 sm:px-2 py-1 ${theme.effects.cardBorderWidth} ${theme.colors.border} text-[8px] sm:text-xs`}
                    >
                      {column.tasks.length}
                    </span>
                  </h3>
                  <button className={`${theme.colors.text} hover:text-blue-500`}>
                    <MoreVertical className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
                  </button>
                </div>

                <div className="space-y-2 sm:space-y-3">
                  {column.tasks.map((task) => (
                    <div key={task.id} className="relative">
                      <div
                        draggable
                        onDragStart={() => handleDragStart(task, column.id)}
                        onClick={() => setSelectedTask(task)} // 💡 태스크 클릭 시 모달 열기
                        className={`relative ${theme.colors.card} p-3 sm:p-4 ${theme.effects.cardBorderWidth} ${theme.colors.border} hover:border-blue-500 transition cursor-pointer ${theme.effects.borderRadius}`}
                      >
                        <h4
                          className={`font-bold ${theme.colors.text} mb-2 sm:mb-3 ${theme.font.size.xs} break-words`}
                        >
                          {task.title}
                        </h4>
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-6 h-6 sm:w-8 sm:h-8 ${theme.colors.primary} ${theme.effects.cardBorderWidth} ${theme.colors.border} flex items-center justify-center text-white font-bold text-[8px] sm:text-xs flex-shrink-0 ${theme.effects.borderRadius}`}
                          >
                            {task.assignee_id ? task.assignee_id[0].toUpperCase() : '?'}
                          </div>
                          <span className={`${theme.font.size.xs} truncate ${theme.colors.text}`}>
                            {task.assignee || '미배정'}
                          </span>
                        </div>
                      </div>
                    </div>
                  ))}
                  <div className="relative">
                    <button
                      className={`relative w-full py-3 sm:py-4 ${theme.effects.cardBorderWidth} border-dashed ${theme.colors.border} ${theme.colors.card} hover:bg-gray-100 transition flex items-center justify-center gap-2 ${theme.font.size.xs} ${theme.effects.borderRadius}`}
                      onClick={() =>
                        setSelectedTask({
                          id: '',
                          title: '',
                          assignee_id: '',
                          status: 'NEW',
                          assignee: '',
                        })
                      }
                    >
                      <Plus className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
                      태스크 추가
                    </button>
                  </div>
                </div>
              </div>
            </div>
          ))}
          <div className="w-full lg:w-80 lg:flex-shrink-0 relative">
            <button
              className={`relative w-full h-24 sm:h-32 ${theme.effects.cardBorderWidth} border-dashed ${theme.colors.border} ${theme.colors.card} hover:bg-gray-100 transition flex items-center justify-center gap-2 ${theme.font.size.xs} ${theme.effects.borderRadius}`}
            >
              <Plus className="w-4 h-4 sm:w-5 sm:h-5" style={{ strokeWidth: 3 }} />
              새로운 티켓
            </button>
          </div>
        </div>
      </div>

      {/* 모달 */}
      {showUserProfile && userProfile && (
        <UserProfileModal user={userProfile} onClose={() => setShowUserProfile(false)} />
      )}
      {selectedTask && (
        <TaskDetailModal task={selectedTask} onClose={() => setSelectedTask(null)} />
      )}

      {/* 💡 Column Detail Modal 렌더링 */}
      {selectedColumn && (
        <ColumnDetailModal
          column={selectedColumn}
          onClose={() => setSelectedColumn(null)}
          onUpdate={handleColumnUpdate} // 💡 갱신 핸들러 연결
        />
      )}
    </div>
  );
};

export default MainDashboard;
