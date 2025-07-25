'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useTradingStore } from '@/stores/trading-store'
import { marketApi } from '@/lib/api'
import { formatCurrency, formatPercent, getPriceChangeColor } from '@/lib/utils'
import {
  Grid,
  RefreshCw,
  TrendingUp,
  TrendingDown,
  Zap,
} from 'lucide-react'

interface HeatmapItem {
  symbol: string
  name: string
  price: number
  change24h: number
  marketCap: number
  volume24h: number
  size: number // For heatmap sizing
}

export function MarketHeatmap() {
  const { marketData, priceUpdates } = useTradingStore()

  // Fetch market heatmap data
  const { data: heatmapData, isLoading, refetch } = useQuery({
    queryKey: ['market-heatmap'],
    queryFn: marketApi.getMarketHeatmap,
    refetchInterval: 30000,
  })

  // Generate mock heatmap data
  const generateMockData = (): HeatmapItem[] => {
    const cryptos = [
      { symbol: 'BTC', name: 'Bitcoin', marketCap: 800000000000 },
      { symbol: 'ETH', name: 'Ethereum', marketCap: 400000000000 },
      { symbol: 'BNB', name: 'BNB', marketCap: 80000000000 },
      { symbol: 'XRP', name: 'XRP', marketCap: 60000000000 },
      { symbol: 'ADA', name: 'Cardano', marketCap: 40000000000 },
      { symbol: 'SOL', name: 'Solana', marketCap: 35000000000 },
      { symbol: 'DOT', name: 'Polkadot', marketCap: 25000000000 },
      { symbol: 'DOGE', name: 'Dogecoin', marketCap: 20000000000 },
      { symbol: 'AVAX', name: 'Avalanche', marketCap: 18000000000 },
      { symbol: 'SHIB', name: 'Shiba Inu', marketCap: 15000000000 },
      { symbol: 'MATIC', name: 'Polygon', marketCap: 12000000000 },
      { symbol: 'LTC', name: 'Litecoin', marketCap: 10000000000 },
      { symbol: 'UNI', name: 'Uniswap', marketCap: 8000000000 },
      { symbol: 'LINK', name: 'Chainlink', marketCap: 7000000000 },
      { symbol: 'ATOM', name: 'Cosmos', marketCap: 6000000000 },
      { symbol: 'XLM', name: 'Stellar', marketCap: 5000000000 },
    ]

    return cryptos.map(crypto => {
      const change = (Math.random() - 0.5) * 20 // -10% to +10%
      const price = Math.random() * 1000 + 10
      const volume = crypto.marketCap * (Math.random() * 0.1 + 0.05) // 5-15% of market cap
      
      return {
        symbol: crypto.symbol,
        name: crypto.name,
        price,
        change24h: change,
        marketCap: crypto.marketCap,
        volume24h: volume,
        size: Math.sqrt(crypto.marketCap / 1000000000), // Size based on market cap
      }
    })
  }

  const heatmapItems = heatmapData || generateMockData()

  // Calculate grid layout
  const getGridSize = (size: number) => {
    if (size > 20) return 'col-span-4 row-span-3'
    if (size > 15) return 'col-span-3 row-span-2'
    if (size > 10) return 'col-span-2 row-span-2'
    if (size > 5) return 'col-span-2 row-span-1'
    return 'col-span-1 row-span-1'
  }

  const getChangeColor = (change: number) => {
    if (change > 5) return 'bg-green-500'
    if (change > 2) return 'bg-green-400'
    if (change > 0) return 'bg-green-300'
    if (change > -2) return 'bg-red-300'
    if (change > -5) return 'bg-red-400'
    return 'bg-red-500'
  }

  const getTextColor = (change: number) => {
    return Math.abs(change) > 2 ? 'text-white' : 'text-gray-900'
  }

  if (isLoading) {
    return (
      <Card className="animate-pulse">
        <CardHeader>
          <div className="h-6 bg-muted rounded w-1/3"></div>
        </CardHeader>
        <CardContent>
          <div className="h-96 bg-muted rounded"></div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Heatmap Controls */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Grid className="h-5 w-5" />
              <span>Market Heatmap</span>
            </CardTitle>
            <div className="flex items-center space-x-2">
              <div className="flex items-center space-x-2 text-sm">
                <span>Size by:</span>
                <Badge variant="outline">Market Cap</Badge>
              </div>
              <div className="flex items-center space-x-2 text-sm">
                <span>Color by:</span>
                <Badge variant="outline">24h Change</Badge>
              </div>
              <Button variant="outline" size="sm" onClick={() => refetch()}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Heatmap Legend */}
      <Card>
        <CardContent className="py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <span className="text-sm font-medium">24h Change:</span>
              <div className="flex items-center space-x-2">
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-red-500 rounded"></div>
                  <span className="text-xs">-5%+</span>
                </div>
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-red-300 rounded"></div>
                  <span className="text-xs">-2%</span>
                </div>
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-gray-300 rounded"></div>
                  <span className="text-xs">0%</span>
                </div>
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-green-300 rounded"></div>
                  <span className="text-xs">+2%</span>
                </div>
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-green-500 rounded"></div>
                  <span className="text-xs">+5%+</span>
                </div>
              </div>
            </div>
            
            <div className="flex items-center space-x-4">
              <span className="text-sm font-medium">Size by Market Cap</span>
              <div className="flex items-center space-x-2">
                <TrendingUp className="h-4 w-4 text-green-500" />
                <span className="text-sm text-green-500">
                  {heatmapItems.filter((item: any) => item.change24h > 0).length} gaining
                </span>
                <TrendingDown className="h-4 w-4 text-red-500" />
                <span className="text-sm text-red-500">
                  {heatmapItems.filter((item: any) => item.change24h < 0).length} losing
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Heatmap Grid */}
      <Card>
        <CardContent className="p-6">
          <div className="grid grid-cols-8 gap-2 h-96">
            {heatmapItems.map((item: any) => {
              const realtimeUpdate = priceUpdates[item.symbol]
              const currentChange = realtimeUpdate?.changePercent || item.change24h
              const isUpdated = !!realtimeUpdate
              
              return (
                <div
                  key={item.symbol}
                  className={`
                    ${getGridSize(item.size)}
                    ${getChangeColor(currentChange)}
                    ${getTextColor(currentChange)}
                    rounded-lg p-3 cursor-pointer transition-all duration-300 hover:scale-105 hover:shadow-lg
                    ${isUpdated ? 'ring-2 ring-yellow-400 animate-pulse' : ''}
                    flex flex-col justify-between
                  `}
                  title={`${item.name} (${item.symbol})`}
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="font-bold text-lg">{item.symbol}</div>
                      <div className="text-xs opacity-90">{item.name}</div>
                    </div>
                    {isUpdated && (
                      <Zap className="h-4 w-4 text-yellow-300" />
                    )}
                  </div>
                  
                  <div className="mt-auto">
                    <div className="font-semibold">
                      {formatCurrency(realtimeUpdate?.price || item.price)}
                    </div>
                    <div className="text-sm font-medium">
                      {currentChange >= 0 ? '+' : ''}{formatPercent(currentChange)}
                    </div>
                    <div className="text-xs opacity-75">
                      Vol: {formatCurrency(item.volume24h, 'USD', 0)}
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </CardContent>
      </Card>

      {/* Market Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Market Cap
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(heatmapItems.reduce((sum: number, item: any) => sum + item.marketCap, 0), 'USD', 0)}
            </div>
            <div className="text-sm text-muted-foreground">
              Across {heatmapItems.length} assets
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              24h Volume
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(heatmapItems.reduce((sum: number, item: any) => sum + item.volume24h, 0), 'USD', 0)}
            </div>
            <div className="text-sm text-muted-foreground">
              Trading volume
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Gainers
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">
              {heatmapItems.filter((item: any) => item.change24h > 0).length}
            </div>
            <div className="text-sm text-muted-foreground">
              Assets in green
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Losers
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-500">
              {heatmapItems.filter((item: any) => item.change24h < 0).length}
            </div>
            <div className="text-sm text-muted-foreground">
              Assets in red
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
