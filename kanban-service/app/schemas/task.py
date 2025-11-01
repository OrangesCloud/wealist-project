from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional
from uuid import UUID
from app.models.enums import TaskStatus

class TaskBase(BaseModel):
    title: str = Field(..., min_length=1, max_length=300)
    description: Optional[str] = Field(None, max_length=2000)
    status: TaskStatus = TaskStatus.TODO

class TaskCreate(TaskBase):
    ticket_id: UUID = Field(...)
    assignee_id: Optional[UUID] = Field(None, description="담당자 ID")

class TaskUpdate(BaseModel):
    title: Optional[str] = Field(None, min_length=1, max_length=300)
    description: Optional[str] = None
    status: Optional[TaskStatus] = None
    assignee_id: Optional[UUID] = Field(None, description="담당자 ID")

class TaskResponse(TaskBase):
    id: UUID
    ticket_id: UUID
    completed_at: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime
    created_by: UUID
    updated_by: Optional[UUID] = None
    assignee_id: Optional[UUID] = None

    class Config:
        from_attributes = True

class TaskListResponse(BaseModel):
    total: int
    items: list[TaskResponse]
    limit: int
    offset: int

    class Config:
        from_attributes = True
