'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { toast } from '@/hooks/use-toast'

export interface WebSocketMessage {
  type: string
  data: any
  timestamp: string
}

export interface UseWebSocketReturn {
  isConnected: boolean
  lastMessage: WebSocketMessage | null
  sendMessage: (message: any) => void
  connectionState: 'connecting' | 'connected' | 'disconnected' | 'error'
}

export function useWebSocket(url?: string): UseWebSocketReturn {
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)
  const [connectionState, setConnectionState] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected')
  
  const ws = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5
  const reconnectDelay = 3000

  const wsUrl = url || process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8090/ws/realtime'

  const connect = useCallback(() => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      return
    }

    setConnectionState('connecting')
    
    try {
      ws.current = new WebSocket(wsUrl)

      ws.current.onopen = () => {
        console.log('WebSocket connected')
        setIsConnected(true)
        setConnectionState('connected')
        reconnectAttempts.current = 0
        
        // Send initial connection message
        ws.current?.send(JSON.stringify({
          type: 'connection',
          data: { clientId: generateClientId() },
          timestamp: new Date().toISOString()
        }))

        toast({
          title: "Connected",
          description: "Real-time connection established",
          variant: "success",
        })
      }

      ws.current.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          setLastMessage(message)
          
          // Handle different message types
          handleMessage(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      ws.current.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason)
        setIsConnected(false)
        setConnectionState('disconnected')
        
        // Attempt to reconnect if not a manual close
        if (event.code !== 1000 && reconnectAttempts.current < maxReconnectAttempts) {
          reconnectAttempts.current++
          console.log(`Attempting to reconnect (${reconnectAttempts.current}/${maxReconnectAttempts})...`)
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, reconnectDelay * reconnectAttempts.current)
        } else if (reconnectAttempts.current >= maxReconnectAttempts) {
          setConnectionState('error')
          toast({
            title: "Connection Failed",
            description: "Unable to establish real-time connection after multiple attempts",
            variant: "destructive",
          })
        }
      }

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error)
        setConnectionState('error')
      }

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      setConnectionState('error')
    }
  }, [wsUrl])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    
    if (ws.current) {
      ws.current.close(1000, 'Manual disconnect')
      ws.current = null
    }
    
    setIsConnected(false)
    setConnectionState('disconnected')
  }, [])

  const sendMessage = useCallback((message: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      const wsMessage: WebSocketMessage = {
        type: message.type || 'message',
        data: message,
        timestamp: new Date().toISOString()
      }
      
      ws.current.send(JSON.stringify(wsMessage))
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message)
      toast({
        title: "Connection Error",
        description: "Cannot send message - not connected to server",
        variant: "destructive",
      })
    }
  }, [])

  const handleMessage = (message: WebSocketMessage) => {
    switch (message.type) {
      case 'order_update':
        // Handle coffee order updates
        break
      case 'defi_update':
        // Handle DeFi portfolio updates
        break
      case 'agent_status':
        // Handle AI agent status updates
        break
      case 'market_data':
        // Handle market data updates
        break
      case 'notification':
        // Handle notifications
        toast({
          title: message.data.title || "Notification",
          description: message.data.message,
          variant: message.data.type || "default",
        })
        break
      default:
        console.log('Unhandled message type:', message.type)
    }
  }

  useEffect(() => {
    connect()

    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
    }
  }, [])

  return {
    isConnected,
    lastMessage,
    sendMessage,
    connectionState,
  }
}

function generateClientId(): string {
  return `client_${Math.random().toString(36).substring(2)}_${Date.now()}`
}
