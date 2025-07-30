import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { Badge } from '../ui/badge';
import { Separator } from '../ui/separator';
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Percent,
  Calculator,
  AlertTriangle,
  CheckCircle,
  Clock
} from 'lucide-react';

interface TradingPanelProps {
  symbol?: string;
  currentPrice?: number;
  onOrderSubmit?: (order: OrderData) => void;
  className?: string;
}

interface OrderData {
  symbol: string;
  side: 'BUY' | 'SELL';
  type: 'MARKET' | 'LIMIT' | 'STOP_LOSS' | 'TAKE_PROFIT';
  quantity: number;
  price?: number;
  stopPrice?: number;
  timeInForce: 'GTC' | 'IOC' | 'FOK';
}

interface AccountBalance {
  asset: string;
  free: number;
  locked: number;
}

export const EnhancedTradingPanel: React.FC<TradingPanelProps> = ({
  symbol = 'BTCUSDT',
  currentPrice = 43245.75,
  onOrderSubmit,
  className
}) => {
  const [orderType, setOrderType] = useState<'MARKET' | 'LIMIT' | 'STOP_LOSS'>('LIMIT');
  const [side, setSide] = useState<'BUY' | 'SELL'>('BUY');
  const [quantity, setQuantity] = useState<string>('');
  const [price, setPrice] = useState<string>(currentPrice.toString());
  const [stopPrice, setStopPrice] = useState<string>('');
  const [total, setTotal] = useState<number>(0);
  const [percentage, setPercentage] = useState<string>('');
  
  // Mock account balance
  const [balance] = useState<AccountBalance[]>([
    { asset: 'USDT', free: 10000.50, locked: 500.25 },
    { asset: 'BTC', free: 0.25, locked: 0.125 }
  ]);

  const baseAsset = symbol.replace('USDT', '');
  const quoteAsset = 'USDT';
  
  const availableBalance = balance.find(b => 
    b.asset === (side === 'BUY' ? quoteAsset : baseAsset)
  )?.free || 0;

  // Calculate total when quantity or price changes
  useEffect(() => {
    const qty = parseFloat(quantity) || 0;
    const prc = parseFloat(price) || currentPrice;
    setTotal(qty * prc);
  }, [quantity, price, currentPrice]);

  // Handle percentage-based quantity calculation
  const handlePercentageChange = (percent: string) => {
    setPercentage(percent);
    const pct = parseFloat(percent) || 0;
    
    if (side === 'BUY') {
      const maxTotal = availableBalance * (pct / 100);
      const prc = parseFloat(price) || currentPrice;
      const qty = maxTotal / prc;
      setQuantity(qty.toFixed(6));
    } else {
      const maxQty = availableBalance * (pct / 100);
      setQuantity(maxQty.toFixed(6));
    }
  };

  const handleSubmitOrder = () => {
    if (!quantity || parseFloat(quantity) <= 0) return;
    
    const orderData: OrderData = {
      symbol,
      side,
      type: orderType,
      quantity: parseFloat(quantity),
      price: orderType !== 'MARKET' ? parseFloat(price) : undefined,
      stopPrice: orderType === 'STOP_LOSS' ? parseFloat(stopPrice) : undefined,
      timeInForce: 'GTC'
    };
    
    onOrderSubmit?.(orderData);
    
    // Reset form
    setQuantity('');
    setPercentage('');
  };

  const isFormValid = () => {
    if (!quantity || parseFloat(quantity) <= 0) return false;
    if (orderType !== 'MARKET' && (!price || parseFloat(price) <= 0)) return false;
    if (orderType === 'STOP_LOSS' && (!stopPrice || parseFloat(stopPrice) <= 0)) return false;
    return true;
  };

  return (
    <Card className={`trading-card-glass ${className}`}>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between text-lg">
          <span>Trade {baseAsset}</span>
          <Badge variant="outline" className="text-xs">
            ${currentPrice.toLocaleString()}
          </Badge>
        </CardTitle>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Buy/Sell Toggle */}
        <div className="grid grid-cols-2 gap-2">
          <Button
            variant={side === 'BUY' ? 'default' : 'outline'}
            className={`${side === 'BUY' ? 'buy-button' : ''} transition-all`}
            onClick={() => setSide('BUY')}
          >
            <TrendingUp className="w-4 h-4 mr-2" />
            Buy
          </Button>
          <Button
            variant={side === 'SELL' ? 'default' : 'outline'}
            className={`${side === 'SELL' ? 'sell-button' : ''} transition-all`}
            onClick={() => setSide('SELL')}
          >
            <TrendingDown className="w-4 h-4 mr-2" />
            Sell
          </Button>
        </div>

        {/* Order Type Selection */}
        <Tabs value={orderType} onValueChange={(value) => setOrderType(value as any)}>
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="MARKET">Market</TabsTrigger>
            <TabsTrigger value="LIMIT">Limit</TabsTrigger>
            <TabsTrigger value="STOP_LOSS">Stop</TabsTrigger>
          </TabsList>
          
          <TabsContent value="MARKET" className="space-y-4 mt-4">
            <div className="text-sm text-muted-foreground">
              Execute immediately at market price
            </div>
          </TabsContent>
          
          <TabsContent value="LIMIT" className="space-y-4 mt-4">
            <div>
              <Label htmlFor="price">Price ({quoteAsset})</Label>
              <Input
                id="price"
                type="number"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                placeholder="0.00"
                className="mt-1"
              />
            </div>
          </TabsContent>
          
          <TabsContent value="STOP_LOSS" className="space-y-4 mt-4">
            <div>
              <Label htmlFor="stopPrice">Stop Price ({quoteAsset})</Label>
              <Input
                id="stopPrice"
                type="number"
                value={stopPrice}
                onChange={(e) => setStopPrice(e.target.value)}
                placeholder="0.00"
                className="mt-1"
              />
            </div>
          </TabsContent>
        </Tabs>

        {/* Quantity Input */}
        <div>
          <Label htmlFor="quantity">Quantity ({baseAsset})</Label>
          <Input
            id="quantity"
            type="number"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            placeholder="0.00"
            className="mt-1"
          />
        </div>

        {/* Percentage Buttons */}
        <div className="grid grid-cols-4 gap-2">
          {['25', '50', '75', '100'].map((pct) => (
            <Button
              key={pct}
              variant="outline"
              size="sm"
              onClick={() => handlePercentageChange(pct)}
              className={`text-xs ${percentage === pct ? 'bg-primary text-primary-foreground' : ''}`}
            >
              {pct}%
            </Button>
          ))}
        </div>

        <Separator />

        {/* Order Summary */}
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Total:</span>
            <span className="font-medium">
              {total.toFixed(2)} {quoteAsset}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Available:</span>
            <span className="font-medium">
              {availableBalance.toFixed(6)} {side === 'BUY' ? quoteAsset : baseAsset}
            </span>
          </div>
          {orderType !== 'MARKET' && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Est. Fee:</span>
              <span className="font-medium">
                {(total * 0.001).toFixed(4)} {quoteAsset}
              </span>
            </div>
          )}
        </div>

        {/* Submit Button */}
        <Button
          onClick={handleSubmitOrder}
          disabled={!isFormValid()}
          className={`w-full ${side === 'BUY' ? 'buy-button' : 'sell-button'} transition-all`}
        >
          {side === 'BUY' ? (
            <>
              <TrendingUp className="w-4 h-4 mr-2" />
              Buy {baseAsset}
            </>
          ) : (
            <>
              <TrendingDown className="w-4 h-4 mr-2" />
              Sell {baseAsset}
            </>
          )}
        </Button>

        {/* Risk Warning */}
        <div className="flex items-start space-x-2 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
          <AlertTriangle className="w-4 h-4 text-yellow-500 mt-0.5 flex-shrink-0" />
          <div className="text-xs text-yellow-500">
            Trading involves risk. Only trade with funds you can afford to lose.
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
