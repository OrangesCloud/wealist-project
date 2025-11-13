import {
  CreateWorkspaceRequest,
  UpdateProfileRequest,
  UpdateWorkspaceSettingsRequest,
  UserProfileResponse,
  WorkspaceResponse,
  WorkspaceMemberResponse, // New DTO from OpenAPI spec
  WorkspaceSettingsResponse, // New DTO from OpenAPI spec
  JoinRequestResponse, // New DTO from OpenAPI spec
  InviteUserRequest,
} from '../../types/user'; // DTO 인터페이스는 types/user 파일에서 가져온다고 가정
import { userRepoClient } from '../apiConfig';
import { AxiosResponse } from 'axios';

// ========================================
// Workspace API Functions (워크스페이스 전체 관리)
// ========================================

/**
 * 워크스페이스 목록 조회 (현재 사용자가 속한 모든 워크스페이스)
 * [API] GET /api/workspaces/all
 */
export const getMyWorkspaces = async (): Promise<WorkspaceResponse[]> => {
  const response: AxiosResponse<WorkspaceResponse[]> = await userRepoClient.get(
    '/api/workspaces/all',
  );
  return response.data;
};

/**
 * 퍼블릭 워크스페이스 목록 조회
 * [API] GET /api/workspaces
 */
export const getPublicWorkspaces = async (): Promise<WorkspaceResponse[]> => {
  const response: AxiosResponse<WorkspaceResponse[]> = await userRepoClient.get('/api/workspaces');
  return response.data;
};

/**
 * 워크스페이스 검색
 * [API] GET /api/workspaces/search?query={query}
 */
export const searchWorkspaces = async (query: string): Promise<WorkspaceResponse[]> => {
  const response: AxiosResponse<WorkspaceResponse[]> = await userRepoClient.get(
    '/api/workspaces/search',
    { params: { query } },
  );
  return response.data;
};

/**
 * 특정 워크스페이스 조회
 * [API] GET /api/workspaces/{workspaceId}
 */
export const getWorkspace = async (workspaceId: string): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<{ data: WorkspaceResponse }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}`,
  );
  return response.data.data;
};

/**
 * 워크스페이스 생성
 * [API] POST /api/workspaces
 * [Response] { data: WorkspaceResponse }
 */
export const createWorkspace = async (data: CreateWorkspaceRequest): Promise<WorkspaceResponse> => {
  try {
    const response: AxiosResponse<WorkspaceResponse> = await userRepoClient.post(
      '/api/workspaces',
      data,
    );
    return response.data;
  } catch (error) {
    console.error('createWorkspace error:', error);
    throw error;
  }
};

/**
 * 워크스페이스 수정
 * [API] PUT /api/workspaces/{workspaceId}
 * [Body] UpdateWorkspaceRequest (UpdateWorkspaceRequest DTO는 DTO 파일에 정의 필요)
 */
export const updateWorkspace = async (
  workspaceId: string,
  data: { workspaceName?: string; workspaceDescription?: string },
): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<{ data: WorkspaceResponse }> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}`,
    data,
  );
  return response.data.data;
};

/**
 * 워크스페이스 삭제 (소프트 삭제)
 * [API] DELETE /api/workspaces/{workspaceId}
 */
export const deleteWorkspace = async (workspaceId: string): Promise<void> => {
  await userRepoClient.delete(`/api/workspaces/${workspaceId}`);
};

/**
 * 워크스페이스 설정 조회
 * [API] GET /api/workspaces/{workspaceId}/settings
 */
export const getWorkspaceSettings = async (
  workspaceId: string,
): Promise<WorkspaceSettingsResponse> => {
  const response: AxiosResponse<WorkspaceSettingsResponse> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/settings`,
  );
  console.log(response.data);
  return response.data;
};

/**
 * 워크스페이스 설정 수정
 * [API] PUT /api/workspaces/{workspaceId}/settings
 */
export const updateWorkspaceSettings = async (
  workspaceId: string,
  data: UpdateWorkspaceSettingsRequest,
): Promise<WorkspaceSettingsResponse> => {
  const response: AxiosResponse<WorkspaceSettingsResponse> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}/settings`,
    data,
  );
  return response.data;
};

// ========================================
// Member & Join Request API Functions
// ========================================

/**
 * 워크스페이스 회원 목록 조회
 * [API] GET /api/workspaces/{workspaceId}/members
 */
export const getWorkspaceMembers = async (
  workspaceId: string,
): Promise<WorkspaceMemberResponse[]> => {
  const response: AxiosResponse<WorkspaceMemberResponse[]> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/members`,
  );
  return response.data;
};

/**
 * 승인 대기 회원 목록 조회
 * [API] GET /api/workspaces/{workspaceId}/pendingMembers
 */
export const getPendingMembers = async (workspaceId: string): Promise<JoinRequestResponse[]> => {
  const response: AxiosResponse<{ data: JoinRequestResponse[] }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/pendingMembers`,
  );
  return response.data.data;
};

/**
 * 가입 신청 목록 조회 (status 필터 가능)
 * [API] GET /api/workspaces/{workspaceId}/joinRequests
 */
export const getJoinRequests = async (
  workspaceId: string,
  status?: string, // 'PENDING', 'APPROVED', 'REJECTED'
): Promise<JoinRequestResponse[]> => {
  const response: AxiosResponse<{ data: JoinRequestResponse[] }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/joinRequests`,
    { params: { status } },
  );
  return response.data.data;
};

/**
 * 멤버 역할 변경
 * [API] PUT /api/workspaces/{workspaceId}/members/{memberId}/role
 */
export const updateMemberRole = async (
  workspaceId: string,
  memberId: string,
  roleName: 'ADMIN' | 'MEMBER',
): Promise<WorkspaceMemberResponse> => {
  const data = { roleName };

  const response: AxiosResponse<{ data: WorkspaceMemberResponse }> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}/members/${memberId}/role`,
    data,
  );
  return response.data.data;
};

/**
 * 멤버 제거
 * [API] DELETE /api/workspaces/{workspaceId}/members/{memberId}
 */
export const removeMember = async (workspaceId: string, memberId: string): Promise<void> => {
  await userRepoClient.delete(`/api/workspaces/${workspaceId}/members/${memberId}`);
};

/**
 * 가입 신청 승인
 * [API] POST /api/workspaces/{workspaceId}/members/{userId}/approve
 */
export const approveMember = async (workspaceId: string, userId: string): Promise<void> => {
  await userRepoClient.post(`/api/workspaces/${workspaceId}/members/${userId}/approve`, {});
};

/**
 * 가입 신청 거절
 * [API] POST /api/workspaces/{workspaceId}/members/{userId}/reject
 */
export const rejectMember = async (workspaceId: string, userId: string): Promise<void> => {
  await userRepoClient.post(`/api/workspaces/${workspaceId}/members/${userId}/reject`, {});
};

/**
 * 워크스페이스에 사용자 초대 (userId 기준)
 * [API] POST /api/workspaces/{workspaceId}/members/invite
 */
export const inviteUser = async (
  workspaceId: string,
  userId: string,
): Promise<WorkspaceMemberResponse> => {
  const data: InviteUserRequest = { userId };

  const response: AxiosResponse<{ data: WorkspaceMemberResponse }> = await userRepoClient.post(
    `/api/workspaces/${workspaceId}/members/invite`,
    data,
  );
  return response.data.data;
};

/**
 * 워크스페이스 가입 신청
 * [API] POST /api/workspaces/join-requests
 */
export const createJoinRequest = async (workspaceId: string): Promise<JoinRequestResponse> => {
  const data = { workspaceId };
  const response: AxiosResponse<{ data: JoinRequestResponse }> = await userRepoClient.post(
    '/api/workspaces/join-requests',
    data,
  );
  return response.data.data;
};

// ========================================
// UserProfile API Functions
// ========================================

/**
 * 내 프로필 조회 (기본 프로필)
 * [API] GET /api/profiles/me
 */
export const getMyProfile = async (): Promise<UserProfileResponse> => {
  const response: AxiosResponse<{ data: UserProfileResponse }> = await userRepoClient.get(
    '/api/profiles/me',
  );
  return response.data.data;
};

/**
 * 내 모든 프로필 조회 (기본 프로필 + 워크스페이스별 프로필)
 * [API] GET /api/profiles/all/me
 */
export const getAllMyProfiles = async (): Promise<UserProfileResponse[]> => {
  const response: AxiosResponse<{ data: UserProfileResponse[] }> = await userRepoClient.get(
    '/api/profiles/all/me',
  );
  return response.data.data;
};

/**
 * 내 프로필 정보 통합 업데이트 (기본 프로필)
 * [API] PUT /api/profiles/me
 */
export const updateMyProfile = async (data: UpdateProfileRequest): Promise<UserProfileResponse> => {
  const response: AxiosResponse<{ data: UserProfileResponse }> = await userRepoClient.put(
    '/api/profiles/me',
    data,
  );
  return response.data.data;
};

// ========================================
// [제거/대체됨] 워크스페이스 프로필 관리 함수 (호환성 유지용)
// ========================================

/**
 * [제거됨] 워크스페이스 프로필 조회 (GET /api/profiles/workspace/{workspaceId})
 * @deprecated 프론트엔드에서 `getAllMyProfiles()`를 호출하여 필터링해야 합니다.
 */
export const getWorkspaceProfile = async (
  workspaceId: string,
): Promise<UserProfileResponse | null> => {
  return null;
};

/**
 * [제거됨] 워크스페이스 프로필 생성/수정 (PUT /api/profiles/workspace/{workspaceId})
 * @deprecated 이 엔드포인트는 제거되었으며, 백엔드 구현이 필요합니다.
 */
export const updateWorkspaceProfile = async (
  workspaceId: string,
  data: UpdateProfileRequest,
): Promise<UserProfileResponse> => {
  throw new Error('워크스페이스별 프로필 업데이트 엔드포인트가 제거되었습니다. (백엔드 구현 필요)');
};

// ========================================
// New API Functions (기타)
// ========================================

/**
 * 기본 워크스페이스 설정
 * [API] POST /api/workspaces/default
 */
export const setDefaultWorkspace = async (workspaceId: string): Promise<void> => {
  const data = { workspaceId };
  await userRepoClient.post('/api/workspaces/default', data);
};
