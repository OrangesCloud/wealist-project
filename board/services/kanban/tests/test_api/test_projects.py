import pytest
from fastapi import status
from uuid import UUID

def test_create_project_success(client, sample_workspace):
    """Project 생성 성공 테스트"""
    response = client.post(
        "/api/projects/",
        json={
            "name": "회원 시스템",
            "workspace_id": str(sample_workspace.id),
            "status": "ACTIVE",
            "priority": "HIGH"
        }
    )
    assert response.status_code == status.HTTP_201_CREATED
    data = response.json()
    assert data["name"] == "회원 시스템"
    assert data["workspace_id"] == str(sample_workspace.id)
    assert data["status"] == "ACTIVE"
    assert data["priority"] == "HIGH"

def test_create_project_workspace_not_found(client):
    """존재하지 않는 Workspace로 Project 생성 시 실패"""
    response = client.post(
        "/api/projects/",
        json={
            "name": "Test Project",
            "workspace_id": str(UUID('00000000-0000-0000-0000-000000000001'))
        }
    )
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_list_projects(client, sample_project):
    """Project 목록 조회"""
    response = client.get("/api/projects/")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 1
    assert len(data["items"]) == 1

def test_list_projects_with_filters(client, sample_project):
    """필터링으로 Project 목록 조회"""
    response = client.get(
        f"/api/projects/?workspace_id={sample_project.workspace_id}&status=ACTIVE"
    )
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 1

def test_get_project_success(client, sample_project):
    """Project 조회 성공"""
    response = client.get(f"/api/projects/{sample_project.id}")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["id"] == str(sample_project.id)

def test_get_project_not_found(client):
    """존재하지 않는 Project 조회 시 실패"""
    response = client.get("/api/projects/" + str(UUID('00000000-0000-0000-0000-000000000001')))
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_update_project_success(client, sample_project):
    """Project 업데이트 성공"""
    response = client.patch(
        f"/api/projects/{sample_project.id}",
        json={"status": "COMPLETED"}
    )
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["status"] == "COMPLETED"

def test_delete_project_success(client, sample_project):
    """Project 삭제 성공 (Cascade로 하위 Ticket, Task 삭제)"""
    response = client.delete(f"/api/projects/{sample_project.id}")
    assert response.status_code == status.HTTP_204_NO_CONTENT
