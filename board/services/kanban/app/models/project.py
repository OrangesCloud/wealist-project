from sqlalchemy import Column, String, Text, Integer, Enum as SQLEnum, Date, Boolean
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel
from app.models.enums import ProjectStatus, Priority

class Project(BaseModel):
    __tablename__ = "projects"

    name = Column(String(200), nullable=False, index=True)
    description = Column(Text, nullable=True)
    status = Column(
        SQLEnum(ProjectStatus),
        default=ProjectStatus.PLANNING,
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
    workspace_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References workspaces.id (no FK for sharding)"
    )

    # ERD 추가 필드
    start_date = Column(Date, nullable=True, comment="프로젝트 시작일")
    end_date = Column(Date, nullable=True, comment="프로젝트 종료일")
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<Project(id={self.id}, name={self.name})>"
