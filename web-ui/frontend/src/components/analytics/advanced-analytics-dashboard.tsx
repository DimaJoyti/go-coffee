'use client'

import React, { useState, useEffect, useMemo } from 'react'
import { motion } from 'framer-motion'
import {
  BarChart3,
  TrendingUp,
  DollarSign,
  Coffee,
  AlertTriangle,
  Shield,
  Cpu,
  Database,
  Activity,
  Download,
  Settings,
  Globe,
  Target,
  Clock,
  Bell,
  RefreshCw,
  Maximize,
  ArrowUpRight,
  ArrowDownRight,
  Bitcoin,
  TrendingUp as TrendingUpIcon
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, BarChart, Bar, PieChart as RechartsPieChart, Pie, Cell, AreaChart, Area, ComposedChart } from 'recharts'
import { useWebSocket } from '@/hooks/use-websocket'
import { formatCurrency, formatNumber } from '@/lib/utils'

interface AnalyticsDashboardProps {
  className?: string
}

interface RealtimeData {
  timestamp: string
  active_orders: number
  revenue: number
  orders_per_hour: number
  system_load: {
    cpu: number
    memory: number
    disk: number
    network: number
    healthy: boolean
    uptime: string
  }
  defi_metrics: {
    portfolio_value: number
    daily_pnl: number
    active_positions: number
    arbitrage_opportunities: number
    yield_apy: number
  }
  locations: Array<{
    id: string
    name: string
    orders: number
    revenue: number
    wait_time: number
    satisfaction: number
    status: string
  }>
  alerts_count: number
}

interface KPI {
  title: string
  value: string | number
  change: number
  icon: React.ComponentType<any>
  color: string
  trend: 'up' | 'down' | 'neutral'
}

interface DefiPosition {
  symbol: string
  value: number
  pnl: number
  pnlPercentage: number
}

interface ChartDataPoint {
  time: string
  revenue: number
  orders: number
  cpu: number
  memory: number
  network: number
}

interface LocationData {
  name: string
  revenue: number
  orders: number
  satisfaction: number
  status: string
}

const COLORS = ['#8884d8', '#82ca9d', '#ffc658', '#ff7300', '#8dd1e1', '#d084d0']

export function AdvancedAnalyticsDashboard({ className }: AnalyticsDashboardProps) {
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d' | '30d'>('24h')
  const [selectedTab, setSelectedTab] = useState('overview')
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [refreshInterval, setRefreshInterval] = useState(30)
  const [realtimeData, setRealtimeData] = useState<RealtimeData | null>(null)
  const [businessData, setBusinessData] = useState<any>(null)
  const [defiData, setDefiData] = useState<any>(null)
  const [technicalData, setTechnicalData] = useState<any>(null)
  const [predictiveData, setPredictiveData] = useState<any>(null)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // WebSocket connection for real-time updates
  const { lastMessage, isConnected, connect } = useWebSocket('ws://localhost:8090/api/v1/ws')

  useEffect(() => {
    if (lastMessage?.data) {
      setRealtimeData(lastMessage.data as RealtimeData)
    }
  }, [lastMessage])

  // Connect to WebSocket on mount
  useEffect(() => {
    connect()
  }, [connect])

  // Fetch data from API
  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true)
        setError(null)

        const [businessResponse, defiResponse, technicalResponse, predictiveResponse] = await Promise.all([
          fetch(`/api/v1/business/overview?range=${timeRange}`),
          fetch(`/api/v1/defi/portfolio`),
          fetch(`/api/v1/technical/performance?range=${timeRange}`),
          fetch(`/api/v1/predictions/demand?horizon=${timeRange}`)
        ])

        // Check if all responses are ok
        if (!businessResponse.ok || !defiResponse.ok || !technicalResponse.ok || !predictiveResponse.ok) {
          throw new Error('One or more API requests failed')
        }

        const [business, defi, technical, predictive] = await Promise.all([
          businessResponse.json(),
          defiResponse.json(),
          technicalResponse.json(),
          predictiveResponse.json()
        ])

        setBusinessData(business)
        setDefiData(defi)
        setTechnicalData(technical)
        setPredictiveData(predictive)
      } catch (error) {
        console.error('Failed to fetch analytics data:', error)
        setError(error instanceof Error ? error.message : 'Failed to fetch analytics data')

        // Set mock data for development
        setBusinessData({ revenue: 50000, orders: 1200 })
        setDefiData({
          positions: [
            { token_symbol: 'ETH', value: 25000, pnl: 1250, pnl_percentage: 5.2 },
            { token_symbol: 'BTC', value: 35000, pnl: -850, pnl_percentage: -2.4 },
            { token_symbol: 'USDC', value: 15000, pnl: 0, pnl_percentage: 0 }
          ]
        })
        setTechnicalData({ cpu: 45, memory: 62, disk: 78 })
        setPredictiveData({ demand_forecast: [120, 135, 98, 156] })
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [timeRange])

  // Auto-refresh functionality
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      // Refresh data
      const event = new CustomEvent('refreshAnalytics')
      window.dispatchEvent(event)
    }, refreshInterval * 1000)

    return () => clearInterval(interval)
  }, [autoRefresh, refreshInterval])

  // Generate mock realtime data if not available
  const mockRealtimeData: RealtimeData = {
    timestamp: new Date().toISOString(),
    active_orders: 23,
    revenue: 12450,
    orders_per_hour: 45,
    system_load: {
      cpu: 45,
      memory: 62,
      disk: 78,
      network: 23,
      healthy: true,
      uptime: "99.9%"
    },
    defi_metrics: {
      portfolio_value: 75000,
      daily_pnl: 1250,
      active_positions: 8,
      arbitrage_opportunities: 3,
      yield_apy: 12.5
    },
    locations: [
      { id: '1', name: 'Downtown', orders: 45, revenue: 3200, wait_time: 5, satisfaction: 4.8, status: 'active' },
      { id: '2', name: 'Mall', orders: 32, revenue: 2100, wait_time: 3, satisfaction: 4.6, status: 'active' },
      { id: '3', name: 'Airport', orders: 67, revenue: 4800, wait_time: 8, satisfaction: 4.2, status: 'busy' }
    ],
    alerts_count: 2
  }

  // Calculate KPIs
  const kpis = useMemo<KPI[]>(() => {
    const data = realtimeData || mockRealtimeData
    const business = businessData || { revenue: 50000, orders: 1200 }

    return [
      {
        title: 'Revenue Today',
        value: formatCurrency(data.revenue),
        change: 12.5,
        icon: DollarSign,
        color: 'text-green-500',
        trend: 'up'
      },
      {
        title: 'Active Orders',
        value: data.active_orders,
        change: -2.1,
        icon: Coffee,
        color: 'text-blue-500',
        trend: 'down'
      },
      {
        title: 'Orders/Hour',
        value: data.orders_per_hour,
        change: 8.3,
        icon: Activity,
        color: 'text-purple-500',
        trend: 'up'
      },
      {
        title: 'DeFi Portfolio',
        value: formatCurrency(data.defi_metrics.portfolio_value),
        change: data.defi_metrics.daily_pnl > 0 ? 5.7 : -3.2,
        icon: Bitcoin,
        color: 'text-orange-500',
        trend: data.defi_metrics.daily_pnl > 0 ? 'up' : 'down'
      },
      {
        title: 'System Health',
        value: data.system_load.healthy ? '99.9%' : '95.2%',
        change: 0.1,
        icon: Shield,
        color: data.system_load.healthy ? 'text-green-500' : 'text-red-500',
        trend: data.system_load.healthy ? 'up' : 'down'
      },
      {
        title: 'Alerts',
        value: data.alerts_count,
        change: -15.3,
        icon: AlertTriangle,
        color: data.alerts_count > 5 ? 'text-red-500' : 'text-yellow-500',
        trend: 'down'
      }
    ]
  }, [realtimeData, businessData, mockRealtimeData])

  // Generate sample chart data
  const generateTimeSeriesData = (points: number = 24) => {
    return Array.from({ length: points }, (_, i) => ({
      time: new Date(Date.now() - (points - i) * 60000).toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit' 
      }),
      revenue: 1000 + Math.random() * 2000 + i * 50,
      orders: 20 + Math.random() * 40 + i * 2,
      cpu: 30 + Math.random() * 40,
      memory: 40 + Math.random() * 30,
      network: 10 + Math.random() * 20
    }))
  }

  const timeSeriesData = generateTimeSeriesData()

  const locationData: LocationData[] = (realtimeData || mockRealtimeData).locations.map((loc: any) => ({
    name: loc.name,
    revenue: loc.revenue,
    orders: loc.orders,
    satisfaction: loc.satisfaction,
    status: loc.status
  }))

  const defiPositionsData: DefiPosition[] = defiData?.positions?.slice(0, 6).map((pos: any) => ({
    symbol: pos.token_symbol,
    value: pos.value,
    pnl: pos.pnl,
    pnlPercentage: pos.pnl_percentage
  })) || []

  // Show loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-4" />
          <p className="text-muted-foreground">Loading analytics data...</p>
        </div>
      </div>
    )
  }

  return (
    <motion.div
      className={`space-y-6 ${className} ${isFullscreen ? 'fixed inset-0 z-50 bg-background p-4 overflow-auto' : ''}`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Error Banner */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-red-500" />
            <div>
              <h3 className="font-medium text-red-800">API Connection Issue</h3>
              <p className="text-sm text-red-600 mt-1">
                {error}. Showing mock data for demonstration.
              </p>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => window.location.reload()}
              className="ml-auto"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <BarChart3 className="h-8 w-8" />
            Analytics Dashboard
          </h1>
          <p className="text-muted-foreground mt-1">
            Real-time insights across your entire Go Coffee ecosystem
          </p>
        </div>
        
        <div className="flex items-center gap-3">
          {/* Connection Status */}
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
            <span className="text-sm text-muted-foreground">
              {isConnected ? 'Live' : 'Offline'}
            </span>
          </div>

          {/* Time Range Selector */}
          <Select value={timeRange} onValueChange={(value: any) => setTimeRange(value)}>
            <SelectTrigger className="w-24">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1h">1h</SelectItem>
              <SelectItem value="24h">24h</SelectItem>
              <SelectItem value="7d">7d</SelectItem>
              <SelectItem value="30d">30d</SelectItem>
            </SelectContent>
          </Select>

          {/* Auto Refresh Toggle */}
          <Button
            variant={autoRefresh ? "default" : "outline"}
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto
          </Button>

          {/* Fullscreen Toggle */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsFullscreen(!isFullscreen)}
          >
            <Maximize className="h-4 w-4" />
          </Button>

          {/* Export */}
          <Button variant="outline" size="sm">
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
        {kpis.map((kpi, index) => (
          <motion.div
            key={kpi.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <Card className="relative overflow-hidden">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">
                      {kpi.title}
                    </p>
                    <p className="text-2xl font-bold">
                      {kpi.value}
                    </p>
                  </div>
                  <div className={`${kpi.color}`}>
                    <kpi.icon className="h-6 w-6" />
                  </div>
                </div>
                
                <div className="flex items-center mt-2 text-sm">
                  {kpi.trend === 'up' ? (
                    <ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
                  ) : kpi.trend === 'down' ? (
                    <ArrowDownRight className="h-4 w-4 text-red-500 mr-1" />
                  ) : null}
                  <span className={`${
                    kpi.trend === 'up' ? 'text-green-600' : 
                    kpi.trend === 'down' ? 'text-red-600' : 
                    'text-muted-foreground'
                  }`}>
                    {kpi.change > 0 ? '+' : ''}{kpi.change.toFixed(1)}%
                  </span>
                  <span className="text-muted-foreground ml-1">vs last period</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>

      {/* Main Dashboard Tabs */}
      <Tabs value={selectedTab} onValueChange={setSelectedTab} className="space-y-6">
        <TabsList className="grid w-full grid-cols-3 lg:grid-cols-6">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="business">Business</TabsTrigger>
          <TabsTrigger value="defi">DeFi</TabsTrigger>
          <TabsTrigger value="technical">Technical</TabsTrigger>
          <TabsTrigger value="predictive">AI/ML</TabsTrigger>
          <TabsTrigger value="custom">Custom</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Revenue Chart */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <TrendingUp className="h-5 w-5" />
                  Revenue & Orders Trend
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <ComposedChart data={timeSeriesData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis yAxisId="left" />
                    <YAxis yAxisId="right" orientation="right" />
                    <Tooltip />
                    <Legend />
                    <Area
                      yAxisId="left"
                      type="monotone"
                      dataKey="revenue"
                      fill="#8884d8"
                      stroke="#8884d8"
                      fillOpacity={0.3}
                    />
                    <Bar yAxisId="right" dataKey="orders" fill="#82ca9d" />
                  </ComposedChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* System Health */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-5 w-5" />
                  System Performance
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {Object.entries((realtimeData || mockRealtimeData).system_load)
                    .filter(([key]) => ['cpu', 'memory', 'disk', 'network'].includes(key))
                    .map(([key, value]) => (
                      <div key={key} className="space-y-2">
                        <div className="flex justify-between text-sm">
                          <span className="capitalize">{key}</span>
                          <span>{(value as number).toFixed(1)}%</span>
                        </div>
                        <Progress 
                          value={value as number} 
                          className={`h-2 ${
                            (value as number) > 80 ? 'bg-red-100' : 
                            (value as number) > 60 ? 'bg-yellow-100' : 
                            'bg-green-100'
                          }`} 
                        />
                      </div>
                    ))}
                </div>
              </CardContent>
            </Card>

            {/* Location Performance */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Globe className="h-5 w-5" />
                  Location Performance
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={locationData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="revenue" fill="#8884d8" />
                    <Bar dataKey="orders" fill="#82ca9d" />
                  </BarChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* DeFi Portfolio Overview */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Bitcoin className="h-5 w-5" />
                  DeFi Positions
                </CardTitle>
              </CardHeader>
              <CardContent>
                {defiPositionsData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={300}>
                    <RechartsPieChart>
                      <Pie
                        data={defiPositionsData}
                        cx="50%"
                        cy="50%"
                        outerRadius={80}
                        dataKey="value"
                        label={({ symbol, value }: { symbol: string; value: number }) => `${symbol}: ${formatCurrency(value)}`}
                      >
                        {defiPositionsData.map((entry: any, index: number) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value: any, name: string) => [
                          formatCurrency(value),
                          name
                        ]}
                      />
                      <Legend />
                    </RechartsPieChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                    <div className="text-center">
                      <Bitcoin className="h-12 w-12 mx-auto mb-2 opacity-50" />
                      <p>No DeFi positions available</p>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Recent Activity & Alerts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  Recent Activity
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { time: '2 min ago', event: 'New order #12345 placed', type: 'order' },
                    { time: '5 min ago', event: 'DeFi arbitrage opportunity executed', type: 'defi' },
                    { time: '8 min ago', event: 'AI model updated: demand-forecast', type: 'ai' },
                    { time: '12 min ago', event: 'System backup completed', type: 'system' },
                    { time: '15 min ago', event: 'New customer registered', type: 'customer' }
                  ].map((activity, index) => (
                    <div key={index} className="flex items-center justify-between p-2 rounded-lg bg-muted/50">
                      <div>
                        <p className="text-sm font-medium">{activity.event}</p>
                        <p className="text-xs text-muted-foreground">{activity.time}</p>
                      </div>
                      <Badge variant={
                        activity.type === 'order' ? 'default' :
                        activity.type === 'defi' ? 'secondary' :
                        activity.type === 'ai' ? 'outline' :
                        'destructive'
                      }>
                        {activity.type}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Bell className="h-5 w-5" />
                  Active Alerts
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {(realtimeData || mockRealtimeData).alerts_count > 0 ? [
                    { severity: 'high', message: 'High CPU usage on auth-service', time: '1 min ago' },
                    { severity: 'medium', message: 'Unusual trading volume detected', time: '3 min ago' },
                    { severity: 'low', message: 'SSL certificate expires in 30 days', time: '1 hour ago' }
                  ].map((alert, index) => (
                    <div key={index} className="flex items-center justify-between p-2 rounded-lg border">
                      <div className="flex items-center gap-2">
                        <AlertTriangle className={`h-4 w-4 ${
                          alert.severity === 'high' ? 'text-red-500' :
                          alert.severity === 'medium' ? 'text-yellow-500' :
                          'text-blue-500'
                        }`} />
                        <div>
                          <p className="text-sm font-medium">{alert.message}</p>
                          <p className="text-xs text-muted-foreground">{alert.time}</p>
                        </div>
                      </div>
                      <Badge variant={
                        alert.severity === 'high' ? 'destructive' :
                        alert.severity === 'medium' ? 'secondary' :
                        'outline'
                      }>
                        {alert.severity}
                      </Badge>
                    </div>
                  )) : (
                    <div className="text-center py-8 text-muted-foreground">
                      <Shield className="h-8 w-8 mx-auto mb-2" />
                      <p>All systems operating normally</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Business Tab */}
        <TabsContent value="business" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle>Revenue Analytics</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={400}>
                  <AreaChart data={timeSeriesData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Area type="monotone" dataKey="revenue" stroke="#8884d8" fill="#8884d8" fillOpacity={0.6} />
                  </AreaChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Top Products</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {['Latte', 'Cappuccino', 'Americano', 'Mocha', 'Espresso'].map((product, index) => (
                    <div key={product} className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-coffee-100 rounded-full flex items-center justify-center text-coffee-600 font-bold text-sm">
                          {index + 1}
                        </div>
                        <span className="font-medium">{product}</span>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold">${(Math.random() * 5000 + 1000).toFixed(0)}</div>
                        <div className="text-sm text-muted-foreground">{Math.floor(Math.random() * 200 + 50)} sales</div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* DeFi Tab */}
        <TabsContent value="defi" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Bitcoin className="h-5 w-5" />
                  Portfolio Performance
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm text-muted-foreground">Total Value</p>
                      <p className="text-2xl font-bold">{formatCurrency((realtimeData || mockRealtimeData).defi_metrics.portfolio_value)}</p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Daily P&L</p>
                      <p className={`text-2xl font-bold ${(realtimeData || mockRealtimeData).defi_metrics.daily_pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                        {formatCurrency((realtimeData || mockRealtimeData).defi_metrics.daily_pnl)}
                      </p>
                    </div>
                  </div>
                  <Separator />
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm text-muted-foreground">Active Positions</p>
                      <p className="text-xl font-semibold">{(realtimeData || mockRealtimeData).defi_metrics.active_positions}</p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Yield APY</p>
                      <p className="text-xl font-semibold">{(realtimeData || mockRealtimeData).defi_metrics.yield_apy.toFixed(2)}%</p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Arbitrage Opportunities</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="text-center p-4">
                    <div className="text-3xl font-bold text-green-500">{(realtimeData || mockRealtimeData).defi_metrics.arbitrage_opportunities}</div>
                    <div className="text-sm text-muted-foreground">Active Opportunities</div>
                  </div>
                  <Separator />
                  <div className="space-y-2">
                    {['ETH/USDC', 'BTC/USDT', 'UNI/ETH'].map((pair, index) => (
                      <div key={pair} className="flex items-center justify-between p-2 bg-muted/50 rounded">
                        <span className="font-medium">{pair}</span>
                        <Badge variant="secondary">{(Math.random() * 2 + 0.5).toFixed(2)}%</Badge>
                      </div>
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Technical Tab */}
        <TabsContent value="technical" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Cpu className="h-5 w-5" />
                  System Resources
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={timeSeriesData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line type="monotone" dataKey="cpu" stroke="#8884d8" name="CPU %" />
                    <Line type="monotone" dataKey="memory" stroke="#82ca9d" name="Memory %" />
                    <Line type="monotone" dataKey="network" stroke="#ffc658" name="Network %" />
                  </LineChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Database className="h-5 w-5" />
                  Service Status
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {['API Gateway', 'Auth Service', 'Order Service', 'Kitchen Service', 'Payment Service'].map((service) => (
                    <div key={service} className="flex items-center justify-between p-2 border rounded">
                      <div className="flex items-center gap-2">
                        <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                        <span className="font-medium">{service}</span>
                      </div>
                      <Badge variant="outline">Healthy</Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Predictive/AI Tab */}
        <TabsContent value="predictive" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Target className="h-5 w-5" />
                  Demand Predictions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {['Latte', 'Cappuccino', 'Americano'].map((product) => (
                    <div key={product} className="space-y-2">
                      <div className="flex justify-between">
                        <span className="font-medium">{product}</span>
                        <span className="text-sm text-muted-foreground">
                          {Math.floor(Math.random() * 200 + 100)} predicted orders
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Progress value={Math.random() * 100} className="flex-1" />
                        <Badge variant="outline">{(Math.random() * 0.3 + 0.7).toFixed(2)} confidence</Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <TrendingUpIcon className="h-5 w-5" />
                  Market Predictions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {['BTC', 'ETH', 'USDC'].map((asset) => (
                    <div key={asset} className="flex items-center justify-between p-3 border rounded">
                      <div>
                        <div className="font-medium">{asset}</div>
                        <div className="text-sm text-muted-foreground">
                          ${(Math.random() * 50000 + 1000).toFixed(2)}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className={`font-semibold ${Math.random() > 0.5 ? 'text-green-500' : 'text-red-500'}`}>
                          {Math.random() > 0.5 ? '+' : '-'}{(Math.random() * 10 + 1).toFixed(1)}%
                        </div>
                        <div className="text-sm text-muted-foreground">24h prediction</div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Custom Tab */}
        <TabsContent value="custom" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Custom Dashboard Builder</CardTitle>
              <p className="text-sm text-muted-foreground">
                Create and customize your own dashboard views
              </p>
            </CardHeader>
            <CardContent>
              <div className="text-center py-12">
                <Settings className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                <h3 className="text-lg font-semibold mb-2">Dashboard Builder</h3>
                <p className="text-muted-foreground mb-4">
                  Drag and drop widgets to create your perfect analytics view
                </p>
                <Button>
                  <Settings className="h-4 w-4 mr-2" />
                  Start Building
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </motion.div>
  )
}