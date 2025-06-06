import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { RefreshCw, TrendingUp, AlertTriangle, DollarSign } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ArbitrageOpportunity {
  symbol: string;
  buy_exchange: string;
  sell_exchange: string;
  buy_price: number;
  sell_price: number;
  price_difference: number;
  profit_percent: number;
  volume: number;
  timestamp: string;
  confidence: number;
}

interface ArbitrageOpportunitiesProps {
  className?: string;
}

const ArbitrageOpportunities: React.FC<ArbitrageOpportunitiesProps> = ({ className }) => {
  const [opportunities, setOpportunities] = useState<ArbitrageOpportunity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const fetchOpportunities = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v2/arbitrage/opportunities?min_profit=0.5');
      
      if (!response.ok) {
        throw new Error('Failed to fetch arbitrage opportunities');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setOpportunities(data.data || []);
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
    fetchOpportunities();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchOpportunities, 30000);
    
    return () => clearInterval(interval);
  }, []);

  const getProfitColor = (profit: number) => {
    if (profit >= 2) return 'text-green-600 bg-green-50';
    if (profit >= 1) return 'text-yellow-600 bg-yellow-50';
    return 'text-blue-600 bg-blue-50';
  };

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.8) return 'text-green-600';
    if (confidence >= 0.6) return 'text-yellow-600';
    return 'text-red-600';
  };

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

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-xl font-bold flex items-center gap-2">
          <TrendingUp className="h-5 w-5 text-green-600" />
          Arbitrage Opportunities
        </CardTitle>
        <div className="flex items-center gap-2">
          {lastUpdate && (
            <span className="text-sm text-gray-500">
              Last updated: {lastUpdate.toLocaleTimeString()}
            </span>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={fetchOpportunities}
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
            <AlertTriangle className="h-5 w-5 text-red-600" />
            <span className="text-red-700">{error}</span>
          </div>
        )}
        
        {loading && opportunities.length === 0 ? (
          <div className="flex items-center justify-center py-8">
            <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
            <span className="ml-2 text-gray-500">Loading opportunities...</span>
          </div>
        ) : opportunities.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <TrendingUp className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p>No arbitrage opportunities found</p>
            <p className="text-sm">Try lowering the minimum profit threshold</p>
          </div>
        ) : (
          <div className="space-y-4">
            {opportunities.slice(0, 10).map((opportunity, index) => (
              <div
                key={`${opportunity.symbol}-${opportunity.buy_exchange}-${opportunity.sell_exchange}-${index}`}
                className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <h3 className="font-semibold text-lg">{opportunity.symbol}</h3>
                    <Badge className={cn('font-medium', getProfitColor(opportunity.profit_percent))}>
                      +{opportunity.profit_percent.toFixed(2)}%
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={cn('text-sm font-medium', getConfidenceColor(opportunity.confidence))}>
                      {(opportunity.confidence * 100).toFixed(0)}% confidence
                    </span>
                  </div>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Buy from:</span>
                      <Badge variant="outline" className="text-green-600 border-green-200">
                        {opportunity.buy_exchange}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Price:</span>
                      <span className="font-medium">{formatCurrency(opportunity.buy_price)}</span>
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Sell to:</span>
                      <Badge variant="outline" className="text-red-600 border-red-200">
                        {opportunity.sell_exchange}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Price:</span>
                      <span className="font-medium">{formatCurrency(opportunity.sell_price)}</span>
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Profit:</span>
                      <span className="font-medium text-green-600 flex items-center gap-1">
                        <DollarSign className="h-4 w-4" />
                        {formatCurrency(opportunity.price_difference)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Volume:</span>
                      <span className="font-medium">{formatVolume(opportunity.volume)}</span>
                    </div>
                  </div>
                </div>
                
                <div className="mt-3 pt-3 border-t border-gray-100">
                  <div className="flex items-center justify-between text-sm text-gray-500">
                    <span>Updated: {new Date(opportunity.timestamp).toLocaleTimeString()}</span>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-xs"
                      onClick={() => {
                        // In a real implementation, this would open a trading interface
                        alert(`Execute arbitrage for ${opportunity.symbol}`);
                      }}
                    >
                      Execute Trade
                    </Button>
                  </div>
                </div>
              </div>
            ))}
            
            {opportunities.length > 10 && (
              <div className="text-center pt-4">
                <Button variant="outline" size="sm">
                  View All {opportunities.length} Opportunities
                </Button>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default ArbitrageOpportunities;
