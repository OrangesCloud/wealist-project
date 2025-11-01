import React, { useState, useEffect, useMemo } from 'react';
import { useTheme } from '../contexts/ThemeContext';
// 💡 실제 API 대신 Mock 함수를 사용합니다.
import { GroupResponse } from '../api/userService';
import { Search } from 'lucide-react';

interface SelectGroupPageProps {
  userId: string;
  accessToken: string;
  onGroupSelected: (groupId: string) => void;
}

// 🚧 Mock 데이터 정의 (조직 검색을 위한 더미 데이터)
const MOCK_GROUPS: GroupResponse[] = [
  { groupId: '1111-a', name: 'Wealist Dev Team (Mock)', companyName: 'Wealist Inc.' },
  { groupId: '2222-b', name: 'Orange Cloud Design (Mock)', companyName: 'KT Cloud' },
  { groupId: '3333-c', name: 'Project Kanban Alpha (Mock)', companyName: 'Self-Employed' },
  { groupId: '4444-d', name: 'Data Engineer Study (Mock)', companyName: 'Personal' },
];

const SelectGroupPage: React.FC<SelectGroupPageProps> = ({
  // userId,
  accessToken,
  onGroupSelected,
}) => {
  const { theme } = useTheme();

  const [groups, setGroups] = useState<GroupResponse[] | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [isCreatingNewGroup, setIsCreatingNewGroup] = useState(false);
  const [newGroupName, setNewGroupName] = useState('');
  const [newCompany, setNewCompany] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  // 1. 그룹 목록 조회 및 초기화 (MOCK 처리)
  useEffect(() => {
    const mockFetchGroups = () => {
      setIsLoading(true);
      setError(null);

      setTimeout(() => {
        // 🚧 [Mock] 미리 정의된 조직 목록을 반환합니다. (사용자가 속한 그룹이 있다면 목록에 나타납니다.)
        // 현재는 '처음 접속한 사용자' 시나리오에 맞게 빈 목록을 반환하는 대신,
        // 선택할 수 있는 조직 목록을 Mock으로 제공합니다.
        setGroups(MOCK_GROUPS);
        setIsLoading(false);
      }, 500);
    };

    mockFetchGroups();
  }, [accessToken]);

  // 2. 조직 검색 필터링 로직 (useMemo로 성능 최적화)
  const availableGroups = useMemo(() => {
    if (!groups) return [];
    const query = searchQuery.toLowerCase().trim();

    // 💡 변경된 로직: 검색어가 없으면 (false) groups 배열 전체를 반환합니다.
    if (!query) {
      return groups;
    }

    // 이름, 회사 이름으로 필터링합니다.
    return groups.filter(
      (group) =>
        group.name.toLowerCase().includes(query) || group.companyName.toLowerCase().includes(query),
    );
  }, [searchQuery, groups]);

  // 3. 새로운 그룹 생성 및 등록 핸들러 (MOCK)
  const handleCreateAndSelectGroup = async () => {
    if (!newGroupName.trim()) {
      setError('그룹 이름을 입력해 주세요.');
      return;
    }

    setIsLoading(true);
    setError(null);

    setTimeout(() => {
      const newGroupId = 'mock-new-group-' + Math.random().toString(36).substring(2, 9);
      alert(`[Mock] 조직 '${newGroupName}' 생성 완료!`);
      setIsLoading(false);
      onGroupSelected(newGroupId);
    }, 1500);
  };

  // 4. 기존 그룹 선택 핸들러 (MOCK)
  const handleSelectExistingGroup = async (group: GroupResponse) => {
    setIsLoading(true);
    setError(null);
    setTimeout(() => {
      setIsLoading(false);
      alert(`[Mock] 그룹 '${group.name}' 선택 완료!`);
      // 🚀 최종 핸들러 호출 -> Workspace 생성 단계로 이동
      onGroupSelected(group.groupId);
    }, 500);
  };

  // --- 로딩 화면 ---
  if (isLoading || groups === null) {
    return (
      <div
        className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4`}
      >
        <div className="p-8">
          <p className={`${theme.font.size.lg} ${theme.colors.text}`}>조직 정보를 확인 중...</p>
        </div>
      </div>
    );
  }

  // --- 메인 렌더링 ---
  return (
    <div className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4`}>
      <div
        className={`${theme.colors.card} ${theme.effects.borderRadius} p-6 sm:p-8 w-full max-w-lg relative z-10 shadow-xl ${theme.effects.cardBorderWidth} ${theme.colors.border}`}
      >
        <h2
          className={`${theme.font.size.xl} font-extrabold ${theme.colors.text} mb-2 text-center`}
        >
          {isCreatingNewGroup ? '새로운 조직 만들기 🏗️' : '워크스페이스 조직 선택'}
        </h2>

        <p className={`text-center mb-6 ${theme.font.size.sm} ${theme.colors.subText}`}>
          <span className={`${theme.colors.text} font-bold mr-1`}>소속된 조직에 참여하거나,</span>새
          조직을 생성하여 시작해 보세요.
        </p>

        {error && (
          <p
            className={`${theme.colors.danger} text-center mb-4 ${theme.font.size.sm} border border-red-300 p-2 rounded-md bg-red-50`}
          >
            {error}
          </p>
        )}

        {isCreatingNewGroup ? (
          /* ------------------- 조직 생성 폼 ------------------- */
          <div className="space-y-4">
            <input
              type="text"
              placeholder="그룹 이름 (예: Orange Cloud 개발팀)"
              value={newGroupName}
              onChange={(e) => setNewGroupName(e.target.value)}
              className={`w-full px-4 py-3 ${theme.colors.secondary} ${theme.font.size.sm} rounded-lg border-2 border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition`}
              disabled={isLoading}
            />
            <input
              type="text"
              placeholder="회사 이름 (선택 사항)"
              value={newCompany}
              onChange={(e) => setNewCompany(e.target.value)}
              className={`w-full px-4 py-3 ${theme.colors.secondary} ${theme.font.size.sm} rounded-lg border-2 border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition`}
              disabled={isLoading}
            />
            <button
              onClick={handleCreateAndSelectGroup}
              disabled={isLoading || !newGroupName.trim()}
              className={`w-full ${theme.colors.success} text-white py-3 font-bold rounded-lg ${theme.colors.successHover} transition disabled:opacity-50 disabled:cursor-not-allowed shadow-md`}
            >
              {isLoading ? '생성 및 등록 중...' : '새 조직 생성 및 시작 (Mock)'}
            </button>

            <button
              onClick={() => setIsCreatingNewGroup(false)}
              className={`w-full ${theme.colors.info} py-2 mt-2 hover:text-blue-700 underline ${theme.font.size.sm}`}
              disabled={isLoading}
            >
              &larr; 돌아가서 기존 조직 검색하기
            </button>
          </div>
        ) : (
          /* ------------------- 조직 검색/선택 UI ------------------- */
          <div className="space-y-4">
            {/* 1. 검색 입력 필드 */}
            <div className="relative">
              <input
                type="text"
                placeholder="조직 이름 또는 코드로 검색"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className={`w-full px-4 pl-10 py-3 ${theme.colors.secondary} ${theme.font.size.sm} rounded-lg border-2 border-gray-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition`}
                disabled={isLoading}
              />
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            </div>

            {/* 2. 조직 목록 표시 영역 */}
            <div className={`max-h-60 overflow-y-auto border-2 ${theme.colors.border} rounded-lg`}>
              {availableGroups.length > 0 ? (
                availableGroups.map((group) => (
                  <button
                    key={group.groupId}
                    onClick={() => handleSelectExistingGroup(group)}
                    className={`w-full text-left p-3 hover:bg-blue-50 border-b border-gray-100 ${theme.colors.text} ${theme.font.size.sm} transition flex justify-between items-center last:border-b-0`}
                    disabled={isLoading}
                  >
                    <div>
                      <span className="font-semibold">{group.name}</span>
                      <p className={`${theme.colors.subText} ${theme.font.size.xs}`}>
                        {group.companyName}
                      </p>
                    </div>
                    <span
                      className={`${theme.colors.info} ${theme.font.size.xs} px-2 py-1 border border-blue-200 rounded`}
                    >
                      선택
                    </span>
                  </button>
                ))
              ) : (
                <p className={`p-4 text-center ${theme.colors.subText} ${theme.font.size.sm}`}>
                  {searchQuery.trim()
                    ? '검색 결과가 없습니다. 이름을 확인하거나 새로 생성해 보세요.'
                    : '소속된 조직이 없습니다. 아래 버튼으로 새로 생성하거나, 이름을 검색하세요.'}
                </p>
              )}
            </div>

            {/* 3. + 새 조직 생성하기 버튼 (강조) */}
            <div className="mt-6 pt-4 border-t border-gray-100">
              <button
                onClick={() => setIsCreatingNewGroup(true)}
                className={`w-full ${theme.colors.primary} text-white py-3 font-bold rounded-lg ${theme.colors.primaryHover} transition disabled:opacity-50 shadow-lg`}
                disabled={isLoading}
              >
                <span className="text-xl mr-2">+</span> 새 조직 생성하기
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default SelectGroupPage;
