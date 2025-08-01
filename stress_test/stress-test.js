import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';

import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.1/index.js";


// Custom metrics
const dashboardResponseTime = new Trend('dashboard_response_time');
const authResponseTime = new Trend('auth_response_time');
const errorRate = new Rate('error_rate');
const successfulLogins = new Counter('successful_logins');
const successfulDashboardCalls = new Counter('successful_dashboard_calls');

const BASE_URL = 'http://localhost:8080';
// export const options = {
//     scenarios: {
//       // Light load - baseline test
//        light_load: {
//         executor: 'constant-vus',
//         vus: 5,
//         duration: '2m',
//         tags: { scenario: 'light_load' },
//       },
//       // light_load: {
//       //   executor: 'constant-vus',
//       //   vus: 5,
//       //   duration: '2m',
//       //   tags: { scenario: 'light_load' },
//       // },

//       // Normal load - typical usage
//       // normal_load: {
//       //   executor: 'ramping-vus',
//       //   startVUs: 0,
//       //   stages: [
//       //     { duration: '1m', target: 20 }, // Ramp up
//       //     { duration: '3m', target: 20 }, // Stay steady
//       //     { duration: '1m', target: 0 },  // Ramp down
//       //   ],
//       //   tags: { scenario: 'normal_load' },
//       // },

//       // Heavy load - stress test
//       // heavy_load: {
//       //   executor: 'ramping-vus',
//       //   startVUs: 0,
//       //   stages: [
//       //     { duration: '2m', target: 50 },  // Ramp up
//       //     { duration: '5m', target: 50 },  // Stay steady
//       //     { duration: '2m', target: 100 }, // Peak load
//       //     { duration: '2m', target: 50 },  // Scale back
//       //     { duration: '1m', target: 0 },   // Ramp down
//       //   ],
//       //   tags: { scenario: 'heavy_load' },
//       // },
//     },

//     thresholds: {
//       // Response time thresholds
//       'dashboard_response_time': [
//         'p(95)<2000', // 95% of requests should be under 2s (current baseline)
//         'p(99)<5000', // 99% of requests should be under 5s
//       ],
//       'auth_response_time': [
//         'p(95)<500',  // Auth should be fast
//         'p(99)<1000',
//       ],

//       // Error rate thresholds
//       'error_rate': ['rate<0.05'], // Less than 5% error rate

//       // HTTP duration thresholds
//       'http_req_duration': ['p(95)<3000'],
//       'http_req_failed': ['rate<0.05'],
//     },
//   };

export const options = {
    scenarios: {
        light_load: {
            executor: 'constant-vus',
            vus: 10,
            duration: '1m',
            tags: { scenario: 'light_load' },
        },
        // Normal load - typical usage
        normal_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 20 }, // Ramp up
                { duration: '3m', target: 20 }, // Stay steady
                { duration: '1m', target: 0 },  // Ramp down
            ],
            tags: { scenario: 'normal_load' },
        },
        // Heavy load - stress test
        heavy_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 50 },  // Ramp up
                { duration: '5m', target: 50 },  // Stay steady
                { duration: '2m', target: 100 }, // Peak load
                { duration: '2m', target: 50 },  // Scale back
                { duration: '1m', target: 0 },   // Ramp down
            ],
            tags: { scenario: 'heavy_load' },
        },
    },

    thresholds: {
        // Response time thresholds
        'dashboard_response_time': [
            'p(95)<2000', // 95% of requests should be under 2s (current baseline)
            'p(99)<5000', // 99% of requests should be under 5s
        ],
        'auth_response_time': [
            'p(95)<500',  // Auth should be fast
            'p(99)<1000',
        ],

        // Error rate thresholds
        'error_rate': ['rate<0.05'], // Less than 5% error rate

        // HTTP duration thresholds
        'http_req_duration': ['p(95)<3000'],
        'http_req_failed': ['rate<0.05'],
    },
};

const TEST_USERS = new SharedArray('users', function () {
    try {
        // Read users from file
        const usersFile = open('./users.txt');
        const usernames = usersFile.split('\n').filter(line => line.trim().length > 0);

        // Convert usernames to user objects with default PIN
        return usernames.map(username => ({
            username: username.trim(),
            pin: '123456'
        }));
    } catch (error) {
        console.error('Could not load sample-users.txt, falling back to default users:', error);
        return;
    }
});

// Get random test user
function getRandomUser() {
    return TEST_USERS[Math.floor(Math.random() * TEST_USERS.length)];
}

// Authenticate and get token
function authenticate() {
    const user = getRandomUser();
    const payload = {
        username: user.username,
        pin: user.pin
    };

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        tags: { endpoint: 'auth' },
    };

    const authStart = Date.now();
    const response = http.post(`${BASE_URL}/api/v1/auth/verify-pin`, JSON.stringify(payload), params);
    const authDuration = Date.now() - authStart;

    authResponseTime.add(authDuration);

    const authSuccess = check(response, {
        'auth status is 200': (r) => r.status === 200,
        'auth has token': (r) => {
            try {
                const data = JSON.parse(r.body);
                return data.data && data.data.token;
            } catch (e) {
                return false;
            }
        },
    });

    if (!authSuccess) {
        errorRate.add(1);
        console.error(`Auth failed for ${user.username}: ${response.status} - ${response.body}`);
        return null;
    }

    successfulLogins.add(1);

    try {
        const resp = JSON.parse(response.body);
        return {
            token: resp.data.token,
            refreshToken: resp.data.refreshToken,
            userID: resp.data.userID
        };
    } catch (e) {
        errorRate.add(1);
        return null;
    }
}

// Test dashboard endpoint
function testDashboard(authData) {
    if (!authData || !authData.token) {
        errorRate.add(1);
        return;
    }

    const params = {
        headers: {
            'Authorization': `Bearer ${authData.token}`,
            'Content-Type': 'application/json',
        },
        tags: { endpoint: 'dashboard' },
    };

    const payload = {
        userID: authData.userID,
    };

    const url = `${BASE_URL}/api/v1/dashboard/accounts`;
    const dashboardStart = Date.now();
    const response = http.get(url, params);
    const dashboardDuration = Date.now() - dashboardStart;

    dashboardResponseTime.add(dashboardDuration);
    const dashboardSuccess = check(response, {
        'dashboard status is 200': (r) => r.status === 200,
        'dashboard has data': (r) => {
            try {
                const data = JSON.parse(r.body);
                return data.data !== undefined;
            } catch (e) {
                return false;
            }
        },
        'dashboard response time < 5s': (r) => dashboardDuration < 5000,
    });

    if (!dashboardSuccess) {
        errorRate.add(1);
        console.error(`Dashboard failed: ${response.status} - ${response.body}`);
    } else {
        successfulDashboardCalls.add(1);
    }

    if (dashboardDuration > 2000) {
        console.warn(`Slow dashboard response: ${dashboardDuration}ms`);
    }
}

export function handleSummary(data) {
    return {
        "summary.html": htmlReport(data, {
            title: "Load Test K6 - Result",
        }),
        stdout: textSummary(data, { indent: " ", enableColors: true }),
    };
}

// Main test function
export default function () {
    // Authenticate
    const authData = authenticate();
    if (authData) {
        // Test dashboard
        testDashboard(authData);

        // Small delay between requests
        sleep(Math.random() * 2 + 1); // 1-3 seconds
    } else {
        // If auth fails, wait before retrying
        sleep(2);
    }
}
