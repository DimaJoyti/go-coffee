'use client'

import React, { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import {
  Bitcoin,
  TrendingUp,
  TrendingDown,
  PieChart,
  BarChart3,
  AlertTriangle,
  Shield,
  Zap,
  RefreshCw,
  ArrowUpRight,
  ArrowDownRight
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ResponsiveContainer, PieChart as RechartsPieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip } from 'recharts'
import { formatCurrency, formatNumber } from '@/lib/utils'

interface DeFiPortfolioDashboardProps {
  className?: string
}

interface PortfolioAllocation {
  name: string
  value: number
  percentage: string
}

interface Position {
  id: string
  protocol: string
  chain: string
  tokenSymbol: string
  amount: number
  value: number
  entryPrice: number
  currentPrice: number
  pnl: number
  pnlPercentage: number
  status: string
}

interface YieldFarm {
  id: string
  protocol: string
  pool: string
  apy: number
  tvl: number
  deposited: number
  earned: number
  impermanentLoss: number
  riskLevel: string
}

interface ArbitrageResult {
  id: string
  timestamp: string
  tokenPair: string
  profit: number
  volume: number
  executionTime: number
  success: boolean
  protocol1: string
  protocol2: string
}

const COLORS = ['#8884d8', '#82ca9d', '#ffc658', '#ff7300', '#8dd1e1', '#d084d0']


export function DeFiPortfolioDashboard({ className }: DeFiPortfolioDashboardProps) {
  const [portfolioData, setPortfolioData] = useState<any>(null)
  const [isLoading, setIsLoading] = useState(true)

  // Mock data - in real implementation, this would come from API
  const mockPositions: Position[] = [
    {
      id: '1',
      protocol: 'Uniswap',
      chain: 'Ethereum',
      tokenSymbol: 'ETH',
      amount: 10.5,
      value: 35670,
      entryPrice: 3200,
      currentPrice: 3397,
      pnl: 2067.5,
      pnlPercentage: 6.15,
      status: 'active'
    },
    {
      id: '2',
      protocol: 'Aave',
      chain: 'Polygon',
      tokenSymbol: 'USDC',
      amount: 50000,
      value: 50000,
      entryPrice: 1.0,
      currentPrice: 1.0,
      pnl: 1250,
      pnlPercentage: 2.5,
      status: 'active'
    },
    {
      id: '3',
      protocol: 'Compound',
      chain: 'Ethereum',
      tokenSymbol: 'COMP',
      amount: 125,
      value: 8750,
      entryPrice: 65,
      currentPrice: 70,
      pnl: 625,
      pnlPercentage: 7.69,
      status: 'active'
    }
  ]

  const mockYieldFarms: YieldFarm[] = [
    {
      id: '1',
      protocol: 'Uniswap',
      pool: 'ETH/USDC',
      apy: 24.5,
      tvl: 1250000,
      deposited: 25000,
      earned: 1875,
      impermanentLoss: 125,
      riskLevel: 'medium'
    },
    {
      id: '2',
      protocol: 'Curve',
      pool: 'USDC/USDT',
      apy: 8.2,
      tvl: 850000,
      deposited: 50000,
      earned: 1025,
      impermanentLoss: 0,
      riskLevel: 'low'
    },
    {
      id: '3',
      protocol: 'SushiSwap',
      pool: 'AAVE/ETH',
      apy: 45.8,
      tvl: 125000,
      deposited: 10000,
      earned: 2290,
      impermanentLoss: 450,
      riskLevel: 'high'
    }
  ]

  const mockArbitrageResults: ArbitrageResult[] = [
    {
      id: '1',
      timestamp: '2024-01-15T10:30:00Z',
      tokenPair: 'ETH/USDC',
      profit: 145.67,
      volume: 25000,
      executionTime: 12.5,
      success: true,
      protocol1: 'Uniswap',
      protocol2: 'SushiSwap'
    },
    {
      id: '2',
      timestamp: '2024-01-15T09:45:00Z',
      tokenPair: 'BTC/USDT',
      profit: 287.34,
      volume: 50000,
      executionTime: 8.2,
      success: true,
      protocol1: 'Curve',
      protocol2: '1inch'
    }
  ]

  useEffect(() => {
    // Simulate API call
    setTimeout(() => {
      setPortfolioData({
        totalValue: 94420,
        dailyPnL: 1543.21,
        weeklyPnL: -2341.67,
        monthlyPnL: 8567.89,
        positions: mockPositions,
        yieldFarms: mockYieldFarms,
        arbitrageHistory: mockArbitrageResults,
        riskMetrics: {
          var: 12500,
          sharpe: 1.85,
          maxDrawdown: 8.5,
          volatility: 24.7,
          beta: 1.12,
          alpha: 3.2,
          riskScore: 7
        }
      })
      setIsLoading(false)
    }, 1000)
  }, [])

  const portfolioAllocation = portfolioData?.positions.map((pos: Position) => ({
    name: pos.tokenSymbol,
    value: pos.value,
    percentage: (pos.value / portfolioData.totalValue * 100).toFixed(1)
  })) || []

  const profitLossData = [
    { period: 'Daily', value: portfolioData?.dailyPnL || 0, fill: (portfolioData?.dailyPnL || 0) >= 0 ? '#10b981' : '#ef4444' },
    { period: 'Weekly', value: portfolioData?.weeklyPnL || 0, fill: (portfolioData?.weeklyPnL || 0) >= 0 ? '#10b981' : '#ef4444' },
    { period: 'Monthly', value: portfolioData?.monthlyPnL || 0, fill: (portfolioData?.monthlyPnL || 0) >= 0 ? '#10b981' : '#ef4444' }
  ]


  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardContent className="p-6">
                <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
                <div className="h-8 bg-muted rounded w-1/2"></div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  return (
    <motion.div
      className={`space-y-6 ${className}`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Portfolio Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Total Portfolio</p>
                <p className="text-2xl font-bold">{formatCurrency(portfolioData.totalValue)}</p>
              </div>
              <Bitcoin className="h-8 w-8 text-orange-500" />
            </div>
            <div className="flex items-center mt-2 text-sm">
              <TrendingUp className="h-4 w-4 text-green-500 mr-1" />
              <span className="text-green-600">+12.5%</span>
              <span className="text-muted-foreground ml-1">all time</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Daily P&L</p>
                <p className={`text-2xl font-bold ${portfolioData.dailyPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                  {formatCurrency(portfolioData.dailyPnL)}
                </p>
              </div>
              {portfolioData.dailyPnL >= 0 ? (
                <TrendingUp className="h-8 w-8 text-green-500" />
              ) : (
                <TrendingDown className="h-8 w-8 text-red-500" />
              )}
            </div>
            <div className="flex items-center mt-2 text-sm">
              {portfolioData.dailyPnL >= 0 ? (
                <ArrowUpRight className="h-4 w-4 text-green-500 mr-1" />
              ) : (
                <ArrowDownRight className="h-4 w-4 text-red-500 mr-1" />
              )}
              <span className={portfolioData.dailyPnL >= 0 ? 'text-green-600' : 'text-red-600'}>
                {((portfolioData.dailyPnL / portfolioData.totalValue) * 100).toFixed(2)}%
              </span>
              <span className="text-muted-foreground ml-1">today</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Active Positions</p>
                <p className="text-2xl font-bold">{portfolioData.positions.length}</p>
              </div>
              <BarChart3 className="h-8 w-8 text-blue-500" />
            </div>
            <div className="flex items-center mt-2 text-sm">
              <Shield className="h-4 w-4 text-blue-500 mr-1" />
              <span className="text-blue-600">Risk Score: {portfolioData.riskMetrics.riskScore}/10</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Yield Earned</p>
                <p className="text-2xl font-bold text-green-500">
                  {formatCurrency(mockYieldFarms.reduce((sum, farm) => sum + farm.earned, 0))}
                </p>
              </div>
              <Zap className="h-8 w-8 text-yellow-500" />
            </div>
            <div className="flex items-center mt-2 text-sm">
              <TrendingUp className="h-4 w-4 text-green-500 mr-1" />
              <span className="text-green-600">
                {(mockYieldFarms.reduce((sum, farm) => sum + farm.apy, 0) / mockYieldFarms.length).toFixed(1)}% avg APY
              </span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Portfolio Allocation */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PieChart className="h-5 w-5" />
              Portfolio Allocation
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col lg:flex-row items-center gap-4">
              <div className="w-full lg:w-1/2">
                <ResponsiveContainer width="100%" height={200}>
                  <RechartsPieChart>
                    <Pie
                      data={portfolioAllocation}
                      cx="50%"
                      cy="50%"
                      innerRadius={40}
                      outerRadius={80}
                      paddingAngle={5}
                      dataKey="value"
                    >
                      {portfolioAllocation.map((_entry: PortfolioAllocation, index: number) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value: number) => formatCurrency(value)} />
                  </RechartsPieChart>
                </ResponsiveContainer>
              </div>
              <div className="w-full lg:w-1/2 space-y-2">
                {portfolioAllocation.map((item: PortfolioAllocation, index: number) => (
                  <div key={item.name} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: COLORS[index % COLORS.length] }}
                      />
                      <span className="font-medium">{item.name}</span>
                    </div>
                    <div className="text-right">
                      <div className="font-semibold">{item.percentage}%</div>
                      <div className="text-sm text-muted-foreground">{formatCurrency(item.value)}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* P&L Chart */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Profit & Loss
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={profitLossData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="period" />
                <YAxis />
                <Tooltip formatter={(value: number) => formatCurrency(value)} />
                <Bar 
                  dataKey="value" 
                  fill="#8884d8"
                  radius={[4, 4, 0, 0]}
                >
                  {profitLossData.map((entry: any, index: number) => (
                    <Cell key={`cell-${index}`} fill={entry.fill} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Positions and Yield Farms */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Active Positions */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Active Positions</CardTitle>
              <Button variant="outline" size="sm">
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {portfolioData.positions.map((position: Position) => (
                <div key={position.id} className="p-4 border rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{position.protocol}</Badge>
                      <span className="font-medium">{position.tokenSymbol}</span>
                      <span className="text-sm text-muted-foreground">on {position.chain}</span>
                    </div>
                    <Badge variant={position.pnl >= 0 ? "default" : "destructive"}>
                      {position.pnl >= 0 ? '+' : ''}{position.pnlPercentage.toFixed(2)}%
                    </Badge>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-muted-foreground">Amount</p>
                      <p className="font-medium">{formatNumber(position.amount)} {position.tokenSymbol}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Value</p>
                      <p className="font-medium">{formatCurrency(position.value)}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Entry Price</p>
                      <p className="font-medium">{formatCurrency(position.entryPrice)}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Current Price</p>
                      <p className="font-medium">{formatCurrency(position.currentPrice)}</p>
                    </div>
                  </div>
                  
                  <Separator className="my-2" />
                  
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">P&L:</span>
                    <span className={`font-semibold ${position.pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                      {formatCurrency(position.pnl)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Yield Farming */}
        <Card>
          <CardHeader>
            <CardTitle>Yield Farming</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {mockYieldFarms.map((farm) => (
                <div key={farm.id} className="p-4 border rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{farm.protocol}</Badge>
                      <span className="font-medium">{farm.pool}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary">{farm.apy.toFixed(1)}% APY</Badge>
                      <Badge 
                        variant={farm.riskLevel === 'low' ? 'default' : farm.riskLevel === 'medium' ? 'secondary' : 'destructive'}
                      >
                        {farm.riskLevel} risk
                      </Badge>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4 text-sm mb-3">
                    <div>
                      <p className="text-muted-foreground">Deposited</p>
                      <p className="font-medium">{formatCurrency(farm.deposited)}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground">Earned</p>
                      <p className="font-medium text-green-500">{formatCurrency(farm.earned)}</p>
                    </div>
                  </div>
                  
                  {farm.impermanentLoss > 0 && (
                    <div className="flex items-center gap-2 text-sm">
                      <AlertTriangle className="h-4 w-4 text-amber-500" />
                      <span className="text-amber-600">IL: {formatCurrency(farm.impermanentLoss)}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Arbitrage Results */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5" />
            Recent Arbitrage Results
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {mockArbitrageResults.map((result) => (
              <div key={result.id} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-4">
                  <div className={`w-2 h-2 rounded-full ${result.success ? 'bg-green-500' : 'bg-red-500'}`} />
                  <div>
                    <div className="font-medium">{result.tokenPair}</div>
                    <div className="text-sm text-muted-foreground">
                      {result.protocol1} → {result.protocol2}
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-semibold text-green-500">+{formatCurrency(result.profit)}</div>
                  <div className="text-sm text-muted-foreground">{result.executionTime}s execution</div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}