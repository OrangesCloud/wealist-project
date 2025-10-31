import pytest
from fastapi import status
from uuid import UUID

def test_create_task_success(client, sample_ticket):
    """Task 생성 성공 테스트"""
    response = client.post(
        "/api/tasks/",
        json={
            "title": "JWT 발급",
            "ticket_id": str(sample_ticket.id),
            "status": "TODO"
        }
    )
    assert response.status_code == status.HTTP_201_CREATED
    data = response.json()
    assert data["title"] == "JWT 발급"
    assert data["ticket_id"] == str(sample_ticket.id)
    assert data["status"] == "TODO"
    assert data["completed_at"] is None

def test_create_task_ticket_not_found(client):
    """존재하지 않는 Ticket으로 Task 생성 시 실패"""
    response = client.post(
        "/api/tasks/",
        json={
            "title": "Test Task",
            "ticket_id": str(UUID('00000000-0000-0000-0000-000000000001'))
        }
    )
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_list_tasks(client, sample_task):
    """Task 목록 조회"""
    response = client.get("/api/tasks/")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 1
    assert len(data["items"]) == 1

def test_list_tasks_with_filters(client, sample_task):
    """필터링으로 Task 목록 조회"""
    response = client.get(
        f"/api/tasks/?ticket_id={sample_task.ticket_id}&status=TODO"
    )
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 1

def test_get_task_success(client, sample_task):
    """Task 조회 성공"""
    response = client.get(f"/api/tasks/{sample_task.id}")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["id"] == str(sample_task.id)

def test_get_task_not_found(client):
    """존재하지 않는 Task 조회 시 실패"""
    response = client.get("/api/tasks/" + str(UUID('00000000-0000-0000-0000-000000000001')))
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_update_task_success(client, sample_task):
    """Task 업데이트 성공"""
    response = client.patch(
        f"/api/tasks/{sample_task.id}",
        json={"status": "IN_PROGRESS"}
    )
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["status"] == "IN_PROGRESS"

def test_complete_task_success(client, sample_task):
    """Task 완료 처리 성공"""
    response = client.patch(f"/api/tasks/{sample_task.id}/complete")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["status"] == "DONE"
    assert data["completed_at"] is not None

def test_complete_task_already_completed(client, sample_task):
    """이미 완료된 Task 완료 처리 시 실패"""
    # 먼저 완료 처리
    client.patch(f"/api/tasks/{sample_task.id}/complete")

    # 다시 완료 처리 시도
    response = client.patch(f"/api/tasks/{sample_task.id}/complete")
    assert response.status_code == status.HTTP_400_BAD_REQUEST

def test_delete_task_success(client, sample_task):
    """Task 삭제 성공"""
    response = client.delete(f"/api/tasks/{sample_task.id}")
    assert response.status_code == status.HTTP_204_NO_CONTENT
