import React, { useEffect, useState, useCallback, useRef } from 'react';
import {
  // Menu,
  // User,
  ChevronDown,
  Plus,
  MoreVertical,
  X,
  // Search,
  Home,
  Bell,
  MessageSquare,
  Briefcase,
  Settings,
  // Edit,
  // AlignLeft,
  File,
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import UserProfileModal from '../components/modals/UserProfileModal';
import TaskDetailModal from '../components/modals/TaskDetailModal';
import ColumnDetailModal from '../components/modals/ColumnDetailModal';
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

const MainDashboard: React.FC<MainDashboardProps> = ({ onLogout, accessToken }) => {
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
  // const [isLoadingData, setIsLoadingData] = useState<boolean>(true);
  // const [dataError, setDataError] = useState<string | null>(null);

  const [showUserMenu, setShowUserMenu] = useState<boolean>(false);
  const [showWorkspaceSelector, setShowWorkspaceSelector] = useState<boolean>(false);
  const [showProjectSelector, setShowProjectSelector] = useState<boolean>(false);
  const [showUserProfile, setShowUserProfile] = useState<boolean>(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [selectedColumn, setSelectedColumn] = useState<Column | null>(null);

  // Ref for Menu/Selector
  const userMenuRef = useRef<HTMLDivElement>(null);
  const workspaceSelectorRef = useRef<HTMLDivElement>(null);
  const projectSelectorRef = useRef<HTMLDivElement>(null);

  // --- 4. 데이터 연동 (useEffect 연쇄) ---
  const fetchProjectData = useCallback(
    async (workspaceId: string) => {
      // setIsLoadingData(true);
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
        // setDataError('프로젝트 목록을 불러오지 못했습니다.');
      } finally {
        // setIsLoadingData(false);
      }
    },
    [accessToken],
  );

  const initDataFetch = useCallback(async () => {
    // setIsLoadingData(true);
    // setDataError(null);
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
      // setDataError('초기 데이터 로드에 실패했습니다. (Kanban API Mock)');
    } finally {
      // setIsLoadingData(false);
    }
  }, [accessToken]);

  useEffect(() => {
    initDataFetch();
  }, [initDataFetch]);

  useEffect(() => {
    if (selectedWorkspace) {
      if (!selectedProject || selectedProject.workspace_id !== selectedWorkspace.id) {
        fetchProjectData(selectedWorkspace.id);
      }
    }
  }, [selectedWorkspace, selectedProject, fetchProjectData]);

  useEffect(() => {
    if (!selectedProject) {
      setColumns([]);
      return;
    }

    console.log(`[Phase 3] 🔄 프로젝트 변경 감지: ${selectedProject.name}. 칸반 보드 로드 시작.`);

    // setIsLoadingData(true);
    mockFetchKanbanBoard(selectedProject.id, accessToken)
      .then((data) => {
        setColumns(data);
      })
      .catch((err) => console.error('칸반 보드 로드 실패', err))
      .finally(() => {
        // setIsLoadingData(false);
      });
  }, [selectedProject, accessToken]);

  const handleColumnUpdate = (updatedColumn: Column) => {
    setColumns((prev) => prev.map((col) => (col.id === updatedColumn.id ? updatedColumn : col)));
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

  // --- 외부 클릭 감지 (모달/메뉴 닫기) ---
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      // User Menu
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setShowUserMenu(false);
      }
      // Workspace Selector
      if (
        showWorkspaceSelector &&
        workspaceSelectorRef.current &&
        !workspaceSelectorRef.current.contains(event.target as Node)
      ) {
        const workspaceLogoButton = document.getElementById('workspace-logo-button');
        if (workspaceLogoButton && workspaceLogoButton.contains(event.target as Node)) {
          return;
        }
        setShowWorkspaceSelector(false);
      }
      // Project Selector
      if (
        showProjectSelector &&
        projectSelectorRef.current &&
        !projectSelectorRef.current.contains(event.target as Node)
      ) {
        setShowProjectSelector(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showWorkspaceSelector, showProjectSelector]);

  const currentWorkspaceInitial = selectedWorkspace?.name.slice(0, 1) || 'W';
  const sidebarWidth = 'w-16 sm:w-20'; // 사이드바 너비 정의 (예: w-20 = 5rem = 80px)

  // --- 8. UI 렌더링 ---
  return (
    <div className={`min-h-screen flex ${theme.colors.background} relative`}>
      {/* 백그라운드 패턴 (전체 배경에 적용) */}
      <div
        className="fixed inset-0 opacity-5"
        style={{
          backgroundImage:
            'linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
      ></div>

      {/* 1. 💡 왼쪽 사이드바 (Fixed Navigation) */}
      <aside
        className={`${sidebarWidth} fixed top-0 left-0 h-full flex flex-col justify-between ${theme.colors.primary} text-white shadow-xl z-50 flex-shrink-0`}
      >
        <div className="flex flex-col flex-grow items-center">
          {/* 1-1. 최상단 Workspace 로고 (Selector Trigger) */}
          <div className={`py-3 flex justify-center w-full relative`}>
            <button
              id="workspace-logo-button"
              onClick={() => setShowWorkspaceSelector(!showWorkspaceSelector)}
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center text-xl font-bold transition 
                    bg-white text-blue-800 ring-2 ring-white/50 hover:opacity-90`}
              title={selectedWorkspace?.name || '워크스페이스 선택'}
            >
              {currentWorkspaceInitial}
            </button>
          </div>

          {/* 1-2. 고정 내비게이션 아이콘 (홈, DM, 알림, 파일) - Opacity 적용 */}
          <div className="flex flex-col gap-2 mt-4 flex-grow px-2 w-full pt-4">
            {/* 홈 (기본 대시보드) - 활성화 상태 */}
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center transition bg-blue-600 text-white ring-2 ring-white/50`}
              title="홈 (대시보드)"
            >
              <Home className="w-6 h-6" />
            </button>

            {/* DM (비활성) */}
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="DM"
            >
              <MessageSquare className="w-6 h-6" />
            </button>

            {/* 알림 (비활성) */}
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="알림"
            >
              <Bell className="w-6 h-6" />
            </button>

            {/* 파일 (비활성) */}
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="파일"
            >
              <File className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* 1-3. 계정/유저 메뉴 (하단) */}
        <div className={`py-3 px-2 border-t border-blue-700`}>
          <button
            onClick={() => setShowUserMenu(!showUserMenu)}
            className={`w-full flex items-center justify-center py-2 text-sm rounded-lg hover:bg-blue-600 transition relative`}
            title="계정 메뉴"
          >
            <div
              className={`w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center text-sm font-bold ring-2 ring-white/50 text-gray-700`}
            >
              {userProfile.avatar}
            </div>
          </button>
        </div>
      </aside>

      {/* 2. Workspace Selector Overlay (Fixed) */}
      {showWorkspaceSelector && (
        <>
          <div
            onClick={() => setShowWorkspaceSelector(false)}
            className="absolute inset-0 bg-black opacity-30 z-40"
          />
          <div
            ref={workspaceSelectorRef}
            className={`fixed top-0 left-16 sm:left-20 h-full w-72 ${theme.colors.card} border-r ${theme.colors.border} z-50 transition-transform duration-300 ease-out`}
            style={{ boxShadow: theme.effects.shadow }}
          >
            <div className="p-4 flex flex-col h-full">
              <div className="flex items-center justify-between pb-3 border-b mb-3">
                <h2 className="font-bold text-lg text-black">워크스페이스</h2>
                <button
                  onClick={() => setShowWorkspaceSelector(false)}
                  className="text-gray-500 hover:text-gray-700"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <div className="overflow-y-auto flex-grow">
                <h3 className="text-xs text-gray-400 mb-2 px-2 font-semibold">내 워크스페이스</h3>
                <div className="space-y-1">
                  {workspaces.map((workspace) => (
                    <button
                      key={workspace.id}
                      onClick={() => {
                        setSelectedWorkspace(workspace);
                        setShowWorkspaceSelector(false);
                      }}
                      className={`w-full px-3 py-2 text-left text-sm rounded transition flex items-center gap-2 ${
                        selectedWorkspace?.id === workspace.id
                          ? 'bg-blue-100 text-blue-700 font-semibold'
                          : 'hover:bg-gray-100 text-gray-800'
                      }`}
                    >
                      <span className="w-6 h-6 rounded bg-gray-300 flex items-center justify-center text-xs font-bold text-gray-700">
                        {workspace.name.slice(0, 1)}
                      </span>
                      {workspace.name}
                    </button>
                  ))}
                </div>
              </div>

              <div className="pt-3 border-t mt-3">
                <button className="w-full px-3 py-2 text-left text-blue-500 text-sm flex items-center gap-2 hover:bg-gray-100 rounded transition">
                  <Plus className="w-4 h-4" /> 새로운 워크스페이스 생성
                </button>
                <button className="w-full px-3 py-2 text-left text-sm flex items-center gap-2 hover:bg-gray-100 rounded transition text-gray-700">
                  <Settings className="w-4 h-4 text-gray-500" /> 워크스페이스 설정
                </button>
              </div>
            </div>
          </div>
        </>
      )}

      {/* 3. 메인 콘텐츠 영역 (사이드바만큼 margin/padding으로 공간 확보) */}
      <div
        className="flex-grow flex flex-col relative z-10"
        style={{ marginLeft: sidebarWidth, minHeight: '100vh' }}
      >
        {/* 3-1. Header (Fixed, 전체 100% width of main content area) */}
        <header
          className={`fixed top-0 left-0 h-16 flex items-center justify-between px-6 sm:px-28 py-2 sm:py-3 ${theme.colors.card} shadow-md z-20 w-full`}
          style={{
            boxShadow: theme.effects.headerShadow,
            width: `calc(100% - ${sidebarWidth})`,
            left: sidebarWidth,
          }}
        >
          <div className="flex items-center gap-2 relative">
            {/* 💡 프로젝트 선택 드롭다운 버튼 */}
            <button
              onClick={() => setShowProjectSelector(!showProjectSelector)}
              className={`flex items-center gap-2 font-bold text-xl ${theme.colors.text} hover:opacity-80 transition`}
            >
              {selectedProject?.name || '프로젝트를 선택하세요'}
              <ChevronDown
                className={`w-5 h-5 text-gray-500 transition-transform ${
                  showProjectSelector ? 'rotate-180' : 'rotate-0'
                }`}
                style={{ strokeWidth: 2.5 }}
              />
            </button>

            {/* 💡 프로젝트 선택 오버레이 */}
            {showProjectSelector && (
              <div
                ref={projectSelectorRef}
                className={`absolute top-full -left-4 mt-3 w-80 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius}`}
                style={{ boxShadow: theme.effects.shadow }}
              >
                <div className="p-3 max-h-80 overflow-y-auto">
                  <h3 className="text-xs text-gray-400 mb-2 px-1 font-semibold">
                    내 프로젝트 ({projects.length})
                  </h3>
                  {projects.map((project) => (
                    <button
                      key={project.id}
                      onClick={() => {
                        setSelectedProject(project);
                        setShowProjectSelector(false);
                      }}
                      className={`w-full px-3 py-2 text-left text-sm rounded transition truncate ${
                        selectedProject?.id === project.id
                          ? 'bg-blue-100 text-blue-700 font-semibold'
                          : 'hover:bg-gray-100 text-gray-800'
                      }`}
                    >
                      # {project.name}
                    </button>
                  ))}
                </div>
                <div className="pt-2 border-t">
                  <button className="w-full px-3 py-2 text-left text-sm flex items-center gap-2 text-blue-500 hover:bg-gray-100 rounded-b-lg transition">
                    <Plus className="w-4 h-4" /> 새 프로젝트 생성
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* 우측의 검색 및 추가 버튼은 제거됨 */}
        </header>

        {/* 3-2. 칸반 보드 (스크롤 가능한 메인 콘텐츠) */}
        {/* mt-16을 통해 Fixed Header 높이(h-16) 만큼 공간 확보 */}
        <div className="flex-grow flex flex-col p-3 sm:p-6 overflow-auto mt-16 ml-20">
          {selectedProject ? (
            <div className="flex flex-col lg:flex-row gap-3 sm:gap-4 min-w-max pb-4">
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
                        <MoreVertical
                          className="w-3 h-3 sm:w-4 sm:h-4"
                          style={{ strokeWidth: 3 }}
                        />
                      </button>
                    </div>

                    <div className="space-y-2 sm:space-y-3">
                      {column.tasks.map((task) => (
                        <div key={task.id} className="relative">
                          <div
                            draggable
                            onDragStart={() => handleDragStart(task, column.id)}
                            onClick={() => setSelectedTask(task)}
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
                              <span
                                className={`${theme.font.size.xs} truncate ${theme.colors.text}`}
                              >
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
                          칸반 추가
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-full text-center p-8">
              <Briefcase className="w-16 h-16 mb-4 text-gray-400" />
              <h2 className={`${theme.font.size.xl} ${theme.colors.text} mb-2`}>
                프로젝트를 선택하세요
              </h2>
              <p className={`${theme.colors.subText}`}>
                왼쪽 워크스페이스 메뉴를 통해 프로젝트 목록을 불러오고 선택하세요.
              </p>
            </div>
          )}
        </div>
      </div>

      {/* 모달 */}
      {showUserMenu && (
        <div
          ref={userMenuRef}
          className={`absolute bottom-20 left-20 w-48 sm:w-56 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius}`}
          style={{ boxShadow: theme.effects.shadow }}
        >
          <div className={`px-3 sm:px-4 py-2 sm:py-3 ${theme.colors.primary} text-white`}>
            <p className={`font-bold ${theme.font.size.xs}`}>{userProfile.email}</p>
          </div>
          <button
            onClick={() => {
              setShowUserProfile(true);
              setShowUserMenu(false);
            }}
            className={`w-full px-3 sm:px-4 py-2 sm:py-3 text-left hover:bg-gray-100 transition text-black text-sm`}
          >
            프로필
          </button>
          <button
            className={`w-full px-3 sm:px-4 py-2 sm:py-3 text-left bg-red-500 hover:bg-red-600 transition text-white text-sm rounded-b-lg`}
            onClick={onLogout}
          >
            로그아웃
          </button>
        </div>
      )}
      {showUserProfile && userProfile && (
        <UserProfileModal user={userProfile} onClose={() => setShowUserProfile(false)} />
      )}
      {selectedTask && (
        <TaskDetailModal task={selectedTask} onClose={() => setSelectedTask(null)} />
      )}

      {selectedColumn && (
        <ColumnDetailModal
          column={selectedColumn}
          onClose={() => setSelectedColumn(null)}
          onUpdate={handleColumnUpdate}
        />
      )}
    </div>
  );
};

export default MainDashboard;
