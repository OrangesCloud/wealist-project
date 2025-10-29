from sqlalchemy import Column, Text, Boolean, Enum as SQLEnum
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel
from app.models.enums import TargetType

class Comment(BaseModel):
    """댓글 - 티켓/태스크에 대한 댓글 및 대댓글"""
    __tablename__ = "comments"

    target_type = Column(
        SQLEnum(TargetType),
        nullable=False,
        index=True,
        comment="댓글 대상 타입 (TICKET, TASK)"
    )
    target_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="댓글 대상 ID (tickets.id 또는 tasks.id)"
    )
    author_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References users.id from User service - 댓글 작성자"
    )
    content = Column(Text, nullable=False, comment="댓글 내용")
    parent_id = Column(
        PG_UUID,
        nullable=True,
        index=True,
        comment="대댓글인 경우 부모 댓글 ID (self-reference)"
    )
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<Comment(id={self.id}, target_type={self.target_type}, target_id={self.target_id}, author_id={self.author_id})>"
