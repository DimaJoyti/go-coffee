import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card"
import { Badge } from "../ui/badge"
import { Button } from "../ui/button"
import { Separator } from "../ui/separator"
import { cn, formatCurrency, formatNumber } from "../../lib/utils"
import { Activity, TrendingUp, TrendingDown, Settings } from "lucide-react"

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
  data: OrderBookData
  symbol?: string
  precision?: number
  maxRows?: number
  loading?: boolean
  className?: string
  onPriceClick?: (price: number, side: "buy" | "sell") => void
}

const OrderBook: React.FC<OrderBookProps> = ({
  data,
  symbol = "BTC/USD",
  precision = 2,
  maxRows = 15,
  loading = false,
  className,
  onPriceClick,
}) => {
  const [grouping, setGrouping] = React.useState(0.01)
  const [showDepth, setShowDepth] = React.useState(true)

  // Group orders by price level
  const groupOrders = (orders: OrderBookEntry[], groupSize: number) => {
    const grouped = new Map<number, OrderBookEntry>()
    
    orders.forEach(order => {
      const groupedPrice = Math.floor(order.price / groupSize) * groupSize
      const existing = grouped.get(groupedPrice)
      
      if (existing) {
        existing.size += order.size
        existing.total += order.total
        existing.count = (existing.count || 0) + (order.count || 1)
      } else {
        grouped.set(groupedPrice, {
          price: groupedPrice,
          size: order.size,
          total: order.total,
          count: order.count || 1
        })
      }
    })
    
    return Array.from(grouped.values())
  }

  const groupedBids = React.useMemo(() => 
    groupOrders(data.bids, grouping)
      .sort((a, b) => b.price - a.price)
      .slice(0, maxRows),
    [data.bids, grouping, maxRows]
  )

  const groupedAsks = React.useMemo(() => 
    groupOrders(data.asks, grouping)
      .sort((a, b) => a.price - b.price)
      .slice(0, maxRows),
    [data.asks, grouping, maxRows]
  )

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
      <div
        className={cn(
          "relative grid grid-cols-3 gap-2 py-1 px-2 text-xs font-mono cursor-pointer transition-colors hover:bg-muted/50",
          "border-l-2 border-transparent hover:border-primary"
        )}
        onClick={() => onPriceClick?.(order.price, isBid ? "buy" : "sell")}
      >
        {/* Depth Background */}
        {showDepth && (
          <div
            className={cn(
              "absolute inset-0 opacity-20",
              isBid ? "bg-green-500" : "bg-red-500"
            )}
            style={{
              width: `${depthPercent}%`,
              right: isBid ? 0 : "auto",
              left: isBid ? "auto" : 0,
            }}
          />
        )}
        
        {/* Price */}
        <div className={cn(
          "relative z-10 text-right",
          isBid ? "text-green-400" : "text-red-400"
        )}>
          {formatCurrency(order.price, "USD", precision)}
        </div>
        
        {/* Size */}
        <div className="relative z-10 text-right text-foreground">
          {formatNumber(order.size, 0, 4)}
        </div>
        
        {/* Total */}
        <div className="relative z-10 text-right text-muted-foreground">
          {formatNumber(order.total, 0, 2)}
        </div>
      </div>
    )
  }

  const LoadingRow = () => (
    <div className="grid grid-cols-3 gap-2 py-1 px-2">
      <div className="h-4 bg-muted rounded animate-pulse"></div>
      <div className="h-4 bg-muted rounded animate-pulse"></div>
      <div className="h-4 bg-muted rounded animate-pulse"></div>
    </div>
  )

  return (
    <Card className={cn("w-full max-w-md", className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center space-x-2">
            <Activity className="h-4 w-4" />
            <span>Order Book</span>
          </CardTitle>
          <div className="flex items-center space-x-2">
            <Badge variant="outline" className="text-xs">
              {symbol}
            </Badge>
            <Button variant="ghost" size="icon" className="h-6 w-6">
              <Settings className="h-3 w-3" />
            </Button>
          </div>
        </div>
        
        {/* Spread Info */}
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Spread:</span>
          <div className="flex items-center space-x-2">
            <span className="font-mono">
              {formatCurrency(data.spread, "USD", precision)}
            </span>
            <Badge variant="outline" className="text-xs">
              {formatNumber(data.spreadPercent, 2)}%
            </Badge>
          </div>
        </div>
        
        {/* Controls */}
        <div className="flex items-center justify-between text-xs">
          <div className="flex items-center space-x-2">
            <span className="text-muted-foreground">Group:</span>
            <select
              value={grouping}
              onChange={(e) => setGrouping(Number(e.target.value))}
              className="bg-background border border-border rounded px-1 py-0.5"
            >
              <option value={0.01}>0.01</option>
              <option value={0.1}>0.1</option>
              <option value={1}>1.0</option>
              <option value={10}>10</option>
            </select>
          </div>
          <Button
            variant={showDepth ? "default" : "outline"}
            size="sm"
            onClick={() => setShowDepth(!showDepth)}
            className="h-6 text-xs"
          >
            Depth
          </Button>
        </div>
      </CardHeader>
      
      <CardContent className="p-0">
        {/* Header */}
        <div className="grid grid-cols-3 gap-2 py-2 px-2 text-xs font-medium text-muted-foreground border-b">
          <div className="text-right">Price</div>
          <div className="text-right">Size</div>
          <div className="text-right">Total</div>
        </div>
        
        {/* Asks (Sell Orders) */}
        <div className="max-h-48 overflow-y-auto scrollbar-thin">
          {loading ? (
            Array.from({ length: maxRows }).map((_, index) => (
              <LoadingRow key={`ask-${index}`} />
            ))
          ) : (
            [...groupedAsks].reverse().map((ask, index) => (
              <OrderRow
                key={`ask-${ask.price}-${index}`}
                order={ask}
                side="ask"
                maxSize={maxSize}
              />
            ))
          )}
        </div>
        
        {/* Current Price */}
        <div className="py-3 px-2 bg-muted/30 border-y">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">Last Price</span>
              {data.lastPrice > data.bids[0]?.price ? (
                <TrendingUp className="h-3 w-3 text-green-500" />
              ) : (
                <TrendingDown className="h-3 w-3 text-red-500" />
              )}
            </div>
            <span className={cn(
              "font-mono font-bold",
              data.lastPrice > data.bids[0]?.price ? "text-green-500" : "text-red-500"
            )}>
              {formatCurrency(data.lastPrice, "USD", precision)}
            </span>
          </div>
        </div>
        
        {/* Bids (Buy Orders) */}
        <div className="max-h-48 overflow-y-auto scrollbar-thin">
          {loading ? (
            Array.from({ length: maxRows }).map((_, index) => (
              <LoadingRow key={`bid-${index}`} />
            ))
          ) : (
            groupedBids.map((bid, index) => (
              <OrderRow
                key={`bid-${bid.price}-${index}`}
                order={bid}
                side="bid"
                maxSize={maxSize}
              />
            ))
          )}
        </div>
        
        {/* Footer */}
        <div className="p-2 text-xs text-muted-foreground text-center border-t">
          Last updated: {new Date(data.lastUpdate).toLocaleTimeString()}
        </div>
      </CardContent>
    </Card>
  )
}

export { OrderBook, type OrderBookData, type OrderBookEntry }
