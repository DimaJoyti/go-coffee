// Go Coffee Performance Testing Suite with k6
// This script tests the optimized Go Coffee service performance

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const orderCreationTime = new Trend('order_creation_duration');
const orderRetrievalTime = new Trend('order_retrieval_duration');
const cacheHitRate = new Rate('cache_hits');
const totalRequests = new Counter('total_requests');

// Test configuration
export const options = {
  scenarios: {
    // Smoke test - basic functionality
    smoke_test: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
      tags: { test_type: 'smoke' },
    },
    
    // Load test - normal expected load
    load_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 20 },  // Ramp up
        { duration: '5m', target: 20 },  // Stay at 20 users
        { duration: '2m', target: 0 },   // Ramp down
      ],
      tags: { test_type: 'load' },
    },
    
    // Stress test - beyond normal capacity
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },   // Ramp up to stress level
        { duration: '5m', target: 50 },   // Stay at stress level
        { duration: '2m', target: 100 },  // Ramp up to breaking point
        { duration: '3m', target: 100 },  // Stay at breaking point
        { duration: '2m', target: 0 },    // Ramp down
      ],
      tags: { test_type: 'stress' },
    },
    
    // Spike test - sudden traffic spikes
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 20 },   // Normal load
        { duration: '30s', target: 200 }, // Spike!
        { duration: '1m', target: 20 },   // Back to normal
        { duration: '30s', target: 0 },   // Ramp down
      ],
      tags: { test_type: 'spike' },
    },
  },
  
  thresholds: {
    // Performance requirements
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.01'],   // Error rate under 1%
    
    // Custom thresholds
    'order_creation_duration': ['p(95)<200'],
    'order_retrieval_duration': ['p(95)<100'],
    'cache_hits': ['rate>0.8'], // 80% cache hit rate
    'errors': ['rate<0.01'],
  },
};

const BASE_URL = 'http://localhost:8080';

// Test data generators
function generateOrderId() {
  return `order-${Math.random().toString(36).substr(2, 9)}`;
}

function generateCustomerId() {
  return `customer-${Math.random().toString(36).substr(2, 9)}`;
}

function generateOrder() {
  const items = [
    { product_id: 'coffee-1', name: 'Espresso', quantity: Math.floor(Math.random() * 3) + 1, price: 4.50 },
    { product_id: 'coffee-2', name: 'Latte', quantity: Math.floor(Math.random() * 2) + 1, price: 5.00 },
    { product_id: 'coffee-3', name: 'Cappuccino', quantity: Math.floor(Math.random() * 2) + 1, price: 4.75 },
  ];
  
  const selectedItems = items.slice(0, Math.floor(Math.random() * 3) + 1);
  const total = selectedItems.reduce((sum, item) => sum + (item.quantity * item.price), 0);
  
  return {
    id: generateOrderId(),
    customer_id: generateCustomerId(),
    items: selectedItems,
    total: Math.round(total * 100) / 100,
  };
}

// Store created orders for retrieval tests
let createdOrders = [];

export default function() {
  totalRequests.add(1);
  
  // Test scenario selection based on iteration
  const iteration = __ITER % 10;
  
  if (iteration < 3) {
    // 30% - Create new orders
    testOrderCreation();
  } else if (iteration < 7) {
    // 40% - Retrieve existing orders (test cache)
    testOrderRetrieval();
  } else if (iteration < 8) {
    // 10% - Health check
    testHealthEndpoint();
  } else if (iteration < 9) {
    // 10% - Metrics check
    testMetricsEndpoint();
  } else {
    // 10% - Home page
    testHomeEndpoint();
  }
  
  // Random sleep between 1-3 seconds to simulate real user behavior
  sleep(Math.random() * 2 + 1);
}

function testOrderCreation() {
  const order = generateOrder();
  
  const response = http.post(`${BASE_URL}/orders`, JSON.stringify(order), {
    headers: { 'Content-Type': 'application/json' },
    tags: { endpoint: 'create_order' },
  });
  
  const success = check(response, {
    'order creation status is 200': (r) => r.status === 200,
    'order creation response has id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id === order.id;
      } catch (e) {
        return false;
      }
    },
    'order creation response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  if (success) {
    createdOrders.push(order.id);
    orderCreationTime.add(response.timings.duration);
  } else {
    errorRate.add(1);
  }
}

function testOrderRetrieval() {
  // Use existing order or create a fallback
  const orderId = createdOrders.length > 0 
    ? createdOrders[Math.floor(Math.random() * createdOrders.length)]
    : 'order-123'; // Fallback to test order
  
  const response = http.get(`${BASE_URL}/orders/get?id=${orderId}`, {
    tags: { endpoint: 'get_order' },
  });
  
  const success = check(response, {
    'order retrieval status is 200': (r) => r.status === 200,
    'order retrieval has valid response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id && body.customer_id;
      } catch (e) {
        return false;
      }
    },
    'order retrieval response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  if (success) {
    orderRetrievalTime.add(response.timings.duration);
    // Assume cache hit if response time is very fast
    if (response.timings.duration < 50) {
      cacheHitRate.add(1);
    } else {
      cacheHitRate.add(0);
    }
  } else {
    errorRate.add(1);
  }
}

function testHealthEndpoint() {
  const response = http.get(`${BASE_URL}/health`, {
    tags: { endpoint: 'health' },
  });
  
  const success = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response is valid': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'healthy';
      } catch (e) {
        return false;
      }
    },
  });
  
  if (!success) {
    errorRate.add(1);
  }
}

function testMetricsEndpoint() {
  const response = http.get(`${BASE_URL}/metrics`, {
    tags: { endpoint: 'metrics' },
  });
  
  const success = check(response, {
    'metrics status is 200': (r) => r.status === 200,
    'metrics response is valid': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.timestamp && body.service;
      } catch (e) {
        return false;
      }
    },
  });
  
  if (!success) {
    errorRate.add(1);
  }
}

function testHomeEndpoint() {
  const response = http.get(`${BASE_URL}/`, {
    tags: { endpoint: 'home' },
  });
  
  const success = check(response, {
    'home status is 200': (r) => r.status === 200,
    'home response is valid': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.service && body.endpoints;
      } catch (e) {
        return false;
      }
    },
  });
  
  if (!success) {
    errorRate.add(1);
  }
}

// Setup function - runs once before the test
export function setup() {
  console.log('🚀 Starting Go Coffee Performance Tests');
  
  // Verify service is running
  const response = http.get(`${BASE_URL}/health`);
  if (response.status !== 200) {
    throw new Error('Service is not running or not healthy');
  }
  
  console.log('✅ Service health check passed');
  return { serviceReady: true };
}

// Teardown function - runs once after the test
export function teardown(data) {
  console.log('📊 Performance test completed');
  console.log(`📈 Total requests made: ${totalRequests.count}`);
  console.log(`⚡ Created orders: ${createdOrders.length}`);
}

// Handle summary - custom summary output
export function handleSummary(data) {
  const summary = {
    'performance-test-results.json': JSON.stringify(data, null, 2),
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
  
  // Add custom performance analysis
  const httpReqDuration = data.metrics.http_req_duration;
  const errorRate = data.metrics.errors ? data.metrics.errors.rate : 0;
  
  console.log('\n🎯 Performance Analysis:');
  console.log(`   Average Response Time: ${httpReqDuration.avg.toFixed(2)}ms`);
  console.log(`   95th Percentile: ${httpReqDuration['p(95)'].toFixed(2)}ms`);
  console.log(`   Error Rate: ${(errorRate * 100).toFixed(2)}%`);
  
  if (httpReqDuration['p(95)'] < 500) {
    console.log('✅ Performance target met (p95 < 500ms)');
  } else {
    console.log('❌ Performance target missed (p95 >= 500ms)');
  }
  
  if (errorRate < 0.01) {
    console.log('✅ Reliability target met (error rate < 1%)');
  } else {
    console.log('❌ Reliability target missed (error rate >= 1%)');
  }
  
  return summary;
}

function textSummary(data, options) {
  // Simple text summary implementation
  return `
Go Coffee Performance Test Summary
==================================

Test Duration: ${data.state.testRunDurationMs}ms
Total Requests: ${data.metrics.http_reqs ? data.metrics.http_reqs.count : 0}
Failed Requests: ${data.metrics.http_req_failed ? data.metrics.http_req_failed.count : 0}

Response Times:
- Average: ${data.metrics.http_req_duration ? data.metrics.http_req_duration.avg.toFixed(2) : 0}ms
- 95th Percentile: ${data.metrics.http_req_duration ? data.metrics.http_req_duration['p(95)'].toFixed(2) : 0}ms

Custom Metrics:
- Order Creation (p95): ${data.metrics.order_creation_duration ? data.metrics.order_creation_duration['p(95)'].toFixed(2) : 0}ms
- Order Retrieval (p95): ${data.metrics.order_retrieval_duration ? data.metrics.order_retrieval_duration['p(95)'].toFixed(2) : 0}ms
- Cache Hit Rate: ${data.metrics.cache_hits ? (data.metrics.cache_hits.rate * 100).toFixed(1) : 0}%
`;
}
