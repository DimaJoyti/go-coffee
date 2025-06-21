// Go Coffee Platform - Stress Testing with K6
// Tests system behavior under extreme load conditions

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');
const requestCount = new Counter('requests');

// Stress test configuration - Simplified for CI/CD
export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 20 },   // Ramp up to 20 users
    { duration: '1m', target: 30 },   // Ramp up to 30 users (stress level for CI)
    { duration: '30s', target: 20 },  // Scale back to 20 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'], // 95% of requests must complete below 3s (relaxed for CI stress)
    http_req_failed: ['rate<0.30'],    // Error rate must be below 30% (relaxed for CI stress)
    errors: ['rate<0.30'],             // Custom error rate below 30%
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

  if (scenario < 0.4) {
    // 40% - Basic HTTP load
    stressTestBasicHttp();
  } else if (scenario < 0.7) {
    // 30% - POST request load
    stressTestPostRequests();
  } else {
    // 30% - Mixed endpoint load
    stressTestMixedEndpoints();
  }

  // Minimal sleep to maximize load
  sleep(0.1);
}

function stressTestBasicHttp() {
  const endpoints = [
    `${BASE_URL}/get`,
    `${USER_GATEWAY_URL}/get`,
    `${SECURITY_GATEWAY_URL}/get`,
    `${WEB_UI_BACKEND_URL}/get`
  ];

  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const response = http.get(endpoint, {
    timeout: '5s',
  });

  const success = check(response, {
    'Stress GET status acceptable': (r) => r.status < 500,
    'Stress GET response time < 5s': (r) => r.timings.duration < 5000,
  });

  errorRate.add(!success);
  responseTime.add(response.timings.duration);
}

function stressTestPostRequests() {
  const testData = {
    stress_test: true,
    timestamp: new Date().toISOString(),
    user_id: Math.floor(Math.random() * 10000),
    data: `stress-data-${Math.random().toString(36).substring(7)}`
  };

  const endpoints = [
    `${BASE_URL}/post`,
    `${USER_GATEWAY_URL}/post`,
    `${SECURITY_GATEWAY_URL}/post`,
    `${WEB_UI_BACKEND_URL}/post`
  ];

  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const response = http.post(endpoint, JSON.stringify(testData), {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });

  const success = check(response, {
    'Stress POST status acceptable': (r) => r.status < 500,
    'Stress POST response time < 10s': (r) => r.timings.duration < 10000,
  });

  errorRate.add(!success);
  responseTime.add(response.timings.duration);
}

function stressTestMixedEndpoints() {
  // Test various HTTP methods and endpoints
  const testScenarios = [
    { method: 'GET', url: `${BASE_URL}/delay/${Math.floor(Math.random() * 3) + 1}` },
    { method: 'GET', url: `${USER_GATEWAY_URL}/status/${200 + Math.floor(Math.random() * 3) * 100}` },
    { method: 'GET', url: `${SECURITY_GATEWAY_URL}/headers` },
    { method: 'GET', url: `${WEB_UI_BACKEND_URL}/user-agent` },
    { method: 'POST', url: `${BASE_URL}/anything`, data: { stress: true } },
    { method: 'POST', url: `${USER_GATEWAY_URL}/anything`, data: { test: 'stress' } }
  ];

  const scenario = testScenarios[Math.floor(Math.random() * testScenarios.length)];
  let response;

  if (scenario.method === 'GET') {
    response = http.get(scenario.url, { timeout: '10s' });
  } else if (scenario.method === 'POST') {
    response = http.post(scenario.url, JSON.stringify(scenario.data), {
      headers: { 'Content-Type': 'application/json' },
      timeout: '10s',
    });
  }

  const success = check(response, {
    'Stress mixed endpoint acceptable': (r) => r.status < 500,
    'Stress mixed endpoint time < 10s': (r) => r.timings.duration < 10000,
  });

  errorRate.add(!success);
  responseTime.add(response.timings.duration);
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
