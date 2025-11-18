// src/components/modals/BoardManageModal.tsx

import React, { useState, useEffect } from 'react';
import { X, Tag, CheckSquare, AlertCircle, Plus, Settings } from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import {
  CreateBoardRequest,
  FieldOption,
  IEditCustomFields,
  UpdateBoardRequest,
} from '../../../types/board';
import { createBoard, updateBoard } from '../../../api/board/boardService';
import { getWorkspaceMembers } from '../../../api/user/userService';
import { WorkspaceMemberResponse } from '../../../types/user';

interface BoardManageModalProps {
  projectId: string;
  editData?: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stage: string;
    role: string;
    importance: string;
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

export const BoardManageModal: React.FC<BoardManageModalProps> = ({
  projectId,
  editData,
  workspaceId,
  onClose,
  // onBoardCreated,
  fieldOptionsLookup,
  handleCustomField,
}) => {
  const { theme } = useTheme();
  console.log(fieldOptionsLookup);
  // Form state
  const [title, setTitle] = useState(editData?.title || '');
  const [content, setContent] = useState(editData?.content || '');
  const [selectedStageId, setSelectedStageId] = useState(
    editData?.stage || fieldOptionsLookup.stages?.[0]?.optionValue || '',
  );
  // Role과 Importance는 초기값이 없으면 빈 문자열로 설정 (사용자가 선택하도록 유도하거나, 옵션이 아닐 수 있음)
  const [selectedRoleId, setSelectedRoleId] = useState(
    editData?.role || fieldOptionsLookup.roles?.[0]?.optionValue || '', // 기존의 fieldOptionsLookup.roles?.[0]?.optionValue 제거
  );
  const [selectedImportanceId, setSelectedImportanceId] = useState(
    editData?.importance || fieldOptionsLookup.importances?.[0]?.optionValue || '', // 기존의 fieldOptionsLookup.importances?.[0]?.optionValue 제거
  );
  // Assignee search state
  const [assigneeSearch, _setAssigneeSearch] = useState('');
  const [_workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingFields, _setIsLoadingFields] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Dropdown states
  const [showRoleDropdown, setShowRoleDropdown] = useState(false);
  const [showStageDropdown, setShowStageDropdown] = useState(false);
  const [showImportanceDropdown, setShowImportanceDropdown] = useState(false);

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
    console.log(editData);
    if (workspaceId) {
      fetchMembers();
    }
  }, [workspaceId]);

  // 드롭다운 외부 클릭 감지
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('.role-dropdown-container')) {
        setShowRoleDropdown(false);
      }
      if (!target.closest('.stage-dropdown-container')) {
        setShowStageDropdown(false);
      }
      if (!target.closest('.importance-dropdown-container')) {
        setShowImportanceDropdown(false);
      }
    };

    if (showRoleDropdown || showStageDropdown || showImportanceDropdown || assigneeSearch.trim()) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showRoleDropdown, showStageDropdown, showImportanceDropdown, assigneeSearch]);

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

    setIsLoading(true);
    setError(null);

    try {
      const customFields: Record<string, any> = {
        stage: selectedStageId,
      };

      if (selectedRoleId) {
        customFields.role = selectedRoleId;
      }

      if (selectedImportanceId) {
        customFields.importance = selectedImportanceId;
      }

      console.log(customFields);

      const boardData: CreateBoardRequest | UpdateBoardRequest = {
        projectId,
        title: title.trim(),
        content: content.trim() || undefined,
        customFields,
      };
      console.log(boardData);
      if (editData?.boardId) {
        await updateBoard(editData!.boardId, boardData);
      } else {
        await createBoard(boardData as CreateBoardRequest);
      }

      alert(`✅  보드 ${editData?.boardId ? '수정' : '생성'} 완료!`);
      // onBoardCreated();
      // onClose();
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error(`❌ 보드 ${editData?.boardId ? '수정' : '생성'} 실패:`, errorMsg);
      setError(errorMsg || `보드 ${editData?.boardId ? '수정' : '생성'}에 실패했습니다.`);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
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
              </div>

              {/* Stage and Role Selection */}
              <div className="grid grid-cols-2 gap-4">
                {/* Stage Selection */}
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
                    <CheckSquare className="w-4 h-4 text-gray-400" />
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
                              fieldOptionsLookup.roles.find((r) => r.optionValue === selectedRoleId)
                                ?.optionLabel
                            }
                          </>
                        )}
                    </span>
                    <Tag className="w-4 h-4 text-gray-400" />
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

              {/* Importance and Field Management */}
              <div className="grid grid-cols-2 gap-4">
                {/* Importance Selection */}
                <div className="relative importance-dropdown-container">
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    <AlertCircle className="w-4 h-4 inline mr-1" />
                    중요도
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
                        <span className="text-gray-500">없음</span>
                      )}
                    </span>
                    <AlertCircle className="w-4 h-4 text-gray-400" />
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

                {/* Field Management */}
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
              <div className="flex gap-3 pt-4 border-t sticky bottom-0 bg-white">
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
  );
};
