// src/components/modals/board/BoardDetailModal.tsx

import React, { useState, useEffect, useRef } from 'react';
import {
  X,
  AlertCircle,
  Tag,
  CheckSquare,
  MessageSquare,
  Send,
  Edit2,
  Trash2,
  Paperclip,
  User,
  Users,
  Download,
  Calendar, // 💡 Calendar 아이콘 재사용
} from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
// BoardDetailResponse 타입을 사용하며, CommentResponse, ParticipantResponse도 가져옵니다.
import {
  BoardDetailResponse,
  FieldOption,
  CommentResponse,
  ParticipantResponse,
} from '../../../types/board';
import {
  getBoard,
  deleteBoard,
  // createCommentWithFile
} from '../../../api/board/boardService';
import { getWorkspaceMembers } from '../../../api/user/userService';
import { WorkspaceMemberResponse } from '../../../types/user';
import { AvatarStack } from '../../common/AvartarStack';
import { formatDate } from '../../../utils/date';
import Portal from '../../common/Portal';

// 💡 1. 정적 데이터를 담을 인터페이스 정의 (startDate 추가)
interface BoardState {
  projectId: string;
  title: string;
  content: string;
  selectedStageId: string;
  selectedRoleId: string;
  selectedImportanceId: string;
  selectedAssigneeId: string;
  dueDate: string;
  startDate: string; // 💡 [추가] 시작일
  createdAt: string;
  updatedAt: string;
  participants: ParticipantResponse[];
  fileUrl?: string;
  fileName?: string;
}

// 💡 2. 초기 상태 정의 (startDate 초기화)
const initialBoardState: BoardState = {
  projectId: '',
  title: '',
  content: '',
  selectedStageId: '',
  selectedRoleId: '',
  selectedImportanceId: '',
  selectedAssigneeId: '',
  dueDate: '',
  startDate: '', // 💡 [추가] 시작일 초기화
  createdAt: '',
  updatedAt: '',
  participants: [],
  fileUrl: undefined,
  fileName: undefined,
};

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
    stage: string;
    assigneeId?: string;
    role: string;
    importance?: string;
    dueDate?: string;
    startDate?: string; // 💡 [추가] onEdit에도 startDate 추가
  }) => void;
  fieldOptionsLookup: {
    stages?: FieldOption[];
    roles?: FieldOption[];
    importances?: FieldOption[];
  };
}

export const BoardDetailModal: React.FC<BoardDetailModalProps> = ({
  boardId,
  workspaceId,
  onClose,
  onBoardDeleted,
  onEdit,
  fieldOptionsLookup,
}) => {
  const { theme } = useTheme();

  // 💡 3. 정적 보드 데이터를 하나의 객체로 묶음
  const [boardData, setBoardData] = useState<BoardState>(initialBoardState);

  // 💡 워크스페이스 멤버 목록 (참여자/할당자 정보 매핑용)
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingBoard, setIsLoadingBoard] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Comment state (동적 상태 유지)
  const [comments, setComments] = useState<CommentResponse[]>([]);
  const [newComment, setNewComment] = useState('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 보드 데이터 조회
  useEffect(() => {
    const fetchBoard = async () => {
      setIsLoadingBoard(true);
      try {
        const data: BoardDetailResponse = await getBoard(boardId);

        const customFields = data.customFields || {};

        // 💡 4. boardData 상태에 모든 정적 데이터 한 번에 설정
        setBoardData({
          projectId: data.projectId || '',
          title: data.title || '',
          content: data.content || '',
          selectedStageId: customFields.stage || '',
          selectedRoleId: customFields.role || '',
          selectedImportanceId: customFields.importance || '',
          selectedAssigneeId: data.assigneeId || '',
          dueDate: (data as any).dueDate || '', // 💡 [수정] DueDate 로드
          startDate: (data as any).startDate || '', // 💡 [추가] StartDate 로드
          createdAt: data.createdAt,
          updatedAt: data.updatedAt,
          participants: data.participants || [],
        });

        // 💡 동적 데이터 (댓글) 설정
        setComments(data.comments || []);

        console.log('✅ 보드 데이터 로드 성공:', data);
      } catch (err) {
        console.error('❌ 보드 데이터 로드 실패:', err);
        setError('보드 정보를 불러오는데 실패했습니다.');
      } finally {
        setIsLoadingBoard(false);
      }
    };

    fetchBoard();
  }, [boardId]);

  // 💡 [수정 필요] Props에서 받은 fieldOptionsLookup에서 필요한 배열들을 구조 분해하여 정의합니다.
  const { stages = [], roles = [], importances = [] } = fieldOptionsLookup;

  // 워크스페이스 멤버 조회 (이전과 동일)
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
    if (!window.confirm('정말 이 보드를 삭제하시겠습니까?')) return;

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
  // 💡 [추가] 파일 다운로드 핸들러
  const handleFileDownload = (fileUrl: string, fileName: string) => {
    if (!fileUrl) return;

    const link = document.createElement('a');
    link.href = fileUrl;
    link.setAttribute('download', fileName);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // 파일 선택 핸들러 (이전과 동일)
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];
      const MAX_SIZE = 20 * 1024 * 1024;
      if (file.size > MAX_SIZE) {
        alert('파일 크기는 20MB를 초과할 수 없습니다.');
        setSelectedFile(null);
        e.target.value = '';
        return;
      }
      setSelectedFile(file);
    }
  };

  const handleAddComment = async () => {
    if (!newComment.trim() && !selectedFile) return;
    if (isLoading) return;

    setIsLoading(true);

    try {
      const formData = new FormData();
      formData.append('boardId', boardId);
      formData.append('content', newComment.trim());

      if (selectedFile) {
        formData.append('file', selectedFile);
      }

      // 💡 API 호출
      // const addedComment = await createCommentWithFile(formData);

      // 댓글 목록에 새 댓글 추가 (최신순 또는 서버의 정렬 기준에 따라)
      // setComments((prevComments) => [...prevComments, addedComment]);
      setNewComment('');
      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error('❌ 댓글 등록 실패:', errorMsg);
    } finally {
      setIsLoading(false);
    }
  };

  // Helper function to get option by ID
  const getFieldOption = (options: FieldOption[], id: string) => {
    return options.find((opt) => opt.optionValue === id);
  };

  if (isLoadingBoard) {
    return (
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[200]"
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

  const currentStage = getFieldOption(stages, boardData.selectedStageId);
  const currentRole = getFieldOption(roles, boardData.selectedRoleId);
  const currentImportance = getFieldOption(importances, boardData.selectedImportanceId);

  const assigneeMember = workspaceMembers.find((m) => m.userId === boardData.selectedAssigneeId);
  const participantMembers = workspaceMembers.filter((m) =>
    boardData.participants.some((p) => p.userId === m.userId),
  );

  return (
    <Portal>
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[9999]" // 💡 z-index를 최상위로 설정
        onClick={onClose}
      >
        <div
          className={`relative w-full max-w-2xl ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-start justify-between mb-4 pb-4">
            <div className="flex-1 pr-4">
              <h2 className="text-xl font-bold text-gray-800 mb-2">
                {boardData.title || '제목 없음'}
              </h2>
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
            <div className="relative">
              {/* Dates (생성일, 수정일) - 우상단에 absolute로 배치 */}
              <div className="absolute top-0 right-0 text-right space-y-1 text-xs text-gray-500 pt-0">
                <p>
                  <span className="font-medium text-gray-700">생성일:</span>{' '}
                  {formatDate(boardData.createdAt)}
                </p>
                <p>
                  <span className="font-medium text-gray-700">수정일:</span>{' '}
                  {formatDate(boardData.updatedAt)}
                </p>
              </div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">설명</label>
              <p className="text-sm text-gray-600 whitespace-pre-wrap">
                {boardData.content || '설명이 없습니다.'}
              </p>
              {/* 💡 [추가] 보드 파일 다운로드 UI (파일 유무에 관계없이 표시) */}
              <div className="mt-4 p-2 bg-gray-50 border border-gray-200 rounded-lg flex items-center justify-between text-sm">
                <span className="text-gray-700 truncate flex items-center gap-1">
                  <Paperclip className="w-4 h-4 text-gray-500 flex-shrink-0" />
                  {boardData.fileUrl ? (
                    <span className="text-gray-700">
                      {boardData.fileName || '첨부된 보드 파일'}
                    </span>
                  ) : (
                    <span className="text-gray-500">첨부 파일 없음</span>
                  )}
                </span>

                {boardData.fileUrl ? (
                  <button
                    type="button"
                    onClick={() => {
                      if (boardData?.fileUrl)
                        handleFileDownload(boardData.fileUrl, boardData.fileName || 'board_file');
                    }}
                    className="flex items-center gap-1 text-blue-600 hover:text-blue-700 transition font-medium ml-2 flex-shrink-0"
                  >
                    <Download className="w-4 h-4" />
                    <span className="text-xs">다운로드</span>
                  </button>
                ) : (
                  <span className="text-gray-400 text-xs flex-shrink-0">첨부 가능</span>
                )}
              </div>
            </div>

            <hr className="mt-4 border-gray-100" />

            {/* Stage, Role, Importance (2컬럼 레이아웃) */}
            <div className="grid grid-cols-2 gap-4">
              {/* Stage */}
              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <CheckSquare className="w-4 h-4 inline mr-1 text-blue-500" />
                  진행 단계
                </label>
                {currentStage ? (
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{
                        backgroundColor: (currentStage as any).color || '#6B7280',
                      }}
                    />
                    <span className="text-sm truncate">{currentStage.optionLabel}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">미정</span>
                )}
              </div>

              {/* Role */}
              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Tag className="w-4 h-4 inline mr-1 text-purple-500" />
                  역할
                </label>
                {currentRole ? (
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{
                        backgroundColor: (currentRole as any).color || '#6B7280',
                      }}
                    />
                    <span className="text-sm truncate">{currentRole.optionLabel}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">미정</span>
                )}
              </div>

              {/* Importance (별도 행에 배치) */}
              <div className="col-span-2">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <AlertCircle className="w-4 h-4 inline mr-1 text-red-500" />
                  중요도
                </label>
                {currentImportance ? (
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{
                        backgroundColor: (currentImportance as any).color || '#6B7280',
                      }}
                    />
                    <span className="text-sm truncate">{currentImportance.optionLabel}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">없음</span>
                )}
              </div>
            </div>
            {/* Assignee and Participants - 2 columns */}
            <div className="grid grid-cols-2 gap-4">
              {/* Assignee (작업 할당자) */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <User className="w-4 h-4 inline mr-1 text-green-500" />
                  작업 할당자 (1명)
                </label>
                {assigneeMember ? (
                  <div className="flex items-center gap-2">
                    <AvatarStack members={[assigneeMember]} />
                    <span className="text-sm">{assigneeMember.userName}</span>
                  </div>
                ) : (
                  <span className="text-sm text-gray-500">할당되지 않음</span>
                )}
              </div>

              {/* Participants (작업자) */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Users className="w-4 h-4 inline mr-1 text-orange-500" />
                  작업자 ({participantMembers.length}명)
                </label>
                {participantMembers.length > 0 ? (
                  <AvatarStack members={participantMembers} />
                ) : (
                  <span className="text-sm text-gray-500">없음</span>
                )}
              </div>
            </div>

            {/* 💡 [추가] 시작일/마감일 표시 섹션 (Stage/Role/Importance 스타일 유사하게) */}
            <div className="grid grid-cols-2 gap-4">
              {/* 1. 시작일 (Start Date) */}
              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Calendar className="w-4 h-4 inline mr-1 text-gray-500" />
                  시작일
                </label>
                <div className="flex items-center gap-2">
                  <span className="text-sm">
                    {boardData.startDate ? formatDate(boardData.startDate) : '미정'}
                  </span>
                </div>
              </div>

              {/* 2. 마감일 (Due Date) */}
              <div className="col-span-1">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  <Calendar className="w-4 h-4 inline mr-1 text-red-500" />
                  마감일
                </label>
                <div className="flex items-center gap-2">
                  <span
                    className={`text-sm ${
                      boardData.dueDate ? 'text-red-600 font-semibold' : 'text-gray-500'
                    }`}
                  >
                    {boardData.dueDate ? formatDate(boardData.dueDate) : '미정'}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Comments Section (이전과 동일) */}
          <div className="pt-4 border-t border-gray-200">
            <div className="flex items-center gap-2">
              <MessageSquare className="w-5 h-5 text-gray-700" />
              <h3 className="text-base font-bold text-gray-800">댓글 ({comments.length}개)</h3>
            </div>

            <div className="space-y-3 mb-4 max-h-40 overflow-y-auto">
              {comments.map((comment) => (
                <div
                  key={comment.commentId}
                  className="p-3 bg-gray-100 border border-gray-200 rounded-lg"
                >
                  <div className="flex items-start gap-2">
                    <div className="w-6 h-6 bg-blue-500 flex items-center justify-center text-white text-xs font-bold rounded-full flex-shrink-0">
                      {comment.userId?.[0] || '?'}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-xs font-bold">사용자 ID: {comment.userId}</span>
                        <span className="text-[10px] text-gray-500">
                          {formatDate(comment.createdAt)}
                        </span>
                      </div>
                      <p className="text-sm break-words text-gray-700">{comment.content}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* 파일 input (숨김) */}
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              accept="image/*,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
              style={{ display: 'none' }}
            />

            <div className="flex flex-col gap-2">
              {/* 파일 미리보기 및 제거 버튼 */}
              {selectedFile && (
                <div className="flex items-center justify-between p-2 bg-blue-50 border border-blue-200 rounded-lg text-sm">
                  <span className="text-blue-700 truncate">{selectedFile.name}</span>
                  <button
                    onClick={() => {
                      setSelectedFile(null);
                      if (fileInputRef.current) fileInputRef.current.value = '';
                    }}
                    className="ml-2 text-blue-500 hover:text-blue-700"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              )}

              <div className="flex gap-2">
                {/* 댓글 입력 필드 */}
                <input
                  type="text"
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleAddComment()}
                  placeholder="댓글을 입력하세요..."
                  className="flex-1 px-3 py-2 border border-gray-300 text-sm rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  disabled={isLoading}
                />

                {/* 파일 선택 버튼 */}
                <button
                  onClick={() => fileInputRef.current?.click()}
                  type="button"
                  className="bg-gray-200 text-gray-700 px-3 py-2 hover:bg-gray-300 transition flex items-center justify-center gap-1 rounded-lg disabled:bg-gray-400"
                  disabled={isLoading}
                >
                  <Paperclip className="w-4 h-4" />
                </button>

                {/* 등록 버튼 */}
                <button
                  onClick={handleAddComment}
                  disabled={isLoading || (!newComment.trim() && !selectedFile)}
                  className="bg-blue-500 text-white px-4 py-2 hover:bg-blue-600 transition flex items-center justify-center gap-1 rounded-lg disabled:bg-gray-400"
                >
                  <Send className="w-4 h-4" />
                  <span className="text-xs">등록</span>
                </button>
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3 mt-6 pt-4 border-t border-gray-300">
            <button
              onClick={handleDelete}
              className="flex-1 px-4 py-2 bg-red-500 text-white font-semibold rounded-lg hover:bg-red-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
              disabled={isLoading}
            >
              <Trash2 className="w-4 h-4" />
              보드 삭제
            </button>
            <button
              onClick={() => {
                // 💡 onEdit에 boardData의 속성 사용
                onEdit({
                  boardId,
                  projectId: boardData.projectId,
                  title: boardData.title || '',
                  content: boardData.content || '',
                  stage: boardData.selectedStageId,
                  role: boardData.selectedRoleId,
                  importance: boardData.selectedImportanceId,
                  assigneeId: boardData.selectedAssigneeId,
                  dueDate: boardData.dueDate,
                  startDate: boardData.startDate, // 💡 [추가] startDate 전달
                });
              }}
              className="flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition disabled:opacity-50 flex items-center justify-center gap-2"
              disabled={isLoading}
            >
              <Edit2 className="w-4 h-4" />
              보드 수정
            </button>
          </div>
        </div>
      </div>
    </Portal>
  );
};
