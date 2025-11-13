// src/components/modals/CustomFieldAddModal.tsx

import React, { useState, useCallback } from 'react';
import { X, ChevronDown, Check, Tag, Menu, Trash2, Plus } from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import { CreateFieldRequest, FieldResponse } from '../../../types/board';
import { MODERN_CUSTOM_FIELD_COLORS } from './constants/colors';
// 💡 [S3 설정 및 API 호출 함수 import는 생략]

// 💡 [수정] 옵션 상태 구조 정의
interface FieldOption {
  label: string;
  color: string;
}

// 💡 필드 유형 정의 (유지)
const FIELD_TYPES = [
  { type: 'text', label: '텍스트', icon: '01' },
  { type: 'number', label: '숫자', icon: '02' },
  { type: 'single_select', label: '선택', icon: '03' },
  { type: 'date', label: '날짜', icon: '04' },
  { type: 'single_user', label: '담당자/사용자', icon: '05' },
];

interface CustomFieldAddModalProps {
  projectId: string;
  onClose: () => void;
  onFieldCreated: (newField: FieldResponse | null) => void;
}

export const CustomFieldAddModal: React.FC<CustomFieldAddModalProps> = ({
  projectId,
  onClose,
  onFieldCreated,
}) => {
  const { theme } = useTheme();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 필드 상태 통합
  const [fieldType, setFieldType] = useState<CreateFieldRequest['fieldType'] | ''>('single_select');
  const [fieldName, setFieldName] = useState('');
  const [fieldOptions, setFieldOptions] = useState<FieldOption[]>([]);
  const [newOption, setNewOption] = useState('');
  const [decimalPlaces, setDecimalPlaces] = useState<number | null>(null);

  // 옵션 편집 상태
  const [editingOption, setEditingOption] = useState<{ option: FieldOption; index: number } | null>(
    null,
  );

  // 드래그 상태
  const [draggedOption, setDraggedOption] = useState<FieldOption | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  const selectedTypeObj = FIELD_TYPES.find((t) => t.type === fieldType);

  // 💡 [수정] 옵션 추가 핸들러: 색상 지정 및 Enter 키 처리
  const handleAddOption = (
    e: React.KeyboardEvent<HTMLInputElement> | React.MouseEvent<HTMLButtonElement>,
  ) => {
    if ('key' in e && e.key !== 'Enter') return;

    e.preventDefault();
    const optionText = newOption.trim();
    if (!optionText) return;

    if (fieldOptions.some((opt) => opt.label.toLowerCase() === optionText.toLowerCase())) {
      setError(`옵션 '${optionText}'은(는) 이미 존재합니다.`);
      setNewOption('');
      return;
    }

    const defaultColor = MODERN_CUSTOM_FIELD_COLORS[0].hex;
    setFieldOptions((prev) => [...prev, { label: optionText, color: defaultColor }]);
    setNewOption('');
    setError(null);
  };

  // 💡 옵션 삭제 핸들러
  const handleRemoveOption = useCallback((optionToRemove: FieldOption) => {
    setFieldOptions((prev) => prev.filter((opt) => opt.label !== optionToRemove.label));
  }, []);

  // 💡 저장 핸들러 (Mock 구현 유지)
  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!fieldType || !fieldName.trim()) {
      setError('필드 유형과 필드 이름은 필수입니다.');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // ⚠️ [TODO: API 호출]
      const mockNewField: FieldResponse = {
        fieldId: `new-field-${Date.now()}`,
        projectId: projectId,
        name: fieldName.trim(),
        description: '사용자 정의 필드',
        fieldType: fieldType as CreateFieldRequest['fieldType'],
        isRequired: false,
        isSystemDefault: false,
        displayOrder: 100,
        config: {
          options: fieldOptions,
          decimal: decimalPlaces,
        },
      };

      setTimeout(() => {
        onFieldCreated(mockNewField);
        onClose();
      }, 500);
    } catch (err) {
      setError('필드 생성에 실패했습니다.');
    } finally {
      // setLoading(false);
    }
  };

  // 💡 [추가] 드래그 앤 드롭 핸들러
  const handleDragStart = (option: FieldOption, index: number) => {
    setDraggedOption(option);
  };

  const handleDrop = (targetIndex: number) => {
    if (!draggedOption) return;

    const newOptions = [...fieldOptions];
    const draggedIndex = newOptions.findIndex((opt) => opt.label === draggedOption.label);

    if (draggedIndex === -1) return;

    // 옵션 순서 변경
    const [removed] = newOptions.splice(draggedIndex, 1);
    newOptions.splice(targetIndex, 0, removed);

    setFieldOptions(newOptions);
    setDraggedOption(null);
    setDragOverIndex(null);
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    setDragOverIndex(index);
  };

  // ========================================
  // 렌더링 헬퍼: 동적 필드 유형에 따른 콘텐츠
  // ========================================
  const renderDynamicFields = () => {
    switch (fieldType) {
      case 'single_select':
      case 'multi_select':
        return (
          <div className="space-y-4">
            {/* 옵션 입력 섹션 */}
            <div className="space-y-2">
              <label className="block text-sm font-semibold text-gray-700">옵션 추가</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newOption}
                  onChange={(e) => setNewOption(e.target.value)}
                  onKeyDown={handleAddOption}
                  placeholder="입력하고 Enter를 눌러 추가"
                  className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm`}
                  disabled={loading}
                />
                <button
                  type="button"
                  onClick={handleAddOption}
                  className="px-3 py-1 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition text-sm font-medium"
                  disabled={loading || !newOption.trim()}
                >
                  +
                </button>
              </div>
            </div>

            {/* 💡 [핵심 수정] 추가된 옵션 목록 (순서 변경 및 편집 가능) */}
            <div className="flex flex-col gap-1.5 pt-1 max-h-40 overflow-y-auto border border-gray-200 p-2 rounded-md bg-gray-50">
              {fieldOptions.length === 0 ? (
                <span className="text-sm text-gray-500">옵션을 추가해주세요.</span>
              ) : (
                fieldOptions.map((option, index) => (
                  <div
                    key={option.label}
                    draggable
                    onDragStart={() => handleDragStart(option, index)}
                    onDragOver={(e) => handleDragOver(e, index)}
                    onDrop={() => handleDrop(index)}
                    onDragEnd={() => {
                      setDraggedOption(null);
                      setDragOverIndex(null);
                    }}
                    className={`flex items-center justify-between p-2 rounded-md transition-all 
                                ${
                                  draggedOption?.label === option.label
                                    ? 'opacity-50 border-2 border-dashed border-gray-400'
                                    : 'bg-white border border-gray-200'
                                }
                                ${
                                  dragOverIndex === index
                                    ? 'border-2 border-blue-500 bg-blue-50'
                                    : ''
                                }
                            `}
                  >
                    <div className="flex items-center gap-3 cursor-move">
                      <Menu className="w-4 h-4 text-gray-400 flex-shrink-0" />
                      <span
                        className="w-4 h-4 rounded-full flex-shrink-0"
                        style={{ backgroundColor: option.color }}
                      ></span>
                      <span className="text-sm font-medium">{option.label}</span>
                    </div>

                    {/* 옵션 편집/삭제 버튼 */}
                    <div className="relative flex gap-2 items-center">
                      {/* 현재 색상 버튼 (클릭 시 팔레트 토글) */}
                      <button
                        type="button"
                        onClick={() =>
                          setEditingOption((prev) =>
                            prev?.option.label === option.label ? null : { option, index },
                          )
                        }
                        className={`px-2 py-1 text-xs rounded-md border transition-colors ${
                          editingOption?.option.label === option.label
                            ? 'bg-gray-200'
                            : 'hover:bg-gray-100'
                        }`}
                      >
                        색상
                      </button>

                      {/* 색상 선택 팔레트 드롭다운 */}
                      {editingOption?.option.label === option.label && (
                        <div className="absolute right-0 top-full mt-2 p-3 bg-white border border-gray-300 rounded-lg shadow-lg z-20 w-64">
                          <div className="grid grid-cols-8 gap-1.5">
                            {MODERN_CUSTOM_FIELD_COLORS.map((color) => (
                              <button
                                key={color.hex}
                                type="button"
                                className={`w-6 h-6 rounded-full border-2 ${
                                  option.color === color.hex
                                    ? 'ring-2 ring-blue-500'
                                    : 'hover:scale-110'
                                }`}
                                style={{ backgroundColor: color.hex }}
                                onClick={() => {
                                  // 색상 업데이트
                                  setFieldOptions((prev) =>
                                    prev.map((opt, i) =>
                                      i === index ? { ...opt, color: color.hex } : opt,
                                    ),
                                  );
                                  setEditingOption(null); // 선택 후 닫기
                                }}
                                title={color.name}
                              />
                            ))}
                          </div>
                        </div>
                      )}

                      <button
                        type="button"
                        onClick={() => handleRemoveOption(option)}
                        className="p-1 rounded-md hover:bg-red-100 text-red-600"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* 기본값 드롭다운 */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">기본값</label>
              <select
                className={`w-full px-3 py-2 border border-gray-300 rounded-lg bg-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500`}
                disabled={loading || fieldOptions.length === 0}
              >
                <option value="">옵션을 선택해주세요.</option>
                {fieldOptions.map((option) => (
                  <option key={option.label} value={option.label}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        );

      case 'number':
        return (
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">소수 자릿수</label>
            <select
              value={decimalPlaces === null ? '없음' : decimalPlaces.toString()}
              onChange={(e) =>
                setDecimalPlaces(e.target.value === '없음' ? null : parseInt(e.target.value))
              }
              className={`w-full px-3 py-2 border border-gray-300 rounded-lg bg-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500`}
              disabled={loading}
            >
              <option value="없음">없음</option>
              <option value="1">01 숫자</option>
              <option value="2">0.01</option>
            </select>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[120]"
      onClick={onClose}
    >
      <form
        onSubmit={handleSave}
        className={`relative w-full max-w-lg ${theme.colors.card} ${theme.effects.borderRadius} shadow-xl`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6 space-y-6">
          {/* 💡 [수정] 헤더 타이틀 크기 조정 및 border 제거 */}
          <h2 className="text-xl font-bold text-gray-800">
            {selectedTypeObj ? selectedTypeObj.label : '새 필드'} 추가
          </h2>

          {/* Error Message */}
          {error && (
            <div className="p-3 bg-red-100 border border-red-300 text-red-700 text-sm rounded-lg">
              {error}
            </div>
          )}

          {/* 1. 필드 유형 선택 */}
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">필드 유형</label>
            <select
              value={fieldType}
              onChange={(e) => {
                setFieldType(e.target.value as CreateFieldRequest['fieldType']);
                setDecimalPlaces(null);
                setFieldOptions([]);
              }}
              className={`w-full px-3 py-2 border border-gray-300 rounded-lg bg-white text-sm appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500`}
              disabled={loading}
            >
              <option value="">유형 선택</option>
              {FIELD_TYPES.map((type) => (
                <option key={type.type} value={type.type}>
                  {type.icon} {type.label}
                </option>
              ))}
            </select>
          </div>

          {/* 2. 필드 이름 입력 */}
          {fieldType && (
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">필드 이름</label>
              <input
                type="text"
                value={fieldName}
                onChange={(e) => setFieldName(e.target.value)}
                placeholder="필드 이름(선택 사항)"
                className={`w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500`}
                disabled={loading}
              />
            </div>
          )}

          {/* 3. 동적 속성 섹션 */}
          {fieldType && renderDynamicFields()}
        </div>

        {/* Action Buttons */}
        <div className="p-6 border-t flex justify-end gap-3">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-gray-700 font-semibold rounded-lg hover:bg-gray-100"
            disabled={loading}
          >
            취소
          </button>
          <button
            type="submit"
            className={`px-4 py-2 bg-green-600 text-white font-semibold rounded-lg hover:bg-green-700 transition ${
              loading || !fieldName.trim() || !fieldType ? 'opacity-50 cursor-not-allowed' : ''
            }`}
            disabled={loading || !fieldName.trim() || !fieldType}
          >
            {loading ? '저장 중...' : '저장'}
          </button>
        </div>
      </form>
    </div>
  );
};
