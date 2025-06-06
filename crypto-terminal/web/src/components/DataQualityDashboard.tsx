import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { RefreshCw, Shield, AlertTriangle, CheckCircle, Clock, Zap } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DataQualityMetrics {
  quality_score: number;
  availability: number;
  latency_ms: number;
  error_rate: number;
  last_update: string;
}

interface ExchangeStatus {
  [exchange: string]: string;
}

interface DataQuality {
  [exchange: string]: {
    [symbol: string]: DataQualityMetrics;
  };
}

interface DataQualityDashboardProps {
  className?: string;
}

const DataQualityDashboard: React.FC<DataQualityDashboardProps> = ({ className }) => {
  const [dataQuality, setDataQuality] = useState<DataQuality>({});
  const [exchangeStatus, setExchangeStatus] = useState<ExchangeStatus>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const fetchDataQuality = async () => {
    try {
      setLoading(true);
      
      // Fetch data quality metrics
      const qualityResponse = await fetch('/api/v2/market/data-quality');
      const statusResponse = await fetch('/api/v2/market/exchanges/status');
      
      if (!qualityResponse.ok || !statusResponse.ok) {
        throw new Error('Failed to fetch data quality metrics');
      }
      
      const qualityData = await qualityResponse.json();
      const statusData = await statusResponse.json();
      
      if (qualityData.success && statusData.success) {
        setDataQuality(qualityData.data || {});
        setExchangeStatus(statusData.data || {});
        setLastUpdate(new Date());
        setError(null);
      } else {
        throw new Error(qualityData.error || statusData.error || 'Unknown error');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDataQuality();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchDataQuality, 30000);
    
    return () => clearInterval(interval);
  }, []);

  const getQualityColor = (score: number) => {
    if (score >= 0.8) return 'text-green-600';
    if (score >= 0.6) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getQualityBadgeColor = (score: number) => {
    if (score >= 0.8) return 'bg-green-100 text-green-800 border-green-200';
    if (score >= 0.6) return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    return 'bg-red-100 text-red-800 border-red-200';
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'connected':
        return 'text-green-600';
      case 'disconnected':
        return 'text-red-600';
      default:
        return 'text-yellow-600';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'connected':
        return <CheckCircle className="h-4 w-4" />;
      case 'disconnected':
        return <AlertTriangle className="h-4 w-4" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  const formatLatency = (latency: number) => {
    if (latency < 1000) {
      return `${latency.toFixed(0)}ms`;
    }
    return `${(latency / 1000).toFixed(1)}s`;
  };

  const calculateOverallQuality = () => {
    let totalScore = 0;
    let count = 0;
    
    Object.values(dataQuality).forEach(exchange => {
      Object.values(exchange).forEach(metrics => {
        totalScore += metrics.quality_score;
        count++;
      });
    });
    
    return count > 0 ? totalScore / count : 0;
  };

  const overallQuality = calculateOverallQuality();

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-xl font-bold flex items-center gap-2">
          <Shield className="h-5 w-5 text-blue-600" />
          Data Quality Dashboard
        </CardTitle>
        <div className="flex items-center gap-3">
          {lastUpdate && (
            <span className="text-sm text-gray-500">
              Last updated: {lastUpdate.toLocaleTimeString()}
            </span>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={fetchDataQuality}
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
        
        {loading && Object.keys(dataQuality).length === 0 ? (
          <div className="flex items-center justify-center py-8">
            <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
            <span className="ml-2 text-gray-500">Loading data quality metrics...</span>
          </div>
        ) : (
          <div className="space-y-6">
            {/* Overall Quality Score */}
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-semibold text-lg">Overall Data Quality</h3>
                <Badge className={cn('font-medium', getQualityBadgeColor(overallQuality))}>
                  {(overallQuality * 100).toFixed(0)}%
                </Badge>
              </div>
              <Progress 
                value={overallQuality * 100} 
                className="h-3"
              />
              <p className="text-sm text-gray-600 mt-2">
                Aggregated quality score across all exchanges and symbols
              </p>
            </div>
            
            {/* Exchange Status */}
            <div className="space-y-3">
              <h3 className="font-semibold text-lg">Exchange Status</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {Object.entries(exchangeStatus).map(([exchange, status]) => (
                  <div
                    key={exchange}
                    className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                  >
                    <div className="flex items-center gap-2">
                      <div className={cn('flex items-center gap-1', getStatusColor(status))}>
                        {getStatusIcon(status)}
                      </div>
                      <span className="font-medium">{exchange}</span>
                    </div>
                    <Badge 
                      variant="outline" 
                      className={cn(getStatusColor(status), 'border-current')}
                    >
                      {status}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
            
            {/* Detailed Quality Metrics */}
            <div className="space-y-4">
              <h3 className="font-semibold text-lg">Quality Metrics by Exchange</h3>
              {Object.entries(dataQuality).map(([exchange, symbols]) => (
                <div key={exchange} className="border border-gray-200 rounded-lg overflow-hidden">
                  <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
                    <h4 className="font-semibold flex items-center gap-2">
                      {exchange}
                      <Badge 
                        variant="outline" 
                        className={cn(getStatusColor(exchangeStatus[exchange] || 'unknown'), 'border-current')}
                      >
                        {exchangeStatus[exchange] || 'Unknown'}
                      </Badge>
                    </h4>
                  </div>
                  
                  <div className="p-4 space-y-3">
                    {Object.entries(symbols).map(([symbol, metrics]) => (
                      <div
                        key={symbol}
                        className="flex items-center justify-between p-3 bg-white border border-gray-100 rounded-lg"
                      >
                        <div className="flex items-center gap-3">
                          <span className="font-medium">{symbol}</span>
                          <Badge className={cn('text-xs', getQualityBadgeColor(metrics.quality_score))}>
                            {(metrics.quality_score * 100).toFixed(0)}%
                          </Badge>
                        </div>
                        
                        <div className="grid grid-cols-4 gap-4 text-right text-sm">
                          <div>
                            <p className="text-gray-600">Availability</p>
                            <p className={cn('font-medium', getQualityColor(metrics.availability))}>
                              {(metrics.availability * 100).toFixed(1)}%
                            </p>
                          </div>
                          <div>
                            <p className="text-gray-600">Latency</p>
                            <p className="font-medium flex items-center gap-1">
                              <Zap className="h-3 w-3" />
                              {formatLatency(metrics.latency_ms)}
                            </p>
                          </div>
                          <div>
                            <p className="text-gray-600">Error Rate</p>
                            <p className={cn('font-medium', metrics.error_rate > 0.1 ? 'text-red-600' : 'text-green-600')}>
                              {(metrics.error_rate * 100).toFixed(1)}%
                            </p>
                          </div>
                          <div>
                            <p className="text-gray-600">Last Update</p>
                            <p className="font-medium text-gray-500">
                              {new Date(metrics.last_update).toLocaleTimeString()}
                            </p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
            
            {Object.keys(dataQuality).length === 0 && (
              <div className="text-center py-8 text-gray-500">
                <Shield className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p>No data quality metrics available</p>
                <p className="text-sm">Check exchange connections and try again</p>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default DataQualityDashboard;
