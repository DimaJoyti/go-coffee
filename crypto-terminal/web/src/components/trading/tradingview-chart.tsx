import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs"
import { cn } from "../../lib/utils"
import { 
  Maximize2, 
  Settings, 
  TrendingUp, 
  BarChart3, 
  Candlestick,
  LineChart,
  Activity
} from "lucide-react"

interface TradingViewChartProps {
  symbol?: string
  interval?: string
  theme?: "light" | "dark"
  height?: number
  showToolbar?: boolean
  showVolumeProfile?: boolean
  className?: string
  onSymbolChange?: (symbol: string) => void
  onIntervalChange?: (interval: string) => void
}

const TradingViewChart: React.FC<TradingViewChartProps> = ({
  symbol = "BTCUSD",
  interval = "1D",
  theme = "dark",
  height = 500,
  showToolbar = true,
  showVolumeProfile = false,
  className,
  onSymbolChange,
  onIntervalChange,
}) => {
  const chartContainerRef = React.useRef<HTMLDivElement>(null)
  const [isFullscreen, setIsFullscreen] = React.useState(false)
  const [chartType, setChartType] = React.useState<"candlestick" | "line" | "area">("candlestick")
  const [activeInterval, setActiveInterval] = React.useState(interval)

  const intervals = [
    { label: "1m", value: "1" },
    { label: "5m", value: "5" },
    { label: "15m", value: "15" },
    { label: "1h", value: "60" },
    { label: "4h", value: "240" },
    { label: "1D", value: "1D" },
    { label: "1W", value: "1W" },
    { label: "1M", value: "1M" },
  ]

  // Initialize TradingView widget
  React.useEffect(() => {
    if (!chartContainerRef.current) return

    // Clear previous widget
    chartContainerRef.current.innerHTML = ""

    // Create TradingView widget script
    const script = document.createElement("script")
    script.src = "https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js"
    script.type = "text/javascript"
    script.async = true
    script.innerHTML = JSON.stringify({
      autosize: true,
      symbol: symbol,
      interval: activeInterval,
      timezone: "Etc/UTC",
      theme: theme,
      style: chartType === "candlestick" ? "1" : chartType === "line" ? "2" : "3",
      locale: "en",
      toolbar_bg: "#f1f3f6",
      enable_publishing: false,
      allow_symbol_change: true,
      container_id: "tradingview_chart",
      hide_top_toolbar: !showToolbar,
      hide_legend: false,
      save_image: false,
      studies: showVolumeProfile ? ["Volume Profile@tv-volumebyprice"] : [],
      overrides: {
        "paneProperties.background": theme === "dark" ? "#1a1a1a" : "#ffffff",
        "paneProperties.vertGridProperties.color": theme === "dark" ? "#2a2a2a" : "#e1e1e1",
        "paneProperties.horzGridProperties.color": theme === "dark" ? "#2a2a2a" : "#e1e1e1",
        "symbolWatermarkProperties.transparency": 90,
        "scalesProperties.textColor": theme === "dark" ? "#d1d4dc" : "#131722",
        "mainSeriesProperties.candleStyle.upColor": "#10b981",
        "mainSeriesProperties.candleStyle.downColor": "#ef4444",
        "mainSeriesProperties.candleStyle.borderUpColor": "#10b981",
        "mainSeriesProperties.candleStyle.borderDownColor": "#ef4444",
        "mainSeriesProperties.candleStyle.wickUpColor": "#10b981",
        "mainSeriesProperties.candleStyle.wickDownColor": "#ef4444",
      }
    })

    chartContainerRef.current.appendChild(script)

    return () => {
      if (chartContainerRef.current) {
        chartContainerRef.current.innerHTML = ""
      }
    }
  }, [symbol, activeInterval, theme, chartType, showToolbar, showVolumeProfile])

  const handleIntervalChange = (newInterval: string) => {
    setActiveInterval(newInterval)
    onIntervalChange?.(newInterval)
  }

  const handleFullscreen = () => {
    setIsFullscreen(!isFullscreen)
    // In a real implementation, you'd handle fullscreen mode
  }

  const getChartTypeIcon = (type: string) => {
    switch (type) {
      case "candlestick":
        return <Candlestick className="h-4 w-4" />
      case "line":
        return <LineChart className="h-4 w-4" />
      case "area":
        return <BarChart3 className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  return (
    <Card className={cn("w-full", className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center space-x-2">
            <TrendingUp className="h-5 w-5" />
            <span>Price Chart</span>
            <Badge variant="outline">{symbol}</Badge>
          </CardTitle>
          
          <div className="flex items-center space-x-2">
            {/* Chart Type Selector */}
            <div className="flex items-center space-x-1 border rounded-md p-1">
              {["candlestick", "line", "area"].map((type) => (
                <Button
                  key={type}
                  variant={chartType === type ? "default" : "ghost"}
                  size="sm"
                  className="h-7 w-7 p-0"
                  onClick={() => setChartType(type as any)}
                >
                  {getChartTypeIcon(type)}
                </Button>
              ))}
            </div>
            
            <Button variant="outline" size="sm" onClick={handleFullscreen}>
              <Maximize2 className="h-4 w-4" />
            </Button>
            
            <Button variant="outline" size="sm">
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </div>
        
        {/* Time Intervals */}
        <div className="flex items-center space-x-1">
          {intervals.map((int) => (
            <Button
              key={int.value}
              variant={activeInterval === int.value ? "default" : "outline"}
              size="sm"
              className="h-7 text-xs"
              onClick={() => handleIntervalChange(int.value)}
            >
              {int.label}
            </Button>
          ))}
        </div>
      </CardHeader>
      
      <CardContent className="p-0">
        <div 
          ref={chartContainerRef}
          className="tradingview-widget-container"
          style={{ height: `${height}px` }}
        >
          <div 
            id="tradingview_chart"
            className="tradingview-widget-container__widget"
            style={{ height: "100%", width: "100%" }}
          />
          
          {/* Loading State */}
          <div className="flex items-center justify-center h-full bg-muted/10">
            <div className="text-center space-y-2">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
              <p className="text-sm text-muted-foreground">Loading chart...</p>
            </div>
          </div>
        </div>
        
        {/* Chart Footer */}
        <div className="p-3 border-t bg-muted/30">
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <div className="flex items-center space-x-4">
              <span>Powered by TradingView</span>
              <span>•</span>
              <span>Real-time data</span>
            </div>
            <div className="flex items-center space-x-2">
              <Button
                variant="ghost"
                size="sm"
                className="h-6 text-xs"
                onClick={() => setShowVolumeProfile(!showVolumeProfile)}
              >
                Volume Profile: {showVolumeProfile ? "ON" : "OFF"}
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export { TradingViewChart }
