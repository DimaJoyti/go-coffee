'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { MarketOverview } from './market-overview'
import { MarketHeatmap } from './market-heatmap'
import { TradingViewChart } from './tradingview-chart'
import { useTradingStore } from '@/stores/trading-store'
import { marketApi, enhancedMarketApi } from '@/lib/api'
import { formatCurrency, formatPercent, formatCompactNumber } from '@/lib/utils'
import {
  TrendingUp,
  TrendingDown,
  BarChart3,
  Globe,
  Zap,
  RefreshCw,
  Search,
  Filter,
  Grid,
  List,
} from 'lucide-react'

export function MarketAnalysis() {
  const { marketData, priceUpdates } = useTradingStore()
  const [activeTab, setActiveTab] = useState<'overview' | 'heatmap' | 'charts' | 'analysis'>('overview')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [searchTerm, setSearchTerm] = useState('')

  // Fetch enhanced market data
  const { data: exchangeStatus } = useQuery({
    queryKey: ['exchange-status'],
    queryFn: enhancedMarketApi.getExchangeStatus,
    refetchInterval: 60000,
  })

  const { data: dataQuality } = useQuery({
    queryKey: ['data-quality'],
    queryFn: enhancedMarketApi.getDataQuality,
    refetchInterval: 30000,
  })

  const tabs = [
    { id: 'overview', label: 'Market Overview', icon: Globe },
    { id: 'heatmap', label: 'Heatmap', icon: Grid },
    { id: 'charts', label: 'Charts', icon: BarChart3 },
    { id: 'analysis', label: 'Analysis', icon: TrendingUp },
  ]

  return (
    <div className="space-y-6">
      {/* Market Status Bar */}
      <Card>
        <CardContent className="py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2">
                <div className="h-2 w-2 rounded-full bg-green-500"></div>
                <span className="text-sm font-medium">Markets Open</span>
              </div>
              
              {dataQuality && (
                <div className="flex items-center space-x-2">
                  <Zap className="h-4 w-4 text-yellow-500" />
                  <span className="text-sm">
                    Data Quality: {(dataQuality.overallScore * 100).toFixed(1)}%
                  </span>
                </div>
              )}
              
              {exchangeStatus && (
                <div className="flex items-center space-x-2">
                  <span className="text-sm">
                    {exchangeStatus.filter((ex: any) => ex.status === 'online').length} exchanges online
                  </span>
                </div>
              )}
            </div>
            
            <div className="flex items-center space-x-2">
              <div className="flex items-center space-x-1">
                <Button
                  variant={viewMode === 'grid' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setViewMode('grid')}
                >
                  <Grid className="h-4 w-4" />
                </Button>
                <Button
                  variant={viewMode === 'list' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setViewMode('list')}
                >
                  <List className="h-4 w-4" />
                </Button>
              </div>
              
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <input
                  type="text"
                  placeholder="Search markets..."
                  className="pl-10 pr-4 py-2 border border-border rounded-md bg-background text-sm"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              
              <Button variant="outline" size="sm">
                <Filter className="h-4 w-4 mr-2" />
                Filter
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Market Tabs */}
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
        {activeTab === 'overview' && <MarketOverview />}
        
        {activeTab === 'heatmap' && <MarketHeatmap />}
        
        {activeTab === 'charts' && (
          <div className="space-y-6">
            <TradingViewChart />
            
            {/* Popular Trading Pairs */}
            <Card>
              <CardHeader>
                <CardTitle>Popular Trading Pairs</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {['BTC/USDT', 'ETH/USDT', 'BNB/USDT', 'ADA/USDT', 'SOL/USDT', 'XRP/USDT'].map((pair) => (
                    <div key={pair} className="p-4 rounded-lg border hover:bg-muted/50 cursor-pointer">
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="font-medium">{pair}</div>
                          <div className="text-sm text-muted-foreground">
                            {formatCurrency(Math.random() * 100000 + 10000)}
                          </div>
                        </div>
                        <Badge variant={Math.random() > 0.5 ? 'profit' : 'loss'}>
                          {Math.random() > 0.5 ? '+' : ''}{(Math.random() * 10 - 5).toFixed(2)}%
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        )}
        
        {activeTab === 'analysis' && (
          <div className="space-y-6">
            {/* Market Sentiment */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <TrendingUp className="h-5 w-5" />
                  <span>Market Sentiment Analysis</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-profit">72</div>
                    <div className="text-sm text-muted-foreground">Fear & Greed Index</div>
                    <Badge variant="profit" className="mt-2">Greed</Badge>
                  </div>
                  
                  <div className="text-center">
                    <div className="text-3xl font-bold">68%</div>
                    <div className="text-sm text-muted-foreground">Bullish Sentiment</div>
                    <Badge variant="profit" className="mt-2">Bullish</Badge>
                  </div>
                  
                  <div className="text-center">
                    <div className="text-3xl font-bold">2.4M</div>
                    <div className="text-sm text-muted-foreground">Social Mentions</div>
                    <Badge variant="secondary" className="mt-2">24h</Badge>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Technical Indicators */}
            <Card>
              <CardHeader>
                <CardTitle>Technical Indicators Summary</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {[
                    { name: 'RSI (14)', value: '65.2', signal: 'Neutral', color: 'secondary' },
                    { name: 'MACD', value: 'Bullish', signal: 'Buy', color: 'profit' },
                    { name: 'Moving Average (50)', value: 'Above', signal: 'Bullish', color: 'profit' },
                    { name: 'Bollinger Bands', value: 'Middle', signal: 'Neutral', color: 'secondary' },
                    { name: 'Stochastic', value: '78.5', signal: 'Overbought', color: 'warning' },
                  ].map((indicator, index) => (
                    <div key={index} className="flex items-center justify-between p-3 rounded-lg bg-muted/50">
                      <div>
                        <div className="font-medium">{indicator.name}</div>
                        <div className="text-sm text-muted-foreground">{indicator.value}</div>
                      </div>
                      <Badge variant={indicator.color as any}>
                        {indicator.signal}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* Market News & Events */}
            <Card>
              <CardHeader>
                <CardTitle>Market News & Events</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {[
                    {
                      title: 'Bitcoin ETF Approval Expected This Week',
                      time: '2 hours ago',
                      impact: 'high',
                      sentiment: 'positive'
                    },
                    {
                      title: 'Ethereum Network Upgrade Scheduled',
                      time: '4 hours ago',
                      impact: 'medium',
                      sentiment: 'positive'
                    },
                    {
                      title: 'Regulatory Clarity on DeFi Protocols',
                      time: '6 hours ago',
                      impact: 'medium',
                      sentiment: 'neutral'
                    },
                  ].map((news, index) => (
                    <div key={index} className="p-4 rounded-lg border hover:bg-muted/50">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="font-medium">{news.title}</div>
                          <div className="text-sm text-muted-foreground">{news.time}</div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <Badge
                            variant={
                              news.impact === 'high' ? 'destructive' :
                              news.impact === 'medium' ? 'warning' : 'secondary'
                            }
                            className="text-xs"
                          >
                            {news.impact} impact
                          </Badge>
                          <Badge
                            variant={
                              news.sentiment === 'positive' ? 'profit' :
                              news.sentiment === 'negative' ? 'loss' : 'secondary'
                            }
                            className="text-xs"
                          >
                            {news.sentiment}
                          </Badge>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </div>
  )
}
