from sqlalchemy import Column, String, Text, Boolean, Integer
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel

class TicketType(BaseModel):
    """프로젝트별 커스텀 티켓 타입/카테고리"""
    __tablename__ = "ticket_types"

    project_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="프로젝트 ID (projects.id 참조)"
    )
    type_name = Column(
        String(100),
        nullable=False,
        index=True,
        comment="티켓 타입 이름 (예: '개발', '디자인', '버그', '긴급 이슈')"
    )
    description = Column(
        Text,
        nullable=True,
        comment="타입 설명"
    )
    color = Column(
        String(7),
        nullable=True,
        comment="UI 표시용 색상 (HEX, 예: #FF5733)"
    )
    icon = Column(
        String(50),
        nullable=True,
        comment="아이콘 이름 (예: 'bug', 'feature', 'design', 'urgent')"
    )
    display_order = Column(
        Integer,
        nullable=True,
        comment="표시 순서 (정렬용)"
    )
    is_deleted = Column(
        Boolean,
        default=False,
        nullable=False,
        index=True,
        comment="소프트 삭제 플래그"
    )

    def __repr__(self):
        return f"<TicketType(id={self.id}, project_id={self.project_id}, type_name={self.type_name})>"
