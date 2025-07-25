'use client'

import * as React from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn, formatCurrency, formatCompactNumber } from "@/lib/utils"
import { Activity, TrendingUp, Settings, Layers } from "lucide-react"

interface OrderBookEntry {
  price: number
  size: number
  total: number
  count?: number
}

interface OrderBookData {
  bids: OrderBookEntry[]
  asks: OrderBookEntry[]
  spread: number
  spreadPercent: number
  lastPrice: number
  lastUpdate: number
}

interface OrderBookProps {
  symbol?: string
  precision?: number
  maxRows?: number
  loading?: boolean
  className?: string
  onPriceClick?: (price: number, side: "buy" | "sell") => void
}

// Mock data for demonstration
const mockOrderBookData: OrderBookData = {
  bids: Array.from({ length: 20 }, (_, i) => ({
    price: 43000 - i * 10,
    size: Math.random() * 5 + 0.1,
    total: Math.random() * 100 + 10,
    count: Math.floor(Math.random() * 50) + 1
  })),
  asks: Array.from({ length: 20 }, (_, i) => ({
    price: 43010 + i * 10,
    size: Math.random() * 5 + 0.1,
    total: Math.random() * 100 + 10,
    count: Math.floor(Math.random() * 50) + 1
  })),
  spread: 10,
  spreadPercent: 0.023,
  lastPrice: 43005,
  lastUpdate: Date.now()
}

const OrderBook: React.FC<OrderBookProps> = ({
  symbol = "BTC/USD",
  precision = 2,
  maxRows = 15,
  loading = false,
  className,
  onPriceClick,
}) => {
  const [data, setData] = React.useState<OrderBookData>(mockOrderBookData)
  const [showDepth, setShowDepth] = React.useState(true)

  // Simulate real-time updates
  React.useEffect(() => {
    const interval = setInterval(() => {
      setData(prev => ({
        ...prev,
        bids: prev.bids.map(bid => ({
          ...bid,
          size: Math.max(0.01, bid.size + (Math.random() - 0.5) * 0.5)
        })),
        asks: prev.asks.map(ask => ({
          ...ask,
          size: Math.max(0.01, ask.size + (Math.random() - 0.5) * 0.5)
        })),
        lastUpdate: Date.now()
      }))
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  const groupedBids = data.bids.slice(0, maxRows)
  const groupedAsks = data.asks.slice(0, maxRows)

  // Calculate max size for depth visualization
  const maxBidSize = Math.max(...groupedBids.map(bid => bid.size), 0)
  const maxAskSize = Math.max(...groupedAsks.map(ask => ask.size), 0)
  const maxSize = Math.max(maxBidSize, maxAskSize)

  const OrderRow = ({ 
    order, 
    side, 
    maxSize 
  }: { 
    order: OrderBookEntry
    side: "bid" | "ask"
    maxSize: number 
  }) => {
    const depthPercent = showDepth ? (order.size / maxSize) * 100 : 0
    const isBid = side === "bid"
    
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className={cn(
          "relative grid grid-cols-3 gap-2 py-1 px-2 text-xs font-mono cursor-pointer hover:bg-muted/50 transition-colors",
          isBid ? "hover:bg-bull/10" : "hover:bg-bear/10"
        )}
        onClick={() => onPriceClick?.(order.price, isBid ? "buy" : "sell")}
      >
        {/* Depth visualization */}
        {showDepth && (
          <div
            className={cn(
              "absolute inset-y-0 right-0 opacity-20 transition-all",
              isBid ? "bg-bull" : "bg-bear"
            )}
            style={{ width: `${depthPercent}%` }}
          />
        )}
        
        <div className={cn(
          "text-right font-medium",
          isBid ? "text-bull" : "text-bear"
        )}>
          {formatCurrency(order.price, 'USD', precision)}
        </div>
        <div className="text-right text-muted-foreground">
          {order.size.toFixed(4)}
        </div>
        <div className="text-right text-muted-foreground">
          {formatCompactNumber(order.total)}
        </div>
      </motion.div>
    )
  }

  if (loading) {
    return (
      <Card variant="glass" className={cn("w-full", className)}>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Activity className="h-5 w-5 animate-pulse" />
            <span>Order Book</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.5 }}
    >
      <Card variant="glass" className={cn("w-full", className)}>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Layers className="h-5 w-5 text-primary" />
              <span>Order Book</span>
              <Badge variant="glass">{symbol}</Badge>
            </CardTitle>
            
            <div className="flex items-center space-x-2">
              <Button
                variant="glass"
                size="sm"
                onClick={() => setShowDepth(!showDepth)}
              >
                Depth: {showDepth ? "ON" : "OFF"}
              </Button>
              <Button variant="glass" size="sm">
                <Settings className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        
        <CardContent className="p-0">
          {/* Header */}
          <div className="grid grid-cols-3 gap-2 py-2 px-2 text-xs font-medium text-muted-foreground border-b">
            <div className="text-right">Price (USD)</div>
            <div className="text-right">Size (BTC)</div>
            <div className="text-right">Total</div>
          </div>
          
          {/* Asks (Sell Orders) */}
          <div className="max-h-48 overflow-y-auto scrollbar-thin">
            {groupedAsks.reverse().map((ask, index) => (
              <OrderRow
                key={`ask-${ask.price}-${index}`}
                order={ask}
                side="ask"
                maxSize={maxSize}
              />
            ))}
          </div>
          
          {/* Spread */}
          <div className="py-3 px-2 border-y glass">
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center space-x-2">
                <span className="text-muted-foreground">Spread:</span>
                <span className="font-mono font-medium">
                  {formatCurrency(data.spread, 'USD', precision)}
                </span>
                <span className="text-xs text-muted-foreground">
                  ({data.spreadPercent.toFixed(3)}%)
                </span>
              </div>
              <div className="flex items-center space-x-1">
                <TrendingUp className="h-4 w-4 text-bull" />
                <span className="font-mono font-medium">
                  {formatCurrency(data.lastPrice, 'USD', precision)}
                </span>
              </div>
            </div>
          </div>
          
          {/* Bids (Buy Orders) */}
          <div className="max-h-48 overflow-y-auto scrollbar-thin">
            {groupedBids.map((bid, index) => (
              <OrderRow
                key={`bid-${bid.price}-${index}`}
                order={bid}
                side="bid"
                maxSize={maxSize}
              />
            ))}
          </div>
          
          {/* Footer */}
          <div className="p-2 border-t glass">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>Real-time order book</span>
              <span>Updated: {new Date(data.lastUpdate).toLocaleTimeString()}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}

export { OrderBook }
