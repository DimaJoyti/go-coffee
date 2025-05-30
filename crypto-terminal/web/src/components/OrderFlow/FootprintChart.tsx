import React, { useState, useEffect } from 'react';
import { ChartBarIcon, CogIcon } from '@heroicons/react/24/outline';

interface FootprintBar {
  id: string;
  symbol: string;
  price_level: number;
  buy_volume: number;
  sell_volume: number;
  total_volume: number;
  delta: number;
  is_imbalanced: boolean;
  is_point_of_control: boolean;
  start_time: string;
  end_time: string;
}

interface FootprintChartProps {
  symbol: string;
  timeframe?: string;
  className?: string;
}

const FootprintChart: React.FC<FootprintChartProps> = ({
  symbol,
  timeframe = '1h',
  className = ''
}) => {
  const [footprintData, setFootprintData] = useState<FootprintBar[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [config, setConfig] = useState({
    aggregationMethod: 'TIME',
    ticksPerRow: 100,
    showImbalances: true,
    showPOC: true,
    showDelta: true
  });

  useEffect(() => {
    fetchFootprintData();
  }, [symbol, timeframe, config.aggregationMethod]);

  const fetchFootprintData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch(
        `/api/v1/orderflow/footprint/${symbol}?timeframe=${timeframe}`
      );
      
      if (!response.ok) {
        throw new Error('Failed to fetch footprint data');
      }
      
      const data = await response.json();
      if (data.success && data.data?.bars) {
        setFootprintData(data.data.bars);
      } else {
        setFootprintData([]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      setFootprintData([]);
    } finally {
      setLoading(false);
    }
  };

  const getVolumeColor = (buyVolume: number, sellVolume: number) => {
    const total = buyVolume + sellVolume;
    if (total === 0) return 'bg-gray-500';
    
    const buyRatio = buyVolume / total;
    if (buyRatio > 0.7) return 'bg-green-500';
    if (buyRatio < 0.3) return 'bg-red-500';
    return 'bg-yellow-500';
  };

  const getDeltaColor = (delta: number) => {
    if (delta > 0) return 'text-green-400';
    if (delta < 0) return 'text-red-400';
    return 'text-gray-400';
  };

  const formatVolume = (volume: number) => {
    if (volume >= 1000000) return `${(volume / 1000000).toFixed(1)}M`;
    if (volume >= 1000) return `${(volume / 1000).toFixed(1)}K`;
    return volume.toFixed(0);
  };

  const formatPrice = (price: number) => {
    return price.toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    });
  };

  if (loading) {
    return (
      <div className={`card ${className}`}>
        <div className="card-header">
          <div className="flex items-center space-x-2">
            <ChartBarIcon className="h-5 w-5" />
            <h3 className="text-lg font-semibold">Footprint Chart - {symbol}</h3>
          </div>
        </div>
        <div className="card-content">
          <div className="flex items-center justify-center h-64">
            <div className="loading-spinner"></div>
            <span className="ml-2">Loading footprint data...</span>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`card ${className}`}>
        <div className="card-header">
          <div className="flex items-center space-x-2">
            <ChartBarIcon className="h-5 w-5" />
            <h3 className="text-lg font-semibold">Footprint Chart - {symbol}</h3>
          </div>
        </div>
        <div className="card-content">
          <div className="text-center text-red-400 py-8">
            <p>Error loading footprint data: {error}</p>
            <button
              onClick={fetchFootprintData}
              className="btn-primary mt-4"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`card ${className}`}>
      <div className="card-header">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <ChartBarIcon className="h-5 w-5" />
            <h3 className="text-lg font-semibold">Footprint Chart - {symbol}</h3>
            <span className="text-sm text-gray-400">({timeframe})</span>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setConfig(prev => ({ ...prev, showImbalances: !prev.showImbalances }))}
              className={`px-3 py-1 rounded text-xs ${
                config.showImbalances ? 'bg-blue-600 text-white' : 'bg-gray-600 text-gray-300'
              }`}
            >
              Imbalances
            </button>
            <button
              onClick={() => setConfig(prev => ({ ...prev, showPOC: !prev.showPOC }))}
              className={`px-3 py-1 rounded text-xs ${
                config.showPOC ? 'bg-purple-600 text-white' : 'bg-gray-600 text-gray-300'
              }`}
            >
              POC
            </button>
            <button
              onClick={() => setConfig(prev => ({ ...prev, showDelta: !prev.showDelta }))}
              className={`px-3 py-1 rounded text-xs ${
                config.showDelta ? 'bg-orange-600 text-white' : 'bg-gray-600 text-gray-300'
              }`}
            >
              Delta
            </button>
            <CogIcon className="h-5 w-5 text-gray-400 cursor-pointer" />
          </div>
        </div>
      </div>
      
      <div className="card-content">
        {footprintData.length === 0 ? (
          <div className="text-center text-gray-400 py-8">
            <p>No footprint data available for {symbol}</p>
            <p className="text-sm mt-2">Try a different timeframe or symbol</p>
          </div>
        ) : (
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {/* Header */}
            <div className="grid grid-cols-6 gap-2 text-xs font-medium text-gray-400 border-b border-gray-700 pb-2">
              <div>Price</div>
              <div>Buy Vol</div>
              <div>Sell Vol</div>
              <div>Total Vol</div>
              {config.showDelta && <div>Delta</div>}
              <div>Indicators</div>
            </div>
            
            {/* Footprint Bars */}
            {footprintData
              .sort((a, b) => b.price_level - a.price_level)
              .map((bar) => (
                <div
                  key={bar.id}
                  className={`grid grid-cols-6 gap-2 text-xs py-2 px-2 rounded ${
                    config.showPOC && bar.is_point_of_control
                      ? 'bg-purple-900/30 border border-purple-500'
                      : config.showImbalances && bar.is_imbalanced
                      ? 'bg-yellow-900/30 border border-yellow-500'
                      : 'hover:bg-slate-700/50'
                  }`}
                >
                  {/* Price */}
                  <div className="font-mono text-white">
                    ${formatPrice(bar.price_level)}
                  </div>
                  
                  {/* Buy Volume */}
                  <div className="text-green-400 font-mono">
                    {formatVolume(bar.buy_volume)}
                  </div>
                  
                  {/* Sell Volume */}
                  <div className="text-red-400 font-mono">
                    {formatVolume(bar.sell_volume)}
                  </div>
                  
                  {/* Total Volume */}
                  <div className="font-mono">
                    <div className="flex items-center space-x-1">
                      <div
                        className={`w-2 h-2 rounded ${getVolumeColor(bar.buy_volume, bar.sell_volume)}`}
                      ></div>
                      <span>{formatVolume(bar.total_volume)}</span>
                    </div>
                  </div>
                  
                  {/* Delta */}
                  {config.showDelta && (
                    <div className={`font-mono ${getDeltaColor(bar.delta)}`}>
                      {bar.delta > 0 ? '+' : ''}{formatVolume(bar.delta)}
                    </div>
                  )}
                  
                  {/* Indicators */}
                  <div className="flex space-x-1">
                    {config.showPOC && bar.is_point_of_control && (
                      <span className="px-1 py-0.5 bg-purple-600 text-white rounded text-xs">
                        POC
                      </span>
                    )}
                    {config.showImbalances && bar.is_imbalanced && (
                      <span className="px-1 py-0.5 bg-yellow-600 text-white rounded text-xs">
                        IMB
                      </span>
                    )}
                  </div>
                </div>
              ))}
          </div>
        )}
        
        {/* Summary */}
        {footprintData.length > 0 && (
          <div className="mt-4 pt-4 border-t border-gray-700">
            <div className="grid grid-cols-4 gap-4 text-sm">
              <div>
                <span className="text-gray-400">Total Bars:</span>
                <span className="ml-2 font-mono">{footprintData.length}</span>
              </div>
              <div>
                <span className="text-gray-400">POC Levels:</span>
                <span className="ml-2 font-mono text-purple-400">
                  {footprintData.filter(bar => bar.is_point_of_control).length}
                </span>
              </div>
              <div>
                <span className="text-gray-400">Imbalances:</span>
                <span className="ml-2 font-mono text-yellow-400">
                  {footprintData.filter(bar => bar.is_imbalanced).length}
                </span>
              </div>
              <div>
                <span className="text-gray-400">Total Volume:</span>
                <span className="ml-2 font-mono">
                  {formatVolume(footprintData.reduce((sum, bar) => sum + bar.total_volume, 0))}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default FootprintChart;
