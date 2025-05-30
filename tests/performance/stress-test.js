import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration');
const requestCount = new Counter('requests');

// Stress test configuration - gradually increase load to find breaking point
export const options = {
  stages: [
    { duration: '1m', target: 10 },    // Warm up
    { duration: '2m', target: 50 },    // Ramp up to 50 users
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '2m', target: 200 },   // Ramp up to 200 users
    { duration: '2m', target: 300 },   // Ramp up to 300 users
    { duration: '2m', target: 400 },   // Ramp up to 400 users
    { duration: '2m', target: 500 },   // Ramp up to 500 users
    { duration: '5m', target: 500 },   // Stay at 500 users for 5 minutes
    { duration: '2m', target: 600 },   // Push to 600 users
    { duration: '2m', target: 700 },   // Push to 700 users
    { duration: '2m', target: 800 },   // Push to 800 users
    { duration: '5m', target: 800 },   // Stay at 800 users
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2s (relaxed for stress test)
    http_req_failed: ['rate<0.2'],     // Error rate must be below 20% (relaxed for stress test)
    errors: ['rate<0.2'],              // Custom error rate must be below 20%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Pre-created test accounts for stress testing
const TEST_ACCOUNTS = [
  { email: 'stress1@test.com', password: 'StressTest123!' },
  { email: 'stress2@test.com', password: 'StressTest123!' },
  { email: 'stress3@test.com', password: 'StressTest123!' },
  { email: 'stress4@test.com', password: 'StressTest123!' },
  { email: 'stress5@test.com', password: 'StressTest123!' },
];

export function setup() {
  console.log('Starting stress test setup...');
  
  // Health check
  const healthResponse = http.get(`${BASE_URL}/health`);
  check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
  });
  
  // Create test accounts if they don't exist
  createTestAccounts();
  
  return { baseUrl: BASE_URL };
}

function createTestAccounts() {
  console.log('Creating test accounts...');
  
  TEST_ACCOUNTS.forEach((account, index) => {
    const payload = {
      email: account.email,
      phone: `+123456789${index}`,
      first_name: 'Stress',
      last_name: `Test${index}`,
      password: account.password,
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
    
    // Don't fail if account already exists
    if (response.status !== 201 && response.status !== 409) {
      console.warn(`Failed to create test account ${account.email}: ${response.status}`);
    }
    
    sleep(0.1); // Small delay between account creations
  });
}

export default function (data) {
  requestCount.add(1);
  
  // Use random test account
  const account = TEST_ACCOUNTS[Math.floor(Math.random() * TEST_ACCOUNTS.length)];
  
  // Stress test scenario: Heavy concurrent operations
  const token = stressLogin(account);
  if (token) {
    // Perform multiple concurrent operations
    stressAccountOperations(token);
    stressPaymentOperations(token);
    stressYieldOperations(token);
    stressTradingOperations(token);
    
    // Random logout (not always)
    if (Math.random() > 0.3) {
      stressLogout(token);
    }
  }
  
  // Minimal sleep to maximize stress
  sleep(Math.random() * 0.5);
}

function stressLogin(account) {
  const payload = {
    email: account.email,
    password: account.password,
    device_id: `stress_device_${__VU}_${__ITER}`,
    remember_me: false,
  };
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify(payload), params);
  
  requestDuration.add(response.timings.duration);
  
  const success = check(response, {
    'stress login status is 200': (r) => r.status === 200,
  });
  
  errorRate.add(!success);
  
  if (success) {
    try {
      const body = JSON.parse(response.body);
      return body.access_token;
    } catch (e) {
      return null;
    }
  }
  
  return null;
}

function stressAccountOperations(token) {
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  
  // Rapid-fire account operations
  const operations = [
    () => http.get(`${BASE_URL}/api/v1/accounts/profile`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/accounts/security-events`, { headers }),
    () => http.put(`${BASE_URL}/api/v1/accounts/profile`, JSON.stringify({
      first_name: `Stress${Math.random()}`,
    }), { headers }),
  ];
  
  // Execute random operations
  for (let i = 0; i < 3; i++) {
    const operation = operations[Math.floor(Math.random() * operations.length)];
    const response = operation();
    
    requestDuration.add(response.timings.duration);
    errorRate.add(response.status >= 400);
  }
}

function stressPaymentOperations(token) {
  const headers = {
    'Authorization': `Bearer ${token}`,
  };
  
  // Concurrent payment-related requests
  const operations = [
    () => http.get(`${BASE_URL}/api/v1/payments/currencies`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/payments/methods`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/payments/transactions`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/payments/balance`, { headers }),
  ];
  
  operations.forEach(operation => {
    const response = operation();
    requestDuration.add(response.timings.duration);
    errorRate.add(response.status >= 400);
  });
}

function stressYieldOperations(token) {
  const headers = {
    'Authorization': `Bearer ${token}`,
  };
  
  // Yield farming operations
  const operations = [
    () => http.get(`${BASE_URL}/api/v1/yield/opportunities`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/yield/portfolio`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/yield/strategies`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/yield/analytics`, { headers }),
  ];
  
  operations.forEach(operation => {
    const response = operation();
    requestDuration.add(response.timings.duration);
    errorRate.add(response.status >= 400);
  });
}

function stressTradingOperations(token) {
  const headers = {
    'Authorization': `Bearer ${token}`,
  };
  
  // Trading operations
  const operations = [
    () => http.get(`${BASE_URL}/api/v1/trading/markets`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/trading/orders`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/trading/portfolio`, { headers }),
    () => http.get(`${BASE_URL}/api/v1/trading/analytics`, { headers }),
  ];
  
  operations.forEach(operation => {
    const response = operation();
    requestDuration.add(response.timings.duration);
    errorRate.add(response.status >= 400);
  });
}

function stressLogout(token) {
  const headers = {
    'Authorization': `Bearer ${token}`,
  };
  
  const response = http.post(`${BASE_URL}/api/v1/auth/logout`, null, { headers });
  
  requestDuration.add(response.timings.duration);
  errorRate.add(response.status >= 400);
}

export function teardown(data) {
  console.log('Stress test completed');
  
  // Final system check
  const healthResponse = http.get(`${data.baseUrl}/health`);
  const metricsResponse = http.get(`${data.baseUrl}/metrics`);
  
  console.log(`Final health check: ${healthResponse.status}`);
  console.log(`Metrics endpoint: ${metricsResponse.status}`);
  
  // Log some final statistics
  console.log('=== STRESS TEST RESULTS ===');
  console.log(`Total requests: ${requestCount.count}`);
  console.log(`Error rate: ${(errorRate.rate * 100).toFixed(2)}%`);
  console.log(`Average response time: ${requestDuration.avg.toFixed(2)}ms`);
  console.log(`95th percentile: ${requestDuration.p95.toFixed(2)}ms`);
  console.log(`99th percentile: ${requestDuration.p99.toFixed(2)}ms`);
}
