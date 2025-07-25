'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useTradingStore } from '@/stores/trading-store'
import { portfolioApi } from '@/lib/api'
import { formatCurrency, formatPercent } from '@/lib/utils'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar,
} from 'recharts'
import {
  Shield,
  AlertTriangle,
  TrendingDown,
  Target,
  Activity,
  Zap,
} from 'lucide-react'

export function PortfolioRiskMetrics() {
  const { selectedPortfolioId, riskAlerts } = useTradingStore()

  // Fetch portfolio risk metrics
  const { data: riskData, isLoading } = useQuery({
    queryKey: ['portfolio-risk-metrics', selectedPortfolioId],
    queryFn: () => portfolioApi.getPortfolioRiskMetrics(selectedPortfolioId!),
    enabled: !!selectedPortfolioId,
    refetchInterval: 60000,
  })

  // Mock risk data for demonstration
  const mockRiskData = {
    overallRiskScore: 6.5,
    riskLevel: 'medium',
    valueAtRisk: {
      var95: -2500,
      var99: -4200,
      expectedShortfall: -5100,
    },
    diversification: {
      score: 7.2,
      concentrationRisk: 0.35,
      correlationRisk: 0.42,
    },
    volatilityMetrics: {
      portfolioVolatility: 0.28,
      averageVolatility: 0.31,
      volatilityContribution: [
        { asset: 'BTC', contribution: 0.45 },
        { asset: 'ETH', contribution: 0.32 },
        { asset: 'ADA', contribution: 0.12 },
        { asset: 'SOL', contribution: 0.11 },
      ],
    },
    riskFactors: [
      { factor: 'Market Risk', score: 7, maxScore: 10 },
      { factor: 'Liquidity Risk', score: 3, maxScore: 10 },
      { factor: 'Concentration Risk', score: 6, maxScore: 10 },
      { factor: 'Correlation Risk', score: 5, maxScore: 10 },
      { factor: 'Volatility Risk', score: 8, maxScore: 10 },
    ],
    stressTests: [
      { scenario: 'Market Crash (-30%)', impact: -15000, impactPercent: -30 },
      { scenario: 'Crypto Winter (-50%)', impact: -25000, impactPercent: -50 },
      { scenario: 'Flash Crash (-20%)', impact: -10000, impactPercent: -20 },
      { scenario: 'Regulatory Ban (-40%)', impact: -20000, impactPercent: -40 },
    ],
    recommendations: [
      {
        type: 'diversification',
        severity: 'medium',
        message: 'Consider reducing BTC concentration below 40%',
        action: 'Rebalance portfolio allocation',
      },
      {
        type: 'volatility',
        severity: 'high',
        message: 'Portfolio volatility is above target range',
        action: 'Add stable assets or reduce position sizes',
      },
      {
        type: 'correlation',
        severity: 'low',
        message: 'Asset correlations are within acceptable range',
        action: 'Monitor correlation changes during market stress',
      },
    ],
  }

  const metrics = riskData || mockRiskData

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

  const getRiskBadgeVariant = (level: string) => {
    switch (level) {
      case 'low': return 'profit'
      case 'medium': return 'warning'
      case 'high': return 'loss'
      default: return 'secondary'
    }
  }

  return (
    <div className="space-y-6">
      {/* Risk Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <Shield className="h-4 w-4" />
              <span>Risk Score</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {metrics.overallRiskScore.toFixed(1)}/10
            </div>
            <Badge variant={getRiskBadgeVariant(metrics.riskLevel)} className="mt-2">
              {metrics.riskLevel.toUpperCase()} RISK
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <TrendingDown className="h-4 w-4" />
              <span>VaR (95%)</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-loss">
              {formatCurrency(metrics.valueAtRisk.var95)}
            </div>
            <div className="text-sm text-muted-foreground">
              1-day potential loss
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <Target className="h-4 w-4" />
              <span>Diversification</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metrics.diversification.score.toFixed(1)}/10
            </div>
            <div className="text-sm text-muted-foreground">
              Portfolio spread
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <Activity className="h-4 w-4" />
              <span>Volatility</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatPercent(metrics.volatilityMetrics.portfolioVolatility)}
            </div>
            <div className="text-sm text-muted-foreground">
              Annualized volatility
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Risk Factor Radar Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Zap className="h-5 w-5" />
            <span>Risk Factor Analysis</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <RadarChart data={metrics.riskFactors}>
                <PolarGrid />
                <PolarAngleAxis dataKey="factor" />
                <PolarRadiusAxis angle={90} domain={[0, 10]} />
                <Radar
                  name="Risk Score"
                  dataKey="score"
                  stroke="#ef4444"
                  fill="#ef4444"
                  fillOpacity={0.3}
                />
              </RadarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Volatility Contribution */}
      <Card>
        <CardHeader>
          <CardTitle>Volatility Contribution by Asset</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={metrics.volatilityMetrics.volatilityContribution}>
                <XAxis dataKey="asset" />
                <YAxis />
                <Tooltip
                  formatter={(value: number) => [
                    `${(value * 100).toFixed(1)}%`,
                    'Volatility Contribution'
                  ]}
                />
                <Bar dataKey="contribution" fill="#f59e0b" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Stress Tests */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <AlertTriangle className="h-5 w-5" />
            <span>Stress Test Scenarios</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {metrics.stressTests.map((test: any, index: number) => (
              <div key={index} className="flex items-center justify-between p-4 rounded-lg bg-muted/50">
                <div>
                  <div className="font-medium">{test.scenario}</div>
                  <div className="text-sm text-muted-foreground">
                    Potential Impact: {formatPercent(test.impactPercent)}
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-lg font-semibold text-loss">
                    {formatCurrency(test.impact)}
                  </div>
                  <Badge variant="loss" className="text-xs">
                    {formatPercent(test.impactPercent)}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Risk Recommendations */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Shield className="h-5 w-5" />
            <span>Risk Management Recommendations</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {metrics.recommendations.map((rec: any, index: number) => (
              <div
                key={index}
                className={`p-4 rounded-lg border-l-4 ${
                  rec.severity === 'high' ? 'border-l-red-500 bg-red-50 dark:bg-red-950/20' :
                  rec.severity === 'medium' ? 'border-l-yellow-500 bg-yellow-50 dark:bg-yellow-950/20' :
                  'border-l-green-500 bg-green-50 dark:bg-green-950/20'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-2">
                      <Badge
                        variant={
                          rec.severity === 'high' ? 'destructive' :
                          rec.severity === 'medium' ? 'warning' : 'profit'
                        }
                        className="text-xs"
                      >
                        {rec.severity.toUpperCase()}
                      </Badge>
                      <span className="text-sm font-medium capitalize">
                        {rec.type.replace('_', ' ')} Risk
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground mb-2">
                      {rec.message}
                    </p>
                    <p className="text-sm font-medium">
                      💡 {rec.action}
                    </p>
                  </div>
                  <Button variant="outline" size="sm" className="ml-4">
                    Apply
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Active Risk Alerts */}
      {riskAlerts.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <AlertTriangle className="h-5 w-5 text-destructive" />
              <span>Active Risk Alerts</span>
              <Badge variant="destructive">{riskAlerts.length}</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {riskAlerts.slice(0, 5).map((alert, index) => (
                <div key={index} className="flex items-start space-x-3 p-3 rounded-lg bg-destructive/10 border border-destructive/20">
                  <AlertTriangle className="h-4 w-4 text-destructive mt-0.5" />
                  <div className="flex-1">
                    <div className="font-medium text-destructive">
                      {alert.type.replace('_', ' ').toUpperCase()}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {alert.message}
                    </div>
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
          </CardContent>
        </Card>
      )}
    </div>
  )
}
