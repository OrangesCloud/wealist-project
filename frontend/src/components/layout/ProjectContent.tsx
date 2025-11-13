// src/components/layout/ProjectContent.tsx

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Plus, Settings, Briefcase, ArrowUp, ArrowDown } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { getDefaultColorByIndex } from '../../constants/colors';
import { AssigneeAvatarStack } from '../common/AvartarStack';
import {
  CustomRoleResponse,
  CustomImportanceResponse,
  CustomStageResponse,
  ProjectResponse,
  BoardResponse,
  Column,
  ViewState,
  FieldOptionsLookup,
} from '../../types/board';
import { getBoards } from '../../api/board/boardService';
import { MOCK_IMPORTANCES, MOCK_ROLES, MOCK_STAGES } from '../../mocks/board';
import { BoardDetailModal } from '../modals/board/BoardDetailModal';
import { FilterBar } from '../modals/board/FilterBar';

interface ProjectContentProps {
  // Data
  selectedProject: ProjectResponse;
  workspaceId: string;
  fieldOptionsLookup: FieldOptionsLookup;

  // Handlers
  onProjectContentUpdate: () => void;
  onManageModalOpen: () => void;

  // Initial States for Modals
  onEditBoard: (data: any) => void;

  // 💡 [추가] MainDashboard에서 모달 상태를 넘겨받음
  showCreateBoard: boolean;
  setShowCreateBoard: (show: boolean) => void;
}

export const ProjectContent: React.FC<ProjectContentProps> = ({
  selectedProject,
  workspaceId,
  onProjectContentUpdate,
  onManageModalOpen,
  onEditBoard,
  showCreateBoard,
  setShowCreateBoard,
}) => {
  const { theme } = useTheme();

  // 💡 [Board Data States]
  const [columns, setColumns] = useState<Column[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 💡 [Custom Field Lookups]
  const [roleOptions] = useState<CustomRoleResponse[]>(MOCK_ROLES);
  const [importanceOptions] = useState<CustomImportanceResponse[]>(MOCK_IMPORTANCES);

  // 💡 [통합된 View/Filter 상태]
  const [viewState, setViewState] = useState<ViewState>({
    currentView: 'stage',
    searchQuery: '',
    filterOption: 'all',
    currentLayout: 'board',
    showCompleted: false,
    sortColumn: null,
    sortDirection: 'asc',
  });

  // 💡 [UI States]
  const [selectedBoardId, setSelectedBoardId] = useState<string | null>(null);
  const [editBoardData, setEditBoardData] = useState<any>(null); // MainDashboard로 전달하기 위한 데이터 복사

  // Drag state
  const [draggedBoard, setDraggedBoard] = useState<BoardResponse | null>(null);
  const [draggedFromColumn, setDraggedFromColumn] = useState<string | null>(null);
  const [draggedColumn, setDraggedColumn] = useState<Column | null>(null);
  const [dragOverBoardId, setDragOverBoardId] = useState<string | null>(null);
  const [dragOverColumn, setDragOverColumn] = useState<string | null>(null);

  // 💡 [추가] View State Setter Helper
  const setViewField = useCallback(<K extends keyof ViewState>(key: K, value: ViewState[K]) => {
    setViewState((prev) => ({ ...prev, [key]: value }));
  }, []);

  // 💡 [추가] Custom Field Option Lookup Helper (유지)
  const getRoleOption = (roleId: string | undefined) =>
    roleId ? roleOptions.find((r) => r.roleId === roleId) : undefined;
  const getImportanceOption = (importanceId: string | undefined) =>
    importanceId ? importanceOptions.find((i) => i.importanceId === importanceId) : undefined;

  // 4. 보드 목록 조회 함수 (useCallback)
  const fetchBoards = useCallback(async () => {
    if (!selectedProject) {
      setColumns([]);
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      const stages = MOCK_STAGES;
      // 💡 API 호출 시 필터링을 viewState 기준으로 적용할 수 있지만, 현재는 전체 로드
      const boardsResponse = await getBoards(selectedProject.projectId);

      // 데이터 처리 로직 (유지)
      const stageMap = new Map<string, { stage: CustomStageResponse; boards: BoardResponse[] }>();
      stages.forEach((stage: CustomStageResponse) => {
        stageMap.set(stage.stageId, { stage, boards: [] });
      });

      boardsResponse?.boards?.forEach((board: BoardResponse) => {
        const stageId = board.customFields?.stageId;
        const targetStageId = stageId || MOCK_STAGES[0].stageId;

        if (stageMap.has(targetStageId)) {
          stageMap.get(targetStageId)!.boards.push(board);
        } else {
          console.warn(
            `[Board Load] 보드 ${board.boardId}에 유효하지 않은 Stage ID (${targetStageId})가 할당되었습니다.`,
          );
        }
      });

      const sortedStages = Array.from(stageMap.values()).sort(
        (a, b) => a.stage.displayOrder - b.stage.displayOrder,
      );

      const newColumns: Column[] = sortedStages.map(({ stage, boards }) => ({
        stageId: stage.stageId,
        title: stage.label,
        color: stage.color,
        boards: boards,
      }));

      setColumns(newColumns);
      onProjectContentUpdate();
    } catch (err) {
      const error = err as Error;
      console.error('❌ 보드 로드 실패:', error);
      setError(`보드 로드 실패: ${error.message}`);
      setColumns([]);
    } finally {
      setIsLoading(false);
    }
  }, [selectedProject, onProjectContentUpdate]);

  // 4.1. 프로젝트 변경 시 보드 로드 트리거
  useEffect(() => {
    fetchBoards();
  }, [fetchBoards]);

  // 5. 드래그 앤 드롭 및 정렬 로직 (useCallback 유지)

  const handleDragStart = (board: BoardResponse, columnId: string): void => {
    setDraggedBoard(board);
    setDraggedFromColumn(columnId);
  };
  const handleDragOver = (e: React.DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
  };

  const handleDragEnd = (): void => {
    setDraggedBoard(null);
    setDraggedFromColumn(null);
    setDraggedColumn(null);
    setDragOverBoardId(null);
    setDragOverColumn(null);
  };

  const handleDrop = useCallback(
    async (targetColumnId: string): Promise<void> => {
      if (!draggedBoard || !draggedFromColumn) return;

      setDragOverColumn(null);

      // Same column: reorder boards within column
      if (draggedFromColumn === targetColumnId) {
        if (!dragOverBoardId || dragOverBoardId === draggedBoard.boardId) {
          handleDragEnd();
          return;
        }

        const targetColumn = columns.find((col) => col.stageId === targetColumnId);
        if (!targetColumn) {
          handleDragEnd();
          return;
        }

        // Reorder boards (Local Optimistic Update)
        const draggedIndex = targetColumn.boards.findIndex(
          (b) => b.boardId === draggedBoard.boardId,
        );
        const targetIndex = targetColumn.boards.findIndex((b) => b.boardId === dragOverBoardId);

        if (draggedIndex !== -1 && targetIndex !== -1) {
          const newBoards = [...targetColumn.boards];
          const [removed] = newBoards.splice(draggedIndex, 1);
          newBoards.splice(targetIndex, 0, removed);
          setColumns(
            columns.map((col) =>
              col.stageId === targetColumnId ? { ...col, boards: newBoards } : col,
            ),
          );
        }
      } else {
        // Different column: Stage Change + Optimistic Update
        const updatedBoard: BoardResponse = {
          ...draggedBoard,
          customFields: { ...draggedBoard.customFields, stageId: targetColumnId },
        };

        const newColumns = columns.map((col) => {
          if (col.stageId === draggedFromColumn) {
            return { ...col, boards: col.boards.filter((t) => t.boardId !== draggedBoard.boardId) };
          }
          if (col.stageId === targetColumnId) {
            if (dragOverBoardId) {
              const targetIndex = col.boards.findIndex((b) => b.boardId === dragOverBoardId);
              if (targetIndex !== -1) {
                const newBoards = [...col.boards];
                newBoards.splice(targetIndex, 0, updatedBoard);
                return { ...col, boards: newBoards };
              }
            }
            return { ...col, boards: [...col.boards, updatedBoard] };
          }
          return col;
        });

        setColumns(newColumns);
      }

      handleDragEnd();

      console.log(`[API CALL] moveBoard 호출: ${draggedBoard.boardId} to ${targetColumnId}`);
    },
    [draggedBoard, draggedFromColumn, dragOverBoardId, columns],
  );

  const handleColumnDragStart = (column: Column): void => {
    setDraggedColumn(column);
  };
  const handleColumnDrop = useCallback(
    async (targetColumn: Column): Promise<void> => {
      if (!draggedColumn || draggedColumn.stageId === targetColumn.stageId) {
        setDraggedColumn(null);
        return;
      }

      const draggedIndex = columns.findIndex((col) => col.stageId === draggedColumn.stageId);
      const targetIndex = columns.findIndex((col) => col.stageId === targetColumn.stageId);

      if (draggedIndex !== -1 && targetIndex !== -1) {
        const newColumns = [...columns];
        const [removed] = newColumns.splice(draggedIndex, 1);
        newColumns.splice(targetIndex, 0, removed);
        setColumns(newColumns);
      }

      handleDragEnd();

      console.log(`[API CALL] updateFieldOrder 호출: Stage 순서 변경`);
    },
    [draggedColumn, columns],
  );

  // Table sorting handler (handleSort)
  const handleSort = (
    column: 'title' | 'stage' | 'role' | 'importance' | 'importance' | 'assignee' | 'dueDate',
  ) => {
    // 💡 [수정] viewState Setter 사용
    if (viewState?.sortColumn === column) {
      setViewField('sortDirection', viewState?.sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setViewField('sortColumn', column);
      setViewField('sortDirection', 'asc');
    }
  };

  // 모달에 전달할 onEdit 핸들러 래핑
  const handleBoardEdit = (boardData: any) => {
    onEditBoard(boardData);
    setSelectedBoardId(null);
  };

  // 6. Table/Board View 공통 데이터 필터링/정렬 로직 (useMemo)
  const { sortedBoards, allBoards } = useMemo(() => {
    const { searchQuery, sortColumn, sortDirection } = viewState;

    const boardsToProcess = columns.flatMap((column) =>
      column.boards.map((board) => {
        const roleId = board.customFields?.roleIds?.[0];
        const importanceId = board.customFields?.importanceId;
        return {
          ...board,
          stageName: column.title,
          stageColor: column.color,
          roleOption: getRoleOption(roleId),
          importanceOption: getImportanceOption(importanceId),
        };
      }),
    );

    const filteredBoards = searchQuery?.trim()
      ? boardsToProcess.filter((board) => {
          const query = searchQuery.toLowerCase();
          const titleMatch = board.title.toLowerCase().includes(query);
          const contentMatch = board.content?.toLowerCase().includes(query);
          return titleMatch || contentMatch;
        })
      : boardsToProcess;

    const sortedBoards = [...filteredBoards].sort((a, b) => {
      // [정렬 로직 유지]
      if (!sortColumn) return 0;

      let aValue: any;
      let bValue: any;

      // 💡 [수정] 실제 정렬 로직을 useMemo 내부에서 viewState를 사용하여 구현해야 함
      switch (sortColumn) {
        case 'title':
          aValue = a.title.toLowerCase();
          bValue = b.title.toLowerCase();
          break;
        // ... (나머지 정렬 케이스 생략)
        default:
          aValue = 0;
          bValue = 0;
          break;
      }

      if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1;
      if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1;
      return 0;
    });

    return { sortedBoards, allBoards: boardsToProcess };
  }, [columns, viewState, roleOptions, importanceOptions]); // 💡 viewState 객체를 의존성 배열에 포함

  // 로딩 상태 처리
  if (isLoading && columns.length === 0) {
    return <LoadingSpinner message="보드를 로드 중..." />;
  }

  if (error) {
    return (
      <div className="mt-4 p-4 bg-red-50 border border-red-300 rounded-lg text-red-700">
        {error}
      </div>
    );
  }

  return (
    <>
      {/* FilterBar */}
      <FilterBar
        onSearchChange={(query) => setViewField('searchQuery', query)}
        onViewChange={(view) => setViewField('currentView', view)}
        onFilterChange={(filter) => setViewField('filterOption', filter)}
        onManageClick={onManageModalOpen}
        currentView={viewState?.currentView}
        onLayoutChange={(layout) => setViewField('currentLayout', layout)}
        onShowCompletedChange={(show) => setViewField('showCompleted', show)}
        currentLayout={viewState?.currentLayout}
        showCompleted={viewState?.showCompleted}
      />

      {/* Boards or Table View */}
      {viewState?.currentLayout === 'table' ? (
        // =============================================================
        // 1. Table Layout
        // =============================================================
        <div className="mt-4 overflow-x-auto">
          <table
            className={`w-full ${theme.colors.card} ${theme.effects.borderRadius} overflow-hidden shadow-lg`}
          >
            <thead className="bg-gray-100 border-b border-gray-200">
              <tr>
                {['title', 'stage', 'role', 'importance', 'assignee', 'dueDate'].map((col) => (
                  <th key={col} className="px-4 py-3 text-left">
                    <button
                      onClick={() =>
                        handleSort(
                          col as
                            | 'title'
                            | 'stage'
                            | 'role'
                            | 'importance'
                            | 'importance'
                            | 'assignee'
                            | 'dueDate',
                        )
                      }
                      className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                    >
                      {col === 'title' && '제목'}
                      {col === 'stage' && '진행 단계'}
                      {col === 'role' && '역할'}
                      {col === 'importance' && '중요도'}
                      {col === 'assignee' && '담당자'}
                      {col === 'dueDate' && '마감일'}
                      {viewState?.sortColumn === col &&
                        (viewState?.sortDirection === 'asc' ? (
                          <ArrowUp className="w-4 h-4" />
                        ) : (
                          <ArrowDown className="w-4 h-4" />
                        ))}
                    </button>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {sortedBoards.map((board) => (
                <tr
                  key={board.boardId}
                  onClick={() => setSelectedBoardId(board.boardId)}
                  className="border-b border-gray-200 hover:bg-gray-50 cursor-pointer transition"
                >
                  {/* Title */}
                  <td className="px-4 py-3 font-semibold text-gray-800">{board.title}</td>
                  {/* Stage */}
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: board.stageColor || '#6B7280' }}
                      />
                      <span className="text-sm">{board.stageName}</span>
                    </div>
                  </td>
                  {/* Role */}
                  <td className="px-4 py-3">
                    {board.roleOption ? (
                      <div className="flex items-center gap-2">
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: board.roleOption.color || '#6B7280' }}
                        />
                        <span className="text-sm">{board.roleOption.label}</span>
                      </div>
                    ) : (
                      <span className="text-sm text-gray-500">없음</span>
                    )}
                  </td>
                  {/* Importance */}
                  <td className="px-4 py-3">
                    {board.importanceOption ? (
                      <div className="flex items-center gap-2">
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: board.importanceOption.color || '#6B7280' }}
                        />
                        <span className="text-sm">{board.importanceOption.label}</span>
                      </div>
                    ) : (
                      <span className="text-sm text-gray-500">없음</span>
                    )}
                  </td>
                  {/* Assignee */}
                  <td className="px-4 py-3">
                    <AssigneeAvatarStack assignees={board.assignee?.name || 'Unassigned'} />
                  </td>
                  {/* Due Date */}
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {board.dueDate ? new Date(board.dueDate).toLocaleDateString('ko-KR') : '없음'}
                  </td>
                </tr>
              ))}

              {/* Add Board Row */}
              <tr
                onClick={() => {
                  // 💡 [수정] setCreateBoardStageId 제거 후, CreateBoardModal을 열도록 지시
                  setShowCreateBoard(true);
                }}
                className="border-t-2 border-gray-300 hover:bg-blue-50 cursor-pointer transition"
              >
                <td colSpan={6} className="px-4 py-4">
                  <div className="flex items-center justify-center gap-2 text-blue-600 font-semibold">
                    <Plus className="w-5 h-5" />
                    <span>보드 추가</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          {allBoards.length === 0 && (
            <div className="text-center py-12 text-gray-500">
              보드가 없습니다. 보드를 추가해보세요.
            </div>
          )}
        </div>
      ) : (
        // =============================================================
        // 2. Board Layout (Kanban)
        // =============================================================
        <div className="flex flex-col lg:flex-row gap-3 sm:gap-4 min-w-max pb-4 mt-4">
          {columns.map((column, idx) => {
            // 필터링된 보드 목록 (테이블 뷰에서 사용한 searchQuery 필터링을 컬럼별로 재적용)
            const columnBoards = viewState?.searchQuery?.trim()
              ? column.boards.filter((board) => {
                  const query = viewState?.searchQuery?.toLowerCase();
                  return (
                    board.title.toLowerCase().includes(query || '') ||
                    board.content?.toLowerCase().includes(query || '')
                  );
                })
              : column.boards;

            return (
              <div
                key={column.stageId}
                draggable
                onDragStart={() => handleColumnDragStart(column)}
                onDragEnd={handleDragEnd}
                onDragOver={(e) => {
                  handleDragOver(e);
                  if (draggedBoard && !draggedColumn) {
                    setDragOverColumn(column.stageId);
                  }
                }}
                onDragLeave={() => {
                  if (draggedBoard && !draggedColumn) {
                    setDragOverColumn(null);
                  }
                }}
                onDrop={() => {
                  draggedColumn ? handleColumnDrop(column) : handleDrop(column.stageId);
                }}
                className={`w-full lg:w-80 lg:flex-shrink-0 relative transition-all cursor-move ${
                  draggedColumn?.stageId === column.stageId
                    ? 'opacity-50 scale-95 shadow-2xl rotate-2'
                    : 'opacity-100'
                }`}
              >
                <div
                  className={`relative ${theme.effects.cardBorderWidth} ${
                    dragOverColumn === column.stageId && draggedFromColumn !== column.stageId
                      ? 'border-blue-500 border-2 bg-blue-50 dark:bg-blue-900/20 shadow-lg'
                      : theme.colors.border
                  } p-3 sm:p-4 ${theme.colors.card} ${
                    theme.effects.borderRadius
                  } transition-all duration-200`}
                >
                  <div className={`flex items-center justify-between pb-2`}>
                    <h3
                      className={`font-bold ${theme.colors.text} flex items-center gap-2 ${theme.font.size.xs}`}
                    >
                      <span
                        className={`w-3 h-3 sm:w-4 sm:h-4 ${theme.effects.cardBorderWidth} ${theme.colors.border}`}
                        style={{
                          backgroundColor: column.color || getDefaultColorByIndex(idx).hex,
                        }}
                      ></span>
                      {column.title}
                      <span
                        className={`bg-black text-white px-1 sm:px-2 py-1 ${theme.effects.cardBorderWidth} ${theme.colors.border} text-[8px] sm:text-xs`}
                      >
                        {columnBoards.length}
                      </span>
                    </h3>
                  </div>

                  <div className="space-y-2 sm:space-y-3">
                    {columnBoards.map((board) => (
                      <div
                        onDragEnd={handleDragEnd}
                        key={board.boardId + column.stageId}
                        className="relative"
                        onDragOver={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          if (draggedBoard && draggedBoard.boardId !== board.boardId) {
                            setDragOverBoardId(board.boardId);
                          }
                        }}
                        onDragLeave={(e) => {
                          e.stopPropagation();
                          setDragOverBoardId(null);
                        }}
                      >
                        {/* Drop indicator line */}
                        {dragOverBoardId === board.boardId &&
                          draggedBoard &&
                          draggedBoard.boardId !== board.boardId && (
                            <div className="absolute -top-2 left-0 right-0 h-1 bg-blue-500 rounded-full shadow-lg shadow-blue-500/50 z-10"></div>
                          )}
                        <div
                          draggable
                          onDragStart={(e) => {
                            e.stopPropagation();
                            handleDragStart(board, column.stageId);
                          }}
                          onDragEnd={handleDragEnd}
                          onClick={() => setSelectedBoardId(board.boardId)}
                          className={`relative ${theme.colors.card} p-3 sm:p-4 ${
                            theme.effects.cardBorderWidth
                          } ${
                            theme.colors.border
                          } hover:border-blue-500 transition-all cursor-pointer ${
                            theme.effects.borderRadius
                          } 
                            ${
                              draggedBoard?.boardId === board.boardId
                                ? 'opacity-50 scale-95 shadow-2xl rotate-1'
                                : 'opacity-100'
                            }
                          `}
                        >
                          <h3
                            className={`font-bold ${theme.colors.text} mb-2 sm:mb-3 ${theme.font.size.xs} break-words`}
                          >
                            {board.title}
                          </h3>
                          <div className="flex items-center justify-between">
                            <AssigneeAvatarStack assignees={board.assignee?.name || 'Unassigned'} />
                          </div>
                        </div>
                      </div>
                    ))}

                    {/* Drop indicator for empty column or below all boards */}
                    {columnBoards.length === 0 &&
                      dragOverColumn === column.stageId &&
                      draggedBoard &&
                      !draggedColumn && (
                        <div className="relative py-2">
                          <div className="h-1 bg-blue-500 rounded-full shadow-lg shadow-blue-500/50"></div>
                        </div>
                      )}

                    <button
                      className={`relative w-full py-3 sm:py-4 ${theme.effects.cardBorderWidth} border-dashed ${theme.colors.border} ${theme.colors.card} hover:bg-gray-100 transition flex items-center justify-center gap-2 ${theme.font.size.xs} ${theme.effects.borderRadius}`}
                      onClick={() => {
                        // 💡 [수정] 모달을 열고, 스테이지 ID는 editData에 포함하여 상위로 전달 준비
                        onEditBoard({ stageId: column.stageId });
                        setShowCreateBoard(true);
                      }}
                      onDragOver={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        if (draggedBoard && !draggedColumn) {
                          setDragOverColumn(column.stageId);
                          setDragOverBoardId(null);
                        }
                      }}
                    >
                      <Plus className="w-3 h-3 sm:w-4 sm:h-4" style={{ strokeWidth: 3 }} />
                      보드 추가
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Board Detail Modal */}
      {selectedBoardId && (
        <BoardDetailModal
          boardId={selectedBoardId}
          workspaceId={workspaceId}
          onClose={() => setSelectedBoardId(null)}
          onBoardUpdated={fetchBoards}
          onBoardDeleted={fetchBoards}
          onEdit={handleBoardEdit} // 래핑된 핸들러 사용
        />
      )}
    </>
  );
};
