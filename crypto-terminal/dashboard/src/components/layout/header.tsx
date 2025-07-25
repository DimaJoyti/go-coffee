'use client'

import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useTradingStore } from '@/stores/trading-store'
import { formatCurrency } from '@/lib/utils'
import {
  Sun,
  Moon,
  Bell,
  User,
  Wifi,
  WifiOff,
  TrendingUp,
  TrendingDown,
} from 'lucide-react'

export function Header() {
  const { theme, setTheme } = useTheme()
  const {
    isConnected,
    portfolios,
    selectedPortfolioId,
    marketOverview,
    signalAlerts,
    riskAlerts,
  } = useTradingStore()

  const selectedPortfolio = portfolios.find(p => p.id === selectedPortfolioId) || portfolios[0]
  const unreadAlerts = signalAlerts.length + riskAlerts.length

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-16 items-center justify-between px-6">
        {/* Left Section - Market Overview */}
        <div className="flex items-center space-x-6">
          {marketOverview && (
            <>
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">Market Cap:</span>
                <span className="text-sm text-muted-foreground">
                  {formatCurrency(marketOverview.totalMarketCap, 'USD', 0)}
                </span>
                {marketOverview.marketCapChange24h > 0 ? (
                  <TrendingUp className="h-4 w-4 text-profit" />
                ) : (
                  <TrendingDown className="h-4 w-4 text-loss" />
                )}
              </div>
              
              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">24h Vol:</span>
                <span className="text-sm text-muted-foreground">
                  {formatCurrency(marketOverview.totalVolume24h, 'USD', 0)}
                </span>
              </div>

              <div className="flex items-center space-x-2">
                <span className="text-sm font-medium">Fear & Greed:</span>
                <Badge
                  variant={
                    marketOverview.fearGreedIndex > 75
                      ? 'profit'
                      : marketOverview.fearGreedIndex > 50
                      ? 'warning'
                      : 'loss'
                  }
                >
                  {marketOverview.fearGreedIndex}
                </Badge>
              </div>
            </>
          )}
        </div>

        {/* Center Section - Portfolio Summary */}
        {selectedPortfolio && (
          <div className="flex items-center space-x-6">
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">Portfolio:</span>
              <span className="text-sm font-semibold">
                {formatCurrency(selectedPortfolio.totalValue)}
              </span>
              <Badge
                variant={selectedPortfolio.totalPnL >= 0 ? 'profit' : 'loss'}
              >
                {selectedPortfolio.totalPnL >= 0 ? '+' : ''}
                {formatCurrency(selectedPortfolio.totalPnL)}
              </Badge>
            </div>
            
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">24h:</span>
              <Badge
                variant={selectedPortfolio.dayChange >= 0 ? 'profit' : 'loss'}
              >
                {selectedPortfolio.dayChange >= 0 ? '+' : ''}
                {selectedPortfolio.dayChangePercent.toFixed(2)}%
              </Badge>
            </div>
          </div>
        )}

        {/* Right Section - Controls */}
        <div className="flex items-center space-x-4">
          {/* Connection Status */}
          <div className="flex items-center space-x-2">
            {isConnected ? (
              <Wifi className="h-4 w-4 text-profit" />
            ) : (
              <WifiOff className="h-4 w-4 text-loss" />
            )}
            <span className="text-xs text-muted-foreground">
              {isConnected ? 'Live' : 'Offline'}
            </span>
          </div>

          {/* Notifications */}
          <Button variant="ghost" size="icon" className="relative">
            <Bell className="h-4 w-4" />
            {unreadAlerts > 0 && (
              <Badge
                variant="destructive"
                className="absolute -top-1 -right-1 h-5 w-5 rounded-full p-0 text-xs"
              >
                {unreadAlerts > 99 ? '99+' : unreadAlerts}
              </Badge>
            )}
          </Button>

          {/* Theme Toggle */}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          >
            <Sun className="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
            <Moon className="absolute h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
            <span className="sr-only">Toggle theme</span>
          </Button>

          {/* User Menu */}
          <Button variant="ghost" size="icon">
            <User className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </header>
  )
}
