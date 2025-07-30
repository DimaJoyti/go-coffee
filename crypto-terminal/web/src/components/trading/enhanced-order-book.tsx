import React, { useState, useEffect, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Separator } from '../ui/separator';
import { 
  TrendingUp, 
  TrendingDown, 
  MoreVertical,
  Layers,
  BarChart3,
  Activity
} from 'lucide-react';

interface OrderBookLevel {
  price: number;
  quantity: number;
  total: number;
  count?: number;
}

interface OrderBookData {
  symbol: string;
  lastUpdateId: number;
  bids: OrderBookLevel[];
  asks: OrderBookLevel[];
}

interface EnhancedOrderBookProps {
  data?: OrderBookData;
  symbol?: string;
  precision?: number;
  maxLevels?: number;
  showDepth?: boolean;
  showSpread?: boolean;
  onPriceClick?: (price: number, side: 'bid' | 'ask') => void;
  className?: string;
}

// Mock data generator
const generateMockOrderBook = (symbol: string, basePrice: number): OrderBookData => {
  const bids: OrderBookLevel[] = [];
  const asks: OrderBookLevel[] = [];
  
  // Generate bids (buy orders) - prices below current price
  let totalBid = 0;
  for (let i = 0; i < 15; i++) {
    const price = basePrice - (i + 1) * 0.25;
    const quantity = Math.random() * 5 + 0.1;
    totalBid += quantity;
    bids.push({
      price,
      quantity,
      total: totalBid,
      count: Math.floor(Math.random() * 10) + 1
    });
  }
  
  // Generate asks (sell orders) - prices above current price
  let totalAsk = 0;
  for (let i = 0; i < 15; i++) {
    const price = basePrice + (i + 1) * 0.25;
    const quantity = Math.random() * 5 + 0.1;
    totalAsk += quantity;
    asks.push({
      price,
      quantity,
      total: totalAsk,
      count: Math.floor(Math.random() * 10) + 1
    });
  }
  
  return {
    symbol,
    lastUpdateId: Date.now(),
    bids,
    asks
  };
};

export const EnhancedOrderBook: React.FC<EnhancedOrderBookProps> = ({
  data,
  symbol = 'BTCUSDT',
  precision = 2,
  maxLevels = 15,
  showDepth = true,
  showSpread = true,
  onPriceClick,
  className
}) => {
  const [grouping, setGrouping] = useState(0.01);
  const [viewMode, setViewMode] = useState<'combined' | 'bids' | 'asks'>('combined');
  
  // Use mock data if no data provided
  const orderBookData = data || generateMockOrderBook(symbol, 43245.75);
  
  // Group order book levels by price grouping
  const groupedData = useMemo(() => {
    const groupLevel = (level: OrderBookLevel) => {
      const groupedPrice = Math.floor(level.price / grouping) * grouping;
      return { ...level, price: groupedPrice };
    };
    
    const groupLevels = (levels: OrderBookLevel[]) => {
      const grouped = new Map<number, OrderBookLevel>();
      
      levels.forEach(level => {
        const groupedLevel = groupLevel(level);
        const existing = grouped.get(groupedLevel.price);
        
        if (existing) {
          existing.quantity += groupedLevel.quantity;
          existing.count = (existing.count || 0) + (groupedLevel.count || 0);
        } else {
          grouped.set(groupedLevel.price, { ...groupedLevel });
        }
      });
      
      return Array.from(grouped.values());
    };
    
    return {
      bids: groupLevels(orderBookData.bids).slice(0, maxLevels),
      asks: groupLevels(orderBookData.asks).slice(0, maxLevels)
    };
  }, [orderBookData, grouping, maxLevels]);
  
  // Calculate spread
  const spread = useMemo(() => {
    if (groupedData.asks.length === 0 || groupedData.bids.length === 0) return null;
    
    const bestAsk = Math.min(...groupedData.asks.map(a => a.price));
    const bestBid = Math.max(...groupedData.bids.map(b => b.price));
    const spreadValue = bestAsk - bestBid;
    const spreadPercent = (spreadValue / bestBid) * 100;
    
    return { value: spreadValue, percent: spreadPercent, bestBid, bestAsk };
  }, [groupedData]);
  
  // Calculate max quantity for depth visualization
  const maxQuantity = useMemo(() => {
    const allQuantities = [...groupedData.bids, ...groupedData.asks].map(l => l.quantity);
    return Math.max(...allQuantities);
  }, [groupedData]);
  
  const OrderBookRow: React.FC<{
    level: OrderBookLevel;
    side: 'bid' | 'ask';
    maxQty: number;
  }> = ({ level, side, maxQty }) => {
    const depthPercent = showDepth ? (level.quantity / maxQty) * 100 : 0;
    const isBid = side === 'bid';
    
    return (
      <div
        className={`
          relative grid grid-cols-3 gap-2 py-1 px-2 text-xs cursor-pointer
          hover:bg-muted/30 transition-colors
          ${isBid ? 'order-book-bid' : 'order-book-ask'}
        `}
        onClick={() => onPriceClick?.(level.price, side)}
      >
        {/* Depth visualization */}
        {showDepth && (
          <div
            className={`
              absolute inset-y-0 ${isBid ? 'right-0' : 'left-0'}
              ${isBid ? 'bg-green-500/10' : 'bg-red-500/10'}
              transition-all duration-300
            `}
            style={{ width: `${depthPercent}%` }}
          />
        )}
        
        {/* Price */}
        <div className={`text-right font-mono ${isBid ? 'text-green-400' : 'text-red-400'}`}>
          {level.price.toFixed(precision)}
        </div>
        
        {/* Quantity */}
        <div className="text-right font-mono text-muted-foreground">
          {level.quantity.toFixed(6)}
        </div>
        
        {/* Total */}
        <div className="text-right font-mono text-muted-foreground">
          {level.total.toFixed(6)}
        </div>
      </div>
    );
  };
  
  return (
    <Card className={`trading-card-glass ${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Order Book</CardTitle>
          <div className="flex items-center space-x-2">
            <Badge variant="outline" className="text-xs">
              {symbol}
            </Badge>
            <Button variant="ghost" size="sm">
              <MoreVertical className="w-4 h-4" />
            </Button>
          </div>
        </div>
        
        {/* Controls */}
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center space-x-2">
            <span className="text-muted-foreground">Group:</span>
            <select
              value={grouping}
              onChange={(e) => setGrouping(parseFloat(e.target.value))}
              className="bg-background border border-border rounded px-2 py-1 text-xs"
            >
              <option value={0.01}>0.01</option>
              <option value={0.1}>0.1</option>
              <option value={1}>1.0</option>
              <option value={10}>10.0</option>
            </select>
          </div>
          
          <div className="flex items-center space-x-1">
            <Button
              variant={viewMode === 'combined' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setViewMode('combined')}
              className="h-6 px-2 text-xs"
            >
              <Layers className="w-3 h-3" />
            </Button>
            <Button
              variant={viewMode === 'bids' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setViewMode('bids')}
              className="h-6 px-2 text-xs"
            >
              <TrendingUp className="w-3 h-3" />
            </Button>
            <Button
              variant={viewMode === 'asks' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setViewMode('asks')}
              className="h-6 px-2 text-xs"
            >
              <TrendingDown className="w-3 h-3" />
            </Button>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="p-0">
        {/* Header */}
        <div className="grid grid-cols-3 gap-2 py-2 px-2 text-xs font-medium text-muted-foreground border-b border-border">
          <div className="text-right">Price</div>
          <div className="text-right">Size</div>
          <div className="text-right">Total</div>
        </div>
        
        {/* Order Book Content */}
        <div className="max-h-96 overflow-y-auto scrollbar-thin">
          {/* Asks (Sell Orders) */}
          {(viewMode === 'combined' || viewMode === 'asks') && (
            <div className="space-y-0">
              {groupedData.asks.reverse().map((ask, index) => (
                <OrderBookRow
                  key={`ask-${ask.price}-${index}`}
                  level={ask}
                  side="ask"
                  maxQty={maxQuantity}
                />
              ))}
            </div>
          )}
          
          {/* Spread */}
          {viewMode === 'combined' && showSpread && spread && (
            <div className="py-2 px-2 border-y border-border bg-muted/20">
              <div className="text-center text-xs">
                <div className="text-muted-foreground">Spread</div>
                <div className="font-mono">
                  {spread.value.toFixed(precision)} ({spread.percent.toFixed(3)}%)
                </div>
              </div>
            </div>
          )}
          
          {/* Bids (Buy Orders) */}
          {(viewMode === 'combined' || viewMode === 'bids') && (
            <div className="space-y-0">
              {groupedData.bids.map((bid, index) => (
                <OrderBookRow
                  key={`bid-${bid.price}-${index}`}
                  level={bid}
                  side="bid"
                  maxQty={maxQuantity}
                />
              ))}
            </div>
          )}
        </div>
        
        {/* Footer Stats */}
        <div className="border-t border-border p-2">
          <div className="grid grid-cols-2 gap-4 text-xs">
            <div className="text-center">
              <div className="text-muted-foreground">Total Bids</div>
              <div className="font-mono text-green-400">
                {groupedData.bids.reduce((sum, bid) => sum + bid.quantity, 0).toFixed(4)}
              </div>
            </div>
            <div className="text-center">
              <div className="text-muted-foreground">Total Asks</div>
              <div className="font-mono text-red-400">
                {groupedData.asks.reduce((sum, ask) => sum + ask.quantity, 0).toFixed(4)}
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
