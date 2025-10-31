from sqlalchemy import Column, String, Text, Boolean
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel

class ProjectRole(BaseModel):
    """프로젝트별 커스텀 역할 정의"""
    __tablename__ = "project_roles"

    project_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="References projects.id (no FK for sharding)"
    )
    role_name = Column(
        String(100),
        nullable=False,
        index=True,
        comment="역할 이름 (예: 개발1팀, 디자이너, 기획자)"
    )
    description = Column(Text, nullable=True, comment="역할 설명")
    color = Column(String(7), nullable=True, comment="UI 표시 색상 (HEX, 예: #FF5733)")
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<ProjectRole(id={self.id}, project_id={self.project_id}, role_name={self.role_name})>"
