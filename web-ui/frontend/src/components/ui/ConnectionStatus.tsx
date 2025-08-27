'use client'

import { motion } from 'framer-motion'
import { useConnectionStatus } from '@/contexts/RealTimeDataContext'

interface ConnectionStatusProps {
  className?: string
  showText?: boolean
  size?: 'sm' | 'md' | 'lg'
}

export default function ConnectionStatus({ 
  className = '', 
  showText = true, 
  size = 'md' 
}: ConnectionStatusProps) {
  const { isConnected, connectionStatus } = useConnectionStatus()

  const getStatusColor = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'bg-green-500 text-green-400 border-green-500/30'
      case 'connecting':
        return 'bg-yellow-500 text-yellow-400 border-yellow-500/30'
      case 'disconnected':
        return 'bg-red-500 text-red-400 border-red-500/30'
      case 'error':
        return 'bg-red-500 text-red-400 border-red-500/30'
      default:
        return 'bg-slate-500 text-slate-400 border-slate-500/30'
    }
  }

  const getStatusText = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'Connected'
      case 'connecting':
        return 'Connecting...'
      case 'disconnected':
        return 'Disconnected'
      case 'error':
        return 'Connection Error'
      default:
        return 'Unknown'
    }
  }

  const getStatusIcon = () => {
    switch (connectionStatus) {
      case 'connected':
        return '🟢'
      case 'connecting':
        return '🟡'
      case 'disconnected':
        return '🔴'
      case 'error':
        return '❌'
      default:
        return '⚪'
    }
  }

  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return {
          container: 'px-2 py-1 text-xs',
          dot: 'w-1.5 h-1.5',
          text: 'text-xs'
        }
      case 'lg':
        return {
          container: 'px-4 py-3 text-base',
          dot: 'w-3 h-3',
          text: 'text-base'
        }
      default:
        return {
          container: 'px-3 py-2 text-sm',
          dot: 'w-2 h-2',
          text: 'text-sm'
        }
    }
  }

  const sizeClasses = getSizeClasses()
  const statusColor = getStatusColor()

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.8 }}
      animate={{ opacity: 1, scale: 1 }}
      className={`
        flex items-center space-x-2 rounded-lg border backdrop-blur-sm
        ${statusColor.replace('bg-', 'bg-').replace('/30', '/20')} 
        ${statusColor.replace('bg-', 'border-')}
        ${sizeClasses.container}
        ${className}
      `}
    >
      <motion.div
        animate={{
          scale: isConnected && connectionStatus === 'connected' ? [1, 1.2, 1] : 1,
          opacity: connectionStatus === 'connecting' ? [0.5, 1, 0.5] : 1
        }}
        transition={{
          duration: connectionStatus === 'connecting' ? 1.5 : 0.6,
          repeat: connectionStatus === 'connecting' ? Infinity : 0,
          ease: "easeInOut"
        }}
        className={`
          ${sizeClasses.dot} 
          rounded-full
          ${statusColor.split(' ')[0]}
        `}
      />
      
      {showText && (
        <span className={`font-medium ${statusColor.split(' ')[1]} ${sizeClasses.text}`}>
          {getStatusText()}
        </span>
      )}
    </motion.div>
  )
}

// Minimal dot indicator
export function ConnectionDot({ className = '' }: { className?: string }) {
  const { isConnected, connectionStatus } = useConnectionStatus()

  const getDotColor = () => {
    switch (connectionStatus) {
      case 'connected':
        return 'bg-green-400'
      case 'connecting':
        return 'bg-yellow-400'
      case 'disconnected':
        return 'bg-red-400'
      case 'error':
        return 'bg-red-400'
      default:
        return 'bg-slate-400'
    }
  }

  return (
    <motion.div
      animate={{
        scale: isConnected && connectionStatus === 'connected' ? [1, 1.3, 1] : 1,
        opacity: connectionStatus === 'connecting' ? [0.3, 1, 0.3] : 1
      }}
      transition={{
        duration: connectionStatus === 'connecting' ? 1.5 : 0.6,
        repeat: connectionStatus === 'connecting' ? Infinity : 0,
        ease: "easeInOut"
      }}
      className={`w-2 h-2 rounded-full ${getDotColor()} ${className}`}
    />
  )
}

// Detailed connection info panel
export function ConnectionInfo({ className = '' }: { className?: string }) {
  const { isConnected, connectionStatus } = useConnectionStatus()

  return (
    <div className={`bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-xl p-4 ${className}`}>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-white font-semibold">Connection Status</h3>
        <ConnectionDot />
      </div>
      
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-slate-400">Status:</span>
          <span className={`font-medium ${
            isConnected ? 'text-green-400' : 'text-red-400'
          }`}>
            {connectionStatus.charAt(0).toUpperCase() + connectionStatus.slice(1)}
          </span>
        </div>
        
        <div className="flex justify-between">
          <span className="text-slate-400">Data Feed:</span>
          <span className={`font-medium ${
            isConnected ? 'text-green-400' : 'text-red-400'
          }`}>
            {isConnected ? 'Active' : 'Inactive'}
          </span>
        </div>
        
        <div className="flex justify-between">
          <span className="text-slate-400">Last Update:</span>
          <span className="text-white font-medium">
            {new Date().toLocaleTimeString()}
          </span>
        </div>
      </div>
      
      {!isConnected && (
        <div className="mt-3 p-2 bg-red-500/10 border border-red-500/20 rounded-lg">
          <p className="text-red-400 text-xs">
            Real-time data unavailable. Some features may be limited.
          </p>
        </div>
      )}
    </div>
  )
}

// Connection status for header/navbar
export function HeaderConnectionStatus({ className = '' }: { className?: string }) {
  const { isConnected, connectionStatus } = useConnectionStatus()

  return (
    <div className={`flex items-center space-x-2 ${className}`}>
      <ConnectionDot />
      <span className={`text-sm font-medium ${
        isConnected ? 'text-green-400' : 'text-red-400'
      }`}>
        {isConnected ? 'Live' : 'Offline'}
      </span>
    </div>
  )
}
