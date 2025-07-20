'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useTradingStore } from '@/stores/trading-store'
import { marketApi } from '@/lib/api'
import { formatCurrency, formatPercent, getPriceChangeColor, formatCompactNumber } from '@/lib/utils'
import {
  TrendingUp,
  TrendingDown,
  BarChart3,
  Globe,
  Zap,
  RefreshCw,
} from 'lucide-react'

export function MarketOverview() {
  const { marketData, priceUpdates } = useTradingStore()

  // Fetch market overview
  const { data: marketOverview, isLoading: overviewLoading, refetch: refetchOverview } = useQuery({
    queryKey: ['market-overview'],
    queryFn: marketApi.getMarketOverview,
    refetchInterval: 60000,
  })

  // Fetch top gainers and losers
  const { data: topGainers, isLoading: gainersLoading } = useQuery({
    queryKey: ['top-gainers'],
    queryFn: marketApi.getTopGainers,
    refetchInterval: 30000,
  })

  const { data: topLosers, isLoading: losersLoading } = useQuery({
    queryKey: ['top-losers'],
    queryFn: marketApi.getTopLosers,
    refetchInterval: 30000,
  })

  // Fetch market prices
  const { data: marketPrices, isLoading: pricesLoading } = useQuery({
    queryKey: ['market-prices'],
    queryFn: marketApi.getPrices,
    refetchInterval: 10000,
  })

  const isLoading = overviewLoading || gainersLoading || losersLoading || pricesLoading

  // Get major cryptocurrencies with real-time updates
  const majorCryptos = ['BTC', 'ETH', 'BNB', 'ADA', 'SOL', 'XRP', 'DOT', 'AVAX']
  const majorCryptoData = majorCryptos.map(symbol => {
    const marketInfo = marketData[symbol] || marketPrices?.find(p => p.symbol === symbol)
    const realtimeUpdate = priceUpdates[symbol]
    
    return {
      symbol,
      name: marketInfo?.name || symbol,
      price: realtimeUpdate?.price || marketInfo?.currentPrice || 0,
      change24h: realtimeUpdate?.changePercent || marketInfo?.change24h || 0,
      volume24h: marketInfo?.volume24h || 0,
      marketCap: marketInfo?.marketCap || 0,
      isUpdated: !!realtimeUpdate,
    }
  }).filter(crypto => crypto.price > 0)

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Card className="animate-pulse">
          <CardHeader>
            <div className="h-6 bg-muted rounded w-1/3"></div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="h-16 bg-muted rounded"></div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Market Statistics */}
      {marketOverview && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center space-x-2">
                <Globe className="h-5 w-5" />
                <span>Global Market Overview</span>
              </CardTitle>
              <Button variant="outline" size="sm" onClick={() => refetchOverview()}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <div className="text-center">
                <div className="text-2xl font-bold">
                  {formatCompactNumber(marketOverview.totalMarketCap)}
                </div>
                <div className="text-sm text-muted-foreground">Total Market Cap</div>
                <div className={`text-xs ${getPriceChangeColor(marketOverview.marketCapChange24h)}`}>
                  {marketOverview.marketCapChange24h >= 0 ? '+' : ''}
                  {formatPercent(marketOverview.marketCapChange24h)}
                </div>
              </div>

              <div className="text-center">
                <div className="text-2xl font-bold">
                  {formatCompactNumber(marketOverview.totalVolume24h)}
                </div>
                <div className="text-sm text-muted-foreground">24h Volume</div>
                <div className={`text-xs ${getPriceChangeColor(marketOverview.volumeChange24h)}`}>
                  {marketOverview.volumeChange24h >= 0 ? '+' : ''}
                  {formatPercent(marketOverview.volumeChange24h)}
                </div>
              </div>

              <div className="text-center">
                <div className="text-2xl font-bold">
                  {marketOverview.btcDominance.toFixed(1)}%
                </div>
                <div className="text-sm text-muted-foreground">BTC Dominance</div>
                <div className="text-xs text-muted-foreground">
                  ETH: {marketOverview.ethDominance.toFixed(1)}%
                </div>
              </div>

              <div className="text-center">
                <div className="text-2xl font-bold">
                  {marketOverview.fearGreedIndex}
                </div>
                <div className="text-sm text-muted-foreground">Fear & Greed Index</div>
                <Badge
                  variant={
                    marketOverview.fearGreedIndex > 75 ? 'profit' :
                    marketOverview.fearGreedIndex > 50 ? 'warning' :
                    marketOverview.fearGreedIndex > 25 ? 'secondary' : 'loss'
                  }
                  className="text-xs"
                >
                  {marketOverview.fearGreedIndex > 75 ? 'Extreme Greed' :
                   marketOverview.fearGreedIndex > 50 ? 'Greed' :
                   marketOverview.fearGreedIndex > 25 ? 'Neutral' : 'Fear'}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Major Cryptocurrencies */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <BarChart3 className="h-5 w-5" />
            <span>Major Cryptocurrencies</span>
            <Badge variant="secondary" className="ml-2">
              Live Prices
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-2">Asset</th>
                  <th className="text-right py-2">Price</th>
                  <th className="text-right py-2">24h Change</th>
                  <th className="text-right py-2">Market Cap</th>
                  <th className="text-right py-2">Volume</th>
                  <th className="text-center py-2">Status</th>
                </tr>
              </thead>
              <tbody>
                {majorCryptoData.map((crypto) => (
                  <tr
                    key={crypto.symbol}
                    className={`border-b hover:bg-muted/50 transition-colors ${
                      crypto.isUpdated ? 'animate-pulse-profit' : ''
                    }`}
                  >
                    <td className="py-3">
                      <div className="flex items-center space-x-2">
                        <div className="font-medium">{crypto.symbol}</div>
                        <div className="text-sm text-muted-foreground">{crypto.name}</div>
                      </div>
                    </td>
                    <td className="text-right py-3 font-mono">
                      {formatCurrency(crypto.price)}
                    </td>
                    <td className={`text-right py-3 ${getPriceChangeColor(crypto.change24h)}`}>
                      {crypto.change24h >= 0 ? '+' : ''}
                      {formatPercent(crypto.change24h)}
                    </td>
                    <td className="text-right py-3">
                      {formatCompactNumber(crypto.marketCap)}
                    </td>
                    <td className="text-right py-3">
                      {formatCompactNumber(crypto.volume24h)}
                    </td>
                    <td className="text-center py-3">
                      {crypto.isUpdated ? (
                        <Badge variant="profit" className="text-xs">
                          <Zap className="h-3 w-3 mr-1" />
                          Live
                        </Badge>
                      ) : (
                        <Badge variant="secondary" className="text-xs">
                          Static
                        </Badge>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Top Gainers and Losers */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Gainers */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2 text-profit">
              <TrendingUp className="h-5 w-5" />
              <span>Top Gainers</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {topGainers && topGainers.length > 0 ? (
              <div className="space-y-3">
                {topGainers.slice(0, 5).map((gainer: any, index: number) => (
                  <div key={index} className="flex items-center justify-between p-3 rounded-lg bg-profit/10">
                    <div>
                      <div className="font-medium">{gainer.symbol}</div>
                      <div className="text-sm text-muted-foreground">
                        {formatCurrency(gainer.price)}
                      </div>
                    </div>
                    <Badge variant="profit">
                      +{formatPercent(gainer.change24h)}
                    </Badge>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <TrendingUp className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No gainers data available</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Top Losers */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2 text-loss">
              <TrendingDown className="h-5 w-5" />
              <span>Top Losers</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {topLosers && topLosers.length > 0 ? (
              <div className="space-y-3">
                {topLosers.slice(0, 5).map((loser: any, index: number) => (
                  <div key={index} className="flex items-center justify-between p-3 rounded-lg bg-loss/10">
                    <div>
                      <div className="font-medium">{loser.symbol}</div>
                      <div className="text-sm text-muted-foreground">
                        {formatCurrency(loser.price)}
                      </div>
                    </div>
                    <Badge variant="loss">
                      {formatPercent(loser.change24h)}
                    </Badge>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <TrendingDown className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No losers data available</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
