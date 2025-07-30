import React, { useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { 
  BarChart3, 
  TrendingUp, 
  TrendingDown,
  Activity,
  Layers
} from 'lucide-react';

interface MarketDepthLevel {
  price: number;
  quantity: number;
  total: number;
}

interface MarketDepthData {
  symbol: string;
  bids: MarketDepthLevel[];
  asks: MarketDepthLevel[];
  lastUpdateId: number;
}

interface MarketDepthChartProps {
  data?: MarketDepthData;
  symbol?: string;
  height?: number;
  maxLevels?: number;
  className?: string;
}

// Mock data generator
const generateMockDepthData = (symbol: string, basePrice: number): MarketDepthData => {
  const bids: MarketDepthLevel[] = [];
  const asks: MarketDepthLevel[] = [];
  
  let totalBid = 0;
  for (let i = 0; i < 20; i++) {
    const price = basePrice - (i + 1) * 0.5;
    const quantity = Math.random() * 10 + 1;
    totalBid += quantity;
    bids.push({ price, quantity, total: totalBid });
  }
  
  let totalAsk = 0;
  for (let i = 0; i < 20; i++) {
    const price = basePrice + (i + 1) * 0.5;
    const quantity = Math.random() * 10 + 1;
    totalAsk += quantity;
    asks.push({ price, quantity, total: totalAsk });
  }
  
  return {
    symbol,
    bids,
    asks,
    lastUpdateId: Date.now()
  };
};

export const MarketDepthChart: React.FC<MarketDepthChartProps> = ({
  data,
  symbol = 'BTCUSDT',
  height = 300,
  maxLevels = 15,
  className
}) => {
  // Use mock data if no data provided
  const depthData = data || generateMockDepthData(symbol, 43245.75);
  
  // Process data for visualization
  const processedData = useMemo(() => {
    const bids = depthData.bids.slice(0, maxLevels);
    const asks = depthData.asks.slice(0, maxLevels);
    
    // Calculate max total for scaling
    const maxBidTotal = Math.max(...bids.map(b => b.total));
    const maxAskTotal = Math.max(...asks.map(a => a.total));
    const maxTotal = Math.max(maxBidTotal, maxAskTotal);
    
    // Get price range
    const minPrice = Math.min(...bids.map(b => b.price));
    const maxPrice = Math.max(...asks.map(a => a.price));
    const priceRange = maxPrice - minPrice;
    
    return {
      bids,
      asks,
      maxTotal,
      minPrice,
      maxPrice,
      priceRange,
      spread: asks[0]?.price - bids[0]?.price || 0
    };
  }, [depthData, maxLevels]);
  
  // Calculate position and width for depth bars
  const getBarStyle = (level: MarketDepthLevel, side: 'bid' | 'ask') => {
    const { minPrice, priceRange, maxTotal } = processedData;
    
    // Calculate x position based on price
    const xPercent = ((level.price - minPrice) / priceRange) * 100;
    
    // Calculate width based on total quantity
    const widthPercent = (level.total / maxTotal) * 100;
    
    return {
      left: side === 'bid' ? `${xPercent}%` : 'auto',
      right: side === 'ask' ? `${100 - xPercent}%` : 'auto',
      width: `${widthPercent}%`,
      backgroundColor: side === 'bid' ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)',
      borderLeft: side === 'bid' ? '2px solid rgb(16, 185, 129)' : 'none',
      borderRight: side === 'ask' ? '2px solid rgb(239, 68, 68)' : 'none',
    };
  };
  
  // Calculate spread percentage
  const spreadPercent = processedData.spread > 0 
    ? (processedData.spread / ((processedData.bids[0]?.price || 0) + (processedData.asks[0]?.price || 0)) * 0.5) * 100
    : 0;
  
  return (
    <Card className={`trading-card-glass ${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center space-x-2">
            <BarChart3 className="w-5 h-5" />
            <span>Market Depth</span>
          </CardTitle>
          <Badge variant="outline" className="text-xs">
            {symbol}
          </Badge>
        </div>
        
        {/* Spread Info */}
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-1">
              <TrendingUp className="w-4 h-4 text-green-400" />
              <span className="text-green-400">Bids</span>
            </div>
            <div className="flex items-center space-x-1">
              <TrendingDown className="w-4 h-4 text-red-400" />
              <span className="text-red-400">Asks</span>
            </div>
          </div>
          <div className="text-muted-foreground">
            Spread: {processedData.spread.toFixed(2)} ({spreadPercent.toFixed(3)}%)
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="p-4">
        {/* Depth Chart */}
        <div 
          className="relative bg-background/50 rounded-lg border border-border overflow-hidden"
          style={{ height: `${height}px` }}
        >
          {/* Price axis labels */}
          <div className="absolute top-0 left-0 right-0 h-6 flex justify-between items-center px-2 text-xs text-muted-foreground border-b border-border">
            <span>${processedData.minPrice.toFixed(2)}</span>
            <span>Price</span>
            <span>${processedData.maxPrice.toFixed(2)}</span>
          </div>
          
          {/* Depth visualization */}
          <div className="absolute inset-0 top-6">
            {/* Bid depth bars */}
            {processedData.bids.map((bid, index) => (
              <div
                key={`bid-${index}`}
                className="absolute bottom-0 transition-all duration-300 hover:opacity-80"
                style={{
                  ...getBarStyle(bid, 'bid'),
                  height: `${((bid.total / processedData.maxTotal) * 100)}%`,
                  zIndex: processedData.bids.length - index
                }}
                title={`Price: $${bid.price.toFixed(2)}, Quantity: ${bid.quantity.toFixed(4)}, Total: ${bid.total.toFixed(4)}`}
              />
            ))}
            
            {/* Ask depth bars */}
            {processedData.asks.map((ask, index) => (
              <div
                key={`ask-${index}`}
                className="absolute bottom-0 transition-all duration-300 hover:opacity-80"
                style={{
                  ...getBarStyle(ask, 'ask'),
                  height: `${((ask.total / processedData.maxTotal) * 100)}%`,
                  zIndex: processedData.asks.length - index
                }}
                title={`Price: $${ask.price.toFixed(2)}, Quantity: ${ask.quantity.toFixed(4)}, Total: ${ask.total.toFixed(4)}`}
              />
            ))}
            
            {/* Mid price line */}
            <div 
              className="absolute top-0 bottom-0 w-0.5 bg-yellow-400 opacity-60"
              style={{ 
                left: '50%',
                transform: 'translateX(-50%)'
              }}
            />
          </div>
          
          {/* Quantity axis */}
          <div className="absolute left-0 top-6 bottom-0 w-12 flex flex-col justify-between py-2">
            <span className="text-xs text-muted-foreground">
              {processedData.maxTotal.toFixed(1)}
            </span>
            <span className="text-xs text-muted-foreground">
              {(processedData.maxTotal * 0.5).toFixed(1)}
            </span>
            <span className="text-xs text-muted-foreground">0</span>
          </div>
        </div>
        
        {/* Summary Stats */}
        <div className="grid grid-cols-3 gap-4 mt-4 text-sm">
          <div className="text-center">
            <div className="text-muted-foreground">Best Bid</div>
            <div className="font-mono text-green-400">
              ${processedData.bids[0]?.price.toFixed(2) || '0.00'}
            </div>
          </div>
          <div className="text-center">
            <div className="text-muted-foreground">Spread</div>
            <div className="font-mono">
              ${processedData.spread.toFixed(2)}
            </div>
          </div>
          <div className="text-center">
            <div className="text-muted-foreground">Best Ask</div>
            <div className="font-mono text-red-400">
              ${processedData.asks[0]?.price.toFixed(2) || '0.00'}
            </div>
          </div>
        </div>
        
        {/* Total Liquidity */}
        <div className="grid grid-cols-2 gap-4 mt-3 text-sm">
          <div className="text-center p-2 bg-green-500/10 rounded">
            <div className="text-muted-foreground">Total Bids</div>
            <div className="font-mono text-green-400">
              {processedData.bids.reduce((sum, bid) => sum + bid.quantity, 0).toFixed(4)}
            </div>
          </div>
          <div className="text-center p-2 bg-red-500/10 rounded">
            <div className="text-muted-foreground">Total Asks</div>
            <div className="font-mono text-red-400">
              {processedData.asks.reduce((sum, ask) => sum + ask.quantity, 0).toFixed(4)}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
