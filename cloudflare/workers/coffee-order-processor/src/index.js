/**
 * Go Coffee Platform - Coffee Order Processor Worker
 * JavaScript implementation for Cloudflare Workers
 */

// Order status constants
const ORDER_STATUS = {
  PENDING: 'pending',
  CONFIRMED: 'confirmed',
  PREPARING: 'preparing',
  READY: 'ready',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled'
};

// Payment status constants
const PAYMENT_STATUS = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  COMPLETED: 'completed',
  FAILED: 'failed',
  REFUNDED: 'refunded'
};

// Coffee Order Processor
class CoffeeOrderProcessor {
  constructor(env) {
    this.env = env;
    this.cache = env.CACHE;
    this.sessions = env.SESSIONS;
    this.orders = env.ORDERS;
    this.assets = env.ASSETS;
    this.orderReceipts = env.ORDER_RECEIPTS;
  }

  async processOrder(orderData) {
    const startTime = Date.now();
    const orderId = orderData.order_id || crypto.randomUUID();
    
    try {
      console.log(`Processing order: ${orderId}`);
      
      // Validate order data
      const validationResult = await this.validateOrder(orderData);
      if (!validationResult.valid) {
        throw new Error(`Order validation failed: ${validationResult.errors.join(', ')}`);
      }
      
      // Create order record
      const order = {
        id: orderId,
        customer_id: orderData.customer_id,
        items: orderData.items,
        subtotal: this.calculateSubtotal(orderData.items),
        tax: 0,
        total: 0,
        status: ORDER_STATUS.PENDING,
        payment_status: PAYMENT_STATUS.PENDING,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        estimated_ready_time: this.calculateReadyTime(orderData.items),
        special_instructions: orderData.special_instructions || '',
        delivery_method: orderData.delivery_method || 'pickup',
        location_id: orderData.location_id
      };
      
      // Calculate tax and total
      order.tax = this.calculateTax(order.subtotal);
      order.total = order.subtotal + order.tax;
      
      // Store order
      await this.orders.put(`order:${orderId}`, JSON.stringify(order));
      
      // Process payment if enabled
      let paymentResult = null;
      if (this.env.PAYMENT_PROCESSING_ENABLED === 'true') {
        paymentResult = await this.processPayment(order, orderData.payment_method);
        order.payment_status = paymentResult.status;
        order.payment_id = paymentResult.payment_id;
      }
      
      // Update inventory if enabled
      if (this.env.INVENTORY_TRACKING_ENABLED === 'true') {
        await this.updateInventory(order.items);
      }
      
      // Send notifications if enabled
      if (this.env.NOTIFICATION_ENABLED === 'true') {
        await this.sendOrderNotification(order);
      }
      
      // Update order status
      order.status = ORDER_STATUS.CONFIRMED;
      order.updated_at = new Date().toISOString();
      await this.orders.put(`order:${orderId}`, JSON.stringify(order));
      
      const processingTime = Date.now() - startTime;
      
      return {
        success: true,
        order: order,
        payment: paymentResult,
        processing_time_ms: processingTime
      };
      
    } catch (error) {
      console.error(`Order processing failed: ${error.message}`);
      
      return {
        success: false,
        order_id: orderId,
        error: error.message,
        processing_time_ms: Date.now() - startTime
      };
    }
  }

  async validateOrder(orderData) {
    const errors = [];
    
    // Check required fields
    if (!orderData.customer_id) errors.push('Customer ID is required');
    if (!orderData.items || !Array.isArray(orderData.items) || orderData.items.length === 0) {
      errors.push('Order must contain at least one item');
    }
    
    // Validate items
    if (orderData.items) {
      orderData.items.forEach((item, index) => {
        if (!item.coffee_id) errors.push(`Item ${index + 1}: Coffee ID is required`);
        if (!item.size) errors.push(`Item ${index + 1}: Size is required`);
        if (!item.quantity || item.quantity < 1) errors.push(`Item ${index + 1}: Valid quantity is required`);
      });
    }
    
    // Check order limits
    const maxOrderValue = parseInt(this.env.MAX_ORDER_VALUE) || 10000; // $100 default
    const maxItemsPerOrder = parseInt(this.env.MAX_ITEMS_PER_ORDER) || 50;
    
    if (orderData.items && orderData.items.length > maxItemsPerOrder) {
      errors.push(`Order exceeds maximum items limit (${maxItemsPerOrder})`);
    }
    
    const subtotal = this.calculateSubtotal(orderData.items || []);
    if (subtotal > maxOrderValue) {
      errors.push(`Order exceeds maximum value limit ($${maxOrderValue / 100})`);
    }
    
    return {
      valid: errors.length === 0,
      errors: errors
    };
  }

  calculateSubtotal(items) {
    return items.reduce((total, item) => {
      const basePrice = this.getItemPrice(item.coffee_id, item.size);
      const customizationPrice = this.getCustomizationPrice(item.customizations || []);
      return total + ((basePrice + customizationPrice) * item.quantity);
    }, 0);
  }

  getItemPrice(coffeeId, size) {
    // Simplified pricing - in real implementation, would fetch from database
    const prices = {
      americano: { small: 250, medium: 300, large: 350 }, // cents
      latte: { small: 400, medium: 450, large: 500 },
      cappuccino: { small: 375, medium: 425, large: 475 },
      espresso: { small: 200, medium: 250, large: 300 },
      macchiato: { small: 425, medium: 475, large: 525 }
    };
    
    return prices[coffeeId]?.[size] || 300; // Default price
  }

  getCustomizationPrice(customizations) {
    // Simplified customization pricing
    const customizationPrices = {
      'extra_shot': 75,
      'decaf': 0,
      'oat_milk': 50,
      'almond_milk': 50,
      'extra_hot': 0,
      'extra_foam': 0,
      'vanilla_syrup': 50,
      'caramel_syrup': 50
    };
    
    return customizations.reduce((total, customization) => {
      return total + (customizationPrices[customization.name] || 0) * (customization.quantity || 1);
    }, 0);
  }

  calculateTax(subtotal) {
    const taxRate = 0.08; // 8% tax rate
    return Math.round(subtotal * taxRate);
  }

  calculateReadyTime(items) {
    // Estimate preparation time based on items
    const baseTime = 5; // 5 minutes base time
    const timePerItem = 2; // 2 minutes per item
    const totalMinutes = baseTime + (items.length * timePerItem);
    
    const readyTime = new Date();
    readyTime.setMinutes(readyTime.getMinutes() + totalMinutes);
    
    return readyTime.toISOString();
  }

  async processPayment(order, paymentMethod) {
    console.log(`Processing payment for order: ${order.id}`);
    
    // Simulate payment processing
    await this.simulateDelay(500, 1500);
    
    // In real implementation, would integrate with Stripe or other payment processor
    const paymentResult = {
      payment_id: crypto.randomUUID(),
      status: PAYMENT_STATUS.COMPLETED,
      amount: order.total,
      currency: 'USD',
      payment_method: paymentMethod?.type || 'card',
      processed_at: new Date().toISOString()
    };
    
    // Store payment record
    await this.cache.put(
      `payment:${paymentResult.payment_id}`,
      JSON.stringify(paymentResult),
      { expirationTtl: 86400 } // 24 hours
    );
    
    return paymentResult;
  }

  async updateInventory(items) {
    console.log('Updating inventory for items:', items.length);
    
    // Simulate inventory update
    for (const item of items) {
      const inventoryKey = `inventory:${item.coffee_id}`;
      const currentInventory = await this.cache.get(inventoryKey);
      const inventory = currentInventory ? JSON.parse(currentInventory) : { quantity: 100 };
      
      inventory.quantity = Math.max(0, inventory.quantity - item.quantity);
      inventory.updated_at = new Date().toISOString();
      
      await this.cache.put(inventoryKey, JSON.stringify(inventory), { expirationTtl: 86400 });
    }
  }

  async sendOrderNotification(order) {
    console.log(`Sending notification for order: ${order.id}`);
    
    // Simulate notification sending
    const notification = {
      type: 'order_confirmation',
      order_id: order.id,
      customer_id: order.customer_id,
      message: `Your order #${order.id} has been confirmed and will be ready at ${new Date(order.estimated_ready_time).toLocaleTimeString()}`,
      sent_at: new Date().toISOString()
    };
    
    // Store notification record
    await this.cache.put(
      `notification:${order.id}`,
      JSON.stringify(notification),
      { expirationTtl: 86400 }
    );
    
    return notification;
  }

  async getOrder(orderId) {
    const orderData = await this.orders.get(`order:${orderId}`);
    return orderData ? JSON.parse(orderData) : null;
  }

  async updateOrderStatus(orderId, status) {
    const order = await this.getOrder(orderId);
    if (!order) {
      throw new Error('Order not found');
    }
    
    order.status = status;
    order.updated_at = new Date().toISOString();
    
    if (status === ORDER_STATUS.COMPLETED) {
      order.completed_at = new Date().toISOString();
    }
    
    await this.orders.put(`order:${orderId}`, JSON.stringify(order));
    
    // Send status update notification
    if (this.env.NOTIFICATION_ENABLED === 'true') {
      await this.sendStatusNotification(order);
    }
    
    return order;
  }

  async sendStatusNotification(order) {
    const statusMessages = {
      [ORDER_STATUS.PREPARING]: 'Your order is being prepared',
      [ORDER_STATUS.READY]: 'Your order is ready for pickup!',
      [ORDER_STATUS.COMPLETED]: 'Thank you for your order!'
    };
    
    const notification = {
      type: 'status_update',
      order_id: order.id,
      customer_id: order.customer_id,
      status: order.status,
      message: statusMessages[order.status] || `Order status updated to ${order.status}`,
      sent_at: new Date().toISOString()
    };
    
    await this.cache.put(
      `notification:${order.id}:${order.status}`,
      JSON.stringify(notification),
      { expirationTtl: 86400 }
    );
    
    return notification;
  }

  async getOrderHistory(customerId, limit = 20) {
    // In real implementation, would query database with proper indexing
    // For now, simulate with cached data
    const historyKey = `order_history:${customerId}`;
    const historyData = await this.cache.get(historyKey);
    const history = historyData ? JSON.parse(historyData) : [];
    
    return history.slice(0, limit);
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
          service: 'coffee-order-processor',
          timestamp: new Date().toISOString(),
          version: '1.0.0',
          features: {
            order_processing: env.ORDER_PROCESSING_ENABLED === 'true',
            payment_processing: env.PAYMENT_PROCESSING_ENABLED === 'true',
            inventory_tracking: env.INVENTORY_TRACKING_ENABLED === 'true',
            notifications: env.NOTIFICATION_ENABLED === 'true'
          }
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
      
      // Process order endpoint
      if (request.method === 'POST' && url.pathname === '/orders') {
        const processor = new CoffeeOrderProcessor(env);
        const orderData = await request.json();
        
        const result = await processor.processOrder(orderData);
        
        return new Response(JSON.stringify({
          success: result.success,
          data: result,
          meta: {
            timestamp: new Date().toISOString(),
            request_id: crypto.randomUUID()
          }
        }), {
          status: result.success ? 200 : 400,
          headers: { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'POST, GET, PUT, OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type, Authorization'
          }
        });
      }
      
      // Get order endpoint
      if (request.method === 'GET' && url.pathname.startsWith('/orders/')) {
        const orderId = url.pathname.split('/')[2];
        const processor = new CoffeeOrderProcessor(env);
        
        const order = await processor.getOrder(orderId);
        
        if (!order) {
          return new Response(JSON.stringify({
            success: false,
            error: { code: 'ORDER_NOT_FOUND', message: 'Order not found' }
          }), {
            status: 404,
            headers: { 'Content-Type': 'application/json' }
          });
        }
        
        return new Response(JSON.stringify({
          success: true,
          data: order,
          meta: { timestamp: new Date().toISOString() }
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
      
      // Update order status endpoint
      if (request.method === 'PUT' && url.pathname.startsWith('/orders/') && url.pathname.endsWith('/status')) {
        const orderId = url.pathname.split('/')[2];
        const processor = new CoffeeOrderProcessor(env);
        const { status } = await request.json();
        
        const order = await processor.updateOrderStatus(orderId, status);
        
        return new Response(JSON.stringify({
          success: true,
          data: order,
          meta: { timestamp: new Date().toISOString() }
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
      
      // Handle CORS preflight
      if (request.method === 'OPTIONS') {
        return new Response(null, {
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'GET, POST, PUT, OPTIONS',
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
