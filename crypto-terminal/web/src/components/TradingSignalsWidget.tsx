import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { TrendingUp, TrendingDown, Activity, Filter, Search, RefreshCw } from 'lucide-react';

interface TradingSignal {
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
  timeFrame: string;
  strategy: string;
  riskLevel: string;
  expectedReturn: number;
  description: string;
  tags: string[];
  createdAt: string;
  status: string;
}

interface TradingBot {
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
}

interface TechnicalAnalysis {
  symbol: string;
  exchange: string;
  timeFrame: string;
  overallSignal: string;
  overallScore: number;
  trendDirection: string;
  trendStrength: number;
  indicators: TechnicalIndicator[];
}

interface TechnicalIndicator {
  name: string;
  value: number;
  signal: string;
  strength: number;
}

const TradingSignalsWidget: React.FC = () => {
  const [signals, setSignals] = useState<TradingSignal[]>([]);
  const [bots, setBots] = useState<TradingBot[]>([]);
  const [analysis, setAnalysis] = useState<Record<string, TechnicalAnalysis>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('signals');
  
  // Filter states
  const [searchTerm, setSearchTerm] = useState('');
  const [sourceFilter, setSourceFilter] = useState('all');
  const [typeFilter, setTypeFilter] = useState('all');
  const [riskFilter, setRiskFilter] = useState('all');

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000); // Update every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      // Fetch trading signals
      const signalsResponse = await fetch('/api/v2/trading/signals?limit=50');
      const signalsData = await signalsResponse.json();
      setSignals(signalsData.signals || []);
      
      // Fetch trading bots
      const botsResponse = await fetch('/api/v2/trading/bots?limit=20');
      const botsData = await botsResponse.json();
      setBots(botsData.bots || []);
      
      // Fetch technical analysis
      const analysisResponse = await fetch('/api/v2/trading/analysis');
      const analysisData = await analysisResponse.json();
      setAnalysis(analysisData.analysis || {});
      
      setError(null);
    } catch (err) {
      setError('Failed to fetch trading data');
      console.error('Error fetching trading data:', err);
    } finally {
      setLoading(false);
    }
  };

  const getSignalIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case 'buy':
      case 'long':
        return <TrendingUp className="h-4 w-4 text-green-500" />;
      case 'sell':
      case 'short':
        return <TrendingDown className="h-4 w-4 text-red-500" />;
      default:
        return <Activity className="h-4 w-4 text-yellow-500" />;
    }
  };

  const getSignalColor = (type: string) => {
    switch (type.toLowerCase()) {
      case 'buy':
      case 'long':
        return 'bg-green-100 text-green-800';
      case 'sell':
      case 'short':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-yellow-100 text-yellow-800';
    }
  };

  const getRiskColor = (risk: string) => {
    switch (risk.toLowerCase()) {
      case 'low':
        return 'bg-green-100 text-green-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'high':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 80) return 'text-green-600';
    if (confidence >= 60) return 'text-yellow-600';
    return 'text-red-600';
  };

  const filteredSignals = signals.filter(signal => {
    const matchesSearch = signal.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         signal.description.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesSource = sourceFilter === 'all' || signal.source === sourceFilter;
    const matchesType = typeFilter === 'all' || signal.type === typeFilter;
    const matchesRisk = riskFilter === 'all' || signal.riskLevel === riskFilter;
    
    return matchesSearch && matchesSource && matchesType && matchesRisk;
  });

  if (loading) {
    return (
      <Card className="w-full">
        <CardContent className="flex items-center justify-center h-64">
          <RefreshCw className="h-8 w-8 animate-spin" />
          <span className="ml-2">Loading trading data...</span>
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
            <Button onClick={fetchData} variant="outline">
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
            <Activity className="h-5 w-5 mr-2" />
            Trading Intelligence
          </CardTitle>
          <Button onClick={fetchData} variant="outline" size="sm">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="signals">Signals ({signals.length})</TabsTrigger>
            <TabsTrigger value="bots">Bots ({bots.length})</TabsTrigger>
            <TabsTrigger value="analysis">Analysis ({Object.keys(analysis).length})</TabsTrigger>
          </TabsList>

          <TabsContent value="signals" className="space-y-4">
            {/* Filters */}
            <div className="flex flex-wrap gap-4 p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center space-x-2">
                <Search className="h-4 w-4" />
                <Input
                  placeholder="Search symbols..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-48"
                />
              </div>
              
              <Select value={sourceFilter} onValueChange={setSourceFilter}>
                <SelectTrigger className="w-32">
                  <SelectValue placeholder="Source" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Sources</SelectItem>
                  <SelectItem value="3commas">3commas</SelectItem>
                  <SelectItem value="tradingview">TradingView</SelectItem>
                  <SelectItem value="custom">Custom</SelectItem>
                </SelectContent>
              </Select>

              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger className="w-32">
                  <SelectValue placeholder="Type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  <SelectItem value="buy">Buy</SelectItem>
                  <SelectItem value="sell">Sell</SelectItem>
                  <SelectItem value="hold">Hold</SelectItem>
                </SelectContent>
              </Select>

              <Select value={riskFilter} onValueChange={setRiskFilter}>
                <SelectTrigger className="w-32">
                  <SelectValue placeholder="Risk" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Risk</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Signals List */}
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {filteredSignals.map((signal) => (
                <div key={signal.id} className="border rounded-lg p-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      {getSignalIcon(signal.type)}
                      <span className="font-semibold">{signal.symbol}</span>
                      <Badge className={getSignalColor(signal.type)}>
                        {signal.type.toUpperCase()}
                      </Badge>
                      <Badge variant="outline">{signal.source}</Badge>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className={`font-semibold ${getConfidenceColor(signal.confidence)}`}>
                        {signal.confidence}%
                      </span>
                      <Badge className={getRiskColor(signal.riskLevel)}>
                        {signal.riskLevel}
                      </Badge>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm text-gray-600 mb-2">
                    <div>Price: ${signal.price.toFixed(4)}</div>
                    {signal.targetPrice && (
                      <div>Target: ${signal.targetPrice.toFixed(4)}</div>
                    )}
                    <div>Return: {signal.expectedReturn.toFixed(2)}%</div>
                    <div>Timeframe: {signal.timeFrame}</div>
                  </div>
                  
                  <p className="text-sm text-gray-700 mb-2">{signal.description}</p>
                  
                  <div className="flex flex-wrap gap-1">
                    {signal.tags.map((tag, index) => (
                      <Badge key={index} variant="secondary" className="text-xs">
                        {tag}
                      </Badge>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </TabsContent>

          <TabsContent value="bots" className="space-y-4">
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {bots.map((bot) => (
                <div key={bot.id} className="border rounded-lg p-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <span className="font-semibold">{bot.name}</span>
                      <Badge variant={bot.status === 'enabled' ? 'default' : 'secondary'}>
                        {bot.status}
                      </Badge>
                      <Badge variant="outline">{bot.exchange}</Badge>
                    </div>
                    <div className="text-right">
                      <div className={`font-semibold ${bot.totalProfitPct >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        {bot.totalProfitPct >= 0 ? '+' : ''}{bot.totalProfitPct.toFixed(2)}%
                      </div>
                      <div className="text-sm text-gray-600">
                        ${bot.totalProfit.toFixed(2)}
                      </div>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-3 gap-4 text-sm text-gray-600">
                    <div>
                      <div className="font-medium">Win Rate</div>
                      <div>{bot.winRate.toFixed(1)}%</div>
                    </div>
                    <div>
                      <div className="font-medium">Active Deals</div>
                      <div>{bot.activeDeals}</div>
                    </div>
                    <div>
                      <div className="font-medium">Completed</div>
                      <div>{bot.completedDeals}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </TabsContent>

          <TabsContent value="analysis" className="space-y-4">
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {Object.entries(analysis).map(([key, ta]) => (
                <div key={key} className="border rounded-lg p-4 hover:bg-gray-50">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <span className="font-semibold">{ta.symbol}</span>
                      <Badge variant="outline">{ta.timeFrame}</Badge>
                      <Badge className={getSignalColor(ta.overallSignal)}>
                        {ta.overallSignal.replace('_', ' ').toUpperCase()}
                      </Badge>
                    </div>
                    <div className="text-right">
                      <div className={`font-semibold ${getConfidenceColor(ta.overallScore)}`}>
                        {ta.overallScore}/100
                      </div>
                      <div className="text-sm text-gray-600">
                        {ta.trendDirection} ({ta.trendStrength}%)
                      </div>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm">
                    {ta.indicators.slice(0, 4).map((indicator, index) => (
                      <div key={index} className="text-center p-2 bg-gray-100 rounded">
                        <div className="font-medium">{indicator.name}</div>
                        <div className={getSignalColor(indicator.signal)}>
                          {indicator.signal.toUpperCase()}
                        </div>
                        <div className="text-xs text-gray-600">
                          {indicator.value.toFixed(2)}
                        </div>
                      </div>
                    ))}
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

export default TradingSignalsWidget;
