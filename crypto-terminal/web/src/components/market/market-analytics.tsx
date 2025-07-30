import React, { useState, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { Input } from '../ui/input';
import { 
  TrendingUp, 
  TrendingDown, 
  Activity,
  BarChart3,
  Globe,
  Zap,
  Bell,
  Search,
  Filter,
  Calendar,
  DollarSign,
  Volume2,
  Target,
  AlertTriangle
} from 'lucide-react';

interface MarketOverview {
  totalMarketCap: number;
  totalVolume24h: number;
  btcDominance: number;
  ethDominance: number;
  defiTvl: number;
  fearGreedIndex: number;
  activeCoins: number;
  exchanges: number;
}

interface TrendingCoin {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  volume24h: number;
  marketCap: number;
  rank: number;
  sparkline: number[];
}

interface MarketSector {
  name: string;
  marketCap: number;
  change24h: number;
  topCoins: string[];
  performance: 'outperforming' | 'underperforming' | 'neutral';
}

interface NewsItem {
  id: string;
  title: string;
  summary: string;
  source: string;
  timestamp: number;
  sentiment: 'positive' | 'negative' | 'neutral';
  impact: 'high' | 'medium' | 'low';
  relatedSymbols: string[];
}

interface MarketAnalyticsProps {
  className?: string;
}

// Mock data
const mockMarketOverview: MarketOverview = {
  totalMarketCap: 1680000000000,
  totalVolume24h: 89500000000,
  btcDominance: 52.3,
  ethDominance: 17.8,
  defiTvl: 45200000000,
  fearGreedIndex: 68,
  activeCoins: 2847,
  exchanges: 156
};

const mockTrendingCoins: TrendingCoin[] = [
  {
    symbol: 'BTC',
    name: 'Bitcoin',
    price: 43245.75,
    change24h: 2.98,
    volume24h: 28500000000,
    marketCap: 847000000000,
    rank: 1,
    sparkline: [42000, 42500, 43000, 42800, 43200, 43245]
  },
  {
    symbol: 'ETH',
    name: 'Ethereum',
    price: 2650.25,
    change24h: -1.67,
    volume24h: 15200000000,
    marketCap: 318000000000,
    rank: 2,
    sparkline: [2700, 2680, 2650, 2670, 2655, 2650]
  },
  {
    symbol: 'SOL',
    name: 'Solana',
    price: 98.45,
    change24h: 8.92,
    volume24h: 2100000000,
    marketCap: 42000000000,
    rank: 4,
    sparkline: [90, 92, 95, 97, 98, 98.45]
  }
];

const mockMarketSectors: MarketSector[] = [
  {
    name: 'DeFi',
    marketCap: 89500000000,
    change24h: 3.45,
    topCoins: ['UNI', 'AAVE', 'COMP'],
    performance: 'outperforming'
  },
  {
    name: 'Layer 1',
    marketCap: 456000000000,
    change24h: 1.23,
    topCoins: ['ETH', 'SOL', 'ADA'],
    performance: 'neutral'
  },
  {
    name: 'AI & Big Data',
    marketCap: 23400000000,
    change24h: 12.67,
    topCoins: ['FET', 'OCEAN', 'GRT'],
    performance: 'outperforming'
  },
  {
    name: 'Gaming',
    marketCap: 15600000000,
    change24h: -2.34,
    topCoins: ['AXS', 'SAND', 'MANA'],
    performance: 'underperforming'
  }
];

const mockNews: NewsItem[] = [
  {
    id: '1',
    title: 'Bitcoin ETF Sees Record Inflows as Institutional Adoption Accelerates',
    summary: 'Major institutional investors continue to pour money into Bitcoin ETFs, with over $2.1B in net inflows this week.',
    source: 'CoinDesk',
    timestamp: Date.now() - 1800000,
    sentiment: 'positive',
    impact: 'high',
    relatedSymbols: ['BTC']
  },
  {
    id: '2',
    title: 'Ethereum Layer 2 Solutions See 300% Growth in TVL',
    summary: 'Arbitrum and Optimism lead the charge as Layer 2 total value locked reaches new all-time highs.',
    source: 'The Block',
    timestamp: Date.now() - 3600000,
    sentiment: 'positive',
    impact: 'medium',
    relatedSymbols: ['ETH', 'ARB', 'OP']
  },
  {
    id: '3',
    title: 'Regulatory Clarity Boosts DeFi Sector Performance',
    summary: 'New regulatory framework provides clearer guidelines for DeFi protocols, leading to sector-wide gains.',
    source: 'CryptoSlate',
    timestamp: Date.now() - 7200000,
    sentiment: 'positive',
    impact: 'medium',
    relatedSymbols: ['UNI', 'AAVE', 'COMP']
  }
];

export const MarketAnalytics: React.FC<MarketAnalyticsProps> = ({ className }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedSector, setSelectedSector] = useState('all');

  const fearGreedStatus = useMemo(() => {
    const index = mockMarketOverview.fearGreedIndex;
    if (index >= 75) return { label: 'Extreme Greed', color: 'text-red-400' };
    if (index >= 55) return { label: 'Greed', color: 'text-orange-400' };
    if (index >= 45) return { label: 'Neutral', color: 'text-yellow-400' };
    if (index >= 25) return { label: 'Fear', color: 'text-blue-400' };
    return { label: 'Extreme Fear', color: 'text-purple-400' };
  }, []);

  const formatNumber = (num: number, decimals: number = 2) => {
    if (num >= 1e12) {
      return `$${(num / 1e12).toFixed(1)}T`;
    } else if (num >= 1e9) {
      return `$${(num / 1e9).toFixed(1)}B`;
    } else if (num >= 1e6) {
      return `$${(num / 1e6).toFixed(1)}M`;
    }
    return `$${num.toFixed(decimals)}`;
  };

  const getSentimentColor = (sentiment: string) => {
    switch (sentiment) {
      case 'positive': return 'text-green-400';
      case 'negative': return 'text-red-400';
      default: return 'text-muted-foreground';
    }
  };

  const getPerformanceColor = (performance: string) => {
    switch (performance) {
      case 'outperforming': return 'text-green-400';
      case 'underperforming': return 'text-red-400';
      default: return 'text-muted-foreground';
    }
  };

  const OverviewCard: React.FC<{
    title: string;
    value: string | number;
    change?: number;
    icon: React.ReactNode;
    subtitle?: string;
  }> = ({ title, value, change, icon, subtitle }) => (
    <Card className="trading-card">
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <p className="text-sm text-muted-foreground">{title}</p>
            <p className="text-xl font-bold">
              {typeof value === 'number' ? value.toLocaleString() : value}
            </p>
            {subtitle && (
              <p className="text-xs text-muted-foreground">{subtitle}</p>
            )}
            {change !== undefined && (
              <p className={`text-sm flex items-center space-x-1 ${
                change >= 0 ? 'text-green-400' : 'text-red-400'
              }`}>
                {change >= 0 ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
                <span>{change >= 0 ? '+' : ''}{change.toFixed(2)}%</span>
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

  const TrendingCoinRow: React.FC<{ coin: TrendingCoin }> = ({ coin }) => (
    <div className="grid grid-cols-6 gap-4 py-3 text-sm border-b border-border/50 last:border-b-0">
      <div className="flex items-center space-x-3">
        <span className="text-muted-foreground">#{coin.rank}</span>
        <div>
          <div className="font-medium">{coin.symbol}</div>
          <div className="text-xs text-muted-foreground">{coin.name}</div>
        </div>
      </div>
      <div className="font-mono">${coin.price.toLocaleString()}</div>
      <div className={`font-mono ${coin.change24h >= 0 ? 'text-green-400' : 'text-red-400'}`}>
        {coin.change24h >= 0 ? '+' : ''}{coin.change24h.toFixed(2)}%
      </div>
      <div className="font-mono text-muted-foreground">
        {formatNumber(coin.volume24h)}
      </div>
      <div className="font-mono text-muted-foreground">
        {formatNumber(coin.marketCap)}
      </div>
      <div className="flex items-center justify-end">
        <div className="w-16 h-8 bg-muted/20 rounded flex items-end space-x-0.5 px-1">
          {coin.sparkline.map((point, index) => {
            const height = ((point - Math.min(...coin.sparkline)) / 
                          (Math.max(...coin.sparkline) - Math.min(...coin.sparkline))) * 100;
            return (
              <div
                key={index}
                className={`w-1 ${coin.change24h >= 0 ? 'bg-green-400' : 'bg-red-400'} opacity-60`}
                style={{ height: `${Math.max(height, 10)}%` }}
              />
            );
          })}
        </div>
      </div>
    </div>
  );

  const SectorCard: React.FC<{ sector: MarketSector }> = ({ sector }) => (
    <Card className="trading-card">
      <CardContent className="p-4">
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">{sector.name}</h4>
            <Badge 
              variant={sector.performance === 'outperforming' ? 'default' : 
                      sector.performance === 'underperforming' ? 'destructive' : 'secondary'}
              className="text-xs"
            >
              {sector.performance}
            </Badge>
          </div>
          
          <div className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Market Cap</span>
              <span className="font-mono">{formatNumber(sector.marketCap)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">24h Change</span>
              <span className={`font-mono ${sector.change24h >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                {sector.change24h >= 0 ? '+' : ''}{sector.change24h.toFixed(2)}%
              </span>
            </div>
          </div>
          
          <div className="flex flex-wrap gap-1">
            {sector.topCoins.map((coin) => (
              <Badge key={coin} variant="outline" className="text-xs">
                {coin}
              </Badge>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );

  const NewsItem: React.FC<{ news: NewsItem }> = ({ news }) => (
    <Card className="trading-card">
      <CardContent className="p-4">
        <div className="space-y-3">
          <div className="flex items-start justify-between">
            <h4 className="font-medium text-sm leading-tight">{news.title}</h4>
            <div className="flex items-center space-x-2 ml-2">
              <Badge 
                variant={news.impact === 'high' ? 'destructive' : 
                        news.impact === 'medium' ? 'secondary' : 'outline'}
                className="text-xs"
              >
                {news.impact}
              </Badge>
            </div>
          </div>
          
          <p className="text-sm text-muted-foreground">{news.summary}</p>
          
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-xs text-muted-foreground">{news.source}</span>
              <span className="text-xs text-muted-foreground">•</span>
              <span className="text-xs text-muted-foreground">
                {new Date(news.timestamp).toLocaleTimeString()}
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <span className={`text-xs ${getSentimentColor(news.sentiment)}`}>
                {news.sentiment}
              </span>
              {news.relatedSymbols.length > 0 && (
                <div className="flex space-x-1">
                  {news.relatedSymbols.slice(0, 3).map((symbol) => (
                    <Badge key={symbol} variant="outline" className="text-xs">
                      {symbol}
                    </Badge>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Market Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <OverviewCard
          title="Total Market Cap"
          value={formatNumber(mockMarketOverview.totalMarketCap)}
          icon={<Globe className="h-6 w-6" />}
        />
        <OverviewCard
          title="24h Volume"
          value={formatNumber(mockMarketOverview.totalVolume24h)}
          icon={<Volume2 className="h-6 w-6" />}
        />
        <OverviewCard
          title="BTC Dominance"
          value={`${mockMarketOverview.btcDominance}%`}
          icon={<Target className="h-6 w-6" />}
        />
        <OverviewCard
          title="Fear & Greed Index"
          value={mockMarketOverview.fearGreedIndex}
          subtitle={fearGreedStatus.label}
          icon={<Activity className="h-6 w-6" />}
        />
      </div>

      {/* Market Analytics Tabs */}
      <Tabs defaultValue="trending" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="trending">Trending</TabsTrigger>
          <TabsTrigger value="sectors">Sectors</TabsTrigger>
          <TabsTrigger value="news">News</TabsTrigger>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
        </TabsList>
        
        <TabsContent value="trending" className="space-y-4">
          <Card className="trading-card">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">Trending Cryptocurrencies</CardTitle>
                <div className="flex items-center space-x-2">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                    <Input
                      placeholder="Search coins..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-10 w-48"
                    />
                  </div>
                  <Button variant="outline" size="sm">
                    <Filter className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {/* Table Header */}
              <div className="grid grid-cols-6 gap-4 py-2 text-xs font-medium text-muted-foreground border-b border-border">
                <div>Rank & Name</div>
                <div>Price</div>
                <div>24h %</div>
                <div>Volume</div>
                <div>Market Cap</div>
                <div className="text-right">7d Chart</div>
              </div>
              
              {/* Coin Rows */}
              <div className="space-y-0">
                {mockTrendingCoins.map((coin) => (
                  <TrendingCoinRow key={coin.symbol} coin={coin} />
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="sectors" className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {mockMarketSectors.map((sector) => (
              <SectorCard key={sector.name} sector={sector} />
            ))}
          </div>
        </TabsContent>
        
        <TabsContent value="news" className="space-y-4">
          <div className="space-y-4">
            {mockNews.map((news) => (
              <NewsItem key={news.id} news={news} />
            ))}
          </div>
        </TabsContent>
        
        <TabsContent value="metrics" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <OverviewCard
              title="DeFi TVL"
              value={formatNumber(mockMarketOverview.defiTvl)}
              icon={<Zap className="h-6 w-6" />}
            />
            <OverviewCard
              title="Active Coins"
              value={mockMarketOverview.activeCoins}
              icon={<BarChart3 className="h-6 w-6" />}
            />
            <OverviewCard
              title="ETH Dominance"
              value={`${mockMarketOverview.ethDominance}%`}
              icon={<Target className="h-6 w-6" />}
            />
            <OverviewCard
              title="Active Exchanges"
              value={mockMarketOverview.exchanges}
              icon={<Globe className="h-6 w-6" />}
            />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
};
