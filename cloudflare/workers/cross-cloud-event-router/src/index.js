/**
 * Go Coffee Platform - Cross-Cloud Event Router Worker
 * JavaScript implementation for Cloudflare Workers
 */

// Event routing configuration
const EVENT_ROUTING_TABLE = {
  'coffee.order.created': {
    aws_target: 'order-processor',
    gcp_target: 'order-handler',
    azure_target: 'order-function',
    priority: 1,
    retry_policy: 'exponential_backoff'
  },
  'coffee.payment.processed': {
    aws_target: 'payment-processor',
    gcp_target: 'payment-handler',
    azure_target: 'payment-function',
    priority: 1,
    retry_policy: 'linear_backoff'
  },
  'coffee.inventory.updated': {
    aws_target: 'inventory-processor',
    gcp_target: 'inventory-handler',
    azure_target: 'inventory-function',
    priority: 2,
    retry_policy: 'exponential_backoff'
  },
  'coffee.user.registered': {
    aws_target: 'user-processor',
    gcp_target: 'user-handler',
    azure_target: 'user-function',
    priority: 3,
    retry_policy: 'linear_backoff'
  }
};

// Cross-Cloud Event Router
class CrossCloudEventRouter {
  constructor(env) {
    this.env = env;
    this.cache = env.CACHE;
    this.eventLogs = env.EVENT_LOGS;
    this.routingTable = EVENT_ROUTING_TABLE;
    
    // Cloud provider configurations
    this.cloudProviders = {
      aws: {
        name: 'aws',
        enabled: env.ENABLE_AWS_ROUTING === 'true',
        region: env.AWS_REGION || 'us-east-1'
      },
      gcp: {
        name: 'gcp',
        enabled: env.ENABLE_GCP_ROUTING === 'true',
        region: env.GCP_REGION || 'us-central1'
      },
      azure: {
        name: 'azure',
        enabled: env.ENABLE_AZURE_ROUTING === 'true',
        region: env.AZURE_LOCATION || 'East US'
      }
    };
  }

  async routeEvent(event) {
    const startTime = Date.now();
    const correlationId = event.correlation_id || crypto.randomUUID();
    
    try {
      console.log(`Routing event: ${event.event_type} with correlation ID: ${correlationId}`);
      
      // Get routing configuration for this event type
      const route = this.routingTable[event.event_type];
      if (!route) {
        throw new Error(`No routing configuration found for event type: ${event.event_type}`);
      }
      
      // Route to enabled cloud providers
      const routingResults = await Promise.allSettled([
        this.cloudProviders.aws.enabled ? this.routeToAWS(event, route, correlationId) : null,
        this.cloudProviders.gcp.enabled ? this.routeToGCP(event, route, correlationId) : null,
        this.cloudProviders.azure.enabled ? this.routeToAzure(event, route, correlationId) : null
      ].filter(Boolean));
      
      // Process results
      const results = {
        correlation_id: correlationId,
        event_type: event.event_type,
        routing_results: [],
        success_count: 0,
        failure_count: 0,
        processing_time_ms: Date.now() - startTime
      };
      
      routingResults.forEach((result, index) => {
        const provider = ['aws', 'gcp', 'azure'][index];
        if (result.status === 'fulfilled') {
          results.routing_results.push({
            provider,
            status: 'success',
            result: result.value
          });
          results.success_count++;
        } else {
          results.routing_results.push({
            provider,
            status: 'failed',
            error: result.reason.message
          });
          results.failure_count++;
        }
      });
      
      // Store routing result in cache
      await this.cache.put(
        `routing_result:${correlationId}`,
        JSON.stringify(results),
        { expirationTtl: 3600 } // 1 hour
      );
      
      // Log event for analytics
      await this.logEvent(event, results);
      
      return results;
      
    } catch (error) {
      console.error(`Event routing failed: ${error.message}`);
      
      const errorResult = {
        correlation_id: correlationId,
        event_type: event.event_type,
        error: error.message,
        processing_time_ms: Date.now() - startTime
      };
      
      await this.logEvent(event, errorResult);
      
      return errorResult;
    }
  }

  async routeToAWS(event, route, correlationId) {
    console.log(`Routing to AWS: ${route.aws_target}`);
    
    // Simulate AWS EventBridge routing
    const eventEntry = {
      source: 'go-coffee.cross-cloud',
      detail_type: event.event_type,
      detail: JSON.stringify({
        correlation_id: correlationId,
        trace_id: event.trace_id,
        payload: event.payload
      }),
      event_bus_name: `go-coffee-${this.env.ENVIRONMENT}-event-bus`,
      resources: [
        `arn:aws:lambda:${this.cloudProviders.aws.region}:*:function:go-coffee-${this.env.ENVIRONMENT}-${route.aws_target}`
      ]
    };
    
    // Simulate successful routing (in real implementation, would call AWS EventBridge API)
    await this.simulateDelay(50, 150); // Simulate network latency
    
    return {
      provider: 'aws',
      target: route.aws_target,
      event_id: crypto.randomUUID(),
      status: 'routed',
      timestamp: new Date().toISOString()
    };
  }

  async routeToGCP(event, route, correlationId) {
    console.log(`Routing to GCP: ${route.gcp_target}`);
    
    // Simulate Google Pub/Sub routing
    const pubsubMessage = {
      data: btoa(JSON.stringify({
        correlation_id: correlationId,
        trace_id: event.trace_id,
        payload: event.payload
      })),
      attributes: {
        event_type: event.event_type,
        source: 'go-coffee.cross-cloud',
        target: route.gcp_target
      }
    };
    
    // Simulate successful routing
    await this.simulateDelay(40, 120);
    
    return {
      provider: 'gcp',
      target: route.gcp_target,
      message_id: crypto.randomUUID(),
      status: 'published',
      timestamp: new Date().toISOString()
    };
  }

  async routeToAzure(event, route, correlationId) {
    console.log(`Routing to Azure: ${route.azure_target}`);
    
    // Simulate Azure Event Grid routing
    const eventGridEvent = {
      id: crypto.randomUUID(),
      eventType: event.event_type,
      subject: `go-coffee/${route.azure_target}`,
      eventTime: new Date().toISOString(),
      data: {
        correlation_id: correlationId,
        trace_id: event.trace_id,
        payload: event.payload
      },
      dataVersion: '1.0'
    };
    
    // Simulate successful routing
    await this.simulateDelay(60, 180);
    
    return {
      provider: 'azure',
      target: route.azure_target,
      event_id: eventGridEvent.id,
      status: 'dispatched',
      timestamp: new Date().toISOString()
    };
  }

  async logEvent(event, result) {
    const logEntry = {
      timestamp: new Date().toISOString(),
      event_type: event.event_type,
      correlation_id: result.correlation_id,
      source: event.source || 'unknown',
      result: result,
      environment: this.env.ENVIRONMENT
    };
    
    // Store in cache for recent events
    const recentEventsKey = 'recent_events';
    const recentEvents = await this.cache.get(recentEventsKey);
    const events = recentEvents ? JSON.parse(recentEvents) : [];
    
    events.unshift(logEntry);
    events.splice(100); // Keep only last 100 events
    
    await this.cache.put(recentEventsKey, JSON.stringify(events), { expirationTtl: 86400 }); // 24 hours
  }

  async getEventHistory(limit = 50) {
    const recentEvents = await this.cache.get('recent_events');
    const events = recentEvents ? JSON.parse(recentEvents) : [];
    return events.slice(0, limit);
  }

  async getRoutingStats() {
    const events = await this.getEventHistory(100);
    
    const stats = {
      total_events: events.length,
      success_rate: 0,
      provider_stats: {
        aws: { total: 0, success: 0 },
        gcp: { total: 0, success: 0 },
        azure: { total: 0, success: 0 }
      },
      event_type_stats: {}
    };
    
    events.forEach(event => {
      // Count by event type
      if (!stats.event_type_stats[event.event_type]) {
        stats.event_type_stats[event.event_type] = { total: 0, success: 0 };
      }
      stats.event_type_stats[event.event_type].total++;
      
      // Count by provider
      if (event.result.routing_results) {
        event.result.routing_results.forEach(result => {
          if (stats.provider_stats[result.provider]) {
            stats.provider_stats[result.provider].total++;
            if (result.status === 'success') {
              stats.provider_stats[result.provider].success++;
              stats.event_type_stats[event.event_type].success++;
            }
          }
        });
      }
    });
    
    // Calculate success rate
    const totalSuccesses = Object.values(stats.provider_stats).reduce((sum, stat) => sum + stat.success, 0);
    const totalAttempts = Object.values(stats.provider_stats).reduce((sum, stat) => sum + stat.total, 0);
    stats.success_rate = totalAttempts > 0 ? (totalSuccesses / totalAttempts) : 0;
    
    return stats;
  }

  async simulateDelay(min, max) {
    const delay = Math.floor(Math.random() * (max - min + 1)) + min;
    return new Promise(resolve => setTimeout(resolve, delay));
  }
}

// Main Worker Handler
export default {
  async fetch(request, env, ctx) {
    try {
      const url = new URL(request.url);
      
      // Health check endpoint
      if (url.pathname === '/health') {
        return new Response(JSON.stringify({
          status: 'healthy',
          service: 'cross-cloud-event-router',
          timestamp: new Date().toISOString(),
          version: '1.0.0',
          cloud_providers: {
            aws: env.ENABLE_AWS_ROUTING === 'true',
            gcp: env.ENABLE_GCP_ROUTING === 'true',
            azure: env.ENABLE_AZURE_ROUTING === 'true'
          }
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
      
      // Route event endpoint
      if (request.method === 'POST' && url.pathname === '/route') {
        const router = new CrossCloudEventRouter(env);
        const event = await request.json();
        
        const result = await router.routeEvent(event);
        
        return new Response(JSON.stringify({
          success: true,
          data: result,
          meta: {
            timestamp: new Date().toISOString(),
            request_id: crypto.randomUUID()
          }
        }), {
          headers: { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'POST, OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type, Authorization'
          }
        });
      }
      
      // Get event history
      if (request.method === 'GET' && url.pathname === '/events') {
        const router = new CrossCloudEventRouter(env);
        const limit = parseInt(url.searchParams.get('limit')) || 50;
        const events = await router.getEventHistory(limit);
        
        return new Response(JSON.stringify({
          success: true,
          data: events,
          meta: {
            timestamp: new Date().toISOString(),
            total: events.length
          }
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
      
      // Get routing statistics
      if (request.method === 'GET' && url.pathname === '/stats') {
        const router = new CrossCloudEventRouter(env);
        const stats = await router.getRoutingStats();
        
        return new Response(JSON.stringify({
          success: true,
          data: stats,
          meta: {
            timestamp: new Date().toISOString()
          }
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
      
      // Handle CORS preflight
      if (request.method === 'OPTIONS') {
        return new Response(null, {
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type, Authorization'
          }
        });
      }
      
      return new Response('Not Found', { status: 404 });
      
    } catch (error) {
      console.error('Worker error:', error);
      
      return new Response(JSON.stringify({
        success: false,
        error: {
          code: 'INTERNAL_ERROR',
          message: error.message
        },
        meta: {
          timestamp: new Date().toISOString(),
          request_id: crypto.randomUUID()
        }
      }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }
};
