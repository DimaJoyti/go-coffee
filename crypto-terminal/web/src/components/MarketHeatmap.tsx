import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  TrendingUp, 
  TrendingDown, 
  RefreshCw, 
  BarChart3, 
  PieChart,
  Activity,
  AlertTriangle,
  Eye
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface HeatmapCoin {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  marketCap: number;
  volume24h: number;
  color: string;
  size: number;
  logoUrl: string;
  lastUpdated: string;
}

interface SectorData {
  name: string;
  marketCap: number;
  change24h: number;
  volume24h: number;
  coinCount: number;
  topCoins: HeatmapCoin[];
  performance: string;
}

interface MarketHeatmapData {
  sectors: SectorData[];
  topMovers: HeatmapCoin[];
  marketSentiment: string;
  totalMarketCap: number;
  lastUpdated: string;
}

interface MarketHeatmapProps {
  className?: string;
}

const MarketHeatmap: React.FC<MarketHeatmapProps> = ({ className }) => {
  const [heatmapData, setHeatmapData] = useState<MarketHeatmapData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedCoin, setSelectedCoin] = useState<HeatmapCoin | null>(null);
  const [viewMode, setViewMode] = useState<'market-cap' | 'performance'>('market-cap');

  const fetchHeatmapData = async () => {
    try {
      setLoading(true);
      
      const response = await fetch('/api/v2/market/heatmap');
      
      if (!response.ok) {
        throw new Error('Failed to fetch market heatmap data');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setHeatmapData(data.data);
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
    fetchHeatmapData();
    
    // Auto-refresh every 60 seconds
    const interval = setInterval(fetchHeatmapData, 60000);
    
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
    if (value > 5) return '#2ED573'; // Strong green
    if (value > 0) return '#7ED321'; // Light green
    if (value > -5) return '#FF6B6B'; // Light red
    return '#FF4757'; // Strong red
  };

  const getSentimentColor = (sentiment: string) => {
    switch (sentiment.toLowerCase()) {
      case 'bullish':
      case 'very positive':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'bearish':
      case 'negative':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'cautiously optimistic':
      case 'positive':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getCoinSize = (coin: HeatmapCoin) => {
    // Size based on market cap relative to largest coin
    const maxSize = 200;
    const minSize = 40;
    const sizeRange = maxSize - minSize;
    
    // Normalize size (0-1) and apply to range
    const normalizedSize = Math.min(coin.size / 100, 1);
    return minSize + (normalizedSize * sizeRange);
  };

  if (loading && !heatmapData) {
    return (
      <Card className={cn('w-full', className)}>
        <CardContent className="flex items-center justify-center py-8">
          <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
          <span className="ml-2 text-gray-500">Loading market heatmap...</span>
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

  if (!heatmapData) {
    return null;
  }

  return (
    <div className={cn('w-full space-y-6', className)}>
      {/* Header */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <CardTitle className="text-xl font-bold flex items-center gap-2">
            <BarChart3 className="h-5 w-5 text-blue-600" />
            Market Heatmap
          </CardTitle>
          <div className="flex items-center gap-3">
            <Badge className={cn('font-medium', getSentimentColor(heatmapData.marketSentiment))}>
              {heatmapData.marketSentiment}
            </Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={fetchHeatmapData}
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
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Market Cap</p>
                <p className="text-2xl font-bold">{formatCurrency(heatmapData.totalMarketCap)}</p>
              </div>
              <PieChart className="h-8 w-8 text-blue-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Active Sectors</p>
                <p className="text-2xl font-bold">{heatmapData.sectors.length}</p>
              </div>
              <Activity className="h-8 w-8 text-green-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Top Movers</p>
                <p className="text-2xl font-bold">{heatmapData.topMovers.length}</p>
              </div>
              <TrendingUp className="h-8 w-8 text-purple-600" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Heatmap Visualization */}
      <Tabs value={viewMode} onValueChange={(value) => setViewMode(value as 'market-cap' | 'performance')}>
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="market-cap">Market Cap View</TabsTrigger>
          <TabsTrigger value="performance">Performance View</TabsTrigger>
        </TabsList>

        <TabsContent value="market-cap" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Market Cap Heatmap</CardTitle>
              <p className="text-sm text-gray-600">
                Size represents market cap, color represents 24h performance
              </p>
            </CardHeader>
            <CardContent>
              <div className="relative min-h-[400px] bg-gray-50 rounded-lg p-4 overflow-hidden">
                <div className="flex flex-wrap gap-2">
                  {heatmapData.topMovers.map((coin) => (
                    <div
                      key={coin.symbol}
                      className="relative cursor-pointer transition-all duration-200 hover:scale-105 hover:z-10"
                      style={{
                        width: `${getCoinSize(coin)}px`,
                        height: `${getCoinSize(coin)}px`,
                        backgroundColor: getChangeColor(coin.change24h),
                        borderRadius: '8px',
                        minWidth: '60px',
                        minHeight: '60px',
                      }}
                      onClick={() => setSelectedCoin(coin)}
                      onMouseEnter={() => setSelectedCoin(coin)}
                    >
                      <div className="absolute inset-0 flex flex-col items-center justify-center text-white p-2">
                        <span className="font-bold text-xs">{coin.symbol}</span>
                        <span className="text-xs">{formatPercent(coin.change24h)}</span>
                        {getCoinSize(coin) > 80 && (
                          <span className="text-xs opacity-75">{formatCurrency(coin.marketCap)}</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="performance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Sector Performance</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {heatmapData.sectors.map((sector) => (
                  <div
                    key={sector.name}
                    className="p-4 border border-gray-200 rounded-lg hover:shadow-md transition-shadow"
                  >
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-3">
                        <h3 className="font-semibold text-lg">{sector.name}</h3>
                        <Badge 
                          variant="outline"
                          className={cn(
                            sector.change24h >= 0 ? 'text-green-600 border-green-200' : 'text-red-600 border-red-200'
                          )}
                        >
                          {formatPercent(sector.change24h)}
                        </Badge>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{formatCurrency(sector.marketCap)}</p>
                        <p className="text-sm text-gray-500">{sector.coinCount} coins</p>
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-3 gap-2">
                      {sector.topCoins.slice(0, 3).map((coin) => (
                        <div
                          key={coin.symbol}
                          className="p-2 bg-gray-50 rounded text-center cursor-pointer hover:bg-gray-100"
                          onClick={() => setSelectedCoin(coin)}
                        >
                          <p className="font-medium text-sm">{coin.symbol}</p>
                          <p className={cn('text-xs', getChangeColor(coin.change24h) === '#2ED573' || getChangeColor(coin.change24h) === '#7ED321' ? 'text-green-600' : 'text-red-600')}>
                            {formatPercent(coin.change24h)}
                          </p>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Selected Coin Details */}
      {selectedCoin && (
        <Card className="border-blue-200 bg-blue-50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Eye className="h-5 w-5 text-blue-600" />
              {selectedCoin.name} ({selectedCoin.symbol})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <p className="text-sm text-gray-600">Price</p>
                <p className="font-bold">{formatCurrency(selectedCoin.price)}</p>
              </div>
              <div>
                <p className="text-sm text-gray-600">24h Change</p>
                <p className={cn('font-bold', selectedCoin.change24h >= 0 ? 'text-green-600' : 'text-red-600')}>
                  {formatPercent(selectedCoin.change24h)}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-600">Market Cap</p>
                <p className="font-bold">{formatCurrency(selectedCoin.marketCap)}</p>
              </div>
              <div>
                <p className="text-sm text-gray-600">24h Volume</p>
                <p className="font-bold">{formatCurrency(selectedCoin.volume24h)}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Legend */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Color Legend</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-6 text-sm">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded" style={{ backgroundColor: '#2ED573' }}></div>
              <span>Strong Gain (+5%)</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded" style={{ backgroundColor: '#7ED321' }}></div>
              <span>Gain (0% to +5%)</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded" style={{ backgroundColor: '#FF6B6B' }}></div>
              <span>Loss (0% to -5%)</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded" style={{ backgroundColor: '#FF4757' }}></div>
              <span>Strong Loss (-5%)</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default MarketHeatmap;
