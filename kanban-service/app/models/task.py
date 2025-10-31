from sqlalchemy import Column, String, Text, Integer, Enum as SQLEnum, DateTime, Date, Boolean
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel
from app.models.enums import TaskStatus

class Task(BaseModel):
    __tablename__ = "tasks"

    title = Column(String(300), nullable=False, index=True)
    description = Column(Text, nullable=True)
    status = Column(
        SQLEnum(TaskStatus),
        default=TaskStatus.TODO,
        nullable=False,
        index=True
    )
    completed_at = Column(DateTime(timezone=True), nullable=True)

    # FK 제거: 샤딩 및 DB 분리 대비
    ticket_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References tickets.id (no FK for sharding)"
    )

    # 담당자 (Member 서비스 users 테이블 참조)
    assignee_id = Column(
        PG_UUID,
        nullable=True,
        index=True,
        comment="References users.id from Member service - 작업 담당자"
    )

    # ERD 추가 필드
    due_date = Column(Date, nullable=True, comment="태스크 마감일")
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<Task(id={self.id}, title={self.title})>"

    def mark_completed(self):
        from datetime import datetime, timezone
        self.status = TaskStatus.DONE
        self.completed_at = datetime.now(timezone.utc)
