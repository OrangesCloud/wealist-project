import React, { useState, useEffect, useMemo } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import {
  GroupResponse,
  CreateGroupRequest,
  getGroups, // 💡 GET /api/groups
  createGroup, // 💡 POST /api/groups
  createUserInfo, // 💡 POST /api/userinfo
} from '../api/userService';
import { Search } from 'lucide-react';
import axios, { AxiosError } from 'axios';

interface SelectGroupPageProps {
  userId: string;
  accessToken: string;
  onGroupSelected: (groupId: string) => void;
}

// Mock 데이터 정의는 이제 불필요하거나, 초기 로딩 시 빈 배열로 시작합니다.
// const MOCK_GROUPS: GroupResponse[] = [ ... ];

const SelectGroupPage: React.FC<SelectGroupPageProps> = ({
  userId,
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

  // 1. 그룹 목록 조회 및 초기화 (실제 API 호출)
  useEffect(() => {
    const fetchGroups = async () => {
      if (!accessToken) return;

      setIsLoading(true);
      setError(null);

      try {
        // 🚀 실제 API 호출: 사용자가 속한 모든 활성 그룹을 조회합니다.
        // 현재는 '모든 활성 그룹'을 조회하지만, 백엔드가 '사용자가 속한 그룹'을 필터링해준다고 가정합니다.
        const fetchedGroups = await getGroups(accessToken);
        setGroups(fetchedGroups);

        // 🚨 참고: Swagger 스펙상 'getGroups'는 MessageApiResponse<any>를 반환합니다.
        //         'userService.ts'에서 data 배열을 추출하는 로직이 필요합니다.
      } catch (e) {
        const err = e as AxiosError;
        setError(`조직 목록 조회 실패: ${err.message}`);
        setGroups([]); // 실패 시 빈 배열로 설정
      } finally {
        setIsLoading(false);
      }
    };

    fetchGroups();
  }, [accessToken]);

  // 2. 조직 검색 필터링 로직 (기존 로직 유지)
  const availableGroups = useMemo(() => {
    if (!groups) return [];
    const query = searchQuery.toLowerCase().trim();

    if (!query) {
      return groups;
    }

    // 이름, 회사 이름으로 필터링합니다.
    return groups.filter(
      (group) =>
        group.name.toLowerCase().includes(query) || group.companyName.toLowerCase().includes(query),
    );
  }, [searchQuery, groups]);

  // 3. 새로운 그룹 생성 및 등록 핸들러 (실제 API 호출)
  const handleCreateAndSelectGroup = async () => {
    if (!newGroupName.trim()) {
      setError('그룹 이름을 입력해 주세요.');
      return;
    }

    setIsLoading(true);
    setError(null);

    const createData: CreateGroupRequest = {
      name: newGroupName,
      companyName: newCompany || 'Personal',
    };

    try {
      // 🚀 1단계: 새 그룹 생성 (POST /api/groups)
      const newGroup = await createGroup(createData, accessToken);
      const newGroupId = newGroup.groupId; // 생성된 그룹의 ID 추출

      // 🚀 2단계: 생성된 그룹에 사용자 정보 등록 (POST /api/userinfo)
      // Spring Security는 JWT의 정보를 사용하여 UserInfo를 자동으로 생성할 수도 있지만,
      // Swagger에 명시된 createUserInfo를 사용하여 명시적으로 등록합니다.
      await createUserInfo(userId, newGroupId, accessToken, 'LEADER');

      alert(`조직 '${newGroupName}' 생성 및 등록 완료!`);

      // 🚀 최종 핸들러 호출 -> Workspace 생성 단계로 이동
      onGroupSelected(newGroupId);
    } catch (e) {
      const err = e as AxiosError;
      setError(`조직 생성 및 등록 실패: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  // 4. 기존 그룹 선택 핸들러 (실제 API 호출)
  const handleSelectExistingGroup = async (group: GroupResponse) => {
    setIsLoading(true);
    setError(null);

    try {
      // 🚀 1단계: 기존 그룹에 사용자 정보 등록/업데이트 (POST /api/userinfo)
      // 이미 등록된 사용자라면 등록 API가 업데이트 역할을 할 수도 있습니다.
      // 현재는 '선택'이 곧 '등록'을 의미한다고 가정합니다.
      await createUserInfo(userId, group.groupId, accessToken, 'MEMBER');

      alert(`그룹 '${group.name}'에 참여 완료!`);

      // 🚀 최종 핸들러 호출 -> Workspace 생성 단계로 이동
      onGroupSelected(group.groupId);
    } catch (e) {
      const err = e as AxiosError;
      setError(`그룹 참여 등록 실패: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  // --- 로딩 화면 ---
  if (isLoading || groups === null) {
    return (
      <div
        className={`min-h-screen ${theme.colors.background} flex items-center justify-center p-4`}
      >
        <div className="p-8">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
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
              {isLoading ? '생성 및 등록 중...' : '새 조직 생성 및 시작'}
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
