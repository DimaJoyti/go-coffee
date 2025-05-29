import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginTrend = new Trend('login_duration');
const accountCreationTrend = new Trend('account_creation_duration');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 },   // Ramp up to 10 users
    { duration: '5m', target: 10 },   // Stay at 10 users
    { duration: '2m', target: 20 },   // Ramp up to 20 users
    { duration: '5m', target: 20 },   // Stay at 20 users
    { duration: '2m', target: 50 },   // Ramp up to 50 users
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],             // Custom error rate must be below 10%
  },
};

// Base URL - can be overridden with environment variable
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data
const testUsers = [];
let userCounter = 0;

export function setup() {
  console.log('Starting performance test setup...');
  
  // Health check
  const healthResponse = http.get(`${BASE_URL}/health`);
  check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
  });
  
  return { baseUrl: BASE_URL };
}

export default function (data) {
  const userId = `user_${__VU}_${__ITER}_${Date.now()}`;
  const email = `${userId}@loadtest.com`;
  
  // Test scenario: Create account, login, perform operations
  testAccountCreation(email);
  sleep(1);
  
  const token = testLogin(email);
  if (token) {
    sleep(1);
    testAccountOperations(token);
    sleep(1);
    testLogout(token);
  }
  
  sleep(Math.random() * 2 + 1); // Random sleep between 1-3 seconds
}

function testAccountCreation(email) {
  const payload = {
    email: email,
    phone: '+1234567890',
    first_name: 'Load',
    last_name: 'Test',
    password: 'LoadTest123!',
    account_type: 'personal',
    country: 'USA',
    accept_terms: true,
  };
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post(`${BASE_URL}/api/v1/accounts`, JSON.stringify(payload), params);
  
  const success = check(response, {
    'account creation status is 201': (r) => r.status === 201,
    'account creation response has id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  accountCreationTrend.add(response.timings.duration);
  errorRate.add(!success);
  
  if (!success) {
    console.error(`Account creation failed: ${response.status} - ${response.body}`);
  }
}

function testLogin(email) {
  const payload = {
    email: email,
    password: 'LoadTest123!',
    device_id: `device_${__VU}`,
    remember_me: false,
  };
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify(payload), params);
  
  const success = check(response, {
    'login status is 200': (r) => r.status === 200,
    'login response has access_token': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.access_token !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  loginTrend.add(response.timings.duration);
  errorRate.add(!success);
  
  if (success) {
    try {
      const body = JSON.parse(response.body);
      return body.access_token;
    } catch (e) {
      console.error(`Failed to parse login response: ${e}`);
      return null;
    }
  } else {
    console.error(`Login failed: ${response.status} - ${response.body}`);
    return null;
  }
}

function testAccountOperations(token) {
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  
  // Get account profile
  const profileResponse = http.get(`${BASE_URL}/api/v1/accounts/profile`, { headers });
  check(profileResponse, {
    'get profile status is 200': (r) => r.status === 200,
  });
  
  // Update account
  const updatePayload = {
    first_name: 'Updated',
    last_name: 'Name',
  };
  
  const updateResponse = http.put(
    `${BASE_URL}/api/v1/accounts/profile`,
    JSON.stringify(updatePayload),
    { headers }
  );
  
  check(updateResponse, {
    'update profile status is 200': (r) => r.status === 200,
  });
  
  // Get security events
  const securityResponse = http.get(`${BASE_URL}/api/v1/accounts/security-events`, { headers });
  check(securityResponse, {
    'get security events status is 200': (r) => r.status === 200,
  });
  
  // Test payments endpoint (if enabled)
  testPaymentsEndpoints(headers);
  
  // Test yield endpoint (if enabled)
  testYieldEndpoints(headers);
}

function testPaymentsEndpoints(headers) {
  // Get supported currencies
  const currenciesResponse = http.get(`${BASE_URL}/api/v1/payments/currencies`, { headers });
  check(currenciesResponse, {
    'get currencies status is 200': (r) => r.status === 200,
  });
  
  // Get payment methods
  const methodsResponse = http.get(`${BASE_URL}/api/v1/payments/methods`, { headers });
  check(methodsResponse, {
    'get payment methods status is 200': (r) => r.status === 200,
  });
  
  // Get transaction history
  const historyResponse = http.get(`${BASE_URL}/api/v1/payments/transactions`, { headers });
  check(historyResponse, {
    'get transaction history status is 200': (r) => r.status === 200,
  });
}

function testYieldEndpoints(headers) {
  // Get yield opportunities
  const opportunitiesResponse = http.get(`${BASE_URL}/api/v1/yield/opportunities`, { headers });
  check(opportunitiesResponse, {
    'get yield opportunities status is 200': (r) => r.status === 200,
  });
  
  // Get portfolio
  const portfolioResponse = http.get(`${BASE_URL}/api/v1/yield/portfolio`, { headers });
  check(portfolioResponse, {
    'get yield portfolio status is 200': (r) => r.status === 200,
  });
}

function testLogout(token) {
  const headers = {
    'Authorization': `Bearer ${token}`,
  };
  
  const response = http.post(`${BASE_URL}/api/v1/auth/logout`, null, { headers });
  
  const success = check(response, {
    'logout status is 200': (r) => r.status === 200,
  });
  
  errorRate.add(!success);
  
  if (!success) {
    console.error(`Logout failed: ${response.status} - ${response.body}`);
  }
}

export function teardown(data) {
  console.log('Performance test completed');
  
  // Final health check
  const healthResponse = http.get(`${data.baseUrl}/health`);
  check(healthResponse, {
    'final health check status is 200': (r) => r.status === 200,
  });
}
