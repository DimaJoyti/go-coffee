import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Format currency values
export function formatCurrency(value: number, currency: string = 'USD', decimals: number = 2): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value)
}

// Format percentage values
export function formatPercentage(value: number, decimals: number = 2): string {
  return new Intl.NumberFormat('en-US', {
    style: 'percent',
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value / 100)
}

// Format percentage values (simple format)
export function formatPercent(value: number, decimals: number = 2): string {
  return `${value >= 0 ? '+' : ''}${value.toFixed(decimals)}%`
}

// Format large numbers with K, M, B suffixes
export function formatCompactNumber(value: number): string {
  return new Intl.NumberFormat('en-US', {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(value)
}

// Format crypto values with appropriate decimals
export function formatCrypto(value: number, symbol: string = '', decimals?: number): string {
  let formatDecimals = decimals

  if (formatDecimals === undefined) {
    if (value >= 1) formatDecimals = 2
    else if (value >= 0.01) formatDecimals = 4
    else if (value >= 0.0001) formatDecimals = 6
    else formatDecimals = 8
  }

  const formatted = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: formatDecimals,
    maximumFractionDigits: formatDecimals,
  }).format(value)

  return symbol ? `${formatted} ${symbol}` : formatted
}

// Get price change color class
export function getPriceChangeColor(change: number): string {
  if (change > 0) return 'text-profit'
  if (change < 0) return 'text-loss'
  return 'text-muted-foreground'
}

// Get price change background color class
export function getPriceChangeBgColor(change: number): string {
  if (change > 0) return 'bg-profit/10 text-profit'
  if (change < 0) return 'bg-loss/10 text-loss'
  return 'bg-muted text-muted-foreground'
}

// Get risk level color class
export function getRiskLevelColor(level: string | number): string {
  if (typeof level === 'number') {
    if (level <= 3) return 'text-green-500'
    if (level <= 6) return 'text-yellow-500'
    return 'text-red-500'
  }

  switch (level.toLowerCase()) {
    case 'low': return 'text-green-500'
    case 'medium': return 'text-yellow-500'
    case 'high': return 'text-red-500'
    default: return 'text-muted-foreground'
  }
}

// Get coffee strategy emoji
export function getCoffeeStrategyEmoji(strategyType: string): string {
  switch (strategyType.toLowerCase()) {
    case 'espresso_scalping': return '☕'
    case 'latte_swing': return '🥛'
    case 'cappuccino_momentum': return '🫖'
    case 'americano_trend': return '🌊'
    case 'mocha_arbitrage': return '🍫'
    case 'frappuccino_grid': return '🧊'
    case 'macchiato_mean_reversion': return '🎯'
    case 'cold_brew_hodl': return '❄️'
    case 'turkish_coffee_volatility': return '🔥'
    case 'french_press_dca': return '⏰'
    default: return '☕'
  }
}

// Get status color class
export function getStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'active':
    case 'running':
    case 'enabled':
    case 'online':
    case 'success':
      return 'text-green-500'
    case 'paused':
    case 'pending':
    case 'warning':
      return 'text-yellow-500'
    case 'stopped':
    case 'disabled':
    case 'offline':
    case 'error':
    case 'failed':
      return 'text-red-500'
    case 'draft':
    case 'inactive':
      return 'text-gray-500'
    default:
      return 'text-muted-foreground'
  }
}

// Debounce function
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout
  return (...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

// Throttle function
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

// Generate random ID
export function generateId(): string {
  return Math.random().toString(36).substring(2) + Date.now().toString(36)
}

// Sleep utility
export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

// Clamp number between min and max
export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max)
}

// Calculate percentage change
export function calculatePercentageChange(oldValue: number, newValue: number): number {
  if (oldValue === 0) return 0
  return ((newValue - oldValue) / oldValue) * 100
}

// Format time ago
export function formatTimeAgo(date: Date | string): string {
  const now = new Date()
  const past = new Date(date)
  const diffInSeconds = Math.floor((now.getTime() - past.getTime()) / 1000)

  if (diffInSeconds < 60) return `${diffInSeconds}s ago`
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`
  return `${Math.floor(diffInSeconds / 86400)}d ago`
}

// Validate email
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

// Copy to clipboard
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch (error) {
    console.error('Failed to copy to clipboard:', error)
    return false
  }
}
