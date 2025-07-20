'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { io, Socket } from 'socket.io-client'
import type { WebSocketMessage } from '@/types/trading'

interface UseWebSocketOptions {
  url?: string
  userId?: string
  autoConnect?: boolean
  reconnectAttempts?: number
  reconnectDelay?: number
}

interface UseWebSocketReturn {
  socket: Socket | null
  isConnected: boolean
  isConnecting: boolean
  error: string | null
  connect: () => void
  disconnect: () => void
  subscribe: (channels: string[]) => void
  unsubscribe: (channels: string[]) => void
  sendMessage: (message: any) => void
}

export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const {
    url = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8090',
    userId = 'default-user',
    autoConnect = true,
    reconnectAttempts = 5,
    reconnectDelay = 1000,
  } = options

  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  const socketRef = useRef<Socket | null>(null)
  const reconnectCountRef = useRef(0)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const connect = useCallback(() => {
    if (socketRef.current?.connected) {
      return
    }

    setIsConnecting(true)
    setError(null)

    try {
      const wsUrl = url.replace('http://', 'ws://').replace('https://', 'wss://')
      
      socketRef.current = io(wsUrl, {
        transports: ['websocket'],
        query: {
          user_id: userId,
        },
        forceNew: true,
      })

      socketRef.current.on('connect', () => {
        console.log('WebSocket connected')
        setIsConnected(true)
        setIsConnecting(false)
        setError(null)
        reconnectCountRef.current = 0
      })

      socketRef.current.on('disconnect', (reason) => {
        console.log('WebSocket disconnected:', reason)
        setIsConnected(false)
        setIsConnecting(false)
        
        // Auto-reconnect logic
        if (reason === 'io server disconnect') {
          // Server initiated disconnect, don't reconnect
          return
        }
        
        if (reconnectCountRef.current < reconnectAttempts) {
          reconnectCountRef.current++
          const delay = reconnectDelay * Math.pow(2, reconnectCountRef.current - 1)
          
          console.log(`Attempting to reconnect in ${delay}ms (attempt ${reconnectCountRef.current}/${reconnectAttempts})`)
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, delay)
        } else {
          setError('Failed to reconnect after maximum attempts')
        }
      })

      socketRef.current.on('connect_error', (err) => {
        console.error('WebSocket connection error:', err)
        setError(`Connection error: ${err.message}`)
        setIsConnecting(false)
      })

      socketRef.current.on('error', (err) => {
        console.error('WebSocket error:', err)
        setError(`WebSocket error: ${err}`)
      })

    } catch (err) {
      console.error('Failed to create WebSocket connection:', err)
      setError(`Failed to connect: ${err}`)
      setIsConnecting(false)
    }
  }, [url, userId, reconnectAttempts, reconnectDelay])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    
    if (socketRef.current) {
      socketRef.current.disconnect()
      socketRef.current = null
    }
    
    setIsConnected(false)
    setIsConnecting(false)
    reconnectCountRef.current = 0
  }, [])

  const subscribe = useCallback((channels: string[]) => {
    if (!socketRef.current?.connected) {
      console.warn('Cannot subscribe: WebSocket not connected')
      return
    }

    const message = {
      type: 'subscribe',
      data: { channels }
    }

    socketRef.current.emit('message', message)
    console.log('Subscribed to channels:', channels)
  }, [])

  const unsubscribe = useCallback((channels: string[]) => {
    if (!socketRef.current?.connected) {
      console.warn('Cannot unsubscribe: WebSocket not connected')
      return
    }

    const message = {
      type: 'unsubscribe',
      data: { channels }
    }

    socketRef.current.emit('message', message)
    console.log('Unsubscribed from channels:', channels)
  }, [])

  const sendMessage = useCallback((message: any) => {
    if (!socketRef.current?.connected) {
      console.warn('Cannot send message: WebSocket not connected')
      return
    }

    socketRef.current.emit('message', message)
  }, [])

  // Auto-connect on mount
  useEffect(() => {
    if (autoConnect) {
      connect()
    }

    return () => {
      disconnect()
    }
  }, [autoConnect, connect, disconnect])

  return {
    socket: socketRef.current,
    isConnected,
    isConnecting,
    error,
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    sendMessage,
  }
}

// Hook for listening to specific WebSocket events
export function useWebSocketEvent<T = any>(
  socket: Socket | null,
  eventName: string,
  handler: (data: T) => void,
  deps: React.DependencyList = []
) {
  useEffect(() => {
    if (!socket) return

    socket.on(eventName, handler)

    return () => {
      socket.off(eventName, handler)
    }
  }, [socket, eventName, ...deps])
}

// Hook for WebSocket messages with type filtering
export function useWebSocketMessages(
  socket: Socket | null,
  messageTypes: string[] = [],
  handler: (message: WebSocketMessage) => void
) {
  useWebSocketEvent(socket, 'message', (message: WebSocketMessage) => {
    if (messageTypes.length === 0 || messageTypes.includes(message.type)) {
      handler(message)
    }
  }, [messageTypes, handler])
}
