import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { 
  TrendingUp, 
  TrendingDown, 
  RefreshCw, 
  Search, 
  Star,
  AlertTriangle,
  BarChart3,
  Activity,
  DollarSign
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface TradingViewCoin {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  changePercent: number;
  marketCap: number;
  volume24h: number;
  circSupply: number;
  volMarketCap: number;
  socialDominance: number;
  category: string[];
  techRating: string;
  rank: number;
  logoUrl: string;
  lastUpdated: string;
}

interface TrendingCoin {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  volume24h: number;
  trendScore: number;
  mentions: number;
  logoUrl: string;
  lastUpdated: string;
}

interface MarketOverview {
  totalMarketCap: number;
  totalVolume24h: number;
  btcDominance: number;
  ethDominance: number;
  activeCoins: number;
  marketSentiment: string;
  fearGreedIndex: number;
  lastUpdated: string;
}

interface TradingViewData {
  coins: TradingViewCoin[];
  marketOverview: MarketOverview;
  trendingCoins: TrendingCoin[];
  gainers: TradingViewCoin[];
  losers: TradingViewCoin[];
  lastUpdated: string;
  dataQuality: number;
}

interface TradingViewWidgetProps {
  className?: string;
}

const TradingViewWidget: React.FC<TradingViewWidgetProps> = ({ className }) => {
  const [tradingViewData, setTradingViewData] = useState<TradingViewData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedTab, setSelectedTab] = useState('overview');

  const fetchTradingViewData = async () => {
    try {
      setLoading(true);
      
      const response = await fetch('/api/v2/tradingview/market-data');
      
      if (!response.ok) {
        throw new Error('Failed to fetch TradingView data');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setTradingViewData(data.data);
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
    fetchTradingViewData();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchTradingViewData, 30000);
    
    return () => clearInterval(interval);
  }, []);

  const formatCurrency = (value: number) => {
    if (value >= 1e12) return `$${(value / 1e12).toFixed(2)}T`;
    if (value >= 1e9) return `$${(value / 1e9).toFixed(2)}B`;
    if (value >= 1e6) return `$${(value / 1e6).toFixed(2)}M`;
    if (value >= 1e3) return `$${(value / 1e3).toFixed(2)}K`;
    return `$${value.toFixed(2)}`;
  };

  const formatPercent = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
  };

  const getChangeColor = (value: number) => {
    if (value > 0) return 'text-green-600';
    if (value < 0) return 'text-red-600';
    return 'text-gray-600';
  };

  const getTechRatingColor = (rating: string) => {
    switch (rating.toLowerCase()) {
      case 'buy':
      case 'strong buy':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'sell':
      case 'strong sell':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'neutral':
        return 'bg-gray-100 text-gray-800 border-gray-200';
      default:
        return 'bg-blue-100 text-blue-800 border-blue-200';
    }
  };

  const filteredCoins = tradingViewData?.coins.filter(coin =>
    coin.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
    coin.name.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  if (loading && !tradingViewData) {
    return (
      <Card className={cn('w-full', className)}>
        <CardContent className="flex items-center justify-center py-8">
          <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
          <span className="ml-2 text-gray-500">Loading TradingView data...</span>
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

  if (!tradingViewData) {
    return null;
  }

  return (
    <div className={cn('w-full space-y-6', className)}>
      {/* Header */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <CardTitle className="text-xl font-bold flex items-center gap-2">
            <BarChart3 className="h-5 w-5 text-blue-600" />
            TradingView Market Data
          </CardTitle>
          <div className="flex items-center gap-3">
            <Badge variant="outline" className="text-green-600 border-green-200">
              Quality: {(tradingViewData.dataQuality * 100).toFixed(0)}%
            </Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={fetchTradingViewData}
              disabled={loading}
              className="flex items-center gap-1"
            >
              <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              Refresh
            </Button>
          </div>
        </CardHeader>
      </Card>

      {/* Market Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Market Cap</p>
                <p className="text-2xl font-bold">{formatCurrency(tradingViewData.marketOverview.totalMarketCap)}</p>
              </div>
              <DollarSign className="h-8 w-8 text-blue-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">24h Volume</p>
                <p className="text-2xl font-bold">{formatCurrency(tradingViewData.marketOverview.totalVolume24h)}</p>
              </div>
              <Activity className="h-8 w-8 text-green-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">BTC Dominance</p>
                <p className="text-2xl font-bold">{tradingViewData.marketOverview.btcDominance.toFixed(1)}%</p>
              </div>
              <TrendingUp className="h-8 w-8 text-orange-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Fear & Greed</p>
                <p className="text-2xl font-bold">{tradingViewData.marketOverview.fearGreedIndex}</p>
                <p className="text-sm text-gray-500">{tradingViewData.marketOverview.marketSentiment}</p>
              </div>
              <Star className="h-8 w-8 text-purple-600" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Data Tabs */}
      <Tabs value={selectedTab} onValueChange={setSelectedTab}>
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="trending">Trending</TabsTrigger>
          <TabsTrigger value="gainers">Gainers</TabsTrigger>
          <TabsTrigger value="losers">Losers</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Top Cryptocurrencies</CardTitle>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
                  <Input
                    placeholder="Search coins..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10 w-64"
                  />
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {filteredCoins.slice(0, 20).map((coin) => (
                  <div
                    key={coin.symbol}
                    className="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:shadow-md transition-shadow"
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex items-center gap-2">
                        <span className="text-sm text-gray-500 w-8">#{coin.rank}</span>
                        <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
                          <span className="text-xs font-bold">{coin.symbol.slice(0, 2)}</span>
                        </div>
                      </div>
                      <div>
                        <p className="font-medium">{coin.name}</p>
                        <p className="text-sm text-gray-500">{coin.symbol}</p>
                      </div>
                      <div className="flex gap-1">
                        {coin.category.slice(0, 2).map((cat) => (
                          <Badge key={cat} variant="outline" className="text-xs">
                            {cat}
                          </Badge>
                        ))}
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-6">
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.price)}</p>
                        <p className={cn('text-sm', getChangeColor(coin.changePercent))}>
                          {formatPercent(coin.changePercent)}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.marketCap)}</p>
                        <p className="text-sm text-gray-500">Market Cap</p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.volume24h)}</p>
                        <p className="text-sm text-gray-500">Volume</p>
                      </div>
                      <Badge className={cn('text-xs', getTechRatingColor(coin.techRating))}>
                        {coin.techRating}
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="trending" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Trending Cryptocurrencies</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {tradingViewData.trendingCoins.map((coin, index) => (
                  <div
                    key={coin.symbol}
                    className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 w-8">#{index + 1}</span>
                      <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
                        <span className="text-xs font-bold">{coin.symbol.slice(0, 2)}</span>
                      </div>
                      <div>
                        <p className="font-medium">{coin.name}</p>
                        <p className="text-sm text-gray-500">{coin.symbol}</p>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-6">
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.price)}</p>
                        <p className={cn('text-sm', getChangeColor(coin.change24h))}>
                          {formatPercent(coin.change24h)}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{coin.trendScore.toFixed(1)}</p>
                        <p className="text-sm text-gray-500">Trend Score</p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{coin.mentions.toLocaleString()}</p>
                        <p className="text-sm text-gray-500">Mentions</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="gainers" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-green-600" />
                Top Gainers (24h)
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {tradingViewData.gainers.map((coin, index) => (
                  <div
                    key={coin.symbol}
                    className="flex items-center justify-between p-3 border border-green-200 bg-green-50 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 w-8">#{index + 1}</span>
                      <div>
                        <p className="font-medium">{coin.name}</p>
                        <p className="text-sm text-gray-500">{coin.symbol}</p>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-6">
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.price)}</p>
                        <p className="text-sm text-green-600 font-bold">
                          {formatPercent(coin.changePercent)}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.marketCap)}</p>
                        <p className="text-sm text-gray-500">Market Cap</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="losers" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingDown className="h-5 w-5 text-red-600" />
                Top Losers (24h)
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {tradingViewData.losers.map((coin, index) => (
                  <div
                    key={coin.symbol}
                    className="flex items-center justify-between p-3 border border-red-200 bg-red-50 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 w-8">#{index + 1}</span>
                      <div>
                        <p className="font-medium">{coin.name}</p>
                        <p className="text-sm text-gray-500">{coin.symbol}</p>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-6">
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.price)}</p>
                        <p className="text-sm text-red-600 font-bold">
                          {formatPercent(coin.changePercent)}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(coin.marketCap)}</p>
                        <p className="text-sm text-gray-500">Market Cap</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default TradingViewWidget;
