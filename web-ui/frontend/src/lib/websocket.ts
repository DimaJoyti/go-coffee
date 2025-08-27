// WebSocket connection manager for real-time data
export class WebSocketManager {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private subscribers: Map<string, Set<(data: any) => void>> = new Map()
  private isConnecting = false

  constructor(private url: string) {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      if (this.isConnecting) {
        return
      }

      this.isConnecting = true

      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.isConnecting = false
          this.reconnectAttempts = 0
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data)
            this.handleMessage(data)
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error)
          }
        }

        this.ws.onclose = () => {
          console.log('WebSocket disconnected')
          this.isConnecting = false
          this.attemptReconnect()
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error)
          this.isConnecting = false
          reject(error)
        }
      } catch (error) {
        this.isConnecting = false
        reject(error)
      }
    })
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)

    setTimeout(() => {
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
      this.connect().catch(() => {
        // Reconnection failed, will try again
      })
    }, delay)
  }

  private handleMessage(data: any) {
    const { type, payload } = data
    const subscribers = this.subscribers.get(type)
    
    if (subscribers) {
      subscribers.forEach(callback => {
        try {
          callback(payload)
        } catch (error) {
          console.error('Error in WebSocket subscriber:', error)
        }
      })
    }
  }

  subscribe(type: string, callback: (data: any) => void) {
    if (!this.subscribers.has(type)) {
      this.subscribers.set(type, new Set())
    }
    this.subscribers.get(type)!.add(callback)

    // Send subscription message to server
    this.send({
      type: 'subscribe',
      channel: type
    })

    // Return unsubscribe function
    return () => {
      const subscribers = this.subscribers.get(type)
      if (subscribers) {
        subscribers.delete(callback)
        if (subscribers.size === 0) {
          this.subscribers.delete(type)
          // Send unsubscribe message to server
          this.send({
            type: 'unsubscribe',
            channel: type
          })
        }
      }
    }
  }

  send(data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    } else {
      console.warn('WebSocket not connected, message not sent:', data)
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.subscribers.clear()
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// Singleton instance
let wsManager: WebSocketManager | null = null

export function getWebSocketManager(): WebSocketManager {
  if (!wsManager) {
    // Use mock WebSocket server for development
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws'
    wsManager = new WebSocketManager(wsUrl)
  }
  return wsManager
}

// Mock WebSocket server for development
export class MockWebSocketServer {
  private intervals: NodeJS.Timeout[] = []
  private subscribers: Set<(data: any) => void> = new Set()

  start() {
    console.log('🚀 Starting Go Coffee Mock WebSocket Server...')

    // Mock price updates
    const priceInterval = setInterval(() => {
      const prices = {
        'BTC/USDT': 45000 + (Math.random() - 0.5) * 2000,
        'ETH/USDT': 2800 + (Math.random() - 0.5) * 200,
        'SOL/USDT': 98 + (Math.random() - 0.5) * 10,
        'COFFEE/USDT': 12 + (Math.random() - 0.5) * 2
      }

      console.log('📊 Broadcasting price updates:', Object.keys(prices))

      Object.entries(prices).forEach(([symbol, price]) => {
        this.broadcast({
          type: 'price_update',
          payload: {
            symbol,
            price: Number(price.toFixed(2)),
            change24h: (Math.random() - 0.5) * 10,
            volume: Math.random() * 1000000,
            timestamp: Date.now()
          }
        })
      })
    }, 2000)

    // Mock trade updates
    const tradeInterval = setInterval(() => {
      const symbols = ['BTC/USDT', 'ETH/USDT', 'SOL/USDT', 'COFFEE/USDT']
      const symbol = symbols[Math.floor(Math.random() * symbols.length)]
      
      this.broadcast({
        type: 'trade_update',
        payload: {
          symbol,
          price: 45000 + (Math.random() - 0.5) * 2000,
          amount: Math.random() * 2,
          side: Math.random() > 0.5 ? 'buy' : 'sell',
          timestamp: Date.now()
        }
      })
    }, 1000)

    // Mock AI agent updates
    const aiInterval = setInterval(() => {
      this.broadcast({
        type: 'ai_update',
        payload: {
          agentId: 'trading-alpha',
          status: 'active',
          profit: (Math.random() - 0.5) * 100,
          action: 'Executed trade order',
          timestamp: Date.now()
        }
      })
    }, 5000)

    this.intervals.push(priceInterval, tradeInterval, aiInterval)
  }

  subscribe(callback: (data: any) => void) {
    this.subscribers.add(callback)
    return () => this.subscribers.delete(callback)
  }

  private broadcast(data: any) {
    this.subscribers.forEach(callback => {
      try {
        callback(data)
      } catch (error) {
        console.error('Error in mock WebSocket subscriber:', error)
      }
    })
  }

  stop() {
    this.intervals.forEach(interval => clearInterval(interval))
    this.intervals = []
    this.subscribers.clear()
  }
}

// Global mock server instance
let mockServer: MockWebSocketServer | null = null

export function getMockWebSocketServer(): MockWebSocketServer {
  if (!mockServer) {
    mockServer = new MockWebSocketServer()
  }
  return mockServer
}
