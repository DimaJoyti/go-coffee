'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { CoffeeStrategies } from './coffee-strategies'
// import { StrategyCreator } from './strategy-creator'
// import { StrategyPerformance } from './strategy-performance'
import { useTradingStore } from '@/stores/trading-store'
import { strategyApi } from '@/lib/api'
import { formatCurrency, formatPercent, getCoffeeStrategyEmoji } from '@/lib/utils'
import {
  Coffee,
  Plus,
  Settings,
  BarChart3,
  Play,
  Pause,
  Trash2,
  Edit,
  TrendingUp,
  Target,
  Shield,
} from 'lucide-react'

export function CoffeeStrategiesManager() {
  const queryClient = useQueryClient()
  const { strategies } = useTradingStore()
  const [activeTab, setActiveTab] = useState<'overview' | 'create' | 'performance' | 'settings'>('overview')
  const [selectedStrategy, setSelectedStrategy] = useState<string | null>(null)

  // Fetch strategies
  const { data: strategiesData, isLoading } = useQuery({
    queryKey: ['coffee-strategies'],
    queryFn: strategyApi.getStrategies,
    refetchInterval: 30000,
  })

  // Fetch coffee menu
  const { data: coffeeMenu } = useQuery({
    queryKey: ['coffee-menu'],
    queryFn: strategyApi.getCoffeeMenu,
  })

  const tabs = [
    { id: 'overview', label: 'Overview', icon: Coffee },
    { id: 'create', label: 'Create Strategy', icon: Plus },
    { id: 'performance', label: 'Performance', icon: BarChart3 },
    { id: 'settings', label: 'Settings', icon: Settings },
  ]

  // Calculate strategy statistics
  const activeStrategies = strategies.filter(s => s.status === 'active')
  const totalTrades = strategies.reduce((sum, s) => sum + (s.performance?.totalTrades || 0), 0)
  const avgWinRate = strategies.length > 0 
    ? strategies.reduce((sum, s) => sum + (s.performance?.winRate || 0), 0) / strategies.length
    : 0
  const totalPnL = strategies.reduce((sum, s) => sum + (s.performance?.totalPnL || 0), 0)

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
      {/* Strategy Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <Coffee className="h-4 w-4" />
              <span>Active Strategies</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{activeStrategies.length}</div>
            <div className="text-sm text-muted-foreground">
              {strategies.length} total strategies
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <BarChart3 className="h-4 w-4" />
              <span>Total Trades</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{totalTrades}</div>
            <div className="text-sm text-muted-foreground">
              Across all strategies
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <Target className="h-4 w-4" />
              <span>Average Win Rate</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-profit">
              {avgWinRate.toFixed(1)}%
            </div>
            <div className="text-sm text-muted-foreground">
              Success rate
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <TrendingUp className="h-4 w-4" />
              <span>Total P&L</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className={`text-3xl font-bold ${totalPnL >= 0 ? 'text-profit' : 'text-loss'}`}>
              {totalPnL >= 0 ? '+' : ''}{formatCurrency(totalPnL)}
            </div>
            <div className="text-sm text-muted-foreground">
              All-time performance
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Strategy Tabs */}
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
        {activeTab === 'overview' && (
          <div className="space-y-6">
            <CoffeeStrategies />
            
            {/* Strategy List */}
            <Card>
              <CardHeader>
                <CardTitle>All Strategies</CardTitle>
              </CardHeader>
              <CardContent>
                {strategies.length > 0 ? (
                  <div className="space-y-4">
                    {strategies.map((strategy) => (
                      <div
                        key={strategy.id}
                        className={`p-4 rounded-lg border transition-colors ${
                          selectedStrategy === strategy.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'
                        }`}
                        onClick={() => setSelectedStrategy(strategy.id)}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center space-x-4">
                            <span className="text-3xl">{getCoffeeStrategyEmoji(strategy.type)}</span>
                            <div>
                              <div className="font-medium">{strategy.name}</div>
                              <div className="text-sm text-muted-foreground">
                                {strategy.symbol} • {strategy.type} • Created {new Date(strategy.createdAt).toLocaleDateString()}
                              </div>
                            </div>
                          </div>
                          
                          <div className="flex items-center space-x-4">
                            <div className="text-right">
                              <div className={`font-medium ${
                                (strategy.performance?.totalPnL || 0) >= 0 ? 'text-profit' : 'text-loss'
                              }`}>
                                {strategy.performance?.totalPnL >= 0 ? '+' : ''}
                                {formatCurrency(strategy.performance?.totalPnL || 0)}
                              </div>
                              <div className="text-sm text-muted-foreground">
                                {strategy.performance?.totalTrades || 0} trades • {(strategy.performance?.winRate || 0).toFixed(1)}% win rate
                              </div>
                            </div>
                            
                            <div className="flex items-center space-x-2">
                              <Badge
                                variant={
                                  strategy.status === 'active' ? 'active' :
                                  strategy.status === 'paused' ? 'paused' : 'inactive'
                                }
                              >
                                {strategy.status}
                              </Badge>
                              
                              <div className="flex items-center space-x-1">
                                <Button variant="ghost" size="sm">
                                  <Edit className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="sm">
                                  <Settings className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="sm" className="text-destructive">
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </div>
                            </div>
                          </div>
                        </div>
                        
                        {/* Strategy Settings Preview */}
                        {selectedStrategy === strategy.id && (
                          <div className="mt-4 pt-4 border-t border-border">
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                              <div className="flex items-center space-x-2">
                                <Shield className="h-4 w-4 text-muted-foreground" />
                                <span>Risk: {strategy.settings?.riskLevel || 'medium'}</span>
                              </div>
                              <div className="flex items-center space-x-2">
                                <Target className="h-4 w-4 text-muted-foreground" />
                                <span>Max Size: {formatCurrency(strategy.settings?.maxPositionSize || 0)}</span>
                              </div>
                              <div className="flex items-center space-x-2">
                                <TrendingUp className="h-4 w-4 text-profit" />
                                <span>Take Profit: {formatPercent(strategy.settings?.takeProfit || 0)}</span>
                              </div>
                              <div className="flex items-center space-x-2">
                                <TrendingUp className="h-4 w-4 text-loss rotate-180" />
                                <span>Stop Loss: {formatPercent(strategy.settings?.stopLoss || 0)}</span>
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-12">
                    <Coffee className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                    <h3 className="text-lg font-medium mb-2">No Strategies Found</h3>
                    <p className="text-muted-foreground mb-4">
                      Create your first coffee strategy to start automated trading
                    </p>
                    <Button onClick={() => setActiveTab('create')}>
                      <Plus className="h-4 w-4 mr-2" />
                      Create Strategy
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        )}
        
        {activeTab === 'create' && (
          <Card>
            <CardHeader>
              <CardTitle>Create New Coffee Strategy</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-center py-12">
                <Coffee className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                <h3 className="text-lg font-medium mb-2">Strategy Creator</h3>
                <p className="text-muted-foreground mb-4">
                  Coming soon! Create custom coffee trading strategies.
                </p>
              </div>
            </CardContent>
          </Card>
        )}
        {activeTab === 'performance' && (
          <Card>
            <CardHeader>
              <CardTitle>Strategy Performance Analysis</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-center py-12">
                <BarChart3 className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                <h3 className="text-lg font-medium mb-2">Performance Analytics</h3>
                <p className="text-muted-foreground mb-4">
                  Coming soon! Detailed strategy performance analysis.
                </p>
              </div>
            </CardContent>
          </Card>
        )}
        {activeTab === 'settings' && (
          <Card>
            <CardHeader>
              <CardTitle>Strategy Settings</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                <div>
                  <h4 className="font-medium mb-4">Global Strategy Settings</h4>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">Auto-start new strategies</div>
                        <div className="text-sm text-muted-foreground">
                          Automatically start strategies after creation
                        </div>
                      </div>
                      <Button variant="outline" size="sm">Configure</Button>
                    </div>
                    
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">Risk management</div>
                        <div className="text-sm text-muted-foreground">
                          Global risk limits and controls
                        </div>
                      </div>
                      <Button variant="outline" size="sm">Configure</Button>
                    </div>
                    
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">Notifications</div>
                        <div className="text-sm text-muted-foreground">
                          Strategy alerts and performance reports
                        </div>
                      </div>
                      <Button variant="outline" size="sm">Configure</Button>
                    </div>
                  </div>
                </div>
                
                <div className="pt-4 border-t">
                  <div className="flex items-center space-x-2">
                    <Button>Save Settings</Button>
                    <Button variant="outline">Reset to Defaults</Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}
