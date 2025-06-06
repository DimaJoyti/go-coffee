import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  PieChart, 
  BarChart3, 
  Shield, 
  AlertTriangle,
  RefreshCw,
  Eye,
  Target,
  Activity
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface PortfolioHolding {
  symbol: string;
  name: string;
  quantity: number;
  avgCost: number;
  currentPrice: number;
  marketValue: number;
  unrealizedPnL: number;
  unrealizedPct: number;
  weight: number;
  dayChange: number;
  dayChangePct: number;
  lastUpdated: string;
}

interface PortfolioAnalytics {
  portfolioId: string;
  totalValue: number;
  totalReturn: number;
  totalReturnPct: number;
  dayReturn: number;
  dayReturnPct: number;
  holdings: PortfolioHolding[];
  allocation: Record<string, number>;
  performance: {
    sharpeRatio: number;
    maxDrawdown: number;
    volatility: number;
    winRate: number;
  };
  risk: {
    var95: number;
    portfolioVol: number;
    riskScore: number;
  };
  lastUpdated: string;
}

interface PortfolioManagerDashboardProps {
  portfolioId: string;
  className?: string;
}

const PortfolioManagerDashboard: React.FC<PortfolioManagerDashboardProps> = ({ 
  portfolioId, 
  className 
}) => {
  const [analytics, setAnalytics] = useState<PortfolioAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const fetchPortfolioAnalytics = async () => {
    try {
      setLoading(true);
      
      const response = await fetch(`/api/v2/portfolio/${portfolioId}/analytics`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch portfolio analytics');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setAnalytics(data.data);
        setLastUpdate(new Date());
        setError(null);
      } else {
        throw new Error(data.error || 'Unknown error');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPortfolioAnalytics();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchPortfolioAnalytics, 30000);
    
    return () => clearInterval(interval);
  }, [portfolioId]);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const formatPercent = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
  };

  const getChangeColor = (value: number) => {
    if (value > 0) return 'text-green-600';
    if (value < 0) return 'text-red-600';
    return 'text-gray-600';
  };

  const getRiskColor = (score: number) => {
    if (score <= 3) return 'text-green-600';
    if (score <= 6) return 'text-yellow-600';
    return 'text-red-600';
  };

  if (loading && !analytics) {
    return (
      <Card className={cn('w-full', className)}>
        <CardContent className="flex items-center justify-center py-8">
          <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
          <span className="ml-2 text-gray-500">Loading portfolio analytics...</span>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className={cn('w-full', className)}>
        <CardContent className="flex items-center gap-2 p-4 bg-red-50 border border-red-200 rounded-lg">
          <AlertTriangle className="h-5 w-5 text-red-600" />
          <span className="text-red-700">{error}</span>
        </CardContent>
      </Card>
    );
  }

  if (!analytics) {
    return null;
  }

  return (
    <div className={cn('w-full space-y-6', className)}>
      {/* Header */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <CardTitle className="text-2xl font-bold flex items-center gap-2">
            <PieChart className="h-6 w-6 text-blue-600" />
            Portfolio Manager Dashboard
          </CardTitle>
          <div className="flex items-center gap-3">
            {lastUpdate && (
              <span className="text-sm text-gray-500">
                Last updated: {lastUpdate.toLocaleTimeString()}
              </span>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={fetchPortfolioAnalytics}
              disabled={loading}
              className="flex items-center gap-1"
            >
              <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              Refresh
            </Button>
          </div>
        </CardHeader>
      </Card>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Value</p>
                <p className="text-2xl font-bold">{formatCurrency(analytics.totalValue)}</p>
              </div>
              <DollarSign className="h-8 w-8 text-blue-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Return</p>
                <p className={cn('text-2xl font-bold', getChangeColor(analytics.totalReturn))}>
                  {formatCurrency(analytics.totalReturn)}
                </p>
                <p className={cn('text-sm', getChangeColor(analytics.totalReturnPct))}>
                  {formatPercent(analytics.totalReturnPct)}
                </p>
              </div>
              {analytics.totalReturn >= 0 ? (
                <TrendingUp className="h-8 w-8 text-green-600" />
              ) : (
                <TrendingDown className="h-8 w-8 text-red-600" />
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Day Return</p>
                <p className={cn('text-2xl font-bold', getChangeColor(analytics.dayReturn))}>
                  {formatCurrency(analytics.dayReturn)}
                </p>
                <p className={cn('text-sm', getChangeColor(analytics.dayReturnPct))}>
                  {formatPercent(analytics.dayReturnPct)}
                </p>
              </div>
              <Activity className="h-8 w-8 text-purple-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Risk Score</p>
                <p className={cn('text-2xl font-bold', getRiskColor(analytics.risk.riskScore))}>
                  {analytics.risk.riskScore.toFixed(1)}/10
                </p>
                <p className="text-sm text-gray-500">
                  VaR: {formatCurrency(analytics.risk.var95)}
                </p>
              </div>
              <Shield className="h-8 w-8 text-orange-600" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Detailed Analytics Tabs */}
      <Tabs defaultValue="holdings" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="holdings">Holdings</TabsTrigger>
          <TabsTrigger value="allocation">Allocation</TabsTrigger>
          <TabsTrigger value="performance">Performance</TabsTrigger>
          <TabsTrigger value="risk">Risk Analysis</TabsTrigger>
        </TabsList>

        <TabsContent value="holdings" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Portfolio Holdings</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {analytics.holdings.map((holding) => (
                  <div
                    key={holding.symbol}
                    className="flex items-center justify-between p-4 border border-gray-200 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center">
                        <span className="text-sm font-bold">{holding.symbol}</span>
                      </div>
                      <div>
                        <p className="font-medium">{holding.name}</p>
                        <p className="text-sm text-gray-500">
                          {holding.quantity.toFixed(4)} @ {formatCurrency(holding.avgCost)}
                        </p>
                      </div>
                    </div>
                    
                    <div className="text-right">
                      <p className="font-medium">{formatCurrency(holding.marketValue)}</p>
                      <p className={cn('text-sm', getChangeColor(holding.unrealizedPnL))}>
                        {formatCurrency(holding.unrealizedPnL)} ({formatPercent(holding.unrealizedPct)})
                      </p>
                      <p className="text-xs text-gray-500">{holding.weight.toFixed(1)}% weight</p>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="allocation" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Asset Allocation</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {Object.entries(analytics.allocation).map(([symbol, percentage]) => (
                  <div key={symbol} className="space-y-2">
                    <div className="flex justify-between">
                      <span className="font-medium">{symbol}</span>
                      <span className="text-sm text-gray-600">{percentage.toFixed(1)}%</span>
                    </div>
                    <Progress value={percentage} className="h-2" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="performance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Performance Metrics</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-4">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Sharpe Ratio</span>
                    <span className="font-medium">{analytics.performance.sharpeRatio.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Max Drawdown</span>
                    <span className="font-medium text-red-600">
                      {analytics.performance.maxDrawdown.toFixed(2)}%
                    </span>
                  </div>
                </div>
                <div className="space-y-4">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Volatility</span>
                    <span className="font-medium">{analytics.performance.volatility.toFixed(2)}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Win Rate</span>
                    <span className="font-medium text-green-600">
                      {analytics.performance.winRate.toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="risk" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Risk Analysis</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                <div className="grid grid-cols-2 gap-6">
                  <div>
                    <p className="text-sm text-gray-600 mb-2">Value at Risk (95%)</p>
                    <p className="text-2xl font-bold text-red-600">
                      {formatCurrency(analytics.risk.var95)}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600 mb-2">Portfolio Volatility</p>
                    <p className="text-2xl font-bold">
                      {(analytics.risk.portfolioVol * 100).toFixed(1)}%
                    </p>
                  </div>
                </div>
                
                <div className="p-4 bg-gray-50 rounded-lg">
                  <div className="flex items-center gap-2 mb-2">
                    <Target className="h-5 w-5 text-blue-600" />
                    <span className="font-medium">Risk Score</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <Progress 
                      value={(analytics.risk.riskScore / 10) * 100} 
                      className="flex-1 h-3"
                    />
                    <span className={cn('font-bold', getRiskColor(analytics.risk.riskScore))}>
                      {analytics.risk.riskScore.toFixed(1)}/10
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 mt-2">
                    {analytics.risk.riskScore <= 3 && "Low Risk - Conservative portfolio"}
                    {analytics.risk.riskScore > 3 && analytics.risk.riskScore <= 6 && "Medium Risk - Balanced portfolio"}
                    {analytics.risk.riskScore > 6 && "High Risk - Aggressive portfolio"}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default PortfolioManagerDashboard;
