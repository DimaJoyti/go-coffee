import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatCurrency(
  amount: number,
  currency: string = 'USD',
  minimumFractionDigits: number = 2,
  maximumFractionDigits: number = 2
): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(amount)
}

export function formatNumber(
  number: number,
  minimumFractionDigits: number = 0,
  maximumFractionDigits: number = 2
): string {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(number)
}

export function formatPercent(
  number: number,
  minimumFractionDigits: number = 2,
  maximumFractionDigits: number = 2
): string {
  return new Intl.NumberFormat('en-US', {
    style: 'percent',
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(number / 100)
}

export function formatCompactNumber(number: number): string {
  return new Intl.NumberFormat('en-US', {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(number)
}

export function formatTimeAgo(date: string | Date): string {
  const now = new Date()
  const past = new Date(date)
  const diffInSeconds = Math.floor((now.getTime() - past.getTime()) / 1000)

  if (diffInSeconds < 60) {
    return `${diffInSeconds}s ago`
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60)
    return `${minutes}m ago`
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600)
    return `${hours}h ago`
  } else {
    const days = Math.floor(diffInSeconds / 86400)
    return `${days}d ago`
  }
}

export function getPriceChangeColor(change: number): string {
  if (change > 0) return 'text-profit'
  if (change < 0) return 'text-loss'
  return 'text-muted-foreground'
}

export function getPriceChangeIcon(change: number): string {
  if (change > 0) return '↗'
  if (change < 0) return '↘'
  return '→'
}

export function getCoffeeStrategyEmoji(type: string): string {
  const emojis = {
    espresso: '☕',
    latte: '🥛',
    'cold-brew': '🧊',
    cappuccino: '☕',
  }
  return emojis[type as keyof typeof emojis] || '☕'
}

export function getCoffeeStrategyColor(type: string): string {
  const colors = {
    espresso: 'bg-coffee-600',
    latte: 'bg-amber-500',
    'cold-brew': 'bg-blue-500',
    cappuccino: 'bg-coffee-500',
  }
  return colors[type as keyof typeof colors] || 'bg-coffee-600'
}

export function getRiskLevelColor(level: string): string {
  const colors = {
    low: 'text-green-500',
    medium: 'text-yellow-500',
    high: 'text-red-500',
    critical: 'text-red-600',
  }
  return colors[level as keyof typeof colors] || 'text-muted-foreground'
}

export function getStatusColor(status: string): string {
  const colors = {
    active: 'text-green-500',
    inactive: 'text-gray-500',
    paused: 'text-yellow-500',
    filled: 'text-green-500',
    pending: 'text-yellow-500',
    cancelled: 'text-gray-500',
    rejected: 'text-red-500',
  }
  return colors[status as keyof typeof colors] || 'text-muted-foreground'
}

export function calculatePnL(currentPrice: number, averagePrice: number, amount: number) {
  const totalValue = currentPrice * amount
  const totalCost = averagePrice * amount
  const pnl = totalValue - totalCost
  const pnlPercent = totalCost > 0 ? (pnl / totalCost) * 100 : 0
  
  return {
    pnl,
    pnlPercent,
    totalValue,
    totalCost,
  }
}

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null
  
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

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
