// src/components/modals/CustomFieldManageModal.tsx

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Menu, Trash2 } from 'lucide-react';
import { useTheme } from '../../../../contexts/ThemeContext';
import {
  FieldTypeInfo,
  IEditCustomFields,
  FieldOption,
  CreateFieldOptionRequest,
} from '../../../../types/board';
import { MODERN_CUSTOM_FIELD_COLORS } from './constants/colors';
import { createFieldOption } from '../../../../api/board/boardService';

interface LocalFieldOption {
  label: string;
  color: string;
}

interface CustomFieldManageModalProps {
  projectId: string;
  editFieldData: IEditCustomFields;
  onClose: () => void;
  afterFieldCreated: (newField: any | null) => void;
  filedTypesLookup: FieldTypeInfo[];
}

export const CustomFieldManageModal: React.FC<CustomFieldManageModalProps> = ({
  projectId,
  editFieldData,
  onClose,
  afterFieldCreated,
  filedTypesLookup,
}) => {
  const { theme } = useTheme();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 💡 UI 레벨 필드 타입 (FieldTypeInfo.typeId와 매칭)
  const [fieldType, setFieldType] = useState<string>('');
  const [fieldName, setFieldName] = useState('');
  const [fieldOptions, setFieldOptions] = useState<LocalFieldOption[]>([]);
  const [newOption, setNewOption] = useState('');
  const [isRequired, _setIsRequired] = useState(false);

  const [editingOption, setEditingOption] = useState<{
    option: LocalFieldOption;
    index: number;
    targetRect: DOMRect;
  } | null>(null);

  const colorButtonRef = useRef<HTMLButtonElement>(null);

  const [draggedOption, setDraggedOption] = useState<LocalFieldOption | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  const selectedTypeObj = filedTypesLookup?.find((t) => t.typeId === fieldType);
  const isSelectType = fieldType === 'single_select' || fieldType === 'multi_select';

  const handleAddOption = (
    e: React.MouseEvent<HTMLButtonElement> | React.KeyboardEvent<HTMLInputElement>,
  ) => {
    if ('key' in e && e.key === 'Enter') {
      e.preventDefault();
    } else if ('key' in e) {
      return;
    }

    const optionText = newOption.trim();
    if (!optionText) return;

    if (fieldOptions.some((opt) => opt.label.toLowerCase() === optionText.toLowerCase())) {
      setError(`옵션 '${optionText}'은(는) 이미 존재합니다.`);
      setNewOption('');
      return;
    }

    const nextColorIndex = fieldOptions.length % MODERN_CUSTOM_FIELD_COLORS.length;
    const defaultColor = MODERN_CUSTOM_FIELD_COLORS[nextColorIndex].hex;

    setFieldOptions((prev) => [...prev, { label: optionText, color: defaultColor }]);

    setNewOption('');
    setError(null);
  };

  useEffect(() => {
    if (editFieldData) {
      console.log(editFieldData);
      setFieldName(editFieldData.name);
      setFieldType(editFieldData.fieldType);

      if (
        (editFieldData.fieldType === 'single_select' ||
          editFieldData.fieldType === 'multi_select') &&
        editFieldData.options
      ) {
        const optionsFromData = editFieldData.options.map((opt: FieldOption) => ({
          label: opt.optionLabel || opt.optionValue,
          color: (opt as any).color || MODERN_CUSTOM_FIELD_COLORS[0]?.hex,
        }));
        setFieldOptions(optionsFromData);
      }
    } else {
      setFieldName('');
      setFieldType('');
      setFieldOptions([]);
    }
  }, [editFieldData]);

  const handleRemoveOption = useCallback((optionToRemove: LocalFieldOption) => {
    setFieldOptions((prev) => prev.filter((opt) => opt.label !== optionToRemove.label));
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!fieldType || !fieldName?.trim()) {
      setError('필드 유형과 필드 이름은 필수입니다.');
      return;
    }
    if (isSelectType && fieldOptions.length === 0) {
      setError('선택 유형 필드는 최소한 하나의 옵션을 가져야 합니다.');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // 💡 UI 레벨 fieldType을 API 레벨 fieldType으로 매핑
      // fieldName을 기준으로 API fieldType 결정
      let apiFieldType: CreateFieldOptionRequest['fieldType'];

      if (fieldName === '진행 단계' || fieldName.toLowerCase().includes('stage')) {
        apiFieldType = 'stage';
      } else if (fieldName === '역할' || fieldName.toLowerCase().includes('role')) {
        apiFieldType = 'role';
      } else if (fieldName === '중요도' || fieldName.toLowerCase().includes('importance')) {
        apiFieldType = 'importance';
      } else {
        // 💡 기본값: 사용자가 입력한 fieldName을 소문자로 변환하여 사용
        const normalized = fieldName.toLowerCase();
        if (normalized.includes('중요')) {
          apiFieldType = 'importance';
        } else if (normalized.includes('역할') || normalized.includes('담당')) {
          apiFieldType = 'role';
        } else {
          apiFieldType = 'stage'; // 기본값
        }
      }

      // 💡 각 옵션을 개별적으로 생성
      if (isSelectType && fieldOptions.length > 0) {
        for (let i = 0; i < fieldOptions.length; i++) {
          const option = fieldOptions[i];
          const requestData: CreateFieldOptionRequest = {
            fieldType: apiFieldType,
            value: option.label.toLowerCase().replace(/\s+/g, '_'), // value 생성
            label: option.label,
            color: option.color,
            displayOrder: i,
          };

          await createFieldOption(requestData);
        }

        console.log('✅ 필드 옵션 생성 완료');
      }

      afterFieldCreated(null); // 💡 필드 옵션 생성 후 콜백
      onClose();
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      setError(`필드 옵션 생성에 실패했습니다: ${errorMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDragStart = (option: LocalFieldOption) => {
    setDraggedOption(option);
  };

  const handleDrop = (targetIndex: number) => {
    if (!draggedOption) return;

    const newOptions = [...fieldOptions];
    const draggedIndex = newOptions.findIndex((opt) => opt.label === draggedOption.label);

    if (draggedIndex === -1) return;

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

  const renderDynamicFields = () => {
    switch (fieldType) {
      case 'single_select':
      case 'multi_select':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="block text-sm font-semibold text-gray-700">옵션 추가</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newOption}
                  onChange={(e) => setNewOption(e.target.value)}
                  onKeyUp={(e) => {
                    if (e.key === 'Enter') handleAddOption(e);
                  }}
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

            <div className="flex flex-col gap-1.5 pt-1 max-h-40 overflow-y-auto border border-gray-200 p-2 rounded-md bg-gray-50">
              {fieldOptions.length === 0 ? (
                <span className="text-sm text-gray-500">옵션을 추가해주세요.</span>
              ) : (
                fieldOptions.map((option, index) => (
                  <div
                    key={option.label}
                    draggable
                    onDragStart={() => handleDragStart(option)}
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

                    <div className="relative flex gap-2 items-center">
                      <button
                        type="button"
                        ref={editingOption?.option.label === option.label ? colorButtonRef : null}
                        onClick={(e) => {
                          const rect = e.currentTarget.getBoundingClientRect();
                          setEditingOption((prev) =>
                            prev?.option.label === option.label
                              ? null
                              : { option, index, targetRect: rect },
                          );
                          e.stopPropagation();
                        }}
                        className={`px-2 py-1 text-xs rounded-md border transition-colors`}
                      >
                        색상
                      </button>

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
          <h2 className="text-xl font-bold text-gray-800">
            {selectedTypeObj ? selectedTypeObj.typeName : '새 필드'} 추가
          </h2>

          {error && (
            <div className="p-3 bg-red-100 border border-red-300 text-red-700 text-sm rounded-lg">
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">필드 유형</label>
            <select
              value={fieldType}
              onChange={(e) => {
                setFieldType(e.target.value);
                setFieldOptions([]);
              }}
              className={`w-full px-3 py-2 border border-gray-300 rounded-lg bg-white text-sm appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500`}
              disabled={loading}
            >
              <option value="" disabled>
                유형 선택
              </option>
              {filedTypesLookup?.map((type) => (
                <option key={type.typeId} value={type.typeId}>
                  {type.typeName}
                </option>
              ))}
            </select>
          </div>

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

          {fieldType && renderDynamicFields()}
        </div>

        <div className="p-6 flex justify-end gap-3">
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

      {editingOption && (
        <ColorPickerPortal
          option={editingOption.option}
          index={editingOption.index}
          targetRect={editingOption.targetRect}
          setFieldOptions={setFieldOptions}
          onClose={() => setEditingOption(null)}
        />
      )}
    </div>
  );
};

interface ColorPickerPortalProps {
  option: LocalFieldOption;
  index: number;
  targetRect: DOMRect;
  setFieldOptions: React.Dispatch<React.SetStateAction<LocalFieldOption[]>>;
  onClose: () => void;
}

const ColorPickerPortal: React.FC<ColorPickerPortalProps> = ({
  option,
  index,
  targetRect,
  setFieldOptions,
  onClose,
}) => {
  const handleColorSelect = (newColor: string) => {
    setFieldOptions((prev) =>
      prev.map((opt, i) => (i === index ? { ...opt, color: newColor } : opt)),
    );
    onClose();
  };

  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest('.color-picker-palette') || target.closest('.color-button-trigger')) {
        return;
      }
      onClose();
    };
    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [onClose]);

  return (
    <div
      className="fixed color-picker-palette z-[150] w-64 p-3 bg-white border border-gray-300 rounded-lg shadow-xl"
      style={{
        top: targetRect.bottom + 5,
        left: targetRect.left - 180,
      }}
      onMouseDown={(e) => e.stopPropagation()}
    >
      <div className="grid grid-cols-8 gap-1.5">
        {MODERN_CUSTOM_FIELD_COLORS.map((color) => (
          <button
            key={color.hex}
            type="button"
            className={`w-6 h-6 rounded-full border-2 ${
              option.color === color.hex ? 'ring-2 ring-blue-500' : 'hover:scale-110'
            }`}
            style={{ backgroundColor: color.hex }}
            onClick={() => handleColorSelect(color.hex)}
            title={color.name}
          />
        ))}
      </div>
      <p className="mt-2 text-xs text-gray-500">색상 선택</p>
    </div>
  );
};
