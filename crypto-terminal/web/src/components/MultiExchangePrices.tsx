import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { RefreshCw, TrendingUp, TrendingDown, Activity, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ExchangePrice {
  exchange: string;
  price: number;
  volume: number;
  timestamp: string;
}

interface MarketSummary {
  symbol: string;
  best_bid: {
    exchange: string;
    price: number;
    volume: number;
    timestamp: string;
  };
  best_ask: {
    exchange: string;
    price: number;
    volume: number;
    timestamp: string;
  };
  weighted_price: number;
  price_spread: number;
  spread_percent: number;
  total_volume: number;
  exchange_prices: Record<string, {
    last_price: number;
    bid_price: number;
    ask_price: number;
    volume_24h: number;
    change_percent_24h: number;
    timestamp: string;
  }>;
  data_quality: number;
  timestamp: string;
}

interface MultiExchangePricesProps {
  className?: string;
}

const MultiExchangePrices: React.FC<MultiExchangePricesProps> = ({ className }) => {
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT');
  const [marketSummary, setMarketSummary] = useState<MarketSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const symbols = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'ADAUSDT', 'SOLUSDT'];

  const fetchMarketSummary = async (symbol: string) => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v2/market/summary/${symbol}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch market summary');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setMarketSummary(data.data);
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
    fetchMarketSummary(selectedSymbol);
    
    // Auto-refresh every 15 seconds
    const interval = setInterval(() => {
      fetchMarketSummary(selectedSymbol);
    }, 15000);
    
    return () => clearInterval(interval);
  }, [selectedSymbol]);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 6,
    }).format(value);
  };

  const formatVolume = (volume: number) => {
    if (volume >= 1000000) {
      return `${(volume / 1000000).toFixed(2)}M`;
    }
    if (volume >= 1000) {
      return `${(volume / 1000).toFixed(2)}K`;
    }
    return volume.toFixed(2);
  };

  const getChangeColor = (change: number) => {
    if (change > 0) return 'text-green-600';
    if (change < 0) return 'text-red-600';
    return 'text-gray-600';
  };

  const getQualityColor = (quality: number) => {
    if (quality >= 0.8) return 'text-green-600 bg-green-50';
    if (quality >= 0.6) return 'text-yellow-600 bg-yellow-50';
    return 'text-red-600 bg-red-50';
  };

  const getBestPriceExchange = (prices: Record<string, any>, type: 'bid' | 'ask') => {
    let bestExchange = '';
    let bestPrice = type === 'bid' ? 0 : Infinity;
    
    Object.entries(prices).forEach(([exchange, data]) => {
      const price = type === 'bid' ? data.bid_price : data.ask_price;
      if (type === 'bid' && price > bestPrice) {
        bestPrice = price;
        bestExchange = exchange;
      } else if (type === 'ask' && price < bestPrice) {
        bestPrice = price;
        bestExchange = exchange;
      }
    });
    
    return { exchange: bestExchange, price: bestPrice };
  };

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-xl font-bold flex items-center gap-2">
          <Activity className="h-5 w-5 text-blue-600" />
          Multi-Exchange Prices
        </CardTitle>
        <div className="flex items-center gap-3">
          <Select value={selectedSymbol} onValueChange={setSelectedSymbol}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {symbols.map((symbol) => (
                <SelectItem key={symbol} value={symbol}>
                  {symbol}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          
          {lastUpdate && (
            <span className="text-sm text-gray-500">
              {lastUpdate.toLocaleTimeString()}
            </span>
          )}
          
          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchMarketSummary(selectedSymbol)}
            disabled={loading}
            className="flex items-center gap-1"
          >
            <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
            Refresh
          </Button>
        </div>
      </CardHeader>
      
      <CardContent>
        {error && (
          <div className="flex items-center gap-2 p-4 bg-red-50 border border-red-200 rounded-lg mb-4">
            <AlertCircle className="h-5 w-5 text-red-600" />
            <span className="text-red-700">{error}</span>
          </div>
        )}
        
        {loading && !marketSummary ? (
          <div className="flex items-center justify-center py-8">
            <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
            <span className="ml-2 text-gray-500">Loading market data...</span>
          </div>
        ) : marketSummary ? (
          <div className="space-y-6">
            {/* Market Overview */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 p-4 bg-gray-50 rounded-lg">
              <div className="text-center">
                <p className="text-sm text-gray-600">Weighted Price</p>
                <p className="text-lg font-bold">{formatCurrency(marketSummary.weighted_price)}</p>
              </div>
              <div className="text-center">
                <p className="text-sm text-gray-600">Spread</p>
                <p className="text-lg font-bold text-orange-600">
                  {marketSummary.spread_percent.toFixed(2)}%
                </p>
              </div>
              <div className="text-center">
                <p className="text-sm text-gray-600">Total Volume</p>
                <p className="text-lg font-bold">{formatVolume(marketSummary.total_volume)}</p>
              </div>
              <div className="text-center">
                <p className="text-sm text-gray-600">Data Quality</p>
                <Badge className={cn('font-medium', getQualityColor(marketSummary.data_quality))}>
                  {(marketSummary.data_quality * 100).toFixed(0)}%
                </Badge>
              </div>
            </div>
            
            {/* Best Prices */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-4 border border-green-200 bg-green-50 rounded-lg">
                <h3 className="font-semibold text-green-800 mb-2">Best Bid</h3>
                <div className="space-y-1">
                  <p className="text-lg font-bold text-green-600">
                    {formatCurrency(marketSummary.best_bid.price)}
                  </p>
                  <p className="text-sm text-green-700">
                    {marketSummary.best_bid.exchange} • {formatVolume(marketSummary.best_bid.volume)}
                  </p>
                </div>
              </div>
              
              <div className="p-4 border border-red-200 bg-red-50 rounded-lg">
                <h3 className="font-semibold text-red-800 mb-2">Best Ask</h3>
                <div className="space-y-1">
                  <p className="text-lg font-bold text-red-600">
                    {formatCurrency(marketSummary.best_ask.price)}
                  </p>
                  <p className="text-sm text-red-700">
                    {marketSummary.best_ask.exchange} • {formatVolume(marketSummary.best_ask.volume)}
                  </p>
                </div>
              </div>
            </div>
            
            {/* Exchange Prices */}
            <div className="space-y-3">
              <h3 className="font-semibold text-lg">Exchange Prices</h3>
              <div className="space-y-2">
                {Object.entries(marketSummary.exchange_prices).map(([exchange, data]) => {
                  const bestBid = getBestPriceExchange(marketSummary.exchange_prices, 'bid');
                  const bestAsk = getBestPriceExchange(marketSummary.exchange_prices, 'ask');
                  
                  return (
                    <div
                      key={exchange}
                      className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:shadow-sm transition-shadow"
                    >
                      <div className="flex items-center gap-3">
                        <Badge variant="outline" className="font-medium">
                          {exchange}
                        </Badge>
                        {(bestBid.exchange === exchange || bestAsk.exchange === exchange) && (
                          <Badge className="bg-blue-100 text-blue-800 text-xs">
                            Best {bestBid.exchange === exchange ? 'Bid' : 'Ask'}
                          </Badge>
                        )}
                      </div>
                      
                      <div className="grid grid-cols-4 gap-4 text-right">
                        <div>
                          <p className="text-sm text-gray-600">Last</p>
                          <p className="font-medium">{formatCurrency(data.last_price)}</p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-600">Bid</p>
                          <p className="font-medium text-green-600">{formatCurrency(data.bid_price)}</p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-600">Ask</p>
                          <p className="font-medium text-red-600">{formatCurrency(data.ask_price)}</p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-600">24h Change</p>
                          <p className={cn('font-medium flex items-center gap-1', getChangeColor(data.change_percent_24h))}>
                            {data.change_percent_24h > 0 ? (
                              <TrendingUp className="h-3 w-3" />
                            ) : data.change_percent_24h < 0 ? (
                              <TrendingDown className="h-3 w-3" />
                            ) : null}
                            {data.change_percent_24h.toFixed(2)}%
                          </p>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <Activity className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p>No market data available</p>
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default MultiExchangePrices;
