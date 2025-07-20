'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useTradingStore } from '@/stores/trading-store'
import { formatCurrency, formatPercent } from '@/lib/utils'
import {
  BarChart3,
  TrendingUp,
  TrendingDown,
  Maximize2,
  Settings,
  RefreshCw,
} from 'lucide-react'

const timeframes = [
  { id: '1m', label: '1m' },
  { id: '5m', label: '5m' },
  { id: '15m', label: '15m' },
  { id: '1h', label: '1h' },
  { id: '4h', label: '4h' },
  { id: '1d', label: '1D' },
  { id: '1w', label: '1W' },
]

const symbols = [
  { symbol: 'BTCUSDT', name: 'Bitcoin' },
  { symbol: 'ETHUSDT', name: 'Ethereum' },
  { symbol: 'BNBUSDT', name: 'BNB' },
  { symbol: 'ADAUSDT', name: 'Cardano' },
  { symbol: 'SOLUSDT', name: 'Solana' },
  { symbol: 'XRPUSDT', name: 'XRP' },
]

export function TradingViewChart() {
  const { marketData, priceUpdates } = useTradingStore()
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT')
  const [selectedTimeframe, setSelectedTimeframe] = useState('1h')
  const [isFullscreen, setIsFullscreen] = useState(false)

  // Mock chart data - in real implementation, this would come from TradingView API
  const currentPrice = priceUpdates[selectedSymbol.replace('USDT', '')]?.price || 
                      marketData[selectedSymbol.replace('USDT', '')]?.currentPrice || 
                      45000

  const change24h = priceUpdates[selectedSymbol.replace('USDT', '')]?.changePercent || 
                   marketData[selectedSymbol.replace('USDT', '')]?.change24h || 
                   2.5

  return (
    <div className="space-y-6">
      {/* Chart Controls */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <BarChart3 className="h-5 w-5" />
              <span>Advanced Charts</span>
            </CardTitle>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm">
                <Settings className="h-4 w-4 mr-2" />
                Indicators
              </Button>
              <Button variant="outline" size="sm">
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
              <Button 
                variant="outline" 
                size="sm"
                onClick={() => setIsFullscreen(!isFullscreen)}
              >
                <Maximize2 className="h-4 w-4 mr-2" />
                Fullscreen
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between mb-4">
            {/* Symbol Selector */}
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">Symbol:</span>
              <div className="flex items-center space-x-1">
                {symbols.map((symbol) => (
                  <Button
                    key={symbol.symbol}
                    variant={selectedSymbol === symbol.symbol ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSelectedSymbol(symbol.symbol)}
                  >
                    {symbol.symbol.replace('USDT', '')}
                  </Button>
                ))}
              </div>
            </div>
            
            {/* Timeframe Selector */}
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">Timeframe:</span>
              <div className="flex items-center space-x-1">
                {timeframes.map((tf) => (
                  <Button
                    key={tf.id}
                    variant={selectedTimeframe === tf.id ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSelectedTimeframe(tf.id)}
                  >
                    {tf.label}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Chart Container */}
      <Card className={isFullscreen ? 'fixed inset-4 z-50' : ''}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div>
                <div className="text-2xl font-bold">
                  {selectedSymbol.replace('USDT', '')}/USDT
                </div>
                <div className="text-sm text-muted-foreground">
                  {symbols.find(s => s.symbol === selectedSymbol)?.name}
                </div>
              </div>
              
              <div className="text-right">
                <div className="text-2xl font-bold">
                  {formatCurrency(currentPrice)}
                </div>
                <div className={`text-sm ${change24h >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                  {change24h >= 0 ? '+' : ''}{formatPercent(change24h)} (24h)
                </div>
              </div>
            </div>
            
            {isFullscreen && (
              <Button 
                variant="outline" 
                size="sm"
                onClick={() => setIsFullscreen(false)}
              >
                Exit Fullscreen
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {/* Chart Placeholder */}
          <div className={`bg-muted/30 rounded-lg flex items-center justify-center ${
            isFullscreen ? 'h-[calc(100vh-200px)]' : 'h-96'
          }`}>
            <div className="text-center">
              <BarChart3 className="h-16 w-16 mx-auto mb-4 text-muted-foreground" />
              <h3 className="text-lg font-medium mb-2">TradingView Chart</h3>
              <p className="text-muted-foreground mb-4">
                Advanced charting with technical indicators
              </p>
              <div className="space-y-2 text-sm text-muted-foreground">
                <div>Symbol: {selectedSymbol}</div>
                <div>Timeframe: {selectedTimeframe}</div>
                <div>Price: {formatCurrency(currentPrice)}</div>
                <div>24h Change: {change24h >= 0 ? '+' : ''}{formatPercent(change24h)}</div>
              </div>
              <div className="mt-4">
                <Badge variant="secondary">
                  TradingView Integration Coming Soon
                </Badge>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Chart Analysis */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Technical Analysis</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm">Overall Signal</span>
                <Badge variant="profit">Strong Buy</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Moving Averages</span>
                <Badge variant="profit">Buy (8/12)</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Oscillators</span>
                <Badge variant="secondary">Neutral (4/8)</Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Key Levels</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm">Resistance</span>
                <span className="text-sm font-medium">
                  {formatCurrency(currentPrice * 1.05)}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Current</span>
                <span className="text-sm font-bold">
                  {formatCurrency(currentPrice)}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Support</span>
                <span className="text-sm font-medium">
                  {formatCurrency(currentPrice * 0.95)}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Volume Analysis</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm">24h Volume</span>
                <span className="text-sm font-medium">
                  {formatCurrency(Math.random() * 1000000000 + 500000000, 'USD', 0, 0)}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Volume Trend</span>
                <Badge variant="profit">
                  <TrendingUp className="h-3 w-3 mr-1" />
                  Increasing
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Avg Volume (7d)</span>
                <span className="text-sm font-medium">
                  {formatCurrency(Math.random() * 800000000 + 400000000, 'USD', 0, 0)}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
