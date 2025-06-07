import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Progress } from '@/components/ui/progress';
import { 
  Bot, 
  TrendingUp, 
  TrendingDown, 
  Activity, 
  DollarSign, 
  Target, 
  RefreshCw,
  ExternalLink,
  AlertTriangle,
  CheckCircle,
  Clock
} from 'lucide-react';

interface CommasBot {
  id: string;
  name: string;
  type: string;
  status: string;
  exchange: string;
  totalProfit: number;
  totalProfitPct: number;
  winRate: number;
  activeDeals: number;
  completedDeals: number;
  maxDrawdown: number;
  avgDealTime: string;
}

interface CommasSignal {
  id: string;
  source: string;
  type: string;
  symbol: string;
  exchange: string;
  price: number;
  targetPrice?: number;
  stopLoss?: number;
  confidence: number;
  strength: string;
  riskLevel: string;
  expectedReturn: number;
  description: string;
  createdAt: string;
  status: string;
}

interface CommasDeal {
  id: string;
  botId: string;
  botName: string;
  symbol: string;
  exchange: string;
  status: string;
  type: string;
  averagePrice: number;
  currentPrice: number;
  unrealizedPnL: number;
  totalInvested: number;
  createdAt: string;
}

const CommasIntegrationWidget: React.FC = () => {
  const [bots, setBots] = useState<CommasBot[]>([]);
  const [signals, setSignals] = useState<CommasSignal[]>([]);
  const [deals, setDeals] = useState<CommasDeal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('bots');

  useEffect(() => {
    fetchCommasData();
    const interval = setInterval(fetchCommasData, 60000); // Update every minute
    return () => clearInterval(interval);
  }, []);

  const fetchCommasData = async () => {
    try {
      setLoading(true);
      
      // Fetch 3commas bots
      const botsResponse = await fetch('/api/v2/3commas/bots');
      const botsData = await botsResponse.json();
      setBots(botsData.bots || []);
      
      // Fetch 3commas signals
      const signalsResponse = await fetch('/api/v2/3commas/signals');
      const signalsData = await signalsResponse.json();
      setSignals(signalsData.signals || []);
      
      // Fetch 3commas deals
      const dealsResponse = await fetch('/api/v2/3commas/deals');
      const dealsData = await dealsResponse.json();
      setDeals(dealsData.deals || []);
      
      setError(null);
    } catch (err) {
      setError('Failed to fetch 3commas data');
      console.error('Error fetching 3commas data:', err);
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'enabled':
      case 'active':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'disabled':
      case 'paused':
        return <Clock className="h-4 w-4 text-yellow-500" />;
      case 'error':
      case 'failed':
        return <AlertTriangle className="h-4 w-4 text-red-500" />;
      default:
        return <Activity className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'enabled':
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'disabled':
      case 'paused':
        return 'bg-yellow-100 text-yellow-800';
      case 'error':
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getProfitColor = (profit: number) => {
    if (profit > 0) return 'text-green-600';
    if (profit < 0) return 'text-red-600';
    return 'text-gray-600';
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
  };

  const calculateTotalStats = () => {
    const totalProfit = bots.reduce((sum, bot) => sum + bot.totalProfit, 0);
    const avgWinRate = bots.length > 0 ? bots.reduce((sum, bot) => sum + bot.winRate, 0) / bots.length : 0;
    const activeBots = bots.filter(bot => bot.status === 'enabled').length;
    const totalDeals = bots.reduce((sum, bot) => sum + bot.activeDeals, 0);
    
    return { totalProfit, avgWinRate, activeBots, totalDeals };
  };

  const stats = calculateTotalStats();

  if (loading) {
    return (
      <Card className="w-full">
        <CardContent className="flex items-center justify-center h-64">
          <RefreshCw className="h-8 w-8 animate-spin" />
          <span className="ml-2">Loading 3commas data...</span>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="w-full">
        <CardContent className="flex items-center justify-center h-64">
          <div className="text-center">
            <p className="text-red-500 mb-4">{error}</p>
            <Button onClick={fetchCommasData} variant="outline">
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center">
            <Bot className="h-5 w-5 mr-2" />
            3commas Integration
          </CardTitle>
          <div className="flex items-center space-x-2">
            <Button onClick={fetchCommasData} variant="outline" size="sm">
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button variant="outline" size="sm" asChild>
              <a href="https://3commas.io" target="_blank" rel="noopener noreferrer">
                <ExternalLink className="h-4 w-4" />
              </a>
            </Button>
          </div>
        </div>
        
        {/* Summary Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
          <div className="text-center p-3 bg-blue-50 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">{stats.activeBots}</div>
            <div className="text-sm text-blue-600">Active Bots</div>
          </div>
          <div className="text-center p-3 bg-green-50 rounded-lg">
            <div className={`text-2xl font-bold ${getProfitColor(stats.totalProfit)}`}>
              {formatCurrency(stats.totalProfit)}
            </div>
            <div className="text-sm text-green-600">Total Profit</div>
          </div>
          <div className="text-center p-3 bg-purple-50 rounded-lg">
            <div className="text-2xl font-bold text-purple-600">{stats.avgWinRate.toFixed(1)}%</div>
            <div className="text-sm text-purple-600">Avg Win Rate</div>
          </div>
          <div className="text-center p-3 bg-orange-50 rounded-lg">
            <div className="text-2xl font-bold text-orange-600">{stats.totalDeals}</div>
            <div className="text-sm text-orange-600">Active Deals</div>
          </div>
        </div>
      </CardHeader>
      
      <CardContent>
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="bots">Bots ({bots.length})</TabsTrigger>
            <TabsTrigger value="signals">Signals ({signals.length})</TabsTrigger>
            <TabsTrigger value="deals">Deals ({deals.length})</TabsTrigger>
          </TabsList>

          <TabsContent value="bots" className="space-y-4">
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {bots.map((bot) => (
                <div key={bot.id} className="border rounded-lg p-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(bot.status)}
                      <span className="font-semibold">{bot.name}</span>
                      <Badge className={getStatusColor(bot.status)}>
                        {bot.status}
                      </Badge>
                      <Badge variant="outline">{bot.type}</Badge>
                    </div>
                    <div className="text-right">
                      <div className={`font-semibold ${getProfitColor(bot.totalProfitPct)}`}>
                        {formatPercentage(bot.totalProfitPct)}
                      </div>
                      <div className="text-sm text-gray-600">
                        {formatCurrency(bot.totalProfit)}
                      </div>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-3">
                    <div className="text-center p-2 bg-gray-50 rounded">
                      <div className="text-sm font-medium">Win Rate</div>
                      <div className="text-lg font-semibold text-green-600">
                        {bot.winRate.toFixed(1)}%
                      </div>
                    </div>
                    <div className="text-center p-2 bg-gray-50 rounded">
                      <div className="text-sm font-medium">Active Deals</div>
                      <div className="text-lg font-semibold text-blue-600">
                        {bot.activeDeals}
                      </div>
                    </div>
                    <div className="text-center p-2 bg-gray-50 rounded">
                      <div className="text-sm font-medium">Completed</div>
                      <div className="text-lg font-semibold text-purple-600">
                        {bot.completedDeals}
                      </div>
                    </div>
                    <div className="text-center p-2 bg-gray-50 rounded">
                      <div className="text-sm font-medium">Max DD</div>
                      <div className="text-lg font-semibold text-red-600">
                        {bot.maxDrawdown.toFixed(1)}%
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex items-center justify-between text-sm text-gray-600">
                    <span>Exchange: {bot.exchange}</span>
                    <span>Avg Deal Time: {bot.avgDealTime}</span>
                  </div>
                </div>
              ))}
            </div>
          </TabsContent>

          <TabsContent value="signals" className="space-y-4">
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {signals.map((signal) => (
                <div key={signal.id} className="border rounded-lg p-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      {signal.type === 'buy' ? (
                        <TrendingUp className="h-4 w-4 text-green-500" />
                      ) : (
                        <TrendingDown className="h-4 w-4 text-red-500" />
                      )}
                      <span className="font-semibold">{signal.symbol}</span>
                      <Badge className={signal.type === 'buy' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}>
                        {signal.type.toUpperCase()}
                      </Badge>
                    </div>
                    <div className="text-right">
                      <div className="font-semibold text-blue-600">
                        {signal.confidence}%
                      </div>
                      <div className="text-sm text-gray-600">
                        {signal.strength}
                      </div>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-2 text-sm text-gray-600 mb-2">
                    <div>Price: {formatCurrency(signal.price)}</div>
                    {signal.targetPrice && (
                      <div>Target: {formatCurrency(signal.targetPrice)}</div>
                    )}
                    <div>Return: {formatPercentage(signal.expectedReturn)}</div>
                  </div>
                  
                  <p className="text-sm text-gray-700 mb-2">{signal.description}</p>
                  
                  <div className="flex items-center justify-between">
                    <Badge className={`${signal.riskLevel === 'low' ? 'bg-green-100 text-green-800' : 
                                      signal.riskLevel === 'medium' ? 'bg-yellow-100 text-yellow-800' : 
                                      'bg-red-100 text-red-800'}`}>
                      {signal.riskLevel} risk
                    </Badge>
                    <span className="text-xs text-gray-500">
                      {new Date(signal.createdAt).toLocaleString()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </TabsContent>

          <TabsContent value="deals" className="space-y-4">
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {deals.map((deal) => (
                <div key={deal.id} className="border rounded-lg p-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(deal.status)}
                      <span className="font-semibold">{deal.symbol}</span>
                      <Badge className={getStatusColor(deal.status)}>
                        {deal.status}
                      </Badge>
                      <Badge variant="outline">{deal.type}</Badge>
                    </div>
                    <div className="text-right">
                      <div className={`font-semibold ${getProfitColor(deal.unrealizedPnL)}`}>
                        {formatCurrency(deal.unrealizedPnL)}
                      </div>
                      <div className="text-sm text-gray-600">
                        {formatPercentage((deal.unrealizedPnL / deal.totalInvested) * 100)}
                      </div>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-2 text-sm text-gray-600 mb-2">
                    <div>Avg Price: {formatCurrency(deal.averagePrice)}</div>
                    <div>Current: {formatCurrency(deal.currentPrice)}</div>
                    <div>Invested: {formatCurrency(deal.totalInvested)}</div>
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">Bot: {deal.botName}</span>
                    <span className="text-xs text-gray-500">
                      {new Date(deal.createdAt).toLocaleString()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
};

export default CommasIntegrationWidget;
