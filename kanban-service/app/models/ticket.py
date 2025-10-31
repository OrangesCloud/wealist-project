from sqlalchemy import Column, String, Text, Integer, Enum as SQLEnum, Date, Boolean
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel
from app.models.enums import TicketStatus, Priority

class Ticket(BaseModel):
    __tablename__ = "tickets"

    title = Column(String(300), nullable=False, index=True)
    description = Column(Text, nullable=True)
    status = Column(
        SQLEnum(TicketStatus),
        default=TicketStatus.OPEN,
        nullable=False,
        index=True
    )
    priority = Column(
        SQLEnum(Priority),
        default=Priority.MEDIUM,
        nullable=False,
        index=True
    )

    # FK 제거: 샤딩 및 DB 분리 대비
    project_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References projects.id (no FK for sharding)"
    )

    # 담당자 (Member 서비스 users 테이블 참조)
    assignee_id = Column(
        PG_UUID,
        nullable=True,
        index=True,
        comment="References users.id from Member service - 티켓 담당자"
    )

    # 티켓 타입 (프로젝트별 커스텀 카테고리)
    ticket_type_id = Column(
        PG_UUID,
        nullable=True,
        index=True,
        comment="티켓 타입 ID (ticket_types.id 참조) - 프로젝트별 커스텀 카테고리 (예: '개발', '디자인', '버그')"
    )

    # ERD 추가 필드
    parent_ticket = Column(
        PG_UUID,
        nullable=True,
        index=True,
        comment="상위 티켓 ID (서브태스크용, self-reference)"
    )
    due_date = Column(Date, nullable=True, comment="티켓 마감일")
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<Ticket(id={self.id}, title={self.title})>"
