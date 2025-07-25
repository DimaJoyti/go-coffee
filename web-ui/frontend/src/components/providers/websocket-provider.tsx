// Simplified WebSocket provider - no dependencies required

interface WebSocketProviderProps {
  children: any
  url?: string
}

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  // Simple wrapper that just returns children
  return children
}

export function useWebSocketContext() {
  // Return mock websocket context
  return {
    socket: null,
    isConnected: false,
    connect: () => {},
    disconnect: () => {},
    send: () => {},
  }
}
