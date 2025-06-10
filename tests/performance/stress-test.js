// Go Coffee Platform - Stress Testing with K6
// Tests system behavior under extreme load conditions

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');
const requestCount = new Counter('requests');

// Stress test configuration
export const options = {
  stages: [
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '3m', target: 200 },  // Ramp up to 200 users (stress level)
    { duration: '5m', target: 300 },  // Ramp up to 300 users (breaking point)
    { duration: '2m', target: 200 },  // Scale back to 200 users
    { duration: '2m', target: 100 },  // Scale back to 100 users
    { duration: '1m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2s (relaxed for stress)
    http_req_failed: ['rate<0.20'],    // Error rate must be below 20% (relaxed for stress)
    errors: ['rate<0.20'],             // Custom error rate below 20%
  },
};

// Base URLs
const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';
const USER_GATEWAY_URL = __ENV.USER_GATEWAY_URL || 'http://localhost:8081';
const SECURITY_GATEWAY_URL = __ENV.SECURITY_GATEWAY_URL || 'http://localhost:8082';
const WEB_UI_BACKEND_URL = __ENV.WEB_UI_BACKEND_URL || 'http://localhost:8090';

export function setup() {
  console.log('Starting Go Coffee Platform Stress Test');
  console.log('WARNING: This test will push the system to its limits');
  console.log(`Target: ${BASE_URL}`);
  
  return {};
}

export default function() {
  requestCount.add(1);
  
  // Randomly select test scenario to simulate real user behavior
  const scenario = Math.random();
  
  if (scenario < 0.3) {
    // 30% - Heavy authentication load
    stressTestAuthentication();
  } else if (scenario < 0.5) {
    // 20% - Coffee order processing
    stressTestCoffeeOrders();
  } else if (scenario < 0.7) {
    // 20% - DeFi operations
    stressTestDefiOperations();
  } else if (scenario < 0.9) {
    // 20% - Data scraping load
    stressTestDataScraping();
  } else {
    // 10% - Analytics queries
    stressTestAnalytics();
  }
  
  // Minimal sleep to maximize load
  sleep(0.1);
}

function stressTestAuthentication() {
  const userId = Math.floor(Math.random() * 10000);
  const testUser = {
    email: `stress-user-${userId}@example.com`,
    password: 'StressTest123!',
    name: `Stress User ${userId}`
  };
  
  // Rapid registration attempts
  const registerPayload = JSON.stringify(testUser);
  const registerResponse = http.post(`${USER_GATEWAY_URL}/api/v1/auth/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '5s',
  });
  
  const registerSuccess = check(registerResponse, {
    'Stress register status acceptable': (r) => r.status < 500,
    'Stress register response time < 5s': (r) => r.timings.duration < 5000,
  });
  
  errorRate.add(!registerSuccess);
  responseTime.add(registerResponse.timings.duration);
  
  // Rapid login attempts
  const loginPayload = JSON.stringify({
    email: testUser.email,
    password: testUser.password
  });
  
  const loginResponse = http.post(`${USER_GATEWAY_URL}/api/v1/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '5s',
  });
  
  const loginSuccess = check(loginResponse, {
    'Stress login status acceptable': (r) => r.status < 500,
    'Stress login response time < 5s': (r) => r.timings.duration < 5000,
  });
  
  errorRate.add(!loginSuccess);
  responseTime.add(loginResponse.timings.duration);
}

function stressTestCoffeeOrders() {
  // Rapid order creation
  const orderPayload = JSON.stringify({
    items: [
      { id: 1, name: 'Espresso', quantity: Math.floor(Math.random() * 5) + 1 },
      { id: 2, name: 'Latte', quantity: Math.floor(Math.random() * 3) + 1 }
    ],
    customerName: `Stress Customer ${Math.floor(Math.random() * 1000)}`,
    paymentMethod: 'crypto'
  });
  
  const createOrderResponse = http.post(`${WEB_UI_BACKEND_URL}/api/v1/coffee/orders`, orderPayload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });
  
  const createSuccess = check(createOrderResponse, {
    'Stress order creation acceptable': (r) => r.status < 500,
    'Stress order creation time < 10s': (r) => r.timings.duration < 10000,
  });
  
  errorRate.add(!createSuccess);
  responseTime.add(createOrderResponse.timings.duration);
  
  // Rapid inventory checks
  const inventoryResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/coffee/inventory`, {
    timeout: '5s',
  });
  
  const inventorySuccess = check(inventoryResponse, {
    'Stress inventory check acceptable': (r) => r.status < 500,
    'Stress inventory check time < 5s': (r) => r.timings.duration < 5000,
  });
  
  errorRate.add(!inventorySuccess);
  responseTime.add(inventoryResponse.timings.duration);
}

function stressTestDefiOperations() {
  // Rapid portfolio queries
  const portfolioResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/defi/portfolio`, {
    timeout: '10s',
  });
  
  const portfolioSuccess = check(portfolioResponse, {
    'Stress portfolio query acceptable': (r) => r.status < 500,
    'Stress portfolio query time < 10s': (r) => r.timings.duration < 10000,
  });
  
  errorRate.add(!portfolioSuccess);
  responseTime.add(portfolioResponse.timings.duration);
  
  // Rapid strategy toggles
  const strategyId = Math.floor(Math.random() * 10) + 1;
  const toggleResponse = http.post(`${WEB_UI_BACKEND_URL}/api/v1/defi/strategies/${strategyId}/toggle`, null, {
    timeout: '5s',
  });
  
  const toggleSuccess = check(toggleResponse, {
    'Stress strategy toggle acceptable': (r) => r.status < 500,
    'Stress strategy toggle time < 5s': (r) => r.timings.duration < 5000,
  });
  
  errorRate.add(!toggleSuccess);
  responseTime.add(toggleResponse.timings.duration);
}

function stressTestDataScraping() {
  // Rapid data refresh requests
  const refreshResponse = http.post(`${WEB_UI_BACKEND_URL}/api/v1/scraping/refresh`, null, {
    timeout: '15s',
  });
  
  const refreshSuccess = check(refreshResponse, {
    'Stress data refresh acceptable': (r) => r.status < 500,
    'Stress data refresh time < 15s': (r) => r.timings.duration < 15000,
  });
  
  errorRate.add(!refreshSuccess);
  responseTime.add(refreshResponse.timings.duration);
  
  // Rapid URL scraping requests
  const scrapePayload = JSON.stringify({
    url: 'https://example.com',
    format: 'json'
  });
  
  const scrapeResponse = http.post(`${WEB_UI_BACKEND_URL}/api/v1/scraping/url`, scrapePayload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '20s',
  });
  
  const scrapeSuccess = check(scrapeResponse, {
    'Stress URL scraping acceptable': (r) => r.status < 500,
    'Stress URL scraping time < 20s': (r) => r.timings.duration < 20000,
  });
  
  errorRate.add(!scrapeSuccess);
  responseTime.add(scrapeResponse.timings.duration);
}

function stressTestAnalytics() {
  // Rapid analytics queries
  const analyticsEndpoints = [
    '/api/v1/analytics/sales',
    '/api/v1/analytics/revenue',
    '/api/v1/analytics/products',
    '/api/v1/analytics/locations'
  ];
  
  const endpoint = analyticsEndpoints[Math.floor(Math.random() * analyticsEndpoints.length)];
  const analyticsResponse = http.get(`${WEB_UI_BACKEND_URL}${endpoint}`, {
    timeout: '10s',
  });
  
  const analyticsSuccess = check(analyticsResponse, {
    'Stress analytics query acceptable': (r) => r.status < 500,
    'Stress analytics query time < 10s': (r) => r.timings.duration < 10000,
  });
  
  errorRate.add(!analyticsSuccess);
  responseTime.add(analyticsResponse.timings.duration);
}

export function teardown(data) {
  console.log('Go Coffee Platform Stress Test Completed');
  console.log('Review the results to identify system breaking points and bottlenecks');
  console.log('Key metrics to analyze:');
  console.log('- Response time degradation under load');
  console.log('- Error rate increase patterns');
  console.log('- Resource utilization peaks');
  console.log('- Recovery time after load reduction');
}
