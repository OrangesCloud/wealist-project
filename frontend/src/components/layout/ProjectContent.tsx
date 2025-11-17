// src/components/layout/ProjectContent.tsx

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Plus, ArrowUp, ArrowDown } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { getDefaultColorByIndex } from '../../constants/colors';
import { AssigneeAvatarStack } from '../common/AvartarStack';
import { ProjectResponse, BoardResponse, Column, ViewState, FieldOption } from '../../types/board';
import { getBoardsByProject, updateBoard } from '../../api/board/boardService';
import { BoardDetailModal } from '../modals/board/BoardDetailModal';
import { FilterBar } from '../modals/board/FilterBar';

interface ProjectContentProps {
  // Data
  selectedProject: ProjectResponse;
  workspaceId: string;
  fieldOptionsLookup: {
    stages?: FieldOption[];
    roles?: FieldOption[];
    importances?: FieldOption[];
  };

  // Handlers
  onProjectContentUpdate: () => void;
  onManageModalOpen: () => void;

  // Initial States for Modals
  onEditBoard: (data: any) => void;

  showCreateBoard: boolean;
  setShowCreateBoard: (show: boolean) => void;
}

export const ProjectContent: React.FC<ProjectContentProps> = ({
  selectedProject,
  workspaceId,
  fieldOptionsLookup,
  // onProjectContentUpdate,
  onManageModalOpen,
  onEditBoard,
  // showCreateBoard,
  setShowCreateBoard,
}) => {
  const { theme } = useTheme();

  // 💡 [Board Data States]
  const [columns, setColumns] = useState<Column[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const {
    roles: roleOptions,
    stages: stageOptions,
    importances: importanceOptions,
  } = fieldOptionsLookup;

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

  // Drag state
  const [draggedBoard, setDraggedBoard] = useState<BoardResponse | null>(null);
  const [draggedFromColumn, setDraggedFromColumn] = useState<string | null>(null);
  const [draggedColumn, setDraggedColumn] = useState<Column | null>(null);
  const [dragOverBoardId, setDragOverBoardId] = useState<string | null>(null);
  const [dragOverColumn, setDragOverColumn] = useState<string | null>(null);

  // 💡 [추가] View State Setter Helper (유지)
  const setViewField = useCallback(<K extends keyof ViewState>(key: K, value: ViewState[K]) => {
    setViewState((prev) => ({ ...prev, [key]: value }));
  }, []);

  const getRoleOption = (roleId: string | undefined) =>
    roleId ? roleOptions?.find((r) => r.optionValue === roleId) : undefined;
  const getImportanceOption = (importanceId: string | undefined) =>
    importanceId ? importanceOptions?.find((i) => i.optionValue === importanceId) : undefined;
  const getStageOption = (stageId: string | undefined) =>
    stageId ? stageOptions?.find((i) => i.optionValue === stageId) : undefined;

  const fetchBoards = useCallback(async () => {
    if (!selectedProject || !stageOptions || stageOptions.length === 0) {
      setColumns([]);
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      const stages = stageOptions;
      const boardsResponse = await getBoardsByProject(selectedProject.projectId);

      // 데이터 처리 로직
      const stageMap = new Map<string, { stage: FieldOption; boards: BoardResponse[] }>();

      // 🔥 미분류 컬럼 추가 - fieldType 제거
      const UNASSIGNED_ID = 'UNASSIGNED';
      stageMap.set(UNASSIGNED_ID, {
        stage: {
          optionId: UNASSIGNED_ID,
          optionValue: UNASSIGNED_ID,
          optionLabel: '미분류',
        } as FieldOption,
        boards: [],
      });

      stages.forEach((stage: FieldOption) => {
        stageMap.set(stage.optionValue, { stage, boards: [] });
      });

      boardsResponse?.forEach((board: BoardResponse) => {
        const stageId = board.customFields?.stage;

        if (stageId && stageMap.has(stageId)) {
          stageMap.get(stageId)!.boards.push(board);
        } else {
          // 🔥 stage가 없거나 유효하지 않으면 미분류로
          stageMap.get(UNASSIGNED_ID)!.boards.push(board);
          console.warn(
            `⚠️ 보드 "${board.title}"가 미분류로 이동되었습니다. (Stage ID: ${stageId})`,
          );
        }
      });

      // 🔥 미분류를 제외한 나머지만 정렬
      const sortedStages = Array.from(stageMap.values())
        .filter(({ stage }) => stage.optionValue !== UNASSIGNED_ID)
        .sort((a, b) => (a.stage as any).displayOrder - (b.stage as any).displayOrder);

      // 🔥 미분류를 맨 앞에 추가
      const unassignedColumn = stageMap.get(UNASSIGNED_ID);
      const newColumns: Column[] = [
        {
          stageId: UNASSIGNED_ID,
          title: '미분류',
          color: '#B3B3B3',
          boards: unassignedColumn?.boards || [],
        },
        ...sortedStages.map(({ stage, boards }) => ({
          stageId: stage.optionValue,
          title: stage.optionLabel,
          color: (stage as any).color,
          boards: boards,
        })),
      ];

      console.log('📌 최종 newColumns:', newColumns);
      setColumns(newColumns);
    } catch (err) {
      const fetchError = err as Error;
      console.error('❌ 보드 로드 실패:', fetchError);
      setError(`보드 로드 실패: ${fetchError.message}`);
      setColumns([]);
    } finally {
      setIsLoading(false);
    }
  }, [selectedProject, stageOptions]);

  useEffect(() => {
    if (selectedProject && stageOptions && stageOptions.length > 0) {
      fetchBoards();
    }
  }, [fetchBoards, selectedProject, stageOptions]);

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

      const targetColumn = columns.find((col) => col.stageId === targetColumnId);
      if (!targetColumn) {
        handleDragEnd();
        return;
      }

      // 같은 컬럼이면 무시
      if (draggedFromColumn === targetColumnId) {
        handleDragEnd();
        return;
      }

      // 🔥 현재 뷰 타입에 따라 업데이트할 필드 결정
      const fieldKeyName = viewState.currentView as 'stage' | 'role' | 'importance';

      // 🔥 UNASSIGNED로 이동하는 경우 해당 필드를 undefined로 설정
      const newFieldValue = targetColumnId === 'UNASSIGNED' ? undefined : targetColumnId;

      // 1. 로컬 상태 업데이트를 위한 새 보드 생성
      const updatedBoard: BoardResponse = {
        ...draggedBoard,
        customFields: {
          ...draggedBoard.customFields,
          [fieldKeyName]: newFieldValue,
        },
      };

      const newColumns = columns.map((col) => {
        if (col.stageId === draggedFromColumn) {
          return { ...col, boards: col.boards.filter((t) => t.boardId !== draggedBoard.boardId) };
        }
        if (col.stageId === targetColumnId) {
          return { ...col, boards: [...col.boards, updatedBoard] };
        }
        return col;
      });

      // 2. 낙관적 UI 업데이트
      setColumns(newColumns);
      handleDragEnd();

      // 3. 🔥 API 호출 - 실제 DB 업데이트
      try {
        await updateBoard(draggedBoard.boardId, {
          customFields: {
            ...draggedBoard.customFields,
            [fieldKeyName]: newFieldValue,
          },
        });
        console.log(
          `✅ [API SUCCESS] 보드 이동 완료: ${draggedBoard.title} → ${targetColumn.title} (${fieldKeyName}: ${newFieldValue})`,
        );
      } catch (error) {
        console.error('❌ [API ERROR] 보드 이동 실패:', error);
        // 🔥 실패 시 롤백
        setColumns(columns);
        alert('보드 이동에 실패했습니다. 다시 시도해주세요.');
      }
    },
    [draggedBoard, draggedFromColumn, columns, viewState.currentView],
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

  const handleSort = (column: 'title' | 'stage' | 'role' | 'importance') => {
    if (viewState.sortColumn === column) {
      setViewField('sortDirection', viewState.sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setViewField('sortColumn', column);
      setViewField('sortDirection', 'asc');
    }
  };

  const handleBoardEdit = (boardData: any) => {
    onEditBoard(boardData);
    setSelectedBoardId(null);
  };

  // 6. Table/Board View 공통 데이터 필터링/정렬 로직 (useMemo)
  const allProcessedBoards = useMemo(() => {
    const { searchQuery, sortColumn, showCompleted } = viewState;

    const boardsToProcess = columns.flatMap((column) =>
      column.boards.map((board) => {
        const roleId = board.customFields?.role;
        const importanceId = board.customFields?.importance;
        const stageId = board.customFields?.stage;
        return {
          ...board,
          stageName: getStageOption(stageId)?.optionLabel || column.title,
          stageColor: (getStageOption(stageId) as any)?.color || column.color,
          stageId: stageId,
          roleOption: getRoleOption(roleId),
          importanceOption: getImportanceOption(importanceId),
        };
      }),
    );

    let filteredBoardsByCompletion = boardsToProcess;

    if (!showCompleted) {
      const completedStageIds = stageOptions
        ?.filter((s) => s.optionLabel === '완료')
        .map((s) => s.optionValue);

      filteredBoardsByCompletion = boardsToProcess.filter(
        (board) => !completedStageIds?.includes(board.stageId),
      );
    }

    const finalFilteredBoards = searchQuery?.trim()
      ? filteredBoardsByCompletion.filter((board) => {
          const query = searchQuery.toLowerCase();
          const titleMatch = board.title.toLowerCase().includes(query);
          const contentMatch = board.content?.toLowerCase().includes(query);
          return titleMatch || contentMatch;
        })
      : filteredBoardsByCompletion;

    const sortedBoards = [...finalFilteredBoards].sort((a, b) => {
      if (!sortColumn) return 0;
      let aValue: any;
      let bValue: any;
      const direction = viewState.sortDirection === 'asc' ? 1 : -1;

      switch (sortColumn) {
        case 'title':
          aValue = a.title.toLowerCase();
          bValue = b.title.toLowerCase();
          break;
        case 'stage':
          aValue = a.stageName;
          bValue = b.stageName;
          break;
        case 'role':
          aValue = a.roleOption?.optionLabel || '';
          bValue = b.roleOption?.optionLabel || '';
          break;
        case 'importance':
          aValue = (a.importanceOption as any)?.level || 0;
          bValue = (b.importanceOption as any)?.level || 0;
          break;
        case 'assignee':
          aValue = a.assigneeId?.toLowerCase() || '';
          bValue = b.assigneeId?.toLowerCase() || '';
          break;
        case 'dueDate':
          aValue = a.dueDate ? new Date(a.dueDate).getTime() : 0;
          bValue = b.dueDate ? new Date(b.dueDate).getTime() : 0;
          break;
        default:
          return 0;
      }

      if (aValue < bValue) return -1 * direction;
      if (aValue > bValue) return 1 * direction;
      return 0;
    });

    return sortedBoards;
  }, [
    columns,
    viewState,
    roleOptions,
    importanceOptions,
    stageOptions,
    getRoleOption,
    getImportanceOption,
    getStageOption,
  ]);

  // 7. 뷰 기준에 따라 컬럼을 재구성 (useMemo)
  const currentViewColumns = useMemo(() => {
    // 🔥 보드가 없을 때 빈 컬럼 구조 반환
    if (!allProcessedBoards || allProcessedBoards.length === 0) {
      if (
        fieldOptionsLookup?.stages?.length &&
        fieldOptionsLookup?.stages?.length > 0 &&
        viewState?.currentLayout === 'board'
      ) {
        const stages = fieldOptionsLookup.stages;
        const UNASSIGNED_ID = 'UNASSIGNED';
        const emptyColumns = [
          {
            stageId: UNASSIGNED_ID,
            title: '미분류',
            color: '#B3B3B3',
            boards: [],
          },
          ...stages.map((stage) => ({
            stageId: stage.optionValue,
            title: stage.optionLabel,
            color: (stage as any).color,
            boards: [],
          })),
        ];
        return emptyColumns;
      }
      return [];
    }

    if (stageOptions?.length === 0 && viewState.currentView === 'stage') {
      return [];
    }

    const groupByField = viewState.currentView;

    // 🔥 stage일 때는 원본 columns 그대로 사용
    if (groupByField === 'stage') {
      return columns;
    }

    // role이나 importance일 때만 재그룹화
    let baseOptions: any[] = [];
    let lookupField: 'roleOption' | 'importanceOption';

    if (groupByField === 'role') {
      baseOptions = fieldOptionsLookup?.roles || [];
      lookupField = 'roleOption';
    } else if (groupByField === 'importance') {
      baseOptions = fieldOptionsLookup?.importances || [];
      lookupField = 'importanceOption';
    } else {
      return [];
    }

    const groupedMap = new Map<string, Column>();
    const UNASSIGNED_ID = 'UNASSIGNED';
    groupedMap.set(UNASSIGNED_ID, {
      stageId: UNASSIGNED_ID,
      title: '미분류',
      color: '#B3B3B3',
      boards: [],
    });

    baseOptions?.forEach((option) => {
      const id = option.optionValue;
      groupedMap?.set(id, {
        stageId: id,
        title: option?.optionLabel,
        color: (option as any).color,
        boards: [],
      });
    });

    allProcessedBoards?.forEach((board) => {
      const optionValue = (board as any)[lookupField]?.optionValue;

      if (optionValue && groupedMap.has(optionValue)) {
        groupedMap.get(optionValue)!.boards.push(board as any);
      } else {
        groupedMap.get(UNASSIGNED_ID)!.boards.push(board as any);
      }
    });

    const result = Array.from(groupedMap.values()).sort((a, b) => {
      if (a.stageId === UNASSIGNED_ID) return -1; // 🔥 미분류를 맨 앞으로
      if (b.stageId === UNASSIGNED_ID) return 1;

      const orderA =
        (baseOptions.find((o) => o.optionValue === a.stageId) as any)?.displayOrder || 0;
      const orderB =
        (baseOptions.find((o) => o.optionValue === b.stageId) as any)?.displayOrder || 0;
      return orderA - orderB;
    });

    return result;
  }, [allProcessedBoards, viewState, fieldOptionsLookup, columns, stageOptions]);

  // 로딩 상태 처리
  if (isLoading && (stageOptions === undefined || stageOptions.length === 0)) {
    return <LoadingSpinner message="보드와 필드 데이터를 로드 중..." />;
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
        currentView={viewState.currentView}
        onLayoutChange={(layout) => setViewField('currentLayout', layout)}
        onShowCompletedChange={(show) => setViewField('showCompleted', show)}
        currentLayout={viewState.currentLayout}
        showCompleted={viewState.showCompleted}
        stageOptions={fieldOptionsLookup?.stages || []}
        roleOptions={fieldOptionsLookup?.roles || []}
        importanceOptions={fieldOptionsLookup?.importances || []}
      />

      {/* Boards or Table View */}
      {viewState?.currentLayout === 'table' ? (
        // Table Layout
        <div className="mt-4 overflow-x-auto">
          <table
            className={`w-full ${theme.colors.card} ${theme.effects.borderRadius} overflow-hidden shadow-lg`}
          >
            <thead className="bg-gray-100 border-b border-gray-200">
              <tr>
                {['title', 'stage', 'role', 'importance', 'assignee', 'dueDate'].map((col) => (
                  <th key={col} className="px-4 py-3 text-left">
                    <button
                      onClick={() => handleSort(col as 'title' | 'stage' | 'role' | 'importance')}
                      className="flex items-center gap-2 font-semibold text-sm text-gray-700 hover:text-blue-600 transition"
                    >
                      {col === 'title' && '제목'}
                      {col === 'stage' && '진행 단계'}
                      {col === 'role' && '역할'}
                      {col === 'importance' && '중요도'}
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
              {allProcessedBoards?.map((board) => (
                <tr
                  key={board.boardId}
                  onClick={() => setSelectedBoardId(board.boardId)}
                  className="border-b border-gray-200 hover:bg-gray-50 cursor-pointer transition"
                >
                  <td className="px-4 py-3 font-semibold text-gray-800">{board.title}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="w-3 h-3 rounded-full" />
                      <span className="text-sm">{board.stageName}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    {board?.roleOption ? (
                      <div className="flex items-center gap-2">
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: (board?.roleOption as any).color || '#6B7280' }}
                        />
                        <span className="text-sm">{board?.roleOption.optionLabel}</span>
                      </div>
                    ) : (
                      <span className="text-sm text-gray-500">없음</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    {board?.importanceOption ? (
                      <div className="flex items-center gap-2">
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{
                            backgroundColor: (board.importanceOption as any).color || '#6B7280',
                          }}
                        />
                        <span className="text-sm">{board.importanceOption.optionLabel}</span>
                      </div>
                    ) : (
                      <span className="text-sm text-gray-500">없음</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <AssigneeAvatarStack assignees={board.assigneeId || 'Unassigned'} />
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {board.dueDate ? new Date(board.dueDate).toLocaleDateString('ko-KR') : '없음'}
                  </td>
                </tr>
              ))}

              <tr
                onClick={() => {
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
          {allProcessedBoards?.length === 0 && (
            <div className="text-center py-12 text-gray-500">
              보드가 없습니다. 보드를 추가해보세요.
            </div>
          )}
        </div>
      ) : (
        // Board Layout (Kanban)
        <div className="flex flex-col lg:flex-row gap-3 sm:gap-4 min-w-max pb-4 mt-4">
          {currentViewColumns?.map((column, idx) => {
            const columnBoards = column.boards;
            const fieldKeyName = viewState.currentView as 'stage' | 'role' | 'importance';
            const initialData: any = {};

            // 🔥 UNASSIGNED 컬럼에서 보드 추가 시 해당 필드를 설정하지 않음
            if (column.stageId !== 'UNASSIGNED') {
              initialData[fieldKeyName] = column.stageId;
            }

            return (
              <div
                key={column?.stageId}
                draggable={column.stageId !== 'UNASSIGNED'} // 🔥 미분류 컬럼은 드래그 불가
                onDragStart={() => {
                  if (column.stageId !== 'UNASSIGNED') {
                    handleColumnDragStart(column);
                  }
                }}
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
                className={`w-full lg:w-80 lg:flex-shrink-0 relative transition-all ${
                  column.stageId !== 'UNASSIGNED' ? 'cursor-move' : ''
                } ${
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
                        {columnBoards?.length}
                      </span>
                    </h3>
                  </div>

                  <div className="space-y-2 sm:space-y-3">
                    {columnBoards?.map((board, idx) => (
                      <div
                        onDragEnd={handleDragEnd}
                        key={board.boardId + column.stageId + idx}
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
                            <AssigneeAvatarStack assignees={board.assigneeId || 'Unassigned'} />
                          </div>
                        </div>
                      </div>
                    ))}

                    {columnBoards?.length === 0 &&
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
                        const defaultInitialData: any = {
                          stage: fieldOptionsLookup.stages?.[0]?.optionValue,
                          role: fieldOptionsLookup.roles?.[0]?.optionValue,
                          importance: fieldOptionsLookup.importances?.[0]?.optionValue,
                        };

                        // 🔥 UNASSIGNED가 아닐 때만 해당 필드 설정
                        if (column.stageId !== 'UNASSIGNED') {
                          defaultInitialData[fieldKeyName] = column.stageId;
                        }

                        onEditBoard(defaultInitialData);
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
          onEdit={handleBoardEdit}
          fieldOptionsLookup={fieldOptionsLookup}
        />
      )}
    </>
  );
};
