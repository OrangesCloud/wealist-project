import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend } from 'k6/metrics';

// Custom trends to measure duration of specific requests
const authTestTrend = new Trend('T01_AuthTest');
const getMyProfileTrend = new Trend('T02_GetMyProfile');
const createWorkspaceTrend = new Trend('T03_CreateWorkspace');
const getWorkspaceTrend = new Trend('T04_GetWorkspace');

export const options = {
  stages: [
    { duration: '30s', target: 10 }, // Ramp up to 10 VUs over 30s
    { duration: '1m', target: 10 },  // Stay at 10 VUs for 1 minute
    { duration: '10s', target: 0 },   // Ramp down to 0 VUs
  ],
  thresholds: {
    'http_req_failed': ['rate<0.01'], // http errors should be less than 1%
    'http_req_duration': ['p(95)<500'], // 95% of requests should be below 500ms
    'T01_AuthTest': ['p(95)<300'],
    'T02_GetMyProfile': ['p(95)<300'],
    'T03_CreateWorkspace': ['p(95)<400'],
    'T04_GetWorkspace': ['p(95)<300'],
  },
};

const BASE_URL = 'http://localhost'; // Nginx is on port 80

export default function () {
  let accessToken;
  let userId;

  // Group 1: Authentication
  group('1. Authentication', function () {
    const res = http.get(`${BASE_URL}/api/auth/test`);

    check(res, {
      'Auth: status is 200': (r) => r.status === 200,
      'Auth: response contains accessToken': (r) => r.json('accessToken') !== null,
    });
    
    authTestTrend.add(res.timings.duration);

    if (res.status === 200 && res.json('accessToken')) {
      accessToken = res.json('accessToken');
      userId = res.json('user_id');
    }
  });

  sleep(1);

  if (accessToken) {
    const headers = {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    };

    // Group 2: User and Profile retrieval
    group('2. Get User Profile & Info', function () {
      const profileRes = http.get(`${BASE_URL}/api/profiles/me`, { headers });
      check(profileRes, {
        'Profile: status is 200': (r) => r.status === 200,
        'Profile: user_id matches': (r) => r.json('user_id') === userId,
      });
      getMyProfileTrend.add(profileRes.timings.duration);

      const userRes = http.get(`${BASE_URL}/api/users/me`, { headers });
      check(userRes, {
        'User: status is 200': (r) => r.status === 200,
        'User: user_id matches': (r) => r.json('user_id') === userId,
      });
    });

    sleep(2); // Simulate user looking at their profile

    let workspaceId;

    // Group 3: Workspace Creation
    group('3. Create Workspace', function () {
      const workspaceName = `k6-test-ws-${__VU}-${__ITER}`;
      const payload = JSON.stringify({
        name: workspaceName,
        description: 'Workspace created by k6 test',
      });

      const createRes = http.post(`${BASE_URL}/api/workspaces`, payload, { headers });

      check(createRes, {
        'Create Workspace: status is 200': (r) => r.status === 200,
        'Create Workspace: response contains ID': (r) => r.json('id') !== null,
      });

      createWorkspaceTrend.add(createRes.timings.duration);
      
      if (createRes.status === 200 && createRes.json('id')) {
        workspaceId = createRes.json('id');
      }
    });

    sleep(1);

    // Group 4: Access Created Workspace
    if (workspaceId) {
      group('4. Get Workspace Details', function () {
        const getRes = http.get(`${BASE_URL}/api/workspaces/${workspaceId}`, { headers });
        check(getRes, {
          'Get Workspace: status is 200': (r) => r.status === 200,
          'Get Workspace: ID matches': (r) => r.json('id') === workspaceId,
        });
        getWorkspaceTrend.add(getRes.timings.duration);
      });
    }
  }
}
