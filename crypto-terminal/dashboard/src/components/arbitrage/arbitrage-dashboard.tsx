'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useTradingStore } from '@/stores/trading-store'
import { arbitrageApi } from '@/lib/api'
import { formatCurrency, formatPercent } from '@/lib/utils'
import {
  ArrowLeftRight,
  TrendingUp,
  Clock,
  DollarSign,
  RefreshCw,
  Filter,
  AlertTriangle,
  CheckCircle,
  Zap,
} from 'lucide-react'

export function ArbitrageDashboard() {
  const { arbitrageOpportunities } = useTradingStore()
  const [sortBy, setSortBy] = useState<'spread' | 'volume' | 'profit'>('spread')
  const [minSpread, setMinSpread] = useState(0.5)

  // Fetch arbitrage opportunities
  const { data: opportunities, isLoading, refetch } = useQuery({
    queryKey: ['arbitrage-opportunities'],
    queryFn: arbitrageApi.getOpportunities,
    refetchInterval: 10000, // Refresh every 10 seconds for arbitrage
  })

  // Generate mock arbitrage data
  const generateMockOpportunities = () => {
    const symbols = ['BTC', 'ETH', 'BNB', 'ADA', 'SOL', 'XRP', 'DOT', 'AVAX']
    const exchanges = ['Binance', 'Coinbase', 'Kraken', 'KuCoin', 'Huobi', 'OKX']
    
    return symbols.map(symbol => {
      const buyExchange = exchanges[Math.floor(Math.random() * exchanges.length)]
      let sellExchange = exchanges[Math.floor(Math.random() * exchanges.length)]
      while (sellExchange === buyExchange) {
        sellExchange = exchanges[Math.floor(Math.random() * exchanges.length)]
      }
      
      const basePrice = Math.random() * 1000 + 100
      const spread = Math.random() * 5 + 0.5 // 0.5% to 5.5%
      const buyPrice = basePrice
      const sellPrice = basePrice * (1 + spread / 100)
      const volume = Math.random() * 50000 + 10000
      const profitPotential = volume * (spread / 100)
      
      return {
        symbol,
        buyExchange,
        sellExchange,
        buyPrice,
        sellPrice,
        spread: sellPrice - buyPrice,
        spreadPercent: spread,
        volume,
        profitPotential,
        timestamp: new Date().toISOString(),
      }
    })
  }

  const mockOpportunities = generateMockOpportunities()
  const displayOpportunities = opportunities || mockOpportunities

  // Filter and sort opportunities
  const filteredOpportunities = displayOpportunities
    .filter(opp => opp.spreadPercent >= minSpread)
    .sort((a, b) => {
      switch (sortBy) {
        case 'spread':
          return b.spreadPercent - a.spreadPercent
        case 'volume':
          return b.volume - a.volume
        case 'profit':
          return b.profitPotential - a.profitPotential
        default:
          return 0
      }
    })

  // Calculate statistics
  const totalOpportunities = filteredOpportunities.length
  const avgSpread = filteredOpportunities.reduce((sum, opp) => sum + opp.spreadPercent, 0) / totalOpportunities || 0
  const totalProfitPotential = filteredOpportunities.reduce((sum, opp) => sum + opp.profitPotential, 0)
  const bestOpportunity = filteredOpportunities[0]

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
      {/* Arbitrage Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <ArrowLeftRight className="h-4 w-4" />
              <span>Opportunities</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{totalOpportunities}</div>
            <div className="text-sm text-muted-foreground">
              Active arbitrage opportunities
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <TrendingUp className="h-4 w-4" />
              <span>Average Spread</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-profit">
              {avgSpread.toFixed(2)}%
            </div>
            <div className="text-sm text-muted-foreground">
              Across all opportunities
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <DollarSign className="h-4 w-4" />
              <span>Profit Potential</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-profit">
              {formatCurrency(totalProfitPotential)}
            </div>
            <div className="text-sm text-muted-foreground">
              Total potential profit
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center space-x-2">
              <Zap className="h-4 w-4" />
              <span>Best Opportunity</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {bestOpportunity ? (
              <>
                <div className="text-2xl font-bold">{bestOpportunity.symbol}</div>
                <div className="text-sm text-profit">
                  {formatPercent(bestOpportunity.spreadPercent)} spread
                </div>
              </>
            ) : (
              <div className="text-sm text-muted-foreground">No opportunities</div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Controls */}
      <Card>
        <CardContent className="py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">Sort by:</span>
                <div className="flex items-center space-x-1">
                  <Button
                    variant={sortBy === 'spread' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSortBy('spread')}
                  >
                    Spread
                  </Button>
                  <Button
                    variant={sortBy === 'volume' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSortBy('volume')}
                  >
                    Volume
                  </Button>
                  <Button
                    variant={sortBy === 'profit' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSortBy('profit')}
                  >
                    Profit
                  </Button>
                </div>
              </div>
              
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">Min Spread:</span>
                <select
                  value={minSpread}
                  onChange={(e) => setMinSpread(Number(e.target.value))}
                  className="px-3 py-1 border border-border rounded-md bg-background text-sm"
                >
                  <option value={0}>0%</option>
                  <option value={0.5}>0.5%</option>
                  <option value={1}>1%</option>
                  <option value={2}>2%</option>
                  <option value={3}>3%</option>
                </select>
              </div>
            </div>
            
            <div className="flex items-center space-x-2">
              <Badge variant="secondary" className="flex items-center space-x-1">
                <Clock className="h-3 w-3" />
                <span>Auto-refresh: 10s</span>
              </Badge>
              <Button variant="outline" size="sm" onClick={() => refetch()}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Arbitrage Opportunities Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <ArrowLeftRight className="h-5 w-5" />
            <span>Arbitrage Opportunities</span>
            <Badge variant="secondary">{filteredOpportunities.length}</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {filteredOpportunities.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-3">Asset</th>
                    <th className="text-left py-3">Buy Exchange</th>
                    <th className="text-right py-3">Buy Price</th>
                    <th className="text-left py-3">Sell Exchange</th>
                    <th className="text-right py-3">Sell Price</th>
                    <th className="text-right py-3">Spread</th>
                    <th className="text-right py-3">Volume</th>
                    <th className="text-right py-3">Profit Potential</th>
                    <th className="text-center py-3">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredOpportunities.map((opportunity, index) => (
                    <tr key={index} className="border-b hover:bg-muted/50">
                      <td className="py-4">
                        <div className="font-medium">{opportunity.symbol}</div>
                      </td>
                      <td className="py-4">
                        <Badge variant="outline" className="text-xs">
                          {opportunity.buyExchange}
                        </Badge>
                      </td>
                      <td className="text-right py-4 font-mono">
                        {formatCurrency(opportunity.buyPrice)}
                      </td>
                      <td className="py-4">
                        <Badge variant="outline" className="text-xs">
                          {opportunity.sellExchange}
                        </Badge>
                      </td>
                      <td className="text-right py-4 font-mono">
                        {formatCurrency(opportunity.sellPrice)}
                      </td>
                      <td className="text-right py-4">
                        <Badge
                          variant={
                            opportunity.spreadPercent > 3 ? 'profit' :
                            opportunity.spreadPercent > 1 ? 'warning' : 'secondary'
                          }
                        >
                          {formatPercent(opportunity.spreadPercent)}
                        </Badge>
                      </td>
                      <td className="text-right py-4">
                        {formatCurrency(opportunity.volume, 'USD', 0)}
                      </td>
                      <td className="text-right py-4 font-semibold text-profit">
                        {formatCurrency(opportunity.profitPotential)}
                      </td>
                      <td className="text-center py-4">
                        <div className="flex items-center justify-center space-x-1">
                          <Button variant="outline" size="sm">
                            Execute
                          </Button>
                          <Button variant="ghost" size="sm">
                            <AlertTriangle className="h-4 w-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-12">
              <ArrowLeftRight className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <h3 className="text-lg font-medium mb-2">No Arbitrage Opportunities</h3>
              <p className="text-muted-foreground mb-4">
                No opportunities found with the current filters. Try lowering the minimum spread.
              </p>
              <Button onClick={() => setMinSpread(0)}>
                Reset Filters
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Risk Warning */}
      <Card className="border-yellow-200 bg-yellow-50 dark:bg-yellow-950/20">
        <CardContent className="py-4">
          <div className="flex items-start space-x-3">
            <AlertTriangle className="h-5 w-5 text-yellow-600 mt-0.5" />
            <div>
              <div className="font-medium text-yellow-800 dark:text-yellow-200">
                Arbitrage Trading Risks
              </div>
              <div className="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                Arbitrage opportunities can disappear quickly due to market movements, execution delays, 
                and network congestion. Always consider transaction fees, withdrawal limits, and execution time 
                when evaluating opportunities. Past performance does not guarantee future results.
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
