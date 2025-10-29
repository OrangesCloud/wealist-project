from fastapi import APIRouter, Depends, HTTPException, status, Query, Path
from sqlalchemy.orm import Session
from typing import Optional
from uuid import UUID
from app.database import get_db
from app.auth import get_current_user_id
from app.models.ticket_type import TicketType
from app.models.project import Project
from app.schemas.ticket_type import (
    TicketTypeCreate,
    TicketTypeUpdate,
    TicketTypeResponse,
    TicketTypeListResponse
)

router = APIRouter()

@router.post("/", response_model=TicketTypeResponse, status_code=status.HTTP_201_CREATED)
async def create_ticket_type(
    project_id: UUID = Path(..., description="프로젝트 ID"),
    ticket_type_in: TicketTypeCreate = ...,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    프로젝트에 새로운 티켓 타입 생성

    - **type_name**: 타입 이름 (예: "개발 작업", "디자인 작업", "버그 수정")
    - **description**: 타입 설명 (optional)
    - **color**: HEX 색상 코드 (optional, 예: #FF5733)
    - **icon**: 아이콘 이름 (optional, 예: "code", "bug", "palette")
    - **display_order**: 표시 순서 (optional)
    """
    # 프로젝트 존재 확인
    project = db.query(Project).filter(
        Project.id == project_id,
        Project.is_deleted == False
    ).first()
    if not project:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Project {project_id} not found"
        )

    # 동일한 프로젝트 내에서 타입 이름 중복 확인
    existing_type = db.query(TicketType).filter(
        TicketType.project_id == project_id,
        TicketType.type_name == ticket_type_in.type_name,
        TicketType.is_deleted == False
    ).first()
    if existing_type:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Ticket type '{ticket_type_in.type_name}' already exists in this project"
        )

    db_ticket_type = TicketType(
        project_id=project_id,
        **ticket_type_in.model_dump(),
        created_by=current_user_id
    )
    db.add(db_ticket_type)
    db.commit()
    db.refresh(db_ticket_type)
    return db_ticket_type

@router.get("/", response_model=TicketTypeListResponse)
async def list_ticket_types(
    project_id: UUID = Path(..., description="프로젝트 ID"),
    include_deleted: bool = Query(False, description="삭제된 항목 포함 여부"),
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    프로젝트의 티켓 타입 목록 조회

    display_order 순으로 정렬되어 반환됩니다.
    """
    query = db.query(TicketType).filter(TicketType.project_id == project_id)

    if not include_deleted:
        query = query.filter(TicketType.is_deleted == False)

    total = query.count()
    ticket_types = query.order_by(
        TicketType.display_order.asc().nulls_last(),
        TicketType.created_at.asc()
    ).limit(limit).offset(offset).all()

    return {
        "total": total,
        "items": ticket_types,
        "limit": limit,
        "offset": offset
    }

@router.get("/{type_id}", response_model=TicketTypeResponse)
async def get_ticket_type(
    project_id: UUID = Path(..., description="프로젝트 ID"),
    type_id: UUID = Path(..., description="티켓 타입 ID"),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """특정 티켓 타입 조회"""
    ticket_type = db.query(TicketType).filter(
        TicketType.id == type_id,
        TicketType.project_id == project_id
    ).first()
    if not ticket_type:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Ticket type {type_id} not found in project {project_id}"
        )
    return ticket_type

@router.patch("/{type_id}", response_model=TicketTypeResponse)
async def update_ticket_type(
    project_id: UUID = Path(..., description="프로젝트 ID"),
    type_id: UUID = Path(..., description="티켓 타입 ID"),
    ticket_type_in: TicketTypeUpdate = ...,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """티켓 타입 정보 수정"""
    ticket_type = db.query(TicketType).filter(
        TicketType.id == type_id,
        TicketType.project_id == project_id,
        TicketType.is_deleted == False
    ).first()
    if not ticket_type:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Ticket type {type_id} not found"
        )

    # 타입 이름 변경 시 중복 확인
    if ticket_type_in.type_name and ticket_type_in.type_name != ticket_type.type_name:
        existing_type = db.query(TicketType).filter(
            TicketType.project_id == project_id,
            TicketType.type_name == ticket_type_in.type_name,
            TicketType.is_deleted == False,
            TicketType.id != type_id
        ).first()
        if existing_type:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Ticket type '{ticket_type_in.type_name}' already exists in this project"
            )

    # 수정 가능한 필드만 업데이트
    update_data = ticket_type_in.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(ticket_type, field, value)

    ticket_type.updated_by = current_user_id
    db.commit()
    db.refresh(ticket_type)
    return ticket_type

@router.delete("/{type_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_ticket_type(
    project_id: UUID = Path(..., description="프로젝트 ID"),
    type_id: UUID = Path(..., description="티켓 타입 ID"),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    티켓 타입 삭제 (소프트 삭제)

    is_deleted 플래그를 True로 설정합니다.
    이미 이 타입을 사용하는 티켓들은 영향받지 않습니다.
    """
    ticket_type = db.query(TicketType).filter(
        TicketType.id == type_id,
        TicketType.project_id == project_id
    ).first()
    if not ticket_type:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Ticket type {type_id} not found"
        )

    if ticket_type.is_deleted:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Ticket type {type_id} is already deleted"
        )

    ticket_type.is_deleted = True
    ticket_type.updated_by = current_user_id
    db.commit()
    return None
