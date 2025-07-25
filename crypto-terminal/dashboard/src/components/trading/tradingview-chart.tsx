'use client'

import * as React from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import {
  Maximize2,
  Settings,
  TrendingUp,
  BarChart3,
  BarChart2,
  LineChart,
  Activity,
  Volume2,
  Zap
} from "lucide-react"

interface TradingViewChartProps {
  symbol?: string
  interval?: string
  theme?: "light" | "dark"
  height?: number
  showToolbar?: boolean
  showVolumeProfile?: boolean
  className?: string
  onIntervalChange?: (interval: string) => void
  onVolumeProfileToggle?: (enabled: boolean) => void
}

const TradingViewChart: React.FC<TradingViewChartProps> = ({
  symbol = "BTCUSD",
  interval = "1D",
  theme = "dark",
  height = 500,
  showToolbar = true,
  showVolumeProfile = false,
  className,
  onIntervalChange,
  onVolumeProfileToggle,
}) => {
  const chartContainerRef = React.useRef<HTMLDivElement>(null)
  const [isFullscreen, setIsFullscreen] = React.useState(false)
  const [chartType, setChartType] = React.useState<"candlestick" | "line" | "area">("candlestick")
  const [activeInterval, setActiveInterval] = React.useState(interval)
  const [isLoading, setIsLoading] = React.useState(true)
  const [volumeProfileEnabled, setVolumeProfileEnabled] = React.useState(showVolumeProfile)

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

    setIsLoading(true)
    
    // Clear previous widget
    chartContainerRef.current.innerHTML = ""

    // Create TradingView widget script
    const script = document.createElement("script")
    script.src = "https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js"
    script.type = "text/javascript"
    script.async = true
    script.onload = () => setIsLoading(false)
    script.innerHTML = JSON.stringify({
      autosize: true,
      symbol: symbol,
      interval: activeInterval,
      timezone: "Etc/UTC",
      theme: theme,
      style: chartType === "candlestick" ? "1" : chartType === "line" ? "2" : "3",
      locale: "en",
      toolbar_bg: theme === "dark" ? "#1a1a1a" : "#f1f3f6",
      enable_publishing: false,
      allow_symbol_change: true,
      container_id: "tradingview_chart",
      hide_top_toolbar: !showToolbar,
      hide_legend: false,
      save_image: false,
      studies: volumeProfileEnabled ? ["Volume Profile@tv-volumebyprice"] : [],
      overrides: {
        "paneProperties.background": theme === "dark" ? "#0a0a0a" : "#ffffff",
        "paneProperties.vertGridProperties.color": theme === "dark" ? "#1a1a1a" : "#e1e1e1",
        "paneProperties.horzGridProperties.color": theme === "dark" ? "#1a1a1a" : "#e1e1e1",
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
  }, [symbol, activeInterval, theme, chartType, showToolbar, volumeProfileEnabled])

  // Sync internal state with prop changes
  React.useEffect(() => {
    setVolumeProfileEnabled(showVolumeProfile)
  }, [showVolumeProfile])

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
        return <BarChart2 className="h-4 w-4" />
      case "line":
        return <LineChart className="h-4 w-4" />
      case "area":
        return <BarChart3 className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <Card variant="glass" className={cn("w-full chart-container", className)}>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <TrendingUp className="h-5 w-5 text-primary" />
              <span>Price Chart</span>
              <Badge variant="glass">{symbol}</Badge>
            </CardTitle>
            
            <div className="flex items-center space-x-2">
              {/* Chart Type Selector */}
              <div className="flex items-center space-x-1 glass rounded-md p-1">
                {["candlestick", "line", "area"].map((type) => (
                  <Button
                    key={type}
                    variant={chartType === type ? "epic" : "ghost"}
                    size="sm"
                    className="h-7 w-7 p-0"
                    onClick={() => setChartType(type as any)}
                  >
                    {getChartTypeIcon(type)}
                  </Button>
                ))}
              </div>
              
              <Button variant="glass" size="sm" onClick={handleFullscreen}>
                <Maximize2 className="h-4 w-4" />
              </Button>
              
              <Button variant="glass" size="sm">
                <Settings className="h-4 w-4" />
              </Button>
            </div>
          </div>
          
          {/* Time Intervals */}
          <div className="flex items-center space-x-1 flex-wrap gap-1">
            {intervals.map((int) => (
              <Button
                key={int.value}
                variant={activeInterval === int.value ? "epic" : "glass"}
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
            className="tradingview-widget-container relative"
            style={{ height: `${height}px` }}
          >
            <div 
              id="tradingview_chart"
              className="tradingview-widget-container__widget"
              style={{ height: "100%", width: "100%" }}
            />
            
            {/* Loading State */}
            {isLoading && (
              <div className="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm">
                <div className="text-center space-y-2">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
                  <p className="text-sm text-muted-foreground">Loading epic chart...</p>
                </div>
              </div>
            )}
          </div>
          
          {/* Chart Footer */}
          <div className="p-3 border-t glass">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <div className="flex items-center space-x-4">
                <span className="flex items-center space-x-1">
                  <Zap className="h-3 w-3" />
                  <span>Powered by TradingView</span>
                </span>
                <span>•</span>
                <span>Real-time data</span>
              </div>
              <div className="flex items-center space-x-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 text-xs"
                  onClick={() => {
                    const newValue = !volumeProfileEnabled
                    setVolumeProfileEnabled(newValue)
                    onVolumeProfileToggle?.(newValue)
                  }}
                >
                  <Volume2 className="h-3 w-3 mr-1" />
                  Volume Profile: {volumeProfileEnabled ? "ON" : "OFF"}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}

export { TradingViewChart }
