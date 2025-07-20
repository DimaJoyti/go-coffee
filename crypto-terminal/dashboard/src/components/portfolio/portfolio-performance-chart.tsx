'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useTradingStore } from '@/stores/trading-store'
import { portfolioApi } from '@/lib/api'
import { formatCurrency, formatPercent } from '@/lib/utils'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
} from 'recharts'
import {
  TrendingUp,
  Calendar,
  BarChart3,
  Activity,
  Target,
} from 'lucide-react'

const timeRanges = [
  { id: '1D', label: '1 Day', value: '1D' },
  { id: '7D', label: '7 Days', value: '7D' },
  { id: '30D', label: '30 Days', value: '30D' },
  { id: '90D', label: '90 Days', value: '90D' },
  { id: '1Y', label: '1 Year', value: '1Y' },
  { id: 'ALL', label: 'All Time', value: 'ALL' },
]

export function PortfolioPerformanceChart() {
  const { selectedPortfolioId } = useTradingStore()
  const [selectedTimeRange, setSelectedTimeRange] = useState('30D')
  const [chartType, setChartType] = useState<'line' | 'area' | 'bar'>('area')

  // Fetch portfolio performance data
  const { data: performanceData, isLoading } = useQuery({
    queryKey: ['portfolio-performance', selectedPortfolioId, selectedTimeRange],
    queryFn: () => portfolioApi.getPortfolioPerformance(selectedPortfolioId!, selectedTimeRange),
    enabled: !!selectedPortfolioId,
    refetchInterval: 60000,
  })

  // Generate mock data for demonstration
  const generateMockData = () => {
    const days = selectedTimeRange === '1D' ? 24 : 
                 selectedTimeRange === '7D' ? 7 :
                 selectedTimeRange === '30D' ? 30 :
                 selectedTimeRange === '90D' ? 90 : 365

    const data = []
    let baseValue = 45000
    
    for (let i = 0; i < days; i++) {
      const change = (Math.random() - 0.5) * 1000
      baseValue += change
      
      data.push({
        date: new Date(Date.now() - (days - i) * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        value: Math.max(baseValue, 1000),
        pnl: baseValue - 45000,
        pnlPercent: ((baseValue - 45000) / 45000) * 100,
        volume: Math.random() * 10000 + 5000,
      })
    }
    
    return data
  }

  const chartData = performanceData?.chartData || generateMockData()
  const metrics = performanceData || {
    startValue: 45000,
    endValue: 50000,
    totalReturn: 5000,
    totalReturnPercent: 11.11,
    annualizedReturn: 22.22,
    volatility: 0.35,
    sharpeRatio: 1.25,
    maxDrawdown: -3500,
    maxDrawdownPercent: -7.78,
    bestDay: 1250,
    worstDay: -890,
    winningDays: 18,
    totalDays: 30,
  }

  if (isLoading) {
    return (
      <Card className="animate-pulse">
        <CardHeader>
          <div className="h-6 bg-muted rounded w-1/3"></div>
        </CardHeader>
        <CardContent>
          <div className="h-64 bg-muted rounded"></div>
        </CardContent>
      </Card>
    )
  }

  const renderChart = () => {
    const commonProps = {
      data: chartData,
      margin: { top: 5, right: 30, left: 20, bottom: 5 },
    }

    switch (chartType) {
      case 'line':
        return (
          <LineChart {...commonProps}>
            <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip
              formatter={(value: number) => [formatCurrency(value), 'Portfolio Value']}
              labelFormatter={(label) => `Date: ${label}`}
            />
            <Line
              type="monotone"
              dataKey="value"
              stroke="#10b981"
              strokeWidth={2}
              dot={false}
            />
          </LineChart>
        )
      case 'area':
        return (
          <AreaChart {...commonProps}>
            <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip
              formatter={(value: number) => [formatCurrency(value), 'Portfolio Value']}
              labelFormatter={(label) => `Date: ${label}`}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke="#10b981"
              fill="#10b981"
              fillOpacity={0.3}
            />
          </AreaChart>
        )
      case 'bar':
        return (
          <BarChart {...commonProps}>
            <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip
              formatter={(value: number) => [formatCurrency(value), 'Daily P&L']}
              labelFormatter={(label) => `Date: ${label}`}
            />
            <Bar
              dataKey="pnl"
              fill={(entry: any) => entry > 0 ? '#10b981' : '#ef4444'}
            />
          </BarChart>
        )
      default:
        return null
    }
  }

  return (
    <div className="space-y-6">
      {/* Performance Chart */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <TrendingUp className="h-5 w-5" />
              <span>Portfolio Performance</span>
            </CardTitle>
            <div className="flex items-center space-x-2">
              {/* Chart Type Selector */}
              <div className="flex items-center space-x-1">
                <Button
                  variant={chartType === 'area' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setChartType('area')}
                >
                  Area
                </Button>
                <Button
                  variant={chartType === 'line' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setChartType('line')}
                >
                  Line
                </Button>
                <Button
                  variant={chartType === 'bar' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setChartType('bar')}
                >
                  Bar
                </Button>
              </div>
              
              {/* Time Range Selector */}
              <div className="flex items-center space-x-1">
                {timeRanges.map((range) => (
                  <Button
                    key={range.id}
                    variant={selectedTimeRange === range.value ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSelectedTimeRange(range.value)}
                  >
                    {range.label}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              {renderChart()}
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Performance Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Return
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-profit">
              {formatCurrency(metrics.totalReturn)}
            </div>
            <div className="text-sm text-muted-foreground">
              {formatPercent(metrics.totalReturnPercent)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Annualized Return
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatPercent(metrics.annualizedReturn)}
            </div>
            <div className="text-sm text-muted-foreground">
              vs {selectedTimeRange} period
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Sharpe Ratio
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metrics.sharpeRatio.toFixed(2)}
            </div>
            <div className="text-sm text-muted-foreground">
              Risk-adjusted return
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Max Drawdown
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-loss">
              {formatCurrency(metrics.maxDrawdown)}
            </div>
            <div className="text-sm text-muted-foreground">
              {formatPercent(metrics.maxDrawdownPercent)}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Additional Metrics */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <BarChart3 className="h-5 w-5" />
            <span>Performance Statistics</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="text-center">
              <div className="text-lg font-semibold text-profit">
                {formatCurrency(metrics.bestDay)}
              </div>
              <div className="text-sm text-muted-foreground">Best Day</div>
            </div>
            
            <div className="text-center">
              <div className="text-lg font-semibold text-loss">
                {formatCurrency(metrics.worstDay)}
              </div>
              <div className="text-sm text-muted-foreground">Worst Day</div>
            </div>
            
            <div className="text-center">
              <div className="text-lg font-semibold">
                {((metrics.winningDays / metrics.totalDays) * 100).toFixed(1)}%
              </div>
              <div className="text-sm text-muted-foreground">
                Win Rate ({metrics.winningDays}/{metrics.totalDays})
              </div>
            </div>
            
            <div className="text-center">
              <div className="text-lg font-semibold">
                {(metrics.volatility * 100).toFixed(1)}%
              </div>
              <div className="text-sm text-muted-foreground">Volatility</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
