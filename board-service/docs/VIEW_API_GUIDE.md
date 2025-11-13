# 프론트엔드를 위한 View API 가이드

> **간단 요약**: View는 "프로젝트 보드를 어떻게 볼 것인가"를 저장한 설정입니다.
> 노션의 "데이터베이스 뷰" 또는 지라의 "필터/보드/리스트 뷰"와 같은 개념입니다.

---

## 목차

1. [View가 뭔가요?](#view가-뭔가요)
2. [화면별 API 호출 가이드](#화면별-api-호출-가이드)
3. [API 상세 명세](#api-상세-명세)
4. [실전 예제 코드](#실전-예제-코드)
5. [자주 묻는 질문](#자주-묻는-질문)

---

## View가 뭔가요?

### 개념 이해

```
프로젝트 = 전체 보드 데이터

View = 보드를 보는 방법
├─ 필터: "어떤 보드만 볼까?" (예: 내가 담당한 것만)
├─ 정렬: "어떤 순서로 볼까?" (예: 최신순)
└─ 그룹핑: "어떻게 묶어서 볼까?" (예: 상태별로 칸반처럼)
```

### 실제 사용 예시

**프로젝트**: "모바일 앱 개발 프로젝트" (보드 100개)

**View 1: "내 작업"**
- 필터: 담당자 = 나
- 정렬: 생성일 내림차순
- 그룹핑: 없음 (리스트 형태)
→ 결과: 내가 담당한 보드 10개를 리스트로 표시

**View 2: "칸반 보드"**
- 필터: 없음 (전체 보드)
- 정렬: 없음
- 그룹핑: 상태 필드 (할일/진행중/완료)
→ 결과: 100개 보드를 상태별로 묶어서 칸반 형태로 표시

**View 3: "긴급 버그"**
- 필터: 우선순위 = "긴급" AND 타입 = "버그"
- 정렬: 생성일 내림차순
- 그룹핑: 없음
→ 결과: 긴급 버그 5개를 리스트로 표시

---

## 화면별 API 호출 가이드

### 1️⃣ 프로젝트 페이지 진입

**화면**: 사용자가 프로젝트를 처음 열었을 때

**해야 할 일**:
1. 프로젝트의 모든 뷰 목록 가져오기
2. 기본 뷰가 있으면 자동으로 적용, 없으면 첫 번째 뷰 선택

**API 호출**:

```typescript
// 1. 뷰 목록 가져오기
GET /api/views?projectId={project_id}

// 응답 예시
{
  "data": [
    {
      "viewId": "view-111",
      "name": "전체 보드",
      "isDefault": true,
      "isShared": true,
      "filters": {},
      "sortBy": "created_at",
      "sortDirection": "desc",
      "groupByFieldId": ""
    },
    {
      "viewId": "view-222",
      "name": "내 작업",
      "isDefault": false,
      "isShared": false,
      "filters": {
        "assignee": { "operator": "eq", "value": "my-user-id" }
      },
      "sortBy": "priority",
      "sortDirection": "desc",
      "groupByFieldId": ""
    },
    {
      "viewId": "view-333",
      "name": "칸반 보드",
      "isDefault": false,
      "isShared": true,
      "filters": {},
      "sortBy": "",
      "sortDirection": "asc",
      "groupByFieldId": "status-field-id"
    }
  ]
}
```

**프론트 코드 예시**:

```typescript
// 프로젝트 진입 시
async function onProjectEnter(projectId: string) {
  // 1. 뷰 목록 가져오기
  const views = await fetchViews(projectId);

  // 2. 기본 뷰 찾기 (없으면 첫 번째 뷰)
  const defaultView = views.find(v => v.isDefault) || views[0];

  // 3. 기본 뷰 적용
  if (defaultView) {
    await applyView(defaultView.viewId);
  }
}
```

---

### 2️⃣ 뷰 선택/변경

**화면**: 사용자가 뷰 드롭다운에서 다른 뷰를 선택했을 때

**해야 할 일**: 선택한 뷰의 설정대로 보드 데이터 가져오기

**API 호출**:

```typescript
// 뷰 적용 (보드 데이터 가져오기)
GET /api/views/{view_id}/apply?page=1&limit=20

// 응답 예시 1: 그룹핑 없음 (리스트 형태)
{
  "boards": [
    {
      "id": "board-1",
      "project_id": "project-123",
      "title": "로그인 기능 구현",
      "content": "OAuth 2.0 사용",
      "custom_fields": {
        "status-field-id": "option-in-progress",
        "assignee-field-id": "user-456"
      },
      "position": "a0",
      "created_at": "2025-01-10T10:00:00Z",
      "updated_at": "2025-01-10T10:00:00Z"
    },
    {
      "id": "board-2",
      "title": "회원가입 페이지 디자인",
      "position": "a1",
      // ...
    }
  ],
  "total": 45,
  "page": 1,
  "limit": 20
}

// 응답 예시 2: 그룹핑 있음 (칸반 형태)
{
  "groupByField": {
    "fieldId": "status-field-id",
    "name": "상태",
    "fieldType": "single_select"
  },
  "groups": [
    {
      "groupValue": {
        "option_id": "option-todo",
        "label": "할 일",
        "color": "#gray"
      },
      "boards": [
        { "id": "board-1", "title": "...", "position": "a0" },
        { "id": "board-2", "title": "...", "position": "a1" }
      ],
      "count": 2
    },
    {
      "groupValue": {
        "option_id": "option-in-progress",
        "label": "진행중",
        "color": "#blue"
      },
      "boards": [
        { "id": "board-3", "title": "...", "position": "b0" }
      ],
      "count": 1
    },
    {
      "groupValue": {
        "option_id": "option-done",
        "label": "완료",
        "color": "#green"
      },
      "boards": [],
      "count": 0
    }
  ],
  "total": 3
}
```

**프론트 코드 예시**:

```typescript
async function applyView(viewId: string, page = 1, limit = 20) {
  const response = await axios.get(`/api/views/${viewId}/apply`, {
    params: { page, limit }
  });

  // 그룹핑이 있는지 확인
  if (response.data.groups) {
    // 칸반 보드 렌더링
    renderKanbanBoard(response.data.groups);
  } else {
    // 리스트 렌더링
    renderBoardList(response.data.boards);
  }
}
```

---

### 3️⃣ 새 뷰 만들기

**화면**: 사용자가 "새 뷰 만들기" 버튼을 눌렀을 때

**해야 할 일**:
1. 뷰 생성 모달 열기
2. 사용자가 이름, 필터, 정렬, 그룹핑 설정
3. API 호출해서 뷰 저장

**API 호출**:

```typescript
POST /api/views

// 요청 바디 예시 1: 리스트 뷰
{
  "projectId": "project-123",
  "name": "내가 담당한 작업",
  "description": "나에게 할당된 보드만 표시",
  "isDefault": false,
  "isShared": false,
  "filters": {
    "assignee-field-id": {
      "operator": "eq",
      "value": "my-user-id"
    }
  },
  "sortBy": "created_at",
  "sortDirection": "desc",
  "groupByFieldId": ""
}

// 요청 바디 예시 2: 칸반 뷰
{
  "projectId": "project-123",
  "name": "개발 칸반 보드",
  "description": "상태별로 묶어서 보기",
  "isDefault": false,
  "isShared": true,  // 팀원도 볼 수 있게
  "filters": {},
  "sortBy": "",
  "sortDirection": "asc",
  "groupByFieldId": "status-field-id"  // 상태 필드로 그룹핑
}

// 응답
{
  "data": {
    "viewId": "new-view-id",
    "projectId": "project-123",
    "name": "내가 담당한 작업",
    "createdBy": "my-user-id",
    // ... 저장된 뷰 정보
  }
}
```

**프론트 코드 예시**:

```typescript
async function createView(formData: {
  name: string;
  filters: Record<string, FilterCondition>;
  sortBy: string;
  groupByFieldId: string;
}) {
  const response = await axios.post('/api/views', {
    projectId: currentProjectId,
    name: formData.name,
    description: '',
    isDefault: false,
    isShared: false,
    filters: formData.filters,
    sortBy: formData.sortBy,
    sortDirection: 'asc',
    groupByFieldId: formData.groupByFieldId
  });

  // 뷰 목록 갱신
  await refreshViewList();

  // 새로 만든 뷰로 전환
  await applyView(response.data.viewId);
}
```

---

### 4️⃣ 뷰 수정하기

**화면**: 사용자가 현재 뷰의 설정을 변경했을 때

**해야 할 일**: 뷰 설정 업데이트

**API 호출**:

```typescript
PUT /api/views/{view_id}

// 요청 바디 (변경할 필드만 보내면 됨)
{
  "name": "진행중인 작업",  // 이름만 변경
}

// 또는 필터 추가
{
  "filters": {
    "status-field-id": {
      "operator": "eq",
      "value": "in-progress-option-id"
    }
  }
}

// 응답: 업데이트된 뷰 정보
```

**프론트 코드 예시**:

```typescript
async function updateViewFilters(viewId: string, newFilters: any) {
  await axios.put(`/api/views/${viewId}`, {
    filters: newFilters
  });

  // 뷰 다시 적용
  await applyView(viewId);
}
```

---

### 5️⃣ 뷰 삭제하기

**화면**: 사용자가 "뷰 삭제" 버튼을 눌렀을 때

**API 호출**:

```typescript
DELETE /api/views/{view_id}

// 응답: 성공 메시지
```

**프론트 코드 예시**:

```typescript
async function deleteView(viewId: string) {
  if (!confirm('정말 이 뷰를 삭제하시겠습니까?')) return;

  await axios.delete(`/api/views/${viewId}`);

  // 뷰 목록 갱신
  const views = await fetchViews(currentProjectId);

  // 다른 뷰로 전환 (기본 뷰 또는 첫 번째 뷰)
  const nextView = views.find(v => v.isDefault) || views[0];
  if (nextView) {
    await applyView(nextView.viewId);
  }
}
```

---

### 6️⃣ 보드 순서 변경 (드래그앤드롭)

**화면**: 사용자가 보드를 드래그해서 순서를 바꿨을 때

**중요**: 보드 순서는 **뷰별, 사용자별**로 다릅니다!
- 같은 뷰를 봐도 철수와 영희의 보드 순서가 다를 수 있음
- 다른 뷰에서는 같은 보드의 순서가 다를 수 있음

**API 호출 방법 2가지**:

#### 방법 1: 보드 이동 API (권장)

```typescript
// 한 개 보드 이동 (칸반 보드의 컬럼 간 이동 포함)
POST /api/v1/boards/{board_id}/move

// 요청 바디
{
  "view_id": "view-123",
  "group_by_field_id": "status-field-id",  // 칸반인 경우
  "new_field_value": "in-progress-option-id",  // 새 컬럼
  "before_position": "a0",  // 이전 보드의 position
  "after_position": "a1"    // 다음 보드의 position
}

// 응답
{
  "board_id": "board-123",
  "new_position": "a0V",  // 자동 생성된 새 position
  "message": "Board moved successfully"
}
```

자세한 내용은 [FRONTEND_API_GUIDE.md](./FRONTEND_API_GUIDE.md) 참고

#### 방법 2: 일괄 순서 업데이트 (특수한 경우만)

```typescript
// 여러 보드의 순서를 한 번에 업데이트
PUT /api/views/board-order

// 요청 바디
{
  "viewId": "view-123",
  "boardOrders": [
    { "boardId": "board-1", "position": "a0" },
    { "boardId": "board-2", "position": "a1" },
    { "boardId": "board-3", "position": "a2" }
  ]
}
```

**언제 사용하나요?**
- 방법 1 (권장): 일반적인 드래그앤드롭 (99% 경우)
- 방법 2: 전체 순서를 재정렬할 때 (매우 드물게)

---

## API 상세 명세

### 1. 뷰 목록 조회

```
GET /api/views?projectId={project_id}
```

**응답**:
```json
{
  "data": [
    {
      "viewId": "uuid",
      "projectId": "uuid",
      "createdBy": "uuid",
      "name": "뷰 이름",
      "description": "설명",
      "isDefault": false,
      "isShared": true,
      "filters": {},
      "sortBy": "created_at",
      "sortDirection": "desc",
      "groupByFieldId": "",
      "createdAt": "2025-01-10T10:00:00Z",
      "updatedAt": "2025-01-10T10:00:00Z"
    }
  ]
}
```

**참고**:
- `isDefault: true` → 프로젝트 진입 시 자동 선택되는 뷰
- `isShared: true` → 팀 전체가 볼 수 있음, `false` → 본인만
- `groupByFieldId`가 있으면 → 칸반 형태, 없으면 → 리스트 형태

---

### 2. 뷰 적용 (보드 데이터 가져오기)

```
GET /api/views/{view_id}/apply?page=1&limit=20
```

**쿼리 파라미터**:
- `page`: 페이지 번호 (기본: 1)
- `limit`: 페이지당 항목 수 (기본: 20, 최대: 100)

**응답 (그룹핑 없음)**:
```json
{
  "boards": [ /* 보드 배열 */ ],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

**응답 (그룹핑 있음)**:
```json
{
  "groupByField": {
    "fieldId": "uuid",
    "name": "상태",
    "fieldType": "single_select"
  },
  "groups": [
    {
      "groupValue": {
        "option_id": "uuid",
        "label": "할 일",
        "color": "#gray"
      },
      "boards": [ /* 이 그룹의 보드들 */ ],
      "count": 5
    }
  ],
  "total": 100
}
```

---

### 3. 뷰 생성

```
POST /api/views
```

**요청 바디**:
```json
{
  "projectId": "uuid",           // 필수
  "name": "뷰 이름",              // 필수
  "description": "설명",          // 선택
  "isDefault": false,            // 선택 (기본: false)
  "isShared": false,             // 선택 (기본: false)
  "filters": {},                 // 선택
  "sortBy": "created_at",        // 선택
  "sortDirection": "desc",       // 선택 (asc 또는 desc)
  "groupByFieldId": ""           // 선택 (single_select 또는 multi_select 필드 ID)
}
```

**필터 예시**:
```json
{
  "filters": {
    "title": {
      "operator": "contains",
      "value": "버그"
    },
    "field-id-123": {
      "operator": "in",
      "value": ["option-id-1", "option-id-2"]
    }
  }
}
```

**지원하는 연산자**:
- `eq`, `ne`: 같음/다름
- `contains`: 포함 (문자열)
- `in`, `not_in`: 배열 포함
- `gt`, `gte`, `lt`, `lte`: 크기 비교
- `is_null`, `is_not_null`: 값 존재 여부

---

### 4. 뷰 수정

```
PUT /api/views/{view_id}
```

**요청 바디**: 변경할 필드만 포함

```json
{
  "name": "새 이름",
  "filters": { /* 새 필터 */ }
}
```

**주의**: 본인이 만든 뷰만 수정 가능 (`createdBy`가 본인인 경우)

---

### 5. 뷰 삭제

```
DELETE /api/views/{view_id}
```

**주의**: 본인이 만든 뷰만 삭제 가능

---

## 실전 예제 코드

### React + TypeScript 전체 구현 예시

```typescript
import axios from 'axios';
import { useState, useEffect } from 'react';

// 타입 정의
interface View {
  viewId: string;
  name: string;
  isDefault: boolean;
  isShared: boolean;
  filters: Record<string, any>;
  sortBy: string;
  sortDirection: 'asc' | 'desc';
  groupByFieldId: string;
}

interface Board {
  id: string;
  title: string;
  position: string;
  custom_fields: Record<string, any>;
}

interface BoardGroup {
  groupValue: {
    option_id: string;
    label: string;
    color: string;
  };
  boards: Board[];
  count: number;
}

// API 클라이언트
class ViewAPI {
  private baseURL = process.env.REACT_APP_API_URL;

  // 뷰 목록 조회
  async fetchViews(projectId: string): Promise<View[]> {
    const response = await axios.get(`${this.baseURL}/api/views`, {
      params: { projectId }
    });
    return response.data.data;
  }

  // 뷰 적용
  async applyView(viewId: string, page = 1, limit = 20) {
    const response = await axios.get(
      `${this.baseURL}/api/views/${viewId}/apply`,
      { params: { page, limit } }
    );
    return response.data;
  }

  // 뷰 생성
  async createView(data: Partial<View>): Promise<View> {
    const response = await axios.post(`${this.baseURL}/api/views`, data);
    return response.data.data;
  }

  // 뷰 수정
  async updateView(viewId: string, data: Partial<View>): Promise<View> {
    const response = await axios.put(
      `${this.baseURL}/api/views/${viewId}`,
      data
    );
    return response.data.data;
  }

  // 뷰 삭제
  async deleteView(viewId: string): Promise<void> {
    await axios.delete(`${this.baseURL}/api/views/${viewId}`);
  }
}

// React 컴포넌트
function ProjectBoardPage({ projectId }: { projectId: string }) {
  const [views, setViews] = useState<View[]>([]);
  const [currentView, setCurrentView] = useState<View | null>(null);
  const [boards, setBoards] = useState<Board[]>([]);
  const [groups, setGroups] = useState<BoardGroup[]>([]);
  const [isKanban, setIsKanban] = useState(false);

  const api = new ViewAPI();

  // 프로젝트 진입 시 뷰 목록 로드
  useEffect(() => {
    loadViews();
  }, [projectId]);

  async function loadViews() {
    const viewList = await api.fetchViews(projectId);
    setViews(viewList);

    // 기본 뷰 선택
    const defaultView = viewList.find(v => v.isDefault) || viewList[0];
    if (defaultView) {
      selectView(defaultView);
    }
  }

  // 뷰 선택
  async function selectView(view: View) {
    setCurrentView(view);

    // 뷰 적용
    const result = await api.applyView(view.viewId);

    // 그룹핑 여부 확인
    if (result.groups) {
      // 칸반 형태
      setIsKanban(true);
      setGroups(result.groups);
    } else {
      // 리스트 형태
      setIsKanban(false);
      setBoards(result.boards);
    }
  }

  // 뷰 생성
  async function createNewView() {
    const newView = await api.createView({
      projectId,
      name: '새 뷰',
      description: '',
      isDefault: false,
      isShared: false,
      filters: {},
      sortBy: 'created_at',
      sortDirection: 'desc',
      groupByFieldId: ''
    });

    // 뷰 목록 갱신
    await loadViews();

    // 새 뷰 선택
    selectView(newView);
  }

  return (
    <div>
      {/* 뷰 선택 드롭다운 */}
      <select
        value={currentView?.viewId}
        onChange={(e) => {
          const view = views.find(v => v.viewId === e.target.value);
          if (view) selectView(view);
        }}
      >
        {views.map(view => (
          <option key={view.viewId} value={view.viewId}>
            {view.name} {view.isDefault && '(기본)'}
          </option>
        ))}
      </select>

      <button onClick={createNewView}>새 뷰 만들기</button>

      {/* 보드 표시 */}
      {isKanban ? (
        <KanbanBoard groups={groups} />
      ) : (
        <BoardList boards={boards} />
      )}
    </div>
  );
}

// 칸반 보드 컴포넌트
function KanbanBoard({ groups }: { groups: BoardGroup[] }) {
  return (
    <div style={{ display: 'flex', gap: '16px' }}>
      {groups.map(group => (
        <div key={group.groupValue.option_id} style={{
          minWidth: '300px',
          backgroundColor: '#f5f5f5',
          padding: '16px',
          borderRadius: '8px'
        }}>
          <h3 style={{ color: group.groupValue.color }}>
            {group.groupValue.label} ({group.count})
          </h3>
          {group.boards
            .sort((a, b) => a.position.localeCompare(b.position))
            .map(board => (
              <div key={board.id} style={{
                backgroundColor: 'white',
                padding: '12px',
                marginTop: '8px',
                borderRadius: '4px'
              }}>
                {board.title}
              </div>
            ))}
        </div>
      ))}
    </div>
  );
}

// 리스트 컴포넌트
function BoardList({ boards }: { boards: Board[] }) {
  return (
    <div>
      {boards
        .sort((a, b) => a.position.localeCompare(b.position))
        .map(board => (
          <div key={board.id} style={{
            padding: '12px',
            borderBottom: '1px solid #eee'
          }}>
            {board.title}
          </div>
        ))}
    </div>
  );
}
```

---

## 자주 묻는 질문

### Q1: View와 Board의 차이는 뭔가요?

**A**:
- **Board**: 실제 작업 데이터 (예: "로그인 기능 구현" 작업)
- **View**: 보드를 보는 방식/설정 (예: "내가 담당한 작업만 보기")

비유:
- Board = 책상 위의 서류들
- View = 서류를 정리하는 방법 (날짜순, 중요한 것만, 카테고리별로 묶기 등)

---

### Q2: 필터는 어떻게 만드나요?

**A**: 커스텀 필드의 ID와 값으로 필터 객체를 만듭니다.

```typescript
// 예시: 상태가 "진행중"이고 담당자가 나인 보드만
{
  "filters": {
    "status-field-id": {
      "operator": "eq",
      "value": "in-progress-option-id"
    },
    "assignee-field-id": {
      "operator": "eq",
      "value": "my-user-id"
    }
  }
}
```

**프론트에서 구현 팁**:
1. 프로젝트의 커스텀 필드 목록 가져오기 (Project Init API)
2. 사용자가 UI에서 필드 선택 → 연산자 선택 → 값 입력
3. 위 형식의 JSON 객체로 변환

---

### Q3: 칸반 보드는 어떻게 만드나요?

**A**: `groupByFieldId`에 single_select 또는 multi_select 필드를 지정하면 됩니다.

```typescript
// 예시: 상태 필드로 그룹핑 (칸반)
{
  "groupByFieldId": "status-field-id"  // 상태 필드 ID
}
```

그러면 응답이 `groups` 배열로 오고, 각 그룹이 칸반의 컬럼이 됩니다.

---

### Q4: 뷰마다 보드 순서가 다를 수 있나요?

**A**: 네! 뷰별, 사용자별로 순서가 다릅니다.

```
"전체 보드" 뷰에서 철수의 순서:
보드A (position: a0)
보드B (position: a1)
보드C (position: a2)

"내 작업" 뷰에서 철수의 순서:
보드C (position: x0)  ← 다른 position!
보드A (position: x1)
```

같은 뷰를 봐도 영희는 다른 순서로 볼 수 있습니다.

---

### Q5: isDefault와 isShared의 차이는?

**A**:
- **isDefault**: 프로젝트 진입 시 자동으로 선택되는 뷰 (프로젝트당 1개만)
- **isShared**: 팀 전체에게 보이는 뷰인지 여부
  - `true`: 팀원 모두 볼 수 있음
  - `false`: 나만 볼 수 있음 (개인 뷰)

---

### Q6: 페이지네이션은 어떻게 하나요?

**A**: `page`와 `limit` 쿼리 파라미터 사용

```typescript
// 1페이지 (1~20번 보드)
GET /api/views/{view_id}/apply?page=1&limit=20

// 2페이지 (21~40번 보드)
GET /api/views/{view_id}/apply?page=2&limit=20
```

무한 스크롤 구현:
```typescript
let page = 1;
const limit = 20;
let allBoards = [];

async function loadMore() {
  const result = await api.applyView(viewId, page, limit);
  allBoards = [...allBoards, ...result.boards];
  page++;
}
```

---

### Q7: 에러 처리는 어떻게 하나요?

**A**: 주요 에러 코드:

```typescript
try {
  await api.applyView(viewId);
} catch (error) {
  if (error.response?.status === 404) {
    alert('뷰를 찾을 수 없습니다');
  } else if (error.response?.status === 403) {
    alert('접근 권한이 없습니다');
  } else {
    alert('오류가 발생했습니다');
  }
}
```

---

## 요약

### 핵심 API 3개만 기억하세요!

1. **뷰 목록 조회**: `GET /api/views?projectId=xxx`
   - 프로젝트 진입 시 1번만

2. **뷰 적용**: `GET /api/views/{view_id}/apply`
   - 뷰 선택할 때마다

3. **뷰 생성**: `POST /api/views`
   - 새 뷰 만들 때

### 화면별 호출 순서

```
1. 프로젝트 진입
   └─ GET /api/views?projectId=xxx
       └─ GET /api/views/{view_id}/apply  (기본 뷰 또는 첫 번째 뷰)

2. 뷰 변경
   └─ GET /api/views/{view_id}/apply

3. 뷰 생성
   └─ POST /api/views
       └─ GET /api/views/{new_view_id}/apply
```

### 응답 형태 구분법

```typescript
const result = await applyView(viewId);

if (result.groups) {
  // 칸반 형태 → groups 배열 렌더링
  renderKanban(result.groups);
} else {
  // 리스트 형태 → boards 배열 렌더링
  renderList(result.boards);
}
```

---

더 궁금한 점이 있으면 백엔드 팀에게 문의하세요!
