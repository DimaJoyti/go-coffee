import React, { useState, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { 
  TrendingUp, 
  TrendingDown, 
  BarChart3,
  Target,
  Activity,
  Calendar,
  DollarSign,
  Percent,
  Award,
  Clock
} from 'lucide-react';

interface PerformanceMetric {
  name: string;
  value: number;
  benchmark?: number;
  period: string;
  trend: 'up' | 'down' | 'neutral';
  format: 'currency' | 'percentage' | 'number' | 'ratio';
}

interface TradeAnalysis {
  totalTrades: number;
  winningTrades: number;
  losingTrades: number;
  winRate: number;
  avgWin: number;
  avgLoss: number;
  profitFactor: number;
  largestWin: number;
  largestLoss: number;
  avgHoldTime: number;
}

interface PeriodPerformance {
  period: string;
  returns: number;
  volatility: number;
  sharpeRatio: number;
  maxDrawdown: number;
  trades: number;
}

interface PerformanceAnalyticsProps {
  className?: string;
}

// Mock data
const mockMetrics: PerformanceMetric[] = [
  {
    name: 'Total Return',
    value: 15.47,
    benchmark: 12.30,
    period: 'YTD',
    trend: 'up',
    format: 'percentage'
  },
  {
    name: 'Sharpe Ratio',
    value: 1.85,
    benchmark: 1.20,
    period: '1Y',
    trend: 'up',
    format: 'ratio'
  },
  {
    name: 'Max Drawdown',
    value: -8.2,
    benchmark: -12.5,
    period: '1Y',
    trend: 'up',
    format: 'percentage'
  },
  {
    name: 'Alpha',
    value: 3.17,
    period: '1Y',
    trend: 'up',
    format: 'percentage'
  },
  {
    name: 'Beta',
    value: 1.15,
    benchmark: 1.00,
    period: '1Y',
    trend: 'neutral',
    format: 'ratio'
  },
  {
    name: 'Volatility',
    value: 24.8,
    benchmark: 28.5,
    period: '1Y',
    trend: 'up',
    format: 'percentage'
  }
];

const mockTradeAnalysis: TradeAnalysis = {
  totalTrades: 247,
  winningTrades: 169,
  losingTrades: 78,
  winRate: 68.4,
  avgWin: 2.85,
  avgLoss: -1.42,
  profitFactor: 2.01,
  largestWin: 15.6,
  largestLoss: -8.3,
  avgHoldTime: 4.2
};

const mockPeriodPerformance: PeriodPerformance[] = [
  {
    period: '1D',
    returns: 2.34,
    volatility: 18.5,
    sharpeRatio: 1.26,
    maxDrawdown: -1.2,
    trades: 8
  },
  {
    period: '1W',
    returns: 5.67,
    volatility: 22.1,
    sharpeRatio: 1.45,
    maxDrawdown: -3.4,
    trades: 23
  },
  {
    period: '1M',
    returns: 12.34,
    volatility: 24.8,
    sharpeRatio: 1.67,
    maxDrawdown: -5.8,
    trades: 89
  },
  {
    period: '3M',
    returns: 18.92,
    volatility: 26.2,
    sharpeRatio: 1.72,
    maxDrawdown: -8.2,
    trades: 247
  },
  {
    period: '1Y',
    returns: 45.67,
    volatility: 28.5,
    sharpeRatio: 1.85,
    maxDrawdown: -12.1,
    trades: 892
  }
];

export const PerformanceAnalytics: React.FC<PerformanceAnalyticsProps> = ({ className }) => {
  const [selectedPeriod, setSelectedPeriod] = useState('1Y');

  const formatValue = (value: number, format: string) => {
    switch (format) {
      case 'currency':
        return `$${value.toLocaleString()}`;
      case 'percentage':
        return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
      case 'ratio':
        return value.toFixed(2);
      case 'number':
        return value.toLocaleString();
      default:
        return value.toString();
    }
  };

  const getTrendColor = (trend: string) => {
    switch (trend) {
      case 'up': return 'text-green-400';
      case 'down': return 'text-red-400';
      default: return 'text-muted-foreground';
    }
  };

  const MetricCard: React.FC<{ metric: PerformanceMetric }> = ({ metric }) => (
    <Card className="trading-card">
      <CardContent className="p-4">
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h4 className="font-medium text-sm">{metric.name}</h4>
            <Badge variant="outline" className="text-xs">
              {metric.period}
            </Badge>
          </div>
          
          <div className="space-y-1">
            <div className={`text-xl font-bold ${getTrendColor(metric.trend)}`}>
              {formatValue(metric.value, metric.format)}
            </div>
            
            {metric.benchmark && (
              <div className="flex items-center space-x-2 text-xs">
                <span className="text-muted-foreground">vs Benchmark:</span>
                <span className={metric.value > metric.benchmark ? 'text-green-400' : 'text-red-400'}>
                  {formatValue(metric.benchmark, metric.format)}
                </span>
                <span className={metric.value > metric.benchmark ? 'text-green-400' : 'text-red-400'}>
                  ({metric.value > metric.benchmark ? '+' : ''}{(metric.value - metric.benchmark).toFixed(2)})
                </span>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );

  const PeriodRow: React.FC<{ period: PeriodPerformance }> = ({ period }) => (
    <div className="grid grid-cols-6 gap-4 py-3 text-sm border-b border-border/50 last:border-b-0">
      <div className="font-medium">{period.period}</div>
      <div className={`font-mono ${period.returns >= 0 ? 'text-green-400' : 'text-red-400'}`}>
        {period.returns >= 0 ? '+' : ''}{period.returns.toFixed(2)}%
      </div>
      <div className="font-mono">{period.volatility.toFixed(1)}%</div>
      <div className="font-mono">{period.sharpeRatio.toFixed(2)}</div>
      <div className="font-mono text-red-400">{period.maxDrawdown.toFixed(1)}%</div>
      <div className="font-mono">{period.trades}</div>
    </div>
  );

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Performance Analytics</h2>
        <div className="flex items-center space-x-2">
          {['1D', '1W', '1M', '3M', '1Y'].map((period) => (
            <Button
              key={period}
              variant={selectedPeriod === period ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSelectedPeriod(period)}
              className="h-8 px-3 text-xs"
            >
              {period}
            </Button>
          ))}
        </div>
      </div>

      {/* Performance Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {mockMetrics.map((metric, index) => (
          <MetricCard key={index} metric={metric} />
        ))}
      </div>

      {/* Detailed Analytics */}
      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="trades">Trade Analysis</TabsTrigger>
          <TabsTrigger value="periods">Period Performance</TabsTrigger>
          <TabsTrigger value="attribution">Attribution</TabsTrigger>
        </TabsList>
        
        <TabsContent value="overview" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card className="trading-card">
              <CardHeader>
                <CardTitle className="text-lg flex items-center space-x-2">
                  <Award className="w-5 h-5" />
                  <span>Key Achievements</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Best Month</span>
                    <span className="font-mono text-green-400">+23.4%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Longest Win Streak</span>
                    <span className="font-mono">12 trades</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Best Single Trade</span>
                    <span className="font-mono text-green-400">+15.6%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Recovery Time</span>
                    <span className="font-mono">3.2 days</span>
                  </div>
                </div>
              </CardContent>
            </Card>
            
            <Card className="trading-card">
              <CardHeader>
                <CardTitle className="text-lg flex items-center space-x-2">
                  <Target className="w-5 h-5" />
                  <span>Risk Metrics</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Value at Risk (95%)</span>
                    <span className="font-mono text-red-400">-$2,389</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Conditional VaR</span>
                    <span className="font-mono text-red-400">-$3,567</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Calmar Ratio</span>
                    <span className="font-mono">1.28</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Sortino Ratio</span>
                    <span className="font-mono">2.14</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
        
        <TabsContent value="trades" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card className="trading-card">
              <CardHeader>
                <CardTitle className="text-lg">Trade Statistics</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Total Trades</span>
                    <span className="font-mono">{mockTradeAnalysis.totalTrades}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Winning Trades</span>
                    <span className="font-mono text-green-400">{mockTradeAnalysis.winningTrades}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Losing Trades</span>
                    <span className="font-mono text-red-400">{mockTradeAnalysis.losingTrades}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Win Rate</span>
                    <span className="font-mono">{mockTradeAnalysis.winRate.toFixed(1)}%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Profit Factor</span>
                    <span className="font-mono">{mockTradeAnalysis.profitFactor.toFixed(2)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
            
            <Card className="trading-card">
              <CardHeader>
                <CardTitle className="text-lg">Trade Performance</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Average Win</span>
                    <span className="font-mono text-green-400">+{mockTradeAnalysis.avgWin.toFixed(2)}%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Average Loss</span>
                    <span className="font-mono text-red-400">{mockTradeAnalysis.avgLoss.toFixed(2)}%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Largest Win</span>
                    <span className="font-mono text-green-400">+{mockTradeAnalysis.largestWin.toFixed(1)}%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Largest Loss</span>
                    <span className="font-mono text-red-400">{mockTradeAnalysis.largestLoss.toFixed(1)}%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Avg Hold Time</span>
                    <span className="font-mono">{mockTradeAnalysis.avgHoldTime.toFixed(1)} days</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
        
        <TabsContent value="periods" className="space-y-4">
          <Card className="trading-card">
            <CardHeader>
              <CardTitle className="text-lg">Performance by Period</CardTitle>
            </CardHeader>
            <CardContent>
              {/* Table Header */}
              <div className="grid grid-cols-6 gap-4 py-2 text-xs font-medium text-muted-foreground border-b border-border">
                <div>Period</div>
                <div>Returns</div>
                <div>Volatility</div>
                <div>Sharpe</div>
                <div>Max DD</div>
                <div>Trades</div>
              </div>
              
              {/* Period Rows */}
              <div className="space-y-0">
                {mockPeriodPerformance.map((period) => (
                  <PeriodRow key={period.period} period={period} />
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="attribution" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card className="trading-card">
              <CardHeader>
                <CardTitle className="text-lg">Asset Contribution</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">BTC</span>
                    <span className="font-mono text-green-400">+8.2%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">ETH</span>
                    <span className="font-mono text-green-400">+5.1%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">SOL</span>
                    <span className="font-mono text-green-400">+2.2%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Others</span>
                    <span className="font-mono text-red-400">-0.1%</span>
                  </div>
                </div>
              </CardContent>
            </Card>
            
            <Card className="trading-card">
              <CardHeader>
                <CardTitle className="text-lg">Strategy Attribution</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Trend Following</span>
                    <span className="font-mono text-green-400">+6.8%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Mean Reversion</span>
                    <span className="font-mono text-green-400">+4.2%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Arbitrage</span>
                    <span className="font-mono text-green-400">+2.1%</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Market Making</span>
                    <span className="font-mono text-green-400">+2.3%</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
};
