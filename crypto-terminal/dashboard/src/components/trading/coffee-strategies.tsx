'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useTradingStore } from '@/stores/trading-store'
import { strategyApi } from '@/lib/api'
import { formatCurrency, formatPercent, getCoffeeStrategyEmoji, getStatusColor } from '@/lib/utils'
import toast from 'react-hot-toast'
import {
  Coffee,
  Play,
  Pause,
  Settings,
  TrendingUp,
  Target,
  Shield,
  Clock,
} from 'lucide-react'

export function CoffeeStrategies() {
  const queryClient = useQueryClient()
  const { strategies, setStrategies } = useTradingStore()
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

  // Start strategy mutation
  const startStrategyMutation = useMutation({
    mutationFn: (strategyId: string) => strategyApi.startStrategy(strategyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['coffee-strategies'] })
      toast.success('☕ Strategy started successfully!')
    },
    onError: (error: any) => {
      toast.error(`Failed to start strategy: ${error.message}`)
    },
  })

  // Stop strategy mutation
  const stopStrategyMutation = useMutation({
    mutationFn: (strategyId: string) => strategyApi.stopStrategy(strategyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['coffee-strategies'] })
      toast.success('Strategy stopped successfully')
    },
    onError: (error: any) => {
      toast.error(`Failed to stop strategy: ${error.message}`)
    },
  })

  const handleStartStrategy = (strategyId: string) => {
    startStrategyMutation.mutate(strategyId)
  }

  const handleStopStrategy = (strategyId: string) => {
    stopStrategyMutation.mutate(strategyId)
  }

  if (isLoading) {
    return (
      <Card className="animate-pulse">
        <CardHeader>
          <div className="h-6 bg-muted rounded w-1/2"></div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-20 bg-muted rounded"></div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Coffee Menu */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Coffee className="h-5 w-5 text-coffee-500" />
            <span>Coffee Strategy Menu</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-3">
            {[
              { type: 'espresso', name: 'Espresso', emoji: '☕', description: 'Quick & Strong' },
              { type: 'latte', name: 'Latte', emoji: '🥛', description: 'Smooth & Balanced' },
              { type: 'cold-brew', name: 'Cold Brew', emoji: '🧊', description: 'Patient & Steady' },
              { type: 'cappuccino', name: 'Cappuccino', emoji: '☕', description: 'Rich & Frothy' },
            ].map((coffee) => (
              <Button
                key={coffee.type}
                variant="outline"
                className="h-auto p-4 flex flex-col items-center space-y-2"
                onClick={() => {
                  // TODO: Open strategy creation dialog
                  toast.success(`${coffee.emoji} ${coffee.name} strategy coming soon!`)
                }}
              >
                <span className="text-2xl">{coffee.emoji}</span>
                <div className="text-center">
                  <div className="font-medium">{coffee.name}</div>
                  <div className="text-xs text-muted-foreground">{coffee.description}</div>
                </div>
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Active Strategies */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>Active Strategies</span>
            <Badge variant="secondary">
              {strategies.filter(s => s.status === 'active').length} running
            </Badge>
          </CardTitle>
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
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center space-x-3">
                      <span className="text-2xl">{getCoffeeStrategyEmoji(strategy.type)}</span>
                      <div>
                        <div className="font-medium">{strategy.name}</div>
                        <div className="text-sm text-muted-foreground">
                          {strategy.symbol} • {strategy.type}
                        </div>
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
                      {strategy.status === 'active' ? (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleStopStrategy(strategy.id)
                          }}
                          disabled={stopStrategyMutation.isPending}
                        >
                          <Pause className="h-4 w-4" />
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          variant="coffee"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleStartStrategy(strategy.id)
                          }}
                          disabled={startStrategyMutation.isPending}
                        >
                          <Play className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </div>

                  {/* Strategy Performance */}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                    <div className="text-center">
                      <div className="font-medium">{strategy.performance?.totalTrades || 0}</div>
                      <div className="text-muted-foreground">Trades</div>
                    </div>
                    <div className="text-center">
                      <div className="font-medium text-profit">
                        {(strategy.performance?.winRate || 0).toFixed(1)}%
                      </div>
                      <div className="text-muted-foreground">Win Rate</div>
                    </div>
                    <div className="text-center">
                      <div className={`font-medium ${
                        (strategy.performance?.totalPnL || 0) >= 0 ? 'text-profit' : 'text-loss'
                      }`}>
                        {strategy.performance?.totalPnL >= 0 ? '+' : ''}
                        {formatCurrency(strategy.performance?.totalPnL || 0)}
                      </div>
                      <div className="text-muted-foreground">P&L</div>
                    </div>
                    <div className="text-center">
                      <div className="font-medium">
                        {formatPercent(strategy.performance?.totalPnLPercent || 0)}
                      </div>
                      <div className="text-muted-foreground">Return</div>
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
                          <Clock className="h-4 w-4 text-loss" />
                          <span>Stop Loss: {formatPercent(strategy.settings?.stopLoss || 0)}</span>
                        </div>
                      </div>
                      <div className="flex justify-end mt-4">
                        <Button variant="outline" size="sm">
                          <Settings className="h-4 w-4 mr-2" />
                          Configure
                        </Button>
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Coffee className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <h3 className="text-lg font-medium mb-2">No Active Strategies</h3>
              <p className="mb-4">Start brewing your first coffee strategy!</p>
              <Button variant="coffee">
                <Coffee className="h-4 w-4 mr-2" />
                Create Strategy
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
