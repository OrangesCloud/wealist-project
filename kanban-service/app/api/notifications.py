from fastapi import APIRouter, Depends, HTTPException, status, Query, Path
from sqlalchemy.orm import Session
from sqlalchemy import func
from typing import Optional
from uuid import UUID
from datetime import datetime
from app.database import get_db
from app.auth import get_current_user_id
from app.models.notification import Notification
from app.models.enums import NotificationType
from app.schemas.notification import (
    NotificationCreate,
    NotificationResponse,
    NotificationListResponse,
    NotificationMarkReadResponse,
    NotificationUnreadCountResponse
)

router = APIRouter()

@router.post("/", response_model=NotificationResponse, status_code=status.HTTP_201_CREATED)
async def create_notification(
    notification_in: NotificationCreate,
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    새로운 알림 생성

    일반적으로 시스템에서 자동으로 생성되지만,
    수동으로 생성할 수도 있습니다.
    """
    db_notification = Notification(
        **notification_in.model_dump(),
        created_by=current_user_id
    )
    db.add(db_notification)
    db.commit()
    db.refresh(db_notification)
    return db_notification

@router.get("/", response_model=NotificationListResponse)
async def list_notifications(
    is_read: Optional[bool] = Query(None, description="읽음 여부 필터 (None=전체, True=읽음, False=안읽음)"),
    notification_type: Optional[NotificationType] = Query(None, description="알림 타입 필터"),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    현재 사용자의 알림 목록 조회

    - 최신순으로 정렬됩니다
    - is_read 파라미터로 읽음/안읽음 필터링 가능
    - notification_type으로 특정 타입만 필터링 가능
    """
    # 현재 사용자의 알림만 조회
    query = db.query(Notification).filter(Notification.user_id == current_user_id)

    if is_read is not None:
        query = query.filter(Notification.is_read == is_read)

    if notification_type:
        query = query.filter(Notification.notification_type == notification_type)

    total = query.count()

    # 읽지 않은 알림 개수 계산
    unread_count = db.query(func.count(Notification.id)).filter(
        Notification.user_id == current_user_id,
        Notification.is_read == False
    ).scalar()

    notifications = query.order_by(Notification.created_at.desc()).limit(limit).offset(offset).all()

    return {
        "total": total,
        "unread_count": unread_count,
        "items": notifications,
        "limit": limit,
        "offset": offset
    }

@router.get("/unread-count", response_model=NotificationUnreadCountResponse)
async def get_unread_count(
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    읽지 않은 알림 개수 조회

    헤더의 알림 뱃지 표시 등에 사용됩니다.
    """
    unread_count = db.query(func.count(Notification.id)).filter(
        Notification.user_id == current_user_id,
        Notification.is_read == False
    ).scalar()

    return {"unread_count": unread_count}

@router.get("/{notification_id}", response_model=NotificationResponse)
async def get_notification(
    notification_id: UUID = Path(..., description="알림 ID"),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """특정 알림 조회"""
    notification = db.query(Notification).filter(
        Notification.id == notification_id,
        Notification.user_id == current_user_id
    ).first()
    if not notification:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Notification {notification_id} not found"
        )
    return notification

@router.patch("/{notification_id}/read", response_model=NotificationMarkReadResponse)
async def mark_notification_as_read(
    notification_id: UUID = Path(..., description="알림 ID"),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    알림을 읽음으로 표시

    is_read를 True로 설정하고 read_at에 현재 시각을 저장합니다.
    """
    notification = db.query(Notification).filter(
        Notification.id == notification_id,
        Notification.user_id == current_user_id
    ).first()
    if not notification:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Notification {notification_id} not found"
        )

    if notification.is_read:
        # 이미 읽음 상태
        return {
            "id": notification.id,
            "is_read": notification.is_read,
            "read_at": notification.read_at
        }

    notification.is_read = True
    notification.read_at = datetime.utcnow()
    notification.updated_by = current_user_id
    db.commit()
    db.refresh(notification)

    return {
        "id": notification.id,
        "is_read": notification.is_read,
        "read_at": notification.read_at
    }

@router.post("/mark-all-read", status_code=status.HTTP_200_OK)
async def mark_all_notifications_as_read(
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    모든 알림을 읽음으로 표시

    현재 사용자의 읽지 않은 알림을 모두 읽음 처리합니다.
    """
    updated_count = db.query(Notification).filter(
        Notification.user_id == current_user_id,
        Notification.is_read == False
    ).update({
        "is_read": True,
        "read_at": datetime.utcnow(),
        "updated_by": current_user_id
    }, synchronize_session=False)

    db.commit()

    return {
        "message": f"{updated_count} notifications marked as read",
        "count": updated_count
    }

@router.delete("/{notification_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_notification(
    notification_id: UUID = Path(..., description="알림 ID"),
    db: Session = Depends(get_db),
    current_user_id: UUID = Depends(get_current_user_id)
):
    """
    알림 삭제 (영구 삭제)

    알림은 소프트 삭제 대신 영구 삭제됩니다.
    사용자가 알림을 삭제하면 복구할 수 없습니다.
    """
    notification = db.query(Notification).filter(
        Notification.id == notification_id,
        Notification.user_id == current_user_id
    ).first()
    if not notification:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Notification {notification_id} not found"
        )

    db.delete(notification)
    db.commit()
    return None
