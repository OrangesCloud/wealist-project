#!/bin/bash
# cleanup_test_data.sh

TOKEN="your_token_here"
WORKSPACE_ID="your_workspace_id_here"

echo "테스트 데이터 정리 중..."

# Delete workspace (cascade delete projects and kanbans)
curl -X DELETE "http://localhost:8000/api/workspaces/$WORKSPACE_ID" \
     -H "Authorization: Bearer $TOKEN"

echo "정리 완료"
