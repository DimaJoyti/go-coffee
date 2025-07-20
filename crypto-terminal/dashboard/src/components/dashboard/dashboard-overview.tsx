'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useTradingStore } from '@/stores/trading-store'
import { analyticsApi, marketApi } from '@/lib/api'
import { formatCurrency, formatPercent, getPriceChangeColor } from '@/lib/utils'
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Activity,
  Coffee,
  AlertTriangle,
} from 'lucide-react'

export function DashboardOverview() {
  const { portfolios, strategies, signalAlerts, riskAlerts } = useTradingStore()

  // Fetch dashboard analytics
  const { data: dashboardData, isLoading } = useQuery({
    queryKey: ['dashboard-analytics'],
    queryFn: analyticsApi.getDashboard,
    refetchInterval: 30000, // Refetch every 30 seconds
  })

  // Fetch market overview
  const { data: marketOverview } = useQuery({
    queryKey: ['market-overview'],
    queryFn: marketApi.getMarketOverview,
    refetchInterval: 60000, // Refetch every minute
  })

  // Calculate portfolio totals
  const totalPortfolioValue = portfolios.reduce((sum, p) => sum + p.totalValue, 0)
  const totalPortfolioPnL = portfolios.reduce((sum, p) => sum + p.totalPnL, 0)
  const totalDayChange = portfolios.reduce((sum, p) => sum + p.dayChange, 0)

  // Calculate strategy stats
  const activeStrategies = strategies.filter(s => s.status === 'active').length
  const totalTrades = strategies.reduce((sum, s) => sum + (s.performance?.totalTrades || 0), 0)
  const avgWinRate = strategies.length > 0 
    ? strategies.reduce((sum, s) => sum + (s.performance?.winRate || 0), 0) / strategies.length
    : 0

  // Recent alerts
  const recentSignals = signalAlerts.slice(0, 3)
  const criticalRiskAlerts = riskAlerts.filter(alert => alert.severity === 'critical' || alert.severity === 'high')

  const overviewCards = [
    {
      title: 'Total Portfolio Value',
      value: formatCurrency(totalPortfolioValue),
      change: totalPortfolioPnL,
      changePercent: totalPortfolioValue > 0 ? (totalPortfolioPnL / (totalPortfolioValue - totalPortfolioPnL)) * 100 : 0,
      icon: DollarSign,
      color: totalPortfolioPnL >= 0 ? 'profit' : 'loss',
    },
    {
      title: '24h Change',
      value: formatCurrency(totalDayChange),
      change: totalDayChange,
      changePercent: totalPortfolioValue > 0 ? (totalDayChange / totalPortfolioValue) * 100 : 0,
      icon: totalDayChange >= 0 ? TrendingUp : TrendingDown,
      color: totalDayChange >= 0 ? 'profit' : 'loss',
    },
    {
      title: 'Active Strategies',
      value: activeStrategies.toString(),
      subtitle: `${strategies.length} total`,
      icon: Coffee,
      color: 'coffee',
    },
    {
      title: 'Win Rate',
      value: `${avgWinRate.toFixed(1)}%`,
      subtitle: `${totalTrades} trades`,
      icon: Activity,
      color: avgWinRate > 60 ? 'profit' : avgWinRate > 40 ? 'warning' : 'loss',
    },
  ]

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[...Array(4)].map((_, i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="pb-2">
              <div className="h-4 bg-muted rounded w-3/4"></div>
            </CardHeader>
            <CardContent>
              <div className="h-8 bg-muted rounded w-1/2 mb-2"></div>
              <div className="h-3 bg-muted rounded w-1/3"></div>
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {overviewCards.map((card, index) => (
          <Card key={index} className="hover:shadow-md transition-shadow">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {card.title}
              </CardTitle>
              <card.icon className={`h-4 w-4 ${getPriceChangeColor(card.change || 0)}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{card.value}</div>
              {card.change !== undefined && (
                <div className="flex items-center space-x-2 text-xs text-muted-foreground">
                  <span className={getPriceChangeColor(card.change)}>
                    {card.change >= 0 ? '+' : ''}{formatPercent(card.changePercent || 0)}
                  </span>
                  <span>from yesterday</span>
                </div>
              )}
              {card.subtitle && (
                <p className="text-xs text-muted-foreground mt-1">
                  {card.subtitle}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Alerts and Signals */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Signals */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Activity className="h-5 w-5" />
              <span>Recent Signals</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {recentSignals.length > 0 ? (
              <div className="space-y-3">
                {recentSignals.map((signal, index) => (
                  <div key={index} className="flex items-center justify-between p-3 rounded-lg bg-muted/50">
                    <div className="flex items-center space-x-3">
                      <span className="text-lg">{signal.emoji}</span>
                      <div>
                        <div className="font-medium">{signal.strategy}</div>
                        <div className="text-sm text-muted-foreground">
                          {signal.symbol} • {signal.signal}
                        </div>
                      </div>
                    </div>
                    <Badge
                      variant={
                        signal.confidence > 0.8 ? 'profit' :
                        signal.confidence > 0.6 ? 'warning' : 'secondary'
                      }
                    >
                      {(signal.confidence * 100).toFixed(0)}%
                    </Badge>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                <Activity className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No recent signals</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Risk Alerts */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <AlertTriangle className="h-5 w-5" />
              <span>Risk Alerts</span>
              {criticalRiskAlerts.length > 0 && (
                <Badge variant="destructive">{criticalRiskAlerts.length}</Badge>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {criticalRiskAlerts.length > 0 ? (
              <div className="space-y-3">
                {criticalRiskAlerts.slice(0, 3).map((alert, index) => (
                  <div key={index} className="flex items-start space-x-3 p-3 rounded-lg bg-destructive/10 border border-destructive/20">
                    <AlertTriangle className="h-4 w-4 text-destructive mt-0.5" />
                    <div className="flex-1">
                      <div className="font-medium text-destructive">{alert.type.replace('_', ' ').toUpperCase()}</div>
                      <div className="text-sm text-muted-foreground">{alert.message}</div>
                      {alert.recommendation && (
                        <div className="text-xs text-muted-foreground mt-1">
                          💡 {alert.recommendation}
                        </div>
                      )}
                    </div>
                    <Badge variant="destructive" className="text-xs">
                      {alert.severity}
                    </Badge>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                <AlertTriangle className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No critical alerts</p>
                <p className="text-xs">Your portfolio is looking good! ☕</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
