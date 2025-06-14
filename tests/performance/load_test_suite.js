import http from 'k6/http';
import ws from 'k6/ws';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics
const errorRate = new Rate('error_rate');
const responseTime = new Trend('response_time');
const throughput = new Rate('throughput');
const activeConnections = new Gauge('active_connections');
const businessMetrics = {
  ordersCreated: new Counter('orders_created'),
  paymentsProcessed: new Counter('payments_processed'),
  cacheHits: new Rate('cache_hits'),
  dbQueries: new Counter('db_queries')
};

// Test configuration
export const options = {
  scenarios: {
    // Smoke test - basic functionality
    smoke_test: {
      executor: 'constant-vus',
      vus: 1,
      duration: '1m',
      tags: { test_type: 'smoke' },
      env: { TEST_TYPE: 'smoke' }
    },
    
    // Load test - normal expected load
    load_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 10 },   // Ramp up
        { duration: '5m', target: 10 },   // Stay at 10 users
        { duration: '2m', target: 20 },   // Ramp up to 20 users
        { duration: '5m', target: 20 },   // Stay at 20 users
        { duration: '2m', target: 0 },    // Ramp down
      ],
      tags: { test_type: 'load' },
      env: { TEST_TYPE: 'load' }
    },
    
    // Stress test - beyond normal capacity
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 20 },   // Ramp up to normal load
        { duration: '5m', target: 20 },   // Stay at normal load
        { duration: '2m', target: 50 },   // Ramp up to stress level
        { duration: '5m', target: 50 },   // Stay at stress level
        { duration: '2m', target: 100 },  // Ramp up to breaking point
        { duration: '5m', target: 100 },  // Stay at breaking point
        { duration: '2m', target: 0 },    // Ramp down
      ],
      tags: { test_type: 'stress' },
      env: { TEST_TYPE: 'stress' }
    },
    
    // Spike test - sudden traffic spikes
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 10 },   // Normal load
        { duration: '30s', target: 100 }, // Spike
        { duration: '1m', target: 10 },   // Back to normal
        { duration: '30s', target: 200 }, // Bigger spike
        { duration: '1m', target: 10 },   // Back to normal
        { duration: '1m', target: 0 },    // Ramp down
      ],
      tags: { test_type: 'spike' },
      env: { TEST_TYPE: 'spike' }
    },
    
    // Volume test - large amounts of data
    volume_test: {
      executor: 'constant-vus',
      vus: 5,
      duration: '10m',
      tags: { test_type: 'volume' },
      env: { TEST_TYPE: 'volume' }
    },
    
    // Soak test - extended duration
    soak_test: {
      executor: 'constant-vus',
      vus: 10,
      duration: '1h',
      tags: { test_type: 'soak' },
      env: { TEST_TYPE: 'soak' }
    }
  },
  
  thresholds: {
    // Response time thresholds
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    'http_req_duration{test_type:smoke}': ['p(95)<200'],
    'http_req_duration{test_type:load}': ['p(95)<300'],
    'http_req_duration{test_type:stress}': ['p(95)<1000'],
    
    // Error rate thresholds
    'http_req_failed': ['rate<0.01'], // Less than 1% errors
    'http_req_failed{test_type:smoke}': ['rate<0.001'], // Less than 0.1% errors
    'http_req_failed{test_type:load}': ['rate<0.005'], // Less than 0.5% errors
    
    // Throughput thresholds
    'http_reqs': ['rate>100'], // More than 100 requests per second
    'http_reqs{test_type:load}': ['rate>200'],
    'http_reqs{test_type:stress}': ['rate>500'],
    
    // Business metrics thresholds
    'orders_created': ['rate>10'],
    'payments_processed': ['rate>8'],
    'cache_hits': ['rate>0.8'], // 80% cache hit rate
    
    // System resource thresholds
    'error_rate': ['rate<0.01'],
    'response_time': ['p(95)<500']
  }
};

// Test data generators
const testData = {
  generateUser: () => ({
    id: randomString(10),
    name: `User_${randomString(5)}`,
    email: `user_${randomString(8)}@example.com`,
    phone: `+1${randomIntBetween(1000000000, 9999999999)}`
  }),
  
  generateOrder: () => ({
    id: randomString(12),
    customer_name: `Customer_${randomString(6)}`,
    items: [
      {
        product_id: `coffee_${randomIntBetween(1, 20)}`,
        quantity: randomIntBetween(1, 5),
        price: randomIntBetween(300, 800) / 100 // $3.00 - $8.00
      }
    ],
    shop_id: `shop_${randomIntBetween(1, 10)}`,
    payment_method: ['credit_card', 'crypto', 'cash'][randomIntBetween(0, 2)]
  }),
  
  generatePayment: (orderId, amount) => ({
    order_id: orderId,
    amount: amount,
    currency: ['USD', 'ETH', 'BTC'][randomIntBetween(0, 2)],
    payment_method: ['stripe', 'crypto_wallet', 'cash'][randomIntBetween(0, 2)]
  })
};

// Base URL configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_VERSION = __ENV.API_VERSION || 'v1';
const API_BASE = `${BASE_URL}/api/${API_VERSION}`;

// Authentication setup
let authToken = '';

export function setup() {
  console.log(`Starting performance tests against: ${BASE_URL}`);
  console.log(`Test type: ${__ENV.TEST_TYPE || 'default'}`);
  
  // Authenticate and get token
  const authResponse = http.post(`${API_BASE}/auth/login`, JSON.stringify({
    username: 'test_user',
    password: 'test_password'
  }), {
    headers: { 'Content-Type': 'application/json' }
  });
  
  if (authResponse.status === 200) {
    const authData = JSON.parse(authResponse.body);
    authToken = authData.token;
    console.log('Authentication successful');
  } else {
    console.warn('Authentication failed, proceeding without token');
  }
  
  return { authToken };
}

export default function(data) {
  const testType = __ENV.TEST_TYPE || 'default';
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': data.authToken ? `Bearer ${data.authToken}` : ''
  };
  
  group('Coffee Order Flow', () => {
    // 1. Health check
    group('Health Check', () => {
      const response = http.get(`${BASE_URL}/health`);
      check(response, {
        'health check status is 200': (r) => r.status === 200,
        'health check response time < 100ms': (r) => r.timings.duration < 100
      });
      responseTime.add(response.timings.duration);
      errorRate.add(response.status !== 200);
    });
    
    // 2. Create order
    let orderId = '';
    group('Create Order', () => {
      const orderData = testData.generateOrder();
      const response = http.post(`${API_BASE}/orders`, JSON.stringify(orderData), { headers });
      
      const success = check(response, {
        'order creation status is 201': (r) => r.status === 201,
        'order creation response time < 500ms': (r) => r.timings.duration < 500,
        'order has valid ID': (r) => {
          try {
            const body = JSON.parse(r.body);
            return body.order_id && body.order_id.length > 0;
          } catch (e) {
            return false;
          }
        }
      });
      
      if (success && response.status === 201) {
        const responseBody = JSON.parse(response.body);
        orderId = responseBody.order_id;
        businessMetrics.ordersCreated.add(1);
      }
      
      responseTime.add(response.timings.duration);
      errorRate.add(response.status !== 201);
    });
    
    // 3. Get order details
    if (orderId) {
      group('Get Order Details', () => {
        const response = http.get(`${API_BASE}/orders/${orderId}`, { headers });
        
        check(response, {
          'get order status is 200': (r) => r.status === 200,
          'get order response time < 200ms': (r) => r.timings.duration < 200,
          'order details are valid': (r) => {
            try {
              const body = JSON.parse(r.body);
              return body.id === orderId && body.status;
            } catch (e) {
              return false;
            }
          }
        });
        
        responseTime.add(response.timings.duration);
        errorRate.add(response.status !== 200);
      });
    }
    
    // 4. Process payment
    if (orderId) {
      group('Process Payment', () => {
        const paymentData = testData.generatePayment(orderId, 5.50);
        const response = http.post(`${API_BASE}/payments`, JSON.stringify(paymentData), { headers });
        
        const success = check(response, {
          'payment processing status is 200': (r) => r.status === 200,
          'payment processing response time < 1000ms': (r) => r.timings.duration < 1000,
          'payment is successful': (r) => {
            try {
              const body = JSON.parse(r.body);
              return body.status === 'success' || body.status === 'completed';
            } catch (e) {
              return false;
            }
          }
        });
        
        if (success) {
          businessMetrics.paymentsProcessed.add(1);
        }
        
        responseTime.add(response.timings.duration);
        errorRate.add(response.status !== 200);
      });
    }
    
    // 5. Cache performance test
    group('Cache Performance', () => {
      const cacheKey = `menu_${randomIntBetween(1, 5)}`; // Limited keys for cache hits
      const response = http.get(`${API_BASE}/menu/${cacheKey}`, { headers });
      
      check(response, {
        'cache request status is 200': (r) => r.status === 200,
        'cache request response time < 50ms': (r) => r.timings.duration < 50
      });
      
      // Check for cache hit header
      const cacheHit = response.headers['X-Cache-Status'] === 'HIT';
      businessMetrics.cacheHits.add(cacheHit ? 1 : 0);
      
      responseTime.add(response.timings.duration);
      errorRate.add(response.status !== 200);
    });
    
    // 6. Database query simulation
    group('Database Queries', () => {
      const response = http.get(`${API_BASE}/analytics/orders?limit=10`, { headers });
      
      check(response, {
        'analytics query status is 200': (r) => r.status === 200,
        'analytics query response time < 300ms': (r) => r.timings.duration < 300
      });
      
      businessMetrics.dbQueries.add(1);
      responseTime.add(response.timings.duration);
      errorRate.add(response.status !== 200);
    });
  });
  
  // WebSocket connection test (for real-time features)
  if (testType === 'load' || testType === 'stress') {
    group('WebSocket Connection', () => {
      const wsUrl = `ws://localhost:8080/ws/orders`;
      const response = ws.connect(wsUrl, {}, function (socket) {
        socket.on('open', () => {
          activeConnections.add(1);
          socket.send(JSON.stringify({ type: 'subscribe', channel: 'orders' }));
        });
        
        socket.on('message', (data) => {
          check(data, {
            'websocket message is valid': (msg) => {
              try {
                const parsed = JSON.parse(msg);
                return parsed.type && parsed.data;
              } catch (e) {
                return false;
              }
            }
          });
        });
        
        socket.on('close', () => {
          activeConnections.add(-1);
        });
        
        // Keep connection open for a short time
        sleep(randomIntBetween(1, 3));
      });
    });
  }
  
  // Volume test specific logic
  if (testType === 'volume') {
    group('Bulk Operations', () => {
      // Create multiple orders in batch
      const batchSize = 10;
      const orders = [];
      
      for (let i = 0; i < batchSize; i++) {
        orders.push(testData.generateOrder());
      }
      
      const response = http.post(`${API_BASE}/orders/batch`, JSON.stringify({ orders }), { headers });
      
      check(response, {
        'batch order creation status is 201': (r) => r.status === 201,
        'batch order creation response time < 2000ms': (r) => r.timings.duration < 2000,
        'all orders created successfully': (r) => {
          try {
            const body = JSON.parse(r.body);
            return body.created_count === batchSize;
          } catch (e) {
            return false;
          }
        }
      });
      
      if (response.status === 201) {
        businessMetrics.ordersCreated.add(batchSize);
      }
      
      responseTime.add(response.timings.duration);
      errorRate.add(response.status !== 201);
    });
  }
  
  // Random sleep to simulate user think time
  sleep(randomIntBetween(1, 3));
}

export function teardown(data) {
  console.log('Performance test completed');
  
  // Cleanup operations if needed
  if (data.authToken) {
    http.post(`${API_BASE}/auth/logout`, {}, {
      headers: { 'Authorization': `Bearer ${data.authToken}` }
    });
  }
}

// Custom check functions
export function checkResponseTime(response, threshold) {
  return response.timings.duration < threshold;
}

export function checkErrorRate(responses, maxErrorRate) {
  const errors = responses.filter(r => r.status >= 400).length;
  return (errors / responses.length) <= maxErrorRate;
}

// Performance test scenarios for different endpoints
export const scenarios = {
  // API Gateway performance
  apiGateway: () => {
    group('API Gateway Load', () => {
      const endpoints = [
        '/health',
        '/metrics',
        '/api/v1/status',
        '/api/v1/version'
      ];
      
      endpoints.forEach(endpoint => {
        const response = http.get(`${BASE_URL}${endpoint}`);
        check(response, {
          [`${endpoint} status is 200`]: (r) => r.status === 200,
          [`${endpoint} response time < 100ms`]: (r) => r.timings.duration < 100
        });
      });
    });
  },
  
  // Database performance
  database: () => {
    group('Database Performance', () => {
      const queries = [
        '/api/v1/orders?limit=100',
        '/api/v1/users?limit=50',
        '/api/v1/analytics/revenue',
        '/api/v1/analytics/popular-items'
      ];
      
      queries.forEach(query => {
        const response = http.get(`${BASE_URL}${query}`, {
          headers: { 'Authorization': `Bearer ${authToken}` }
        });
        check(response, {
          [`${query} status is 200`]: (r) => r.status === 200,
          [`${query} response time < 500ms`]: (r) => r.timings.duration < 500
        });
        
        businessMetrics.dbQueries.add(1);
      });
    });
  },
  
  // Cache performance
  cache: () => {
    group('Cache Performance', () => {
      const cacheableEndpoints = [
        '/api/v1/menu',
        '/api/v1/shops',
        '/api/v1/products',
        '/api/v1/categories'
      ];
      
      cacheableEndpoints.forEach(endpoint => {
        const response = http.get(`${BASE_URL}${endpoint}`);
        const cacheHit = response.headers['X-Cache-Status'] === 'HIT';
        
        check(response, {
          [`${endpoint} status is 200`]: (r) => r.status === 200,
          [`${endpoint} response time < 100ms`]: (r) => r.timings.duration < 100
        });
        
        businessMetrics.cacheHits.add(cacheHit ? 1 : 0);
      });
    });
  }
};
