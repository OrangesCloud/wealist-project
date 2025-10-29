from sqlalchemy import Column, String, Boolean, Enum as SQLEnum, BigInteger
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from app.models.base import BaseModel
from app.models.enums import TargetType

class Attachment(BaseModel):
    """첨부파일 - 프로젝트/티켓/태스크에 대한 파일 첨부"""
    __tablename__ = "attachments"

    target_type = Column(
        SQLEnum(TargetType),
        nullable=False,
        index=True,
        comment="첨부 대상 타입 (PROJECT, TICKET, TASK)"
    )
    target_id = Column(
        PG_UUID,
        nullable=False,
        index=True,
        comment="첨부 대상 ID (projects.id, tickets.id, 또는 tasks.id)"
    )
    file_name = Column(String(255), nullable=False, comment="파일명")
    file_path = Column(String(500), nullable=False, comment="파일 저장 경로")
    file_size = Column(BigInteger, nullable=True, comment="파일 크기 (bytes)")
    mime_type = Column(String(100), nullable=True, comment="MIME 타입 (예: image/png, application/pdf)")
    is_deleted = Column(Boolean, default=False, nullable=False, index=True, comment="소프트 삭제 플래그")

    def __repr__(self):
        return f"<Attachment(id={self.id}, target_type={self.target_type}, target_id={self.target_id}, file_name={self.file_name})>"
