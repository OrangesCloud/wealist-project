// src/components/modals/board/BoardManageModal.tsx

import React, { useState, useEffect, useRef } from 'react';
import {
  X,
  Tag,
  CheckSquare,
  AlertCircle,
  Plus,
  Settings,
  User,
  Users,
  ChevronDown,
  CheckSquare as CheckSquareIcon,
  Paperclip,
  Calendar,
} from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import {
  CreateBoardRequest,
  FieldOption,
  IEditCustomFields,
  UpdateBoardRequest,
} from '../../../types/board';
import { createBoard, updateBoard } from '../../../api/boardService';
import { getWorkspaceMembers } from '../../../api/userService';
import { WorkspaceMemberResponse } from '../../../types/user';
import { AvatarStack } from '../../common/AvartarStack';
import Portal from '../../common/Portal';

interface BoardManageModalProps {
  projectId: string;
  editData?: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stage: string;
    role: string;
    dueDate: string;
    startDate: string; // 💡 [추가] startDate
    importance: string;
    assigneeId?: string;
    participantIds?: string[];
  } | null;
  workspaceId: string;
  onClose: () => void;
  onBoardCreated: () => void;
  fieldOptionsLookup: {
    stages?: FieldOption[];
    roles?: FieldOption[];
    importances?: FieldOption[];
  };
  handleCustomField: (editFieldData: IEditCustomFields | null) => void;
}

// 💡 Assignee, Participant 아바타 표시를 위한 헬퍼 함수
const getMember = (members: WorkspaceMemberResponse[], userId: string) => {
  return members.find((m) => m.userId === userId);
};

export const BoardManageModal: React.FC<BoardManageModalProps> = ({
  projectId,
  editData,
  workspaceId,
  onClose,
  onBoardCreated,
  fieldOptionsLookup,
  handleCustomField,
}) => {
  const { theme } = useTheme();

  // Form state
  const [title, setTitle] = useState(editData?.title || '');
  const [content, setContent] = useState(editData?.content || '');
  const [selectedStageId, setSelectedStageId] = useState(
    editData?.stage || fieldOptionsLookup.stages?.[0]?.optionValue || '',
  );
  const [selectedRoleId, setSelectedRoleId] = useState(
    editData?.role || fieldOptionsLookup.roles?.[0]?.optionValue || '',
  );
  const [selectedImportanceId, setSelectedImportanceId] = useState(
    editData?.importance || fieldOptionsLookup.importances?.[0]?.optionValue || '',
  );

  const [selectedAssigneeId, setSelectedAssigneeId] = useState(editData?.assigneeId || '');
  const [selectedParticipantIds, setSelectedParticipantIds] = useState<string[]>(
    editData?.participantIds || [],
  );

  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);

  // 💡 [수정] 마감일 상태
  const [dueDate, setDueDate] = useState(
    editData?.dueDate ? editData.dueDate.substring(0, 10) : '',
  );
  // 💡 [추가] 시작일 상태
  const [startDate, setStartDate] = useState(
    editData?.startDate ? editData.startDate.substring(0, 10) : '',
  );

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingFields, _setIsLoadingFields] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Dropdown states
  const [showRoleDropdown, setShowRoleDropdown] = useState(false);
  const [showStageDropdown, setShowStageDropdown] = useState(false);
  const [showImportanceDropdown, setShowImportanceDropdown] = useState(false);
  const [showAssigneeDropdown, setShowAssigneeDropdown] = useState(false);
  const [showParticipantDropdown, setShowParticipantDropdown] = useState(false);

  // 💡 [추가] 파일 상태
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 드롭다운 외부 클릭 감지 (기존 로직 유지)
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;

      if (!target.closest('.role-dropdown-container')) setShowRoleDropdown(false);
      if (!target.closest('.stage-dropdown-container')) setShowStageDropdown(false);
      if (!target.closest('.importance-dropdown-container')) setShowImportanceDropdown(false);
      if (!target.closest('.assignee-dropdown-container')) setShowAssigneeDropdown(false);
      if (!target.closest('.participant-dropdown-container')) setShowParticipantDropdown(false);
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // 워크스페이스 멤버 조회 (기존 로직 유지)
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

  // 💡 작업자 다중 선택 토글 핸들러
  const toggleParticipant = (userId: string) => {
    setSelectedParticipantIds((prev) =>
      prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId],
    );
  };

  // 💡 파일 선택 핸들러 (20MB 제한)
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

  // 제출 핸들러
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    if (!title.trim()) {
      setError('보드 제목은 필수입니다.');
      return;
    }
    if (!selectedStageId) {
      setError('진행 단계를 선택해주세요.');
      return;
    }
    if (!selectedRoleId) {
      setError('역할을 선택해주세요.');
      return;
    }
    if (!selectedImportanceId) {
      setError('중요도를 선택해주세요.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const customFields: Record<string, any> = {
        stage: selectedStageId,
        role: selectedRoleId,
        importance: selectedImportanceId,
      };

      const isEditing = !!editData?.boardId;

      // 💡 파일이 첨부된 경우와 그렇지 않은 경우를 분리
      if (selectedFile) {
        // 파일이 있는 경우: FormData를 사용하여 멀티파트 요청 전송
        const formData = new FormData();

        formData.append('title', title.trim());
        if (content.trim()) formData.append('content', content.trim());
        formData.append('projectId', projectId);
        formData.append('assigneeId', selectedAssigneeId);
        formData.append('dueDate', dueDate);
        formData.append('startDate', startDate); // 💡 [추가] StartDate FormData에 추가
        formData.append('participants', JSON.stringify(selectedParticipantIds));
        formData.append('customFields', JSON.stringify(customFields));
        formData.append('file', selectedFile);

        // 💡 API 호출 (Mock 유지)
        if (isEditing) {
          console.warn('Update with file API call is mocked. Using FormData:', formData);
        } else {
          console.warn('Create with file API call is mocked. Using FormData:', formData);
        }
      } else {
        // 파일이 없는 경우: 일반 JSON 요청 전송
        const boardData: CreateBoardRequest | UpdateBoardRequest = {
          projectId,
          title: title.trim(),
          content: content.trim() || undefined,
          customFields,
          assigneeId: selectedAssigneeId || undefined,
          participants: selectedParticipantIds,
          dueDate: dueDate || undefined,
          startDate: startDate || undefined, // 💡 [추가] StartDate Payload에 추가
        };

        if (isEditing) {
          await updateBoard(editData!.boardId, boardData);
        } else {
          await createBoard(boardData as CreateBoardRequest);
        }
      }

      alert(`✅  보드 ${isEditing ? '수정' : '생성'} 완료!`);
      onBoardCreated();
      onClose();
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error(`❌ 보드 ${editData?.boardId ? '수정' : '생성'} 실패:`, errorMsg);
      setError(errorMsg || `보드 ${editData?.boardId ? '수정' : '생성'}에 실패했습니다.`);
    } finally {
      setIsLoading(false);
    }
  };
  // 현재 선택된 할당자 정보
  const currentAssignee = getMember(workspaceMembers, selectedAssigneeId);

  return (
    <Portal>
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[9999] modal-container"
        onClick={onClose}
      >
        <div
          className={`relative w-full max-w-2xl ${theme.colors.card} ${theme.effects.borderRadius} shadow-xl max-h-[90vh] flex flex-col overflow-hidden`}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 pt-6 pb-4 flex-shrink-0">
            <h2 className="text-xl font-bold text-gray-800">
              {editData?.boardId ? '보드 수정' : '새 보드 만들기'}
            </h2>
            <button
              onClick={onClose}
              className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Scrollable Content Area */}
          <div className="flex-1 overflow-y-auto px-6">
            {/* Error Message */}
            {error && (
              <div className="mt-4 mb-4 p-3 bg-red-50 border border-red-300 rounded-lg text-red-700 text-sm">
                {error}
              </div>
            )}

            {/* Loading State */}
            {isLoadingFields ? (
              <div className="flex items-center justify-center py-12">
                <div className="text-center">
                  <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
                  <p className="text-gray-600">커스텀 필드를 불러오는 중...</p>
                </div>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4 pb-4">
                {/* Title */}
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    보드 제목 <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    placeholder="예: 사용자 인증 API 구현"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                    disabled={isLoading}
                    maxLength={200}
                  />
                </div>

                {/* Content */}
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    설명 (선택)
                  </label>
                  <textarea
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    placeholder="보드에 대한 자세한 설명을 입력하세요"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm resize-none"
                    rows={3}
                    disabled={isLoading}
                    maxLength={5000}
                  />
                  {/* 💡 파일 첨부 버튼 및 미리보기 */}
                  <div className="mt-2 flex items-center gap-3">
                    {/* 숨겨진 파일 Input */}
                    <input
                      type="file"
                      ref={fileInputRef}
                      onChange={handleFileChange}
                      accept="image/*,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                      style={{ display: 'none' }}
                      disabled={isLoading}
                    />

                    {/* 파일 선택 버튼 */}
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      className="flex items-center gap-1 text-sm text-blue-600 font-medium hover:text-blue-700 transition disabled:text-gray-400"
                      disabled={isLoading}
                    >
                      <Paperclip className="w-4 h-4" />
                      {selectedFile ? '파일 변경' : '파일 첨부 (최대 20MB)'}
                    </button>

                    {/* 선택된 파일 이름 표시 */}
                    {selectedFile && (
                      <div className="flex items-center gap-1 text-xs text-gray-700 p-1 px-2 border border-gray-300 rounded-full bg-gray-50 max-w-[200px] truncate">
                        {selectedFile.name}
                        <button
                          type="button"
                          onClick={() => {
                            setSelectedFile(null);
                            if (fileInputRef.current) fileInputRef.current.value = '';
                          }}
                          className="ml-1 text-gray-500 hover:text-gray-700 transition"
                          disabled={isLoading}
                        >
                          <X className="w-3 h-3" />
                        </button>
                      </div>
                    )}
                  </div>
                </div>

                <hr className="my-4 border-gray-100" />

                {/* 💡 [추가] 시작일 / 마감일 (2컬럼 배치) */}
                <div className="grid grid-cols-2 gap-4">
                  {/* 시작일 캘린더 입력 */}
                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <Calendar className="w-4 h-4 inline mr-1 text-gray-500" />
                      시작일 (선택)
                    </label>
                    <input
                      type="date"
                      value={startDate}
                      onChange={(e) => setStartDate(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                      disabled={isLoading}
                    />
                  </div>

                  {/* 마감일 캘린더 입력 */}
                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <Calendar className="w-4 h-4 inline mr-1 text-red-500" />
                      마감일 (선택)
                    </label>
                    <input
                      type="date"
                      value={dueDate}
                      onChange={(e) => setDueDate(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                      disabled={isLoading}
                    />
                  </div>
                </div>

                {/* Assignee / Participants Selection (2컬럼) */}
                <div className="grid grid-cols-2 gap-4">
                  {/* 💡 Assignee Selection (단건 선택) */}
                  <div className="relative assignee-dropdown-container">
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <User className="w-4 h-4 inline mr-1 text-green-500" />
                      작업 할당자 (Assignee)
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowAssigneeDropdown(!showAssigneeDropdown)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                      disabled={isLoading}
                    >
                      <span className="flex items-center gap-2">
                        {currentAssignee ? (
                          <>
                            <div
                              className={`w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold bg-blue-500 text-white flex-shrink-0`}
                            >
                              {currentAssignee.userName[0]}
                            </div>
                            {currentAssignee.userName}
                          </>
                        ) : (
                          <span className="text-gray-500">할당자 선택</span>
                        )}
                      </span>
                      <ChevronDown className="w-4 h-4 text-gray-400" />
                    </button>
                    {showAssigneeDropdown && (
                      <div className="absolute z-20 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                        {/* 할당 해제 옵션 */}
                        <button
                          type="button"
                          onClick={() => {
                            setSelectedAssigneeId('');
                            setShowAssigneeDropdown(false);
                          }}
                          className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                            selectedAssigneeId === '' ? 'bg-blue-50' : ''
                          }`}
                        >
                          <X className="w-4 h-4 text-gray-400" /> 할당 해제
                        </button>

                        {workspaceMembers.map((member) => (
                          <button
                            key={member.userId}
                            type="button"
                            onClick={() => {
                              setSelectedAssigneeId(member.userId);
                              setShowAssigneeDropdown(false);
                            }}
                            className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                              selectedAssigneeId === member.userId ? 'bg-blue-50' : ''
                            }`}
                          >
                            <div
                              className={`w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold bg-gray-400 text-white flex-shrink-0`}
                            >
                              {member.userName[0]}
                            </div>
                            {member.userName}
                            {selectedAssigneeId === member.userId && (
                              <CheckSquareIcon className="w-4 h-4 ml-auto text-blue-500" />
                            )}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* 💡 Participants Selection (다중 선택) */}
                  <div className="relative participant-dropdown-container">
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <Users className="w-4 h-4 inline mr-1 text-orange-500" />
                      작업자 (Participants)
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowParticipantDropdown(!showParticipantDropdown)}
                      // 💡 [수정] 높이 고정 (h-10) 및 내부 요소 정렬 (items-center)
                      className="w-full px-3 py-2 h-10 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                      disabled={isLoading}
                    >
                      <span className="flex items-center gap-2">
                        {selectedParticipantIds.length > 0 ? (
                          <>
                            <AvatarStack
                              members={workspaceMembers.filter((m) =>
                                selectedParticipantIds.includes(m.userId),
                              )}
                            />
                            {selectedParticipantIds.length}명 선택됨
                          </>
                        ) : (
                          <span className="text-gray-500">작업자 선택</span>
                        )}
                      </span>
                      <ChevronDown className="w-4 h-4 text-gray-400" />
                    </button>
                    {showParticipantDropdown && (
                      <div className="absolute z-20 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                        {workspaceMembers.map((member) => (
                          <button
                            key={member.userId}
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation(); // 다중 선택 시 드롭다운이 닫히지 않도록 이벤트 전파 중단
                              toggleParticipant(member.userId);
                            }}
                            className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                              selectedParticipantIds.includes(member.userId) ? 'bg-blue-50' : ''
                            }`}
                          >
                            <div
                              className={`w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold bg-gray-400 text-white flex-shrink-0`}
                            >
                              {member.userName[0]}
                            </div>
                            {member.userName}
                            {selectedParticipantIds.includes(member.userId) && (
                              <CheckSquareIcon className="w-4 h-4 ml-auto text-blue-500" />
                            )}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                {/* Stage / Role Selection (2컬럼) */}
                <div className="grid grid-cols-2 gap-4">
                  {/* Stage Selection (기존 코드 유지) */}
                  <div className="relative stage-dropdown-container">
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <CheckSquare className="w-4 h-4 inline mr-1" />
                      진행 단계 <span className="text-red-500">*</span>
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowStageDropdown(!showStageDropdown)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                      disabled={isLoading}
                    >
                      <span className="flex items-center gap-2">
                        {selectedStageId &&
                          fieldOptionsLookup?.stages?.find(
                            (s) => s.optionValue === selectedStageId,
                          ) && (
                            <>
                              <span
                                className="w-3 h-3 rounded-full"
                                style={{
                                  backgroundColor:
                                    (
                                      fieldOptionsLookup?.stages?.find(
                                        (s) => s.optionValue === selectedStageId,
                                      ) as any
                                    )?.color || '#6B7280',
                                }}
                              />
                              {
                                fieldOptionsLookup?.stages?.find(
                                  (s) => s.optionValue === selectedStageId,
                                )?.optionLabel
                              }
                            </>
                          )}
                      </span>
                      <ChevronDown className="w-4 h-4 text-gray-400" />
                    </button>
                    {showStageDropdown && (
                      <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                        {fieldOptionsLookup?.stages?.map((stage) => (
                          <button
                            key={stage.optionId}
                            type="button"
                            onClick={() => {
                              setSelectedStageId(stage.optionValue);
                              setShowStageDropdown(false);
                            }}
                            className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                              selectedStageId === stage.optionValue ? 'bg-blue-50' : ''
                            }`}
                          >
                            <span
                              className="w-3 h-3 rounded-full"
                              style={{ backgroundColor: (stage as any).color || '#6B7280' }}
                            />
                            {stage?.optionLabel}
                            {selectedStageId === stage.optionValue && (
                              <CheckSquareIcon className="w-4 h-4 ml-auto text-blue-500" />
                            )}
                          </button>
                        ))}
                        <button
                          type="button"
                          onClick={() => {
                            setShowStageDropdown(false);
                            handleCustomField({
                              name: '진행 단계',
                              fieldType: 'multi_select',
                              options: fieldOptionsLookup?.stages,
                            });
                          }}
                          className="w-full px-3 py-2 text-left transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2 disabled:text-gray-400 disabled:cursor-not-allowed"
                        >
                          <Settings className="w-4 h-4" /> 진행 단계 관리
                        </button>
                      </div>
                    )}
                  </div>

                  {/* Role Selection (기존 코드 유지) */}
                  <div className="relative role-dropdown-container">
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <Tag className="w-4 h-4 inline mr-1" />
                      역할 <span className="text-red-500">*</span>
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowRoleDropdown(!showRoleDropdown)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                      disabled={isLoading}
                    >
                      <span className="flex items-center gap-2">
                        {selectedRoleId &&
                          fieldOptionsLookup?.roles?.find(
                            (r) => r.optionValue === selectedRoleId,
                          ) && (
                            <>
                              <span
                                className="w-3 h-3 rounded-full"
                                style={{
                                  backgroundColor:
                                    (
                                      fieldOptionsLookup.roles.find(
                                        (r) => r.optionValue === selectedRoleId,
                                      ) as any
                                    )?.color || '#6B7280',
                                }}
                              />
                              {
                                fieldOptionsLookup.roles.find(
                                  (r) => r.optionValue === selectedRoleId,
                                )?.optionLabel
                              }
                            </>
                          )}
                      </span>
                      <ChevronDown className="w-4 h-4 text-gray-400" />
                    </button>
                    {showRoleDropdown && (
                      <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                        {fieldOptionsLookup?.roles?.map((role) => (
                          <button
                            key={role.optionId}
                            type="button"
                            onClick={() => {
                              setSelectedRoleId(role.optionValue);
                              setShowRoleDropdown(false);
                            }}
                            className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                              selectedRoleId === role.optionValue ? 'bg-blue-50' : ''
                            }`}
                          >
                            <span
                              className="w-3 h-3 rounded-full"
                              style={{ backgroundColor: (role as any).color || '#6B7280' }}
                            />
                            {role.optionLabel}
                            {selectedRoleId === role.optionValue && (
                              <CheckSquareIcon className="w-4 h-4 ml-auto text-blue-500" />
                            )}
                          </button>
                        ))}
                        <button
                          type="button"
                          onClick={() => {
                            setShowRoleDropdown(false);
                            handleCustomField({
                              name: '역할',
                              fieldType: 'multi_select',
                              options: fieldOptionsLookup?.roles,
                            });
                          }}
                          className="w-full px-3 py-2 text-left transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2 disabled:text-gray-400 disabled:cursor-not-allowed"
                        >
                          <Settings className="w-4 h-4" /> 역할 관리
                        </button>
                      </div>
                    )}
                  </div>
                </div>

                {/* Importance and Field Management (2컬럼) */}
                <div className="grid grid-cols-2 gap-4">
                  {/* Importance Selection (기존 코드 유지) */}
                  <div className="relative importance-dropdown-container">
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <AlertCircle className="w-4 h-4 inline mr-1" />
                      중요도 <span className="text-red-500">*</span>
                    </label>
                    <button
                      type="button"
                      onClick={() => setShowImportanceDropdown(!showImportanceDropdown)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between"
                      disabled={isLoading}
                    >
                      <span className="flex items-center gap-2">
                        {selectedImportanceId ? (
                          fieldOptionsLookup?.importances?.find(
                            (i) => i.optionValue === selectedImportanceId,
                          ) && (
                            <>
                              <span
                                className="w-3 h-3 rounded-full"
                                style={{
                                  backgroundColor:
                                    (
                                      fieldOptionsLookup.importances.find(
                                        (i) => i.optionValue === selectedImportanceId,
                                      ) as any
                                    )?.color || '#6B7280',
                                }}
                              />
                              {
                                fieldOptionsLookup.importances.find(
                                  (i) => i.optionValue === selectedImportanceId,
                                )?.optionLabel
                              }
                            </>
                          )
                        ) : (
                          <span className="text-gray-500">선택</span>
                        )}
                      </span>
                      <ChevronDown className="w-4 h-4 text-gray-400" />
                    </button>
                    {showImportanceDropdown && (
                      <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                        {fieldOptionsLookup?.importances?.map((importance) => (
                          <button
                            key={importance.optionId}
                            type="button"
                            onClick={() => {
                              setSelectedImportanceId(importance.optionValue);
                              setShowImportanceDropdown(false);
                            }}
                            className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                              selectedImportanceId === importance.optionValue ? 'bg-blue-50' : ''
                            }`}
                          >
                            <span
                              className="w-3 h-3 rounded-full"
                              style={{ backgroundColor: (importance as any).color || '#6B7280' }}
                            />
                            {importance.optionLabel}
                            {selectedImportanceId === importance.optionValue && (
                              <CheckSquareIcon className="w-4 h-4 ml-auto text-blue-500" />
                            )}
                          </button>
                        ))}
                        <button
                          type="button"
                          onClick={() => {
                            setShowImportanceDropdown(false);
                            handleCustomField({
                              name: '중요도',
                              fieldType: 'multi_select',
                              options: fieldOptionsLookup?.importances,
                            });
                          }}
                          className="w-full px-3 py-2 text-left transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2 disabled:text-gray-400 disabled:cursor-not-allowed"
                        >
                          <Settings className="w-4 h-4" /> 중요도 관리
                        </button>
                      </div>
                    )}
                  </div>

                  {/* Field Management (기존 코드 유지) */}
                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      <Plus className="w-4 h-4 inline mr-1" />
                      필드 추가
                    </label>
                    <button
                      type="button"
                      onClick={() => handleCustomField(null)}
                      className="w-full px-3 py-2 border border-dashed border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between font-medium"
                      disabled={isLoading}
                    >
                      <span className="text-gray-600">필드 생성하기</span>
                      <Settings className="w-4 h-4 text-gray-400" />
                    </button>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex gap-3 pt-4 sticky bottom-0 bg-white">
                  <button
                    type="button"
                    onClick={onClose}
                    className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 font-semibold hover:bg-gray-50 transition"
                    disabled={isLoading}
                  >
                    취소
                  </button>
                  <button
                    type="submit"
                    className={`flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition ${
                      isLoading ? 'opacity-50 cursor-not-allowed' : ''
                    }`}
                    disabled={isLoading}
                  >
                    {isLoading
                      ? editData?.boardId
                        ? '수정 중...'
                        : '생성 중...'
                      : editData?.boardId
                      ? '보드 수정'
                      : '보드 만들기'}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      </div>
    </Portal>
  );
};
