import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,       // 10 virtual users
  duration: '30s', // for 30 seconds
};

export default function () {
  const res = http.get('http://localhost/health');
  check(res, {
    'is status 200': (r) => r.status === 200,
  });
  sleep(1); // Think time between requests
}
