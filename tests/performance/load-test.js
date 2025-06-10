// Go Coffee Platform - Load Testing with K6
// Tests normal load conditions and performance baselines

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 },   // Ramp up to 10 users
    { duration: '5m', target: 10 },   // Stay at 10 users
    { duration: '2m', target: 20 },   // Ramp up to 20 users
    { duration: '5m', target: 20 },   // Stay at 20 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.05'],   // Error rate must be below 5%
    errors: ['rate<0.05'],            // Custom error rate below 5%
  },
};

// Base URLs
const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';
const USER_GATEWAY_URL = __ENV.USER_GATEWAY_URL || 'http://localhost:8081';
const SECURITY_GATEWAY_URL = __ENV.SECURITY_GATEWAY_URL || 'http://localhost:8082';
const WEB_UI_BACKEND_URL = __ENV.WEB_UI_BACKEND_URL || 'http://localhost:8090';

// Test data
const testUser = {
  email: `test-${Math.random().toString(36).substring(7)}@example.com`,
  password: 'TestPassword123!',
  name: 'Test User'
};

export function setup() {
  console.log('Starting Go Coffee Platform Load Test');
  console.log(`Base URL: ${BASE_URL}`);
  console.log(`User Gateway: ${USER_GATEWAY_URL}`);
  console.log(`Security Gateway: ${SECURITY_GATEWAY_URL}`);
  console.log(`Web UI Backend: ${WEB_UI_BACKEND_URL}`);
  
  return { testUser };
}

export default function(data) {
  // Test 1: Health Check Endpoints
  testHealthChecks();
  
  // Test 2: User Registration and Authentication
  testUserAuthentication(data.testUser);
  
  // Test 3: Coffee Order Management
  testCoffeeOrders();
  
  // Test 4: DeFi Operations
  testDefiOperations();
  
  // Test 5: Web Scraping Services
  testScrapingServices();
  
  // Test 6: Analytics Endpoints
  testAnalytics();
  
  sleep(1); // Think time between iterations
}

function testHealthChecks() {
  const services = [
    { name: 'User Gateway', url: `${USER_GATEWAY_URL}/health` },
    { name: 'Security Gateway', url: `${SECURITY_GATEWAY_URL}/health` },
    { name: 'Web UI Backend', url: `${WEB_UI_BACKEND_URL}/health` }
  ];
  
  services.forEach(service => {
    const response = http.get(service.url);
    
    const success = check(response, {
      [`${service.name} health check status is 200`]: (r) => r.status === 200,
      [`${service.name} health check response time < 100ms`]: (r) => r.timings.duration < 100,
    });
    
    errorRate.add(!success);
    responseTime.add(response.timings.duration);
  });
}

function testUserAuthentication(user) {
  // Register user
  const registerPayload = JSON.stringify({
    email: user.email,
    password: user.password,
    name: user.name
  });
  
  const registerResponse = http.post(`${USER_GATEWAY_URL}/api/v1/auth/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const registerSuccess = check(registerResponse, {
    'User registration status is 201 or 409': (r) => r.status === 201 || r.status === 409,
    'User registration response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  errorRate.add(!registerSuccess);
  responseTime.add(registerResponse.timings.duration);
  
  // Login user
  const loginPayload = JSON.stringify({
    email: user.email,
    password: user.password
  });
  
  const loginResponse = http.post(`${USER_GATEWAY_URL}/api/v1/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const loginSuccess = check(loginResponse, {
    'User login status is 200': (r) => r.status === 200,
    'User login response time < 200ms': (r) => r.timings.duration < 200,
    'Login response contains token': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.token !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  errorRate.add(!loginSuccess);
  responseTime.add(loginResponse.timings.duration);
  
  // Extract token for subsequent requests
  let token = '';
  if (loginResponse.status === 200) {
    try {
      const body = JSON.parse(loginResponse.body);
      token = body.token;
    } catch (e) {
      console.error('Failed to parse login response');
    }
  }
  
  return token;
}

function testCoffeeOrders() {
  // Get coffee inventory
  const inventoryResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/coffee/inventory`);
  
  const inventorySuccess = check(inventoryResponse, {
    'Coffee inventory status is 200': (r) => r.status === 200,
    'Coffee inventory response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!inventorySuccess);
  responseTime.add(inventoryResponse.timings.duration);
  
  // Get coffee orders
  const ordersResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/coffee/orders`);
  
  const ordersSuccess = check(ordersResponse, {
    'Coffee orders status is 200': (r) => r.status === 200,
    'Coffee orders response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!ordersSuccess);
  responseTime.add(ordersResponse.timings.duration);
}

function testDefiOperations() {
  // Get DeFi portfolio
  const portfolioResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/defi/portfolio`);
  
  const portfolioSuccess = check(portfolioResponse, {
    'DeFi portfolio status is 200': (r) => r.status === 200,
    'DeFi portfolio response time < 400ms': (r) => r.timings.duration < 400,
  });
  
  errorRate.add(!portfolioSuccess);
  responseTime.add(portfolioResponse.timings.duration);
  
  // Get DeFi strategies
  const strategiesResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/defi/strategies`);
  
  const strategiesSuccess = check(strategiesResponse, {
    'DeFi strategies status is 200': (r) => r.status === 200,
    'DeFi strategies response time < 400ms': (r) => r.timings.duration < 400,
  });
  
  errorRate.add(!strategiesSuccess);
  responseTime.add(strategiesResponse.timings.duration);
}

function testScrapingServices() {
  // Get market data
  const marketDataResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/scraping/data`);
  
  const marketDataSuccess = check(marketDataResponse, {
    'Market data status is 200': (r) => r.status === 200,
    'Market data response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  
  errorRate.add(!marketDataSuccess);
  responseTime.add(marketDataResponse.timings.duration);
  
  // Get data sources
  const sourcesResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/scraping/sources`);
  
  const sourcesSuccess = check(sourcesResponse, {
    'Data sources status is 200': (r) => r.status === 200,
    'Data sources response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  errorRate.add(!sourcesSuccess);
  responseTime.add(sourcesResponse.timings.duration);
}

function testAnalytics() {
  // Get sales data
  const salesResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/analytics/sales`);
  
  const salesSuccess = check(salesResponse, {
    'Sales analytics status is 200': (r) => r.status === 200,
    'Sales analytics response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  errorRate.add(!salesSuccess);
  responseTime.add(salesResponse.timings.duration);
  
  // Get revenue data
  const revenueResponse = http.get(`${WEB_UI_BACKEND_URL}/api/v1/analytics/revenue`);
  
  const revenueSuccess = check(revenueResponse, {
    'Revenue analytics status is 200': (r) => r.status === 200,
    'Revenue analytics response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  errorRate.add(!revenueSuccess);
  responseTime.add(revenueResponse.timings.duration);
}

export function teardown(data) {
  console.log('Go Coffee Platform Load Test Completed');
  console.log('Check the results for performance metrics and thresholds');
}
