import React, { useState } from 'react';
import { X, Edit, Trash2, Users, FileText, Upload, Plus, Search } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { Column } from '../../types';
// 💡 Column 타입을 MainDashboard와 동일하게 사용

// 💡 담당자 Mock 타입 (실제 User Service에서 올 정보)
interface ColumnMember {
  id: string;
  name: string;
  role: string;
}

interface ColumnDetailModalProps {
  column: Column;
  onClose: () => void;
  onUpdate: (updatedColumn: Column) => void;
}

// 🚧 Mock 데이터 (전체 조직 구성원 목록)
const ALL_MOCK_MEMBERS: ColumnMember[] = [
  { id: 'user-1', name: '김개발', role: 'Owner' },
  { id: 'user-2', name: '박보안', role: 'Contributor' },
  { id: 'user-3', name: '이디자인', role: 'Viewer' },
  { id: 'user-4', name: '최데브옵스', role: 'Contributor' },
  { id: 'user-5', name: '정기획', role: 'Viewer' },
  { id: 'user-6', name: '하인턴', role: 'Viewer' },
];

const ColumnDetailModal: React.FC<ColumnDetailModalProps> = ({ column, onClose, onUpdate }) => {
  const { theme } = useTheme();

  // 💡 상태 초기화
  const [title, setTitle] = useState(column.title);
  const [initialTitle] = useState(column.title);

  const [description, setDescription] = useState(
    `[${column.title} 목표] 이 컬럼은 작업 시작 전의 '대기열' 역할을 담당합니다. 이곳에 있는 태스크는 백로그 및 다음 스프린트의 후보입니다.\n\n[규칙] 담당자(Assignee)가 지정된 태스크만 이 컬럼에 추가될 수 있습니다.`,
  );
  const [initialDescription] = useState(description);

  const [isEditing, setIsEditing] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  // 💡 수정된 상태: 현재 멤버와 검색 상태
  const [members, setMembers] = useState<ColumnMember[]>(ALL_MOCK_MEMBERS.slice(0, 3)); // 초기 멤버 (Mock)
  const [searchQuery, setSearchQuery] = useState('');

  // --- Mock 검색 로직 ---
  const filteredAvailableMembers = ALL_MOCK_MEMBERS.filter(
    (member) =>
      // 1. 검색어 필터링
      member.name.toLowerCase().includes(searchQuery.toLowerCase()) &&
      // 2. 이미 추가된 멤버 제외
      !members.some((m) => m.id === member.id),
  );

  // 수정 처리 핸들러 (Mock)
  const handleSave = () => {
    if (!title.trim()) return;

    setIsLoading(true);

    setTimeout(() => {
      const updatedColumn: Column = {
        ...column,
        title: title.trim(),
      };

      onUpdate(updatedColumn);
      alert(`[Mock] '${updatedColumn.title}' 수정 요청 완료!`);
      setIsLoading(false);
      setIsEditing(false); // 저장 후 편집 모드 해제
    }, 500);
  };

  // 삭제 처리 핸들러 (Mock)
  const handleDelete = () => {
    if (
      window.confirm(
        `정말로 "${column.title}"을(를) 삭제하시겠습니까? (이 컬럼의 ${column.tasks.length}개 태스크는 다음 컬럼으로 이동됩니다.)`,
      )
    ) {
      alert(`[Mock] '${column.title}' 삭제 처리 완료.`);
      onClose();
    }
  };

  // 편집 모드 토글 핸들러
  const handleEditToggle = () => {
    if (isEditing) {
      if (window.confirm('편집을 취소하시겠습니까? 변경 사항은 저장되지 않습니다.')) {
        setTitle(initialTitle);
        setDescription(initialDescription);
        setIsEditing(false);
      }
    } else {
      setIsEditing(true);
    }
  };

  // 담당자 추가/제거 핸들러 (Mock)
  const handleAddMember = (member: ColumnMember) => {
    setMembers((prev) => [...prev, member]);
    setSearchQuery(''); // 추가 후 검색창 초기화
  };

  const handleRemoveMember = (memberId: string) => {
    if (window.confirm('담당자를 목록에서 제외하시겠습니까?')) {
      setMembers((prev) => prev.filter((m) => m.id !== memberId));
    }
  };

  // 저장 버튼 활성화 조건
  const isSaveDisabled =
    !isEditing ||
    isLoading ||
    !title.trim() ||
    (title === initialTitle && description === initialDescription);

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50 overflow-y-auto"
      onClick={onClose}
    >
      <div className="relative w-full max-w-4xl my-8" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} p-6 sm:p-8 max-h-[90vh] overflow-y-auto ${theme.effects.borderRadius} shadow-xl`}
        >
          {/* 모달 헤더 및 제목 */}
          <div className="flex items-start justify-between mb-4">
            <div className="flex-1 pr-4">
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className={`w-full ${theme.font.size.xl} font-extrabold ${
                  theme.colors.text
                } focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  isEditing ? 'border-b-2 border-blue-200' : 'bg-transparent'
                }`}
                disabled={!isEditing || isLoading}
              />
            </div>
            <button
              onClick={onClose}
              className={`bg-red-500 ${theme.effects.cardBorderWidth} ${theme.colors.border} p-2 hover:bg-red-600 flex-shrink-0 ${theme.effects.borderRadius} transition text-white`}
              disabled={isLoading}
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <div className="border-b border-gray-200 mb-6 pb-4 flex justify-end">
            <button
              onClick={handleEditToggle}
              className={`flex items-center gap-2 ${
                isEditing ? 'bg-gray-500 hover:bg-gray-600' : 'bg-blue-600 hover:bg-blue-700'
              } text-white px-3 py-2 ${theme.effects.borderRadius} transition ${
                theme.font.size.xs
              }`}
              disabled={isLoading}
            >
              {isEditing ? <X className="w-4 h-4" /> : <Edit className="w-4 h-4" />}
              {isEditing ? '편집 취소' : '편집 모드'}
            </button>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            {/* 1. 메인 콘텐츠 (목표/설명) */}
            <div className="md:col-span-2 space-y-4">
              <h3 className={`${theme.font.size.base} font-bold flex items-center gap-2 mb-2`}>
                <FileText className="w-5 h-5 text-gray-700" />
                목표 및 비전
              </h3>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="이 컬럼의 목표, 규칙, 책임 영역 등을 상세히 기재하세요."
                className={`w-full px-4 py-3 border ${theme.font.size.sm} min-h-[250px] ${
                  theme.effects.borderRadius
                } resize-y focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  !isEditing ? 'bg-gray-50 cursor-default' : 'bg-white'
                }`}
                disabled={!isEditing || isLoading}
              />
            </div>

            {/* 2. 사이드바 (담당자 및 첨부파일) */}
            <div className="md:col-span-1 space-y-6">
              {/* 담당자 목록 */}
              <div>
                <h3
                  className={`${theme.font.size.sm} font-bold flex items-center gap-2 mb-3 ${theme.colors.text}`}
                >
                  <Users className="w-4 h-4" />
                  담당자 ({members.length})
                </h3>
                <div className="space-y-2 max-h-48 overflow-y-auto border p-3 rounded-lg bg-gray-50">
                  {members.map((member) => (
                    <div key={member.id} className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div
                          className={`w-6 h-6 bg-blue-100 text-blue-800 ${theme.effects.borderRadius} flex items-center justify-center font-semibold text-xs`}
                        >
                          {member.name[0]}
                        </div>
                        <span>{member.name}</span>
                      </div>
                      {isEditing && (
                        <button
                          onClick={() => handleRemoveMember(member.id)}
                          className="text-red-500 hover:text-red-700 transition"
                          title="담당자 제외"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      )}
                      {!isEditing && (
                        <span
                          className={`text-xs px-2 py-1 rounded-full ${
                            member.role === 'Owner' ? 'bg-yellow-200' : 'bg-gray-200'
                          }`}
                        >
                          {member.role}
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              {/* 💡 담당자 검색 및 추가 필드 (편집 모드 시 활성화) */}
              {isEditing && (
                <div className="border p-3 rounded-lg bg-white shadow-sm">
                  <h4 className={`${theme.font.size.xs} font-bold mb-2`}>조직원 검색/추가</h4>
                  <div className="relative mb-2">
                    <input
                      type="text"
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      placeholder="이름으로 검색..."
                      className={`w-full px-3 py-2 pl-8 border ${theme.font.size.xs} rounded-md focus:border-blue-500`}
                    />
                    <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 w-3 h-3 text-gray-400" />
                  </div>

                  {/* 검색 결과 목록 */}
                  <div className="space-y-1 max-h-24 overflow-y-auto">
                    {filteredAvailableMembers.length > 0 ? (
                      filteredAvailableMembers.map((member) => (
                        <button
                          key={member.id}
                          onClick={() => handleAddMember(member)}
                          className={`w-full flex justify-between items-center p-1.5 text-left bg-gray-50 hover:bg-gray-100 rounded transition ${theme.font.size.xs}`}
                        >
                          <span>{member.name}</span>
                          <Plus className="w-3 h-3 text-green-600" />
                        </button>
                      ))
                    ) : (
                      <p
                        className={`text-center py-1 ${theme.font.size.xs} ${theme.colors.subText}`}
                      >
                        {searchQuery ? '결과 없음' : '검색어를 입력하세요.'}
                      </p>
                    )}
                  </div>
                </div>
              )}

              {/* 첨부 파일 (Mock) */}
              <div>
                <h3
                  className={`${theme.font.size.sm} font-bold flex items-center gap-2 mb-3 ${theme.colors.text}`}
                >
                  <Upload className="w-4 h-4" />
                  관련 파일 첨부
                </h3>
                <div className="space-y-2 border p-3 rounded-lg bg-gray-50">
                  <p className={`text-xs ${theme.colors.subText}`}>
                    파일 첨부 기능은 구현 예정입니다. (Mock)
                  </p>
                  {isEditing && (
                    <input
                      type="file"
                      multiple
                      className={`w-full text-sm ${theme.font.size.xs}`}
                      disabled={isLoading}
                    />
                  )}
                  <p className={`text-xs text-gray-700 font-medium`}>- 프로젝트 명세서.pdf</p>
                </div>
              </div>
            </div>
          </div>

          {/* --- 하단 액션 버튼 --- */}
          <div className="flex justify-between gap-3 pt-4 mt-6 border-t border-gray-200">
            <button
              onClick={handleSave}
              disabled={isSaveDisabled}
              className={`flex-1 ${theme.colors.primary} text-white py-3 font-bold ${theme.colors.primaryHover} transition ${theme.font.size.sm} ${theme.effects.borderRadius} disabled:opacity-50`}
            >
              {isLoading ? '저장 중...' : '수정 및 저장'}
            </button>
            <button
              onClick={handleDelete}
              className={`bg-red-500 text-white px-4 py-3 font-bold hover:bg-red-600 transition ${theme.font.size.sm} ${theme.effects.borderRadius} flex items-center gap-2`}
              disabled={isLoading || isEditing}
            >
              <Trash2 className="w-4 h-4" />
              삭제
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ColumnDetailModal;
