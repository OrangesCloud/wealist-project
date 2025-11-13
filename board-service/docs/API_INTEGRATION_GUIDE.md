# API 통합 가이드 - 프로젝트 페이지 로딩부터 보드 표시까지

> **이 문서의 목적**: Project Init Settings API와 View API를 어떻게 함께 사용하는지, 언제 무엇을 호출해야 하는지 명확하게 설명합니다.

---

## 핵심 요약

**프로젝트 페이지를 로딩할 때 4단계**:

```
1. Project Init Settings API 호출
   → 프로젝트 설정 데이터 (프로젝트 정보, 필드 정의, 필드 타입)

2. Members API 호출
   → 프로젝트 멤버 목록 (담당자 할당용)

3. View List API 호출
   → 사용 가능한 뷰 목록

4. View Apply API 호출
   → 실제 보드 데이터 (필터링/정렬/페이징)
```

---

## 목차

1. [API 역할 구분](#api-역할-구분)
2. [프로젝트 페이지 로딩 전체 흐름](#프로젝트-페이지-로딩-전체-흐름)
3. [API 호출 순서도](#api-호출-순서도)
4. [각 API의 역할](#각-api의-역할)
5. [전체 코드 예시](#전체-코드-예시)
6. [성능 최적화 팁](#성능-최적화-팁)
7. [FAQ](#faq)

---

## API 역할 구분

### 1. Project Init Settings API

```
GET /api/projects/{projectId}/init-settings
```

**역할**: 프로젝트의 **정적 설정 데이터** 가져오기

**무엇을 가져오나**:
- ✅ 프로젝트 기본 정보 (이름, 설명, 소유자 등)
- ✅ **필드 정의** (상태, 우선순위 등 커스텀 필드 + 옵션)
- ✅ 필드 타입 정보 (새 필드 만들 때 사용)
- ✅ 기본 뷰 ID

**언제 호출**: 프로젝트 진입 시 **1회만**

**왜 필요한가**:
- 필드 정의 없이는 보드의 커스텀 필드 값을 해석할 수 없음
- 프로젝트 기본 정보 필요
- 변하지 않는 설정값만 포함

---

### 2. Members API

```
GET /api/projects/{projectId}/members
```

**역할**: 프로젝트 **멤버 목록** 가져오기

**무엇을 가져오나**:
- ✅ 멤버 목록 (ID, 이름, 이메일, 역할)

**언제 호출**: 프로젝트 진입 시 **1회만** (또는 멤버 변경 시)

**왜 필요한가**:
- 담당자 할당 드롭다운용
- 멤버는 동적으로 변할 수 있어 별도 API로 분리

---

### 3. View List API

```
GET /api/views?projectId={projectId}
```

**역할**: 사용자가 선택할 수 있는 **뷰 목록** 가져오기

**무엇을 가져오나**:
- ✅ 뷰 목록 (이름, 필터, 정렬, 그룹핑 설정)
- ✅ 각 뷰의 isDefault, isShared 정보

**언제 호출**: 프로젝트 진입 시 **1회만** (또는 뷰 생성/삭제 후)

**왜 필요한가**:
- 사용자가 선택할 뷰 목록을 드롭다운에 표시
- 기본 뷰 찾기

---

### 4. View Apply API

```
GET /api/views/{viewId}/apply?page=1&limit=20
```

**역할**: 선택한 뷰의 설정대로 **실제 보드 데이터** 가져오기

**무엇을 가져오나**:
- ✅ 필터링/정렬/그룹핑된 보드 데이터
- ✅ 페이지네이션 지원 (20개씩)
- ✅ 각 보드의 position 정보 (뷰별 순서)

**언제 호출**: **뷰를 선택/변경할 때마다**

**왜 필요한가**:
- 보드 데이터는 View API를 통해서만 조회 (Init Settings에는 보드 없음)
- 페이지네이션 지원
- 뷰별 순서, 필터, 그룹핑 적용

---

## 프로젝트 페이지 로딩 전체 흐름

### 시나리오: 사용자가 프로젝트 페이지에 처음 진입

```
┌─────────────────────────────────────────────────────────┐
│ 1. 프로젝트 진입                                          │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 2. Project Init Settings API 호출                        │
│    GET /api/projects/{projectId}/init-settings          │
│                                                          │
│    응답:                                                 │
│    - project (프로젝트 정보)                              │
│    - fields (필드 정의)   ← 전역 상태에 저장!             │
│    - fieldTypes           ← 전역 상태에 저장!             │
│    - defaultViewId                                       │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Members API 호출                                      │
│    GET /api/projects/{projectId}/members                │
│                                                          │
│    응답:                                                 │
│    - members[] (멤버 목록)  ← 전역 상태에 저장!           │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 4. View List API 호출                                    │
│    GET /api/views?projectId={projectId}                 │
│                                                          │
│    응답:                                                 │
│    - views[] (뷰 목록)                                   │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 5. 기본 뷰 찾기                                           │
│    - defaultViewId에 해당하는 뷰 찾기                     │
│    - 없으면 첫 번째 뷰 선택                               │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 6. View Apply API 호출                                   │
│    GET /api/views/{viewId}/apply?page=1&limit=20        │
│                                                          │
│    응답:                                                 │
│    - boards[] (필터링/정렬된 보드 20개)                   │
│    또는                                                  │
│    - groups[] (그룹핑된 보드들, 칸반용)                   │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 7. 보드 렌더링                                            │
│    - Step 2의 fields 정의 사용                           │
│    - Step 3의 members 목록 사용                          │
│    - Step 6의 boards 데이터 사용                         │
│    - 커스텀 필드 값 해석                                  │
└─────────────────────────────────────────────────────────┘
```

---

## API 호출 순서도

### 초기 로딩

```typescript
// 1. Project Init Settings API (프로젝트 설정)
const initSettings = await getProjectInitSettings(projectId);
// → fields, project, fieldTypes 저장

// 2. Members API (멤버 목록)
const members = await getProjectMembers(projectId);
// → members 저장

// 3. View List API (뷰 목록)
const views = await getViews(projectId);

// 4. 기본 뷰 선택
const defaultView = views.find(v => v.viewId === initSettings.defaultViewId)
                    || views[0];

// 5. View Apply API (보드 데이터)
const boardData = await applyView(defaultView.viewId, 1, 20);

// 6. 렌더링
render(boardData.boards, initSettings.fields, members);
```

### 뷰 변경 시

```typescript
// 뷰 변경 시에는 View Apply API만 호출!
async function onViewChange(newViewId: string) {
  const boardData = await applyView(newViewId, 1, 20);
  render(boardData.boards, cachedFields); // fields는 캐시된 거 사용
}
```

---

## 각 API의 역할

### Project Init Settings API - "프로젝트 설정 데이터"

**비유**: 게임의 "설정 파일" 또는 "스키마 정의"

```typescript
const initSettings = await getProjectInitSettings(projectId);

// 프로젝트 정보
console.log(initSettings.project.name); // "웹사이트 리뉴얼"

// 필드 정의 (가장 중요!)
initSettings.fields.forEach(field => {
  console.log(field.name);      // "상태"
  console.log(field.fieldType); // "single_select"
  console.log(field.options);   // [{ label: "할 일", color: "#gray" }, ...]
});

// 이 데이터들은 전역 상태에 저장하고 계속 재사용!
```

**이 데이터 어디에 사용?**:
- 필드 정의 → 보드 커스텀 필드 값 해석
- 필드 타입 → 새 필드 만들기 UI
- 프로젝트 정보 → 헤더 표시

### Members API - "프로젝트 멤버 데이터"

```typescript
const members = await getProjectMembers(projectId);

// 멤버 목록
members.forEach(member => {
  console.log(member.name);  // "홍길동"
  console.log(member.role);  // "ADMIN"
});
```

**이 데이터 어디에 사용?**:
- 담당자 할당 드롭다운
- 멤버 표시

---

### View Apply API - "실제 보드 데이터"

**비유**: 게임의 "실제 플레이 데이터" 또는 "쿼리 결과"

```typescript
const viewData = await applyView(viewId, 1, 20);

// 보드 데이터만 있음
viewData.boards.forEach(board => {
  console.log(board.title);           // "로그인 기능 구현"
  console.log(board.custom_fields);   // { "field-id-123": "option-id-456" }
  console.log(board.position);        // "a0"
});

// 필드 정의는 없음! → Project Init에서 가져온 것 사용
const fieldDef = cachedFields["field-id-123"];
const optionDef = fieldDef.options.find(o => o.optionId === "option-id-456");
console.log(optionDef.label); // "진행중"
```

---

## 전체 코드 예시

### React + TypeScript 완전 구현

```typescript
import { useState, useEffect } from 'react';
import axios from 'axios';

// ===== 타입 정의 =====

interface ProjectMetadata {
  project: ProjectInfo;
  fields: Field[];
  members: Member[];
  fieldTypes: FieldType[];
}

interface Field {
  fieldId: string;
  name: string;
  fieldType: string;
  options: Option[];
}

interface Option {
  optionId: string;
  label: string;
  color: string;
}

interface Member {
  userId: string;
  name: string;
  email: string;
  role: string;
}

interface View {
  viewId: string;
  name: string;
  isDefault: boolean;
  isShared: boolean;
  groupByFieldId: string;
}

interface Board {
  id: string;
  title: string;
  custom_fields: Record<string, any>;
  position: string;
}

// ===== API 함수 =====

async function getProjectInitSettings(projectId: string) {
  const response = await axios.get(`/api/projects/${projectId}/init-settings`);
  return response.data.data;
}

async function getProjectMembers(projectId: string) {
  const response = await axios.get(`/api/projects/${projectId}/members`);
  return response.data.data;
}

async function getViews(projectId: string) {
  const response = await axios.get('/api/views', {
    params: { projectId }
  });
  return response.data.data;
}

async function applyView(viewId: string, page = 1, limit = 20) {
  const response = await axios.get(`/api/views/${viewId}/apply`, {
    params: { page, limit }
  });
  return response.data;
}

// ===== 메인 컴포넌트 =====

function ProjectPage({ projectId }: { projectId: string }) {
  // 전역 상태 (한 번만 로드)
  const [metadata, setMetadata] = useState<ProjectMetadata | null>(null);
  const [views, setViews] = useState<View[]>([]);

  // 현재 상태
  const [currentView, setCurrentView] = useState<View | null>(null);
  const [boards, setBoards] = useState<Board[]>([]);
  const [groups, setGroups] = useState<any[]>([]);
  const [isKanban, setIsKanban] = useState(false);

  // 로딩 상태
  const [isLoading, setIsLoading] = useState(true);

  // 프로젝트 초기화 (컴포넌트 마운트 시 1회만)
  useEffect(() => {
    initializeProject();
  }, [projectId]);

  async function initializeProject() {
    setIsLoading(true);

    try {
      // Step 1: Project Init Settings API - 설정 데이터 로드
      console.log('📡 Loading project settings...');
      const initSettings = await getProjectInitSettings(projectId);

      // Step 2: Members API - 멤버 목록 로드
      console.log('📡 Loading members...');
      const memberList = await getProjectMembers(projectId);

      setMetadata({
        project: initSettings.project,
        fields: initSettings.fields,
        members: memberList,
        fieldTypes: initSettings.fieldTypes
      });

      console.log('✅ Settings and members loaded:', {
        fields: initSettings.fields.length,
        members: memberList.length
      });

      // Step 3: View List API - 뷰 목록 로드
      console.log('📡 Loading views...');
      const viewList = await getViews(projectId);
      setViews(viewList);

      console.log('✅ Views loaded:', viewList.length);

      // Step 4: 기본 뷰 찾기
      const defaultView = viewList.find(v => v.viewId === initSettings.defaultViewId)
                          || viewList[0];

      if (defaultView) {
        console.log('🎯 Applying default view:', defaultView.name);
        await loadViewData(defaultView, initSettings.fields);
      }

    } catch (error) {
      console.error('❌ Failed to initialize project:', error);
    } finally {
      setIsLoading(false);
    }
  }

  // 뷰 데이터 로드
  async function loadViewData(view: View, fields: Field[]) {
    setCurrentView(view);

    try {
      // Step 4: View Apply API - 보드 데이터 로드
      console.log('📡 Loading boards for view:', view.name);
      const viewData = await applyView(view.viewId, 1, 20);

      if (viewData.groups) {
        // 칸반 형태
        console.log('✅ Kanban view loaded:', viewData.groups.length, 'groups');
        setIsKanban(true);
        setGroups(viewData.groups);
      } else {
        // 리스트 형태
        console.log('✅ List view loaded:', viewData.boards.length, 'boards');
        setIsKanban(false);
        setBoards(viewData.boards);
      }

    } catch (error) {
      console.error('❌ Failed to load view data:', error);
    }
  }

  // 뷰 변경 핸들러
  async function handleViewChange(viewId: string) {
    const view = views.find(v => v.viewId === viewId);
    if (view && metadata) {
      await loadViewData(view, metadata.fields);
    }
  }

  // 필드 값 해석 헬퍼 함수
  function getFieldValue(board: Board, fieldId: string) {
    if (!metadata) return null;

    const field = metadata.fields.find(f => f.fieldId === fieldId);
    if (!field) return null;

    const value = board.custom_fields[fieldId];
    if (!value) return null;

    // single_select인 경우 option 정보 찾기
    if (field.fieldType === 'single_select') {
      const option = field.options.find(o => o.optionId === value);
      return option ? option.label : value;
    }

    return value;
  }

  if (isLoading) {
    return <div>로딩 중...</div>;
  }

  if (!metadata) {
    return <div>프로젝트를 불러오는데 실패했습니다.</div>;
  }

  return (
    <div>
      {/* 프로젝트 헤더 */}
      <header>
        <h1>{metadata.project.name}</h1>
        <p>{metadata.project.description}</p>
      </header>

      {/* 뷰 선택 드롭다운 */}
      <div className="view-selector">
        <select
          value={currentView?.viewId}
          onChange={(e) => handleViewChange(e.target.value)}
        >
          {views.map(view => (
            <option key={view.viewId} value={view.viewId}>
              {view.name} {view.isDefault && '(기본)'}
            </option>
          ))}
        </select>
      </div>

      {/* 보드 표시 */}
      {isKanban ? (
        <KanbanView
          groups={groups}
          fields={metadata.fields}
        />
      ) : (
        <ListView
          boards={boards}
          fields={metadata.fields}
          getFieldValue={getFieldValue}
        />
      )}
    </div>
  );
}

// ===== 리스트 뷰 컴포넌트 =====

function ListView({
  boards,
  fields,
  getFieldValue
}: {
  boards: Board[];
  fields: Field[];
  getFieldValue: (board: Board, fieldId: string) => any;
}) {
  return (
    <div className="list-view">
      <table>
        <thead>
          <tr>
            <th>제목</th>
            {fields.map(field => (
              <th key={field.fieldId}>{field.name}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {boards
            .sort((a, b) => a.position.localeCompare(b.position))
            .map(board => (
              <tr key={board.id}>
                <td>{board.title}</td>
                {fields.map(field => (
                  <td key={field.fieldId}>
                    {getFieldValue(board, field.fieldId)}
                  </td>
                ))}
              </tr>
            ))}
        </tbody>
      </table>
    </div>
  );
}

// ===== 칸반 뷰 컴포넌트 =====

function KanbanView({
  groups,
  fields
}: {
  groups: any[];
  fields: Field[];
}) {
  return (
    <div className="kanban-view" style={{ display: 'flex', gap: '16px' }}>
      {groups.map(group => (
        <div
          key={group.groupValue.option_id}
          className="kanban-column"
          style={{
            minWidth: '300px',
            backgroundColor: '#f5f5f5',
            padding: '16px',
            borderRadius: '8px'
          }}
        >
          <h3 style={{ color: group.groupValue.color }}>
            {group.groupValue.label} ({group.count})
          </h3>
          {group.boards
            .sort((a: Board, b: Board) => a.position.localeCompare(b.position))
            .map((board: Board) => (
              <div
                key={board.id}
                style={{
                  backgroundColor: 'white',
                  padding: '12px',
                  marginTop: '8px',
                  borderRadius: '4px',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                }}
              >
                <h4>{board.title}</h4>
                {/* 커스텀 필드 표시 */}
                {fields.slice(0, 3).map(field => {
                  const value = board.custom_fields[field.fieldId];
                  if (!value) return null;

                  return (
                    <div key={field.fieldId} style={{ fontSize: '12px', marginTop: '4px' }}>
                      <strong>{field.name}:</strong> {value}
                    </div>
                  );
                })}
              </div>
            ))}
        </div>
      ))}
    </div>
  );
}

export default ProjectPage;
```

---

## 성능 최적화 팁

### 1. 메타데이터 캐싱

```typescript
// ✅ 좋은 예: 전역 상태에 저장 (Redux, Context, Zustand 등)
const [metadata, setMetadata] = useState<ProjectMetadata | null>(null);

// Project Init은 1회만 호출
useEffect(() => {
  if (!metadata) {
    loadMetadata();
  }
}, [projectId]);

// 뷰 변경 시 메타데이터는 재사용
async function changeView(viewId: string) {
  const boardData = await applyView(viewId);
  renderBoards(boardData, metadata); // 캐시된 메타데이터 사용
}
```

### 2. 불필요한 API 호출 방지

```typescript
// ❌ 나쁜 예: 뷰 변경할 때마다 Project Init 호출
async function changeView(viewId: string) {
  const initData = await getProjectInitData(projectId); // 불필요!
  const viewData = await applyView(viewId);
  render(viewData, initData.fields);
}

// ✅ 좋은 예: 캐시된 메타데이터 사용
async function changeView(viewId: string) {
  const viewData = await applyView(viewId);
  render(viewData, cachedFields); // 이미 로드된 필드 사용
}
```

### 3. 병렬 호출

```typescript
// ✅ Init Settings, Members, Views를 병렬로 호출
async function initializeProject() {
  const [initSettings, members, viewList] = await Promise.all([
    getProjectInitSettings(projectId),
    getProjectMembers(projectId),
    getViews(projectId)
  ]);

  // 메타데이터 저장
  setMetadata({
    project: initSettings.project,
    fields: initSettings.fields,
    members: members,
    fieldTypes: initSettings.fieldTypes
  });
  setViews(viewList);

  // 기본 뷰 적용
  const defaultView = viewList.find(v => v.viewId === initSettings.defaultViewId);
  if (defaultView) {
    await loadViewData(defaultView);
  }
}
```

### 4. 페이지네이션 활용

```typescript
// ✅ 무한 스크롤 구현
let currentPage = 1;
const limit = 20;

async function loadMore() {
  currentPage++;
  const viewData = await applyView(viewId, currentPage, limit);
  setBoards(prev => [...prev, ...viewData.boards]);
}
```

---

## FAQ

### Q1: 왜 Init Settings와 Members를 분리했나요?

**A**: 데이터 특성이 다르기 때문입니다:
- **Init Settings**: 정적 설정 데이터 (필드, 타입 등) - 거의 변하지 않음
- **Members**: 동적 데이터 (멤버 추가/삭제 가능) - 자주 변할 수 있음

분리하면:
- 멤버만 변경 시 Settings 재조회 불필요
- 캐싱 전략을 다르게 가져갈 수 있음

### Q2: Init Settings에 보드 데이터가 없나요?

**A**: 네, 이제 없습니다. 보드 데이터는 View Apply API를 통해서만 가져옵니다:
- 필터링/정렬/그룹핑 지원
- 페이지네이션 지원
- 뷰별 순서 관리

### Q3: 매번 모든 API를 다 호출해야 하나요?

**A**: 아니요!

```typescript
// 프로젝트 진입 시 (1회만)
const initSettings = await getProjectInitSettings(projectId); // 1회
const members = await getProjectMembers(projectId);           // 1회
const views = await getViews(projectId);                      // 1회

// 뷰 변경 시 (매번)
const boardData = await applyView(viewId); // 필요할 때마다
```

### Q4: fields 정보가 바뀌면 어떻게 하나요?

**A**: 필드가 추가/수정/삭제되면 Project Init을 다시 호출하거나, Field API를 직접 호출하세요.

```typescript
// 필드 생성 후
await createField(fieldData);

// 필드 목록 갱신
const updatedFields = await getFields(projectId);
setMetadata(prev => ({ ...prev, fields: updatedFields }));
```

### Q5: View Apply가 실패하면 어떻게 하나요?

**A**: 에러 메시지를 표시하세요. Init Settings에는 보드가 없으므로 View API가 필수입니다.

```typescript
try {
  const viewData = await applyView(viewId);
  setBoards(viewData.boards);
} catch (error) {
  showError('뷰를 불러오는데 실패했습니다');
  // 또는 기본 뷰로 재시도
}
```

---

## 요약 체크리스트

프로젝트 페이지 구현 시:

- [ ] Project Init Settings API 호출 (1회)
  - [ ] fields 전역 상태에 저장
  - [ ] project 정보 저장
  - [ ] fieldTypes 저장

- [ ] Members API 호출 (1회)
  - [ ] members 전역 상태에 저장

- [ ] View List API 호출 (1회)
  - [ ] 뷰 목록 저장
  - [ ] 뷰 드롭다운 렌더링

- [ ] View Apply API 호출 (뷰 선택 시마다)
  - [ ] 보드 데이터 렌더링
  - [ ] 캐시된 fields로 커스텀 필드 해석

- [ ] 뷰 변경 시
  - [ ] View Apply만 호출 (Init Settings 재호출 ❌)
  - [ ] 캐시된 메타데이터 재사용

---

## 관련 문서

- [View API 상세](./VIEW_API_GUIDE.md)
- [보드 순서 변경](./ORDER_UPDATE_GUIDE.md)
- [Fractional Indexing](./FRONTEND_API_GUIDE.md)
- [Swagger API 문서](http://localhost:8000/swagger/index.html)
