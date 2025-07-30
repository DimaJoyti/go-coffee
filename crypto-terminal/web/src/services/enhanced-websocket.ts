import { EventEmitter } from 'events';

export interface MarketDataUpdate {
  symbol: string;
  price: number;
  change24h: number;
  changePercent24h: number;
  volume24h: number;
  timestamp: number;
}

export interface OrderBookUpdate {
  symbol: string;
  bids: Array<[number, number]>; // [price, quantity]
  asks: Array<[number, number]>; // [price, quantity]
  lastUpdateId: number;
}

export interface TradeUpdate {
  symbol: string;
  price: number;
  quantity: number;
  time: number;
  isBuyerMaker: boolean;
}

export interface TickerUpdate {
  symbol: string;
  priceChange: number;
  priceChangePercent: number;
  weightedAvgPrice: number;
  prevClosePrice: number;
  lastPrice: number;
  lastQty: number;
  bidPrice: number;
  bidQty: number;
  askPrice: number;
  askQty: number;
  openPrice: number;
  highPrice: number;
  lowPrice: number;
  volume: number;
  quoteVolume: number;
  openTime: number;
  closeTime: number;
  count: number;
}

export type WebSocketMessage = 
  | { type: 'marketData'; data: MarketDataUpdate }
  | { type: 'orderBook'; data: OrderBookUpdate }
  | { type: 'trade'; data: TradeUpdate }
  | { type: 'ticker'; data: TickerUpdate }
  | { type: 'error'; data: { message: string } }
  | { type: 'connected'; data: { timestamp: number } }
  | { type: 'disconnected'; data: { reason: string } };

export class EnhancedWebSocketService extends EventEmitter {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnecting = false;
  private subscriptions = new Set<string>();
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private lastHeartbeat = 0;

  constructor(url: string = 'ws://localhost:8080/ws') {
    super();
    this.url = url;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      if (this.isConnecting) {
        this.once('connected', resolve);
        this.once('error', reject);
        return;
      }

      this.isConnecting = true;

      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('Enhanced WebSocket connected');
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.startHeartbeat();
          this.resubscribeAll();
          this.emit('connected', { timestamp: Date.now() });
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
            this.emit('error', { message: 'Failed to parse message' });
          }
        };

        this.ws.onclose = (event) => {
          console.log('Enhanced WebSocket disconnected:', event.reason);
          this.isConnecting = false;
          this.stopHeartbeat();
          this.emit('disconnected', { reason: event.reason });
          
          if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error('Enhanced WebSocket error:', error);
          this.isConnecting = false;
          this.emit('error', { message: 'WebSocket connection error' });
          reject(error);
        };

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.ws) {
      this.stopHeartbeat();
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
  }

  private handleMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case 'marketData':
        this.emit('marketData', message.data);
        break;
      case 'orderBook':
        this.emit('orderBook', message.data);
        break;
      case 'trade':
        this.emit('trade', message.data);
        break;
      case 'ticker':
        this.emit('ticker', message.data);
        break;
      case 'error':
        this.emit('error', message.data);
        break;
      default:
        console.warn('Unknown message type:', message);
    }
  }

  private scheduleReconnect(): void {
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`);
    
    setTimeout(() => {
      if (this.ws?.readyState !== WebSocket.OPEN) {
        this.connect().catch(console.error);
      }
    }, delay);
  }

  private startHeartbeat(): void {
    this.lastHeartbeat = Date.now();
    this.heartbeatInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping', timestamp: Date.now() });
        
        // Check if we haven't received a heartbeat response
        if (Date.now() - this.lastHeartbeat > 30000) {
          console.warn('Heartbeat timeout, reconnecting...');
          this.disconnect();
          this.connect().catch(console.error);
        }
      }
    }, 10000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  // Subscription methods
  subscribeToMarketData(symbols: string[]): void {
    symbols.forEach(symbol => {
      this.subscriptions.add(`marketData:${symbol}`);
    });
    
    this.send({
      type: 'subscribe',
      channel: 'marketData',
      symbols
    });
  }

  subscribeToOrderBook(symbol: string): void {
    this.subscriptions.add(`orderBook:${symbol}`);
    
    this.send({
      type: 'subscribe',
      channel: 'orderBook',
      symbol
    });
  }

  subscribeToTrades(symbol: string): void {
    this.subscriptions.add(`trades:${symbol}`);
    
    this.send({
      type: 'subscribe',
      channel: 'trades',
      symbol
    });
  }

  subscribeToTicker(symbols: string[]): void {
    symbols.forEach(symbol => {
      this.subscriptions.add(`ticker:${symbol}`);
    });
    
    this.send({
      type: 'subscribe',
      channel: 'ticker',
      symbols
    });
  }

  unsubscribeFromMarketData(symbols: string[]): void {
    symbols.forEach(symbol => {
      this.subscriptions.delete(`marketData:${symbol}`);
    });
    
    this.send({
      type: 'unsubscribe',
      channel: 'marketData',
      symbols
    });
  }

  unsubscribeFromOrderBook(symbol: string): void {
    this.subscriptions.delete(`orderBook:${symbol}`);
    
    this.send({
      type: 'unsubscribe',
      channel: 'orderBook',
      symbol
    });
  }

  unsubscribeFromTrades(symbol: string): void {
    this.subscriptions.delete(`trades:${symbol}`);
    
    this.send({
      type: 'unsubscribe',
      channel: 'trades',
      symbol
    });
  }

  unsubscribeFromTicker(symbols: string[]): void {
    symbols.forEach(symbol => {
      this.subscriptions.delete(`ticker:${symbol}`);
    });
    
    this.send({
      type: 'unsubscribe',
      channel: 'ticker',
      symbols
    });
  }

  private resubscribeAll(): void {
    // Group subscriptions by channel
    const channels: Record<string, string[]> = {};
    
    this.subscriptions.forEach(subscription => {
      const [channel, symbol] = subscription.split(':');
      if (!channels[channel]) {
        channels[channel] = [];
      }
      channels[channel].push(symbol);
    });

    // Resubscribe to all channels
    Object.entries(channels).forEach(([channel, symbols]) => {
      switch (channel) {
        case 'marketData':
          this.subscribeToMarketData(symbols);
          break;
        case 'orderBook':
          symbols.forEach(symbol => this.subscribeToOrderBook(symbol));
          break;
        case 'trades':
          symbols.forEach(symbol => this.subscribeToTrades(symbol));
          break;
        case 'ticker':
          this.subscribeToTicker(symbols);
          break;
      }
    });
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  getConnectionState(): string {
    if (!this.ws) return 'disconnected';
    
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'connecting';
      case WebSocket.OPEN:
        return 'connected';
      case WebSocket.CLOSING:
        return 'closing';
      case WebSocket.CLOSED:
        return 'closed';
      default:
        return 'unknown';
    }
  }
}

// Singleton instance
export const enhancedWebSocketService = new EnhancedWebSocketService();
