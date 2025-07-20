'use client'

import { useEffect } from 'react'
import { useWebSocket, useWebSocketMessages } from '@/hooks/use-websocket'
import { useTradingStore } from '@/stores/trading-store'
import type {
  WebSocketMessage,
  PriceUpdate,
  SignalAlert,
  TradeExecution,
  PortfolioUpdate,
  RiskAlert,
} from '@/types/trading'
import toast from 'react-hot-toast'

export function RealtimeUpdates() {
  const {
    setConnectionStatus,
    setConnectionError,
    updatePrice,
    addSignalAlert,
    addTradeExecution,
    updatePortfolio,
    addRiskAlert,
  } = useTradingStore()

  const { socket, isConnected, isConnecting, error, connect, subscribe } = useWebSocket({
    url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8090',
    userId: 'dashboard-user',
    autoConnect: true,
  })

  // Update connection status in store
  useEffect(() => {
    setConnectionStatus(isConnected)
    setConnectionError(error)
  }, [isConnected, error, setConnectionStatus, setConnectionError])

  // Subscribe to channels when connected
  useEffect(() => {
    if (isConnected && socket) {
      const channels = [
        'price_updates',
        'signal_alerts',
        'trade_executions',
        'portfolio_updates',
        'risk_alerts',
      ]
      subscribe(channels)
    }
  }, [isConnected, socket, subscribe])

  // Handle WebSocket messages
  useWebSocketMessages(socket, [], (message: WebSocketMessage) => {
    switch (message.type) {
      case 'price_update':
        handlePriceUpdate(message.data as PriceUpdate)
        break
      case 'signal_alert':
        handleSignalAlert(message.data as SignalAlert)
        break
      case 'trade_execution':
        handleTradeExecution(message.data as TradeExecution)
        break
      case 'portfolio_update':
        handlePortfolioUpdate(message.data as PortfolioUpdate)
        break
      case 'risk_alert':
        handleRiskAlert(message.data as RiskAlert)
        break
      default:
        console.log('Unknown message type:', message.type)
    }
  })

  const handlePriceUpdate = (priceUpdate: PriceUpdate) => {
    updatePrice(priceUpdate)
  }

  const handleSignalAlert = (alert: SignalAlert) => {
    addSignalAlert(alert)
    
    // Show toast notification for important signals
    if (alert.confidence > 0.8) {
      const emoji = alert.emoji || '📈'
      toast.success(
        `${emoji} ${alert.strategy}: ${alert.signal} ${alert.symbol}`,
        {
          duration: 5000,
          position: 'top-right',
        }
      )
    }
  }

  const handleTradeExecution = (execution: TradeExecution) => {
    addTradeExecution(execution)
    
    // Show toast notification for trade executions
    const emoji = execution.side === 'buy' ? '🟢' : '🔴'
    const action = execution.side === 'buy' ? 'Bought' : 'Sold'
    
    if (execution.status === 'filled') {
      toast.success(
        `${emoji} ${action} ${execution.amount} ${execution.symbol} at $${execution.price}`,
        {
          duration: 4000,
        }
      )
    } else if (execution.status === 'rejected') {
      toast.error(
        `❌ Trade rejected: ${execution.message}`,
        {
          duration: 6000,
        }
      )
    }
  }

  const handlePortfolioUpdate = (portfolioUpdate: PortfolioUpdate) => {
    updatePortfolio(portfolioUpdate)
  }

  const handleRiskAlert = (alert: RiskAlert) => {
    addRiskAlert(alert)
    
    // Show toast notification for risk alerts
    const getAlertEmoji = (severity: string) => {
      switch (severity) {
        case 'critical': return '🚨'
        case 'high': return '⚠️'
        case 'medium': return '⚡'
        default: return 'ℹ️'
      }
    }

    const emoji = getAlertEmoji(alert.severity)
    
    if (alert.severity === 'critical' || alert.severity === 'high') {
      toast.error(
        `${emoji} Risk Alert: ${alert.message}`,
        {
          duration: 8000,
        }
      )
    } else {
      toast(
        `${emoji} ${alert.message}`,
        {
          duration: 5000,
        }
      )
    }
  }

  // Show connection status changes
  useEffect(() => {
    if (isConnected) {
      toast.success('☕ Connected to Coffee Trading!', {
        duration: 3000,
      })
    } else if (error && !isConnecting) {
      toast.error(`Connection failed: ${error}`, {
        duration: 5000,
      })
    }
  }, [isConnected, error, isConnecting])

  // This component doesn't render anything visible
  return null
}
