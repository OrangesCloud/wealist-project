from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from uuid import UUID
from app.database import get_db
from app.auth import get_current_user_id
from app.models.workspace import Workspace
from app.schemas.workspace import (
    WorkspaceCreate,
    WorkspaceUpdate,
    WorkspaceResponse,
    WorkspaceListResponse
)

router = APIRouter()

@router.post("/", response_model=WorkspaceResponse, status_code=status.HTTP_201_CREATED)
async def create_workspace(
    workspace_in: WorkspaceCreate,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """새로운 Workspace 생성"""
    existing = db.query(Workspace).filter(Workspace.name == workspace_in.name).first()
    if existing:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"Workspace '{workspace_in.name}' already exists"
        )

    db_workspace = Workspace(
        **workspace_in.model_dump(),
        created_by=current_user_id
    )
    db.add(db_workspace)
    db.commit()
    db.refresh(db_workspace)
    return db_workspace

@router.get("/", response_model=WorkspaceListResponse)
async def list_workspaces(
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Workspace 목록 조회 (페이지네이션)"""
    total = db.query(Workspace).count()
    workspaces = (
        db.query(Workspace)
        .order_by(Workspace.created_at.desc())
        .limit(limit)
        .offset(offset)
        .all()
    )
    return {
        "total": total,
        "items": workspaces,
        "limit": limit,
        "offset": offset
    }

@router.get("/{workspace_id}", response_model=WorkspaceResponse)
async def get_workspace(
    workspace_id: UUID,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """특정 Workspace 조회"""
    workspace = db.query(Workspace).filter(Workspace.id == workspace_id).first()
    if not workspace:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Workspace {workspace_id} not found"
        )
    return workspace

@router.patch("/{workspace_id}", response_model=WorkspaceResponse)
async def update_workspace(
    workspace_id: UUID,
    workspace_in: WorkspaceUpdate,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Workspace 정보 수정"""
    workspace = db.query(Workspace).filter(Workspace.id == workspace_id).first()
    if not workspace:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Workspace {workspace_id} not found"
        )

    update_data = workspace_in.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(workspace, field, value)

    workspace.updated_by = current_user_id

    db.commit()
    db.refresh(workspace)
    return workspace

@router.delete("/{workspace_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_workspace(
    workspace_id: UUID,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Workspace 삭제 (애플리케이션 레벨에서 CASCADE 처리)"""
    from app.models.project import Project
    from app.models.ticket import Ticket
    from app.models.task import Task

    workspace = db.query(Workspace).filter(Workspace.id == workspace_id).first()
    if not workspace:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Workspace {workspace_id} not found"
        )

    # 애플리케이션 레벨에서 CASCADE 삭제 (샤딩 대비)
    # 1. workspace에 속한 모든 project 찾기
    projects = db.query(Project).filter(Project.workspace_id == workspace_id).all()
    project_ids = [p.id for p in projects]

    if project_ids:
        # 2. project들에 속한 모든 ticket 찾기
        tickets = db.query(Ticket).filter(Ticket.project_id.in_(project_ids)).all()
        ticket_ids = [t.id for t in tickets]

        if ticket_ids:
            # 3. ticket들에 속한 모든 task 삭제
            db.query(Task).filter(Task.ticket_id.in_(ticket_ids)).delete(synchronize_session=False)

        # 4. project들에 속한 모든 ticket 삭제
        db.query(Ticket).filter(Ticket.project_id.in_(project_ids)).delete(synchronize_session=False)

    # 5. workspace에 속한 모든 project 삭제
    db.query(Project).filter(Project.workspace_id == workspace_id).delete(synchronize_session=False)

    # 6. workspace 삭제
    db.delete(workspace)
    db.commit()
    return None
