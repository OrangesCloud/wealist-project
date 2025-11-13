# 순서/위치 변경 API 가이드

> **이 문서의 목적**: 프론트엔드에서 보드, 필드, 옵션의 순서를 변경할 때 어떤 API를 호출해야 하는지 명확하게 설명합니다.

---

## 핵심 요약

```
┌─────────────────────────────────────────────────────────┐
│ 순서를 변경할 수 있는 것들                                 │
├─────────────────────────────────────────────────────────┤
│ 1. 보드 순서 (드래그앤드롭)                                │
│    → POST /api/v1/boards/{board_id}/move                │
│                                                          │
│ 2. 커스텀 필드 순서 (표시 순서)                            │
│    → PUT /api/fields/{field_id}                         │
│                                                          │
│ 3. 필드 옵션 순서 (드롭다운 순서)                          │
│    → PUT /api/field-options/{option_id}                 │
└─────────────────────────────────────────────────────────┘
```

---

## 목차

1. [보드 순서 변경 (드래그앤드롭)](#보드-순서-변경-드래그앤드롭)
2. [커스텀 필드 순서 변경](#커스텀-필드-순서-변경)
3. [필드 옵션 순서 변경](#필드-옵션-순서-변경)
4. [전체 시나리오별 API 정리](#전체-시나리오별-api-정리)
5. [실전 코드 예시](#실전-코드-예시)
6. [주의사항](#주의사항)

---

## 보드 순서 변경 (드래그앤드롭)

### 개념

**보드 순서는 뷰별, 사용자별로 다릅니다!**

```
같은 뷰를 봐도:
- 철수의 보드 순서: A, B, C
- 영희의 보드 순서: C, A, B (다를 수 있음)

같은 사용자가 다른 뷰를 보면:
- "전체 보드" 뷰: A, B, C
- "내 작업" 뷰: B, A, C (다를 수 있음)
```

**저장 위치**: `user_board_order` 테이블
- `view_id`: 어떤 뷰에서
- `user_id`: 어떤 사용자가
- `board_id`: 어떤 보드를
- `position`: 어떤 위치에 (fractional indexing)

---

### API: 보드 이동 (Move Board)

#### 엔드포인트

```
POST /api/v1/boards/{board_id}/move
```

#### 언제 호출?

- ✅ 사용자가 보드를 드래그해서 위치를 바꿨을 때
- ✅ 칸반 보드에서 다른 컬럼으로 이동했을 때
- ✅ 리스트에서 순서를 바꿨을 때

#### 요청 바디

```json
{
  "view_id": "view-uuid",                    // 필수: 현재 보고 있는 뷰
  "group_by_field_id": "status-field-uuid",  // 필수: 칸반인 경우 그룹핑 필드
  "new_field_value": "in-progress-uuid",     // 필수: 새 컬럼(옵션)의 ID
  "before_position": "a0",                   // 선택: 이전 보드의 position
  "after_position": "a1"                     // 선택: 다음 보드의 position
}
```

#### 응답

```json
{
  "board_id": "board-uuid",
  "new_field_value": "in-progress-uuid",
  "new_position": "a0V",  // 자동 계산된 새 position
  "message": "Board moved successfully"
}
```

---

### 시나리오 1: 리스트에서 보드 순서 변경

**상황**: 사용자가 "내 작업" 뷰 (리스트 형태)에서 보드를 드래그

```
현재 순서:
├─ 보드A (position: a0)
├─ 보드B (position: a1) ← 이걸 맨 아래로
└─ 보드C (position: a2)

원하는 순서:
├─ 보드A (position: a0)
├─ 보드C (position: a2)
└─ 보드B (position: ???) ← 새로 계산
```

**API 호출**:

```typescript
// 보드B를 보드C 아래로 이동
await axios.post(`/api/v1/boards/${boardB.id}/move`, {
  view_id: "my-tasks-view-uuid",
  group_by_field_id: "",  // 리스트 뷰는 그룹핑 없음
  new_field_value: "",    // 필드 값 변경 없음
  before_position: "a2",  // 보드C의 position
  after_position: null    // 맨 아래니까 null
});

// 응답:
// {
//   "board_id": "board-b-uuid",
//   "new_position": "a3"  // 자동 계산됨
// }
```

**프론트엔드 코드**:

```typescript
function onBoardDropInList(
  draggedBoardId: string,
  targetIndex: number,
  allBoards: Board[]
) {
  // 타겟 위치의 앞뒤 보드 찾기
  const beforeBoard = targetIndex > 0 ? allBoards[targetIndex - 1] : null;
  const afterBoard = targetIndex < allBoards.length ? allBoards[targetIndex] : null;

  // API 호출
  const response = await axios.post(`/api/v1/boards/${draggedBoardId}/move`, {
    view_id: currentViewId,
    group_by_field_id: "",  // 리스트 뷰
    new_field_value: "",    // 값 변경 없음
    before_position: beforeBoard?.position || null,
    after_position: afterBoard?.position || null
  });

  // UI 업데이트
  updateBoardPosition(draggedBoardId, response.data.new_position);
}
```

---

### 시나리오 2: 칸반 보드에서 같은 컬럼 내 순서 변경

**상황**: "할 일" 컬럼 내에서 보드 순서 변경

```
"할 일" 컬럼:
├─ 보드A (position: a0)
├─ 보드B (position: a1) ← 이걸 아래로
└─ 보드C (position: a2)
```

**API 호출**:

```typescript
await axios.post(`/api/v1/boards/${boardB.id}/move`, {
  view_id: "kanban-view-uuid",
  group_by_field_id: "status-field-uuid",
  new_field_value: "todo-option-uuid",  // 같은 컬럼 (변경 없음)
  before_position: "a2",  // 보드C의 position
  after_position: null    // 맨 아래
});
```

---

### 시나리오 3: 칸반 보드에서 다른 컬럼으로 이동

**상황**: "할 일" → "진행중" 컬럼으로 이동

```
"할 일" 컬럼:
├─ 보드A (position: a0)
├─ 보드B (position: a1) ← 이걸 "진행중"으로
└─ 보드C (position: a2)

"진행중" 컬럼:
├─ 보드D (position: b0)
└─ 보드E (position: b1)
```

**API 호출**:

```typescript
// 보드B를 "진행중" 컬럼의 맨 위로 이동
await axios.post(`/api/v1/boards/${boardB.id}/move`, {
  view_id: "kanban-view-uuid",
  group_by_field_id: "status-field-uuid",        // 상태 필드
  new_field_value: "in-progress-option-uuid",    // "진행중" 옵션 ID
  before_position: null,   // 맨 위
  after_position: "b0"     // 보드D의 position
});

// 이 API는 두 가지를 동시에 수행:
// 1. 보드B의 상태 필드 값을 "진행중"으로 변경
// 2. 보드B의 position을 "aV" (b0보다 앞) 로 설정
```

**프론트엔드 코드 (칸반)**:

```typescript
function onBoardDropInKanban(
  draggedBoardId: string,
  targetColumnId: string,    // 새 컬럼의 option ID
  targetIndex: number,
  columnBoards: Board[]      // 타겟 컬럼의 보드들
) {
  const beforeBoard = targetIndex > 0 ? columnBoards[targetIndex - 1] : null;
  const afterBoard = targetIndex < columnBoards.length ? columnBoards[targetIndex] : null;

  const response = await axios.post(`/api/v1/boards/${draggedBoardId}/move`, {
    view_id: currentViewId,
    group_by_field_id: groupByFieldId,  // "상태" 필드 ID
    new_field_value: targetColumnId,    // 새 컬럼의 option ID
    before_position: beforeBoard?.position || null,
    after_position: afterBoard?.position || null
  });

  // UI 업데이트
  updateBoard(draggedBoardId, {
    position: response.data.new_position,
    custom_fields: {
      [groupByFieldId]: targetColumnId  // 상태 필드 값도 변경됨
    }
  });
}
```

---

### before_position / after_position 계산법

**핵심 규칙**:
- **맨 앞에 삽입**: `before_position: null, after_position: 첫번째보드.position`
- **맨 뒤에 삽입**: `before_position: 마지막보드.position, after_position: null`
- **중간에 삽입**: `before_position: 앞보드.position, after_position: 뒤보드.position`
- **빈 컬럼**: `before_position: null, after_position: null`

**코드로 구현**:

```typescript
function calculatePositions(targetIndex: number, boards: Board[]) {
  // 빈 리스트
  if (boards.length === 0) {
    return { before_position: null, after_position: null };
  }

  // 맨 앞 (index 0)
  if (targetIndex === 0) {
    return {
      before_position: null,
      after_position: boards[0].position
    };
  }

  // 맨 뒤
  if (targetIndex >= boards.length) {
    return {
      before_position: boards[boards.length - 1].position,
      after_position: null
    };
  }

  // 중간
  return {
    before_position: boards[targetIndex - 1].position,
    after_position: boards[targetIndex].position
  };
}
```

---

## 커스텀 필드 순서 변경

### 개념

**커스텀 필드 순서는 프로젝트 전체에 적용됩니다.**

```
필드 순서:
1. 상태 (display_order: 0)
2. 우선순위 (display_order: 1)
3. 담당자 (display_order: 2)

→ 테이블, 칸반, 상세 페이지 모두 이 순서로 표시
```

**저장 위치**: `project_fields` 테이블의 `display_order` 컬럼

---

### API: 필드 수정

#### 엔드포인트

```
PUT /api/fields/{field_id}
```

#### 언제 호출?

- ✅ 사용자가 필드 설정에서 필드 순서를 변경했을 때
- ✅ 드래그앤드롭으로 필드 순서를 바꿨을 때

#### 요청 바디

```json
{
  "display_order": 2  // 새로운 순서
}
```

#### 응답

```json
{
  "data": {
    "field_id": "field-uuid",
    "name": "우선순위",
    "display_order": 2,
    // ... 기타 필드 정보
  }
}
```

---

### 시나리오: 필드 순서 변경

**상황**: 사용자가 필드 설정 화면에서 필드 순서를 변경

```
현재 순서:
0. 상태
1. 우선순위
2. 담당자

원하는 순서:
0. 상태
1. 담당자      ← 위로 올림
2. 우선순위    ← 아래로 내림
```

**방법 1: 개별 업데이트 (비효율적)**

```typescript
// ❌ 비효율: 각 필드마다 API 호출
await updateField(field2.id, { display_order: 1 }); // 담당자
await updateField(field1.id, { display_order: 2 }); // 우선순위
```

**방법 2: 배치 업데이트 (권장)**

현재 배치 업데이트 API는 없으므로, 필요하다면 순차적으로 호출:

```typescript
async function reorderFields(newOrder: Field[]) {
  // 변경된 필드만 업데이트
  for (let i = 0; i < newOrder.length; i++) {
    const field = newOrder[i];
    if (field.display_order !== i) {
      await axios.put(`/api/fields/${field.field_id}`, {
        display_order: i
      });
    }
  }

  // 필드 목록 갱신
  refreshFields();
}
```

**프론트엔드 코드 (드래그앤드롭)**:

```typescript
function FieldSettingsPage({ projectId }: { projectId: string }) {
  const [fields, setFields] = useState<Field[]>([]);

  // 드래그앤드롭 핸들러
  async function onFieldDrop(draggedFieldId: string, targetIndex: number) {
    // 필드 배열 재정렬
    const newOrder = [...fields];
    const draggedField = newOrder.find(f => f.field_id === draggedFieldId);
    const oldIndex = newOrder.indexOf(draggedField);

    newOrder.splice(oldIndex, 1);
    newOrder.splice(targetIndex, 0, draggedField);

    // UI 즉시 업데이트 (낙관적 업데이트)
    setFields(newOrder);

    try {
      // 변경된 필드들만 API 호출
      for (let i = 0; i < newOrder.length; i++) {
        if (newOrder[i].display_order !== i) {
          await axios.put(`/api/fields/${newOrder[i].field_id}`, {
            display_order: i
          });
        }
      }
    } catch (error) {
      // 실패 시 원래 순서로 복구
      setFields(fields);
      alert('필드 순서 변경에 실패했습니다');
    }
  }

  return (
    <div>
      <h2>필드 순서 설정</h2>
      <DraggableList items={fields} onDrop={onFieldDrop}>
        {fields.map(field => (
          <FieldItem key={field.field_id} field={field} />
        ))}
      </DraggableList>
    </div>
  );
}
```

---

## 필드 옵션 순서 변경

### 개념

**필드 옵션 순서는 드롭다운 표시 순서입니다.**

```
"상태" 필드의 옵션 순서:
0. 할 일     (display_order: 0)
1. 진행중    (display_order: 1)
2. 완료      (display_order: 2)

→ 보드 생성/수정 시 드롭다운이 이 순서로 표시됨
```

**저장 위치**: `field_options` 테이블의 `display_order` 컬럼

---

### API: 옵션 수정

#### 엔드포인트

```
PUT /api/field-options/{option_id}
```

#### 언제 호출?

- ✅ 사용자가 옵션 순서를 변경했을 때
- ✅ 드래그앤드롭으로 옵션 순서를 바꿨을 때

#### 요청 바디

```json
{
  "display_order": 1  // 새로운 순서
}
```

#### 응답

```json
{
  "data": {
    "option_id": "option-uuid",
    "field_id": "field-uuid",
    "label": "진행중",
    "display_order": 1,
    // ... 기타 옵션 정보
  }
}
```

---

### 시나리오: 옵션 순서 변경

**상황**: "상태" 필드의 옵션 순서를 변경

```
현재 순서:
0. 할 일
1. 진행중
2. 완료

원하는 순서:
0. 진행중    ← 위로 올림
1. 할 일     ← 아래로 내림
2. 완료
```

**API 호출**:

```typescript
// 각 옵션의 display_order 업데이트
await axios.put(`/api/field-options/${option진행중.id}`, { display_order: 0 });
await axios.put(`/api/field-options/${option할일.id}`, { display_order: 1 });
// 완료는 변경 없음
```

**프론트엔드 코드**:

```typescript
function FieldOptionsEditor({ field }: { field: Field }) {
  const [options, setOptions] = useState<Option[]>(field.options);

  async function onOptionDrop(draggedOptionId: string, targetIndex: number) {
    // 옵션 배열 재정렬
    const newOrder = [...options];
    const draggedOption = newOrder.find(o => o.option_id === draggedOptionId);
    const oldIndex = newOrder.indexOf(draggedOption);

    newOrder.splice(oldIndex, 1);
    newOrder.splice(targetIndex, 0, draggedOption);

    // UI 즉시 업데이트
    setOptions(newOrder);

    try {
      // 변경된 옵션들만 API 호출
      for (let i = 0; i < newOrder.length; i++) {
        if (newOrder[i].display_order !== i) {
          await axios.put(`/api/field-options/${newOrder[i].option_id}`, {
            display_order: i
          });
        }
      }
    } catch (error) {
      // 실패 시 복구
      setOptions(options);
      alert('옵션 순서 변경에 실패했습니다');
    }
  }

  return (
    <div>
      <h3>{field.name} 옵션 순서</h3>
      <DraggableList items={options} onDrop={onOptionDrop}>
        {options.map(option => (
          <OptionItem
            key={option.option_id}
            option={option}
          />
        ))}
      </DraggableList>
    </div>
  );
}
```

---

## 전체 시나리오별 API 정리

| 무엇을 변경? | API | 언제 호출? | 저장 위치 |
|-------------|-----|-----------|----------|
| **보드 순서 (리스트)** | `POST /api/v1/boards/{id}/move` | 리스트에서 보드 드래그앤드롭 | `user_board_order.position` |
| **보드 순서 (칸반 내)** | `POST /api/v1/boards/{id}/move` | 같은 컬럼 내에서 순서 변경 | `user_board_order.position` |
| **보드 컬럼 이동** | `POST /api/v1/boards/{id}/move` | 다른 컬럼으로 이동 | `board_field_values.value` + `user_board_order.position` |
| **커스텀 필드 순서** | `PUT /api/fields/{id}` | 필드 설정에서 순서 변경 | `project_fields.display_order` |
| **필드 옵션 순서** | `PUT /api/field-options/{id}` | 옵션 설정에서 순서 변경 | `field_options.display_order` |

---

## 실전 코드 예시

### 완전한 칸반 보드 구현

```typescript
import React, { useState } from 'react';
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd';
import axios from 'axios';

interface Board {
  id: string;
  title: string;
  position: string;
  custom_fields: Record<string, string>;
}

interface Column {
  optionId: string;
  label: string;
  color: string;
  boards: Board[];
}

function KanbanBoard({
  columns,
  viewId,
  groupByFieldId
}: {
  columns: Column[];
  viewId: string;
  groupByFieldId: string;
}) {
  const [kanbanColumns, setKanbanColumns] = useState(columns);

  async function onDragEnd(result: any) {
    const { source, destination, draggableId } = result;

    // 드롭 위치가 없으면 무시
    if (!destination) return;

    // 같은 위치면 무시
    if (
      source.droppableId === destination.droppableId &&
      source.index === destination.index
    ) {
      return;
    }

    // 보드 ID
    const boardId = draggableId;

    // 출발/도착 컬럼
    const sourceColumn = kanbanColumns.find(c => c.optionId === source.droppableId);
    const destColumn = kanbanColumns.find(c => c.optionId === destination.droppableId);

    // 같은 컬럼 내 이동
    if (source.droppableId === destination.droppableId) {
      await handleSameColumnMove(boardId, destColumn, destination.index);
    } else {
      // 다른 컬럼으로 이동
      await handleCrossColumnMove(
        boardId,
        sourceColumn,
        destColumn,
        destination.index
      );
    }
  }

  // 같은 컬럼 내 순서 변경
  async function handleSameColumnMove(
    boardId: string,
    column: Column,
    newIndex: number
  ) {
    const boards = column.boards;

    // before/after position 계산
    const beforeBoard = newIndex > 0 ? boards[newIndex - 1] : null;
    const afterBoard = newIndex < boards.length ? boards[newIndex] : null;

    try {
      const response = await axios.post(`/api/v1/boards/${boardId}/move`, {
        view_id: viewId,
        group_by_field_id: groupByFieldId,
        new_field_value: column.optionId,  // 같은 컬럼
        before_position: beforeBoard?.position || null,
        after_position: afterBoard?.position || null
      });

      // UI 업데이트
      updateBoardPosition(boardId, response.data.new_position);

    } catch (error) {
      console.error('Failed to move board:', error);
      alert('보드 이동에 실패했습니다');
      // 필요 시 원래 위치로 복구
      revertUI();
    }
  }

  // 다른 컬럼으로 이동
  async function handleCrossColumnMove(
    boardId: string,
    sourceColumn: Column,
    destColumn: Column,
    newIndex: number
  ) {
    const destBoards = destColumn.boards;

    // before/after position 계산
    const beforeBoard = newIndex > 0 ? destBoards[newIndex - 1] : null;
    const afterBoard = newIndex < destBoards.length ? destBoards[newIndex] : null;

    try {
      const response = await axios.post(`/api/v1/boards/${boardId}/move`, {
        view_id: viewId,
        group_by_field_id: groupByFieldId,
        new_field_value: destColumn.optionId,  // 새 컬럼
        before_position: beforeBoard?.position || null,
        after_position: afterBoard?.position || null
      });

      // UI 업데이트 (컬럼 + position + 필드 값)
      updateBoard(boardId, {
        position: response.data.new_position,
        columnId: destColumn.optionId,
        custom_fields: {
          [groupByFieldId]: destColumn.optionId
        }
      });

    } catch (error) {
      console.error('Failed to move board:', error);
      alert('보드 이동에 실패했습니다');
      revertUI();
    }
  }

  function updateBoardPosition(boardId: string, newPosition: string) {
    setKanbanColumns(prev =>
      prev.map(column => ({
        ...column,
        boards: column.boards.map(board =>
          board.id === boardId
            ? { ...board, position: newPosition }
            : board
        )
      }))
    );
  }

  function updateBoard(
    boardId: string,
    updates: { position: string; columnId: string; custom_fields: any }
  ) {
    setKanbanColumns(prev => {
      // 보드를 이전 컬럼에서 제거
      const withoutBoard = prev.map(column => ({
        ...column,
        boards: column.boards.filter(b => b.id !== boardId)
      }));

      // 보드를 새 컬럼에 추가
      return withoutBoard.map(column => {
        if (column.optionId === updates.columnId) {
          const movedBoard = prev
            .flatMap(c => c.boards)
            .find(b => b.id === boardId);

          return {
            ...column,
            boards: [
              ...column.boards,
              {
                ...movedBoard,
                position: updates.position,
                custom_fields: updates.custom_fields
              }
            ].sort((a, b) => a.position.localeCompare(b.position))
          };
        }
        return column;
      });
    });
  }

  function revertUI() {
    // 원래 상태로 복구 (필요 시 구현)
  }

  return (
    <DragDropContext onDragEnd={onDragEnd}>
      <div style={{ display: 'flex', gap: '16px' }}>
        {kanbanColumns.map(column => (
          <Droppable key={column.optionId} droppableId={column.optionId}>
            {(provided) => (
              <div
                ref={provided.innerRef}
                {...provided.droppableProps}
                style={{
                  minWidth: '300px',
                  backgroundColor: '#f5f5f5',
                  padding: '16px',
                  borderRadius: '8px'
                }}
              >
                <h3 style={{ color: column.color }}>
                  {column.label} ({column.boards.length})
                </h3>

                {column.boards
                  .sort((a, b) => a.position.localeCompare(b.position))
                  .map((board, index) => (
                    <Draggable
                      key={board.id}
                      draggableId={board.id}
                      index={index}
                    >
                      {(provided) => (
                        <div
                          ref={provided.innerRef}
                          {...provided.draggableProps}
                          {...provided.dragHandleProps}
                          style={{
                            ...provided.draggableProps.style,
                            backgroundColor: 'white',
                            padding: '12px',
                            marginTop: '8px',
                            borderRadius: '4px',
                            cursor: 'grab'
                          }}
                        >
                          {board.title}
                        </div>
                      )}
                    </Draggable>
                  ))}

                {provided.placeholder}
              </div>
            )}
          </Droppable>
        ))}
      </div>
    </DragDropContext>
  );
}

export default KanbanBoard;
```

---

## 주의사항

### 1. 낙관적 업데이트 (Optimistic Update)

```typescript
// ✅ 권장: UI 먼저 업데이트 → API 호출 → 실패 시 복구
function onDragEnd(result) {
  // 1. UI 즉시 업데이트
  updateUIOptimistically(result);

  // 2. API 호출
  moveBoard(...)
    .then(response => {
      // 3. 서버 응답으로 확정
      confirmUIUpdate(response.new_position);
    })
    .catch(error => {
      // 4. 실패 시 복구
      revertUIUpdate();
    });
}
```

### 2. Position 문자열 비교

```typescript
// ✅ 올바른 정렬
boards.sort((a, b) => a.position.localeCompare(b.position));

// ❌ 잘못된 정렬
boards.sort((a, b) => a.position > b.position ? 1 : -1);  // 버그!
```

### 3. 뷰별 순서 주의

```typescript
// ❌ 잘못된 구현: 뷰 상관없이 순서 저장
await moveBoard(boardId, newPosition);  // view_id가 없음!

// ✅ 올바른 구현: 현재 뷰의 순서 저장
await moveBoard(boardId, {
  view_id: currentViewId,  // 필수!
  // ...
});
```

### 4. 필드 값 변경 주의

칸반에서 컬럼을 이동하면 **필드 값도 함께 변경**됩니다:

```typescript
// 보드를 "할 일" → "진행중"으로 이동
await moveBoard(boardId, {
  new_field_value: "in-progress-option-id"  // 상태가 변경됨!
});

// UI에서도 반영 필요
board.custom_fields[statusFieldId] = "in-progress-option-id";
```

---

## 요약

### 보드 순서 변경
- **API**: `POST /api/v1/boards/{id}/move`
- **파라미터**: `view_id`, `before_position`, `after_position`
- **주의**: 뷰별, 사용자별로 순서가 다름

### 필드 순서 변경
- **API**: `PUT /api/fields/{id}`
- **파라미터**: `display_order`
- **주의**: 프로젝트 전체에 영향

### 옵션 순서 변경
- **API**: `PUT /api/field-options/{id}`
- **파라미터**: `display_order`
- **주의**: 드롭다운 표시 순서만 변경

---

## 관련 문서

- [Fractional Indexing 상세 가이드](./FRONTEND_API_GUIDE.md)
- [View API 가이드](./VIEW_API_GUIDE.md)
- [API 통합 가이드](./API_INTEGRATION_GUIDE.md)
