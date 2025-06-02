'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  TrendingUp, 
  TrendingDown, 
  Wallet, 
  ArrowUpRight,
  ArrowDownRight,
  RefreshCw,
  Eye,
  EyeOff
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { formatCurrency, formatCrypto, formatPercentage, cn } from '@/lib/utils'

interface CryptoAsset {
  symbol: string
  name: string
  balance: number
  usdValue: number
  change24h: number
  price: number
  icon: string
}

interface DefiPortfolioProps {
  className?: string
}

export function DefiPortfolio({ className }: DefiPortfolioProps) {
  const [hideBalances, setHideBalances] = useState(false)
  const [assets] = useState<CryptoAsset[]>([
    {
      symbol: 'BTC',
      name: 'Bitcoin',
      balance: 0.5432,
      usdValue: 23456.78,
      change24h: 2.34,
      price: 43200.50,
      icon: '₿'
    },
    {
      symbol: 'ETH',
      name: 'Ethereum',
      balance: 12.8765,
      usdValue: 28934.12,
      change24h: -1.23,
      price: 2247.89,
      icon: 'Ξ'
    },
    {
      symbol: 'USDC',
      name: 'USD Coin',
      balance: 15000.00,
      usdValue: 15000.00,
      change24h: 0.01,
      price: 1.00,
      icon: '$'
    },
    {
      symbol: 'USDT',
      name: 'Tether',
      balance: 8500.00,
      usdValue: 8500.00,
      change24h: -0.02,
      price: 1.00,
      icon: '₮'
    }
  ])

  const totalValue = assets.reduce((sum, asset) => sum + asset.usdValue, 0)
  const totalChange = assets.reduce((sum, asset) => sum + (asset.usdValue * asset.change24h / 100), 0)
  const totalChangePercent = (totalChange / totalValue) * 100

  const strategies = [
    {
      name: 'Arbitrage Bot',
      status: 'active',
      profit: 1234.56,
      profitPercent: 8.7,
      trades: 45
    },
    {
      name: 'Yield Farming',
      status: 'active',
      profit: 2345.67,
      profitPercent: 12.3,
      trades: 12
    },
    {
      name: 'Grid Trading',
      status: 'paused',
      profit: 567.89,
      profitPercent: 3.2,
      trades: 23
    }
  ]

  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">DeFi Portfolio</h1>
          <p className="text-muted-foreground">
            Manage your cryptocurrency assets and trading strategies
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => setHideBalances(!hideBalances)}
          >
            {hideBalances ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </Button>
          <Button variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Portfolio Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Wallet className="h-5 w-5" />
              Total Portfolio Value
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="text-3xl font-bold">
                {hideBalances ? '••••••' : formatCurrency(totalValue)}
              </div>
              <div className={cn(
                "flex items-center gap-1 text-sm",
                totalChangePercent >= 0 ? "text-green-600" : "text-red-600"
              )}>
                {totalChangePercent >= 0 ? (
                  <ArrowUpRight className="h-4 w-4" />
                ) : (
                  <ArrowDownRight className="h-4 w-4" />
                )}
                {hideBalances ? '••••' : formatPercentage(totalChangePercent)} 
                ({hideBalances ? '••••' : formatCurrency(Math.abs(totalChange))}) 24h
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button className="w-full" variant="coffee">
              Buy Coffee with Crypto
            </Button>
            <Button className="w-full" variant="outline">
              Deposit Funds
            </Button>
            <Button className="w-full" variant="outline">
              Withdraw
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Assets Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <Card>
          <CardHeader>
            <CardTitle>Assets</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {assets.map((asset, index) => (
                <motion.div
                  key={asset.symbol}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="flex items-center justify-between p-3 rounded-lg hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center font-bold">
                      {asset.icon}
                    </div>
                    <div>
                      <div className="font-medium">{asset.symbol}</div>
                      <div className="text-sm text-muted-foreground">{asset.name}</div>
                    </div>
                  </div>
                  
                  <div className="text-right">
                    <div className="font-medium">
                      {hideBalances ? '••••••' : formatCrypto(asset.balance, asset.symbol)}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {hideBalances ? '••••••' : formatCurrency(asset.usdValue)}
                    </div>
                    <div className={cn(
                      "text-xs flex items-center gap-1",
                      asset.change24h >= 0 ? "text-green-600" : "text-red-600"
                    )}>
                      {asset.change24h >= 0 ? (
                        <TrendingUp className="h-3 w-3" />
                      ) : (
                        <TrendingDown className="h-3 w-3" />
                      )}
                      {formatPercentage(asset.change24h)}
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Trading Strategies</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {strategies.map((strategy, index) => (
                <motion.div
                  key={strategy.name}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="p-3 rounded-lg border border-border"
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="font-medium">{strategy.name}</div>
                    <Badge variant={strategy.status === 'active' ? 'success' : 'warning'}>
                      {strategy.status}
                    </Badge>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <div className="text-muted-foreground">Profit</div>
                      <div className="font-medium text-green-600">
                        {hideBalances ? '••••••' : formatCurrency(strategy.profit)}
                      </div>
                    </div>
                    <div>
                      <div className="text-muted-foreground">ROI</div>
                      <div className="font-medium text-green-600">
                        {formatPercentage(strategy.profitPercent)}
                      </div>
                    </div>
                  </div>
                  
                  <div className="mt-2 text-xs text-muted-foreground">
                    {strategy.trades} trades executed
                  </div>
                </motion.div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </motion.div>
  )
}
