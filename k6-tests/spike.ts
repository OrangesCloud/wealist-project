import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// ----------------------------------------------------
// 1. 테스트 설정 (Test Options)
// ----------------------------------------------------
export const options = {
    // 가상 사용자 (VUs) 및 지속 시간 설정
    vus: 3000,  // 10명의 동시 사용자
    duration: '30s', // 30초 동안 테스트 실행
    thresholds: {
        // http 요청 95% 응답 시간이 500ms 미만
        'http_req_duration': ['p(95)<500'],
        // 모든 체크 항목의 성공률이 99% 이상
        'checks': ['rate>0.99'],
    },
};

// ----------------------------------------------------
// 2. 환경 변수 및 상수 설정 (Base URLs)
// ----------------------------------------------------

// Orange Cloud User Repository API Base URL 
const USER_API_BASE_URL = '실제 주소'; 
// Project Board Management API Base URL 
const PROJECT_API_BASE_URL = '실제 주소'; 

// ----------------------------------------------------
// 3. 메인 테스트 함수 (Test Scenario)
// ----------------------------------------------------

export default function () {
    // VU의 번호 (1, 2, 3...)
    const vuId = __VU; 
    // 고유 식별자 (충돌 방지용)
    const shortId = uuidv4().substring(0, 8); 
    
    // User A와 User B의 이메일 및 Google ID 
    const userAEmail = `userA_${vuId}_${shortId}@test.com`; 
    const userBEmail = `userB_${vuId}_${shortId}@test.com`; 
    const googleIdA = `googleA_${vuId}_${shortId}`; 
    const googleIdB = `googleB_${vuId}_${shortId}`; 
    
    let authTokenA = ''; 
    let authTokenB = ''; 
    let workspaceId = '';
    let projectId = '';
    let boardId = '';
    let userAId = '';
    let userBId = '';
    let commentId = '';
    
    // API 호출 결과를 저장할 변수
    let res;

    // ----------------------------------------------------
    // (1) 사용자 A 생성 및 토큰 획득 (USER API)
    // ----------------------------------------------------

    // 1. 사용자 A 생성
    res = http.post(`${USER_API_BASE_URL}/api/users`, 
        JSON.stringify({ 
            email: userAEmail, 
            googleId: googleIdA,
            provider: "google"
        }), 
        { tags: { name: 'createUserA' }, headers: { 'Content-Type': 'application/json' } }
    );
    if (res.status !== 201 && res.status !== 200) {
        console.error(`VU ${vuId}: [ERROR] Step 1 (User A Created) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '1. User A Created': (r) => r.status === 201 || r.status === 200 });
    
    if (res.status === 201 || res.status === 200) {
        try {
            userAId = res.json().userId; 
            if (!userAId) {
                console.error(`VU ${vuId}: [ERROR] Step 1: Failed to extract 'userId'. Response: ${res.body}`);
            }
        } catch (e) {
            console.error(`VU ${vuId}: [ERROR] Step 1: Failed to parse JSON. Response: ${res.body}`);
        }
    }
    sleep(1);

    // 2. User A의 Access Token 획득
    // ------
    // AccessToken 받는 작업 필수
    //-------
    sleep(1);
    
    // ----------------------------------------------------
    // (2) 워크스페이스 및 프로젝트/보드 생성 (USER & PROJECT API)
    // ----------------------------------------------------

    // 3. 워크스페이스 생성 (User A 권한으로)
    if (!authTokenA) return; 
    res = http.post(`${USER_API_BASE_URL}/api/workspaces/create`,
        JSON.stringify({ 
            workspaceName: `WS ${vuId} - ${shortId}`, 
            workspaceDescription: 'Load Test WS',
            isPublic: true
        }), 
        { tags: { name: 'createWorkspace' }, headers: { 'Authorization': authTokenA, 'Content-Type': 'application/json' } }
    );
    if (res.status !== 201 && res.status !== 200) { 
        console.error(`VU ${vuId}: [ERROR] Step 3 (Create Workspace) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '3. Workspace Created': (r) => r.status === 201 || r.status === 200 }); 
    if (res.status === 201 || res.status === 200) { 
        try {
            workspaceId = res.json().workspaceId; 
            if (!workspaceId) {
                console.error(`VU ${vuId}: [ERROR] Step 3: Failed to extract 'workspaceId'. Response: ${res.body}`);
            }
        } catch (e) {
            console.error(`VU ${vuId}: [ERROR] Step 3: Failed to parse JSON. Response: ${res.body}`);
        }
    }
    sleep(1);

    // 4. 프로젝트 생성 (User A 권한으로)
    if (!workspaceId) return; 
    res = http.post(`${PROJECT_API_BASE_URL}/api/projects`,
        JSON.stringify({ 
            workspaceId: workspaceId, 
            name: `Project ${vuId} - ${shortId}`, 
            description: 'Test Project' 
        }),
        { tags: { name: 'createProject' }, headers: { 'Authorization': authTokenA, 'Content-Type': 'application/json' } }
    );
    if (res.status !== 201) {
        console.error(`VU ${vuId}: [ERROR] Step 4 (Create Project) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '4. Project Created': (r) => r.status === 201 });
    if (res.status === 201) {
        try {
            // Project API ID 추출 수정: data 객체 내의 projectId
            projectId = res.json().data?.projectId; 
            if (!projectId) {
                console.error(`VU ${vuId}: [ERROR] Step 4: Failed to extract 'projectId' from 'data'. Response: ${res.body}`);
            }
        } catch (e) {
            console.error(`VU ${vuId}: [ERROR] Step 4: Failed to parse JSON. Response: ${res.body}`);
        }
    }
    sleep(1);

    // 5. 보드 생성 (User A 권한으로) - customFields 없이 성공하는 버전 유지
    if (!projectId) return; 
    res = http.post(`${PROJECT_API_BASE_URL}/api/boards`,
        JSON.stringify({ 
            projectId: projectId, 
            title: `To Do Board - VU ${vuId}`, 
            content: 'Initial task in the project.',
        }),
        { tags: { name: 'createBoard' }, headers: { 'Authorization': authTokenA, 'Content-Type': 'application/json' } }
    );
    if (res.status !== 201) {
        console.error(`VU ${vuId}: [ERROR] Step 5 (Create Board) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '5. Board Created': (r) => r.status === 201 });
    if (res.status === 201) {
        try {
            // Project API ID 추출 수정: data 객체 내의 boardId
            boardId = res.json().data?.boardId; 
            if (!boardId) {
                console.error(`VU ${vuId}: [ERROR] Step 5: Failed to extract 'boardId' from 'data'. Response: ${res.body}`);
            }
        } catch (e) {
            console.error(`VU ${vuId}: [ERROR] Step 5: Failed to parse JSON. Response: ${res.body}`);
        }
    }
    sleep(1);
    
    // ----------------------------------------------------
    // (3) 사용자 B 생성, 토큰 획득 및 초대 (USER & PROJECT API)
    // ----------------------------------------------------

    // 6. 사용자 B 생성
    res = http.post(`${USER_API_BASE_URL}/api/users`, 
        JSON.stringify({ 
            email: userBEmail, 
            googleId: googleIdB,
            provider: "google" 
        }), 
        { tags: { name: 'createUserB' }, headers: { 'Content-Type': 'application/json' } }
    );
    if (res.status !== 201 && res.status !== 200) {
        console.error(`VU ${vuId}: [ERROR] Step 6 (User B Created) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '6. User B Created': (r) => r.status === 201 || r.status === 200 });
    if (res.status === 201 || res.status === 200) {
        try {
            userBId = res.json().userId; 
            if (!userBId) {
                console.error(`VU ${vuId}: [ERROR] Step 6: Failed to extract 'userId'. Response: ${res.body}`);
            }
        } catch (e) {
            console.error(`VU ${vuId}: [ERROR] Step 6: Failed to parse JSON. Response: ${res.body}`);
        }
    }
    sleep(1);

    // 7. User B의 Access Token 획득
    if (userBId) {
        res = http.get(`${USER_API_BASE_URL}/api/users/test/${userBId}`, 
            { tags: { name: 'getAccessTokenB' } }
        );
        if (res.status !== 200) {
            console.error(`VU ${vuId}: [ERROR] Step 7 (Token B) failed. Status: ${res.status}. Response: ${res.body}`);
        }
        check(res, { '7. Token B Status 200': (r) => r.status === 200 });
        if (res.status === 200) {
            const token = res.body;
            if (token && token.length > 10) { 
                 authTokenB = `Bearer ${token}`; 
            } else {
                 console.error(`VU ${vuId}: [ERROR] Step 7: Response body was too short. Expected JWT token. Raw Response: ${res.body}`);
            }
        }
    }
    sleep(1);

    // 8. 워크스페이스에 사용자 B 초대 (User A 권한으로)
    if (!workspaceId) return; 
    res = http.post(`${USER_API_BASE_URL}/api/workspaces/${workspaceId}/members/invite`,
        // 'query: must not be null' 오류 해결을 위해 'userId' 대신 'query' 필드에 'userBEmail' 사용
        JSON.stringify({ 
            query: userBEmail, 
            role: 'MEMBER' 
        }),
        { tags: { name: 'inviteUserB' }, headers: { 'Authorization': authTokenA, 'Content-Type': 'application/json' } }
    );
    if (res.status !== 200 && res.status !== 201) {
        console.error(`VU ${vuId}: [ERROR] Step 8 (Invite User B) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '8. User B Invited': (r) => r.status === 200 || r.status === 201 });
    sleep(1);
    
    // 9. 프로젝트 멤버 목록 조회 (User A 권한으로) - 명세에 따라 GET으로 유지
    if (!projectId) return; 
    res = http.get(`${PROJECT_API_BASE_URL}/api/projects/${projectId}/members`,
        { tags: { name: 'getProjectMembers' }, headers: { 'Authorization': authTokenA, 'Content-Type': 'application/json' } }
    );
    if (res.status !== 200) {
        console.error(`VU ${vuId}: [ERROR] Step 9 (Get Project Members) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '9. Project Members Retrieved (GET)': (r) => r.status === 200 });
    sleep(1);


    // 10. 보드에 참여자 B 추가 (User A 권한으로)
    if (!boardId) return; 
    res = http.post(`${PROJECT_API_BASE_URL}/api/participants`,
        JSON.stringify({ 
            boardId: boardId, 
            userIds: [userBId] 
        }),
        { tags: { name: 'addParticipantB' }, headers: { 'Authorization': authTokenA, 'Content-Type': 'application/json' } }
    );
    if (res.status !== 200 && res.status !== 201) {
        console.error(`VU ${vuId}: [ERROR] Step 10 (Add Participant B) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '10. Participant B Added': (r) => r.status === 200 || r.status === 201 });
    sleep(1);

    // ----------------------------------------------------
    // (4) 협업 및 관리 기능 테스트 (PROJECT API)
    // ----------------------------------------------------
    
    // 11. 댓글 작성 (User B 권한으로)
    if (authTokenB && boardId) {
        res = http.post(`${PROJECT_API_BASE_URL}/api/comments`,
            JSON.stringify({ 
                boardId: boardId, 
                content: `Hello, from User B ${vuId}`
            }),
            { tags: { name: 'createComment' }, headers: { 'Authorization': authTokenB, 'Content-Type': 'application/json' } }
        );
        if (res.status !== 201) {
            console.error(`VU ${vuId}: [ERROR] Step 11 (Comment Created) failed. Status: ${res.status}. Response: ${res.body}`);
        }
        check(res, { '11. Comment Created': (r) => r.status === 201 });

        if (res.status === 201) {
            try {
                // Project API ID 추출 수정: data 객체 내의 commentId
                commentId = res.json().data?.commentId; 
                if (!commentId) {
                    console.error(`VU ${vuId}: [ERROR] Step 11: Failed to extract 'commentId' from 'data'. Response: ${res.body}`);
                }
            } catch (e) {
                console.error(`VU ${vuId}: [ERROR] Step 11: Failed to parse JSON. Response: ${res.body}`);
            }
        }

    } else {
        console.error(`VU ${vuId}: [SKIP] Step 11 skipped (Token B or Board ID missing).`);
    }
    sleep(1);

    // 12. 보드 이동 (User A 권한으로) - API 예시의 필드 이름/값으로 수정
    // groupByFieldName: 'stage', newFieldValue: 'in_progress' 사용
    if (!boardId || !projectId) return;
    res = http.put(`${PROJECT_API_BASE_URL}/api/boards/${boardId}/move`,
        JSON.stringify({ 
            projectId: projectId, 
            groupByFieldName: 'stage', // API 명세 예시 참고하여 수정
            newFieldValue: 'in_progress' // API 명세 예시 참고하여 수정
        }),
        { tags: { name: 'moveBoard' }, headers: { 'Authorization': authTokenA, 'Content-Type': 'application/json' } }
    );
    if (res.status !== 200) {
        console.error(`VU ${vuId}: [ERROR] Step 12 (Board Moved) failed. Status: ${res.status}. Response: ${res.body}`);
    }
    check(res, { '12. Board Moved': (r) => r.status === 200 });
    sleep(1);
}