import { boardServiceClient } from '../apiConfig';
import { AxiosResponse } from 'axios';

import {
  SuccessResponse,
  BoardResponse,
  BoardDetailResponse,
  CreateBoardRequest,
  UpdateBoardRequest,
  BoardFilters,
  ProjectResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  PaginatedProjectsResponse,
  ProjectInitSettingsResponse,
  ProjectMemberResponse,
  UpdateProjectMemberRoleRequest,
  ProjectJoinRequestResponse,
  CreateProjectJoinRequestRequest,
  UpdateProjectJoinRequestRequest,
  CommentResponse,
  CreateCommentRequest,
  UpdateCommentRequest,
  ParticipantResponse,
  AddParticipantRequest,
  FieldOptionResponse,
  CreateFieldOptionRequest,
  UpdateFieldOptionRequest,
} from '../../types/board';

/**
 * ========================================
 * 목업 모드 전환
 * ========================================
 */
const USE_MOCK_DATA = false; // 💡 목업 모드 ON/OFF

// ============================================================================
// 💡 [신규] 프로젝트 초기 데이터 로드 API
// ============================================================================

/**
 * 프로젝트 초기 페이지 로드에 필요한 모든 데이터를 조회합니다.
 * [API] GET /api/projects/{projectId}/init-settings
 */
export const getProjectInitSettings = async (
  projectId: string,
): Promise<ProjectInitSettingsResponse> => {
  if (USE_MOCK_DATA) {
    return {
      project: {
        projectId: 'mock-project-id',
        workspaceId: 'mock-workspace-id',
        workspaceName: 'Mock Workspace',
        workspaceEmail: 'workspace@example.com',
        name: 'Mock Project',
        description: 'Mock project description',
        ownerId: 'mock-owner-id',
        isPublic: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      fields: [],
      fieldTypes: [],
      defaultViewId: 'mock-view-id',
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectInitSettingsResponse>> =
      await boardServiceClient.get(`/projects/${projectId}/init-settings`);
    return response.data.data;
  } catch (error) {
    console.error('getProjectInitSettings error:', error);
    throw error;
  }
};

// ============================================================================
// 프로젝트 관련 API
// ============================================================================

/**
 * 워크스페이스의 모든 프로젝트를 조회합니다.
 * [API] GET /api/projects/workspace/{workspaceId}
 */
export const getProjects = async (workspaceId: string): Promise<ProjectResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        projectId: 'mock-project-1',
        workspaceId: workspaceId,
        name: 'Mock Project 1',
        description: 'First mock project',
        ownerId: 'mock-owner-1',
        ownerName: 'Mock Owner',
        ownerEmail: 'owner@example.com',
        isPublic: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectResponse[]>> =
      await boardServiceClient.get(`/projects/workspace/${workspaceId}`);
    return response.data.data || [];
  } catch (error) {
    console.error('getProjects error:', error);
    throw error;
  }
};

/**
 * 워크스페이스의 기본(default) 프로젝트를 조회합니다.
 * [API] GET /api/projects/workspace/{workspaceId}/default
 */
export const getDefaultProject = async (workspaceId: string): Promise<ProjectResponse> => {
  if (USE_MOCK_DATA) {
    return {
      projectId: 'mock-default-project',
      workspaceId: workspaceId,
      name: 'Default Project',
      description: 'Default mock project',
      ownerId: 'mock-owner',
      ownerName: 'Mock Owner',
      ownerEmail: 'owner@example.com',
      isPublic: true,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectResponse>> = await boardServiceClient.get(
      `/projects/workspace/${workspaceId}/default`,
    );
    return response.data.data;
  } catch (error) {
    console.error('getDefaultProject error:', error);
    throw error;
  }
};

/**
 * 특정 프로젝트를 조회합니다.
 * [API] GET /api/projects/{projectId}
 */
export const getProject = async (projectId: string): Promise<ProjectResponse> => {
  if (USE_MOCK_DATA) {
    return {
      projectId: projectId,
      workspaceId: 'mock-workspace-id',
      name: 'Mock Project',
      description: 'Mock project description',
      ownerId: 'mock-owner-id',
      ownerName: 'Mock Owner',
      ownerEmail: 'owner@example.com',
      isPublic: true,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectResponse>> = await boardServiceClient.get(
      `/projects/${projectId}`,
    );
    return response.data.data;
  } catch (error) {
    console.error('getProject error:', error);
    throw error;
  }
};

/**
 * 새로운 프로젝트를 생성합니다.
 * [API] POST /api/projects
 */
export const createProject = async (data: CreateProjectRequest): Promise<ProjectResponse> => {
  if (USE_MOCK_DATA) {
    return {
      projectId: 'mock-new-project',
      workspaceId: data.workspaceId,
      name: data.name,
      description: data.description || '',
      ownerId: 'mock-owner-id',
      ownerName: 'Mock Owner',
      ownerEmail: 'owner@example.com',
      isPublic: true,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectResponse>> = await boardServiceClient.post(
      '/projects',
      data,
    );
    return response.data.data;
  } catch (error) {
    console.error('createProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 업데이트합니다.
 * [API] PUT /api/projects/{projectId}
 */
export const updateProject = async (
  projectId: string,
  data: UpdateProjectRequest,
): Promise<ProjectResponse> => {
  if (USE_MOCK_DATA) {
    return {
      projectId: projectId,
      workspaceId: 'mock-workspace-id',
      name: data.name || 'Updated Project',
      description: data.description || 'Updated description',
      ownerId: 'mock-owner-id',
      ownerName: 'Mock Owner',
      ownerEmail: 'owner@example.com',
      isPublic: true,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectResponse>> = await boardServiceClient.put(
      `/projects/${projectId}`,
      data,
    );
    return response.data.data;
  } catch (error) {
    console.error('updateProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 삭제합니다.
 * [API] DELETE /api/projects/{projectId}
 */
export const deleteProject = async (projectId: string): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.delete(`/projects/${projectId}`);
  } catch (error) {
    console.error('deleteProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 검색합니다.
 * [API] GET /api/projects/search
 */
export const searchProjects = async (
  workspaceId: string,
  query: string,
  page: number = 1,
  limit: number = 10,
): Promise<PaginatedProjectsResponse> => {
  if (USE_MOCK_DATA) {
    return {
      projects: [],
      total: 0,
      page: page,
      limit: limit,
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<PaginatedProjectsResponse>> =
      await boardServiceClient.get('/projects/search', {
        params: { workspaceId, query, page, limit },
      });
    return response.data.data;
  } catch (error) {
    console.error('searchProjects error:', error);
    throw error;
  }
};

// ============================================================================
// 프로젝트 멤버 관련 API
// ============================================================================

/**
 * 프로젝트의 모든 멤버를 조회합니다.
 * [API] GET /api/projects/{projectId}/members
 */
export const getProjectMembers = async (projectId: string): Promise<ProjectMemberResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        memberId: 'mock-member-1',
        projectId: projectId,
        userId: 'mock-user-1',
        userName: 'Mock User',
        userEmail: 'user@example.com',
        roleName: 'MEMBER',
        joinedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectMemberResponse[]>> =
      await boardServiceClient.get(`/projects/${projectId}/members`);
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectMembers error:', error);
    throw error;
  }
};

/**
 * 프로젝트 멤버의 역할을 변경합니다.
 * [API] PUT /api/projects/{projectId}/members/{memberId}/role
 */
export const updateProjectMemberRole = async (
  projectId: string,
  memberId: string,
  data: UpdateProjectMemberRoleRequest,
): Promise<ProjectMemberResponse> => {
  if (USE_MOCK_DATA) {
    return {
      memberId: memberId,
      projectId: projectId,
      userId: 'mock-user-id',
      userName: 'Mock User',
      userEmail: 'user@example.com',
      roleName: data.roleName,
      joinedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectMemberResponse>> =
      await boardServiceClient.put(`/projects/${projectId}/members/${memberId}/role`, data);
    return response.data.data;
  } catch (error) {
    console.error('updateProjectMemberRole error:', error);
    throw error;
  }
};

/**
 * 프로젝트에서 멤버를 제거합니다.
 * [API] DELETE /api/projects/{projectId}/members/{memberId}
 */
export const removeProjectMember = async (projectId: string, memberId: string): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.delete(`/projects/${projectId}/members/${memberId}`);
  } catch (error) {
    console.error('removeProjectMember error:', error);
    throw error;
  }
};

// ============================================================================
// 프로젝트 가입 요청 관련 API
// ============================================================================

/**
 * 프로젝트 가입 요청 목록을 조회합니다.
 * [API] GET /api/projects/{projectId}/join-requests
 */
export const getProjectJoinRequests = async (
  projectId: string,
  status?: string,
): Promise<ProjectJoinRequestResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        requestId: 'mock-request-1',
        projectId: projectId,
        userId: 'mock-user-1',
        userName: 'Mock User',
        userEmail: 'user@example.com',
        status: 'PENDING',
        requestedAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectJoinRequestResponse[]>> =
      await boardServiceClient.get(`/projects/${projectId}/join-requests`, {
        params: { status },
      });
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectJoinRequests error:', error);
    throw error;
  }
};

/**
 * 프로젝트 가입 요청을 생성합니다.
 * [API] POST /api/join-requests
 */
export const createProjectJoinRequest = async (
  data: CreateProjectJoinRequestRequest,
): Promise<ProjectJoinRequestResponse> => {
  if (USE_MOCK_DATA) {
    return {
      requestId: 'mock-new-request',
      projectId: data.projectId,
      userId: 'mock-user-id',
      userName: 'Mock User',
      userEmail: 'user@example.com',
      status: 'PENDING',
      requestedAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectJoinRequestResponse>> =
      await boardServiceClient.post('/join-requests', data);
    return response.data.data;
  } catch (error) {
    console.error('createProjectJoinRequest error:', error);
    throw error;
  }
};

/**
 * 프로젝트 가입 요청을 승인/거부합니다.
 * [API] PUT /api/join-requests/{joinRequestId}
 */
export const updateProjectJoinRequest = async (
  joinRequestId: string,
  data: UpdateProjectJoinRequestRequest,
): Promise<ProjectJoinRequestResponse> => {
  if (USE_MOCK_DATA) {
    return {
      requestId: joinRequestId,
      projectId: 'mock-project-id',
      userId: 'mock-user-id',
      userName: 'Mock User',
      userEmail: 'user@example.com',
      status: data.status,
      requestedAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<ProjectJoinRequestResponse>> =
      await boardServiceClient.put(`/join-requests/${joinRequestId}`, data);
    return response.data.data;
  } catch (error) {
    console.error('updateProjectJoinRequest error:', error);
    throw error;
  }
};

// ============================================================================
// 보드 관련 API
// ============================================================================

/**
 * 프로젝트의 모든 보드를 조회합니다 (쿼리 파라미터 방식).
 * [API] GET /api/boards?projectId={projectId}
 */
export const getBoards = async (
  projectId: string,
  filters?: BoardFilters,
): Promise<BoardResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        boardId: 'mock-board-1',
        projectId: projectId,
        title: 'Mock Board',
        content: 'Mock content',
        customFields: {
          stage: 'in_progress',
          role: 'developer',
          importance: 'normal',
        },
        authorId: 'mock-author',
        assigneeId: 'mock-assignee',
        dueDate: new Date().toISOString(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const params: any = { projectId };
    if (filters?.customFields) {
      params.customFields = JSON.stringify(filters.customFields);
    }

    const response: AxiosResponse<SuccessResponse<BoardResponse[]>> = await boardServiceClient.get(
      '/boards',
      { params },
    );
    return response.data.data || [];
  } catch (error) {
    console.error('getBoards error:', error);
    throw error;
  }
};

/**
 * 프로젝트의 모든 보드를 조회합니다 (경로 파라미터 방식).
 * [API] GET /api/boards/project/{projectId}
 */
export const getBoardsByProject = async (
  projectId: string,
  filters?: BoardFilters,
): Promise<BoardResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        boardId: 'mock-board-1',
        projectId: projectId,
        title: 'Mock Board',
        content: 'Mock content',
        customFields: {
          stage: 'in_progress',
          role: 'developer',
          importance: 'normal',
        },
        authorId: 'mock-author',
        assigneeId: 'mock-assignee',
        dueDate: new Date().toISOString(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const params: any = {};
    if (filters?.customFields) {
      params.customFields = JSON.stringify(filters.customFields);
    }

    const response: AxiosResponse<SuccessResponse<BoardResponse[]>> = await boardServiceClient.get(
      `/boards/project/${projectId}`,
      { params },
    );
    return response.data.data || [];
  } catch (error) {
    console.error('getBoardsByProject error:', error);
    throw error;
  }
};

/**
 * 특정 보드를 조회합니다 (상세 정보 포함).
 * [API] GET /api/boards/{boardId}
 */
export const getBoard = async (boardId: string): Promise<BoardDetailResponse> => {
  if (USE_MOCK_DATA) {
    return {
      boardId: boardId,
      projectId: 'mock-project-id',
      title: 'Mock Board',
      content: 'Mock content',
      customFields: {
        stage: 'in_progress',
        role: 'developer',
        importance: 'normal',
      },
      authorId: 'mock-author',
      assigneeId: 'mock-assignee',
      dueDate: new Date().toISOString(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      participants: [],
      comments: [],
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<BoardDetailResponse>> =
      await boardServiceClient.get(`/boards/${boardId}`);
    return response.data.data;
  } catch (error) {
    console.error('getBoard error:', error);
    throw error;
  }
};

/**
 * 새로운 보드를 생성합니다.
 * [API] POST /api/boards
 */
export const createBoard = async (data: CreateBoardRequest): Promise<BoardResponse> => {
  if (USE_MOCK_DATA) {
    return {
      boardId: 'mock-new-board',
      projectId: data.projectId,
      title: data.title,
      content: data.content || '',
      customFields: data.customFields || {},
      authorId: 'mock-author',
      assigneeId: data.assigneeId || '',
      dueDate: data.dueDate || '',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }
  console.log(data);
  try {
    const response: AxiosResponse<SuccessResponse<BoardResponse>> = await boardServiceClient.post(
      '/boards',
      data,
    );
    return response.data.data;
  } catch (error) {
    console.error('createBoard error:', error);
    throw error;
  }
};

/**
 * 보드를 업데이트합니다.
 * [API] PUT /api/boards/{boardId}
 */
export const updateBoard = async (
  boardId: string,
  data: UpdateBoardRequest,
): Promise<BoardResponse> => {
  if (USE_MOCK_DATA) {
    return {
      boardId: boardId,
      projectId: 'mock-project-id',
      title: data.title || 'Updated Board',
      content: data.content || 'Updated content',
      customFields: data.customFields || {},
      authorId: 'mock-author',
      assigneeId: data.assigneeId || 'mock-assignee',
      dueDate: data.dueDate || '',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<BoardResponse>> = await boardServiceClient.put(
      `/boards/${boardId}`,
      data,
    );
    return response.data.data;
  } catch (error) {
    console.error('updateBoard error:', error);
    throw error;
  }
};

/**
 * 보드를 삭제합니다.
 * [API] DELETE /api/boards/{boardId}
 */
export const deleteBoard = async (boardId: string): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.delete(`/boards/${boardId}`);
  } catch (error) {
    console.error('deleteBoard error:', error);
    throw error;
  }
};

// ============================================================================
// 필드 옵션 관련 API
// ============================================================================

/**
 * 특정 필드 타입의 옵션 목록을 조회합니다.
 * [API] GET /api/field-options?fieldType={fieldType}
 */
export const getFieldOptions = async (
  fieldType: 'stage' | 'role' | 'importance',
): Promise<FieldOptionResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        optionId: 'mock-option-1',
        fieldType: fieldType,
        value: 'in_progress',
        label: 'In Progress',
        color: '#3b82f6',
        displayOrder: 1,
        isSystemDefault: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const response: AxiosResponse<SuccessResponse<FieldOptionResponse[]>> =
      await boardServiceClient.get('/field-options', {
        params: { fieldType },
      });
    return response.data.data || [];
  } catch (error) {
    console.error('getFieldOptions error:', error);
    throw error;
  }
};

/**
 * 새로운 필드 옵션을 생성합니다.
 * [API] POST /api/field-options
 */
export const createFieldOption = async (
  data: CreateFieldOptionRequest,
): Promise<FieldOptionResponse> => {
  if (USE_MOCK_DATA) {
    return {
      optionId: 'mock-new-option',
      fieldType: data.fieldType,
      value: data.value,
      label: data.label,
      color: data.color,
      displayOrder: data.displayOrder || 0,
      isSystemDefault: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<FieldOptionResponse>> =
      await boardServiceClient.post('/field-options', data);
    return response.data.data;
  } catch (error) {
    console.error('createFieldOption error:', error);
    throw error;
  }
};

/**
 * 필드 옵션을 수정합니다.
 * [API] PATCH /api/field-options/{optionId}
 */
export const updateFieldOption = async (
  optionId: string,
  data: UpdateFieldOptionRequest,
): Promise<FieldOptionResponse> => {
  if (USE_MOCK_DATA) {
    return {
      optionId: optionId,
      fieldType: 'stage',
      value: 'updated_value',
      label: data.label || 'Updated Label',
      color: data.color || '#000000',
      displayOrder: data.displayOrder || 0,
      isSystemDefault: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<FieldOptionResponse>> =
      await boardServiceClient.patch(`/field-options/${optionId}`, data);
    return response.data.data;
  } catch (error) {
    console.error('updateFieldOption error:', error);
    throw error;
  }
};

/**
 * 필드 옵션을 삭제합니다.
 * [API] DELETE /api/field-options/{optionId}
 */
export const deleteFieldOption = async (optionId: string): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.delete(`/field-options/${optionId}`);
  } catch (error) {
    console.error('deleteFieldOption error:', error);
    throw error;
  }
};

// ============================================================================
// 댓글 관련 API
// ============================================================================

/**
 * 보드의 모든 댓글을 조회합니다 (쿼리 파라미터 방식).
 * [API] GET /api/comments?boardId={boardId}
 */
export const getComments = async (boardId: string): Promise<CommentResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        commentId: 'mock-comment-1',
        boardId: boardId,
        userId: 'mock-user-1',
        content: 'Mock comment',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const response: AxiosResponse<SuccessResponse<CommentResponse[]>> =
      await boardServiceClient.get('/comments', {
        params: { boardId },
      });
    return response.data.data || [];
  } catch (error) {
    console.error('getComments error:', error);
    throw error;
  }
};

/**
 * 보드의 모든 댓글을 조회합니다 (경로 파라미터 방식).
 * [API] GET /api/comments/board/{boardId}
 */
export const getCommentsByBoard = async (boardId: string): Promise<CommentResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        commentId: 'mock-comment-1',
        boardId: boardId,
        userId: 'mock-user-1',
        content: 'Mock comment',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const response: AxiosResponse<SuccessResponse<CommentResponse[]>> =
      await boardServiceClient.get(`/comments/board/${boardId}`);
    return response.data.data || [];
  } catch (error) {
    console.error('getCommentsByBoard error:', error);
    throw error;
  }
};

/**
 * 새 댓글을 생성합니다.
 * [API] POST /api/comments
 */
export const createComment = async (data: CreateCommentRequest): Promise<CommentResponse> => {
  if (USE_MOCK_DATA) {
    return {
      commentId: 'mock-new-comment',
      boardId: data.boardId,
      userId: 'mock-user-id',
      content: data.content,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<CommentResponse>> = await boardServiceClient.post(
      '/comments',
      data,
    );
    return response.data.data;
  } catch (error) {
    console.error('createComment error:', error);
    throw error;
  }
};

/**
 * 댓글을 수정합니다.
 * [API] PUT /api/comments/{commentId}
 */
export const updateComment = async (
  commentId: string,
  data: UpdateCommentRequest,
): Promise<CommentResponse> => {
  if (USE_MOCK_DATA) {
    return {
      commentId: commentId,
      boardId: 'mock-board-id',
      userId: 'mock-user-id',
      content: data.content,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
  }

  try {
    const response: AxiosResponse<SuccessResponse<CommentResponse>> = await boardServiceClient.put(
      `/comments/${commentId}`,
      data,
    );
    return response.data.data;
  } catch (error) {
    console.error('updateComment error:', error);
    throw error;
  }
};

/**
 * 댓글을 삭제합니다.
 * [API] DELETE /api/comments/{commentId}
 */
export const deleteComment = async (commentId: string): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.delete(`/comments/${commentId}`);
  } catch (error) {
    console.error('deleteComment error:', error);
    throw error;
  }
};

// ============================================================================
// 참여자 관련 API
// ============================================================================

/**
 * 보드의 모든 참여자를 조회합니다.
 * [API] GET /api/participants/board/{boardId}
 */
export const getParticipants = async (boardId: string): Promise<ParticipantResponse[]> => {
  if (USE_MOCK_DATA) {
    return [
      {
        id: 'mock-participant-1',
        boardId: boardId,
        userId: 'mock-user-1',
        createdAt: new Date().toISOString(),
      },
    ];
  }

  try {
    const response: AxiosResponse<SuccessResponse<ParticipantResponse[]>> =
      await boardServiceClient.get(`/participants/board/${boardId}`);
    return response.data.data || [];
  } catch (error) {
    console.error('getParticipants error:', error);
    throw error;
  }
};

/**
 * 보드에 참여자를 추가합니다.
 * [API] POST /api/participants
 */
export const addParticipant = async (data: AddParticipantRequest): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.post('/participants', data);
  } catch (error) {
    console.error('addParticipant error:', error);
    throw error;
  }
};

/**
 * 보드에서 참여자를 제거합니다.
 * [API] DELETE /api/participants/board/{boardId}/user/{userId}
 */
export const removeParticipant = async (boardId: string, userId: string): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.delete(`/participants/board/${boardId}/user/${userId}`);
  } catch (error) {
    console.error('removeParticipant error:', error);
    throw error;
  }
};

// ============================================================================
// 💡 [신규] 보드 이동 API (WebSocket 실시간 동기화용)
// ============================================================================

/**
 * 보드를 다른 컬럼으로 이동합니다 (실시간 반영).
 * [API] PUT /api/boards/{boardId}/move
 */
export const moveBoard = async (
  boardId: string,
  data: {
    projectId: string;
    groupByFieldName: string;
    newFieldValue?: string;
  },
): Promise<void> => {
  if (USE_MOCK_DATA) {
    return;
  }

  try {
    await boardServiceClient.put(`/${boardId}/move`, data);
  } catch (error) {
    console.error('moveBoard error:', error);
    throw error;
  }
};
export const addCommentToBoardWithFile = async (
  boardId: string,
  formData: FormData,
): Promise<CommentResponse> => {
  // Axios 요청 시 Content-Type을 'multipart/form-data'로 설정해야 합니다.
  // 보통 Axios는 FormData 객체를 사용하면 자동으로 Content-Type을 설정합니다.
  const response = await boardServiceClient.post<CommentResponse>(`/api/boards/${boardId}/comments`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};
