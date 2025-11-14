// src/components/modals/CustomFieldManagerModal.tsx

import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  Edit2,
  Plus,
  Trash2,
  Tag,
  CheckSquare,
  AlertCircle,
  Calendar,
  User,
  Menu,
} from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';

import { FieldResponse, FieldOptionResponse } from '../../../types/board';
import {
  getProjectFields,
  getFieldOptions,
  // updateField,
  // updateFieldOption,
  deleteField,
  deleteFieldOption,
} from '../../../api/board/boardService';
import { LoadingSpinner } from '../../common/LoadingSpinner';
import { MOCK_FIELD_OPTIONS, MOCK_FIELDS } from '../../../mocks/board';

interface CustomFieldManagerModalProps {
  projectId: string;
  onClose: () => void;
  onFieldsUpdated: () => void; // 필드 구조가 변경되었음을 상위 컴포넌트에 알림
}

export const CustomFieldManagerModal: React.FC<CustomFieldManagerModalProps> = ({
  projectId,
  onClose,
  onFieldsUpdated,
}) => {
  const { theme } = useTheme();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 필드 목록 (좌측 리스트)
  const [fields, setFields] = useState<FieldResponse[]>([]);
  const [selectedField, setSelectedField] = useState<FieldResponse | null>(null);

  // 선택된 필드의 옵션 목록 (우측 상세)
  const [fieldOptions, setFieldOptions] = useState<FieldOptionResponse[]>([]);
  const [isEditingFieldName, setIsEditingFieldName] = useState(false);
  const [newFieldName, setNewFieldName] = useState('');

  // ========================================
  // 1. 데이터 로드 및 갱신
  // ========================================

  const fetchFields = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // 💡 [API] 모든 필드 로드
      const fetchedFields = (await getProjectFields(projectId)).sort(
        (a, b) => a.displayOrder - b.displayOrder,
      );

      setFields(fetchedFields);

      // 로드 후, 선택된 필드가 없다면 첫 번째 필드 자동 선택
      if (!selectedField && fetchedFields.length > 0) {
        setSelectedField(fetchedFields[0]);
      } else if (selectedField) {
        // 기존 선택 필드를 업데이트된 목록에서 다시 찾아서 반영
        const updatedSelected = fetchedFields.find((f) => f.fieldId === selectedField.fieldId);
        setSelectedField(updatedSelected || fetchedFields[0]);
      }
    } catch (err: any) {
      setError(`필드 목록 로드 실패: ${err.message}`);
      setFields(MOCK_FIELDS); // Mock fallback
    } finally {
      setLoading(false);
    }
  }, [projectId, selectedField]);

  const fetchOptions = useCallback(async (fieldId: string) => {
    if (
      !fieldId ||
      fieldId === 'dueDate' ||
      fieldId === 'assignee' ||
      fieldId === 'text' ||
      fieldId === 'number'
    ) {
      setFieldOptions([]); // 옵션이 없는 필드 유형
      return;
    }

    setLoading(true);
    try {
      // 💡 [API] 필드 옵션 로드
      const options = await getFieldOptions(fieldId);
      setFieldOptions(options.sort((a, b) => a.displayOrder - b.displayOrder));
    } catch (err: any) {
      setError(`옵션 로드 실패: ${err.message}`);
      setFieldOptions(MOCK_FIELD_OPTIONS?.filter((o) => o.fieldId === fieldId)); // Mock fallback
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchFields();
  }, [fetchFields]);

  useEffect(() => {
    if (selectedField) {
      setNewFieldName(selectedField.name);
      setIsEditingFieldName(false);
      fetchOptions(selectedField.fieldId);
    }
  }, [selectedField, fetchOptions]);

  // ========================================
  // 2. 필드 수정 핸들러
  // ========================================

  const handleUpdateFieldName = async () => {
    if (!selectedField || !newFieldName.trim()) return;

    setLoading(true);
    try {
      // 💡 [API] 필드 이름 업데이트
      // const updatedField = await updateField(selectedField.fieldId, { name: newFieldName });
      onFieldsUpdated(); // 메인 대시보드 갱신
      await fetchFields(); // 로컬 필드 목록 갱신
      setIsEditingFieldName(false);
    } catch (err: any) {
      setError(`필드 이름 수정 실패: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteField = async (field: FieldResponse) => {
    if (!window.confirm(`정말 필드 '${field.name}'을 삭제하시겠습니까? (복구 불가)`)) return;

    setLoading(true);
    try {
      await deleteField(field.fieldId);
      onFieldsUpdated();
      setSelectedField(null); // 삭제 후 선택 해제
      await fetchFields();
    } catch (err: any) {
      setError(`필드 삭제 실패: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 3. 옵션 수정 핸들러 (Mocking/API)
  // ========================================

  // 💡 TODO: 이 부분은 CustomFieldManageModal 로직을 가져와 옵션 CRUD를 구현해야 합니다.
  // const handleOptionUpdate = async (optionId: string, updates: Partial<FieldOptionResponse>) => {
  //   if (!selectedField) return;

  //   setLoading(true);
  //   try {
  //     // 💡 [API] 옵션 업데이트
  //     await updateFieldOption(optionId, updates);
  //     await fetchOptions(selectedField.fieldId);
  //     onFieldsUpdated();
  //   } catch (err: any) {
  //     setError(`옵션 수정 실패: ${err.message}`);
  //   } finally {
  //     setLoading(false);
  //   }
  // };

  // ========================================
  // 4. 렌더링 헬퍼
  // ========================================

  const renderFieldIcon = (fieldType: string) => {
    switch (fieldType) {
      case 'single_select':
      case 'multi_select':
        return <Tag className="w-4 h-4 text-gray-500" />;
      case 'number':
        return <AlertCircle className="w-4 h-4 text-gray-500" />;
      case 'date':
        return <Calendar className="w-4 h-4 text-gray-500" />;
      case 'single_user':
        return <User className="w-4 h-4 text-gray-500" />;
      default:
        return <CheckSquare className="w-4 h-4 text-gray-500" />;
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[120] p-4">
      <div
        className={`${theme.colors.card} rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col`}
      >
        {/* Header */}
        <div
          className={`p-6 border-b ${theme.colors.border} flex justify-between items-center flex-shrink-0`}
        >
          <h2 className={`text-xl font-bold ${theme.colors.text}`}>커스텀 필드 관리</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700 transition-colors">
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Main Content Area (2 Columns) */}
        <div className="flex flex-1 min-h-0">
          {/* Left Column: Field List */}
          <div
            className={`w-64 border-r ${theme.colors.border} p-4 flex flex-col flex-shrink-0 overflow-y-auto`}
          >
            <h3 className="font-semibold text-sm mb-3">프로젝트 필드</h3>

            {loading && <LoadingSpinner message="필드 로드 중" />}
            {error && <p className="text-xs text-red-500">{error}</p>}

            {fields.map((field) => (
              <div
                key={field.fieldId}
                onClick={() => setSelectedField(field)}
                className={`flex items-center gap-2 p-2 rounded-md cursor-pointer transition ${
                  selectedField?.fieldId === field.fieldId
                    ? 'bg-blue-100 text-blue-700 font-medium'
                    : 'hover:bg-gray-50 text-gray-700'
                }`}
              >
                {renderFieldIcon(field.fieldType)}
                <span className="text-sm truncate">{field.name}</span>
                {field.isSystemDefault && (
                  <span className="text-[10px] bg-gray-200 px-1 rounded">시스템</span>
                )}
              </div>
            ))}

            {/* 💡 [추가] 새 필드 정의 버튼 (CustomFieldManageModal로 이동) */}
            <button
              // ⚠️ 여기서는 CustomFieldManageModal을 열지 않고, MainDashboard로 돌아가도록 유도
              // (CustomFieldManageModal의 진입점은 CreateBoardModal 내부여야 함)
              onClick={() => alert('새 필드 추가는 보드 생성/수정 모달에서 접근해주세요.')}
              className="mt-4 w-full py-2 border-2 border-dashed border-gray-300 rounded-md text-gray-600 hover:border-blue-500 transition-colors flex items-center justify-center gap-2 text-sm"
            >
              <Plus className="w-4 h-4" /> 새 필드 추가
            </button>
          </div>

          {/* Right Column: Field Detail & Options */}
          <div className="flex-1 p-6 overflow-y-auto">
            {selectedField ? (
              <div className="space-y-6">
                {/* Field Header & Editing */}
                <div className="border-b pb-4">
                  <div className="flex justify-between items-start">
                    <h3 className="text-xl font-bold">{selectedField.name}</h3>
                    <button
                      onClick={() => handleDeleteField(selectedField)}
                      className={`p-2 rounded-md hover:bg-red-100 ${
                        selectedField.isSystemDefault ? 'opacity-50 cursor-not-allowed' : ''
                      }`}
                      disabled={selectedField.isSystemDefault}
                    >
                      <Trash2 className="w-5 h-5 text-red-600" />
                    </button>
                  </div>
                  <p className="text-sm text-gray-500">유형: {selectedField.fieldType}</p>
                </div>

                {/* Field Name Editing */}
                <div className="flex items-center gap-3">
                  {isEditingFieldName ? (
                    <input
                      type="text"
                      value={newFieldName}
                      onChange={(e) => setNewFieldName(e.target.value)}
                      onBlur={handleUpdateFieldName}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') handleUpdateFieldName();
                      }}
                      className="px-3 py-1 border rounded-md text-lg font-medium"
                      autoFocus
                    />
                  ) : (
                    <p className="text-lg font-medium">{selectedField.name}</p>
                  )}

                  {!selectedField.isSystemDefault && !isEditingFieldName && (
                    <button
                      onClick={() => setIsEditingFieldName(true)}
                      className="p-1 hover:bg-gray-100 rounded-full"
                    >
                      <Edit2 className="w-4 h-4 text-gray-600" />
                    </button>
                  )}
                </div>

                {/* Options Management (for Select/Multi-Select type) */}
                {(selectedField.fieldType === 'single_select' ||
                  selectedField.fieldType === 'multi_select') && (
                  <div className="space-y-4 pt-4 border-t">
                    <h4 className="font-semibold">옵션 목록 ({fieldOptions.length})</h4>

                    {/* 💡 [TODO] 옵션 추가 인풋 (CustomFieldManageModal에서 가져온 로직 필요) */}
                    <div className="p-3 border rounded-md bg-gray-50">
                      <p className="text-sm text-gray-600">옵션 추가 영역</p>
                    </div>

                    {/* 💡 [TODO] 옵션 리스트 및 수정 (Color, Drag) */}
                    <div className="flex flex-col gap-2 max-h-60 overflow-y-auto">
                      {fieldOptions.map((option) => (
                        <div
                          key={option.optionId}
                          className="flex justify-between items-center p-2 bg-white border rounded-md"
                        >
                          <div className="flex items-center gap-2">
                            <Menu className="w-4 h-4 text-gray-400 cursor-move" />
                            <span
                              className="w-4 h-4 rounded-full"
                              style={{ backgroundColor: option.color }}
                            ></span>
                            <span className="text-sm">{option.label}</span>
                          </div>
                          <div className="flex gap-2">
                            <button
                              onClick={() => alert('색상 변경')}
                              className="text-xs text-blue-500 hover:underline"
                            >
                              색상 변경
                            </button>
                            <button
                              onClick={() => deleteFieldOption(option.optionId)}
                              className="p-1 hover:bg-red-100 rounded-full"
                            >
                              <Trash2 className="w-4 h-4 text-red-600" />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-gray-500">필드를 선택하거나, 새 필드를 추가하세요.</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
