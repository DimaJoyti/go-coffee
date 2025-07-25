import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card"
import { Badge } from "../ui/badge"
import { Button } from "../ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs"
import { cn, formatCurrency, formatPercentage } from "../../lib/utils"
import { TrendingUp, TrendingDown, RotateCcw, Maximize2 } from "lucide-react"

interface HeatmapData {
  symbol: string
  name: string
  price: number
  change24h: number
  changePercent24h: number
  marketCap: number
  volume24h: number
  category: string
}

interface MarketHeatmapProps {
  data: HeatmapData[]
  title?: string
  loading?: boolean
  className?: string
  onCellClick?: (item: HeatmapData) => void
}

const MarketHeatmap: React.FC<MarketHeatmapProps> = ({
  data,
  title = "Market Heatmap",
  loading = false,
  className,
  onCellClick,
}) => {
  const [sortBy, setSortBy] = React.useState<"marketCap" | "volume24h" | "changePercent24h">("marketCap")
  const [timeframe, setTimeframe] = React.useState<"24h" | "7d" | "30d">("24h")

  // Sort and filter data
  const sortedData = React.useMemo(() => {
    return [...data]
      .sort((a, b) => {
        switch (sortBy) {
          case "marketCap":
            return b.marketCap - a.marketCap
          case "volume24h":
            return b.volume24h - a.volume24h
          case "changePercent24h":
            return Math.abs(b.changePercent24h) - Math.abs(a.changePercent24h)
          default:
            return 0
        }
      })
      .slice(0, 50) // Show top 50
  }, [data, sortBy])

  // Calculate cell size based on market cap
  const getCellSize = (marketCap: number) => {
    const maxMarketCap = Math.max(...sortedData.map(item => item.marketCap))
    const minMarketCap = Math.min(...sortedData.map(item => item.marketCap))
    const ratio = (marketCap - minMarketCap) / (maxMarketCap - minMarketCap)
    
    // Size between 80px and 200px
    const minSize = 80
    const maxSize = 200
    return minSize + (maxSize - minSize) * ratio
  }

  // Get color based on change percentage
  const getCellColor = (changePercent: number) => {
    const intensity = Math.min(Math.abs(changePercent) / 10, 1) // Max intensity at 10%
    
    if (changePercent > 0) {
      return {
        backgroundColor: `rgba(16, 185, 129, ${0.1 + intensity * 0.4})`,
        borderColor: `rgba(16, 185, 129, ${0.3 + intensity * 0.7})`,
        color: changePercent > 5 ? "#ffffff" : "#10b981"
      }
    } else if (changePercent < 0) {
      return {
        backgroundColor: `rgba(239, 68, 68, ${0.1 + intensity * 0.4})`,
        borderColor: `rgba(239, 68, 68, ${0.3 + intensity * 0.7})`,
        color: changePercent < -5 ? "#ffffff" : "#ef4444"
      }
    } else {
      return {
        backgroundColor: "rgba(107, 114, 128, 0.1)",
        borderColor: "rgba(107, 114, 128, 0.3)",
        color: "#6b7280"
      }
    }
  }

  const LoadingCell = ({ size }: { size: number }) => (
    <div
      className="animate-pulse bg-muted rounded-lg border"
      style={{ width: size, height: size }}
    />
  )

  const HeatmapCell = ({ item }: { item: HeatmapData }) => {
    const size = getCellSize(item.marketCap)
    const colors = getCellColor(item.changePercent24h)
    
    return (
      <div
        className={cn(
          "relative rounded-lg border-2 p-3 cursor-pointer transition-all duration-200 hover:scale-105 hover:shadow-lg",
          "flex flex-col justify-between"
        )}
        style={{
          width: size,
          height: size,
          backgroundColor: colors.backgroundColor,
          borderColor: colors.borderColor,
          color: colors.color,
        }}
        onClick={() => onCellClick?.(item)}
      >
        {/* Symbol */}
        <div className="font-bold text-sm truncate">{item.symbol}</div>
        
        {/* Price */}
        <div className="font-mono text-xs">
          {formatCurrency(item.price, "USD", item.price < 1 ? 4 : 2)}
        </div>
        
        {/* Change */}
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium">
            {formatPercentage(item.changePercent24h / 100)}
          </span>
          {item.changePercent24h > 0 ? (
            <TrendingUp className="h-3 w-3" />
          ) : item.changePercent24h < 0 ? (
            <TrendingDown className="h-3 w-3" />
          ) : null}
        </div>
        
        {/* Market Cap (for larger cells) */}
        {size > 120 && (
          <div className="text-2xs opacity-75 truncate">
            MC: {formatCurrency(item.marketCap, "USD", 0)}
          </div>
        )}
      </div>
    )
  }

  return (
    <Card className={cn("w-full", className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center space-x-2">
            <span>{title}</span>
            <Badge variant="outline">{timeframe}</Badge>
          </CardTitle>
          <div className="flex items-center space-x-2">
            <Button variant="outline" size="sm" onClick={() => window.location.reload()}>
              <RotateCcw className="h-4 w-4 mr-1" />
              Refresh
            </Button>
            <Button variant="outline" size="sm">
              <Maximize2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
        
        <Tabs value={timeframe} onValueChange={(value) => setTimeframe(value as any)}>
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="24h">24H</TabsTrigger>
            <TabsTrigger value="7d">7D</TabsTrigger>
            <TabsTrigger value="30d">30D</TabsTrigger>
          </TabsList>
        </Tabs>
        
        <div className="flex items-center space-x-2">
          <span className="text-sm text-muted-foreground">Sort by:</span>
          <Button
            variant={sortBy === "marketCap" ? "default" : "outline"}
            size="sm"
            onClick={() => setSortBy("marketCap")}
          >
            Market Cap
          </Button>
          <Button
            variant={sortBy === "volume24h" ? "default" : "outline"}
            size="sm"
            onClick={() => setSortBy("volume24h")}
          >
            Volume
          </Button>
          <Button
            variant={sortBy === "changePercent24h" ? "default" : "outline"}
            size="sm"
            onClick={() => setSortBy("changePercent24h")}
          >
            Change %
          </Button>
        </div>
      </CardHeader>
      
      <CardContent>
        <div className="flex flex-wrap gap-3 justify-center">
          {loading ? (
            Array.from({ length: 20 }).map((_, index) => (
              <LoadingCell key={index} size={120} />
            ))
          ) : (
            sortedData.map((item) => (
              <HeatmapCell key={item.symbol} item={item} />
            ))
          )}
        </div>
        
        {!loading && sortedData.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            No market data available
          </div>
        )}
        
        {/* Legend */}
        <div className="mt-6 flex items-center justify-center space-x-6 text-sm text-muted-foreground">
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-green-500/20 border border-green-500/50 rounded"></div>
            <span>Positive</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-red-500/20 border border-red-500/50 rounded"></div>
            <span>Negative</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-gray-500/20 border border-gray-500/50 rounded"></div>
            <span>Neutral</span>
          </div>
          <span>• Size = Market Cap</span>
        </div>
      </CardContent>
    </Card>
  )
}

export { MarketHeatmap, type HeatmapData }
