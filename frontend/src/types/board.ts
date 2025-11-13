// src/types/board.ts

// =======================================================
// Board Service - 공통 DTO
// =======================================================

/**
 * @summary 사용자 정보 (dto.UserInfo)
 */
export interface UserInfo {
  userId: string;
  name: string;
  email: string;
  isActive: boolean;
}

/**
 * @summary 프로젝트 응답 DTO (dto.ProjectResponse)
 * [API: GET /api/projects]
 */
export interface ProjectResponse {
  projectId: string;
  workspaceId: string;
  name: string;
  description?: string;
  ownerId: string;
  ownerName: string;
  ownerEmail: string;
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
}

/**
 * @summary 프로젝트 수정 요청 (dto.UpdateProjectRequest)
 * [API: PUT /api/projects/{projectId}]
 */
export interface UpdateProjectRequest {
  name?: string;
  description?: string;
}

// =======================================================
// Board API 요청/응답 타입
// =======================================================

/**
 * @summary 보드 응답 DTO (dto.BoardResponse)
 * [API: GET /api/boards/{boardId}]
 */
export interface BoardResponse {
  boardId: string;
  title: string;
  content: string;
  projectId: string;
  position: string;
  dueDate: string;
  createdAt: string;
  updatedAt: string;
  author: UserInfo;
  assignee: UserInfo;
  /**
   * @description 파싱된 커스텀 필드 값들. 예: { stageId: "uuid", roleIds: ["uuid1", "uuid2"] }
   */
  customFields: Record<string, any>;
}

/**
 * @summary 보드 생성 요청 (dto.CreateBoardRequest)
 * [API: POST /api/boards]
 * @description 레거시 필드 (stageId, importanceId)는 DTO에 포함시키지만, customFields로 대체 권장
 */
export interface CreateBoardRequest {
  projectId: string;
  title: string;
  content?: string;
  assigneeId?: string; // 단일 사용자 ID
  dueDate?: string;
  // 레거시 필드 (백엔드 호환을 위해 유지)
  stageId?: string;
  importanceId?: string;
  roleIds?: string[]; // 멀티 셀렉트
}

/**
 * @summary 보드 수정 요청 (dto.UpdateBoardRequest)
 * [API: PUT /api/boards/{boardId}]
 */
export interface UpdateBoardRequest extends Partial<CreateBoardRequest> {}

/**
 * @summary 페이징된 보드 목록 응답 (dto.PaginatedBoardsResponse)
 * [API: GET /api/boards]
 */
export interface PaginatedBoardsResponse {
  boards: BoardResponse[];
  total: number;
  page: number;
  limit: number;
}

// =======================================================
// Custom Field API 응답/요청 타입
// =======================================================

/**
 * @summary 커스텀 필드 응답 DTO (dto.FieldResponse)
 */
export interface FieldResponse {
  fieldId: string;
  projectId: string;
  name: string;
  description: string;
  fieldType:
    | 'text'
    | 'number'
    | 'single_select'
    | 'multi_select'
    | 'date'
    | 'datetime'
    | 'single_user'
    | 'multi_user'
    | 'checkbox'
    | 'url';
  isRequired: boolean;
  isSystemDefault: boolean;
  displayOrder: number;
  config: Record<string, any>;
}

/**
 * @summary 필드 옵션 응답 DTO (dto.OptionResponse)
 * @description Stage, Role, Importance 등의 선택지를 포함
 */
export interface FieldOptionResponse {
  optionId: string;
  fieldId: string;
  label: string;
  description: string;
  color: string;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}

// 💡 Mock Data 호환성 및 프론트엔드 LookUp용 타입 (FieldOptionResponse 기반)

export interface BaseFieldOption {
  label: string;
  color: string;
  displayOrder: number;
  level?: number;
  fieldId: string; // 소속 필드 ID
  isSystemDefault: boolean;
  description: string;
}

/**
 * @summary Stage 옵션 타입 (프론트엔드 LookUp용 - optionId를 stageId로 사용)
 */
export interface CustomStageResponse extends BaseFieldOption {
  stageId: string; // FieldOptionResponse.optionId와 동일
}

/**
 * @summary Role 옵션 타입 (프론트엔드 LookUp용 - optionId를 roleId로 사용)
 */
export interface CustomRoleResponse extends BaseFieldOption {
  roleId: string; // FieldOptionResponse.optionId와 동일
}

/**
 * @summary Importance 옵션 타입 (프론트엔드 LookUp용 - optionId를 importanceId로 사용)
 */
export interface CustomImportanceResponse extends BaseFieldOption {
  importanceId: string; // FieldOptionResponse.optionId와 동일
}

// =======================================================
// 그 외 API 요청/응답 타입
// =======================================================

/**
 * @summary 보드 이동 요청 (dto.MoveBoardRequest)
 * [API: PUT /api/boards/{boardId}/move]
 */
export interface MoveBoardRequest {
  viewId: string;
  groupByFieldId: string;
  newFieldValue: string; // 새로운 필드 옵션 ID (예: Stage ID)
  beforePosition?: string;
  afterPosition?: string;
}

/**
 * @summary 보드 이동 응답 (dto.MoveBoardResponse)
 */
export interface MoveBoardResponse {
  boardId: string;
  newFieldValue: string;
  newPosition: string;
  message: string;
}

/**
 * @summary 뷰 응답 DTO (dto.ViewResponse)
 */
export interface ViewResponse {
  viewId: string;
  projectId: string;
  name: string;
  description?: string;
  isDefault: boolean;
  isShared: boolean;
  filters: Record<string, any>; // 필터 조건
  groupByFieldId?: string;
  sortBy?: string;
  sortDirection?: 'asc' | 'desc';
  createdAt: string;
  updatedAt: string;
}

/**
 * @summary 뷰 순서 변경 요청 (dto.UpdateBoardOrderRequest)
 * [API: PUT /api/view-board-orders]
 */
export interface UpdateBoardOrderRequest {
  viewId: string;
  boardOrders: Array<{ boardId: string; position: string }>;
}

/**
 * @summary 댓글 응답 (가정된 DTO)
 */
export interface CommentResponse {
  commentId: string;
  boardId: string;
  userId: string;
  userName: string;
  userAvatar?: string;
  content: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * @summary 댓글 생성 요청 (가정된 DTO)
 */
export interface CreateCommentRequest {
  boardId: string;
  content: string;
}

/**
 * @summary 댓글 수정 요청 (가정된 DTO)
 */
export interface UpdateCommentRequest {
  content: string;
}

// 💡 기존의 프론트엔드 컴포넌트에서 사용하던 타입은 BoardResponse로 대체하거나,
//    필요에 따라 BoardResponse를 확장하여 사용합니다.
export type Priority = 'HIGH' | 'MEDIUM' | 'LOW' | '';
// 💡 [통합된 View/Filter 상태 인터페이스]
export interface Column {
  stageId: string;
  title: string;
  color?: string;
  boards: BoardResponse[];
}
export type TLayout = 'table' | 'board' | undefined;
export type TView = 'stage' | 'role' | 'importance' | undefined;

export interface ViewState {
  currentView?: TView;
  searchQuery?: string;
  filterOption?: string;
  currentLayout?: TLayout;
  showCompleted?: boolean;
  sortColumn?:
    | 'title'
    | 'stage'
    | 'role'
    | 'importance'
    | 'importance'
    | 'assignee'
    | 'dueDate'
    | null;
  sortDirection?: 'asc' | 'desc';
}
// --- 필드/옵션 요청 DTO ---

// 💡 [추가] 룩업 데이터 인터페이스
export interface FieldOptionsLookup {
  stages?: CustomStageResponse[];
  roles?: CustomRoleResponse[];
  importances?: CustomImportanceResponse[];
}

/**
 * @summary 필드 옵션 생성 요청 (dto.CreateOptionRequest)
 * [API: POST /api/field-options]
 */
export interface CreateFieldOptionRequest {
  fieldId: string;
  label: string;
  description?: string;
  color?: string;
}

/**
 * @summary 필드 옵션 수정 요청 (dto.UpdateOptionRequest)
 * [API: PATCH /api/field-options/{optionId}]
 */
export interface UpdateFieldOptionRequest {
  label?: string;
  description?: string;
  color?: string;
}

// =======================================================
// 프로젝트 초기화 데이터 DTO (Init Data)
// =======================================================

/**
 * @summary 필드 유형 정보 (dto.FieldTypeInfo)
 */
export interface FieldTypeInfo {
  type: string;
  displayName: string;
  description: string;
  hasOptions: boolean;
}

/**
 * @summary 필드 및 옵션 정보 통합 (dto.FieldWithOptionsResponse)
 */
export interface FieldWithOptionsResponse extends FieldResponse {
  options: FieldOptionResponse[];
}

/**
 * @summary 프로젝트 초기화 응답 DTO (dto.ProjectInitSettingResponse)
 * [API: GET /api/projects/{projectId}/init-settings]
 */
export interface ProjectInitSettingResponse {
  project: ProjectResponse; // Project Basic Info
  fields: FieldWithOptionsResponse[]; // 모든 필드 정의와 그 옵션
  fieldTypes: FieldTypeInfo[]; // 사용 가능한 필드 유형 목록
  defaultViewId?: string;
}

/**
 * @summary 새 커스텀 필드 생성 요청 (dto.CreateFieldRequest)
 * [API: POST /api/fields]
 */
export interface CreateFieldRequest {
  projectId: string;
  name: string;
  description?: string;
  fieldType: // FieldType 정의
  | 'text'
    | 'number'
    | 'single_select'
    | 'multi_select'
    | 'date'
    | 'datetime'
    | 'single_user'
    | 'multi_user'
    | 'checkbox'
    | 'url';
  isRequired?: boolean;
  config?: Record<string, any>;
}

/**
 * @summary 필드 수정 요청 (dto.UpdateFieldRequest)
 * [API: PATCH /api/fields/{fieldId}]
 */
export interface UpdateFieldRequest {
  name?: string;
  description?: string;
  displayOrder?: number;
  isRequired?: boolean;
  config?: Record<string, any>;
}

// --- 보드 필드 값 요청 DTO ---

/**
 * @summary 보드의 필드 값 설정 요청 (dto.SetFieldValueRequest)
 * [API: POST /api/board-field-values]
 */
export interface SetFieldValueRequest {
  boardId: string;
  fieldId: string;
  value: any; // Type depends on field type
}

/**
 * @summary 보드의 멀티 셀렉트 필드 값 설정 요청 (dto.SetMultiSelectValueRequest)
 * [API: POST /api/board-field-values/multi-select]
 */
export interface SetMultiSelectValueRequest {
  boardId: string;
  fieldId: string;
  values: Array<{ valueId: string; displayOrder: number }>;
}

// --- 뷰 요청 DTO ---

/**
 * @summary 뷰 생성 요청 (dto.CreateViewRequest)
 * [API: POST /api/views]
 */
export interface CreateViewRequest {
  projectId: string;
  name: string;
  description?: string;
  isDefault?: boolean;
  isShared?: boolean;
  filters?: Record<string, any>;
  groupByFieldId?: string;
  sortBy?: string;
  sortDirection?: 'asc' | 'desc';
}

/**
 * @summary 뷰 수정 요청 (dto.UpdateViewRequest)
 * [API: PATCH /api/views/{viewId}]
 */
export interface UpdateViewRequest {
  name?: string;
  description?: string;
  isDefault?: boolean;
  isShared?: boolean;
  filters?: Record<string, any>;
  groupByFieldId?: string;
  sortBy?: string;
  sortDirection?: 'asc' | 'desc';
}
