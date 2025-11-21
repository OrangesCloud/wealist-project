// src/types/board.ts

// =======================================================
// Common Types
// =======================================================

/**
 * @summary 공통 응답 래퍼 (response.SuccessResponse)
 */
export interface SuccessResponse<T = any> {
  data: T;
  requestId: string;
}

/**
 * @summary 공통 에러 응답 (response.ErrorResponse)
 */
export interface ErrorResponse {
  error: any;
  requestId: string;
}

// =======================================================
// Board Types
// =======================================================

/**
 * @summary 보드 응답 DTO (dto.BoardResponse)
 * [API: GET /api/boards/{boardId}, POST /api/boards, PUT /api/boards/{boardId}]
 */
export interface BoardResponse {
  boardId: string;
  projectId: string;
  title: string;
  content: string;
  assigneeId: string;
  authorId: string;
  dueDate: string;
  customFields: Record<string, any>;
  createdAt: string;
  updatedAt: string;
  fileUrl?: string; // 💡 [추가] 보드 파일 URL 설정
  fileName?: string; // 💡 [추가] 보드 파일 이름 설정
}

/**
 * @summary 보드 상세 응답 DTO (dto.BoardDetailResponse)
 * [API: GET /api/boards/{boardId}]
 */
export interface BoardDetailResponse extends BoardResponse {
  participants: ParticipantResponse[];
  comments: CommentResponse[];
}

/**
 * @summary 보드 생성 요청 (dto.CreateBoardRequest)
 * [API: POST /api/boards]
 */
export interface CreateBoardRequest {
  projectId: string;
  title: string;
  content?: string;
  assigneeId?: string;
  startDate?: string;
  dueDate?: string;
  customFields?: Record<string, any>;
  participants?: string[];
  fileUrl?: string; // 💡 [추가] 보드 파일 URL 설정
  fileName?: string; // 💡 [추가] 보드 파일 이름 설정
}

/**
 * @summary 보드 수정 요청 (dto.UpdateBoardRequest)
 * [API: PUT /api/boards/{boardId}]
 */
export interface UpdateBoardRequest {
  title?: string;
  content?: string;
  assigneeId?: string;
  startDate?: string;
  dueDate?: string;
  customFields?: Record<string, any>;
  participants?: string[];
  fileUrl?: string; // 💡 [추가] 보드 파일 URL 설정
  fileName?: string; // 💡 [추가] 보드 파일 이름 설정
}

/**
 * @summary 보드 필터 (dto.BoardFilters)
 */
export interface BoardFilters {
  customFields?: Record<string, any>;
}

/**
 * @summary 페이징된 보드 목록 (dto.PaginatedBoardsResponse)
 */
export interface PaginatedBoardsResponse {
  boards: BoardResponse[];
  total: number;
  page: number;
  limit: number;
}

/**
 * @summary 보드 필드 수정 요청 (dto.UpdateBoardFieldRequest)
 */
export interface UpdateBoardFieldRequest {
  fieldId: 'stage' | 'importance' | 'role';
  value: string;
}

// =======================================================
// Project Types
// =======================================================

/**
 * @summary 프로젝트 응답 DTO (dto.ProjectResponse)
 * [API: GET /api/projects/{projectId}, POST /api/projects, PUT /api/projects/{projectId}]
 */
export interface ProjectResponse {
  projectId: string;
  workspaceId: string;
  name: string;
  description: string;
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
  dueDate?: string;
}

/**
 * @summary 프로젝트 기본 정보 (dto.ProjectBasicInfo)
 */
export interface ProjectBasicInfo {
  projectId: string;
  workspaceId: string;
  workspaceName: string;
  workspaceEmail: string;
  name: string;
  description: string;
  ownerId: string;
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * @summary 프로젝트 생성 요청 (dto.CreateProjectRequest)
 * [API: POST /api/projects]
 */
export interface CreateProjectRequest {
  workspaceId: string;
  name: string;
  description?: string;
  dueDate?: string;
}

/**
 * @summary 프로젝트 수정 요청 (dto.UpdateProjectRequest)
 * [API: PUT /api/projects/{projectId}]
 */
export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  dueDate?: string;
}

/**
 * @summary 페이징된 프로젝트 목록 (dto.PaginatedProjectsResponse)
 * [API: GET /api/projects/search]
 */
export interface PaginatedProjectsResponse {
  projects: ProjectResponse[];
  total: number;
  page: number;
  limit: number;
}

// =======================================================
// Project Init Settings Types
// =======================================================

/**
 * @summary 필드 타입 정보 (dto.FieldTypeInfo)
 */
export interface FieldTypeInfo {
  typeId: string;
  typeName: string;
  description: string;
}

/**
 * @summary 필드 옵션 (dto.FieldOption)
 */
export interface FieldOption {
  optionId: string;
  optionValue: string;
  optionLabel: string;
  color?: string;
}

/**
 * @summary 옵션이 포함된 필드 응답 (dto.FieldWithOptionsResponse)
 */
export interface FieldWithOptionsResponse {
  fieldId: string;
  fieldName: string;
  fieldType: string;
  description: string;
  isRequired: boolean;
  options: FieldOption[];
}

/**
 * @summary 프로젝트 초기 설정 응답 (dto.ProjectInitSettingsResponse)
 * [API: GET /api/projects/{projectId}/init-settings]
 */
export interface ProjectInitSettingsResponse {
  project: ProjectBasicInfo;
  fields: FieldWithOptionsResponse[];
  fieldTypes: FieldTypeInfo[];
  defaultViewId: string;
}

// =======================================================
// Field Option Types
// =======================================================

/**
 * @summary 필드 옵션 응답 (dto.FieldOptionResponse)
 * [API: GET /api/field-options, POST /api/field-options, PATCH /api/field-options/{optionId}]
 */
export interface FieldOptionResponse {
  optionId: string;
  fieldType: string;
  value: string;
  label: string;
  color: string;
  displayOrder: number;
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * @summary 필드 옵션 생성 요청 (dto.CreateFieldOptionRequest)
 * [API: POST /api/field-options]
 */
export interface CreateFieldOptionRequest {
  fieldType: 'stage' | 'role' | 'importance';
  value: string;
  label: string;
  color: string;
  displayOrder?: number;
}

/**
 * @summary 필드 옵션 수정 요청 (dto.UpdateFieldOptionRequest)
 * [API: PATCH /api/field-options/{optionId}]
 */
export interface UpdateFieldOptionRequest {
  label?: string;
  color?: string;
  displayOrder?: number;
}

// =======================================================
// Project Member Types
// =======================================================

/**
 * @summary 프로젝트 멤버 응답 (dto.ProjectMemberResponse)
 * [API: GET /api/projects/{projectId}/members]
 */
export interface ProjectMemberResponse {
  memberId: string;
  projectId: string;
  userId: string;
  userName: string;
  userEmail: string;
  roleName: string;
  joinedAt: string;
}

/**
 * @summary 멤버 역할 변경 요청 (dto.UpdateProjectMemberRoleRequest)
 * [API: PUT /api/projects/{projectId}/members/{memberId}/role]
 */
export interface UpdateProjectMemberRoleRequest {
  roleName: 'OWNER' | 'ADMIN' | 'MEMBER';
}

// =======================================================
// Project Join Request Types
// =======================================================

/**
 * @summary 프로젝트 가입 요청 응답 (dto.ProjectJoinRequestResponse)
 * [API: GET /api/projects/{projectId}/join-requests, POST /api/join-requests]
 */
export interface ProjectJoinRequestResponse {
  requestId: string;
  projectId: string;
  userId: string;
  userName: string;
  userEmail: string;
  status: string;
  requestedAt: string;
  updatedAt: string;
}

/**
 * @summary 프로젝트 가입 요청 생성 (dto.CreateProjectJoinRequestRequest)
 * [API: POST /api/join-requests]
 */
export interface CreateProjectJoinRequestRequest {
  projectId: string;
}

/**
 * @summary 프로젝트 가입 요청 상태 변경 (dto.UpdateProjectJoinRequestRequest)
 * [API: PUT /api/join-requests/{joinRequestId}]
 */
export interface UpdateProjectJoinRequestRequest {
  status: 'APPROVED' | 'REJECTED';
}

// =======================================================
// Comment Types
// =======================================================

/**
 * @summary 댓글 응답 DTO (dto.CommentResponse)
 * [API: GET /api/comments/board/{boardId}, POST /api/comments, PUT /api/comments/{commentId}]
 */
export interface CommentResponse {
  commentId: string;
  boardId: string;
  userId: string;
  content: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * @summary 댓글 생성 요청 (dto.CreateCommentRequest)
 * [API: POST /api/comments]
 */
export interface CreateCommentRequest {
  boardId: string;
  content: string;
}

/**
 * @summary 댓글 수정 요청 (dto.UpdateCommentRequest)
 * [API: PUT /api/comments/{commentId}]
 */
export interface UpdateCommentRequest {
  content: string;
}

// =======================================================
// Participant Types
// =======================================================

/**
 * @summary 참여자 응답 DTO (dto.ParticipantResponse)
 * [API: GET /api/participants/board/{boardId}]
 */
export interface ParticipantResponse {
  id: string;
  boardId: string;
  userId: string;
  createdAt: string;
}

/**
 * @summary 참여자 추가 요청 (dto.AddParticipantRequest)
 * [API: POST /api/participants]
 */
export interface AddParticipantRequest {
  boardId: string;
  userId: string;
}

// =======================================================
// Frontend Utility Types
// =======================================================

export type Priority = 'HIGH' | 'MEDIUM' | 'LOW' | '';
export type TLayout = 'table' | 'board' | undefined;
export type TView = 'stage' | 'role' | 'importance' | undefined;

export interface Column {
  stageId: string;
  title: string;
  color?: string;
  boards: BoardResponse[];
}

export interface ViewState {
  currentView?: TView;
  searchQuery?: string;
  filterOption?: string;
  currentLayout?: TLayout;
  showCompleted?: boolean;
  sortColumn?: 'title' | 'stage' | 'role' | 'importance' | 'assignee' | 'dueDate' | null;
  sortDirection?: 'asc' | 'desc';
}

export interface FieldOptionsLookup {
  [fieldId: string]: FieldOption[] | undefined;
}

export interface IEditCustomFields {
  name: string;
  fieldType:
    | 'text'
    | 'number'
    | 'single_select'
    | 'multi_select'
    | 'date'
    | 'single_user'
    | 'multi_user';
  options?: FieldOption[];
  value?: string | number | null;
}
