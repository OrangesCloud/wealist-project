from sqlalchemy import Column, Boolean
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel

class ProjectMember(BaseModel):
    """프로젝트 멤버 - 프로젝트에 참여하는 사용자와 역할"""
    __tablename__ = "project_members"

    project_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References projects.id (no FK for sharding)"
    )
    user_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References users.id from User service"
    )
    role_id = Column(
        PG_UUID,
        nullable=True,
        index=True,
        comment="References project_roles.id (nullable - 역할 미부여 상태 가능)"
    )
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<ProjectMember(id={self.id}, project_id={self.project_id}, user_id={self.user_id}, role_id={self.role_id})>"
