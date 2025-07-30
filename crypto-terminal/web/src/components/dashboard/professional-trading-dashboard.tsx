import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { DashboardLayout } from '../layout/dashboard-layout';
import { EnhancedTradingViewChart } from '../trading/enhanced-tradingview-chart';
import { EnhancedOrderBook } from '../trading/enhanced-order-book';
import { EnhancedTradingPanel } from '../trading/enhanced-trading-panel';
import { MarketDepthChart } from '../trading/market-depth-chart';
import { EnhancedMarketData } from '../market/enhanced-market-data';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Activity,
  BarChart3,
  Target,
  Zap,
  Settings,
  Maximize2,
  Grid3X3,
  Layout
} from 'lucide-react';

interface Position {
  symbol: string;
  side: 'LONG' | 'SHORT';
  size: number;
  entryPrice: number;
  markPrice: number;
  pnl: number;
  pnlPercent: number;
  margin: number;
  leverage: number;
}

interface TradingStats {
  totalBalance: number;
  availableBalance: number;
  totalPnL: number;
  totalPnLPercent: number;
  openPositions: number;
  todayVolume: number;
}

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1
    }
  }
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 }
};

export const ProfessionalTradingDashboard: React.FC = () => {
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT');
  const [layoutMode, setLayoutMode] = useState<'standard' | 'focus' | 'advanced'>('standard');
  const [isFullscreen, setIsFullscreen] = useState(false);

  // Mock trading stats
  const [tradingStats] = useState<TradingStats>({
    totalBalance: 125750.50,
    availableBalance: 98420.25,
    totalPnL: 3250.75,
    totalPnLPercent: 2.65,
    openPositions: 3,
    todayVolume: 45280.30
  });

  // Mock positions
  const [positions] = useState<Position[]>([
    {
      symbol: 'BTCUSDT',
      side: 'LONG',
      size: 0.125,
      entryPrice: 42000.00,
      markPrice: 43245.75,
      pnl: 155.72,
      pnlPercent: 2.97,
      margin: 525.00,
      leverage: 10
    },
    {
      symbol: 'ETHUSDT',
      side: 'SHORT',
      size: 2.5,
      entryPrice: 2700.00,
      markPrice: 2650.25,
      pnl: 124.38,
      pnlPercent: 1.84,
      margin: 337.50,
      leverage: 8
    }
  ]);

  const handleSymbolSelect = (symbol: string) => {
    setSelectedSymbol(symbol);
  };

  const handleOrderSubmit = (orderData: any) => {
    console.log('Order submitted:', orderData);
    // Here you would integrate with your trading API
  };

  const StatCard: React.FC<{
    title: string;
    value: string | number;
    change?: number;
    icon: React.ReactNode;
    trend?: 'up' | 'down' | 'neutral';
  }> = ({ title, value, change, icon, trend = 'neutral' }) => (
    <Card className="trading-card">
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{title}</p>
            <p className="text-2xl font-bold">
              {typeof value === 'number' ? value.toLocaleString() : value}
            </p>
            {change !== undefined && (
              <p className={`text-sm flex items-center space-x-1 ${
                trend === 'up' ? 'text-green-400' : trend === 'down' ? 'text-red-400' : 'text-muted-foreground'
              }`}>
                {trend === 'up' ? <TrendingUp className="w-3 h-3" /> : 
                 trend === 'down' ? <TrendingDown className="w-3 h-3" /> : null}
                <span>{change >= 0 ? '+' : ''}{change}%</span>
              </p>
            )}
          </div>
          <div className="text-muted-foreground">
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  );

  const PositionsTable: React.FC = () => (
    <Card className="trading-card">
      <CardHeader>
        <CardTitle className="text-lg">Open Positions</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {positions.map((position, index) => (
            <div key={index} className="grid grid-cols-6 gap-4 py-2 text-sm border-b border-border/50 last:border-b-0">
              <div className="font-medium">{position.symbol}</div>
              <div className={`${position.side === 'LONG' ? 'text-green-400' : 'text-red-400'}`}>
                {position.side}
              </div>
              <div className="font-mono">{position.size}</div>
              <div className="font-mono">${position.markPrice.toLocaleString()}</div>
              <div className={`font-mono ${position.pnl >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                ${position.pnl.toFixed(2)}
              </div>
              <div className={`font-mono ${position.pnlPercent >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                {position.pnlPercent >= 0 ? '+' : ''}{position.pnlPercent.toFixed(2)}%
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );

  return (
    <DashboardLayout>
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
        className="space-y-6"
      >
        {/* Header Stats */}
        <motion.div variants={itemVariants}>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <StatCard
              title="Total Balance"
              value={`$${tradingStats.totalBalance.toLocaleString()}`}
              change={tradingStats.totalPnLPercent}
              icon={<DollarSign className="h-6 w-6" />}
              trend={tradingStats.totalPnL >= 0 ? 'up' : 'down'}
            />
            <StatCard
              title="Available Balance"
              value={`$${tradingStats.availableBalance.toLocaleString()}`}
              icon={<Target className="h-6 w-6" />}
            />
            <StatCard
              title="Total P&L"
              value={`$${tradingStats.totalPnL.toFixed(2)}`}
              change={tradingStats.totalPnLPercent}
              icon={<TrendingUp className="h-6 w-6" />}
              trend={tradingStats.totalPnL >= 0 ? 'up' : 'down'}
            />
            <StatCard
              title="Open Positions"
              value={tradingStats.openPositions}
              icon={<Activity className="h-6 w-6" />}
            />
          </div>
        </motion.div>

        {/* Layout Controls */}
        <motion.div variants={itemVariants}>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Button
                variant={layoutMode === 'standard' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setLayoutMode('standard')}
              >
                <Grid3X3 className="w-4 h-4 mr-2" />
                Standard
              </Button>
              <Button
                variant={layoutMode === 'focus' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setLayoutMode('focus')}
              >
                <Layout className="w-4 h-4 mr-2" />
                Focus
              </Button>
              <Button
                variant={layoutMode === 'advanced' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setLayoutMode('advanced')}
              >
                <BarChart3 className="w-4 h-4 mr-2" />
                Advanced
              </Button>
            </div>
            
            <div className="flex items-center space-x-2">
              <Badge variant="outline">
                {selectedSymbol}
              </Badge>
              <Button variant="ghost" size="sm">
                <Settings className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </motion.div>

        {/* Main Trading Interface */}
        {layoutMode === 'standard' && (
          <motion.div variants={itemVariants}>
            <div className="grid grid-cols-1 xl:grid-cols-4 gap-6">
              {/* Chart Section */}
              <div className="xl:col-span-3 space-y-6">
                <EnhancedTradingViewChart
                  symbol={selectedSymbol}
                  height={500}
                  showDrawingTools={true}
                  showIndicators={true}
                />
                
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <MarketDepthChart symbol={selectedSymbol} />
                  <PositionsTable />
                </div>
              </div>

              {/* Side Panel */}
              <div className="space-y-6">
                <EnhancedTradingPanel
                  symbol={selectedSymbol}
                  onOrderSubmit={handleOrderSubmit}
                />
                <EnhancedOrderBook symbol={selectedSymbol} />
              </div>
            </div>
          </motion.div>
        )}

        {layoutMode === 'focus' && (
          <motion.div variants={itemVariants}>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <div className="lg:col-span-2">
                <EnhancedTradingViewChart
                  symbol={selectedSymbol}
                  height={600}
                  showDrawingTools={true}
                  showIndicators={true}
                />
              </div>
              <div className="space-y-6">
                <EnhancedTradingPanel
                  symbol={selectedSymbol}
                  onOrderSubmit={handleOrderSubmit}
                />
                <EnhancedOrderBook symbol={selectedSymbol} maxLevels={10} />
              </div>
            </div>
          </motion.div>
        )}

        {layoutMode === 'advanced' && (
          <motion.div variants={itemVariants}>
            <div className="grid grid-cols-1 xl:grid-cols-6 gap-6">
              {/* Market Data */}
              <div className="xl:col-span-2">
                <EnhancedMarketData onSymbolSelect={handleSymbolSelect} />
              </div>
              
              {/* Chart */}
              <div className="xl:col-span-3">
                <EnhancedTradingViewChart
                  symbol={selectedSymbol}
                  height={500}
                  showDrawingTools={true}
                  showIndicators={true}
                />
              </div>
              
              {/* Trading Panel */}
              <div className="space-y-6">
                <EnhancedTradingPanel
                  symbol={selectedSymbol}
                  onOrderSubmit={handleOrderSubmit}
                />
              </div>
            </div>
            
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-6">
              <EnhancedOrderBook symbol={selectedSymbol} />
              <MarketDepthChart symbol={selectedSymbol} />
              <PositionsTable />
            </div>
          </motion.div>
        )}
      </motion.div>
    </DashboardLayout>
  );
};
