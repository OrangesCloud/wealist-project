from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from typing import Optional
from uuid import UUID
from app.database import get_db
from app.auth import get_current_user_id
from app.models.project import Project
from app.models.workspace import Workspace
from app.models.enums import ProjectStatus, Priority
from app.schemas.project import (
    ProjectCreate,
    ProjectUpdate,
    ProjectResponse,
    ProjectListResponse
)

router = APIRouter()

@router.post("/", response_model=ProjectResponse, status_code=status.HTTP_201_CREATED)
async def create_project(
    project_in: ProjectCreate,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """새로운 Project 생성"""
    # Workspace 존재 확인
    workspace = db.query(Workspace).filter(Workspace.id == project_in.workspace_id).first()
    if not workspace:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Workspace {project_in.workspace_id} not found"
        )

    db_project = Project(
        **project_in.model_dump(),
        created_by=current_user_id
    )
    db.add(db_project)
    db.commit()
    db.refresh(db_project)
    return db_project

@router.get("/", response_model=ProjectListResponse)
async def list_projects(
    workspace_id: Optional[UUID] = Query(None),
    status_filter: Optional[ProjectStatus] = Query(None, alias="status"),
    priority: Optional[Priority] = Query(None),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Project 목록 조회 (필터링 및 페이지네이션)"""
    query = db.query(Project)

    if workspace_id:
        query = query.filter(Project.workspace_id == workspace_id)
    if status_filter:
        query = query.filter(Project.status == status_filter)
    if priority:
        query = query.filter(Project.priority == priority)

    total = query.count()
    projects = query.order_by(Project.created_at.desc()).limit(limit).offset(offset).all()

    return {
        "total": total,
        "items": projects,
        "limit": limit,
        "offset": offset
    }

@router.get("/{project_id}", response_model=ProjectResponse)
async def get_project(
    project_id: UUID,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """특정 Project 조회"""
    project = db.query(Project).filter(Project.id == project_id).first()
    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project {project_id} not found"
        )
    return project

@router.patch("/{project_id}", response_model=ProjectResponse)
async def update_project(
    project_id: UUID,
    project_in: ProjectUpdate,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Project 정보 수정"""
    project = db.query(Project).filter(Project.id == project_id).first()
    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project {project_id} not found"
        )

    update_data = project_in.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(project, field, value)

    project.updated_by = current_user_id

    db.commit()
    db.refresh(project)
    return project

@router.delete("/{project_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_project(
    project_id: UUID,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Project 삭제 (애플리케이션 레벨에서 CASCADE 처리)"""
    from app.models.ticket import Ticket
    from app.models.task import Task

    project = db.query(Project).filter(Project.id == project_id).first()
    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project {project_id} not found"
        )

    # 애플리케이션 레벨에서 CASCADE 삭제 (샤딩 대비)
    # 1. project에 속한 모든 ticket 찾기
    tickets = db.query(Ticket).filter(Ticket.project_id == project_id).all()
    ticket_ids = [t.id for t in tickets]

    if ticket_ids:
        # 2. ticket들에 속한 모든 task 삭제
        db.query(Task).filter(Task.ticket_id.in_(ticket_ids)).delete(synchronize_session=False)

    # 3. project에 속한 모든 ticket 삭제
    db.query(Ticket).filter(Ticket.project_id == project_id).delete(synchronize_session=False)

    # 4. project 삭제
    db.delete(project)
    db.commit()
    return None
