'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useTradingStore } from '@/stores/trading-store'
import { portfolioApi } from '@/lib/api'
import { formatCurrency, formatPercent, getPriceChangeColor } from '@/lib/utils'
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip } from 'recharts'
import {
  Wallet,
  RefreshCw,
  Eye,
  MoreHorizontal,
} from 'lucide-react'

const COLORS = ['#10b981', '#f59e0b', '#ef4444', '#3b82f6', '#8b5cf6', '#f97316']

export function PortfolioSummary() {
  const { portfolios, selectedPortfolioId } = useTradingStore()

  // Fetch portfolios
  const { isLoading, refetch } = useQuery({
    queryKey: ['portfolios'],
    queryFn: portfolioApi.getPortfolios,
    refetchInterval: 30000,
  })

  const selectedPortfolio = portfolios.find(p => p.id === selectedPortfolioId) || portfolios[0]

  // Prepare chart data
  const allocationData = selectedPortfolio?.holdings?.map((holding, index) => ({
    name: holding.symbol,
    value: holding.allocation,
    amount: holding.totalValue,
    color: COLORS[index % COLORS.length],
  })) || []

  const performanceData = selectedPortfolio?.holdings?.map(holding => ({
    symbol: holding.symbol,
    pnl: holding.pnl,
    pnlPercent: holding.pnlPercent,
  })) || []

  if (isLoading) {
    return (
      <Card className="animate-pulse">
        <CardHeader>
          <div className="h-6 bg-muted rounded w-1/3"></div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="h-4 bg-muted rounded w-1/2"></div>
            <div className="h-32 bg-muted rounded"></div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (!selectedPortfolio) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="text-center">
            <Wallet className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
            <h3 className="text-lg font-medium mb-2">No Portfolio Found</h3>
            <p className="text-muted-foreground mb-4">
              Create your first portfolio to start trading
            </p>
            <Button>Create Portfolio</Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Portfolio Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Wallet className="h-5 w-5" />
              <span>{selectedPortfolio.name}</span>
            </CardTitle>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" onClick={() => refetch()}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
              <Button variant="outline" size="sm">
                <Eye className="h-4 w-4 mr-2" />
                View Details
              </Button>
              <Button variant="ghost" size="sm">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            {/* Total Value */}
            <div className="text-center">
              <div className="text-2xl font-bold">
                {formatCurrency(selectedPortfolio.totalValue)}
              </div>
              <div className="text-sm text-muted-foreground">Total Value</div>
            </div>

            {/* Total P&L */}
            <div className="text-center">
              <div className={`text-2xl font-bold ${getPriceChangeColor(selectedPortfolio.totalPnL)}`}>
                {selectedPortfolio.totalPnL >= 0 ? '+' : ''}
                {formatCurrency(selectedPortfolio.totalPnL)}
              </div>
              <div className="text-sm text-muted-foreground">
                Total P&L ({formatPercent(selectedPortfolio.totalPnLPercent)})
              </div>
            </div>

            {/* 24h Change */}
            <div className="text-center">
              <div className={`text-2xl font-bold ${getPriceChangeColor(selectedPortfolio.dayChange)}`}>
                {selectedPortfolio.dayChange >= 0 ? '+' : ''}
                {formatCurrency(selectedPortfolio.dayChange)}
              </div>
              <div className="text-sm text-muted-foreground">
                24h Change ({formatPercent(selectedPortfolio.dayChangePercent)})
              </div>
            </div>

            {/* Holdings Count */}
            <div className="text-center">
              <div className="text-2xl font-bold">
                {selectedPortfolio.holdings?.length || 0}
              </div>
              <div className="text-sm text-muted-foreground">Holdings</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Allocation Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Portfolio Allocation</CardTitle>
          </CardHeader>
          <CardContent>
            {allocationData.length > 0 ? (
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={allocationData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={100}
                      paddingAngle={2}
                      dataKey="value"
                    >
                      {allocationData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value: number, name: string, props: any) => [
                        `${value.toFixed(1)}%`,
                        formatCurrency(props.payload.amount)
                      ]}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <div className="h-64 flex items-center justify-center text-muted-foreground">
                No holdings to display
              </div>
            )}
          </CardContent>
        </Card>

        {/* Performance Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Holdings Performance</CardTitle>
          </CardHeader>
          <CardContent>
            {performanceData.length > 0 ? (
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={performanceData}>
                    <XAxis dataKey="symbol" />
                    <YAxis />
                    <Tooltip
                      formatter={(value: number) => [
                        formatPercent(value),
                        'P&L %'
                      ]}
                    />
                    <Bar
                      dataKey="pnlPercent"
                      fill="#10b981"
                    />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <div className="h-64 flex items-center justify-center text-muted-foreground">
                No performance data to display
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Holdings Table */}
      <Card>
        <CardHeader>
          <CardTitle>Holdings</CardTitle>
        </CardHeader>
        <CardContent>
          {selectedPortfolio.holdings && selectedPortfolio.holdings.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2">Asset</th>
                    <th className="text-right py-2">Amount</th>
                    <th className="text-right py-2">Price</th>
                    <th className="text-right py-2">Value</th>
                    <th className="text-right py-2">P&L</th>
                    <th className="text-right py-2">Allocation</th>
                  </tr>
                </thead>
                <tbody>
                  {selectedPortfolio.holdings.map((holding) => (
                    <tr key={holding.id} className="border-b hover:bg-muted/50">
                      <td className="py-3">
                        <div>
                          <div className="font-medium">{holding.symbol}</div>
                          <div className="text-sm text-muted-foreground">{holding.name}</div>
                        </div>
                      </td>
                      <td className="text-right py-3">
                        {holding.amount.toFixed(6)}
                      </td>
                      <td className="text-right py-3">
                        {formatCurrency(holding.currentPrice)}
                      </td>
                      <td className="text-right py-3">
                        {formatCurrency(holding.totalValue)}
                      </td>
                      <td className="text-right py-3">
                        <div className={getPriceChangeColor(holding.pnl)}>
                          {holding.pnl >= 0 ? '+' : ''}{formatCurrency(holding.pnl)}
                          <div className="text-xs">
                            ({formatPercent(holding.pnlPercent)})
                          </div>
                        </div>
                      </td>
                      <td className="text-right py-3">
                        <Badge variant="outline">
                          {holding.allocation.toFixed(1)}%
                        </Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Wallet className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No holdings in this portfolio</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
