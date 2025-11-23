// src/components/layout/MainLayout.tsx

import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Home, MessageSquare, Bell, File, LogOut, User as UserIcon } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { UserProfileResponse } from '../../types/user';
import { getMyProfile } from '../../api/user/userService'; // 기본 프로필 조회 API

interface MainLayoutProps {
  onLogout: () => void;
  workspaceId: string;
  children: React.ReactNode;
  // 💡 [추가] 프로필 모달을 열기 위한 상위 컴포넌트 핸들러
  onProfileModalOpen: () => void;
}

const MainLayout: React.FC<MainLayoutProps> = ({
  onLogout,
  workspaceId,
  children,
  onProfileModalOpen,
}) => {
  const navigate = useNavigate();
  const { theme } = useTheme();

  // 💡 [수정] DTO 타입 변경 반영 및 userProfile 초기값 설정
  const [userProfile, setUserProfile] = useState<UserProfileResponse | null>(null);

  // UI 상태
  const [showUserMenu, setShowUserMenu] = useState<boolean>(false);
  const [isLoadingProfile, setIsLoadingProfile] = useState<boolean>(true);

  // Ref
  const userMenuRef = useRef<HTMLDivElement>(null);

  const sidebarWidth = 'w-16 sm:w-20';

  // 1. 사용자 기본 프로필 로드 (사이드바 메뉴 및 모달 초기값용)
  useEffect(() => {
    const fetchUserProfile = async () => {
      try {
        // 💡 [수정] 토큰 없이 API 호출 (인터셉터 사용)
        const profile = await getMyProfile();
        console.log(profile);
        setUserProfile(profile);
      } catch (e) {
        console.error('기본 프로필 로드 실패:', e);
      } finally {
        setIsLoadingProfile(false);
      }
    };
    fetchUserProfile();
  }, []);

  // 2. 외부 클릭 감지 (유저 메뉴)
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setShowUserMenu(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  useEffect(() => {
    console.log(showUserMenu);
  }, [showUserMenu]);

  // 3. 워크스페이스 로고 클릭 핸들러
  const handleBackToSelect = () => {
    navigate('/workspaces');
  };

  if (isLoadingProfile) {
    // 레이아웃 로딩 스피너 (프로필이 완전히 로드될 때까지 대기)
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen flex ${theme.colors.background} relative`}>
      {/* 백그라운드 패턴 (레이아웃 일부로 유지) */}
      <div
        className="fixed inset-0 opacity-5"
        style={{
          backgroundImage:
            'linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
      ></div>

      {/* 사이드바 */}
      <aside
        className={`${sidebarWidth} fixed top-0 left-0 h-full flex flex-col justify-between ${theme.colors.primary} text-white shadow-xl z-50 flex-shrink-0`}
      >
        <div className="flex flex-col flex-grow items-center">
          {/* 워크스페이스 로고 */}
          <div className={`py-3 flex justify-center w-full relative`}>
            <button
              onClick={handleBackToSelect}
              title="워크스페이스 목록으로"
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center text-xl font-bold transition 
                    bg-white text-blue-800 ring-2 ring-white/50 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-300`}
            >
              {workspaceId.slice(0, 1).toUpperCase()}
            </button>
          </div>

          {/* 사이드바 메뉴 */}
          <div className="flex flex-col gap-2 mt-4 flex-grow px-2 w-full pt-4">
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center transition bg-blue-600 text-white ring-2 ring-white/50`}
              title="홈"
            >
              <Home className="w-6 h-6" />
            </button>
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="DM"
            >
              <MessageSquare className="w-6 h-6" />
            </button>
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="알림"
            >
              <Bell className="w-6 h-6" />
            </button>
            <button
              className={`w-12 h-12 rounded-lg mx-auto flex items-center justify-center bg-gray-700 hover:bg-gray-600 text-white opacity-50 transition`}
              title="파일"
            >
              <File className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* 하단 유저 메뉴 버튼 */}
        <div className={`py-3 px-2 border-t border-gray-700`}>
          <button
            onClick={(e) => {
              e.stopPropagation(); // 💡 [수정] 이벤트 버블링 차단
              setShowUserMenu(!showUserMenu);
            }}
            className={`w-full flex items-center justify-center py-2 text-sm rounded-lg hover:bg-blue-600 transition relative`}
            title="계정 메뉴"
          >
            <div
              className={`w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center text-sm font-bold ring-2 ring-white/50 text-gray-700 overflow-hidden`}
            >
              {userProfile?.profileImageUrl ? (
                <img
                  src={userProfile.profileImageUrl}
                  alt={userProfile.nickName}
                  className="w-full h-full object-cover"
                />
              ) : (
                userProfile?.nickName[0]?.toUpperCase() || '나'
              )}
            </div>
          </button>
        </div>
      </aside>

      {/* 메인 콘텐츠 영역 */}
      <main
        className="flex-grow flex flex-col relative z-10"
        style={{ marginLeft: sidebarWidth, minHeight: '100vh' }}
      >
        {children}
      </main>

      {/* 유저 메뉴 드롭다운 (사이드바 위에 팝업) */}
      {showUserMenu && (
        <div
          ref={userMenuRef}
          className={`absolute bottom-16 left-12 sm:left-16 w-64 ${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} z-50 ${theme.effects.borderRadius} shadow-2xl`}
          onMouseDown={(e) => e.stopPropagation()} // 💡 [수정] 메뉴 내부 클릭 시 닫히는 현상 방지
        >
          <div className="p-3 pb-3 mb-2 border-b border-gray-200">
            <div className="flex items-center gap-3">
              <div
                className={`w-10 h-10 ${theme.colors.primary} flex items-center justify-center text-white text-base font-bold rounded-md overflow-hidden`}
              >
                {userProfile?.profileImageUrl ? (
                  <img
                    src={userProfile?.profileImageUrl}
                    alt={userProfile?.nickName}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  userProfile?.nickName[0]?.toUpperCase() || 'U'
                )}
              </div>
              <div>
                <h3 className="font-bold text-lg text-gray-900">{userProfile?.nickName}</h3>
                <div className="flex items-center text-green-600 text-xs mt-1">
                  <span className="w-2 h-2 bg-green-500 rounded-full mr-1"></span>
                  대화 가능
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-1 p-2 pt-0">
            <button
              onClick={() => {
                // 💡 [수정] MainDashboard의 Setter를 호출하여 모달을 엽니다.
                onProfileModalOpen();
                setShowUserMenu(false);
              }}
              className="w-full text-left px-2 py-1.5 text-sm text-gray-800 hover:bg-blue-50 hover:text-blue-700 rounded transition flex items-center gap-2"
            >
              <UserIcon className="w-4 h-4" /> 프로필 설정
            </button>
          </div>

          <div className="pt-2 pb-2 border-t border-gray-200 mx-2">
            <button
              onClick={onLogout}
              className="w-full text-left px-2 py-1.5 text-sm text-gray-800 hover:bg-red-50 hover:text-red-700 rounded transition flex items-center gap-2"
            >
              <LogOut className="w-4 h-4" /> 로그아웃
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default MainLayout;
