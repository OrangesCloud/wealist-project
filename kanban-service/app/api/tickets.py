from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from typing import Optional
from uuid import UUID
from app.database import get_db
from app.auth import get_current_user_id
from app.models.ticket import Ticket
from app.models.project import Project
from app.models.enums import TicketStatus, Priority
from app.schemas.ticket import (
    TicketCreate,
    TicketUpdate,
    TicketResponse,
    TicketListResponse
)

router = APIRouter()

@router.post("/", response_model=TicketResponse, status_code=status.HTTP_201_CREATED)
async def create_ticket(
    ticket_in: TicketCreate,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """새로운 Ticket 생성"""
    # Project 존재 확인
    project = db.query(Project).filter(Project.id == ticket_in.project_id).first()
    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project {ticket_in.project_id} not found"
        )

    db_ticket = Ticket(
        **ticket_in.model_dump(),
        created_by=current_user_id
    )
    db.add(db_ticket)
    db.commit()
    db.refresh(db_ticket)
    return db_ticket

@router.get("/", response_model=TicketListResponse)
async def list_tickets(
    project_id: Optional[UUID] = Query(None),
    status_filter: Optional[TicketStatus] = Query(None, alias="status"),
    priority: Optional[Priority] = Query(None),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Ticket 목록 조회 (필터링 및 페이지네이션)"""
    query = db.query(Ticket)

    if project_id:
        query = query.filter(Ticket.project_id == project_id)
    if status_filter:
        query = query.filter(Ticket.status == status_filter)
    if priority:
        query = query.filter(Ticket.priority == priority)

    total = query.count()
    tickets = query.order_by(Ticket.created_at.desc()).limit(limit).offset(offset).all()

    return {
        "total": total,
        "items": tickets,
        "limit": limit,
        "offset": offset
    }

@router.get("/{ticket_id}", response_model=TicketResponse)
async def get_ticket(
    ticket_id: UUID,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """특정 Ticket 조회"""
    ticket = db.query(Ticket).filter(Ticket.id == ticket_id).first()
    if not ticket:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Ticket {ticket_id} not found"
        )
    return ticket

@router.patch("/{ticket_id}", response_model=TicketResponse)
async def update_ticket(
    ticket_id: UUID,
    ticket_in: TicketUpdate,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Ticket 정보 수정"""
    ticket = db.query(Ticket).filter(Ticket.id == ticket_id).first()
    if not ticket:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Ticket {ticket_id} not found"
        )

    update_data = ticket_in.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(ticket, field, value)

    ticket.updated_by = current_user_id

    db.commit()
    db.refresh(ticket)
    return ticket

@router.delete("/{ticket_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_ticket(
    ticket_id: UUID,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """Ticket 삭제 (애플리케이션 레벨에서 CASCADE 처리)"""
    from app.models.task import Task

    ticket = db.query(Ticket).filter(Ticket.id == ticket_id).first()
    if not ticket:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Ticket {ticket_id} not found"
        )

    # 애플리케이션 레벨에서 CASCADE 삭제 (샤딩 대비)
    # 1. ticket에 속한 모든 task 삭제
    db.query(Task).filter(Task.ticket_id == ticket_id).delete(synchronize_session=False)

    # 2. ticket 삭제
    db.delete(ticket)
    db.commit()
    return None
