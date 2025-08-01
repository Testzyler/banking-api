// Load testing script using K6 for auth and dashboard
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';

import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

// Custom metrics to measure performance
const authResponseTime = new Trend('auth_response_time');
const dashboardResponseTime = new Trend('dashboard_response_time');
const errorRate = new Rate('error_rate');
const successfulLogins = new Counter('successful_logins');
const successfulDashboardCalls = new Counter('successful_dashboard_calls');

const BASE_URL = 'http://localhost:8080';

// Configuration: Max 500 concurrent connections on Fiber and DB
// To stay within safe limits, we'll target up to 400 VUs in test
export const options = {
    scenarios: {
      light_load: {
        executor: 'constant-vus',
        vus: 50,
        duration: '1m',
        tags: { scenario: 'light_load' },
      },
      normal_load: {
        executor: 'ramping-vus',
        startVUs: 0,
        stages: [
          { duration: '1m', target: 200 },
          { duration: '2m', target: 400 },
          { duration: '1m', target: 0 },
        ],
        tags: { scenario: 'normal_load' },
      },
      heavy_load: {
        executor: 'ramping-vus',
        startVUs: 0,
        stages: [
          { duration: '2m', target: 300 },
          { duration: '3m', target: 600 },
          { duration: '2m', target: 800 },
          { duration: '1m', target: 0 },
        ],
        tags: { scenario: 'heavy_load' },
      },
    },
  
    thresholds: {
      // Light Load
      'auth_response_time{scenario:light_load}': ['p(95)<4000', 'p(99)<6000'],
      'dashboard_response_time{scenario:light_load}': ['p(95)<2000'],
      'http_req_duration{scenario:light_load}': ['p(95)<3500'],
  
      // Normal Load
      'auth_response_time{scenario:normal_load}': ['p(95)<5000', 'p(99)<8000'],
      'dashboard_response_time{scenario:normal_load}': ['p(95)<5000', 'p(99)<9000'],
      'http_req_duration{scenario:normal_load}': ['p(95)<9000'],
  
      // Heavy Load
      'auth_response_time{scenario:heavy_load}': ['p(95)<12000', 'p(99)<15000'],
      'dashboard_response_time{scenario:heavy_load}': ['p(95)<7000', 'p(99)<10000'],
      'http_req_duration{scenario:heavy_load}': ['p(95)<12000'],
  
      // Error rate tolerance per scenario
      'error_rate{scenario:light_load}': ['rate<0.01'],
      'error_rate{scenario:normal_load}': ['rate<0.03'],
      'error_rate{scenario:heavy_load}': ['rate<0.05'],
    },
  };
  

// Load test users from file
const TEST_USERS = new SharedArray('users', function () {
    try {
        const file = open('./users.txt');
        return file
            .split('\n')
            .filter(Boolean)
            .map((username) => ({ username: username.trim(), pin: '123456' }));
    } catch (err) {
        console.error('Failed to load users.txt:', err);
        return [];
    }
});

function getRandomUser() {
    return TEST_USERS[Math.floor(Math.random() * TEST_USERS.length)];
}

function authenticate() {
    const user = getRandomUser();
    const payload = JSON.stringify({ username: user.username, pin: user.pin });
    const params = { headers: { 'Content-Type': 'application/json' }, tags: { endpoint: 'auth' } };

    const start = Date.now();
    const res = http.post(`${BASE_URL}/api/v1/auth/verify-pin`, payload, params);
    const duration = Date.now() - start;
    authResponseTime.add(duration);

    const ok = check(res, {
        'auth status is 200': (r) => r.status === 200,
        'auth returns token': (r) => {
            try {
                const data = JSON.parse(r.body);
                return data.data && data.data.token;
            } catch (_) {
                return false;
            }
        },
    });

    if (!ok) {
        errorRate.add(1);
        return null;
    }

    successfulLogins.add(1);
    const data = JSON.parse(res.body).data;
    return { token: data.token, userID: data.userID };
}

function testDashboard(authData) {
    if (!authData?.token) return;

    const params = {
        headers: {
            Authorization: `Bearer ${authData.token}`,
            'Content-Type': 'application/json',
        },
        tags: { endpoint: 'dashboard' },
    };

    const start = Date.now();
    const res = http.get(`${BASE_URL}/api/v1/dashboard/accounts`, params);
    const duration = Date.now() - start;
    dashboardResponseTime.add(duration);

    const ok = check(res, {
        'dashboard status is 200': (r) => r.status === 200,
        'dashboard returns data': (r) => {
            try {
                const data = JSON.parse(r.body);
                return data.data !== undefined;
            } catch (_) {
                return false;
            }
        },
    });

    if (!ok) {
        errorRate.add(1);
    } else {
        successfulDashboardCalls.add(1);
    }
}

export default function () {
    const authData = authenticate();
    if (authData) {
        testDashboard(authData);
    }
    sleep(Math.random() * 2 + 1);
}

export function handleSummary(data) {
    return {
        'summary.html': htmlReport(data, { title: 'K6 Load Test Report' }),
        stdout: textSummary(data, { indent: ' ', enableColors: true }),
    };
}