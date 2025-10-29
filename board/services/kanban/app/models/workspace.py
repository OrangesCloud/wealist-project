from sqlalchemy import Column, String, Text
from app.models.base import BaseModel

class Workspace(BaseModel):
    __tablename__ = "workspaces"

    name = Column(String(100), nullable=False, index=True)
    description = Column(Text, nullable=True)

    def __repr__(self):
        return f"<Workspace(id={self.id}, name={self.name})>"
