'use client'

import { motion } from 'framer-motion'
import { cn } from '@/lib/utils'

// Coffee-themed loading spinner
export function CoffeeLoader({ className = '', size = 'md' }: { className?: string; size?: 'sm' | 'md' | 'lg' }) {
  const sizeClasses = {
    sm: 'w-6 h-6',
    md: 'w-8 h-8',
    lg: 'w-12 h-12'
  }

  return (
    <div className={cn('flex items-center justify-center', className)}>
      <motion.div
        className={cn('relative', sizeClasses[size])}
        animate={{ rotate: 360 }}
        transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
      >
        <div className="absolute inset-0 rounded-full border-2 border-coffee-200/30" />
        <div className="absolute inset-0 rounded-full border-2 border-transparent border-t-coffee-500 animate-spin" />
        <div className="absolute inset-2 rounded-full bg-coffee-400/20 flex items-center justify-center">
          <span className="text-coffee-600 text-xs">☕</span>
        </div>
      </motion.div>
    </div>
  )
}

// Crypto-themed loading animation
export function CryptoLoader({ className = '' }: { className?: string }) {
  return (
    <div className={cn('flex items-center space-x-1', className)}>
      {['₿', 'Ξ', '◎'].map((symbol, index) => (
        <motion.div
          key={symbol}
          className="text-2xl font-bold"
          animate={{
            y: [0, -10, 0],
            color: ['#f7931a', '#627eea', '#9945ff']
          }}
          transition={{
            duration: 1.5,
            repeat: Infinity,
            delay: index * 0.2,
            ease: 'easeInOut'
          }}
        >
          {symbol}
        </motion.div>
      ))}
    </div>
  )
}

// Skeleton loading component
interface SkeletonProps {
  className?: string
  variant?: 'text' | 'circular' | 'rectangular'
  width?: string | number
  height?: string | number
  animation?: 'pulse' | 'wave'
}

export function Skeleton({ 
  className = '', 
  variant = 'text',
  width,
  height,
  animation = 'pulse'
}: SkeletonProps) {
  const baseClasses = 'bg-gradient-to-r from-slate-700/50 to-slate-600/50 rounded'
  
  const variantClasses = {
    text: 'h-4 w-full',
    circular: 'rounded-full',
    rectangular: 'rounded-lg'
  }

  const animationClasses = {
    pulse: 'animate-pulse',
    wave: 'shimmer'
  }

  const style = {
    width: width || undefined,
    height: height || undefined
  }

  return (
    <div
      className={cn(
        baseClasses,
        variantClasses[variant],
        animationClasses[animation],
        className
      )}
      style={style}
    />
  )
}

// Loading card with skeleton content
export function LoadingCard({ className = '' }: { className?: string }) {
  return (
    <div className={cn('metric-card space-y-4', className)}>
      <div className="flex items-center justify-between">
        <Skeleton variant="circular" width={40} height={40} />
        <Skeleton width={60} height={20} />
      </div>
      <div className="space-y-2">
        <Skeleton height={32} width="60%" />
        <Skeleton height={16} width="40%" />
      </div>
    </div>
  )
}

// Floating action button with loading state
interface FloatingActionButtonProps {
  onClick?: () => void
  loading?: boolean
  icon?: React.ReactNode
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

export function FloatingActionButton({
  onClick,
  loading = false,
  icon = '☕',
  className = '',
  size = 'md'
}: FloatingActionButtonProps) {
  const sizeClasses = {
    sm: 'w-12 h-12 text-lg',
    md: 'w-14 h-14 text-xl',
    lg: 'w-16 h-16 text-2xl'
  }

  return (
    <motion.button
      whileHover={{ scale: 1.1 }}
      whileTap={{ scale: 0.9 }}
      onClick={onClick}
      disabled={loading}
      className={cn(
        'fixed bottom-6 right-6 rounded-full shadow-glow hover:shadow-glow-lg',
        'bg-gradient-to-r from-coffee-500 to-brand-amber text-white',
        'flex items-center justify-center transition-all duration-300',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        sizeClasses[size],
        className
      )}
    >
      {loading ? (
        <CoffeeLoader size="sm" />
      ) : (
        <motion.div
          animate={{ rotate: [0, 10, -10, 0] }}
          transition={{ duration: 2, repeat: Infinity }}
        >
          {icon}
        </motion.div>
      )}
    </motion.button>
  )
}

// Progress bar with coffee theme
interface ProgressBarProps {
  value: number
  max?: number
  className?: string
  showLabel?: boolean
  animated?: boolean
}

export function ProgressBar({
  value,
  max = 100,
  className = '',
  showLabel = false,
  animated = true
}: ProgressBarProps) {
  const percentage = Math.min((value / max) * 100, 100)

  return (
    <div className={cn('w-full', className)}>
      {showLabel && (
        <div className="flex justify-between text-sm text-slate-400 mb-2">
          <span>Progress</span>
          <span>{Math.round(percentage)}%</span>
        </div>
      )}
      <div className="w-full bg-slate-700/50 rounded-full h-2 overflow-hidden">
        <motion.div
          className="h-full bg-gradient-to-r from-coffee-500 to-brand-amber rounded-full shadow-glow"
          initial={{ width: 0 }}
          animate={{ width: `${percentage}%` }}
          transition={{ duration: animated ? 0.5 : 0, ease: 'easeOut' }}
        />
      </div>
    </div>
  )
}

// Loading overlay
interface LoadingOverlayProps {
  isVisible: boolean
  message?: string
  className?: string
}

export function LoadingOverlay({
  isVisible,
  message = 'Loading...',
  className = ''
}: LoadingOverlayProps) {
  if (!isVisible) return null

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className={cn(
        'fixed inset-0 bg-slate-900/80 backdrop-blur-sm z-50',
        'flex items-center justify-center',
        className
      )}
    >
      <div className="glass-card p-8 text-center">
        <CoffeeLoader size="lg" className="mb-4" />
        <p className="text-slate-300 font-medium">{message}</p>
      </div>
    </motion.div>
  )
}

// Pulse animation wrapper
export function PulseWrapper({ 
  children, 
  className = '',
  intensity = 'normal'
}: { 
  children: React.ReactNode
  className?: string
  intensity?: 'subtle' | 'normal' | 'strong'
}) {
  const intensityClasses = {
    subtle: 'animate-pulse opacity-80',
    normal: 'animate-pulse',
    strong: 'animate-pulse opacity-60'
  }

  return (
    <div className={cn(intensityClasses[intensity], className)}>
      {children}
    </div>
  )
}

// Typing animation
export function TypingAnimation({ 
  text, 
  speed = 100,
  className = ''
}: { 
  text: string
  speed?: number
  className?: string
}) {
  return (
    <motion.div
      className={cn('font-mono', className)}
      initial={{ width: 0 }}
      animate={{ width: 'auto' }}
      transition={{ duration: (text.length * speed) / 1000 }}
    >
      {text}
      <motion.span
        animate={{ opacity: [1, 0] }}
        transition={{ duration: 0.8, repeat: Infinity }}
        className="text-coffee-400"
      >
        |
      </motion.span>
    </motion.div>
  )
}
