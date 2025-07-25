import * as React from "react"
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "./card"
import { Badge } from "./badge"
import { cn, formatCurrency, formatPercentage } from "../../lib/utils"

interface PriceDataPoint {
  timestamp: string | number
  price: number
  volume?: number
  high?: number
  low?: number
  open?: number
  close?: number
}

interface PriceChartProps {
  data: PriceDataPoint[]
  title?: string
  symbol?: string
  currency?: string
  type?: "line" | "area" | "candlestick"
  height?: number
  showVolume?: boolean
  showGrid?: boolean
  showTooltip?: boolean
  color?: string
  gradient?: boolean
  loading?: boolean
  className?: string
  timeframe?: string
  currentPrice?: number
  priceChange?: number
  priceChangePercent?: number
}

const PriceChart: React.FC<PriceChartProps> = ({
  data,
  title,
  symbol,
  currency = "USD",
  type = "area",
  height = 300,
  showVolume = false,
  showGrid = true,
  showTooltip = true,
  color = "#10b981",
  gradient = true,
  loading = false,
  className,
  timeframe,
  currentPrice,
  priceChange,
  priceChangePercent,
}) => {
  const isPositive = priceChange ? priceChange >= 0 : false
  const chartColor = isPositive ? "#10b981" : "#ef4444"

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload
      return (
        <div className="bg-background border border-border rounded-lg p-3 shadow-lg">
          <p className="text-sm text-muted-foreground">
            {new Date(label).toLocaleString()}
          </p>
          <p className="text-sm font-medium">
            Price: {formatCurrency(data.price, currency)}
          </p>
          {data.volume && (
            <p className="text-sm text-muted-foreground">
              Volume: {data.volume.toLocaleString()}
            </p>
          )}
        </div>
      )
    }
    return null
  }

  if (loading) {
    return (
      <Card className={cn("animate-pulse", className)}>
        <CardHeader>
          <div className="h-6 bg-muted rounded w-1/3"></div>
        </CardHeader>
        <CardContent>
          <div className="h-[300px] bg-muted rounded"></div>
        </CardContent>
      </Card>
    )
  }

  const renderChart = () => {
    const commonProps = {
      data,
      margin: { top: 5, right: 30, left: 20, bottom: 5 },
    }

    if (type === "line") {
      return (
        <LineChart {...commonProps}>
          {showGrid && <CartesianGrid strokeDasharray="3 3" className="opacity-30" />}
          <XAxis 
            dataKey="timestamp" 
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => new Date(value).toLocaleDateString()}
          />
          <YAxis 
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => formatCurrency(value, currency, 0)}
          />
          {showTooltip && <Tooltip content={<CustomTooltip />} />}
          <Line
            type="monotone"
            dataKey="price"
            stroke={chartColor}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4, stroke: chartColor, strokeWidth: 2 }}
          />
        </LineChart>
      )
    }

    return (
      <AreaChart {...commonProps}>
        {showGrid && <CartesianGrid strokeDasharray="3 3" className="opacity-30" />}
        <XAxis 
          dataKey="timestamp" 
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          tickFormatter={(value) => new Date(value).toLocaleDateString()}
        />
        <YAxis 
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          tickFormatter={(value) => formatCurrency(value, currency, 0)}
        />
        {showTooltip && <Tooltip content={<CustomTooltip />} />}
        <defs>
          <linearGradient id="colorPrice" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={chartColor} stopOpacity={0.3} />
            <stop offset="95%" stopColor={chartColor} stopOpacity={0} />
          </linearGradient>
        </defs>
        <Area
          type="monotone"
          dataKey="price"
          stroke={chartColor}
          strokeWidth={2}
          fill={gradient ? "url(#colorPrice)" : chartColor}
          fillOpacity={gradient ? 1 : 0.1}
        />
        {currentPrice && (
          <ReferenceLine 
            y={currentPrice} 
            stroke={chartColor} 
            strokeDasharray="5 5" 
            strokeOpacity={0.7}
          />
        )}
      </AreaChart>
    )
  }

  return (
    <Card className={cn("w-full", className)}>
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold">
            {title || `${symbol} Price Chart`}
          </CardTitle>
          <div className="flex items-center space-x-2">
            {timeframe && (
              <Badge variant="outline" className="text-xs">
                {timeframe}
              </Badge>
            )}
            {priceChangePercent !== undefined && (
              <Badge 
                variant={isPositive ? "bull" : "bear"}
                className="text-xs"
              >
                {isPositive ? "+" : ""}{formatPercentage(priceChangePercent / 100)}
              </Badge>
            )}
          </div>
        </div>
        {(currentPrice || priceChange) && (
          <div className="flex items-center space-x-4">
            {currentPrice && (
              <span className="text-2xl font-bold">
                {formatCurrency(currentPrice, currency)}
              </span>
            )}
            {priceChange && (
              <span className={cn(
                "text-sm font-medium",
                isPositive ? "text-bull" : "text-bear"
              )}>
                {isPositive ? "+" : ""}{formatCurrency(priceChange, currency)}
              </span>
            )}
          </div>
        )}
      </CardHeader>
      <CardContent className="pt-0">
        <div style={{ height }}>
          <ResponsiveContainer width="100%" height="100%">
            {renderChart()}
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}

export { PriceChart, type PriceDataPoint }
