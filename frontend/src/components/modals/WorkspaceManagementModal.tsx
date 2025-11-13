/**
 * 워크스페이스 관리 모달 컴포넌트 (OWNER/ADMIN용)
 *
 * [백엔드 개발자 참고사항]
 *
 * 모든 API 로직은 src/api/user/userService.ts에 구현되어 있습니다.
 * userService.ts 파일에서 USE_MOCK_DATA 플래그를 false로 변경하면
 * 자동으로 실제 API를 호출합니다.
 */

import React, { useState, useEffect, useMemo } from 'react';
import { X, Search } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { useAuth } from '../../contexts/AuthContext';

// 💡 DTO 및 함수 Import 업데이트: 토큰 인자 제거 및 DTO 이름 변경 반영
import {
  getWorkspaceSettings,
  updateWorkspaceSettings,
  getWorkspaceMembers,
  getPendingMembers,
  approveMember,
  rejectMember,
  updateMemberRole,
  removeMember,
  // searchInvitableUsers는 새 명세에 없어 제거됨
  inviteUser,
} from '../../api/user/userService';
import {
  JoinRequestResponse,
  WorkspaceMemberResponse,
  WorkspaceMemberRole,
  WorkspaceSettingsResponse,
} from '../../types/user';

interface WorkspaceManagementModalProps {
  workspaceId: string;
  workspaceName: string;
  onClose: () => void;
}

const WorkspaceManagementModal: React.FC<WorkspaceManagementModalProps> = ({
  workspaceId,
  onClose,
}) => {
  const { theme } = useTheme();
  // 💡 토큰은 인터셉터가 처리하므로, 여기서는 인증 상태를 확인하는 용도로만 남겨둡니다.
  const { token } = useAuth();

  // ========================================
  // 상태 관리
  // ========================================

  const [activeTab, setActiveTab] = useState<'settings' | 'members'>('settings');

  // 💡 DTO 타입 변경: WorkspaceSettings -> WorkspaceSettingsResponse
  const [settings, setSettings] = useState<WorkspaceSettingsResponse | null>(null);
  const [settingsForm, setSettingsForm] = useState({
    workspaceName: '',
    workspaceDescription: '',
    isPublic: false,
    requiresApproval: false,
    onlyOwnerCanInvite: false,
  });

  // 💡 DTO 타입 변경: WorkspaceMember -> WorkspaceMemberResponse, PendingMemberType -> JoinRequestResponse
  const [members, setMembers] = useState<WorkspaceMemberResponse[]>([]);
  const [pendingMembers, setPendingMembers] = useState<JoinRequestResponse[]>([]);

  // 💡 searchInvitableUsers API 제거로 인한 상태 제거
  // const [invitableUsers, setInvitableUsers] = useState<InvitableUser[]>([]);
  const [inviteUserId, setInviteUserId] = useState(''); // 초대를 위한 임시 userId 입력 필드

  const [searchQuery, setSearchQuery] = useState(''); // 기존 검색창은 사용자 검색으로 대체될 수 있음 (현재는 미사용)
  const [memberSearchQuery, setMemberSearchQuery] = useState('');

  // 로딩 및 에러
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // ========================================
  // 초기 데이터 로드
  // ========================================

  const fetchWorkspaceData = async () => {
    try {
      setLoading(true);
      setError(null);

      // 💡 API 호출 시 token 인자 제거
      const [settingsData, membersData, pendingData] = await Promise.all([
        getWorkspaceSettings(workspaceId),
        getWorkspaceMembers(workspaceId),
        getPendingMembers(workspaceId),
      ]);

      setSettings(settingsData);
      setSettingsForm({
        workspaceName: settingsData.workspaceName,
        workspaceDescription: settingsData.workspaceDescription,
        isPublic: settingsData.isPublic,
        requiresApproval: settingsData.requiresApproval,
        onlyOwnerCanInvite: settingsData.onlyOwnerCanInvite,
      });
      setMembers(membersData);
      setPendingMembers(pendingData);
    } catch (err) {
      console.error('[WorkspaceManagement] 데이터 로드 실패:', err);
      // 401 에러(토큰 만료)는 인터셉터가 처리하므로, 나머지 에러만 처리합니다.
      setError('워크스페이스 정보를 불러오는데 실패했습니다. (토큰 재시도 실패 등)');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!token) {
      // 토큰이 없으면 인터셉터가 리다이렉트하지만, 혹시 모를 상황 대비
      setError('인증 토큰이 없습니다. 다시 로그인해 주세요.');
      return;
    }
    fetchWorkspaceData();
  }, [workspaceId, token]);

  // 💡 searchInvitableUsers 관련 useEffect 제거됨 (API 미지원)

  // ========================================
  // 기본정보 핸들러
  // ========================================

  const handleSaveSettings = async () => {
    try {
      setLoading(true);
      setError(null);
      // 💡 API 호출 시 token 인자 제거
      const updated = await updateWorkspaceSettings(workspaceId, settingsForm);
      setSettings(updated);
      console.log('워크스페이스 설정이 저장되었습니다.');
    } catch (err) {
      console.error('[WorkspaceManagement] 설정 저장 실패:', err);
      setError('설정 저장에 실패했습니다. (권한 오류 또는 서버 문제)');
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 회원관리 핸들러
  // ========================================

  const handleInviteUserByUserId = async () => {
    if (!inviteUserId.trim()) {
      setError('초대할 사용자의 ID를 입력해주세요.');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      // 💡 API 호출 시 token 인자 제거
      await inviteUser(workspaceId, inviteUserId);

      // 목록 새로고침
      await fetchWorkspaceData();

      setInviteUserId(''); // 입력창 초기화
      console.log(`사용자 ID ${inviteUserId}에 대한 초대가 완료되었습니다.`);
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 초대 실패:', err);
      setError('회원 초대에 실패했습니다. (유효하지 않은 User ID 또는 권한 오류)');
    } finally {
      setLoading(false);
    }
  };

  const handleApproveMember = async (userId: string) => {
    try {
      setLoading(true);
      setError(null);
      // 💡 API 호출 시 token 인자 제거
      await approveMember(workspaceId, userId);

      // 목록 새로고침
      await fetchWorkspaceData();

      console.log('회원 승인이 완료되었습니다.');
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 승인 실패:', err);
      setError('회원 승인에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleRejectMember = async (userId: string) => {
    try {
      setLoading(true);
      setError(null);
      // 💡 API 호출 시 token 인자 제거
      await rejectMember(workspaceId, userId);

      // 목록 새로고침
      await fetchWorkspaceData();

      console.log('회원 요청을 거절했습니다.');
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 거절 실패:', err);
      setError('회원 거절에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateRole = async (
    memberId: string,
    currentRole: WorkspaceMemberRole,
    newRole: 'ADMIN' | 'MEMBER',
  ) => {
    if (currentRole === newRole) return;

    // OWNER는 역할 변경 API의 대상이 아니므로 예외 처리
    if (currentRole === 'OWNER') {
      setError('OWNER 역할은 변경할 수 없습니다.');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      // 💡 API 호출 시 token 인자 제거. memberId 사용.
      await updateMemberRole(workspaceId, memberId, newRole);

      // 목록 새로고침
      await fetchWorkspaceData();

      console.log(`회원 역할이 ${newRole}로 변경되었습니다.`);
    } catch (err) {
      console.error('[WorkspaceManagement] 역할 변경 실패:', err);
      setError('역할 변경에 실패했습니다. (권한 오류 확인)');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveMember = async (memberId: string, userName: string) => {
    // 💡 confirm() 제거 및 사용자에게 안내
    console.warn(`[Confirmation] 사용자 ${userName} 퇴출을 진행합니다.`);

    try {
      setLoading(true);
      setError(null);
      // 💡 API 호출 시 token 인자 제거. memberId 사용.
      await removeMember(workspaceId, memberId);

      // 목록 새로고침
      await fetchWorkspaceData();

      console.log(`${userName}님이 퇴출되었습니다.`);
    } catch (err) {
      console.error('[WorkspaceManagement] 회원 퇴출 실패:', err);
      setError('회원 퇴출에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // 회원 목록 필터링
  // ========================================

  const filteredMembers = useMemo(() => {
    if (!memberSearchQuery.trim()) return members;
    const query = memberSearchQuery.toLowerCase();
    return members.filter(
      (member) =>
        member.userName.toLowerCase().includes(query) ||
        member.userEmail.toLowerCase().includes(query) ||
        member.roleName.toLowerCase().includes(query),
    );
  }, [members, memberSearchQuery]);

  // ========================================
  // 렌더링
  // ========================================

  if (!settings && loading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-700">워크스페이스 정보를 불러오는 중...</p>
        </div>
      </div>
    );
  }

  // 초기 데이터 로드 실패 시
  if (error && !settings) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div className="bg-white p-8 rounded-xl shadow-lg">
          <p className="text-red-700 font-semibold mb-4">오류 발생</p>
          <p className="text-sm text-gray-700">{error}</p>
          <button
            onClick={onClose}
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
      onClick={onClose}
    >
      <div className="relative w-full max-w-2xl" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} ${theme.effects.borderRadius} shadow-xl max-h-[90vh] overflow-y-auto`}
        >
          {/* 헤더 */}
          <div className="flex items-center justify-between p-4 pb-3">
            <h2 className={`${theme.font.size.base} font-bold text-gray-800`}>
              워크스페이스 관리 ({settings?.workspaceName || '불러오는 중...'})
            </h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 rounded-lg transition"
              title="닫기"
            >
              <X className="w-4 h-4 text-gray-600" />
            </button>
          </div>

          {/* 탭 메뉴 */}
          <div className="flex border-b border-gray-200 px-6">
            <button
              onClick={() => setActiveTab('settings')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'settings' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              기본정보
              {activeTab === 'settings' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></div>
              )}
            </button>
            <button
              onClick={() => setActiveTab('members')}
              className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
                activeTab === 'members' ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              회원관리
              {activeTab === 'members' && (
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

            {/* 기본정보 탭 */}
            {activeTab === 'settings' && (
              <div className="space-y-4">
                <div>
                  <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                    워크스페이스 이름:
                  </label>
                  <input
                    type="text"
                    value={settingsForm.workspaceName}
                    onChange={(e) =>
                      setSettingsForm({ ...settingsForm, workspaceName: e.target.value })
                    }
                    className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                    placeholder="워크스페이스 이름"
                  />
                </div>

                <div>
                  <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                    설명:
                  </label>
                  <textarea
                    value={settingsForm.workspaceDescription}
                    onChange={(e) =>
                      setSettingsForm({ ...settingsForm, workspaceDescription: e.target.value })
                    }
                    rows={3}
                    className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                    placeholder="워크스페이스 설명"
                  />
                </div>

                <div className="space-y-3">
                  {/* 공개/비공개 토글 */}
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="text-sm font-medium text-gray-700">공개 워크스페이스</p>
                      <p className="text-xs text-gray-500">
                        공개 시 다른 사용자가 검색하여 가입 신청할 수 있습니다
                      </p>
                    </div>
                    <button
                      onClick={() =>
                        setSettingsForm({ ...settingsForm, isPublic: !settingsForm.isPublic })
                      }
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        settingsForm.isPublic ? 'bg-blue-600' : 'bg-gray-300'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          settingsForm.isPublic ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>

                  {/* 승인제/비승인제 토글 */}
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="text-sm font-medium text-gray-700">가입 승인제</p>
                      <p className="text-xs text-gray-500">
                        활성화 시 가입 신청을 승인해야 회원이 됩니다
                      </p>
                    </div>
                    <button
                      onClick={() =>
                        setSettingsForm({
                          ...settingsForm,
                          requiresApproval: !settingsForm.requiresApproval,
                        })
                      }
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        settingsForm.requiresApproval ? 'bg-blue-600' : 'bg-gray-300'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          settingsForm.requiresApproval ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>

                  {/* OWNER 전용 초대 토글 */}
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div>
                      <p className="text-sm font-medium text-gray-700">OWNER만 초대 가능</p>
                      <p className="text-xs text-gray-500">
                        활성화 시 OWNER만 새 회원을 초대할 수 있습니다
                      </p>
                    </div>
                    <button
                      onClick={() =>
                        setSettingsForm({
                          ...settingsForm,
                          onlyOwnerCanInvite: !settingsForm.onlyOwnerCanInvite,
                        })
                      }
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        settingsForm.onlyOwnerCanInvite ? 'bg-blue-600' : 'bg-gray-300'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          settingsForm.onlyOwnerCanInvite ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>
                </div>

                <button
                  onClick={handleSaveSettings}
                  disabled={loading}
                  className={`w-full ${theme.colors.primary} text-white py-3 ${
                    theme.effects.borderRadius
                  } font-semibold transition ${
                    loading ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-90'
                  }`}
                >
                  {loading ? '저장 중...' : '설정 저장'}
                </button>
              </div>
            )}

            {/* 회원관리 탭 */}
            {activeTab === 'members' && (
              <div className="space-y-6">
                {/* 💡 초대 기능 업데이트: 검색 API 제거, userId 직접 입력으로 변경 */}
                <div>
                  <label className={`block ${theme.font.size.xs} mb-2 text-gray-500 font-medium`}>
                    사용자 ID로 직접 초대 (디버그/테스트용):
                  </label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      placeholder="초대할 사용자의 UUID (userId) 입력"
                      value={inviteUserId}
                      onChange={(e) => setInviteUserId(e.target.value)}
                      className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                    />
                    <button
                      onClick={handleInviteUserByUserId}
                      disabled={loading || !inviteUserId.trim()}
                      className="px-4 py-2 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition disabled:opacity-50 font-semibold"
                    >
                      초대
                    </button>
                  </div>
                </div>

                {/* 💡 초대 가능 회원 목록 제거됨 (API 미지원) */}

                {/* 승인 대기 목록 */}
                {pendingMembers?.length > 0 && (
                  <div className="border border-gray-200 rounded-lg p-4">
                    <p className="text-sm font-semibold text-gray-700 mb-3">
                      승인 대기 목록 ({pendingMembers?.length}명)
                    </p>
                    <div className="space-y-2">
                      {/* pendingMember는 JoinRequestResponse 타입이며, userId와 userName을 가지고 있음 */}
                      {pendingMembers?.map((member) => (
                        <div
                          key={member.id} // 요청 ID로 key 사용
                          className="flex items-center justify-between bg-gray-50 p-3 rounded border border-gray-200"
                        >
                          <div>
                            <p className="text-sm font-medium text-gray-700">{member.userName}</p>
                            <p className="text-xs text-gray-500">{member.userEmail}</p>
                          </div>
                          <div className="flex gap-2">
                            <button
                              onClick={() => handleApproveMember(member.userId)} // userId로 승인
                              disabled={loading}
                              className="px-3 py-1 bg-green-500 text-white text-xs rounded hover:bg-green-600 transition disabled:opacity-50"
                            >
                              승인
                            </button>
                            <button
                              onClick={() => handleRejectMember(member.userId)} // userId로 거절
                              disabled={loading}
                              className="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 transition disabled:opacity-50"
                            >
                              거절
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* 조직 회원 목록 */}
                <div className="border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3 gap-3">
                    <p className="text-sm font-semibold text-gray-700 whitespace-nowrap">
                      조직 회원 목록 ({members?.length}명)
                    </p>
                    <div className="relative flex-1 max-w-xs">
                      <input
                        type="text"
                        placeholder="이름, 이메일, 역할 검색"
                        value={memberSearchQuery}
                        onChange={(e) => setMemberSearchQuery(e.target.value)}
                        className={`w-full px-3 pl-8 py-1.5 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} text-xs ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                      />
                      <Search className="absolute left-2.5 top-1/2 transform -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
                    </div>
                  </div>
                  <div className="space-y-2">
                    {/* member는 WorkspaceMemberResponse 타입이며, id(멤버 ID), userId, roleName을 가지고 있음 */}
                    {filteredMembers?.map((member) => (
                      <div
                        key={member?.id} // 멤버 ID로 key 사용
                        className="flex items-center justify-between bg-gray-50 p-3 rounded border border-gray-200"
                      >
                        <div>
                          <p className="text-sm font-medium text-gray-700">
                            {member?.userName}{' '}
                            <span className="text-xs text-blue-600 font-semibold">
                              ({member?.roleName})
                            </span>
                          </p>
                          <p className="text-xs text-gray-500">{member?.userEmail}</p>
                        </div>
                        {member?.roleName !== 'OWNER' && (
                          <div className="flex gap-2">
                            {member?.roleName === 'MEMBER' && (
                              <button
                                // 💡 member.id (멤버 ID) 사용, 새로운 역할: ADMIN
                                onClick={() =>
                                  handleUpdateRole(member?.id, member?.roleName, 'ADMIN')
                                }
                                disabled={loading}
                                className="px-3 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 transition disabled:opacity-50"
                              >
                                ADMIN
                              </button>
                            )}
                            {member?.roleName === 'ADMIN' && (
                              <button
                                // 💡 member.id (멤버 ID) 사용, 새로운 역할: MEMBER
                                onClick={() =>
                                  handleUpdateRole(member?.id, member?.roleName, 'MEMBER')
                                }
                                disabled={loading}
                                className="px-3 py-1 bg-gray-500 text-white text-xs rounded hover:bg-gray-600 transition disabled:opacity-50"
                              >
                                MEMBER
                              </button>
                            )}
                            {/* OWNER 위임 기능은 OWNER ROLE을 지정하지 않고, 별도 API를 사용할 수 있으나 현재 명세에 PUT OWNER는 ADMIN/MEMBER만 가능하므로 OWNER 버튼 제거 */}

                            <button
                              // 💡 member.id (멤버 ID) 사용
                              onClick={() => handleRemoveMember(member?.id, member?.userName)}
                              disabled={loading}
                              className="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 transition disabled:opacity-50"
                            >
                              퇴출
                            </button>
                          </div>
                        )}
                      </div>
                    ))}
                    {filteredMembers?.length === 0 && (
                      <p className="text-center text-sm text-gray-500 py-4">
                        {memberSearchQuery.trim()
                          ? '검색 결과가 없습니다.'
                          : '조직 회원이 없습니다.'}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkspaceManagementModal;
