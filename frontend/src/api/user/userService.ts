// src/api/user/userService.ts

import {
  CreateWorkspaceRequest,
  UpdateProfileRequest,
  UpdateWorkspaceSettingsRequest,
  UserProfileResponse,
  WorkspaceResponse,
  WorkspaceMemberResponse,
  WorkspaceMemberRole, // DTO 타입에서 역할 타입을 가져옴
  WorkspaceSettingsResponse,
  JoinRequestResponse,
  InviteUserRequest,
  // UpdateWorkspaceRequest DTO가 명시되지 않아 임시로 구조를 정의함
} from '../../types/user';
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
 * * Response: { data: WorkspaceResponse }
 */
export const getWorkspace = async (workspaceId: string): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<{ data: WorkspaceResponse }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}`,
  );
  return response.data.data; // data 필드 추출
};

/**
 * 워크스페이스 생성
 * [API] POST /api/workspaces
 * * Response: WorkspaceResponse (API 스펙에 따라 { data: WorkspaceResponse }가 아닐 수 있음)
 */
export const createWorkspace = async (data: CreateWorkspaceRequest): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<WorkspaceResponse> = await userRepoClient.post(
    '/api/workspaces',
    data,
  );
  return response.data;
};

/**
 * 워크스페이스 수정
 * [API] PUT /api/workspaces/{workspaceId}
 * * Response: { data: WorkspaceResponse }
 */
export const updateWorkspace = async (
  workspaceId: string,
  data: { workspaceName?: string; workspaceDescription?: string },
): Promise<WorkspaceResponse> => {
  const response: AxiosResponse<{ data: WorkspaceResponse }> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}`,
    data,
  );
  return response.data.data; // data 필드 추출
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
// Member & Join Request API Functions (회원/가입 요청 관리)
// * WorkspaceMembersTab.tsx 에서 사용되는 주요 함수 그룹
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
 * * Response: { data: JoinRequestResponse[] }
 */
export const getPendingMembers = async (workspaceId: string): Promise<JoinRequestResponse[]> => {
  const response: AxiosResponse<{ data: JoinRequestResponse[] }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/pendingMembers`,
  );
  return response.data.data; // data 필드 추출
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
 * * Response: { data: WorkspaceMemberResponse }
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
  return response.data.data; // data 필드 추출
};

/**
 * 멤버 역할 변경
 * [API] PUT /api/workspaces/{workspaceId}/members/{memberId}/role
 * * Response: { data: WorkspaceMemberResponse }
 */
export const updateMemberRole = async (
  workspaceId: string,
  memberId: string,
  roleName: WorkspaceMemberRole, // DTO 타입을 사용
): Promise<WorkspaceMemberResponse> => {
  const data = { roleName };

  const response: AxiosResponse<{ data: WorkspaceMemberResponse }> = await userRepoClient.put(
    `/api/workspaces/${workspaceId}/members/${memberId}/role`,
    data,
  );
  return response.data.data; // data 필드 추출
};

/**
 * 멤버 제거
 * [API] DELETE /api/workspaces/{workspaceId}/members/{memberId}
 */
export const removeMember = async (workspaceId: string, memberId: string): Promise<void> => {
  await userRepoClient.delete(`/api/workspaces/${workspaceId}/members/${memberId}`);
};

/**
 * 가입 신청 목록 조회 (status 필터 가능)
 * [API] GET /api/workspaces/{workspaceId}/joinRequests
 * * Response: { data: JoinRequestResponse[] }
 */
export const getJoinRequests = async (
  workspaceId: string,
  status?: string, // 'PENDING', 'APPROVED', 'REJECTED'
): Promise<JoinRequestResponse[]> => {
  const response: AxiosResponse<{ data: JoinRequestResponse[] }> = await userRepoClient.get(
    `/api/workspaces/${workspaceId}/joinRequests`,
    { params: { status } },
  );
  return response.data.data; // data 필드 추출
};

/**
 * 워크스페이스 가입 신청
 * [API] POST /api/workspaces/join-requests
 * * Response: { data: JoinRequestResponse }
 */
export const createJoinRequest = async (workspaceId: string): Promise<JoinRequestResponse> => {
  const data = { workspaceId };
  const response: AxiosResponse<{ data: JoinRequestResponse }> = await userRepoClient.post(
    '/api/workspaces/join-requests',
    data,
  );
  return response.data.data; // data 필드 추출
};

// ========================================
// UserProfile API Functions
// ========================================

/**
 * 내 프로필 조회 (기본 프로필)
 * [API] GET /api/profiles/me
 * * Response: { data: UserProfileResponse }
 */
export const getMyProfile = async (): Promise<UserProfileResponse> => {
  const response: AxiosResponse<{ data: UserProfileResponse }> = await userRepoClient.get(
    '/api/profiles/me',
  );
  return response.data.data; // data 필드 추출
};

/**
 * 내 모든 프로필 조회 (기본 프로필 + 워크스페이스별 프로필)
 * [API] GET /api/profiles/all/me
 * * Response: { data: UserProfileResponse[] }
 */
export const getAllMyProfiles = async (): Promise<UserProfileResponse[]> => {
  const response: AxiosResponse<{ data: UserProfileResponse[] }> = await userRepoClient.get(
    '/api/profiles/all/me',
  );
  return response.data.data; // data 필드 추출
};

/**
 * 내 프로필 정보 통합 업데이트 (기본 프로필)
 * [API] PUT /api/profiles/me
 * * Response: { data: UserProfileResponse }
 */
export const updateMyProfile = async (data: UpdateProfileRequest): Promise<UserProfileResponse> => {
  const response: AxiosResponse<{ data: UserProfileResponse }> = await userRepoClient.put(
    '/api/profiles/me',
    data,
  );
  return response.data.data; // data 필드 추출
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
