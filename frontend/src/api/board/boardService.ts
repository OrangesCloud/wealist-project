import { boardServiceClient } from '../apiConfig';
import { AxiosResponse } from 'axios';

import {
  ProjectResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  BoardResponse,
  CreateBoardRequest,
  UpdateBoardRequest,
  PaginatedBoardsResponse,
  MoveBoardRequest,
  MoveBoardResponse,
  FieldResponse,
  FieldOptionResponse,
  ViewResponse,
  UpdateBoardOrderRequest,
  CommentResponse,
  CreateCommentRequest,
  UpdateCommentRequest,
  UpdateViewRequest,
  CreateViewRequest,
  SetMultiSelectValueRequest,
  SetFieldValueRequest,
  UpdateFieldOptionRequest,
  CreateFieldOptionRequest,
  CreateFieldRequest,
  UpdateFieldRequest,
  ProjectInitSettingResponse,
} from '../../types/board'; // 💡 최신 타입 임포트

/**
 * ========================================
 * 목업 모드 전환
 * ========================================
 */
// ============================================================================
// 💡 [신규] 프로젝트 초기 데이터 로드 API
// ============================================================================

/**
 * 프로젝트 초기 페이지 로드에 필요한 모든 데이터를 조회합니다.
 * (보드, 필드/옵션, 필드 유형 목록 포함)
 * [API] GET /api/projects/{projectId}/init-settings
 */
export const getProjectInitSettings = async (
  projectId: string,
): Promise<ProjectInitSettingResponse> => {
  // ⚠️ Mock data는 삭제하고 실제 API 호출만 남깁니다. (Mock 환경은 USE_MOCK_DATA로 제어)
  try {
    const response: AxiosResponse<{ data: ProjectInitSettingResponse }> =
      await boardServiceClient.get(`/api/projects/${projectId}/init-settings`);
    return response.data.data;
  } catch (error) {
    console.error('getProjectInitData error:', error);
    throw error;
  }
};

// ============================================================================
// 프로젝트 관련 API
// ============================================================================

/**
 * 워크스페이스의 모든 프로젝트를 조회합니다.
 * GET /api/projects
 * @param workspaceId 워크스페이스 ID
 * @returns 프로젝트 배열
 */
// 💡 [수정] token 인자 제거
export const getProjects = async (workspaceId: string): Promise<ProjectResponse[]> => {
  try {
    // 💡 [수정 필요] API 명세에 따르면 PaginatedProjectsResponse DTO의 data 필드입니다.
    // 💡 response.data.data는 PaginatedProjectsResponse 타입이어야 합니다.
    const response: AxiosResponse<{ data: { projects: ProjectResponse[] } }> =
      await boardServiceClient.get('/api/projects', {
        params: { workspaceId },
      });

    // 💡 [핵심 수정] projects 배열을 직접 추출하여 반환합니다.
    return response?.data?.data?.projects || [];
  } catch (error) {
    console.error('getProjects error:', error);
    throw error;
  }
};
/**
 * 특정 프로젝트를 조회합니다.
 * GET /api/projects/{projectId}
 */
// 💡 [수정] token 인자 제거
export const getProject = async (projectId: string): Promise<ProjectResponse> => {
  try {
    const response: AxiosResponse<{ data: ProjectResponse }> = await boardServiceClient.get(
      `/api/projects/${projectId}`,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('getProject error:', error);
    throw error;
  }
};

/**
 * 새로운 프로젝트를 생성합니다.
 * POST /api/projects
 */
// 💡 [수정] token 인자 제거
export const createProject = async (data: CreateProjectRequest): Promise<ProjectResponse> => {
  try {
    const response: AxiosResponse<{ data: ProjectResponse }> = await boardServiceClient.post(
      '/api/projects',
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('createProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 업데이트합니다.
 * PUT /api/projects/{projectId}
 */
// 💡 [수정] token 인자 제거
export const updateProject = async (
  projectId: string,
  data: UpdateProjectRequest, // 💡 UpdateProjectRequest DTO 사용
): Promise<ProjectResponse> => {
  try {
    const response: AxiosResponse<{ data: ProjectResponse }> = await boardServiceClient.put(
      `/api/projects/${projectId}`,
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('updateProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 삭제합니다. (Soft Delete)
 * DELETE /api/projects/{projectId}
 */
// 💡 [수정] token 인자 제거
export const deleteProject = async (projectId: string): Promise<void> => {
  try {
    await boardServiceClient.delete(`/api/projects/${projectId}`, {});
  } catch (error) {
    console.error('deleteProject error:', error);
    throw error;
  }
};

/**
 * 프로젝트를 검색합니다.
 * GET /api/projects/search
 * @returns PaginatedProjectsResponse
 */
// 💡 [수정] token 인자 제거
export const searchProjects = async (
  workspaceId: string,
  query: string,
): Promise<PaginatedBoardsResponse> => {
  try {
    const response: AxiosResponse<{ data: PaginatedBoardsResponse }> = await boardServiceClient.get(
      '/api/projects/search',
      {
        params: { workspaceId, query },
      },
    );
    return response.data.data;
  } catch (error) {
    console.error('searchProjects error:', error);
    throw error;
  }
};

// ============================================================================
// 보드 관련 API
// ============================================================================

/**
 * 프로젝트의 보드를 조회합니다.
 * GET /api/boards
 * @returns 페이징된 보드 응답
 */
// 💡 [수정] token 인자 제거
export const getBoards = async (
  projectId: string,
  filters?: {
    stageId?: string;
    roleId?: string;
    importanceId?: string;
    assigneeId?: string;
    authorId?: string;
    page?: number;
    limit?: number;
  },
): Promise<PaginatedBoardsResponse> => {
  try {
    const params = { projectId, ...filters };
    const response: AxiosResponse<{ data: PaginatedBoardsResponse }> = await boardServiceClient.get(
      '/api/boards',
      {
        params,
      },
    );
    return response.data.data || { boards: [], total: 0, page: 1, limit: 20 };
  } catch (error) {
    console.error('getBoards error:', error);
    throw error;
  }
};

/**
 * 특정 보드를 조회합니다.
 * GET /api/boards/{boardId}
 */
// 💡 [수정] token 인자 제거
export const getBoard = async (boardId: string): Promise<BoardResponse> => {
  try {
    const response: AxiosResponse<{ data: BoardResponse }> = await boardServiceClient.get(
      `/api/boards/${boardId}`,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('getBoard error:', error);
    throw error;
  }
};

/**
 * 새로운 보드를 생성합니다.
 * POST /api/boards
 */
// 💡 [수정] token 인자 제거
export const createBoard = async (data: CreateBoardRequest): Promise<BoardResponse> => {
  try {
    const response: AxiosResponse<{ data: BoardResponse }> = await boardServiceClient.post(
      '/api/boards',
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('createBoard error:', error);
    throw error;
  }
};

/**
 * 보드를 업데이트합니다.
 * PUT /api/boards/{boardId}
 */
// 💡 [수정] token 인자 제거
export const updateBoard = async (
  boardId: string,
  data: UpdateBoardRequest,
): Promise<BoardResponse> => {
  try {
    const response: AxiosResponse<{ data: BoardResponse }> = await boardServiceClient.put(
      `/api/boards/${boardId}`,
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('updateBoard error:', error);
    throw error;
  }
};

/**
 * 보드를 삭제합니다.
 * DELETE /api/boards/{boardId}
 */
// 💡 [수정] token 인자 제거
export const deleteBoard = async (boardId: string): Promise<void> => {
  try {
    await boardServiceClient.delete(`/api/boards/${boardId}`, {});
  } catch (error) {
    console.error('deleteBoard error:', error);
    throw error;
  }
};

/**
 * 보드를 이동합니다.
 * PUT /api/boards/{boardId}/move
 */
// 💡 [수정] token 인자 제거
export const moveBoard = async (
  boardId: string,
  data: MoveBoardRequest,
): Promise<MoveBoardResponse> => {
  try {
    const response: AxiosResponse<{ data: MoveBoardResponse }> = await boardServiceClient.put(
      `/api/boards/${boardId}/move`,
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('moveBoard error:', error);
    throw error;
  }
};

// ============================================================================
// 커스텀 필드 관련 API
// ============================================================================

/**
 * 프로젝트의 모든 커스텀 필드를 조회합니다.
 * GET /api/projects/{projectId}/fields
 */
// 💡 [수정] token 인자 제거
export const getProjectFields = async (projectId: string): Promise<FieldResponse[]> => {
  try {
    const response: AxiosResponse<{ data: FieldResponse[] }> = await boardServiceClient.get(
      `/api/projects/${projectId}/fields`,
      {},
    );
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectFields error:', error);
    throw error;
  }
};

/**
 * 필드 옵션 목록 조회
 * GET /api/fields/{fieldId}/options
 */
// 💡 [수정] token 인자 제거
export const getFieldOptions = async (fieldId: string): Promise<FieldOptionResponse[]> => {
  try {
    const response: AxiosResponse<{ data: FieldOptionResponse[] }> = await boardServiceClient.get(
      `/api/fields/${fieldId}/options`,
      {},
    );
    return response.data.data || [];
  } catch (error) {
    console.error('getFieldOptions error:', error);
    throw error;
  }
};

// ============================================================================
// 필드 CRUD API
// ============================================================================

/**
 * 새 커스텀 필드를 생성합니다.
 * POST /api/fields
 */
// 💡 [수정] token 인자 제거
export const createField = async (data: CreateFieldRequest): Promise<FieldResponse> => {
  try {
    const response: AxiosResponse<{ data: FieldResponse }> = await boardServiceClient.post(
      '/api/fields',
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('createField error:', error);
    throw error;
  }
};

/**
 * 특정 필드를 조회합니다.
 * GET /api/fields/{fieldId}
 */
// 💡 [수정] token 인자 제거
export const getField = async (fieldId: string): Promise<FieldResponse> => {
  try {
    const response: AxiosResponse<{ data: FieldResponse }> = await boardServiceClient.get(
      `/api/fields/${fieldId}`,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('getField error:', error);
    throw error;
  }
};

/**
 * 필드를 수정합니다.
 * PATCH /api/fields/{fieldId}
 */
// 💡 [수정] token 인자 제거
export const updateField = async (
  fieldId: string,
  data: UpdateFieldRequest,
): Promise<FieldResponse> => {
  try {
    const response: AxiosResponse<{ data: FieldResponse }> = await boardServiceClient.patch(
      `/api/fields/${fieldId}`,
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('updateField error:', error);
    throw error;
  }
};

/**
 * 필드를 삭제합니다.
 * DELETE /api/fields/{fieldId}
 */
// 💡 [수정] token 인자 제거
export const deleteField = async (fieldId: string): Promise<void> => {
  try {
    await boardServiceClient.delete(`/api/fields/${fieldId}`, {});
  } catch (error) {
    console.error('deleteField error:', error);
    throw error;
  }
};

// ============================================================================
// 필드 옵션 CRUD API
// ============================================================================

/**
 * 필드 옵션을 생성합니다.
 * POST /api/field-options
 */
// 💡 [수정] token 인자 제거
export const createFieldOption = async (
  data: CreateFieldOptionRequest,
): Promise<FieldOptionResponse> => {
  try {
    const response: AxiosResponse<{ data: FieldOptionResponse }> = await boardServiceClient.post(
      '/api/field-options',
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('createFieldOption error:', error);
    throw error;
  }
};

/**
 * 필드 옵션을 수정합니다.
 * PATCH /api/field-options/{optionId}
 */
// 💡 [수정] token 인자 제거
export const updateFieldOption = async (
  optionId: string,
  data: UpdateFieldOptionRequest,
): Promise<FieldOptionResponse> => {
  try {
    const response: AxiosResponse<{ data: FieldOptionResponse }> = await boardServiceClient.patch(
      `/api/field-options/${optionId}`,
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('updateFieldOption error:', error);
    throw error;
  }
};

/**
 * 필드 옵션을 삭제합니다.
 * DELETE /api/field-options/{optionId}
 */
// 💡 [수정] token 인자 제거
export const deleteFieldOption = async (optionId: string): Promise<void> => {
  try {
    await boardServiceClient.delete(`/api/field-options/${optionId}`, {});
  } catch (error) {
    console.error('deleteFieldOption error:', error);
    throw error;
  }
};

// ============================================================================
// 보드 필드 값 관리 API
// ============================================================================

export interface BoardFieldValuesResponse {
  boardId: string;
  fields: Record<string, any>; // map[field_id]value
}

/**
 * 보드의 모든 필드 값을 조회합니다.
 * GET /api/boards/{boardId}/field-values
 */
// 💡 [수정] token 인자 제거 및 응답 DTO 변경
export const getBoardFieldValues = async (boardId: string): Promise<BoardFieldValuesResponse> => {
  try {
    const response: AxiosResponse<{ data: BoardFieldValuesResponse }> =
      await boardServiceClient.get(`/api/boards/${boardId}/field-values`, {});
    return response.data.data;
  } catch (error) {
    console.error('getBoardFieldValues error:', error);
    throw error;
  }
};

/**
 * 보드의 필드 값을 설정합니다.
 * POST /api/board-field-values
 */
// 💡 [수정] token 인자 제거
export const setFieldValue = async (data: SetFieldValueRequest): Promise<void> => {
  try {
    // API 명세: 204 No Content
    await boardServiceClient.post('/api/board-field-values', data, {});
  } catch (error) {
    console.error('setFieldValue error:', error);
    throw error;
  }
};

/**
 * 보드의 멀티 셀렉트 필드 값을 설정합니다.
 * POST /api/board-field-values/multi-select
 */
// 💡 [수정] token 인자 제거 및 DTO 변경
export const setMultiSelectValue = async (data: SetMultiSelectValueRequest): Promise<void> => {
  try {
    // API 명세: 204 No Content
    await boardServiceClient.post('/api/board-field-values/multi-select', data, {});
  } catch (error) {
    console.error('setMultiSelectValue error:', error);
    throw error;
  }
};

/**
 * 보드의 필드 값을 삭제합니다.
 * DELETE /api/boards/{boardId}/field-values/{fieldId}
 */
// 💡 [수정] token 인자 제거
export const deleteFieldValue = async (boardId: string, fieldId: string): Promise<void> => {
  try {
    // API 명세: 204 No Content
    await boardServiceClient.delete(`/api/boards/${boardId}/field-values/${fieldId}`, {});
  } catch (error) {
    console.error('deleteFieldValue error:', error);
    throw error;
  }
};

// ============================================================================
// 순서 관리 API
// ============================================================================

export interface FieldOrder {
  fieldId: string;
  displayOrder: number;
}
export interface UpdateFieldOrderRequest {
  fieldOrders: FieldOrder[];
}

export interface OptionOrder {
  optionId: string;
  displayOrder: number;
}
export interface UpdateOptionOrderRequest {
  optionOrders: OptionOrder[];
}

/**
 * 프로젝트 필드의 순서를 변경합니다.
 * PUT /api/projects/{projectId}/fields/order
 */
// 💡 [수정] token 인자 제거 및 DTO 변경
export const updateFieldOrder = async (
  projectId: string,
  data: UpdateFieldOrderRequest,
): Promise<void> => {
  try {
    // API 명세: 204 No Content
    await boardServiceClient.put(`/api/projects/${projectId}/fields/order`, data, {});
  } catch (error) {
    console.error('updateFieldOrder error:', error);
    throw error;
  }
};

/**
 * 필드 옵션의 순서를 변경합니다.
 * PUT /api/fields/{fieldId}/options/order
 */
// 💡 [수정] token 인자 제거 및 DTO 변경
export const updateOptionOrder = async (
  fieldId: string,
  data: UpdateOptionOrderRequest,
): Promise<void> => {
  try {
    // API 명세: 204 No Content
    await boardServiceClient.put(`/api/fields/${fieldId}/options/order`, data, {});
  } catch (error) {
    console.error('updateOptionOrder error:', error);
    throw error;
  }
};

// ============================================================================
// 댓글 관리 API (명세에 없으나 기존 코드에 존재하여 임시 유지)
// ============================================================================

/**
 * 보드의 모든 댓글을 조회합니다.
 * GET /api/comments (가정)
 */
// 💡 [수정] token 인자 제거
export const getComments = async (boardId: string): Promise<CommentResponse[]> => {
  try {
    // API 명세에는 없지만, 쿼리 파라미터로 boardId를 사용한다고 가정
    const response: AxiosResponse<{ data: CommentResponse[] }> = await boardServiceClient.get(
      '/api/comments',
      {
        params: { boardId },
      },
    );
    return response.data.data || [];
  } catch (error) {
    console.error('getComments error:', error);
    throw error;
  }
};

/**
 * 새 댓글을 생성합니다.
 * POST /api/comments (가정)
 */
// 💡 [수정] token 인자 제거
export const createComment = async (data: CreateCommentRequest): Promise<CommentResponse> => {
  try {
    const response: AxiosResponse<{ data: CommentResponse }> = await boardServiceClient.post(
      '/api/comments',
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('createComment error:', error);
    throw error;
  }
};

/**
 * 댓글을 수정합니다.
 * PUT /api/comments/{commentId} (가정)
 */
// 💡 [수정] token 인자 제거
export const updateComment = async (
  commentId: string,
  data: UpdateCommentRequest,
): Promise<CommentResponse> => {
  try {
    const response: AxiosResponse<{ data: CommentResponse }> = await boardServiceClient.put(
      `/api/comments/${commentId}`,
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('updateComment error:', error);
    throw error;
  }
};

/**
 * 댓글을 삭제합니다.
 * DELETE /api/comments/{commentId} (가정)
 */
// 💡 [수정] token 인자 제거
export const deleteComment = async (commentId: string): Promise<void> => {
  try {
    await boardServiceClient.delete(`/api/comments/${commentId}`, {});
  } catch (error) {
    console.error('deleteComment error:', error);
    throw error;
  }
};

// ============================================================================
// 뷰 관리 API
// ============================================================================

/**
 * 프로젝트별 뷰 목록 조회
 * GET /api/projects/{projectId}/views
 */
// 💡 [수정] token 인자 제거
export const getProjectViews = async (projectId: string): Promise<ViewResponse[]> => {
  try {
    const response: AxiosResponse<{ data: ViewResponse[] }> = await boardServiceClient.get(
      `/api/projects/${projectId}/views`,
      {},
    );
    return response.data.data || [];
  } catch (error) {
    console.error('getProjectViews error:', error);
    throw error;
  }
};

/**
 * 뷰를 생성합니다.
 * POST /api/views
 */
// 💡 [수정] token 인자 제거
export const createView = async (data: CreateViewRequest): Promise<ViewResponse> => {
  try {
    const response: AxiosResponse<{ data: ViewResponse }> = await boardServiceClient.post(
      '/api/views',
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('createView error:', error);
    throw error;
  }
};

/**
 * 특정 뷰를 조회합니다.
 * GET /api/views/{viewId}
 */
// 💡 [수정] token 인자 제거
export const getView = async (viewId: string): Promise<ViewResponse> => {
  try {
    const response: AxiosResponse<{ data: ViewResponse }> = await boardServiceClient.get(
      `/api/views/${viewId}`,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('getView error:', error);
    throw error;
  }
};

/**
 * 뷰를 수정합니다.
 * PATCH /api/views/{viewId}
 */
// 💡 [수정] token 인자 제거
export const updateView = async (
  viewId: string,
  data: UpdateViewRequest,
): Promise<ViewResponse> => {
  try {
    const response: AxiosResponse<{ data: ViewResponse }> = await boardServiceClient.patch(
      `/api/views/${viewId}`,
      data,
      {},
    );
    return response.data.data;
  } catch (error) {
    console.error('updateView error:', error);
    throw error;
  }
};

/**
 * 뷰를 삭제합니다.
 * DELETE /api/views/{viewId}
 */
// 💡 [수정] token 인자 제거
export const deleteView = async (viewId: string): Promise<void> => {
  try {
    // API 명세: 204 No Content
    await boardServiceClient.delete(`/api/views/${viewId}`, {});
  } catch (error) {
    console.error('deleteView error:', error);
    throw error;
  }
};

/**
 * 뷰 적용하여 보드 조회
 * GET /api/views/{viewId}/boards
 * @returns PaginatedBoardsResponse (가정)
 */
// 💡 [수정] token 인자 제거
export const getBoardsByView = async (
  viewId: string,
  filters?: {
    page?: number;
    limit?: number;
  },
): Promise<PaginatedBoardsResponse> => {
  try {
    const response: AxiosResponse<{ data: PaginatedBoardsResponse }> = await boardServiceClient.get(
      `/api/views/${viewId}/boards`,
      {
        params: filters,
      },
    );
    // 응답 data 필드에는 보드 데이터가 페이징되어 포함된다고 가정
    return response.data.data || { boards: [], total: 0, page: 1, limit: 20 };
  } catch (error) {
    console.error('getBoardsByView error:', error);
    throw error;
  }
};

/**
 * 뷰에서 보드 순서를 변경합니다.
 * PUT /api/view-board-orders
 */
// 💡 [수정] token 인자 제거
export const updateViewBoardOrder = async (data: UpdateBoardOrderRequest): Promise<void> => {
  try {
    // API 명세: 204 No Content
    await boardServiceClient.put('/api/view-board-orders', data, {});
  } catch (error) {
    console.error('updateViewBoardOrder error:', error);
    throw error;
  }
};
