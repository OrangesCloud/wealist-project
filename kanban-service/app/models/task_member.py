from sqlalchemy import Column, Boolean, Enum as SQLEnum
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel
from app.models.enums import TaskParticipationType

class TaskMember(BaseModel):
    """태스크 멤버 - 태스크에 참여하는 사용자 (담당자, 리뷰어)"""
    __tablename__ = "task_members"

    task_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References tasks.id (no FK for sharding)"
    )
    user_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References users.id from User service"
    )
    participation_type = Column(
        SQLEnum(TaskParticipationType),
        default=TaskParticipationType.ASSIGNEE,
        nullable=False,
        index=True,
        comment="참여 타입 (담당자/리뷰어)"
    )
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<TaskMember(id={self.id}, task_id={self.task_id}, user_id={self.user_id}, type={self.participation_type})>"
