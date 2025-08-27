'use client'

import { motion, AnimatePresence } from 'framer-motion'
import { useState, useEffect } from 'react'
import { useRealTimeData } from '@/contexts/RealTimeDataContext'

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message: string
  timestamp: number
  duration?: number
}

interface NotificationSystemProps {
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left'
  maxNotifications?: number
}

export default function NotificationSystem({ 
  position = 'top-right', 
  maxNotifications = 5 
}: NotificationSystemProps) {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const { prices } = useRealTimeData()

  // Monitor price changes and create notifications
  useEffect(() => {
    const priceArray = Array.from(prices.values())
    
    priceArray.forEach(priceData => {
      // Create notification for significant price changes
      if (Math.abs(priceData.change24h) > 5) {
        const notification: Notification = {
          id: `price-${priceData.symbol}-${Date.now()}`,
          type: priceData.change24h > 0 ? 'success' : 'warning',
          title: `${priceData.symbol} Price Alert`,
          message: `${priceData.change24h > 0 ? '📈' : '📉'} ${priceData.change24h.toFixed(2)}% change in 24h`,
          timestamp: Date.now(),
          duration: 5000
        }
        
        addNotification(notification)
      }
    })
  }, [prices])

  const addNotification = (notification: Notification) => {
    setNotifications(prev => {
      // Check if similar notification already exists
      const exists = prev.some(n => 
        n.title === notification.title && 
        Date.now() - n.timestamp < 60000 // Within last minute
      )
      
      if (exists) return prev
      
      const newNotifications = [notification, ...prev.slice(0, maxNotifications - 1)]
      
      // Auto-remove notification after duration
      if (notification.duration) {
        setTimeout(() => {
          removeNotification(notification.id)
        }, notification.duration)
      }
      
      return newNotifications
    })
  }

  const removeNotification = (id: string) => {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  const getPositionClasses = () => {
    switch (position) {
      case 'top-left':
        return 'top-4 left-4'
      case 'bottom-right':
        return 'bottom-4 right-4'
      case 'bottom-left':
        return 'bottom-4 left-4'
      default:
        return 'top-4 right-4'
    }
  }

  const getNotificationIcon = (type: Notification['type']) => {
    switch (type) {
      case 'success':
        return '✅'
      case 'error':
        return '❌'
      case 'warning':
        return '⚠️'
      case 'info':
        return 'ℹ️'
      default:
        return 'ℹ️'
    }
  }

  const getNotificationColors = (type: Notification['type']) => {
    switch (type) {
      case 'success':
        return 'bg-status-success/20 border-status-success/30 text-status-success shadow-lg shadow-status-success/10'
      case 'error':
        return 'bg-status-error/20 border-status-error/30 text-status-error shadow-lg shadow-status-error/10'
      case 'warning':
        return 'bg-status-warning/20 border-status-warning/30 text-status-warning shadow-lg shadow-status-warning/10'
      case 'info':
        return 'bg-status-info/20 border-status-info/30 text-status-info shadow-lg shadow-status-info/10'
      default:
        return 'bg-coffee-500/20 border-coffee-500/30 text-coffee-400 shadow-glow'
    }
  }

  return (
    <div className={`fixed ${getPositionClasses()} z-50 space-y-2 pointer-events-none`}>
      <AnimatePresence>
        {notifications.map((notification) => (
          <motion.div
            key={notification.id}
            initial={{ opacity: 0, x: position.includes('right') ? 300 : -300, scale: 0.8 }}
            animate={{ opacity: 1, x: 0, scale: 1 }}
            exit={{ opacity: 0, x: position.includes('right') ? 300 : -300, scale: 0.8 }}
            transition={{ duration: 0.3, ease: "easeOut" }}
            className={`
              max-w-sm w-full glass-card p-4 pointer-events-auto hover-lift
              ${getNotificationColors(notification.type)}
            `}
          >
            <div className="flex items-start space-x-3">
              <div className="text-xl flex-shrink-0">
                {getNotificationIcon(notification.type)}
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-semibold text-white mb-1">
                  {notification.title}
                </h4>
                <p className="text-sm opacity-90">
                  {notification.message}
                </p>
                <p className="text-xs opacity-60 mt-1">
                  {new Date(notification.timestamp).toLocaleTimeString()}
                </p>
              </div>
              <motion.button
                onClick={() => removeNotification(notification.id)}
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
                className="text-white/60 hover:text-white/90 transition-colors duration-200 flex-shrink-0"
              >
                ✕
              </motion.button>
            </div>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  )
}

// Hook to use notification system from anywhere
export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([])

  const addNotification = (notification: Omit<Notification, 'id' | 'timestamp'>) => {
    const newNotification: Notification = {
      ...notification,
      id: `notification-${Date.now()}-${Math.random()}`,
      timestamp: Date.now()
    }
    
    setNotifications(prev => [newNotification, ...prev.slice(0, 4)])
    
    // Auto-remove after duration
    if (notification.duration) {
      setTimeout(() => {
        removeNotification(newNotification.id)
      }, notification.duration)
    }
  }

  const removeNotification = (id: string) => {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  const clearAll = () => {
    setNotifications([])
  }

  return {
    notifications,
    addNotification,
    removeNotification,
    clearAll
  }
}

// Predefined notification creators
export const createSuccessNotification = (title: string, message: string): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'success',
  title,
  message,
  duration: 4000
})

export const createErrorNotification = (title: string, message: string): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'error',
  title,
  message,
  duration: 6000
})

export const createWarningNotification = (title: string, message: string): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'warning',
  title,
  message,
  duration: 5000
})

export const createInfoNotification = (title: string, message: string): Omit<Notification, 'id' | 'timestamp'> => ({
  type: 'info',
  title,
  message,
  duration: 4000
})
