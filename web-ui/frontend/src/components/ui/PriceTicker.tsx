'use client'

import { motion } from 'framer-motion'
import { usePrices } from '@/contexts/RealTimeDataContext'

interface PriceTickerProps {
  className?: string
  showChange?: boolean
  autoScroll?: boolean
}

export default function PriceTicker({
  className = '',
  showChange = true,
  autoScroll = true
}: PriceTickerProps) {
  const { prices } = usePrices()

  // Fallback data if real-time data isn't loaded yet
  const fallbackPrices = [
    { symbol: 'BTC/USDT', price: 45234.56, change24h: 2.34, volume: 1234567890, timestamp: Date.now() },
    { symbol: 'ETH/USDT', price: 2834.12, change24h: 1.87, volume: 987654321, timestamp: Date.now() },
    { symbol: 'SOL/USDT', price: 98.45, change24h: -0.56, volume: 456789123, timestamp: Date.now() },
    { symbol: 'COFFEE/USDT', price: 12.34, change24h: 5.67, volume: 123456789, timestamp: Date.now() }
  ]

  const displayPrices = prices.length > 0 ? prices : fallbackPrices

  if (displayPrices.length === 0) {
    return (
      <div className={`glass-card p-4 ${className}`}>
        <div className="flex items-center justify-center">
          <div className="animate-pulse text-slate-300 flex items-center gap-2">
            <div className="w-4 h-4 border-2 border-coffee-400 border-t-transparent rounded-full animate-spin" />
            Loading market data...
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={`glass-card overflow-hidden ${className}`}>
      <div className="flex items-center h-16 px-4">
        <div className="flex items-center space-x-2 mr-4">
          <div className="w-2 h-2 bg-status-success rounded-full animate-pulse shadow-glow" />
          <span className="text-slate-300 text-sm font-medium">Live Prices</span>
        </div>
        
        <div className={`flex items-center space-x-8 ${autoScroll ? 'animate-scroll' : ''}`}>
          {displayPrices.map((priceData, index) => (
            <motion.div
              key={priceData.symbol}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="flex items-center space-x-3 flex-shrink-0"
            >
              <div className="flex items-center space-x-2">
                <div className="w-6 h-6 bg-gradient-to-r from-amber-500 to-orange-500 rounded-full flex items-center justify-center text-white text-xs font-bold">
                  {priceData.symbol.split('/')[0].charAt(0)}
                </div>
                <span className="text-white font-medium text-sm">
                  {priceData.symbol.split('/')[0]}
                </span>
              </div>
              
              <div className="text-right">
                <div className="text-white font-semibold">
                  ${priceData.price.toLocaleString('en-US', { 
                    minimumFractionDigits: 2, 
                    maximumFractionDigits: 2 
                  })}
                </div>
                {showChange && (
                  <div className={`text-xs font-medium ${
                    priceData.change24h >= 0 ? 'text-green-400' : 'text-red-400'
                  }`}>
                    {priceData.change24h >= 0 ? '+' : ''}{priceData.change24h.toFixed(2)}%
                  </div>
                )}
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </div>
  )
}

// Compact version for smaller spaces
export function CompactPriceTicker({ className = '' }: { className?: string }) {
  const { prices } = usePrices()

  if (prices.length === 0) {
    return (
      <div className={`flex items-center space-x-2 ${className}`}>
        <div className="animate-pulse text-slate-400 text-sm">Loading...</div>
      </div>
    )
  }

  return (
    <div className={`flex items-center space-x-6 ${className}`}>
      {prices.slice(0, 4).map((priceData) => (
        <motion.div
          key={priceData.symbol}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="flex items-center space-x-2"
        >
          <span className="text-slate-400 text-sm">
            {priceData.symbol.split('/')[0]}
          </span>
          <span className="text-white font-medium text-sm">
            ${priceData.price.toLocaleString()}
          </span>
          <span className={`text-xs ${
            priceData.change24h >= 0 ? 'text-green-400' : 'text-red-400'
          }`}>
            {priceData.change24h >= 0 ? '+' : ''}{priceData.change24h.toFixed(1)}%
          </span>
        </motion.div>
      ))}
    </div>
  )
}

// Single price display component
interface SinglePriceDisplayProps {
  symbol: string
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

export function SinglePriceDisplay({ 
  symbol, 
  className = '', 
  size = 'md' 
}: SinglePriceDisplayProps) {
  const { getPrice } = usePrices()
  const priceData = getPrice(symbol)

  if (!priceData) {
    return (
      <div className={`animate-pulse ${className}`}>
        <div className="bg-slate-700/50 rounded h-6 w-24" />
      </div>
    )
  }

  const sizeClasses = {
    sm: 'text-sm',
    md: 'text-base',
    lg: 'text-lg'
  }

  return (
    <motion.div
      key={priceData.price}
      initial={{ scale: 1 }}
      animate={{ scale: [1, 1.05, 1] }}
      transition={{ duration: 0.3 }}
      className={`flex items-center space-x-2 ${className}`}
    >
      <span className={`text-white font-bold ${sizeClasses[size]}`}>
        ${priceData.price.toLocaleString('en-US', { 
          minimumFractionDigits: 2, 
          maximumFractionDigits: 2 
        })}
      </span>
      <span className={`font-medium ${sizeClasses[size]} ${
        priceData.change24h >= 0 ? 'text-green-400' : 'text-red-400'
      }`}>
        {priceData.change24h >= 0 ? '+' : ''}{priceData.change24h.toFixed(2)}%
      </span>
    </motion.div>
  )
}

// Price change indicator
export function PriceChangeIndicator({ 
  symbol, 
  className = '' 
}: { 
  symbol: string
  className?: string 
}) {
  const { getPrice } = usePrices()
  const priceData = getPrice(symbol)

  if (!priceData) return null

  const isPositive = priceData.change24h >= 0

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.8 }}
      animate={{ opacity: 1, scale: 1 }}
      className={`flex items-center space-x-1 ${className}`}
    >
      <motion.div
        animate={{ 
          rotate: isPositive ? 0 : 180,
          color: isPositive ? '#10b981' : '#ef4444'
        }}
        transition={{ duration: 0.3 }}
      >
        ↗
      </motion.div>
      <span className={`text-sm font-medium ${
        isPositive ? 'text-green-400' : 'text-red-400'
      }`}>
        {Math.abs(priceData.change24h).toFixed(2)}%
      </span>
    </motion.div>
  )
}
