from enum import Enum as PyEnum

class ProjectStatus(str, PyEnum):
    PLANNING = "PLANNING"
    ACTIVE = "ACTIVE"
    COMPLETED = "COMPLETED"
    ON_HOLD = "ON_HOLD"
    CANCELLED = "CANCELLED"

class TicketStatus(str, PyEnum):
    OPEN = "OPEN"
    IN_PROGRESS = "IN_PROGRESS"
    REVIEW = "REVIEW"
    TESTING = "TESTING"
    DONE = "DONE"
    CLOSED = "CLOSED"
    BLOCKED = "BLOCKED"

class TaskStatus(str, PyEnum):
    TODO = "TODO"
    IN_PROGRESS = "IN_PROGRESS"
    REVIEW = "REVIEW"
    DONE = "DONE"

class Priority(str, PyEnum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    URGENT = "URGENT"

class TicketParticipationType(str, PyEnum):
    ASSIGNEE = "ASSIGNEE"
    REVIEWER = "REVIEWER"
    WATCHER = "WATCHER"

class TaskParticipationType(str, PyEnum):
    ASSIGNEE = "ASSIGNEE"
    REVIEWER = "REVIEWER"

class TargetType(str, PyEnum):
    PROJECT = "PROJECT"
    TICKET = "TICKET"
    TASK = "TASK"

class NotificationType(str, PyEnum):
    TICKET_CREATED = "TICKET_CREATED"
    TICKET_UPDATED = "TICKET_UPDATED"
    COMMENT_ADDED = "COMMENT_ADDED"
    TICKET_ASSIGNED = "TICKET_ASSIGNED"
    DUE_DATE_REMINDER = "DUE_DATE_REMINDER"
    MENTION = "MENTION"
