import pytest
from fastapi import status
from uuid import UUID

def test_create_ticket_success(client, sample_project):
    """Ticket 생성 성공 테스트"""
    response = client.post(
        "/api/tickets/",
        json={
            "title": "로그인 API",
            "project_id": str(sample_project.id),
            "status": "IN_PROGRESS",
            "priority": "HIGH"
        }
    )
    assert response.status_code == status.HTTP_201_CREATED
    data = response.json()
    assert data["title"] == "로그인 API"
    assert data["project_id"] == str(sample_project.id)
    assert data["status"] == "IN_PROGRESS"
    assert data["priority"] == "HIGH"

def test_create_ticket_project_not_found(client):
    """존재하지 않는 Project로 Ticket 생성 시 실패"""
    response = client.post(
        "/api/tickets/",
        json={
            "title": "Test Ticket",
            "project_id": str(UUID('00000000-0000-0000-0000-000000000001'))
        }
    )
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_list_tickets(client, sample_ticket):
    """Ticket 목록 조회"""
    response = client.get("/api/tickets/")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 1
    assert len(data["items"]) == 1

def test_list_tickets_with_filters(client, sample_ticket):
    """필터링으로 Ticket 목록 조회"""
    response = client.get(
        f"/api/tickets/?project_id={sample_ticket.project_id}&status=OPEN"
    )
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 1

def test_get_ticket_success(client, sample_ticket):
    """Ticket 조회 성공"""
    response = client.get(f"/api/tickets/{sample_ticket.id}")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["id"] == str(sample_ticket.id)

def test_get_ticket_not_found(client):
    """존재하지 않는 Ticket 조회 시 실패"""
    response = client.get("/api/tickets/" + str(UUID('00000000-0000-0000-0000-000000000001')))
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_update_ticket_success(client, sample_ticket):
    """Ticket 업데이트 성공"""
    response = client.patch(
        f"/api/tickets/{sample_ticket.id}",
        json={"status": "DONE"}
    )
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["status"] == "DONE"

def test_delete_ticket_success(client, sample_ticket):
    """Ticket 삭제 성공 (Cascade로 하위 Task 삭제)"""
    response = client.delete(f"/api/tickets/{sample_ticket.id}")
    assert response.status_code == status.HTTP_204_NO_CONTENT
