'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PortfolioSummary } from './portfolio-summary'
import { PortfolioPerformanceChart } from './portfolio-performance-chart'
import { PortfolioRiskMetrics } from './portfolio-risk-metrics'
import { useTradingStore } from '@/stores/trading-store'
import { portfolioApi } from '@/lib/api'
import { formatCurrency, formatPercent } from '@/lib/utils'
import {
  Wallet,
  Plus,
  Settings,
  Download,
  Upload,
  BarChart3,
  PieChart,
  TrendingUp,
  Shield,
} from 'lucide-react'

export function PortfolioManager() {
  const { portfolios, selectedPortfolioId, setSelectedPortfolioId } = useTradingStore()
  const [activeTab, setActiveTab] = useState<'overview' | 'performance' | 'risk' | 'settings'>('overview')

  // Fetch portfolios
  const { data: portfolioData, isLoading } = useQuery({
    queryKey: ['portfolios'],
    queryFn: portfolioApi.getPortfolios,
    refetchInterval: 30000,
  })

  const selectedPortfolio = portfolios.find(p => p.id === selectedPortfolioId) || portfolios[0]

  const tabs = [
    { id: 'overview', label: 'Overview', icon: PieChart },
    { id: 'performance', label: 'Performance', icon: BarChart3 },
    { id: 'risk', label: 'Risk Analysis', icon: Shield },
    { id: 'settings', label: 'Settings', icon: Settings },
  ]

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Card className="animate-pulse">
          <CardHeader>
            <div className="h-6 bg-muted rounded w-1/3"></div>
          </CardHeader>
          <CardContent>
            <div className="h-32 bg-muted rounded"></div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Portfolio Selector */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Wallet className="h-5 w-5" />
              <span>Portfolio Manager</span>
            </CardTitle>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm">
                <Upload className="h-4 w-4 mr-2" />
                Import
              </Button>
              <Button variant="outline" size="sm">
                <Download className="h-4 w-4 mr-2" />
                Export
              </Button>
              <Button size="sm">
                <Plus className="h-4 w-4 mr-2" />
                New Portfolio
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {portfolios.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {portfolios.map((portfolio) => (
                <div
                  key={portfolio.id}
                  className={`p-4 rounded-lg border cursor-pointer transition-colors ${
                    selectedPortfolioId === portfolio.id
                      ? 'border-primary bg-primary/5'
                      : 'border-border hover:bg-muted/50'
                  }`}
                  onClick={() => setSelectedPortfolioId(portfolio.id)}
                >
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium">{portfolio.name}</h3>
                    <Badge variant={portfolio.totalPnL >= 0 ? 'profit' : 'loss'}>
                      {portfolio.totalPnL >= 0 ? '+' : ''}
                      {formatPercent(portfolio.totalPnLPercent)}
                    </Badge>
                  </div>
                  <div className="text-2xl font-bold mb-1">
                    {formatCurrency(portfolio.totalValue)}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {portfolio.holdings?.length || 0} holdings
                  </div>
                  <div className="text-sm text-muted-foreground">
                    24h: <span className={portfolio.dayChange >= 0 ? 'text-profit' : 'text-loss'}>
                      {portfolio.dayChange >= 0 ? '+' : ''}
                      {formatCurrency(portfolio.dayChange)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-12">
              <Wallet className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <h3 className="text-lg font-medium mb-2">No Portfolios Found</h3>
              <p className="text-muted-foreground mb-4">
                Create your first portfolio to start tracking your investments
              </p>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Portfolio
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Portfolio Tabs */}
      {selectedPortfolio && (
        <>
          <Card>
            <CardHeader>
              <div className="flex items-center space-x-1">
                {tabs.map((tab) => (
                  <Button
                    key={tab.id}
                    variant={activeTab === tab.id ? 'default' : 'ghost'}
                    size="sm"
                    onClick={() => setActiveTab(tab.id as any)}
                    className="flex items-center space-x-2"
                  >
                    <tab.icon className="h-4 w-4" />
                    <span>{tab.label}</span>
                  </Button>
                ))}
              </div>
            </CardHeader>
          </Card>

          {/* Tab Content */}
          <div className="space-y-6">
            {activeTab === 'overview' && <PortfolioSummary />}
            {activeTab === 'performance' && <PortfolioPerformanceChart />}
            {activeTab === 'risk' && <PortfolioRiskMetrics />}
            {activeTab === 'settings' && (
              <Card>
                <CardHeader>
                  <CardTitle>Portfolio Settings</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <h4 className="font-medium mb-2">Portfolio Information</h4>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <label className="text-sm font-medium">Portfolio Name</label>
                          <input
                            type="text"
                            className="w-full mt-1 px-3 py-2 border border-border rounded-md"
                            defaultValue={selectedPortfolio.name}
                          />
                        </div>
                        <div>
                          <label className="text-sm font-medium">Created</label>
                          <div className="mt-1 px-3 py-2 bg-muted rounded-md text-sm">
                            {new Date(selectedPortfolio.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                    </div>

                    <div>
                      <h4 className="font-medium mb-2">Sync Settings</h4>
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-sm">Auto-sync with wallets</span>
                          <Button variant="outline" size="sm">Configure</Button>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm">Real-time price updates</span>
                          <Badge variant="profit">Enabled</Badge>
                        </div>
                      </div>
                    </div>

                    <div>
                      <h4 className="font-medium mb-2">Notifications</h4>
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-sm">Price alerts</span>
                          <Button variant="outline" size="sm">Configure</Button>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm">Performance reports</span>
                          <Button variant="outline" size="sm">Configure</Button>
                        </div>
                      </div>
                    </div>

                    <div className="pt-4 border-t">
                      <div className="flex items-center space-x-2">
                        <Button variant="default">Save Changes</Button>
                        <Button variant="outline">Reset</Button>
                        <Button variant="destructive" className="ml-auto">
                          Delete Portfolio
                        </Button>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </>
      )}
    </div>
  )
}
