from sqlalchemy import Column, Boolean, Enum as SQLEnum
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel
from app.models.enums import TicketParticipationType

class TicketMember(BaseModel):
    """티켓 멤버 - 티켓에 참여하는 사용자 (담당자, 리뷰어, 관찰자)"""
    __tablename__ = "ticket_members"

    ticket_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References tickets.id (no FK for sharding)"
    )
    user_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References users.id from User service"
    )
    participation_type = Column(
        SQLEnum(TicketParticipationType),
        default=TicketParticipationType.ASSIGNEE,
        nullable=False,
        index=True,
        comment="참여 타입 (담당자/리뷰어/관찰자)"
    )
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<TicketMember(id={self.id}, ticket_id={self.ticket_id}, user_id={self.user_id}, type={self.participation_type})>"
