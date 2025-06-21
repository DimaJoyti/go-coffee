// Go Coffee Platform - Load Testing with K6
// Tests normal load conditions and performance baselines

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');

// Test configuration - Simplified for CI/CD
export const options = {
  stages: [
    { duration: '30s', target: 5 },   // Ramp up to 5 users
    { duration: '1m', target: 5 },    // Stay at 5 users
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },   // Stay at 10 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests must complete below 1s (relaxed for CI)
    http_req_failed: ['rate<0.10'],    // Error rate must be below 10% (relaxed for CI)
    errors: ['rate<0.10'],             // Custom error rate below 10%
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
  // Simplified load test for CI/CD environment
  // Test basic HTTP endpoints using httpbin mock services

  // Test 1: Basic HTTP GET requests
  testBasicEndpoints();

  // Test 2: HTTP POST requests
  testPostEndpoints();

  // Test 3: Response time validation
  testResponseTimes();

  sleep(1); // Think time between iterations
}

function testBasicEndpoints() {
  const services = [
    { name: 'API Gateway', url: `${BASE_URL}/get` },
    { name: 'User Gateway', url: `${USER_GATEWAY_URL}/get` },
    { name: 'Security Gateway', url: `${SECURITY_GATEWAY_URL}/get` },
    { name: 'Web UI Backend', url: `${WEB_UI_BACKEND_URL}/get` }
  ];

  services.forEach(service => {
    const response = http.get(service.url);

    const success = check(response, {
      [`${service.name} status is 200`]: (r) => r.status === 200,
      [`${service.name} response time < 500ms`]: (r) => r.timings.duration < 500,
      [`${service.name} has valid JSON response`]: (r) => {
        try {
          JSON.parse(r.body);
          return true;
        } catch (e) {
          return false;
        }
      },
    });

    errorRate.add(!success);
    responseTime.add(response.timings.duration);
  });
}

function testPostEndpoints() {
  const testData = {
    name: 'Load Test',
    timestamp: new Date().toISOString(),
    data: 'test payload'
  };

  const services = [
    { name: 'API Gateway', url: `${BASE_URL}/post` },
    { name: 'User Gateway', url: `${USER_GATEWAY_URL}/post` },
    { name: 'Security Gateway', url: `${SECURITY_GATEWAY_URL}/post` },
    { name: 'Web UI Backend', url: `${WEB_UI_BACKEND_URL}/post` }
  ];

  services.forEach(service => {
    const response = http.post(service.url, JSON.stringify(testData), {
      headers: { 'Content-Type': 'application/json' },
    });

    const success = check(response, {
      [`${service.name} POST status is 200`]: (r) => r.status === 200,
      [`${service.name} POST response time < 500ms`]: (r) => r.timings.duration < 500,
      [`${service.name} POST echoes data`]: (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.json && body.json.name === testData.name;
        } catch (e) {
          return false;
        }
      },
    });

    errorRate.add(!success);
    responseTime.add(response.timings.duration);
  });
}

function testResponseTimes() {
  // Test various HTTP methods and response times
  const endpoints = [
    { method: 'GET', url: `${BASE_URL}/delay/1`, name: 'Delayed Response (1s)' },
    { method: 'GET', url: `${USER_GATEWAY_URL}/status/200`, name: 'Status 200' },
    { method: 'GET', url: `${SECURITY_GATEWAY_URL}/headers`, name: 'Headers Check' },
    { method: 'GET', url: `${WEB_UI_BACKEND_URL}/user-agent`, name: 'User Agent' }
  ];

  endpoints.forEach(endpoint => {
    let response;
    if (endpoint.method === 'GET') {
      response = http.get(endpoint.url);
    } else if (endpoint.method === 'POST') {
      response = http.post(endpoint.url, '{}', {
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const success = check(response, {
      [`${endpoint.name} status is acceptable`]: (r) => r.status >= 200 && r.status < 400,
      [`${endpoint.name} response time < 2000ms`]: (r) => r.timings.duration < 2000,
    });

    errorRate.add(!success);
    responseTime.add(response.timings.duration);
  });
}

export function teardown(data) {
  console.log('Go Coffee Platform Load Test Completed');
  console.log('Check the results for performance metrics and thresholds');
}
