import pytest
from fastapi import status
from uuid import UUID

def test_create_workspace_success(client):
    """Workspace 생성 성공 테스트"""
    response = client.post(
        "/api/workspaces/",
        json={"name": "개발팀", "description": "백엔드 개발팀"}
    )
    assert response.status_code == status.HTTP_201_CREATED
    data = response.json()
    assert data["name"] == "개발팀"
    assert data["description"] == "백엔드 개발팀"
    assert "id" in data
    assert "created_at" in data
    assert "updated_at" in data

def test_create_workspace_duplicate_name(client, sample_workspace):
    """중복된 이름으로 Workspace 생성 시 실패"""
    response = client.post(
        "/api/workspaces/",
        json={"name": sample_workspace.name}
    )
    assert response.status_code == status.HTTP_409_CONFLICT

def test_list_workspaces_empty(client):
    """빈 Workspace 목록 조회"""
    response = client.get("/api/workspaces/")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 0
    assert len(data["items"]) == 0

def test_list_workspaces_with_data(client, sample_workspace):
    """Workspace 목록 조회"""
    response = client.get("/api/workspaces/")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["total"] == 1
    assert len(data["items"]) == 1
    assert data["items"][0]["id"] == str(sample_workspace.id)

def test_get_workspace_success(client, sample_workspace):
    """Workspace 조회 성공"""
    response = client.get(f"/api/workspaces/{sample_workspace.id}")
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["id"] == str(sample_workspace.id)
    assert data["name"] == sample_workspace.name

def test_get_workspace_not_found(client):
    """존재하지 않는 Workspace 조회 시 실패"""
    response = client.get("/api/workspaces/" + str(UUID('00000000-0000-0000-0000-000000000001')))
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_update_workspace_success(client, sample_workspace):
    """Workspace 업데이트 성공"""
    response = client.patch(
        f"/api/workspaces/{sample_workspace.id}",
        json={"name": "Updated Name"}
    )
    assert response.status_code == status.HTTP_200_OK
    data = response.json()
    assert data["name"] == "Updated Name"

def test_update_workspace_not_found(client):
    """존재하지 않는 Workspace 업데이트 시 실패"""
    response = client.patch(
        "/api/workspaces/" + str(UUID('00000000-0000-0000-0000-000000000001')),
        json={"name": "Updated Name"}
    )
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_delete_workspace_success(client, sample_workspace):
    """Workspace 삭제 성공"""
    response = client.delete(f"/api/workspaces/{sample_workspace.id}")
    assert response.status_code == status.HTTP_204_NO_CONTENT

    # 삭제 확인
    response = client.get(f"/api/workspaces/{sample_workspace.id}")
    assert response.status_code == status.HTTP_404_NOT_FOUND

def test_delete_workspace_not_found(client):
    """존재하지 않는 Workspace 삭제 시 실패"""
    response = client.delete("/api/workspaces/" + str(UUID('00000000-0000-0000-0000-000000000001')))
    assert response.status_code == status.HTTP_404_NOT_FOUND
