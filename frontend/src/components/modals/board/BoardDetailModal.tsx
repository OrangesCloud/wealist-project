import React, { useState, useEffect } from 'react';
import { X, AlertCircle, Tag, CheckSquare, MessageSquare, Send, Edit2, Trash2 } from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import {
  CustomStageResponse,
  CustomRoleResponse,
  CustomImportanceResponse,
  BoardResponse,
} from '../../../types/board';
import { getBoard, deleteBoard } from '../../../api/board/boardService';
import { getWorkspaceMembers } from '../../../api/user/userService';
import { WorkspaceMemberResponse } from '../../../types/user';
import { MOCK_STAGES, MOCK_ROLES, MOCK_IMPORTANCES } from '../../../mocks/board';
interface BoardDetailModalProps {
  boardId: string;
  workspaceId: string;
  onClose: () => void;
  onBoardUpdated: () => void;
  onBoardDeleted: () => void;
  onEdit: (boardData: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stageId: string;
    assigneeId?: string; // 단일 담당자 ID
    roleIds: string[]; // 역할 ID 배열
    importanceId?: string;
    dueDate?: string;
  }) => void;
}

export const BoardDetailModal: React.FC<BoardDetailModalProps> = ({
  boardId,
  workspaceId,
  onClose,
  onBoardDeleted,
  onEdit,
}) => {
  const { theme } = useTheme();

  // Form state
  const [projectId, setProjectId] = useState<string>('');
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [selectedStageId, setSelectedStageId] = useState('');
  const [selectedRoleId, setSelectedRoleId] = useState<string>('');
  const [selectedImportanceId, setSelectedImportanceId] = useState<string>('');
  const [selectedAssigneeId, setSelectedAssigneeId] = useState<string>('');
  const [dueDate, setDueDate] = useState<string>('');

  // Data state
  const [stages, _setStages] = useState<CustomStageResponse[]>(MOCK_STAGES);
  const [roles, _setRoles] = useState<CustomRoleResponse[]>(MOCK_ROLES);
  const [importances, _setImportances] = useState<CustomImportanceResponse[]>(MOCK_IMPORTANCES);
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingBoard, setIsLoadingBoard] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Comment state (변경 없음)
  const [comments, setComments] = useState<any[]>([]);
  const [newComment, setNewComment] = useState('');

  // 보드 데이터 조회
  useEffect(() => {
    const fetchBoard = async () => {
      setIsLoadingBoard(true);
      try {
        const boardData: BoardResponse = await getBoard(boardId);

        // 보드 데이터로 상태 초기화
        setProjectId(boardData.projectId || '');
        setTitle(boardData.title || '');
        setContent(boardData.content || '');

        // 💡 [수정] customFields가 null일 경우 안전하게 빈 객체로 초기화
        const customFields = boardData.customFields || {};

        // Custom Field ID 추출 로직
        const stageIdFromCustomField = customFields.stageId || '';
        setSelectedStageId(stageIdFromCustomField);

        const roleIdsFromCustomField: string[] = customFields.roleIds || [];
        setSelectedRoleId(roleIdsFromCustomField[0] || ''); // 단일 Role ID만 사용

        const importanceIdFromCustomField = customFields.importanceId || '';
        setSelectedImportanceId(importanceIdFromCustomField);

        // 💡 [수정] assignee가 null일 경우, .userId 접근 방지
        const assigneeId: string = boardData.assignee?.userId || '';
        setSelectedAssigneeId(assigneeId);

        setDueDate(boardData.dueDate || '');

        console.log('✅ 보드 데이터 로드 성공:', boardData);
      } catch (err) {
        console.error('❌ 보드 데이터 로드 실패:', err);
        setError('보드 정보를 불러오는데 실패했습니다.');
      } finally {
        setIsLoadingBoard(false);
      }
    };

    fetchBoard();
  }, [boardId]);

  // 워크스페이스 멤버 조회
  useEffect(() => {
    const fetchMembers = async () => {
      try {
        const members = await getWorkspaceMembers(workspaceId);
        setWorkspaceMembers(members);
      } catch (err) {
        console.error('❌ 워크스페이스 멤버 로드 실패:', err);
      }
    };

    if (workspaceId) {
      fetchMembers();
    }
  }, [workspaceId]);

  const handleDelete = async () => {
    console.warn('⚠️ 보드 삭제를 진행합니다. (사용자 확인 로직 생략)');

    setIsLoading(true);
    try {
      await deleteBoard(boardId);
      alert('보드가 삭제되었습니다.');
      onBoardDeleted();
      onClose();
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error('❌ 보드 삭제 실패:', errorMsg);
      setError(errorMsg || '보드 삭제에 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddComment = () => {
    if (newComment.trim()) {
      setComments([
        ...comments,
        {
          id: comments.length + 1,
          author: '사용자',
          content: newComment,
          timestamp: '방금 전',
        },
      ]);
      setNewComment('');
    }
  };

  // 💡 필드 정보 조회 헬퍼 함수 (Mock 기반)
  const getFieldOption = (
    options: CustomStageResponse[] | CustomRoleResponse[] | CustomImportanceResponse[],
    id: string,
  ) => {
    return options.find(
      // 💡 [수정] stageId, roleId, importanceId를 명시적으로 체크
      (opt: any) => opt.stageId === id || opt.roleId === id || opt.importanceId === id,
    );
  };

  // 로딩 중이면 로딩 UI 표시
  if (isLoadingBoard) {
    return (
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
        onClick={onClose}
      >
        <div
          className={`relative w-full max-w-2xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl`}
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
              <p className="text-gray-600">보드 정보를 불러오는 중...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // 💡 [추가] UI 표시를 위한 필드 데이터 조회
  const currentStage = getFieldOption(stages, selectedStageId);
  const currentRole = getFieldOption(roles, selectedRoleId);
  const currentImportance = getFieldOption(importances, selectedImportanceId);

  // 💡 [수정] 단일 담당자 조회
  const currentAssignee = workspaceMembers?.find((m) => m.userId === selectedAssigneeId);

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
      onClick={onClose}
    >
      <div
        className={`relative w-full max-w-2xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-start justify-between mb-4 pb-4 border-b border-gray-200">
          <div className="flex-1 pr-4">
            {/* 💡 [수정] title이 null일 경우 대비 */}
            <h2 className="text-xl font-bold text-gray-800 mb-2">{title || '제목 없음'}</h2>
          </div>
          <div className="flex gap-2">
            <button
              onClick={onClose}
              className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
            {error}
          </div>
        )}

        {/* Content */}
        <div className="space-y-4 mb-6">
          {/* Description */}
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">설명</label>
            <p className="text-sm text-gray-600 whitespace-pre-wrap">
              {content || '설명이 없습니다.'}
            </p>
          </div>

          {/* Stage and Role - 2 columns */}
          <div className="grid grid-cols-2 gap-4">
            {/* Stage */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <CheckSquare className="w-4 h-4 inline mr-1" />
                진행 단계
              </label>
              <div className="flex items-center gap-2">
                <span
                  className="w-3 h-3 rounded-full"
                  style={{
                    backgroundColor: currentStage?.color || '#6B7280',
                  }}
                />
                <span className="text-sm">{currentStage?.label || '알 수 없음'}</span>
              </div>
            </div>

            {/* Role (단일 선택) */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                <Tag className="w-4 h-4 inline mr-1" />
                역할
              </label>
              <div className="flex items-center gap-2">
                <span
                  className="w-3 h-3 rounded-full"
                  style={{
                    backgroundColor: currentRole?.color || '#6B7280',
                  }}
                />
                <span className="text-sm">{currentRole?.label || '알 수 없음'}</span>
              </div>
            </div>
          </div>

          {/* Importance */}
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              <AlertCircle className="w-4 h-4 inline mr-1" />
              중요도
            </label>
            <div className="flex items-center gap-2">
              {selectedImportanceId ? (
                <>
                  <span
                    className="w-3 h-3 rounded-full"
                    style={{
                      backgroundColor: currentImportance?.color || '#6B7280',
                    }}
                  />
                  <span className="text-sm">{currentImportance?.label || '알 수 없음'}</span>
                </>
              ) : (
                <span className="text-sm text-gray-500">없음</span>
              )}
            </div>
          </div>
        </div>

        {/* Comments Section (변경 없음) */}
        <div className="pt-4 border-t border-gray-200">
          <div className="flex items-center gap-2 mb-4">
            <MessageSquare className="w-5 h-5 text-gray-700" />
            <h3 className="text-base font-bold text-gray-800">댓글 ({comments.length}개)</h3>
          </div>

          <div className="space-y-3 mb-4 max-h-40 overflow-y-auto">
            {comments.map((comment) => (
              <div key={comment.id} className="p-3 bg-gray-100 border border-gray-200 rounded-lg">
                <div className="flex items-start gap-2">
                  <div className="w-6 h-6 bg-blue-500 flex items-center justify-center text-white text-xs font-bold rounded-full flex-shrink-0">
                    {comment.author[0]}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs font-bold">{comment.author}</span>
                      <span className="text-[10px] text-gray-500">{comment.timestamp}</span>
                    </div>
                    <p className="text-sm break-words text-gray-700">{comment.content}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="flex gap-2">
            <input
              type="text"
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleAddComment()}
              placeholder="댓글을 입력하세요..."
              className="flex-1 px-3 py-2 border border-gray-300 text-sm rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isLoading}
            />
            <button
              onClick={handleAddComment}
              disabled={isLoading || !newComment.trim()}
              className="bg-blue-500 text-white px-4 py-2 hover:bg-blue-600 transition flex items-center justify-center gap-1 rounded-lg disabled:bg-gray-400"
            >
              <Send className="w-4 h-4" />
              <span className="text-xs">등록</span>
            </button>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-3 mt-6 pt-4 border-t border-gray-300">
          <button
            onClick={() => {
              // 💡 [수정] onEdit으로 전달하는 데이터 구조를 단일 담당자 및 역할 배열로 변경
              onEdit({
                boardId,
                projectId,
                title: title || '',
                content: content || '',
                stageId: selectedStageId,
                roleIds: selectedRoleId ? [selectedRoleId] : [], // 단일 선택이어도 배열 형태로 전달
                importanceId: selectedImportanceId,
                assigneeId: selectedAssigneeId,
                dueDate: dueDate,
              });
            }}
            className="flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
            disabled={isLoading}
          >
            <Edit2 className="w-4 h-4" />
            보드 수정
          </button>
          <button
            onClick={handleDelete}
            className="flex-1 px-4 py-2 bg-red-500 text-white font-semibold rounded-lg hover:bg-red-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
            disabled={isLoading}
          >
            <Trash2 className="w-4 h-4" />
            보드 삭제
          </button>
        </div>
      </div>
    </div>
  );
};
