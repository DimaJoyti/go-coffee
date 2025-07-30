import React, { useState, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Badge } from '../ui/badge';
import { 
  Search,
  Star,
  TrendingUp,
  TrendingDown,
  Volume2,
  BarChart3,
  Filter,
  ArrowUpDown
} from 'lucide-react';

interface MarketDataItem {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  changePercent24h: number;
  volume24h: number;
  marketCap: number;
  rank: number;
  logo?: string;
  isFavorite?: boolean;
}

interface EnhancedMarketDataProps {
  data?: MarketDataItem[];
  onSymbolSelect?: (symbol: string) => void;
  onFavoriteToggle?: (symbol: string) => void;
  className?: string;
}

// Mock data
const mockMarketData: MarketDataItem[] = [
  {
    symbol: 'BTC',
    name: 'Bitcoin',
    price: 43245.75,
    change24h: 1250.75,
    changePercent24h: 2.98,
    volume24h: 28500000000,
    marketCap: 847000000000,
    rank: 1,
    isFavorite: true
  },
  {
    symbol: 'ETH',
    name: 'Ethereum',
    price: 2650.25,
    change24h: -45.50,
    changePercent24h: -1.67,
    volume24h: 15200000000,
    marketCap: 318000000000,
    rank: 2,
    isFavorite: false
  },
  {
    symbol: 'BNB',
    name: 'Binance Coin',
    price: 315.80,
    change24h: 8.25,
    changePercent24h: 2.68,
    volume24h: 1200000000,
    marketCap: 47000000000,
    rank: 3,
    isFavorite: true
  },
  {
    symbol: 'SOL',
    name: 'Solana',
    price: 98.45,
    change24h: -2.15,
    changePercent24h: -2.14,
    volume24h: 2100000000,
    marketCap: 42000000000,
    rank: 4,
    isFavorite: false
  },
  {
    symbol: 'ADA',
    name: 'Cardano',
    price: 0.485,
    change24h: 0.025,
    changePercent24h: 5.43,
    volume24h: 850000000,
    marketCap: 17000000000,
    rank: 5,
    isFavorite: false
  }
];

type SortField = 'rank' | 'price' | 'changePercent24h' | 'volume24h' | 'marketCap';
type SortDirection = 'asc' | 'desc';

export const EnhancedMarketData: React.FC<EnhancedMarketDataProps> = ({
  data = mockMarketData,
  onSymbolSelect,
  onFavoriteToggle,
  className
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [showFavoritesOnly, setShowFavoritesOnly] = useState(false);
  const [sortField, setSortField] = useState<SortField>('rank');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

  // Filter and sort data
  const filteredAndSortedData = useMemo(() => {
    let filtered = data.filter(item => {
      const matchesSearch = item.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           item.symbol.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesFavorites = !showFavoritesOnly || item.isFavorite;
      return matchesSearch && matchesFavorites;
    });

    // Sort data
    filtered.sort((a, b) => {
      let aValue = a[sortField];
      let bValue = b[sortField];
      
      if (sortDirection === 'desc') {
        [aValue, bValue] = [bValue, aValue];
      }
      
      return aValue > bValue ? 1 : -1;
    });

    return filtered;
  }, [data, searchTerm, showFavoritesOnly, sortField, sortDirection]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const formatNumber = (num: number, decimals: number = 2) => {
    if (num >= 1e9) {
      return `$${(num / 1e9).toFixed(1)}B`;
    } else if (num >= 1e6) {
      return `$${(num / 1e6).toFixed(1)}M`;
    } else if (num >= 1e3) {
      return `$${(num / 1e3).toFixed(1)}K`;
    }
    return `$${num.toFixed(decimals)}`;
  };

  const SortButton: React.FC<{ field: SortField; children: React.ReactNode }> = ({ field, children }) => (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => handleSort(field)}
      className="h-auto p-1 font-medium text-xs hover:bg-muted/50"
    >
      <span className="flex items-center space-x-1">
        <span>{children}</span>
        <ArrowUpDown className="w-3 h-3" />
      </span>
    </Button>
  );

  return (
    <Card className={`trading-card-glass ${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Market Data</CardTitle>
          <div className="flex items-center space-x-2">
            <Button
              variant={showFavoritesOnly ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setShowFavoritesOnly(!showFavoritesOnly)}
              className="h-8 px-3"
            >
              <Star className={`w-4 h-4 ${showFavoritesOnly ? 'fill-current' : ''}`} />
            </Button>
            <Button variant="ghost" size="sm" className="h-8 px-3">
              <Filter className="w-4 h-4" />
            </Button>
          </div>
        </div>
        
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search cryptocurrencies..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
      </CardHeader>
      
      <CardContent className="p-0">
        {/* Table Header */}
        <div className="grid grid-cols-7 gap-2 py-2 px-4 text-xs font-medium text-muted-foreground border-b border-border">
          <div className="flex items-center">
            <SortButton field="rank">#</SortButton>
          </div>
          <div className="col-span-2">Name</div>
          <div className="text-right">
            <SortButton field="price">Price</SortButton>
          </div>
          <div className="text-right">
            <SortButton field="changePercent24h">24h %</SortButton>
          </div>
          <div className="text-right">
            <SortButton field="volume24h">Volume</SortButton>
          </div>
          <div className="text-right">
            <SortButton field="marketCap">Market Cap</SortButton>
          </div>
        </div>
        
        {/* Table Body */}
        <div className="max-h-96 overflow-y-auto scrollbar-thin">
          {filteredAndSortedData.map((item) => (
            <div
              key={item.symbol}
              className="grid grid-cols-7 gap-2 py-3 px-4 text-sm market-data-row border-b border-border/50 last:border-b-0"
              onClick={() => onSymbolSelect?.(item.symbol)}
            >
              {/* Rank & Favorite */}
              <div className="flex items-center space-x-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    onFavoriteToggle?.(item.symbol);
                  }}
                  className="h-auto p-0 hover:bg-transparent"
                >
                  <Star className={`w-3 h-3 ${item.isFavorite ? 'fill-yellow-400 text-yellow-400' : 'text-muted-foreground'}`} />
                </Button>
                <span className="text-muted-foreground">{item.rank}</span>
              </div>
              
              {/* Name & Symbol */}
              <div className="col-span-2 flex items-center space-x-3">
                <div className="w-8 h-8 bg-muted rounded-full flex items-center justify-center">
                  <span className="text-xs font-bold">{item.symbol.slice(0, 2)}</span>
                </div>
                <div>
                  <div className="font-medium">{item.symbol}</div>
                  <div className="text-xs text-muted-foreground">{item.name}</div>
                </div>
              </div>
              
              {/* Price */}
              <div className="text-right font-mono">
                {item.price < 1 ? `$${item.price.toFixed(4)}` : `$${item.price.toLocaleString()}`}
              </div>
              
              {/* 24h Change */}
              <div className="text-right">
                <div className={`flex items-center justify-end space-x-1 ${
                  item.changePercent24h >= 0 ? 'text-green-400' : 'text-red-400'
                }`}>
                  {item.changePercent24h >= 0 ? (
                    <TrendingUp className="w-3 h-3" />
                  ) : (
                    <TrendingDown className="w-3 h-3" />
                  )}
                  <span className="font-mono">
                    {item.changePercent24h >= 0 ? '+' : ''}{item.changePercent24h.toFixed(2)}%
                  </span>
                </div>
              </div>
              
              {/* Volume */}
              <div className="text-right font-mono text-muted-foreground">
                {formatNumber(item.volume24h)}
              </div>
              
              {/* Market Cap */}
              <div className="text-right font-mono text-muted-foreground">
                {formatNumber(item.marketCap)}
              </div>
            </div>
          ))}
        </div>
        
        {filteredAndSortedData.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            No cryptocurrencies found matching your criteria.
          </div>
        )}
      </CardContent>
    </Card>
  );
};
