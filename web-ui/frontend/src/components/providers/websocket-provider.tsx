'use client'

import React, { createContext, useContext, ReactNode } from 'react'
import { useWebSocket, UseWebSocketReturn } from '@/hooks/use-websocket'

const WebSocketContext = createContext<UseWebSocketReturn | null>(null)

interface WebSocketProviderProps {
  children: ReactNode
  url?: string
}

export function WebSocketProvider({ children, url }: WebSocketProviderProps) {
  const websocket = useWebSocket(url)

  return (
    <WebSocketContext.Provider value={websocket}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocketContext() {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider')
  }
  return context
}
