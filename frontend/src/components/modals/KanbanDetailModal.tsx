import React, { useState, useCallback, useMemo } from 'react';
import { X, Calendar, Tag, User, Settings, Send, Trash, Plus } from 'lucide-react';
import { CustomField, Kanban, KanbanWithCustomFields, Priority } from '../../types/kanban';

// =============================================================================
// MOCK TYPES & CONTEXTS (실제 앱에서는 별도 파일에 정의됨)
// =============================================================================

// 💡 Mock Theme Context (Tailwind CSS를 위한 최소한의 테마 정의)
const useTheme = () => ({
  theme: {
    font: {
      size: { xs: 'text-xs', sm: 'text-sm', base: 'text-lg' },
    },
    colors: {
      primary: 'bg-blue-600',
      primaryText: 'text-blue-600',
      primaryHover: 'hover:bg-blue-700',
      card: 'bg-white',
      border: 'border-gray-200',
      subText: 'text-gray-500',
      success: 'bg-green-500',
    },
    effects: {
      borderRadius: 'rounded-xl',
      cardBorderWidth: 'border',
    },
  },
});

// 💡 Mock Kanban, CustomField, Comment Types
interface Comment {
  id: number;
  author: string;
  authorId: string;
  content: string;
  timestamp: string;
}

interface KanbanDetailModalProps {
  kanban: KanbanWithCustomFields;
  onClose: () => void;
  // onSave: (updatedKanban: KanbanWithCustomFields) => void; // 실제 구현 시 사용
}

// =============================================================================
// CUSTOM FIELD MODAL COMPONENT (Dependency)
// =============================================================================

interface CustomFieldModalProps {
  initialField?: CustomField;
  onSave: (field: CustomField) => void;
  onClose: () => void;
}

const CustomFieldModal: React.FC<CustomFieldModalProps> = ({ initialField, onSave, onClose }) => {
  const { theme } = useTheme();
  const [field, setField] = useState<CustomField>(
    initialField || {
      id: `cf-${Date.now()}`,
      name: '',
      type: 'TEXT',
      options: [],
    },
  );

  const handleSave = () => {
    if (!field.name || !field.type) {
      alert('필드 이름과 타입을 지정해야 합니다.');
      return;
    }
    onSave(field);
    onClose();
  };

  const isSelect = field.type === 'SELECT';

  return (
    <div className="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center p-4 z-[60]">
      <div
        className={`relative w-full max-w-lg ${theme.colors.card} p-6 ${theme.effects.borderRadius} shadow-2xl`}
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-xl font-bold mb-4 border-b pb-2">사용자 정의 필드 설정</h2>
        <div className="space-y-4">
          <div>
            <label className={`${theme.font.size.sm} font-semibold mb-1 block`}>필드 이름</label>
            <input
              type="text"
              value={field.name}
              onChange={(e) => setField({ ...field, name: e.target.value })}
              className="w-full px-3 py-2 border rounded"
              placeholder="예: 스프린트 번호, QA 담당자"
            />
          </div>
          <div>
            <label className={`${theme.font.size.sm} font-semibold mb-1 block`}>필드 타입</label>
            <select
              value={field.type}
              onChange={(e) => {
                setField({ ...field, type: e.target.value as CustomField['type'] });
              }}
              className="w-full px-3 py-2 border rounded"
            >
              <option value="TEXT">텍스트 (Text)</option>
              <option value="NUMBER">숫자 (Number)</option>
              <option value="DATE">날짜 (Date)</option>
              <option value="PERSON">담당자 (Person)</option>
              <option value="SELECT">선택 목록 (Select)</option>
            </select>
          </div>

          {isSelect && (
            <div className="border p-3 rounded bg-gray-50">
              <h3 className="font-semibold mb-2">선택 옵션 ({field.options?.length || 0})</h3>
              {field.options?.map((opt, index) => (
                <div key={index} className="flex items-center gap-2 mb-1">
                  <input
                    type="text"
                    value={opt.value}
                    onChange={(e) => {
                      const newOptions = [...(field.options || [])];
                      newOptions[index].value = e.target.value;
                      setField({ ...field, options: newOptions });
                    }}
                    className="flex-1 px-2 py-1 border rounded text-sm"
                  />
                  <button
                    onClick={() =>
                      setField({
                        ...field,
                        options: field.options?.filter((_, i) => i !== index),
                      })
                    }
                    className="text-red-500 hover:text-red-700 text-sm"
                  >
                    <Trash className="w-4 h-4" />
                  </button>
                </div>
              ))}
              <button
                onClick={() =>
                  setField({
                    ...field,
                    options: [...(field.options || []), { value: '', isDefault: false }],
                  })
                }
                className="mt-2 text-blue-500 hover:text-blue-700 text-sm flex items-center gap-1"
              >
                <Plus className="w-4 h-4" /> 옵션 추가
              </button>
            </div>
          )}

          {/* 추가 옵션 (Select일 때만 보임) */}
          {(isSelect || field.type === 'PERSON') && (
            <div className="flex items-center">
              <input
                type="checkbox"
                id="allowMultipleSections"
                checked={field.allowMultipleSections || false}
                onChange={(e) => setField({ ...field, allowMultipleSections: e.target.checked })}
                className="mr-2"
              />
              <label htmlFor="allowMultipleSections" className={`${theme.font.size.sm}`}>
                다중 값 허용 (쉼표로 구분)
              </label>
            </div>
          )}
        </div>
        <div className="flex justify-end gap-3 mt-6">
          <button
            onClick={onClose}
            className="bg-gray-200 text-gray-700 py-2 px-4 rounded-lg hover:bg-gray-300 transition"
          >
            취소
          </button>
          <button
            onClick={handleSave}
            className={`${theme.colors.primary} text-white py-2 px-4 rounded-lg ${theme.colors.primaryHover} transition`}
          >
            저장
          </button>
        </div>
      </div>
    </div>
  );
};

// =============================================================================
// MAIN MODAL COMPONENT
// =============================================================================

const KanbanDetailModal: React.FC<KanbanDetailModalProps> = ({ kanban, onClose }) => {
  const { theme } = useTheme();
  const isCreating = kanban.id === '';

  // 💡 Mock Custom Fields State
  const [customFields, setCustomFields] = useState<CustomField[]>([
    {
      id: 'cf-status',
      name: '커스텀 진행단계',
      type: 'SELECT',
      options: [
        { value: 'TO DO', isDefault: true },
        { value: 'IN PROGRESS', isDefault: false },
        { value: 'QA', isDefault: false },
      ],
      allowMultipleSections: false,
    },
    { id: 'cf-sprint', name: '스프린트 번호', type: 'NUMBER' },
  ]);

  // 💡 Mock Comments State
  const [comments, setComments] = useState<Comment[]>(
    kanban.id
      ? [
          {
            id: 1,
            author: '김개발',
            authorId: 'kim1',
            content: '데이터 수집 범위에 대해 검토가 필요합니다.',
            timestamp: '2일 전',
          },
          {
            id: 2,
            author: '박기획',
            authorId: 'park2',
            content: '프론트 디자인 시안 초안 공유했습니다.',
            timestamp: '1일 전',
          },
        ]
      : [],
  );
  const [newComment, setNewComment] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showCustomFieldModal, setShowCustomFieldModal] = useState(false);
  const [fieldToEdit, setFieldToEdit] = useState<CustomField | undefined>(undefined);

  // 💡 Kanban State with Custom Field Values initialization
  const [currentKanban, setCurrentKanban] = useState<KanbanWithCustomFields>(() => ({
    ...kanban,
    customFieldValues: kanban.customFieldValues || {},
    title: kanban.title || '',
    assignee: kanban.assignee || '사용자 본인',
    assignee_id: kanban.assignee_id || 'user_id_123',
    status: kanban.status || 'BACKLOG',
    dueDate: kanban.dueDate || '',
    priority: kanban.priority || 'MEDIUM',
    description: kanban.description || '',
  }));

  // =============================================================================
  // HANDLERS
  // =============================================================================

  const handleFieldChange = React.useCallback(
    <T extends keyof Kanban>(field: T, value: Kanban[T]) => {
      setCurrentKanban((prev) => ({ ...prev, [field]: value }));
    },
    [setCurrentKanban],
  );

  const handleCustomFieldChange = useCallback((fieldId: string, value: any) => {
    setCurrentKanban((prev) => ({
      ...prev,
      customFieldValues: {
        ...prev.customFieldValues,
        [fieldId]: value,
      },
    }));
  }, []);

  const handleSaveCustomField = useCallback((newField: CustomField) => {
    setCustomFields((prev) => {
      const existingIndex = prev.findIndex((f) => f.id === newField.id);
      if (existingIndex > -1) {
        return prev.map((f, i) => (i === existingIndex ? newField : f));
      }
      return [...prev, newField];
    });
    setFieldToEdit(undefined);
  }, []);

  const handleAddComment = () => {
    if (!newComment.trim() || isLoading) return;

    const authorName = currentKanban.assignee || '익명 사용자';
    setComments((prev) => [
      ...prev,
      {
        id: prev.length + 1,
        author: authorName,
        authorId: currentKanban.assignee_id,
        content: newComment,
        timestamp: '방금 전',
      },
    ]);
    setNewComment('');
  };

  const handleSave = () => {
    if (!currentKanban.title.trim()) {
      // alert() 대신 커스텀 모달을 사용해야 하지만, 여기서는 mock으로 alert을 사용합니다.
      alert('제목은 필수입니다.');
      return;
    }

    setIsLoading(true);

    // 🚧 [Mock API 호출]
    setTimeout(() => {
      // 부모 컴포넌트에 최종 데이터 전달 (추후 구현)
      // onSave(currentKanban);
      const action = isCreating ? '생성' : '수정 및 저장';
      alert(`[Mock] 칸반 '${currentKanban.title}' ${action} 완료!`);

      setIsLoading(false);
      onClose();
    }, 800);
  };

  const handleDelete = () => {
    if (window.confirm(`정말로 칸반 "${currentKanban.title}"을(를) 삭제하시겠습니까?`)) {
      alert(`[Mock] 칸반 삭제 처리 완료.`);
      onClose();
    }
  };

  // =============================================================================
  // RENDERING HELPERS
  // =============================================================================

  const priorityMap = useMemo(
    () => ({
      HIGH: '높음 (High)',
      MEDIUM: '보통 (Medium)',
      LOW: '낮음 (Low)',
      '': '선택 사항',
    }),
    [],
  );

  // 💡 Custom Field 렌더링 함수
  const renderCustomField = (field: CustomField) => {
    const currentValue = currentKanban.customFieldValues?.[field.id] || field.defaultValue || '';

    // 다중 선택 값을 쉼표로 분리하여 표시 (SELECT + allowMultipleSections)
    const displayValue = Array.isArray(currentValue) ? currentValue.join(', ') : currentValue;

    // 입력/선택 필드 렌더링 로직
    const inputField = () => {
      const baseClasses = `w-full px-3 py-2 border ${theme.colors.border} bg-gray-50 text-sm ${theme.effects.borderRadius} focus:ring-2 focus:ring-blue-500`;

      switch (field.type) {
        case 'TEXT':
        case 'PERSON':
          return (
            <input
              type="text"
              value={displayValue}
              onChange={(e) => handleCustomFieldChange(field.id, e.target.value)}
              className={baseClasses}
              placeholder={field.type === 'PERSON' ? '담당자 이름 입력...' : '텍스트 입력'}
            />
          );
        case 'NUMBER':
          return (
            <input
              type="number"
              value={displayValue}
              onChange={(e) => handleCustomFieldChange(field.id, Number(e.target.value))}
              className={baseClasses}
            />
          );
        case 'DATE':
          return (
            <input
              type="date"
              value={displayValue}
              onChange={(e) => handleCustomFieldChange(field.id, e.target.value)}
              className={baseClasses}
            />
          );
        case 'SELECT':
          if (field.allowMultipleSections) {
            // 다중 선택 (Mock: 텍스트 입력 후 쉼표로 분리)
            return (
              <input
                type="text"
                value={displayValue}
                onChange={(e) =>
                  handleCustomFieldChange(
                    field.id,
                    e.target.value.split(',').map((v) => v.trim()),
                  )
                }
                placeholder="값들을 쉼표(,)로 구분하여 입력"
                className={baseClasses}
              />
            );
          }
          return (
            <select
              value={displayValue}
              onChange={(e) => handleCustomFieldChange(field.id, e.target.value)}
              className={baseClasses}
            >
              <option value="" disabled>
                선택하세요
              </option>
              {field.options?.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.value} {opt.isDefault ? '(기본)' : ''}
                </option>
              ))}
            </select>
          );
        default:
          return null;
      }
    };

    return (
      <div key={field.id} className="w-full">
        <label className={`${theme.font.size.xs} mb-2 ${theme.colors.subText} font-semibold block`}>
          {field.name}
        </label>
        <div className="flex items-center gap-2">
          <div className="flex-1">{inputField()}</div>
          <button
            onClick={() => {
              setFieldToEdit(field);
              setShowCustomFieldModal(true);
            }}
            className="p-1 text-gray-400 hover:text-blue-600 transition flex-shrink-0"
            title="필드 설정/수정"
            disabled={isLoading}
          >
            <Settings className="w-4 h-4" />
          </button>
        </div>
      </div>
    );
  };

  // 💡 시스템 필드 렌더링 함수 (Assignee, DueDate, Priority)
  const renderSystemField = ({
    id,
    label,
    icon: Icon,
    input,
  }: {
    id: string;
    label: string;
    icon: React.ElementType;
    input: React.ReactNode;
  }) => (
    <div key={id} className="w-full">
      <label
        className={`flex items-center gap-1 ${theme.font.size.xs} mb-2 ${theme.colors.subText} font-semibold`}
      >
        <Icon className="w-4 h-4" />
        {label}
      </label>
      <div className="flex items-center gap-2">
        <div className="flex-1">{input}</div>
        <button
          onClick={() => setShowCustomFieldModal(true)} // Mock: 커스텀 필드 모달로 연결
          className="p-1 text-gray-400 hover:text-red-500 transition flex-shrink-0"
          title="시스템 필드 설정을 변경하려면 프로젝트 설정에서 진행하세요"
          disabled={isLoading}
        >
          <Settings className="w-4 h-4" />
        </button>
      </div>
    </div>
  );

  // 💡 System Fields List
  const systemFields = useMemo(
    () => [
      // 1. 담당자 필드 (Assignee)
      {
        id: 'assignee',
        label: '담당자',
        icon: User,
        input: (
          <input
            type="text"
            value={currentKanban.assignee || ''}
            onChange={(e) => handleFieldChange('assignee', e.target.value)}
            placeholder="담당자 이름 검색..."
            className={`w-full px-3 py-2 border ${theme.colors.border} bg-gray-50 ${theme.font.size.sm} ${theme.effects.borderRadius} font-medium focus:outline-none focus:ring-2 focus:ring-blue-500`}
            disabled={isLoading}
          />
        ),
      },
      // 2. 중요도 필드 (Priority)
      {
        id: 'priority',
        label: '중요도 (우선 순위)',
        icon: Tag,
        input: (
          <select
            value={currentKanban.priority}
            onChange={(e) => handleFieldChange('priority', e.target.value as Priority)}
            className={`w-full px-3 py-2 border ${theme.colors.border} bg-gray-50 ${theme.font.size.sm} ${theme.effects.borderRadius} font-bold focus:outline-none focus:ring-2 focus:ring-blue-500`}
            disabled={isLoading}
          >
            {Object.entries(priorityMap).map(([key, value]) => (
              <option key={key} value={key}>
                {value}
              </option>
            ))}
          </select>
        ),
      },
      // 3. 마감일 필드 (DueDate)
      {
        id: 'dueDate',
        label: '마감일',
        icon: Calendar,
        input: (
          <input
            type="date"
            value={currentKanban.dueDate}
            onChange={(e) => handleFieldChange('dueDate', e.target.value)}
            className={`w-full px-3 py-2 border ${theme.colors.border} bg-gray-50 ${theme.font.size.sm} ${theme.effects.borderRadius} font-medium focus:outline-none focus:ring-2 focus:ring-blue-500`}
            disabled={isLoading}
          />
        ),
      },
    ],
    [
      currentKanban,
      handleFieldChange,
      isLoading,
      priorityMap,
      theme.colors.border,
      theme.effects.borderRadius,
      theme.font.size.sm,
    ],
  );

  return (
    <>
      <div
        className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50 overflow-y-auto"
        onClick={onClose}
      >
        <div className="relative w-full max-w-2xl my-8" onClick={(e) => e.stopPropagation()}>
          <div
            className={`relative ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} p-4 sm:p-6 max-h-[90vh] overflow-y-auto ${theme.effects.borderRadius} shadow-xl`}
          >
            {/* --- 1. 제목 및 닫기 버튼 섹션 --- */}
            <div className={`flex items-start justify-between mb-4 pb-4`}>
              <div className="flex-1 pr-4">
                {/* 제목 입력 필드 */}
                <input
                  type="text"
                  value={currentKanban.title}
                  onChange={(e) => handleFieldChange('title', e.target.value)}
                  placeholder={isCreating ? '새 칸반 제목을 입력하세요 (필수)' : '제목'}
                  className={`w-full ${
                    theme.font.size.base
                  } font-bold mb-1 break-words focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                    isCreating ? 'border-b-2 border-blue-200' : 'bg-transparent'
                  }`}
                  disabled={isLoading}
                />
                {/* <div className={`${theme.font.size.sm} ${theme.colors.subText}`}>
                  컬럼: <span className="font-semibold">{currentKanban.status}</span>
                </div> */}
              </div>
              <button
                onClick={onClose}
                className="p-2 rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* --- 2. 상세 정보 및 필드 섹션 --- */}
            <div className="space-y-4 mb-3 border-b border-gray-200 pb-4">
              {/* 시스템 필드 및 커스텀 필드 렌더링 영역 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* 💡 시스템 필드 (담당자, 중요도, 마감일) */}
                {systemFields.map(renderSystemField)}

                {/* 💡 커스텀 필드 */}
                {customFields.map(renderCustomField)}
              </div>

              {/* 💡 새 필드 추가 버튼 */}
              <button
                onClick={() => {
                  setFieldToEdit(undefined); // 새 필드 추가
                  setShowCustomFieldModal(true);
                }}
                className="w-full text-blue-600 hover:text-blue-800 text-sm font-semibold border-dashed border-2 border-blue-200 hover:border-blue-400 p-2 rounded-lg mt-2 transition flex items-center justify-center gap-2"
                disabled={isLoading}
              >
                <Plus className="w-4 h-4" /> 사용자 정의 필드 추가
              </button>

              {/* 💡 상세 설명 (Description) */}
              <div className="pt-4">
                <label
                  className={`${theme.font.size.xs} mb-2 ${theme.colors.subText} font-semibold block`}
                >
                  상세 내용:
                </label>
                <textarea
                  value={currentKanban.description}
                  onChange={(e) => handleFieldChange('description', e.target.value)}
                  placeholder="상세 내용 및 목표를 입력하세요."
                  className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} bg-gray-50 ${theme.font.size.sm} min-h-24 ${theme.effects.borderRadius} resize-none focus:outline-none focus:ring-2 focus:ring-blue-500`}
                  disabled={isLoading}
                />
              </div>
            </div>

            {/* --- 3. 댓글 섹션 (생성 모드에서는 댓글 비활성화) --- */}
            {!isCreating && (
              <div className="mb-6 space-y-4">
                <h3 className="font-bold text-gray-700 pb-2">댓글 ({comments.length})</h3>

                {/* 댓글 목록 */}
                <div className="max-h-60 overflow-y-auto space-y-3 pr-2">
                  {comments.length === 0 ? (
                    <p className={`${theme.font.size.sm} ${theme.colors.subText}`}>
                      아직 댓글이 없습니다. 첫 댓글을 작성해보세요!
                    </p>
                  ) : (
                    comments.map((comment) => (
                      <div key={comment.id} className="flex gap-3">
                        <div
                          className={`w-8 h-8 ${theme.colors.primary} flex items-center justify-center text-white ${theme.font.size.xs} font-bold rounded-full flex-shrink-0`}
                        >
                          {comment.author[0]}
                        </div>
                        <div className="flex-1 p-3 bg-gray-50 rounded-lg">
                          <div className="flex justify-between items-center mb-1">
                            <span className="font-semibold text-sm">{comment.author}</span>
                            <span className={`${theme.font.size.xs} ${theme.colors.subText}`}>
                              {comment.timestamp}
                            </span>
                          </div>
                          <p className={theme.font.size.sm}>{comment.content}</p>
                        </div>
                      </div>
                    ))
                  )}
                </div>

                {/* 댓글 입력 필드 */}
                <div className="flex gap-2 pt-2">
                  <textarea
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                    placeholder="댓글을 입력하세요..."
                    rows={1}
                    className={`flex-1 px-3 py-2 border ${theme.colors.border} bg-gray-50 ${theme.font.size.sm} ${theme.effects.borderRadius} resize-none focus:outline-none focus:ring-2 focus:ring-blue-500`}
                    disabled={isLoading}
                  />
                  <button
                    onClick={handleAddComment}
                    disabled={!newComment.trim() || isLoading}
                    className={`${theme.colors.primary} text-white p-3 ${theme.colors.primaryHover} ${theme.effects.borderRadius} transition disabled:opacity-50 flex items-center justify-center`}
                  >
                    <Send className="w-4 h-4" />
                  </button>
                </div>
              </div>
            )}

            {/* --- 4. 액션 버튼 --- */}
            <div className={`flex gap-3 mt-6 pt-4`}>
              <button
                onClick={handleSave}
                disabled={isLoading || !currentKanban.title.trim()}
                className={`flex-1 ${theme.colors.primary} text-white py-3 font-bold ${theme.colors.primaryHover} transition ${theme.font.size.sm} ${theme.effects.borderRadius} disabled:opacity-50`}
              >
                {isLoading ? '처리 중...' : isCreating ? '생성' : '수정 및 저장'}
              </button>

              {!isCreating && (
                <button
                  onClick={handleDelete}
                  className={`bg-red-500 text-white px-4 py-3 font-bold hover:bg-red-600 transition ${theme.font.size.sm} ${theme.effects.borderRadius} disabled:opacity-50`}
                  disabled={isLoading}
                >
                  삭제
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 💡 CustomFieldModal 렌더링 */}
      {showCustomFieldModal && (
        <CustomFieldModal
          initialField={fieldToEdit}
          onSave={handleSaveCustomField}
          onClose={() => {
            setShowCustomFieldModal(false);
            setFieldToEdit(undefined);
          }}
        />
      )}
    </>
  );
};

export default KanbanDetailModal;
