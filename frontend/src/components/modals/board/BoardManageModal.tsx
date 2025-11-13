// src/components/modals/BoardManageModal.tsx

import React, { useState, useEffect } from 'react';
import { X, Tag, CheckSquare, AlertCircle, Calendar, User, Plus, Settings } from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import { CUSTOM_FIELD_COLORS, ColorOption } from '../../../constants/colors';
import { CreateBoardRequest, FieldOptionsLookup, UpdateBoardRequest } from '../../../types/board';
import { createBoard, updateBoard } from '../../../api/board/boardService';
import { getWorkspaceMembers } from '../../../api/user/userService';
import { WorkspaceMemberResponse } from '../../../types/user';
import { IFieldOption } from '../../../types/common';

interface BoardManageModalProps {
  projectId: string;
  // initial?: IFieldOption;
  editData?: {
    boardId: string;
    projectId: string;
    title: string;
    content: string;
    stageId: string;
    roleId: string;
    importanceId: string;
    assigneeIds: string[];
    dueDate: string;
  } | null;
  workspaceId: string;
  onClose: () => void;
  onBoardCreated: () => void;
  onAddFieldsClick: () => void;
  fieldOptionsLookup: FieldOptionsLookup;
}

export const BoardManageModal: React.FC<BoardManageModalProps> = ({
  projectId,
  editData,
  workspaceId,
  onClose,
  onBoardCreated,
  onAddFieldsClick,
  fieldOptionsLookup,
}) => {
  const { theme } = useTheme();
  // Form state
  const [title, setTitle] = useState(editData?.title || '');
  const [content, setContent] = useState(editData?.content || '');
  const [selectedStageId, setSelectedStageId] = useState(
    editData?.stageId || fieldOptionsLookup.stages?.[0]?.stageId || '',
  );
  const [selectedRoleId, setSelectedRoleId] = useState(
    editData?.roleId || fieldOptionsLookup.roles?.[0]?.roleId || '',
  );
  const [selectedImportanceId, setSelectedImportanceId] = useState(
    editData?.importanceId || fieldOptionsLookup.importances?.[0]?.importanceId || '',
  );
  // Assignee search state
  const [assigneeSearch, setAssigneeSearch] = useState('');
  const [workspaceMembers, setWorkspaceMembers] = useState<WorkspaceMemberResponse[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingFields, setIsLoadingFields] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 💡 [제거] fieldRefreshKey 상태 제거

  // Inline creation state (API 미지원으로 임시 비활성화)
  const [showCreateStage, setShowCreateStage] = useState(false);
  const [showCreateRole, setShowCreateRole] = useState(false);
  const [showCreateImportance, setShowCreateImportance] = useState(false);
  const [newFieldName, setNewFieldName] = useState('');
  const [newFieldColor, setNewFieldColor] = useState(CUSTOM_FIELD_COLORS[0].hex);
  const [newImportanceLevel, setNewImportanceLevel] = useState(1);

  // Dropdown states (변경 없음)
  const [showRoleDropdown, setShowRoleDropdown] = useState(false);
  const [showStageDropdown, setShowStageDropdown] = useState(false);
  const [showImportanceDropdown, setShowImportanceDropdown] = useState(false);

  // 1.2 워크스페이스 멤버 조회 (유지)
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

  // 1.3 드롭다운 외부 클릭 감지 (유지)
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
      if (!target.closest('.assignee-dropdown-container')) {
        setAssigneeSearch('');
      }
    };

    if (showRoleDropdown || showStageDropdown || showImportanceDropdown || assigneeSearch.trim()) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showRoleDropdown, showStageDropdown, showImportanceDropdown, assigneeSearch]);

  // 2. Inline custom field creation handlers (유지)
  const handleCreateCustomField = async (type: 'stage' | 'role' | 'importance') => {
    setError(
      `새 ${type} 필드 추가 기능은 현재 API 스펙 변경으로 인해 비활성화되었습니다. (API 미지원)`,
    );
    setIsLoading(false);
    cancelInlineCreation();
  };

  const cancelInlineCreation = () => {
    setShowCreateStage(false);
    setShowCreateRole(false);
    setShowCreateImportance(false);
    setNewFieldName('');
    setNewFieldColor(CUSTOM_FIELD_COLORS[0].hex);
    setNewImportanceLevel(1);
  };

  // 3. 제출 핸들러 (유지)
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
      const boardData: CreateBoardRequest | UpdateBoardRequest = {
        projectId,
        title: title.trim(),
        content: content.trim() || undefined,
        stageId: selectedStageId,
        roleId: selectedRoleId || undefined,
        importanceId: selectedImportanceId || undefined,
      };
      console.log(boardData);
      if (editData?.boardId) {
        await updateBoard(editData!.boardId, boardData);
      } else {
        await createBoard(boardData as CreateBoardRequest);
      }

      // onBoardCreated();
      // onClose();
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error(`❌ 보드 ${editData ? '수정' : '생성'} 실패:`, errorMsg);
      setError(errorMsg || `보드 ${editData ? '수정' : '생성'}에 실패했습니다.`);
    } finally {
      setIsLoading(false);
    }
  };

  // Helper: Color Picker Component (유지)
  const renderColorPicker = (selectedColor: string, onColorChange: (color: string) => void) => (
    <div className="grid grid-cols-6 gap-2 mt-2">
      {CUSTOM_FIELD_COLORS?.map((color: ColorOption) => (
        <button
          key={color.hex}
          type="button"
          className={`w-8 h-8 rounded-md border-2 transition-all ${
            selectedColor === color.hex
              ? 'border-gray-800 ring-2 ring-blue-500 scale-110'
              : 'border-gray-300 hover:scale-105'
          }`}
          style={{ backgroundColor: color.hex }}
          onClick={() => onColorChange(color.hex)}
          title={color.name}
          disabled={isLoading}
        />
      ))}
    </div>
  );

  // Helper: Creation Modal (작은 모달로 표시) - 인라인 생성 기능 비활성화 (유지)
  const renderCreationModal = (
    type: 'stage' | 'role' | 'importance' | 'importance',
    title: string,
  ) => (
    <div
      className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-[100]"
      onClick={cancelInlineCreation}
    >
      <div
        className="bg-white rounded-lg shadow-xl p-6 w-full max-w-md"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-bold text-gray-800 mb-4">새 {title} 추가</h3>
        <div className="space-y-4">
          <div>
            <label className="text-sm font-semibold text-gray-700 block mb-2">
              {title} 이름 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={newFieldName}
              onChange={(e) => setNewFieldName(e.target.value)}
              placeholder={`예: ${
                type === 'stage' ? '진행중' : type === 'role' ? '디자이너' : '매우 높음'
              }`}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              disabled={isLoading}
              autoFocus
            />
          </div>
          <div>
            <label className="text-sm font-semibold text-gray-700 block mb-2">색상 선택</label>
            {renderColorPicker(newFieldColor, setNewFieldColor)}
          </div>
          {type === 'importance' && (
            <div>
              <label className="text-sm font-semibold text-gray-700 block mb-2">
                중요도 레벨 (1-5)
              </label>
              <input
                type="number"
                min="1"
                max="5"
                value={newImportanceLevel}
                onChange={(e) => setNewImportanceLevel(parseInt(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                disabled={isLoading}
              />
            </div>
          )}
          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={cancelInlineCreation}
              className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 font-semibold rounded-lg hover:bg-gray-50 transition"
              disabled={isLoading}
            >
              취소
            </button>
            <button
              type="button"
              onClick={() => handleCreateCustomField(type)}
              className="flex-1 px-4 py-2 bg-blue-500 text-white font-semibold rounded-lg hover:bg-blue-600 transition"
              // 💡 인라인 생성 기능을 임시로 막음
              disabled={true}
            >
              추가
            </button>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[90]"
      onClick={onClose}
    >
      <div
        className={`relative w-full max-w-2xl ${theme.colors.card} ${theme.effects.borderRadius} shadow-xl max-h-[90vh] flex flex-col overflow-hidden`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header (변경 없음) */}
        <div className="flex items-center justify-between px-6 pt-6 pb-4  flex-shrink-0">
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
              {/* Title, Content (유지) */}
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
                        fieldOptionsLookup?.stages?.find((s) => s.stageId === selectedStageId) && (
                          <>
                            <span
                              className="w-3 h-3 rounded-full"
                              style={{
                                backgroundColor:
                                  fieldOptionsLookup?.stages?.find(
                                    (s) => s.stageId === selectedStageId,
                                  )?.color || '#6B7280',
                              }}
                            />
                            {
                              fieldOptionsLookup?.stages?.find((s) => s.stageId === selectedStageId)
                                ?.label
                            }
                          </>
                        )}
                    </span>
                    <CheckSquare className="w-4 h-4 text-gray-400" />
                  </button>
                  {/* 드롭다운 메뉴 */}
                  {showStageDropdown && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                      {fieldOptionsLookup?.stages?.map((stage) => (
                        <button
                          key={stage.stageId}
                          type="button"
                          onClick={() => {
                            setSelectedStageId(stage.stageId);
                            setShowStageDropdown(false);
                          }}
                          className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                            selectedStageId === stage.stageId ? 'bg-blue-50' : ''
                          }`}
                        >
                          <span
                            className="w-3 h-3 rounded-full"
                            style={{ backgroundColor: stage.color || '#6B7280' }}
                          />
                          {stage.label}
                        </button>
                      ))}
                      <button
                        type="button"
                        // 💡 [수정] 인라인 생성 기능 비활성화
                        onClick={() => {
                          setShowStageDropdown(false);
                          handleCreateCustomField('stage');
                        }}
                        className="w-full px-3 py-2 text-left transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2 disabled:text-gray-400 disabled:cursor-not-allowed"
                        disabled={true}
                      >
                        <Plus className="w-4 h-4" />+ 새 진행 단계 추가
                      </button>
                    </div>
                  )}
                </div>

                {/* Role Selection */}
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
                        fieldOptionsLookup?.roles?.find((r) => r.roleId === selectedRoleId) && (
                          <>
                            <span
                              className="w-3 h-3 rounded-full"
                              style={{
                                backgroundColor:
                                  fieldOptionsLookup.roles.find((r) => r.roleId === selectedRoleId)
                                    ?.color || '#6B7280',
                              }}
                            />
                            {
                              fieldOptionsLookup.roles.find((r) => r.roleId === selectedRoleId)
                                ?.label
                            }
                          </>
                        )}
                    </span>
                    <Tag className="w-4 h-4 text-gray-400" />
                  </button>
                  {/* 드롭다운 메뉴 */}
                  {showRoleDropdown && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                      {fieldOptionsLookup?.roles?.map((role) => (
                        <button
                          key={role.roleId}
                          type="button"
                          onClick={() => {
                            setSelectedRoleId(role.roleId);
                            setShowRoleDropdown(false);
                          }}
                          className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                            selectedRoleId === role.roleId ? 'bg-blue-50' : ''
                          }`}
                        >
                          <span
                            className="w-3 h-3 rounded-full"
                            style={{ backgroundColor: role.color || '#6B7280' }}
                          />
                          {role.label}
                        </button>
                      ))}
                      <button
                        type="button"
                        // 💡 [수정] 인라인 생성 기능 비활성화
                        onClick={() => {
                          setShowRoleDropdown(false);
                          handleCreateCustomField('role');
                        }}
                        className="w-full px-3 py-2 text-left transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2 disabled:text-gray-400 disabled:cursor-not-allowed"
                        disabled={true}
                      >
                        <Plus className="w-4 h-4" />+ 새 역할 추가
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
                    중요도 (선택)
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
                          (i) => i.importanceId === selectedImportanceId,
                        ) && (
                          <>
                            <span
                              className="w-3 h-3 rounded-full"
                              style={{
                                backgroundColor:
                                  fieldOptionsLookup.importances.find(
                                    (i) => i.importanceId === selectedImportanceId,
                                  )?.color || '#6B7280',
                              }}
                            />
                            {
                              fieldOptionsLookup.importances.find(
                                (i) => i.importanceId === selectedImportanceId,
                              )?.label
                            }
                          </>
                        )
                      ) : (
                        <span className="text-gray-500">없음</span>
                      )}
                    </span>
                    <AlertCircle className="w-4 h-4 text-gray-400" />
                  </button>
                  {/* 드롭다운 메뉴 */}
                  {showImportanceDropdown && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto">
                      <button
                        type="button"
                        onClick={() => {
                          setSelectedImportanceId('');
                          setShowImportanceDropdown(false);
                        }}
                        className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                          selectedImportanceId === '' ? 'bg-blue-50' : ''
                        }`}
                      >
                        <span className="w-3 h-3 rounded-full bg-gray-300" />
                        없음
                      </button>
                      {fieldOptionsLookup?.importances?.map((importance) => (
                        <button
                          key={importance.importanceId}
                          type="button"
                          onClick={() => {
                            setSelectedImportanceId(importance.importanceId);
                            setShowImportanceDropdown(false);
                          }}
                          className={`w-full px-3 py-2 text-left hover:bg-gray-100 transition text-sm flex items-center gap-2 ${
                            selectedImportanceId === importance.importanceId ? 'bg-blue-50' : ''
                          }`}
                        >
                          <span
                            className="w-3 h-3 rounded-full"
                            style={{ backgroundColor: importance.color || '#6B7280' }}
                          />
                          {importance.label}
                        </button>
                      ))}
                      <button
                        type="button"
                        // 💡 [수정] 인라인 생성 기능 비활성화
                        onClick={() => {
                          setShowImportanceDropdown(false);
                          handleCreateCustomField('importance');
                        }}
                        className="w-full px-3 py-2 text-left transition text-sm text-blue-600 font-medium border-t border-gray-200 flex items-center gap-2 disabled:text-gray-400 disabled:cursor-not-allowed"
                        disabled={true}
                      >
                        <Plus className="w-4 h-4" />+ 새 중요도 추가
                      </button>
                    </div>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-2">
                    <Plus className="w-4 h-4 inline mr-1" />
                    커스텀 필드 추가
                  </label>
                  <button
                    type="button" // 💡 [수정] 명시적으로 버튼 타입 지정 (폼 충돌 방지)
                    onClick={onAddFieldsClick} // 💡 [수정] Prop 호출
                    className="w-full px-3 py-2 border border-dashed border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition text-sm text-left flex items-center justify-between font-medium"
                    disabled={isLoading}
                  >
                    {/* 💡 [수정] 텍스트 및 아이콘 수정 */}
                    <span className="text-gray-600">+ 새 필드 유형 정의</span>
                    <Settings className="w-4 h-4 text-gray-400" />
                  </button>
                </div>
              </div>
              {/* Actions (변경 없음) */}
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
                  {/* 💡 [수정] isEditMode 변수 사용 */}
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

      {/* Creation Modals (API 제거로 임시 비활성화) */}
      {showCreateStage && renderCreationModal('stage', '진행 단계')}
      {showCreateRole && renderCreationModal('role', '역할')}
      {showCreateImportance && renderCreationModal('importance', '중요도')}
    </div>
  );
};
