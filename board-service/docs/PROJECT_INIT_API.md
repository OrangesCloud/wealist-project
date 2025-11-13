# 프로젝트 초기 로딩 API 문서

## 개요

프로젝트 페이지를 처음 로딩할 때 필요한 모든 데이터를 한 번의 API 호출로 가져오는 엔드포인트입니다.

## API 엔드포인트

```
GET /api/projects/{projectId}/init-data
```

## 인증

- **필수**: Bearer Token
- **권한**: 프로젝트 멤버만 접근 가능

## 요청

### Path Parameters

| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|------|------|
| projectId | UUID | O | 프로젝트 ID |

### Headers

```
Authorization: Bearer <JWT_TOKEN>
```

### 요청 예시

```bash
curl -X GET "http://localhost:8000/api/projects/550e8400-e29b-41d4-a716-446655440000/init-data" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 응답

### 성공 응답 (200 OK)

```json
{
  "data": {
    "project": {
      "projectId": "550e8400-e29b-41d4-a716-446655440000",
      "name": "웹사이트 리뉴얼 프로젝트",
      "description": "회사 홈페이지 전면 리뉴얼",
      "workspaceId": "550e8400-e29b-41d4-a716-446655440099",
      "ownerId": "550e8400-e29b-41d4-a716-446655440003",
      "isPublic": false,
      "createdAt": "2025-11-01T10:00:00Z",
      "updatedAt": "2025-11-12T15:00:00Z"
    },
    "boards": [
      {
        "boardId": "550e8400-e29b-41d4-a716-446655440001",
        "projectId": "550e8400-e29b-41d4-a716-446655440000",
        "title": "프로젝트 기획서 작성",
        "content": "상세 기획서 작성 및 검토",
        "assignee": {
          "userId": "550e8400-e29b-41d4-a716-446655440002",
          "name": "홍길동",
          "email": "hong@example.com",
          "isActive": true
        },
        "author": {
          "userId": "550e8400-e29b-41d4-a716-446655440003",
          "name": "김철수",
          "email": "kim@example.com",
          "isActive": true
        },
        "dueDate": "2025-12-31T23:59:59Z",
        "createdAt": "2025-11-01T10:00:00Z",
        "updatedAt": "2025-11-12T15:30:00Z",
        "customFields": {
          "550e8400-e29b-41d4-a716-446655440010": "550e8400-e29b-41d4-a716-446655440020",
          "550e8400-e29b-41d4-a716-446655440011": "높음"
        },
        "fieldValues": [
          {
            "valueId": "550e8400-e29b-41d4-a716-446655440100",
            "fieldId": "550e8400-e29b-41d4-a716-446655440010",
            "fieldName": "상태",
            "fieldType": "single_select",
            "value": {
              "optionId": "550e8400-e29b-41d4-a716-446655440020",
              "label": "할 일",
              "color": "#94A3B8",
              "description": "아직 시작하지 않은 작업"
            },
            "displayOrder": 0,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          },
          {
            "valueId": "550e8400-e29b-41d4-a716-446655440101",
            "fieldId": "550e8400-e29b-41d4-a716-446655440011",
            "fieldName": "우선순위",
            "fieldType": "single_select",
            "value": {
              "optionId": "550e8400-e29b-41d4-a716-446655440032",
              "label": "높음",
              "color": "#EF4444",
              "description": ""
            },
            "displayOrder": 0,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          }
        ],
        "position": "a0"
      }
    ],
    "fields": [
      {
        "fieldId": "550e8400-e29b-41d4-a716-446655440010",
        "projectId": "550e8400-e29b-41d4-a716-446655440000",
        "name": "상태",
        "fieldType": "single_select",
        "description": "작업 진행 상태",
        "displayOrder": 0,
        "isRequired": true,
        "isSystemDefault": true,
        "config": {},
        "canEditRoles": ["ADMIN", "OWNER"],
        "options": [
          {
            "optionId": "550e8400-e29b-41d4-a716-446655440020",
            "fieldId": "550e8400-e29b-41d4-a716-446655440010",
            "label": "할 일",
            "color": "#94A3B8",
            "description": "아직 시작하지 않은 작업",
            "displayOrder": 0,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          },
          {
            "optionId": "550e8400-e29b-41d4-a716-446655440021",
            "fieldId": "550e8400-e29b-41d4-a716-446655440010",
            "label": "진행 중",
            "color": "#3B82F6",
            "description": "현재 진행 중인 작업",
            "displayOrder": 1,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          },
          {
            "optionId": "550e8400-e29b-41d4-a716-446655440022",
            "fieldId": "550e8400-e29b-41d4-a716-446655440010",
            "label": "완료",
            "color": "#10B981",
            "description": "완료된 작업",
            "displayOrder": 2,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          }
        ],
        "createdAt": "2025-11-01T10:00:00Z",
        "updatedAt": "2025-11-01T10:00:00Z"
      },
      {
        "fieldId": "550e8400-e29b-41d4-a716-446655440011",
        "projectId": "550e8400-e29b-41d4-a716-446655440000",
        "name": "우선순위",
        "fieldType": "single_select",
        "description": "작업 우선순위",
        "displayOrder": 1,
        "isRequired": false,
        "isSystemDefault": true,
        "config": {},
        "canEditRoles": [],
        "options": [
          {
            "optionId": "550e8400-e29b-41d4-a716-446655440030",
            "fieldId": "550e8400-e29b-41d4-a716-446655440011",
            "label": "낮음",
            "color": "#94A3B8",
            "description": "",
            "displayOrder": 0,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          },
          {
            "optionId": "550e8400-e29b-41d4-a716-446655440031",
            "fieldId": "550e8400-e29b-41d4-a716-446655440011",
            "label": "보통",
            "color": "#FBBF24",
            "description": "",
            "displayOrder": 1,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          },
          {
            "optionId": "550e8400-e29b-41d4-a716-446655440032",
            "fieldId": "550e8400-e29b-41d4-a716-446655440011",
            "label": "높음",
            "color": "#EF4444",
            "description": "",
            "displayOrder": 2,
            "createdAt": "2025-11-01T10:00:00Z",
            "updatedAt": "2025-11-01T10:00:00Z"
          }
        ],
        "createdAt": "2025-11-01T10:00:00Z",
        "updatedAt": "2025-11-01T10:00:00Z"
      }
    ],
    "fieldTypes": [
      {
        "type": "text",
        "displayName": "텍스트",
        "description": "짧은 텍스트 입력",
        "hasOptions": false
      },
      {
        "type": "number",
        "displayName": "숫자",
        "description": "숫자 입력",
        "hasOptions": false
      },
      {
        "type": "single_select",
        "displayName": "단일 선택",
        "description": "하나의 옵션 선택",
        "hasOptions": true
      },
      {
        "type": "multi_select",
        "displayName": "다중 선택",
        "description": "여러 옵션 선택",
        "hasOptions": true
      },
      {
        "type": "date",
        "displayName": "날짜",
        "description": "날짜 선택",
        "hasOptions": false
      },
      {
        "type": "datetime",
        "displayName": "날짜/시간",
        "description": "날짜와 시간 선택",
        "hasOptions": false
      },
      {
        "type": "single_user",
        "displayName": "담당자",
        "description": "한 명의 사용자 지정",
        "hasOptions": false
      },
      {
        "type": "multi_user",
        "displayName": "다중 담당자",
        "description": "여러 사용자 지정",
        "hasOptions": false
      },
      {
        "type": "checkbox",
        "displayName": "체크박스",
        "description": "예/아니오 선택",
        "hasOptions": false
      },
      {
        "type": "url",
        "displayName": "URL",
        "description": "웹 링크",
        "hasOptions": false
      }
    ],
    "members": [
      {
        "userId": "550e8400-e29b-41d4-a716-446655440003",
        "name": "김철수",
        "email": "kim@example.com",
        "role": "OWNER",
        "joinedAt": "2025-11-01T10:00:00Z"
      },
      {
        "userId": "550e8400-e29b-41d4-a716-446655440002",
        "name": "홍길동",
        "email": "hong@example.com",
        "role": "ADMIN",
        "joinedAt": "2025-11-02T09:00:00Z"
      },
      {
        "userId": "550e8400-e29b-41d4-a716-446655440004",
        "name": "이영희",
        "email": "lee@example.com",
        "role": "MEMBER",
        "joinedAt": "2025-11-03T14:30:00Z"
      }
    ],
    "defaultViewId": "550e8400-e29b-41d4-a716-446655440088"
  }
}
```

### 에러 응답

#### 400 Bad Request
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "잘못된 프로젝트 ID"
  }
}
```

#### 401 Unauthorized
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "인증이 필요합니다"
  }
}
```

#### 403 Forbidden
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "프로젝트 멤버가 아닙니다"
  }
}
```

#### 500 Internal Server Error
```json
{
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "서버 오류가 발생했습니다"
  }
}
```

## 응답 데이터 구조

### ProjectInitDataResponse

| 필드 | 타입 | 설명 |
|-----|------|------|
| project | ProjectBasicInfo | 프로젝트 기본 정보 |
| boards | BoardResponse[] | 프로젝트의 모든 보드 목록 (position 순서로 정렬됨) |
| fields | FieldWithOptionsResponse[] | 필드 정의 + 옵션 목록 |
| fieldTypes | FieldTypeInfo[] | 사용 가능한 필드 타입 정보 |
| members | ProjectMemberBasicInfo[] | 프로젝트 멤버 목록 (담당자 할당용) |
| defaultViewId | UUID | 기본 뷰 ID (없으면 빈 문자열) |

### ProjectBasicInfo

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| projectId | UUID | O | 프로젝트 ID |
| name | string | O | 프로젝트 이름 |
| description | string | - | 프로젝트 설명 |
| workspaceId | UUID | O | 워크스페이스 ID |
| ownerId | UUID | O | 소유자 ID |
| isPublic | bool | O | 공개 여부 |
| createdAt | string | O | 생성일시 (RFC3339) |
| updatedAt | string | O | 수정일시 (RFC3339) |

### BoardResponse

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| boardId | UUID | O | 보드 ID |
| projectId | UUID | O | 프로젝트 ID |
| title | string | O | 보드 제목 |
| content | string | - | 보드 내용 |
| assignee | UserInfo | - | 담당자 정보 |
| author | UserInfo | O | 작성자 정보 |
| dueDate | timestamp | - | 마감일 |
| createdAt | timestamp | O | 생성일시 |
| updatedAt | timestamp | O | 수정일시 |
| customFields | map[string]interface{} | - | 커스텀 필드 값 (fieldId: value) - Legacy |
| fieldValues | FieldValueWithInfo[] | - | 필드 값 배열 (field 정보 포함) ✨ NEW |
| position | string | - | 보드 순서 (Fractional indexing, 기본 뷰의 순서) |

### FieldWithOptionsResponse

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| fieldId | UUID | O | 필드 ID |
| projectId | UUID | O | 프로젝트 ID |
| name | string | O | 필드 이름 |
| fieldType | string | O | 필드 타입 (text, number, single_select 등) |
| description | string | - | 필드 설명 |
| displayOrder | int | O | 표시 순서 |
| isRequired | bool | O | 필수 여부 |
| isSystemDefault | bool | O | 시스템 기본 필드 여부 |
| config | map[string]interface{} | O | 타입별 설정 |
| canEditRoles | string[] | O | 수정 가능 역할 목록 |
| options | OptionResponse[] | O | 필드 옵션 목록 (single_select, multi_select 타입용) |
| createdAt | timestamp | O | 생성일시 |
| updatedAt | timestamp | O | 수정일시 |

### OptionResponse

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| optionId | UUID | O | 옵션 ID |
| fieldId | UUID | O | 필드 ID |
| label | string | O | 옵션 레이블 |
| color | string | - | 색상 코드 (#RRGGBB) |
| description | string | - | 옵션 설명 |
| displayOrder | int | O | 표시 순서 |
| createdAt | timestamp | O | 생성일시 |
| updatedAt | timestamp | O | 수정일시 |

### FieldValueWithInfo ✨ NEW

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| valueId | UUID | O | 필드 값 ID |
| fieldId | UUID | O | 필드 ID |
| fieldName | string | O | **필드 이름** (예: "상태", "우선순위") |
| fieldType | string | O | **필드 타입** (예: "single_select", "text") |
| value | interface{} | O | 실제 값 (타입에 따라 다름) |
| displayOrder | int | - | 표시 순서 (multi_select, multi_user용) |
| createdAt | timestamp | O | 생성일시 |
| updatedAt | timestamp | O | 수정일시 |

**value 필드 타입별 형식:**
- `text`, `url`: string
- `number`: number
- `date`, `datetime`: timestamp
- `checkbox`: boolean
- `single_select`, `multi_select`: Option 객체 `{ optionId, label, color, description }`
- `single_user`, `multi_user`: string (user ID)

### FieldTypeInfo

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| type | string | O | 필드 타입 키 (text, number, date 등) |
| displayName | string | O | 사용자에게 표시할 이름 |
| description | string | O | 타입 설명 |
| hasOptions | bool | O | 옵션 사용 여부 (single_select, multi_select만 true) |

### UserInfo

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| userId | UUID | O | 사용자 ID |
| name | string | O | 사용자 이름 |
| email | string | O | 이메일 |
| isActive | bool | O | 활성 상태 |

### ProjectMemberBasicInfo

| 필드 | 타입 | 필수 | 설명 |
|-----|------|------|------|
| userId | UUID | O | 사용자 ID |
| name | string | O | 사용자 이름 |
| email | string | O | 이메일 |
| role | string | O | 역할 (OWNER, ADMIN, MEMBER) |
| joinedAt | string | O | 가입일시 (RFC3339) |

## 사용 가능한 필드 타입

| 타입 | 한글명 | 설명 | 옵션 지원 |
|-----|--------|------|----------|
| text | 텍스트 | 짧은 텍스트 입력 | X |
| number | 숫자 | 숫자 입력 | X |
| single_select | 단일 선택 | 하나의 옵션 선택 | O |
| multi_select | 다중 선택 | 여러 옵션 선택 | O |
| date | 날짜 | 날짜 선택 | X |
| datetime | 날짜/시간 | 날짜와 시간 선택 | X |
| single_user | 담당자 | 한 명의 사용자 지정 | X |
| multi_user | 다중 담당자 | 여러 사용자 지정 | X |
| checkbox | 체크박스 | 예/아니오 선택 | X |
| url | URL | 웹 링크 | X |

## 구현 세부사항

### 1. 프로젝트 기본 정보
- 프로젝트의 기본 정보를 조회합니다 (이름, 설명, 소유자 등)
- 페이지 제목, 프로젝트 헤더 표시에 사용됩니다

### 2. 멤버 목록
- 프로젝트의 모든 멤버 정보를 조회합니다
- 각 멤버의 이름, 이메일, 역할(OWNER/ADMIN/MEMBER)을 포함합니다
- **용도**: 보드에 담당자 할당 시 드롭다운 목록으로 사용

### 3. 기본 뷰 및 보드 순서
- 프로젝트의 기본 뷰(`is_default = true`)를 찾습니다
- 기본 뷰가 있으면 해당 뷰의 보드 순서를 조회합니다 (`user_board_order` 테이블)
- 각 보드에 `position` 필드가 포함됩니다 (Fractional indexing 문자열)
- **정렬 순서**:
  1. position이 있는 보드: `position` 문자열 사전순 정렬
  2. position이 없는 보드: `createdAt` 시간순 정렬
  3. position 있는 보드가 먼저, 없는 보드가 나중

### 4. 보드 데이터
- 프로젝트에 속한 모든 보드를 가져옵니다 (최대 1000개)
- 각 보드의 담당자(assignee)와 작성자(author) 정보를 User Service에서 조회하여 포함합니다
- `customFields`는 보드의 `custom_fields_cache` JSONB 컬럼에서 파싱됩니다 (Legacy)
- **✨ NEW: `fieldValues`**: 각 보드의 모든 필드 값을 field 정보와 함께 포함합니다
  - 배치 최적화: 모든 보드의 field values를 한 번에 조회 (N+1 문제 방지)
  - Field 메타데이터 포함: field name, field type이 함께 제공되어 별도 조회 불필요
  - Option 상세 정보 포함: single_select/multi_select 타입의 경우 option의 label, color, description 포함
- 위에서 조회한 순서 정보에 따라 정렬되어 반환됩니다

### 5. 필드 데이터
- 프로젝트의 모든 커스텀 필드를 `displayOrder` 순서로 가져옵니다
- 각 필드에 대해 관련된 모든 옵션을 `displayOrder` 순서로 가져옵니다
- `single_select`와 `multi_select` 타입의 필드만 옵션을 가집니다
- 다른 타입의 필드는 빈 배열(`[]`)이 반환됩니다

### 6. 필드 타입 정보
- 시스템에서 지원하는 모든 필드 타입의 메타데이터를 반환합니다
- 프론트엔드에서 새 필드 생성 시 타입 선택 UI에 사용됩니다

## 성능 최적화

### 캐싱
- 사용자 정보는 Redis 캐시를 통해 조회됩니다 (TTL: 5분)
- 필드 정의도 캐싱되어 반복 조회를 최소화합니다

### N+1 쿼리 방지
- 모든 보드의 작성자 ID를 수집하여 배치로 사용자 정보를 조회합니다
- 필드와 옵션은 프로젝트 단위로 한 번에 조회합니다
- **✨ NEW**: Field values도 배치로 조회합니다
  - 모든 보드의 field values를 한 번의 쿼리로 조회
  - 필요한 field 메타데이터를 한 번에 조회
  - 필요한 option 정보를 한 번에 조회

## 프론트엔드 사용 예시

```typescript
interface ProjectInitData {
  project: ProjectBasicInfo;
  boards: Board[];
  fields: FieldWithOptions[];
  fieldTypes: FieldTypeInfo[];
  members: ProjectMemberBasicInfo[];
  defaultViewId: string;
}

interface ProjectBasicInfo {
  projectId: string;
  name: string;
  description: string;
  workspaceId: string;
  ownerId: string;
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
}

interface ProjectMemberBasicInfo {
  userId: string;
  name: string;
  email: string;
  role: 'OWNER' | 'ADMIN' | 'MEMBER';
  joinedAt: string;
}

async function loadProjectPage(projectId: string, token: string): Promise<ProjectInitData> {
  const response = await fetch(`/api/projects/${projectId}/init-data`, {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to load project data');
  }

  const { data } = await response.json();
  return data;
}

// 사용
const initData = await loadProjectPage('550e8400-e29b-41d4-a716-446655440000', token);

// 프로젝트 제목 표시
document.title = initData.project.name;
document.getElementById('project-header').textContent = initData.project.name;

// 보드 렌더링 (이미 정렬된 순서대로)
renderBoards(initData.boards);

// 담당자 선택 드롭다운 구성
setupAssigneeDropdown(initData.members);

// 필드 설정 UI 구성
setupFieldConfiguration(initData.fields, initData.fieldTypes);
```

## 참고사항

1. **권한**: 프로젝트 멤버만 이 API를 호출할 수 있습니다
2. **페이지네이션**: 보드는 최대 1000개까지만 반환됩니다 (향후 필요시 페이지네이션 추가 가능)
3. **User Service 의존성**: 작성자/담당자/멤버 정보 조회를 위해 User Service가 정상 동작해야 합니다
4. **캐시 무효화**: 필드나 옵션 변경 시 캐시가 자동으로 무효화됩니다
5. **보드 정렬**: 기본 뷰가 있으면 해당 뷰의 순서로, 없으면 생성일시 순으로 정렬됩니다

## 버전 히스토리

- **v1.2.0** (2025-01-13): Field Values 정보 추가 🎉
  - ✨ **Board 응답에 `fieldValues` 필드 추가**
    - Field 명칭 포함 (`fieldName`)
    - Field 타입 포함 (`fieldType`)
    - Option 상세 정보 포함 (label, color, description)
    - 배치 최적화로 성능 개선 (N+1 쿼리 방지)
  - 🔧 `customFields`는 Legacy로 유지 (하위 호환성)
  - 📖 FieldValueWithInfo DTO 추가

- **v1.1.0** (2025-11-12): 멤버 및 순서 정보 추가
  - ✨ 프로젝트 기본 정보 추가 (`project`)
  - ✨ 멤버 목록 추가 (`members`) - 담당자 할당 드롭다운용
  - ✨ 기본 뷰 ID 추가 (`defaultViewId`)
  - ✨ 보드 순서 정보 포함 (`position` 필드) - Fractional indexing
  - 🔧 보드가 position 순서로 정렬되어 반환됨

- **v1.0.0** (2025-11-12): 초기 릴리스
  - 보드, 필드, 필드 타입 정보를 한 번에 반환
  - 필드 옵션 포함 (색상, 레이블)
