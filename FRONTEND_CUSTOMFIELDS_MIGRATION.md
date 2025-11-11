# 프론트엔드 CustomFields 마이그레이션 가이드

## 📋 목차
1. [현재 상황](#현재-상황)
2. [백엔드 데이터 구조](#백엔드-데이터-구조)
3. [프론트엔드 현재 코드 상황](#프론트엔드-현재-코드-상황)
4. [수정해야 할 작업 목록](#수정해야-할-작업-목록)
5. [상세 구현 가이드](#상세-구현-가이드)
6. [테스트 체크리스트](#테스트-체크리스트)

---

## 현재 상황

### ⚠️ 문제점
- **백엔드**: `customFields` 기반 (통합된 커스텀 필드 시스템)
- **프론트엔드**: 레거시 `stage`, `roles`, `importance` 필드 사용
- **결과**: 프론트엔드가 백엔드 응답을 제대로 처리하지 못함

### ✅ 백엔드 완료 작업
1. 프로젝트 생성 시 자동으로 기본 필드 생성 (Stage, Role, Importance)
2. BoardResponse에 `customFields` 포함
3. 기본 필드 옵션:
   - **Stage**: 대기, 진행중, 완료
   - **Role**: 프론트엔드, 백엔드, 디자인
   - **Importance**: 낮음, 보통, 높음

---

## 백엔드 데이터 구조

### 1. Field (프로젝트 필드 정의)

**API**: `GET /api/projects/{projectId}/fields`

**응답 예시**:
```json
{
  "data": [
    {
      "fieldId": "550e8400-e29b-41d4-a716-446655440001",
      "projectId": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Stage",
      "fieldType": "single_select",
      "description": "작업 진행 단계",
      "displayOrder": 0,
      "isRequired": true,
      "isSystemDefault": true,
      "config": {},
      "canEditRoles": null,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    },
    {
      "fieldId": "550e8400-e29b-41d4-a716-446655440002",
      "projectId": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Role",
      "fieldType": "single_select",
      "description": "담당 역할",
      "displayOrder": 1,
      "isRequired": false,
      "isSystemDefault": true,
      "config": {},
      "canEditRoles": null,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    },
    {
      "fieldId": "550e8400-e29b-41d4-a716-446655440003",
      "projectId": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Importance",
      "fieldType": "single_select",
      "description": "작업 중요도",
      "displayOrder": 2,
      "isRequired": false,
      "isSystemDefault": true,
      "config": {},
      "canEditRoles": null,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 2. Field Options (필드 옵션)

**API**: `GET /api/fields/{fieldId}/options`

**응답 예시** (Stage 필드의 옵션들):
```json
{
  "data": [
    {
      "optionId": "650e8400-e29b-41d4-a716-446655440001",
      "fieldId": "550e8400-e29b-41d4-a716-446655440001",
      "label": "대기",
      "color": "#F59E0B",
      "description": "",
      "displayOrder": 0,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    },
    {
      "optionId": "650e8400-e29b-41d4-a716-446655440002",
      "fieldId": "550e8400-e29b-41d4-a716-446655440001",
      "label": "진행중",
      "color": "#3B82F6",
      "description": "",
      "displayOrder": 1,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    },
    {
      "optionId": "650e8400-e29b-41d4-a716-446655440003",
      "fieldId": "550e8400-e29b-41d4-a716-446655440001",
      "label": "완료",
      "color": "#10B981",
      "description": "",
      "displayOrder": 2,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 3. Board Response (보드 데이터)

**API**: `GET /api/boards?projectId={projectId}`

**응답 예시**:
```json
{
  "data": {
    "boards": [
      {
        "boardId": "750e8400-e29b-41d4-a716-446655440001",
        "projectId": "550e8400-e29b-41d4-a716-446655440000",
        "title": "로그인 페이지 구현",
        "content": "JWT 인증 방식으로 로그인/로그아웃 기능 구현",
        "assignee": {
          "userId": "850e8400-e29b-41d4-a716-446655440001",
          "name": "김개발",
          "email": "dev@example.com",
          "isActive": true
        },
        "author": {
          "userId": "850e8400-e29b-41d4-a716-446655440001",
          "name": "김개발",
          "email": "dev@example.com",
          "isActive": true
        },
        "dueDate": "2024-02-15T00:00:00Z",
        "createdAt": "2024-01-20T00:00:00Z",
        "updatedAt": "2024-01-25T00:00:00Z",
        "customFields": {
          "550e8400-e29b-41d4-a716-446655440001": "650e8400-e29b-41d4-a716-446655440002",
          "550e8400-e29b-41d4-a716-446655440002": "650e8400-e29b-41d4-a716-446655440004",
          "550e8400-e29b-41d4-a716-446655440003": "650e8400-e29b-41d4-a716-446655440007"
        },
        "position": "0|hzzzzz:"
      }
    ],
    "total": 10,
    "page": 1,
    "limit": 20
  }
}
```

**customFields 구조 설명**:
```javascript
{
  "[field-id]": "[option-id]",  // single_select 필드의 경우
  "[field-id]": ["[option-id-1]", "[option-id-2]"],  // multi_select 필드의 경우
  "[field-id]": "text value",  // text 필드의 경우
  "[field-id]": 42,  // number 필드의 경우
}
```

현재 기본 필드는 모두 `single_select` 타입이므로:
- `customFields[stageFieldId]` → stageOptionId
- `customFields[roleFieldId]` → roleOptionId
- `customFields[importanceFieldId]` → importanceOptionId

---

## 프론트엔드 현재 코드 상황

### 1. 레거시 필드 사용 위치

#### **Dashboard.tsx** (`frontend/src/pages/Dashboard.tsx`)

**라인 335-337**: 보드를 Stage 컬럼에 매핑
```typescript
const stageId = board.stage?.stage_id;  // ❌ board.stage는 undefined
if (stageId && stageMap.has(stageId)) {
  stageMap.get(stageId)!.boards.push(board);
}
```

**라인 343**: Stage displayOrder로 정렬
```typescript
(a, b) => a.stage.displayOrder - b.stage.displayOrder,  // ❌ a.stage는 undefined
```

**라인 441**: Drag & Drop 시 Stage 업데이트
```typescript
stage: { ...draggedBoard.stage!, id: targetColumnId },  // ❌ draggedBoard.stage는 undefined
```

**라인 860-861**: Role로 정렬
```typescript
aValue = a.roles?.[0]?.name?.toLowerCase() || '';  // ❌ a.roles는 undefined
bValue = b.roles?.[0]?.name?.toLowerCase() || '';
```

**라인 864-865**: Importance로 정렬
```typescript
aValue = a.importance?.level || 0;  // ❌ a.importance는 undefined
bValue = b.importance?.level || 0;
```

**라인 904-910**: Role 표시
```typescript
{board.roles && board.roles.length > 0 ? (  // ❌ board.roles는 undefined
  <div className="flex items-center gap-1">
    <div style={{ backgroundColor: board.roles[0].color || '#6B7280' }} />
    <span>{board.roles[0].name}</span>
  </div>
) : null}
```

**라인 918-924**: Importance 표시
```typescript
{board.importance ? (  // ❌ board.importance는 undefined
  <div className="flex items-center gap-1">
    <div style={{ backgroundColor: board.importance.color || '#6B7280' }} />
    <span>{board.importance.name}</span>
  </div>
) : null}
```

#### **CreateBoardModal.tsx** (`frontend/src/components/modals/CreateBoardModal.tsx`)

**라인 53-55**: 레거시 필드 State
```typescript
const [selectedStageId, setSelectedStageId] = useState(editData?.stageId || initialStageId || '');
const [selectedRoleId, setSelectedRoleId] = useState<string>(editData?.roleId || '');
const [selectedImportanceId, setSelectedImportanceId] = useState<string>(editData?.importanceId || '');
```

**라인 91-95**: 레거시 API 호출
```typescript
const [stagesData, rolesData, importancesData] = await Promise.all([
  getProjectStages(projectId, accessToken),  // ❌ 이제 getProjectFields 사용해야 함
  getProjectRoles(projectId, accessToken),
  getProjectImportances(projectId, accessToken),
]);
```

#### **BoardDetailModal.tsx** (`frontend/src/components/modals/BoardDetailModal.tsx`)

**라인 97-100**: 레거시 필드에서 값 추출
```typescript
setSelectedStageId(boardData.stage?.id || '');  // ❌ boardData.stage는 undefined
setSelectedRoleId(boardData.roles?.[0]?.id || '');  // ❌ boardData.roles는 undefined
setSelectedImportanceId(boardData.importance?.id || '');  // ❌ boardData.importance는 undefined
```

---

## 수정해야 할 작업 목록

### ✅ Phase 1: API 레이어 수정 (frontend/src/api/board/boardService.ts)

#### 1.1 새로운 타입 정의 추가
```typescript
// Field 관련 타입
export interface FieldResponse {
  fieldId: string;
  projectId: string;
  name: string;
  fieldType: 'text' | 'number' | 'single_select' | 'multi_select' | 'date' | 'datetime' | 'single_user' | 'multi_user' | 'checkbox' | 'url';
  description: string;
  displayOrder: number;
  isRequired: boolean;
  isSystemDefault: boolean;
  config: Record<string, any>;
  canEditRoles: string[] | null;
  createdAt: string;
  updatedAt: string;
}

export interface OptionResponse {
  optionId: string;
  fieldId: string;
  label: string;
  color: string;
  description: string;
  displayOrder: number;
  createdAt: string;
  updatedAt: string;
}

// 편의를 위한 파싱된 필드 타입
export interface ParsedField {
  field: FieldResponse;
  options: OptionResponse[];
}

export interface ProjectFieldsResponse {
  fields: ParsedField[];
  stageField: ParsedField | null;
  roleField: ParsedField | null;
  importanceField: ParsedField | null;
}
```

#### 1.2 필드 조회 함수 추가
```typescript
/**
 * 프로젝트의 모든 필드를 조회하고 옵션까지 함께 가져옵니다.
 * GET /api/projects/{projectId}/fields
 * GET /api/fields/{fieldId}/options (각 필드마다)
 */
export const getProjectFieldsWithOptions = async (
  projectId: string,
  token: string,
): Promise<ProjectFieldsResponse> => {
  try {
    // 1. 모든 필드 조회
    const fieldsResponse = await boardService.get(`/api/projects/${projectId}/fields`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const fields: FieldResponse[] = fieldsResponse.data.data || [];

    // 2. 각 필드의 옵션 조회 (병렬 처리)
    const parsedFields: ParsedField[] = await Promise.all(
      fields.map(async (field) => {
        if (field.fieldType === 'single_select' || field.fieldType === 'multi_select') {
          const optionsResponse = await boardService.get(`/api/fields/${field.fieldId}/options`, {
            headers: { Authorization: `Bearer ${token}` },
          });
          return {
            field,
            options: (optionsResponse.data.data || []).sort(
              (a: OptionResponse, b: OptionResponse) => a.displayOrder - b.displayOrder
            ),
          };
        }
        return { field, options: [] };
      })
    );

    // 3. 기본 필드 식별 (이름으로 매칭)
    const stageField = parsedFields.find(pf => pf.field.name === 'Stage') || null;
    const roleField = parsedFields.find(pf => pf.field.name === 'Role') || null;
    const importanceField = parsedFields.find(pf => pf.field.name === 'Importance') || null;

    return {
      fields: parsedFields,
      stageField,
      roleField,
      importanceField,
    };
  } catch (error) {
    console.error('getProjectFieldsWithOptions error:', error);
    throw error;
  }
};
```

#### 1.3 레거시 함수 Deprecated 표시
```typescript
/**
 * @deprecated Use getProjectFieldsWithOptions instead
 */
export const getProjectStages = async (project_id: string, token: string) => {
  // ... 기존 코드
};

/**
 * @deprecated Use getProjectFieldsWithOptions instead
 */
export const getProjectRoles = async (project_id: string, token: string) => {
  // ... 기존 코드
};

/**
 * @deprecated Use getProjectFieldsWithOptions instead
 */
export const getProjectImportances = async (project_id: string, token: string) => {
  // ... 기존 코드
};
```

---

### ✅ Phase 2: 유틸리티 함수 작성 (frontend/src/utils/customFields.ts - 새 파일)

```typescript
import { BoardResponse, FieldResponse, OptionResponse, ParsedField } from '../api/board/boardService';

/**
 * customFields에서 특정 필드의 옵션 ID를 추출합니다.
 */
export function getFieldOptionId(
  board: BoardResponse,
  fieldId: string | undefined
): string | null {
  if (!fieldId || !board.customFields) return null;
  const value = board.customFields[fieldId];
  return typeof value === 'string' ? value : null;
}

/**
 * customFields와 필드 정보를 기반으로 옵션 객체를 반환합니다.
 */
export function getFieldOption(
  board: BoardResponse,
  field: ParsedField | null
): OptionResponse | null {
  if (!field) return null;
  const optionId = getFieldOptionId(board, field.field.fieldId);
  if (!optionId) return null;
  return field.options.find(opt => opt.optionId === optionId) || null;
}

/**
 * Stage 정보 추출 (하위 호환성)
 */
export function getStageInfo(
  board: BoardResponse,
  stageField: ParsedField | null
): { id: string; name: string; color: string; displayOrder: number } | null {
  const option = getFieldOption(board, stageField);
  if (!option) return null;
  return {
    id: option.optionId,
    name: option.label,
    color: option.color,
    displayOrder: option.displayOrder,
  };
}

/**
 * Role 정보 추출 (하위 호환성)
 */
export function getRoleInfo(
  board: BoardResponse,
  roleField: ParsedField | null
): { id: string; name: string; color: string } | null {
  const option = getFieldOption(board, roleField);
  if (!option) return null;
  return {
    id: option.optionId,
    name: option.label,
    color: option.color,
  };
}

/**
 * Importance 정보 추출 (하위 호환성)
 */
export function getImportanceInfo(
  board: BoardResponse,
  importanceField: ParsedField | null
): { id: string; name: string; color: string; level: number } | null {
  const option = getFieldOption(board, importanceField);
  if (!option) return null;
  return {
    id: option.optionId,
    name: option.label,
    color: option.color,
    level: option.displayOrder, // displayOrder를 level로 사용
  };
}

/**
 * 모든 보드에 가상의 stage/role/importance 필드를 추가합니다.
 * (기존 코드와의 호환성을 위해)
 */
export function enrichBoardsWithLegacyFields(
  boards: BoardResponse[],
  stageField: ParsedField | null,
  roleField: ParsedField | null,
  importanceField: ParsedField | null
): BoardResponse[] {
  return boards.map(board => ({
    ...board,
    stage: getStageInfo(board, stageField),
    roles: getRoleInfo(board, roleField) ? [getRoleInfo(board, roleField)!] : [],
    importance: getImportanceInfo(board, importanceField),
  }));
}
```

---

### ✅ Phase 3: Dashboard.tsx 수정

#### 3.1 상태 추가
```typescript
// 기존 상태
const [columns, setColumns] = useState<Column[]>([]);

// 추가 상태
const [projectFields, setProjectFields] = useState<ProjectFieldsResponse | null>(null);
```

#### 3.2 fetchBoards 함수 수정
```typescript
const fetchBoards = React.useCallback(async () => {
  if (!selectedProject || !accessToken) {
    setColumns([]);
    return;
  }

  setIsLoading(true);
  setError(null);

  try {
    console.log(`[Dashboard] 보드 로드 시작 (Project: ${selectedProject.name})`);

    // 1. 프로젝트 필드 조회 (Stage, Role, Importance 포함)
    const fieldsData = await getProjectFieldsWithOptions(selectedProject.projectId, accessToken);
    setProjectFields(fieldsData);
    console.log('✅ Fields loaded:', fieldsData);

    // 1.1 Stage 필드가 없으면 에러 (프로젝트 생성 시 자동으로 만들어지므로 있어야 함)
    if (!fieldsData.stageField) {
      throw new Error('Stage 필드가 없습니다. 프로젝트 초기화에 문제가 있습니다.');
    }

    // 2. 보드 조회
    const boardsResponse = await getBoards(selectedProject.projectId, accessToken);
    console.log('✅ Boards loaded:', boardsResponse);

    // 3. Stage별로 빈 컬럼 먼저 생성
    const stageMap = new Map<string, { stage: OptionResponse; boards: BoardResponse[] }>();
    fieldsData.stageField.options.forEach((option) => {
      stageMap.set(option.optionId, { stage: option, boards: [] });
    });

    // 4. 보드를 해당 Stage 컬럼에 추가
    boardsResponse.boards.forEach((board) => {
      const stageOptionId = getFieldOptionId(board, fieldsData.stageField?.field.fieldId);
      if (stageOptionId && stageMap.has(stageOptionId)) {
        stageMap.get(stageOptionId)!.boards.push(board);
      }
    });

    // 5. 컬럼 데이터 생성 (displayOrder로 정렬)
    const newColumns: Column[] = Array.from(stageMap.values())
      .sort((a, b) => a.stage.displayOrder - b.stage.displayOrder)
      .map(({ stage, boards }) => ({
        id: stage.optionId,
        title: stage.label,
        color: stage.color,
        cards: boards.map((board) => ({
          id: board.boardId,
          title: board.title,
          content: board.content || '',
          assignee: board.assignee?.name,
          dueDate: board.dueDate,
          // 레거시 호환성을 위해 추가
          board: {
            ...board,
            stage: getStageInfo(board, fieldsData.stageField),
            roles: getRoleInfo(board, fieldsData.roleField) ? [getRoleInfo(board, fieldsData.roleField)!] : [],
            importance: getImportanceInfo(board, fieldsData.importanceField),
          },
        })),
      }));

    setColumns(newColumns);
  } catch (err) {
    const error = err as Error;
    console.error('❌ 보드 로드 실패:', error);
    setError(`보드 로드 실패: ${error.message}`);
    setColumns([]);
  } finally {
    setIsLoading(false);
  }
}, [selectedProject, accessToken]);
```

#### 3.3 Drag & Drop 핸들러 수정
```typescript
const handleDragEnd = async (result: any) => {
  // ... 기존 DnD 로직 ...

  // Stage 업데이트 시
  if (source.droppableId !== destination.droppableId) {
    try {
      // ✅ 새로운 방식: FieldValueService API 호출
      await boardService.put(
        `/api/boards/${draggedBoard.boardId}/fields/${projectFields?.stageField?.field.fieldId}/value`,
        {
          boardId: draggedBoard.boardId,
          fieldId: projectFields?.stageField?.field.fieldId,
          value: targetColumnId, // 새로운 stage option ID
        },
        {
          headers: { Authorization: `Bearer ${accessToken}` },
        }
      );

      // 보드 목록 새로고침
      fetchBoards();
    } catch (error) {
      console.error('❌ Stage 업데이트 실패:', error);
      alert('Stage 업데이트에 실패했습니다.');
    }
  }
};
```

#### 3.4 테이블 뷰 정렬 수정
```typescript
// Role로 정렬
case 'role':
  sorted.sort((a, b) => {
    const aRole = getRoleInfo(a.board, projectFields?.roleField);
    const bRole = getRoleInfo(b.board, projectFields?.roleField);
    aValue = aRole?.name?.toLowerCase() || '';
    bValue = bRole?.name?.toLowerCase() || '';
    return direction === 'asc'
      ? aValue.localeCompare(bValue)
      : bValue.localeCompare(aValue);
  });
  break;

// Importance로 정렬
case 'importance':
  sorted.sort((a, b) => {
    const aImportance = getImportanceInfo(a.board, projectFields?.importanceField);
    const bImportance = getImportanceInfo(b.board, projectFields?.importanceField);
    aValue = aImportance?.level || 0;
    bValue = bImportance?.level || 0;
    return direction === 'asc' ? aValue - bValue : bValue - aValue;
  });
  break;
```

#### 3.5 테이블 뷰 렌더링 수정
```typescript
{/* Role 컬럼 */}
<td>
  {(() => {
    const role = getRoleInfo(card.board, projectFields?.roleField);
    return role ? (
      <div className="flex items-center gap-1">
        <div
          className="w-3 h-3 rounded-full"
          style={{ backgroundColor: role.color }}
        />
        <span className="text-sm">{role.name}</span>
      </div>
    ) : (
      <span className="text-sm text-gray-400">-</span>
    );
  })()}
</td>

{/* Importance 컬럼 */}
<td>
  {(() => {
    const importance = getImportanceInfo(card.board, projectFields?.importanceField);
    return importance ? (
      <div className="flex items-center gap-1">
        <div
          className="w-3 h-3 rounded-full"
          style={{ backgroundColor: importance.color }}
        />
        <span className="text-sm">{importance.name}</span>
      </div>
    ) : (
      <span className="text-sm text-gray-400">-</span>
    );
  })()}
</td>
```

---

### ✅ Phase 4: CreateBoardModal.tsx 수정

#### 4.1 State 수정
```typescript
// 기존 레거시 state 제거하고 새로운 state로 교체
const [projectFields, setProjectFields] = useState<ProjectFieldsResponse | null>(null);
const [selectedStageOptionId, setSelectedStageOptionId] = useState<string>('');
const [selectedRoleOptionId, setSelectedRoleOptionId] = useState<string>('');
const [selectedImportanceOptionId, setSelectedImportanceOptionId] = useState<string>('');
```

#### 4.2 필드 로딩 수정
```typescript
useEffect(() => {
  const fetchCustomFields = async () => {
    setIsLoadingFields(true);
    try {
      const fieldsData = await getProjectFieldsWithOptions(projectId, accessToken);
      setProjectFields(fieldsData);

      // editData가 있으면 customFields에서 값 추출
      if (editData && editData.customFields) {
        if (fieldsData.stageField) {
          const stageOptionId = editData.customFields[fieldsData.stageField.field.fieldId];
          setSelectedStageOptionId(stageOptionId || '');
        }
        if (fieldsData.roleField) {
          const roleOptionId = editData.customFields[fieldsData.roleField.field.fieldId];
          setSelectedRoleOptionId(roleOptionId || '');
        }
        if (fieldsData.importanceField) {
          const importanceOptionId = editData.customFields[fieldsData.importanceField.field.fieldId];
          setSelectedImportanceOptionId(importanceOptionId || '');
        }
      } else {
        // 기본값 설정 (첫 번째 옵션)
        if (fieldsData.stageField && fieldsData.stageField.options.length > 0) {
          setSelectedStageOptionId(initialStageId || fieldsData.stageField.options[0].optionId);
        }
        if (fieldsData.roleField && fieldsData.roleField.options.length > 0) {
          setSelectedRoleOptionId(fieldsData.roleField.options[0].optionId);
        }
      }

      console.log('✅ Custom Fields 로드:', fieldsData);
    } catch (err) {
      console.error('❌ Custom Fields 로드 실패:', err);
      setError('커스텀 필드를 불러오는데 실패했습니다.');
    } finally {
      setIsLoadingFields(false);
    }
  };

  if (projectId && accessToken) {
    fetchCustomFields();
  }
}, [projectId, accessToken, editData, initialStageId]);
```

#### 4.3 보드 생성/수정 함수 수정
```typescript
const handleSubmit = async () => {
  // 유효성 검사
  if (!title.trim()) {
    alert('제목을 입력해주세요.');
    return;
  }

  if (!selectedStageOptionId) {
    alert('Stage를 선택해주세요.');
    return;
  }

  setIsLoading(true);
  try {
    const boardData = {
      projectId,
      title: title.trim(),
      content: content.trim() || undefined,
      assigneeId: selectedAssigneeIds[0] || undefined,
      dueDate: dueDate || undefined,

      // ⚠️ 레거시 필드 (백엔드가 아직 지원)
      stageId: selectedStageOptionId,
      roleIds: selectedRoleOptionId ? [selectedRoleOptionId] : undefined,
      importanceId: selectedImportanceOptionId || undefined,
    };

    if (editData) {
      await updateBoard(editData.boardId, boardData, accessToken);
    } else {
      const newBoard = await createBoard(boardData, accessToken);

      // ✅ 생성 후 customFields 설정 (옵션: 백엔드에서 자동으로 처리하면 불필요)
      // 만약 백엔드가 레거시 필드를 처리하지 않으면 아래 코드 사용
      /*
      if (projectFields) {
        const fieldUpdates = [];
        if (projectFields.stageField && selectedStageOptionId) {
          fieldUpdates.push({
            fieldId: projectFields.stageField.field.fieldId,
            value: selectedStageOptionId,
          });
        }
        if (projectFields.roleField && selectedRoleOptionId) {
          fieldUpdates.push({
            fieldId: projectFields.roleField.field.fieldId,
            value: selectedRoleOptionId,
          });
        }
        if (projectFields.importanceField && selectedImportanceOptionId) {
          fieldUpdates.push({
            fieldId: projectFields.importanceField.field.fieldId,
            value: selectedImportanceOptionId,
          });
        }

        for (const update of fieldUpdates) {
          await boardService.put(
            `/api/boards/${newBoard.boardId}/fields/${update.fieldId}/value`,
            {
              boardId: newBoard.boardId,
              fieldId: update.fieldId,
              value: update.value,
            },
            {
              headers: { Authorization: `Bearer ${accessToken}` },
            }
          );
        }
      }
      */
    }

    onBoardCreated();
    onClose();
  } catch (err) {
    console.error('❌ 보드 저장 실패:', err);
    setError('보드 저장에 실패했습니다.');
  } finally {
    setIsLoading(false);
  }
};
```

#### 4.4 UI 렌더링 수정
```typescript
{/* Stage 선택 */}
<div>
  <label className="block text-sm font-semibold mb-2">
    Stage <span className="text-red-500">*</span>
  </label>
  <select
    value={selectedStageOptionId}
    onChange={(e) => setSelectedStageOptionId(e.target.value)}
    className="w-full px-3 py-2 border rounded"
  >
    <option value="">선택하세요</option>
    {projectFields?.stageField?.options.map((option) => (
      <option key={option.optionId} value={option.optionId}>
        {option.label}
      </option>
    ))}
  </select>
</div>

{/* Role 선택 */}
<div>
  <label className="block text-sm font-semibold mb-2">Role</label>
  <select
    value={selectedRoleOptionId}
    onChange={(e) => setSelectedRoleOptionId(e.target.value)}
    className="w-full px-3 py-2 border rounded"
  >
    <option value="">선택하세요</option>
    {projectFields?.roleField?.options.map((option) => (
      <option key={option.optionId} value={option.optionId}>
        {option.label}
      </option>
    ))}
  </select>
</div>

{/* Importance 선택 */}
<div>
  <label className="block text-sm font-semibold mb-2">Importance</label>
  <select
    value={selectedImportanceOptionId}
    onChange={(e) => setSelectedImportanceOptionId(e.target.value)}
    className="w-full px-3 py-2 border rounded"
  >
    <option value="">선택하세요</option>
    {projectFields?.importanceField?.options.map((option) => (
      <option key={option.optionId} value={option.optionId}>
        {option.label}
      </option>
    ))}
  </select>
</div>
```

---

### ✅ Phase 5: BoardDetailModal.tsx 수정

#### 5.1 State 추가
```typescript
const [projectFields, setProjectFields] = useState<ProjectFieldsResponse | null>(null);
```

#### 5.2 보드 로딩 시 필드도 함께 로딩
```typescript
useEffect(() => {
  const fetchBoard = async () => {
    setIsLoadingBoard(true);
    try {
      const boardData = await getBoard(boardId, accessToken);

      // 프로젝트 필드 로딩
      const fieldsData = await getProjectFieldsWithOptions(boardData.projectId, accessToken);
      setProjectFields(fieldsData);

      // 보드 데이터로 상태 초기화
      setProjectId(boardData.projectId);
      setTitle(boardData.title);
      setContent(boardData.content || '');

      // customFields에서 값 추출
      if (fieldsData.stageField) {
        const stageOptionId = getFieldOptionId(boardData, fieldsData.stageField.field.fieldId);
        setSelectedStageId(stageOptionId || '');
      }
      if (fieldsData.roleField) {
        const roleOptionId = getFieldOptionId(boardData, fieldsData.roleField.field.fieldId);
        setSelectedRoleId(roleOptionId || '');
      }
      if (fieldsData.importanceField) {
        const importanceOptionId = getFieldOptionId(boardData, fieldsData.importanceField.field.fieldId);
        setSelectedImportanceId(importanceOptionId || '');
      }

      // ... 나머지 필드들
    } catch (err) {
      console.error('❌ 보드 로드 실패:', err);
    } finally {
      setIsLoadingBoard(false);
    }
  };

  fetchBoard();
}, [boardId, accessToken]);
```

---

## 테스트 체크리스트

### ✅ 기능 테스트
- [ ] 프로젝트 생성 시 Stage/Role/Importance 필드가 자동 생성되는지 확인
- [ ] Dashboard에서 보드가 올바른 Stage 컬럼에 표시되는지 확인
- [ ] Drag & Drop으로 Stage 변경 시 정상 동작하는지 확인
- [ ] 테이블 뷰에서 Role, Importance 정렬이 정상 동작하는지 확인
- [ ] 테이블 뷰에서 Role, Importance가 올바르게 표시되는지 확인
- [ ] CreateBoardModal에서 Stage/Role/Importance 선택이 정상 동작하는지 확인
- [ ] 보드 생성 시 선택한 필드 값이 저장되는지 확인
- [ ] BoardDetailModal에서 기존 보드의 필드 값이 올바르게 표시되는지 확인
- [ ] 보드 수정 시 필드 값이 업데이트되는지 확인

### ✅ 엣지 케이스
- [ ] 프로젝트에 필드가 없는 경우 처리
- [ ] customFields가 비어있는 보드 처리
- [ ] 잘못된 fieldId/optionId 처리
- [ ] 필드는 있지만 옵션이 없는 경우 처리

### ✅ 성능 테스트
- [ ] 보드 100개 이상일 때 렌더링 성능 확인
- [ ] 필드/옵션 API 호출이 중복되지 않는지 확인
- [ ] 캐싱이 제대로 동작하는지 확인

---

## 추가 개선 사항 (선택)

### 1. 필드 데이터 캐싱
React Query나 Context API를 사용해서 프로젝트 필드 데이터를 캐싱하면 성능 향상:

```typescript
// frontend/src/contexts/ProjectFieldsContext.tsx
import React, { createContext, useContext, useState, useEffect } from 'react';

interface ProjectFieldsContextType {
  projectFields: ProjectFieldsResponse | null;
  loading: boolean;
  error: Error | null;
  refreshFields: () => Promise<void>;
}

const ProjectFieldsContext = createContext<ProjectFieldsContextType | undefined>(undefined);

export const ProjectFieldsProvider: React.FC<{ projectId: string; children: React.ReactNode }> = ({
  projectId,
  children,
}) => {
  const [projectFields, setProjectFields] = useState<ProjectFieldsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const loadFields = async () => {
    setLoading(true);
    try {
      const accessToken = localStorage.getItem('accessToken') || '';
      const data = await getProjectFieldsWithOptions(projectId, accessToken);
      setProjectFields(data);
      setError(null);
    } catch (err) {
      setError(err as Error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFields();
  }, [projectId]);

  return (
    <ProjectFieldsContext.Provider
      value={{
        projectFields,
        loading,
        error,
        refreshFields: loadFields,
      }}
    >
      {children}
    </ProjectFieldsContext.Provider>
  );
};

export const useProjectFields = () => {
  const context = useContext(ProjectFieldsContext);
  if (!context) {
    throw new Error('useProjectFields must be used within ProjectFieldsProvider');
  }
  return context;
};
```

### 2. 커스텀 Hook 작성
```typescript
// frontend/src/hooks/useCustomFields.ts
export function useStageInfo(board: BoardResponse | null) {
  const { projectFields } = useProjectFields();
  if (!board || !projectFields) return null;
  return getStageInfo(board, projectFields.stageField);
}

export function useRoleInfo(board: BoardResponse | null) {
  const { projectFields } = useProjectFields();
  if (!board || !projectFields) return null;
  return getRoleInfo(board, projectFields.roleField);
}

export function useImportanceInfo(board: BoardResponse | null) {
  const { projectFields } = useProjectFields();
  if (!board || !projectFields) return null;
  return getImportanceInfo(board, projectFields.importanceField);
}
```

---

## 참고 자료

- **백엔드 API 문서**: `board-service/docs/swagger.yaml`
- **Field 관련 서비스**: `board-service/internal/service/field_service.go`
- **FieldValue 관련 서비스**: `board-service/internal/service/field_value_service.go`
- **Board 응답 구조**: `board-service/internal/dto/board.go`

---

## 마이그레이션 우선순위

**HIGH** (필수):
1. Phase 1: API 레이어 수정
2. Phase 2: 유틸리티 함수 작성
3. Phase 3: Dashboard.tsx 수정

**MEDIUM** (중요):
4. Phase 4: CreateBoardModal.tsx 수정
5. Phase 5: BoardDetailModal.tsx 수정

**LOW** (선택):
6. 추가 개선 사항 (캐싱, 커스텀 Hook 등)

---

**작성일**: 2025-11-11
**작성자**: Claude Code Assistant
**관련 브랜치**: `claude/frontend-endpoint-fix-011CV1TzUXkagZBK1JA9ViCy`
