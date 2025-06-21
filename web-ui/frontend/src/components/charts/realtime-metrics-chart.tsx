'use client'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Activity, TrendingUp, Zap } from 'lucide-react'
import React, { useEffect, useState } from 'react'
import { Area, AreaChart, CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { useWebSocket } from '../../hooks/use-websocket'

interface RealtimeMetricsChartProps {
  data?: Array<{
    time: string
    revenue: number
    orders: number
    cpu: number
    memory: number
    network: number
  }>
  type?: 'line' | 'area'
  height?: number
  metrics?: string[]
  showLegend?: boolean
  className?: string
}

interface MetricsData {
  time: string
  revenue: number
  orders: number
  cpu: number
  memory: number
  network: number
}

export function RealtimeMetricsChart({
  data = [],
  type = 'line',
  height = 300,
  metrics = ['revenue', 'orders'],
  showLegend = true,
  className
}: RealtimeMetricsChartProps) {
  const [animationKey, setAnimationKey] = useState<number>(0)
  const [metricsData, setMetricsData] = useState<MetricsData[]>(data)
  const { lastMessage, connect } = useWebSocket('ws://localhost:8090/api/v1/ws')

  useEffect(() => {
    setAnimationKey((prev: number) => prev + 1)
  }, [data])

  useEffect(() => {
    if (data.length > 0) {
      setMetricsData(data)
    }
  }, [data])

  useEffect(() => {
    connect()
  }, [connect])

  useEffect(() => {
    if (lastMessage?.data) {
      const newMetric: MetricsData = {
        time: new Date().toLocaleTimeString(),
        cpu: lastMessage.data.cpu || Math.random() * 100,
        memory: lastMessage.data.memory || Math.random() * 100,
        revenue: lastMessage.data.revenue || Math.random() * 1000,
        orders: lastMessage.data.orders || Math.floor(Math.random() * 50),
        network: lastMessage.data.network || Math.random() * 100,
      }

      setMetricsData((prev: MetricsData[]) => {
        const updated = [...prev, newMetric]
        return updated.slice(-20)
      })
    }
  }, [lastMessage])

  const formatTooltipValue = (value: number, name: string): [string, string] => {
    switch (name) {
      case 'revenue':
        return [`$${value.toFixed(2)}`, 'Revenue']
      case 'orders':
        return [value.toString(), 'Orders']
      case 'cpu':
      case 'memory':
      case 'network':
        return [`${value.toFixed(1)}%`, name.toUpperCase()]
      default:
        return [value.toString(), name]
    }
  }

  const getMetricColor = (metric: string): string => {
    const colors = {
      revenue: '#8884d8',
      orders: '#82ca9d',
      cpu: '#ffc658',
      memory: '#ff7300',
      network: '#8dd1e1'
    }
    return colors[metric as keyof typeof colors] || '#8884d8'
  }

  const getMetricIcon = (metric: string): React.ReactElement => {
    switch (metric) {
      case 'revenue':
        return <TrendingUp className="h-4 w-4" />
      case 'orders':
        return <Activity className="h-4 w-4" />
      case 'cpu':
      case 'memory':
      case 'network':
        return <Zap className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  const ChartComponent = type === 'area' ? AreaChart : LineChart

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Real-time Metrics</CardTitle>
          <div className="flex gap-1">
            {metrics.map((metric) => (
              <Badge
                key={metric}
                className="flex items-center gap-1 border"
                style={{ borderColor: getMetricColor(metric) }}
              >
                {getMetricIcon(metric)}
                {metric}
              </Badge>
            ))}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={height}>
          <ChartComponent key={animationKey} data={metricsData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
            <XAxis 
              dataKey="time" 
              stroke="#666"
              fontSize={12}
              tickFormatter={(value) => value.split(' ')[1] || value}
            />
            <YAxis stroke="#666" fontSize={12} />
            <Tooltip
              contentStyle={{
                backgroundColor: 'rgba(255, 255, 255, 0.95)',
                border: '1px solid #e2e8f0',
                borderRadius: '8px',
                boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)'
              }}
              formatter={formatTooltipValue}
              labelStyle={{ color: '#374151' }}
            />
            {showLegend && <Legend />}
            
            {metrics.map((metric) => {
              const color = getMetricColor(metric)
              
              if (type === 'area') {
                return (
                  <Area
                    key={metric}
                    type="monotone"
                    dataKey={metric}
                    stroke={color}
                    fill={color}
                    fillOpacity={0.3}
                    strokeWidth={2}
                    dot={{ r: 0 }}
                    activeDot={{ r: 4, stroke: color, strokeWidth: 2 }}
                  />
                )
              } else {
                return (
                  <Line
                    key={metric}
                    type="monotone"
                    dataKey={metric}
                    stroke={color}
                    strokeWidth={2}
                    dot={{ r: 0 }}
                    activeDot={{ r: 4, stroke: color, strokeWidth: 2 }}
                  />
                )
              }
            })}
          </ChartComponent>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}