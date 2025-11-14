/**
 * 사용자 프로필 모달 컴포넌트
 *
 * [최종 로직 목표]
 * 1. 초기 로드 시: GET /api/workspaces/all (워크스페이스 목록) + GET /api/profiles/all/me (모든 프로필)을 호출.
 * 2. 탭 선택 시: 로컬 상태(allProfiles)에서 기본 프로필(workspaceId=null)과 선택된 워크스페이스 프로필을 필터링하여 표시.
 * 3. 저장 시: S3에 이미지를 업로드하고 반환된 URL로 닉네임과 프로필을 업데이트합니다.
 */

import React, { useState, useRef, ChangeEvent, useEffect } from 'react';
import { X, Camera } from 'lucide-react';
import { useTheme } from '../../../contexts/ThemeContext';
import { updateMyProfile, getAllMyProfiles, getMyWorkspaces } from '../../../api/user/userService';
import {
  UserProfileResponse,
  WorkspaceResponse,
  UpdateProfileRequest,
  UserWorkspaceResponse,
} from '../../../types/user';

// 💡 [추가] S3 업로드 헬퍼 함수

interface UserProfileModalProps {
  onClose: () => void;
}

const UserProfileModal: React.FC<UserProfileModalProps> = ({ onClose }) => {
  const { theme } = useTheme();
  const [activeTab, setActiveTab] = useState<'default' | 'workspace'>('default');

  const [allProfiles, setAllProfiles] = useState<UserProfileResponse[]>([]);
  const [workspaces, setWorkspaces] = useState<UserWorkspaceResponse[]>([]);
  const [selectedWorkspaceId, setSelectedWorkspaceId] = useState<string>('');

  const fileInputRef = useRef<HTMLInputElement>(null);

  const [defaultNickName, setDefaultNickName] = useState('');
  const [workspaceNickName, setWorkspaceNickName] = useState('');

  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState<string | null>(null);

  // 💡 [추가] S3에 업로드할 실제 파일 객체 상태
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // ========================================
  // 프로필 데이터 필터링 및 계산된 상태
  // ========================================

  const defaultProfile = allProfiles?.find((p) => p.workspaceId === null) || null;
  const currentWorkspaceProfile =
    allProfiles?.find((p) => p.workspaceId === selectedWorkspaceId) || null;

  const currentProfile =
    activeTab === 'default' ? defaultProfile : currentWorkspaceProfile || defaultProfile;

  const currentNickName = activeTab === 'default' ? defaultNickName : workspaceNickName;
  const setCurrentNickName = activeTab === 'default' ? setDefaultNickName : setWorkspaceNickName;

  // ========================================
  // 초기 데이터 로드 (유지)
  // ========================================

  useEffect(() => {
    const loadInitialData = async () => {
      try {
        setLoading(true);
        const [allProfs, workspaceList] = await Promise.all([
          getAllMyProfiles(),
          getMyWorkspaces(),
        ]);
        console.log(allProfs);
        setAllProfiles(allProfs);
        const initialDefaultProfile = allProfs?.find((p) => p.workspaceId === null);
        if (initialDefaultProfile) {
          setDefaultNickName(initialDefaultProfile?.nickName);
        }

        setWorkspaces(workspaceList);
        if (workspaceList.length > 0) {
          setSelectedWorkspaceId(workspaceList[0].workspaceId);
        }
      } catch (err) {
        console.error('[Initial Data Load Error]', err);
        setError('프로필 정보를 불러오는데 실패했습니다.');
      } finally {
        setLoading(false);
      }
    };
    loadInitialData();
  }, []);

  // 💡 [추가] 워크스페이스 변경 시 닉네임 입력 필드 상태 업데이트 (유지)
  useEffect(() => {
    if (activeTab === 'workspace') {
      const workspace = workspaces?.find((ws) => ws.workspaceId === selectedWorkspaceId);

      if (currentWorkspaceProfile) {
        setWorkspaceNickName(currentWorkspaceProfile.nickName);
      } else if (defaultProfile) {
        setWorkspaceNickName(
          `${defaultProfile.nickName} (${workspace?.workspaceName || '새 조직'})`,
        );
      } else {
        setWorkspaceNickName('');
      }
    }
  }, [selectedWorkspaceId, activeTab, currentWorkspaceProfile, defaultProfile, workspaces]);

  // ========================================
  // 이미지 업로드 핸들러 (S3 파일 상태 추가)
  // ========================================

  const handleAvatarChangeClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl);
      }
      setAvatarPreviewUrl(URL.createObjectURL(file));
      // 💡 [추가] 업로드할 파일 객체를 상태에 저장
      setSelectedFile(file);
      console.log(`[File] 새 프로필 사진 선택: ${file.name}`);
    } else {
      // 파일 선택 취소 시 초기화
      setSelectedFile(null);
      setAvatarPreviewUrl(null);
    }
  };

  // ========================================
  // 워크스페이스 변경 핸들러 (변경 없음)
  // ========================================

  const handleWorkspaceChange = (workspaceId: string) => {
    setSelectedWorkspaceId(workspaceId);
    setAvatarPreviewUrl(null); // 워크스페이스 변경 시 미리보기 초기화
  };

  // ========================================
  // 저장 핸들러 (S3 업로드 로직 포함)
  // ========================================

  const handleSave = async () => {
    try {
      setLoading(true);
      setError(null);

      if (!currentNickName.trim()) {
        setError('닉네임은 필수입니다.');
        setLoading(false);
        return;
      }

      const currentUserId = defaultProfile?.userId;
      if (!currentUserId) {
        throw new Error('사용자 ID를 찾을 수 없습니다. (재로그인 필요)');
      }

      let newImageUrl: string | undefined = undefined;

      // 1. S3 이미지 업로드 필요 시 처리
      // if (selectedFile) {
      //   newImageUrl = await uploadProfileImage(selectedFile, currentUserId);
      // } else {
      //   // 2. 파일 변경이 없다면 기존 URL 유지 (null 또는 undefined 포함)
      //   newImageUrl = currentProfile?.profileImageUrl || undefined;
      // }

      // // 3. API 호출 DTO 구성
      // const data: UpdateProfileRequest = {
      //   nickName: currentNickName,
      //   profileImageUrl: newImageUrl, // S3에서 받은 URL 또는 기존 URL
      // };

      // let updatedProfile: UserProfileResponse;

      // if (activeTab === 'default') {
      //   // PUT /api/profiles/me
      //   updatedProfile = await updateMyProfile(data);
      //   alert('기본 프로필이 저장되었습니다.');
      // } else {
      //   // PUT /api/profiles/workspace/{workspaceId} (Mock 처리)
      //   // ⚠️ [주의] updateWorkspaceProfile은 Mock 함수이거나 백엔드 구현이 필요합니다.
      //   // updatedProfile = await updateWorkspaceProfile(selectedWorkspaceId, data);
      //   const workspaceName_display = workspaces?.find(
      //     (ws) => ws.workspaceId === selectedWorkspaceId,
      //   )?.workspaceName;
      //   alert(`${workspaceName_display} 프로필이 저장되었습니다. (⚠️ 백엔드 구현 확인 필요)`);
      // }

      // // 4. 로컬 상태 업데이트 (모든 프로필)
      // setAllProfiles((prev) => {
      //   const targetId = activeTab === 'default' ? null : selectedWorkspaceId;
      //   const index = prev?.findIndex((p) => p.workspaceId === targetId);

      //   if (index !== -1 && prev) {
      //     const newProfiles = [...prev];
      //     newProfiles[index] = updatedProfile;
      //     return newProfiles;
      //   }
      //   return [...(prev || []), updatedProfile];
      // });

      // 5. 저장 후 파일 상태 초기화
      setSelectedFile(null);
      setAvatarPreviewUrl(null);
    } catch (err: any) {
      const errorMsg = err.response?.data?.error?.message || err.message;
      console.error('[Profile Save Error]', errorMsg);
      setError('프로필 저장에 실패했습니다. (S3 업로드 또는 API 문제)');
      // 에러 시 파일 상태는 유지하여 사용자가 다시 시도하거나 취소할 수 있도록 함.
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 모달 닫기 핸들러 (변경 없음)
  // ========================================

  const handleClose = () => {
    if (avatarPreviewUrl) {
      URL.revokeObjectURL(avatarPreviewUrl);
    }
    onClose();
  };

  // ========================================
  // 렌더링
  // ========================================

  if (!defaultProfile && loading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-700">프로필 정보를 불러오는 중...</p>
        </div>
      </div>
    );
  }

  if (!defaultProfile && !loading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <p className="text-red-700 font-semibold mb-4">프로필 로드 실패</p>
          <p className="text-sm text-gray-700">기본 프로필 정보를 찾을 수 없습니다.</p>
          <button
            onClick={handleClose}
            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            닫기
          </button>
        </div>
      </div>
    );
  }

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      onClick={handleClose}
    >
      <div className="relative w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} ${theme.effects.borderRadius} shadow-xl`}
        >
          {/* 헤더 */}
          <div className="flex items-center justify-between p-4 pb-3">
            <h2 className={`${theme.font.size.base} font-bold text-gray-800`}>
              사용자 프로필 설정
            </h2>
            <button
              onClick={handleClose}
              className="p-2 hover:bg-gray-100 rounded-lg transition"
              title="닫기"
            >
              <X className="w-4 h-4 text-gray-600" />
            </button>
          </div>

          {/* 탭 메뉴 */}
          <div className="flex border-b border-gray-200 px-6">
            <button
              onClick={() => setActiveTab('default')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'default' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              기본 프로필
              {activeTab === 'default' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
            <button
              onClick={() => setActiveTab('workspace')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'workspace' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              워크스페이스별 프로필
              {activeTab === 'workspace' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
          </div>

          {/* 탭 컨텐츠 */}
          <div className="p-6 space-y-5">
            {/* 에러 메시지 */}
            {error && (
              <div className="p-3 bg-red-100 border border-red-400 text-red-700 rounded-md text-sm">
                {error}
              </div>
            )}

            {/* 워크스페이스 선택 */}
            <div className={activeTab === 'default' ? 'hidden' : ''}>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                워크스페이스 선택:
              </label>
              <select
                value={selectedWorkspaceId}
                onChange={(e) => handleWorkspaceChange(e.target.value)}
                className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
              >
                {workspaces.map((workspace) => (
                  <option key={workspace.workspaceId} value={workspace.workspaceId}>
                    {workspace.workspaceName}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-500">
                워크스페이스마다 다른 프로필을 설정할 수 있습니다
              </p>
            </div>
            {/* 기본 탭일 때 높이 유지를 위한 공간 */}
            {activeTab === 'default' && <div style={{ height: '70px' }} className="w-full"></div>}

            {/* 프로필 이미지 */}
            <div className="flex flex-col items-center mb-4">
              <div className="relative">
                {avatarPreviewUrl ? (
                  <img
                    src={avatarPreviewUrl}
                    alt="프로필 미리보기"
                    className="w-24 h-24 object-cover border-2 border-gray-300 rounded-full"
                  />
                ) : currentProfile?.profileImageUrl ? (
                  <img
                    src={currentProfile.profileImageUrl}
                    alt="프로필 이미지"
                    className="w-24 h-24 object-cover border-2 border-gray-300 rounded-full"
                  />
                ) : (
                  <div className="w-24 h-24 bg-blue-500 border-2 border-gray-300 flex items-center justify-center text-white text-3xl font-bold rounded-full">
                    {currentNickName[0] || 'U'}
                  </div>
                )}

                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileChange}
                  accept="image/*"
                  className="hidden"
                />

                <button
                  onClick={handleAvatarChangeClick}
                  className="absolute bottom-0 right-0 p-2 bg-gray-700 hover:bg-gray-800 text-white rounded-full transition shadow-md"
                  title="프로필 사진 변경"
                >
                  <Camera className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* 닉네임 */}
            <div>
              <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                닉네임:
              </label>
              <input
                type="text"
                value={currentNickName}
                onChange={(e) => setCurrentNickName(e.target.value)}
                className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                placeholder="닉네임을 입력하세요"
              />
            </div>

            {/* 버튼 영역 */}
            <div className="flex gap-2 pt-4">
              <button
                onClick={handleSave}
                disabled={loading}
                className={`flex-1 ${theme.colors.primary} text-white py-3 ${
                  theme.effects.borderRadius
                } font-semibold transition ${
                  loading ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-90'
                }`}
              >
                {loading ? '저장 중...' : '저장'}
              </button>
              <button
                onClick={handleClose}
                className="flex-1 bg-gray-300 text-gray-800 py-3 rounded-lg font-semibold hover:bg-gray-400 transition"
              >
                취소
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UserProfileModal;
