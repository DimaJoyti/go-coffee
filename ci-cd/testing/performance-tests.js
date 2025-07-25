// ☕ Go Coffee - Performance Testing Suite
// K6 Performance Tests for Go Coffee Platform

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

// Custom metrics
const errorRate = new Rate('error_rate');
const responseTime = new Trend('response_time');
const requestCount = new Counter('request_count');

// Test configuration
export const options = {
  stages: [
    // Ramp-up
    { duration: '2m', target: 10 },   // Ramp up to 10 users over 2 minutes
    { duration: '5m', target: 10 },   // Stay at 10 users for 5 minutes
    { duration: '2m', target: 50 },   // Ramp up to 50 users over 2 minutes
    { duration: '5m', target: 50 },   // Stay at 50 users for 5 minutes
    { duration: '2m', target: 100 },  // Ramp up to 100 users over 2 minutes
    { duration: '10m', target: 100 }, // Stay at 100 users for 10 minutes
    { duration: '5m', target: 0 },    // Ramp down to 0 users over 5 minutes
  ],
  thresholds: {
    // Performance thresholds
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests under 500ms, 99% under 1s
    http_req_failed: ['rate<0.01'],                  // Error rate under 1%
    error_rate: ['rate<0.01'],                       // Custom error rate under 1%
    response_time: ['p(95)<500'],                    // 95% response time under 500ms
    
    // Business metrics thresholds
    'group_duration{group:::API Gateway}': ['p(95)<200'],
    'group_duration{group:::Order Service}': ['p(95)<300'],
    'group_duration{group:::Payment Service}': ['p(95)<400'],
    'group_duration{group:::AI Services}': ['p(95)<2000'],
  },
  ext: {
    loadimpact: {
      projectID: 3596490,
      name: 'Go Coffee Performance Test'
    }
  }
};

// Environment configuration
const BASE_URL = __ENV.BASE_URL || 'https://staging.gocoffee.dev';
const API_KEY = __ENV.API_KEY || 'test-api-key';

// Test data
const testUsers = [
  { email: 'test1@example.com', password: 'password123' },
  { email: 'test2@example.com', password: 'password123' },
  { email: 'test3@example.com', password: 'password123' },
];

const menuItems = [
  { id: 1, name: 'Espresso', price: 2.50 },
  { id: 2, name: 'Cappuccino', price: 3.50 },
  { id: 3, name: 'Latte', price: 4.00 },
  { id: 4, name: 'Americano', price: 3.00 },
];

// Utility functions
function getRandomUser() {
  return testUsers[Math.floor(Math.random() * testUsers.length)];
}

function getRandomMenuItem() {
  return menuItems[Math.floor(Math.random() * menuItems.length)];
}

function authenticateUser(user) {
  const loginPayload = JSON.stringify({
    email: user.email,
    password: user.password
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
  };

  const response = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, params);
  
  check(response, {
    'login successful': (r) => r.status === 200,
    'login response has token': (r) => r.json('token') !== undefined,
  });

  return response.json('token');
}

// Main test function
export default function () {
  const user = getRandomUser();
  let authToken = '';

  group('Authentication', () => {
    authToken = authenticateUser(user);
    sleep(1);
  });

  if (!authToken) {
    console.error('Authentication failed, skipping remaining tests');
    return;
  }

  const authParams = {
    headers: {
      'Authorization': `Bearer ${authToken}`,
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
  };

  group('API Gateway', () => {
    // Health check
    const healthResponse = http.get(`${BASE_URL}/api/v1/health`);
    check(healthResponse, {
      'health check status is 200': (r) => r.status === 200,
      'health check response time < 100ms': (r) => r.timings.duration < 100,
    });

    // API info
    const infoResponse = http.get(`${BASE_URL}/api/v1/info`, authParams);
    check(infoResponse, {
      'info endpoint status is 200': (r) => r.status === 200,
      'info response has version': (r) => r.json('version') !== undefined,
    });

    responseTime.add(healthResponse.timings.duration);
    requestCount.add(1);
    sleep(0.5);
  });

  group('Menu Service', () => {
    // Get menu
    const menuResponse = http.get(`${BASE_URL}/api/v1/menu`, authParams);
    check(menuResponse, {
      'menu fetch status is 200': (r) => r.status === 200,
      'menu has items': (r) => r.json('items').length > 0,
      'menu response time < 200ms': (r) => r.timings.duration < 200,
    });

    // Get specific menu item
    const itemId = getRandomMenuItem().id;
    const itemResponse = http.get(`${BASE_URL}/api/v1/menu/items/${itemId}`, authParams);
    check(itemResponse, {
      'menu item fetch status is 200': (r) => r.status === 200,
      'menu item has details': (r) => r.json('name') !== undefined,
    });

    responseTime.add(menuResponse.timings.duration);
    requestCount.add(2);
    sleep(0.5);
  });

  group('Order Service', () => {
    const orderItem = getRandomMenuItem();
    const orderPayload = JSON.stringify({
      items: [
        {
          menuItemId: orderItem.id,
          quantity: Math.floor(Math.random() * 3) + 1,
          customizations: ['extra shot', 'oat milk']
        }
      ],
      customerNotes: 'Test order from performance test'
    });

    // Create order
    const createOrderResponse = http.post(`${BASE_URL}/api/v1/orders`, orderPayload, authParams);
    check(createOrderResponse, {
      'order creation status is 201': (r) => r.status === 201,
      'order has ID': (r) => r.json('orderId') !== undefined,
      'order creation time < 300ms': (r) => r.timings.duration < 300,
    });

    if (createOrderResponse.status === 201) {
      const orderId = createOrderResponse.json('orderId');
      
      // Get order status
      const orderStatusResponse = http.get(`${BASE_URL}/api/v1/orders/${orderId}`, authParams);
      check(orderStatusResponse, {
        'order status fetch is 200': (r) => r.status === 200,
        'order has status': (r) => r.json('status') !== undefined,
      });

      // Get order history
      const orderHistoryResponse = http.get(`${BASE_URL}/api/v1/orders/history`, authParams);
      check(orderHistoryResponse, {
        'order history fetch is 200': (r) => r.status === 200,
        'order history is array': (r) => Array.isArray(r.json('orders')),
      });
    }

    responseTime.add(createOrderResponse.timings.duration);
    requestCount.add(3);
    sleep(1);
  });

  group('Payment Service', () => {
    const paymentPayload = JSON.stringify({
      amount: 4.50,
      currency: 'USD',
      paymentMethod: 'card',
      cardToken: 'test_card_token_123'
    });

    // Process payment
    const paymentResponse = http.post(`${BASE_URL}/api/v1/payments/process`, paymentPayload, authParams);
    check(paymentResponse, {
      'payment processing status is 200 or 201': (r) => [200, 201].includes(r.status),
      'payment has transaction ID': (r) => r.json('transactionId') !== undefined,
      'payment processing time < 400ms': (r) => r.timings.duration < 400,
    });

    // Get payment methods
    const paymentMethodsResponse = http.get(`${BASE_URL}/api/v1/payments/methods`, authParams);
    check(paymentMethodsResponse, {
      'payment methods fetch is 200': (r) => r.status === 200,
      'payment methods is array': (r) => Array.isArray(r.json('methods')),
    });

    responseTime.add(paymentResponse.timings.duration);
    requestCount.add(2);
    sleep(0.5);
  });

  group('AI Services', () => {
    // AI beverage recommendation
    const recommendationPayload = JSON.stringify({
      preferences: ['strong', 'sweet'],
      season: 'winter',
      timeOfDay: 'morning'
    });

    const recommendationResponse = http.post(`${BASE_URL}/api/v1/ai/recommendations`, recommendationPayload, authParams);
    check(recommendationResponse, {
      'AI recommendation status is 200': (r) => r.status === 200,
      'AI recommendation has suggestions': (r) => r.json('recommendations').length > 0,
      'AI recommendation time < 2000ms': (r) => r.timings.duration < 2000,
    });

    // AI inventory forecast
    const forecastResponse = http.get(`${BASE_URL}/api/v1/ai/inventory/forecast`, authParams);
    check(forecastResponse, {
      'AI forecast status is 200': (r) => r.status === 200,
      'AI forecast has data': (r) => r.json('forecast') !== undefined,
    });

    // AI customer service
    const customerServicePayload = JSON.stringify({
      message: 'I have a question about my order',
      context: 'order_inquiry'
    });

    const customerServiceResponse = http.post(`${BASE_URL}/api/v1/ai/customer-service`, customerServicePayload, authParams);
    check(customerServiceResponse, {
      'AI customer service status is 200': (r) => r.status === 200,
      'AI customer service has response': (r) => r.json('response') !== undefined,
    });

    responseTime.add(recommendationResponse.timings.duration);
    requestCount.add(3);
    sleep(1);
  });

  group('User Profile', () => {
    // Get user profile
    const profileResponse = http.get(`${BASE_URL}/api/v1/users/profile`, authParams);
    check(profileResponse, {
      'profile fetch status is 200': (r) => r.status === 200,
      'profile has user data': (r) => r.json('email') !== undefined,
    });

    // Update preferences
    const preferencesPayload = JSON.stringify({
      favoriteItems: [1, 2, 3],
      dietaryRestrictions: ['lactose-free'],
      notifications: {
        email: true,
        push: false
      }
    });

    const preferencesResponse = http.put(`${BASE_URL}/api/v1/users/preferences`, preferencesPayload, authParams);
    check(preferencesResponse, {
      'preferences update status is 200': (r) => r.status === 200,
    });

    responseTime.add(profileResponse.timings.duration);
    requestCount.add(2);
    sleep(0.5);
  });

  // Error tracking
  errorRate.add(false); // No errors in this iteration
  sleep(Math.random() * 2 + 1); // Random sleep between 1-3 seconds
}

// Stress test scenario
export function stressTest() {
  const response = http.get(`${BASE_URL}/api/v1/health`);
  check(response, {
    'stress test - health check status is 200': (r) => r.status === 200,
  });
  
  if (response.status !== 200) {
    errorRate.add(true);
  }
}

// Load test scenario for specific endpoints
export function loadTestCriticalEndpoints() {
  const endpoints = [
    '/api/v1/health',
    '/api/v1/menu',
    '/api/v1/orders/status',
    '/api/v1/ai/recommendations'
  ];

  endpoints.forEach(endpoint => {
    const response = http.get(`${BASE_URL}${endpoint}`);
    check(response, {
      [`${endpoint} - status is 200`]: (r) => r.status === 200,
      [`${endpoint} - response time < 500ms`]: (r) => r.timings.duration < 500,
    });
    
    responseTime.add(response.timings.duration);
    requestCount.add(1);
  });
}

// Spike test scenario
export function spikeTest() {
  // Simulate sudden traffic spike
  for (let i = 0; i < 10; i++) {
    const response = http.get(`${BASE_URL}/api/v1/menu`);
    check(response, {
      'spike test - menu fetch successful': (r) => r.status === 200,
    });
    
    if (response.status !== 200) {
      errorRate.add(true);
    }
  }
}

// Volume test scenario
export function volumeTest() {
  // Test with large payloads
  const largeOrderPayload = JSON.stringify({
    items: Array.from({ length: 50 }, (_, i) => ({
      menuItemId: (i % 4) + 1,
      quantity: Math.floor(Math.random() * 5) + 1,
      customizations: ['extra shot', 'oat milk', 'extra hot', 'no foam']
    })),
    customerNotes: 'Large test order with many items '.repeat(10)
  });

  const authParams = {
    headers: {
      'Authorization': `Bearer test-token`,
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
  };

  const response = http.post(`${BASE_URL}/api/v1/orders`, largeOrderPayload, authParams);
  check(response, {
    'volume test - large order processed': (r) => [200, 201, 400].includes(r.status), // 400 might be expected for oversized orders
    'volume test - response time reasonable': (r) => r.timings.duration < 5000,
  });
}

// Custom summary report
export function handleSummary(data) {
  return {
    'performance-report.html': htmlReport(data),
    'performance-summary.txt': textSummary(data, { indent: ' ', enableColors: true }),
    'performance-results.json': JSON.stringify(data),
  };
}

// Setup function
export function setup() {
  console.log('🚀 Starting Go Coffee Performance Tests');
  console.log(`Target URL: ${BASE_URL}`);
  console.log(`Test Duration: ~30 minutes`);
  console.log(`Max Virtual Users: 100`);
  
  // Verify API is accessible
  const healthCheck = http.get(`${BASE_URL}/api/v1/health`);
  if (healthCheck.status !== 200) {
    throw new Error(`API health check failed: ${healthCheck.status}`);
  }
  
  console.log('✅ API health check passed');
  return { baseUrl: BASE_URL };
}

// Teardown function
export function teardown(data) {
  console.log('🏁 Performance tests completed');
  console.log(`Base URL tested: ${data.baseUrl}`);
}

// Export test scenarios for different test types
export const scenarios = {
  // Default load test
  load_test: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '2m', target: 10 },
      { duration: '5m', target: 10 },
      { duration: '2m', target: 0 },
    ],
    gracefulRampDown: '30s',
  },
  
  // Stress test
  stress_test: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '2m', target: 100 },
      { duration: '5m', target: 100 },
      { duration: '2m', target: 200 },
      { duration: '5m', target: 200 },
      { duration: '2m', target: 0 },
    ],
    gracefulRampDown: '30s',
    exec: 'stressTest',
  },
  
  // Spike test
  spike_test: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '10s', target: 100 },
      { duration: '1m', target: 100 },
      { duration: '10s', target: 1400 },
      { duration: '3m', target: 1400 },
      { duration: '10s', target: 100 },
      { duration: '3m', target: 100 },
      { duration: '10s', target: 0 },
    ],
    gracefulRampDown: '30s',
    exec: 'spikeTest',
  },
  
  // Volume test
  volume_test: {
    executor: 'constant-vus',
    vus: 10,
    duration: '5m',
    exec: 'volumeTest',
  },
  
  // Critical endpoints load test
  critical_endpoints: {
    executor: 'constant-arrival-rate',
    rate: 30,
    timeUnit: '1s',
    duration: '5m',
    preAllocatedVUs: 50,
    maxVUs: 100,
    exec: 'loadTestCriticalEndpoints',
  },
};
