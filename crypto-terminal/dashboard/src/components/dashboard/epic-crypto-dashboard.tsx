'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { MarketOverview } from '@/components/market/market-overview'
import { PortfolioSummary } from '@/components/portfolio/portfolio-summary'
import { TradingViewChart } from '@/components/trading/tradingview-chart'
import { OrderBook } from '@/components/trading/order-book'
import { MarketHeatmap } from '@/components/market/market-heatmap'
import { PriceAlerts } from '@/components/trading/price-alerts'
import { NewsWidget } from '@/components/market/news-widget'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  TrendingUp,
  Activity,
  DollarSign,
  BarChart3,
  Zap,
  Globe,
  Shield
} from 'lucide-react'

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.2
    }
  }
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.5,
      ease: "easeOut"
    }
  }
}

export function EpicCryptoDashboard() {
  const [isConnected, setIsConnected] = useState(false)
  const [marketStatus] = useState<'open' | 'closed' | 'pre-market'>('open')

  useEffect(() => {
    // Simulate connection status
    const timer = setTimeout(() => setIsConnected(true), 1000)
    return () => clearTimeout(timer)
  }, [])

  return (
    <DashboardLayout>
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
        className="space-y-6 p-6"
      >
        {/* Epic Header */}
        <motion.div variants={itemVariants} className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-4xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
              Epic Crypto Terminal
            </h1>
            <p className="text-muted-foreground">
              Professional cryptocurrency trading dashboard with real-time market data
            </p>
          </div>
          
          <div className="flex items-center space-x-4">
            <Badge 
              variant={isConnected ? "default" : "destructive"}
              className="animate-pulse"
            >
              <Activity className="w-3 h-3 mr-1" />
              {isConnected ? 'Connected' : 'Connecting...'}
            </Badge>
            
            <Badge variant="outline">
              <Globe className="w-3 h-3 mr-1" />
              {marketStatus === 'open' ? 'Market Open' : 'Market Closed'}
            </Badge>
          </div>
        </motion.div>

        {/* Quick Stats */}
        <motion.div variants={itemVariants}>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card variant="epic" className="group hover:scale-105 transition-transform">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Portfolio</p>
                    <p className="text-2xl font-bold">$124,567.89</p>
                    <p className="text-xs text-profit flex items-center">
                      <TrendingUp className="w-3 h-3 mr-1" />
                      +2.34% (24h)
                    </p>
                  </div>
                  <DollarSign className="w-8 h-8 text-primary opacity-60" />
                </div>
              </CardContent>
            </Card>

            <Card variant="epic" className="group hover:scale-105 transition-transform">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">24h P&L</p>
                    <p className="text-2xl font-bold text-profit">+$2,891.23</p>
                    <p className="text-xs text-profit flex items-center">
                      <TrendingUp className="w-3 h-3 mr-1" />
                      +2.34%
                    </p>
                  </div>
                  <BarChart3 className="w-8 h-8 text-profit opacity-60" />
                </div>
              </CardContent>
            </Card>

            <Card variant="epic" className="group hover:scale-105 transition-transform">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Active Positions</p>
                    <p className="text-2xl font-bold">12</p>
                    <p className="text-xs text-muted-foreground">8 profitable</p>
                  </div>
                  <Activity className="w-8 h-8 text-primary opacity-60" />
                </div>
              </CardContent>
            </Card>

            <Card variant="epic" className="group hover:scale-105 transition-transform">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Risk Score</p>
                    <p className="text-2xl font-bold text-warning">Medium</p>
                    <p className="text-xs text-muted-foreground">7.2/10</p>
                  </div>
                  <Shield className="w-8 h-8 text-warning opacity-60" />
                </div>
              </CardContent>
            </Card>
          </div>
        </motion.div>

        {/* Main Trading Interface */}
        <motion.div variants={itemVariants}>
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
            {/* Chart Section */}
            <div className="xl:col-span-2 space-y-6">
              <Card variant="glass">
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span>BTC/USD</span>
                    <div className="flex items-center space-x-2">
                      <Badge variant="outline">1D</Badge>
                      <Button variant="epic" size="sm">
                        <Zap className="w-4 h-4 mr-1" />
                        Trade
                      </Button>
                    </div>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <TradingViewChart symbol="BTCUSD" />
                </CardContent>
              </Card>

              <MarketHeatmap />
            </div>

            {/* Side Panel */}
            <div className="space-y-6">
              <OrderBook symbol="BTCUSD" />
              <PriceAlerts />
              <NewsWidget />
            </div>
          </div>
        </motion.div>

        {/* Bottom Section */}
        <motion.div variants={itemVariants}>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <PortfolioSummary />
            <MarketOverview />
          </div>
        </motion.div>
      </motion.div>
    </DashboardLayout>
  )
}
