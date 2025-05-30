import React, { useState, useEffect } from 'react';
import { ChartBarIcon, AdjustmentsHorizontalIcon } from '@heroicons/react/24/outline';

interface VolumeProfileLevel {
  price: number;
  volume: number;
  buy_volume: number;
  sell_volume: number;
  delta: number;
  percentage: number;
  is_hvn: boolean;
  is_lvn: boolean;
  is_poc: boolean;
  is_value_area: boolean;
}

interface VolumeProfileData {
  symbol: string;
  profile_type: string;
  start_time: string;
  end_time: string;
  profile: {
    point_of_control: number;
    value_area_high: number;
    value_area_low: number;
    total_volume: number;
  };
  price_levels: VolumeProfileLevel[];
}

interface VolumeProfileProps {
  symbol: string;
  profileType?: string;
  className?: string;
}

const VolumeProfile: React.FC<VolumeProfileProps> = ({
  symbol,
  profileType = 'VPVR',
  className = ''
}) => {
  const [profileData, setProfileData] = useState<VolumeProfileData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedType, setSelectedType] = useState(profileType);

  useEffect(() => {
    fetchVolumeProfile();
  }, [symbol, selectedType]);

  const fetchVolumeProfile = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch(
        `/api/v1/orderflow/volume-profile/${symbol}?type=${selectedType}`
      );
      
      if (!response.ok) {
        throw new Error('Failed to fetch volume profile data');
      }
      
      const data = await response.json();
      if (data.success && data.data) {
        setProfileData(data.data);
      } else {
        setProfileData(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      setProfileData(null);
    } finally {
      setLoading(false);
    }
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

  const getVolumeBarWidth = (volume: number, maxVolume: number) => {
    return Math.max((volume / maxVolume) * 100, 1);
  };

  const getNodeColor = (level: VolumeProfileLevel) => {
    if (level.is_poc) return 'bg-purple-500';
    if (level.is_hvn) return 'bg-blue-500';
    if (level.is_lvn) return 'bg-gray-600';
    if (level.is_value_area) return 'bg-green-500';
    return 'bg-slate-500';
  };

  const getNodeLabel = (level: VolumeProfileLevel) => {
    if (level.is_poc) return 'POC';
    if (level.is_hvn) return 'HVN';
    if (level.is_lvn) return 'LVN';
    return '';
  };

  if (loading) {
    return (
      <div className={`card ${className}`}>
        <div className="card-header">
          <div className="flex items-center space-x-2">
            <ChartBarIcon className="h-5 w-5" />
            <h3 className="text-lg font-semibold">Volume Profile - {symbol}</h3>
          </div>
        </div>
        <div className="card-content">
          <div className="flex items-center justify-center h-64">
            <div className="loading-spinner"></div>
            <span className="ml-2">Loading volume profile...</span>
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
            <h3 className="text-lg font-semibold">Volume Profile - {symbol}</h3>
          </div>
        </div>
        <div className="card-content">
          <div className="text-center text-red-400 py-8">
            <p>Error loading volume profile: {error}</p>
            <button
              onClick={fetchVolumeProfile}
              className="btn-primary mt-4"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  const maxVolume = profileData?.price_levels.reduce((max, level) => 
    Math.max(max, level.volume), 0) || 1;

  return (
    <div className={`card ${className}`}>
      <div className="card-header">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <ChartBarIcon className="h-5 w-5" />
            <h3 className="text-lg font-semibold">Volume Profile - {symbol}</h3>
            <span className="text-sm text-gray-400">({selectedType})</span>
          </div>
          <div className="flex items-center space-x-2">
            <select
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value)}
              className="select-field text-sm"
            >
              <option value="VPVR">VPVR (Visible Range)</option>
              <option value="VPSV">VPSV (Session Volume)</option>
            </select>
            <AdjustmentsHorizontalIcon className="h-5 w-5 text-gray-400 cursor-pointer" />
          </div>
        </div>
      </div>
      
      <div className="card-content">
        {!profileData || profileData.price_levels.length === 0 ? (
          <div className="text-center text-gray-400 py-8">
            <p>No volume profile data available for {symbol}</p>
            <p className="text-sm mt-2">Try a different profile type or symbol</p>
          </div>
        ) : (
          <div className="space-y-4">
            {/* Key Levels Summary */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 bg-slate-800 rounded-lg">
              <div className="text-center">
                <div className="text-xs text-gray-400">Point of Control</div>
                <div className="text-lg font-mono text-purple-400">
                  ${formatPrice(profileData.profile.point_of_control)}
                </div>
              </div>
              <div className="text-center">
                <div className="text-xs text-gray-400">Value Area High</div>
                <div className="text-lg font-mono text-green-400">
                  ${formatPrice(profileData.profile.value_area_high)}
                </div>
              </div>
              <div className="text-center">
                <div className="text-xs text-gray-400">Value Area Low</div>
                <div className="text-lg font-mono text-green-400">
                  ${formatPrice(profileData.profile.value_area_low)}
                </div>
              </div>
              <div className="text-center">
                <div className="text-xs text-gray-400">Total Volume</div>
                <div className="text-lg font-mono">
                  {formatVolume(profileData.profile.total_volume)}
                </div>
              </div>
            </div>

            {/* Volume Profile Chart */}
            <div className="space-y-1 max-h-96 overflow-y-auto">
              {profileData.price_levels
                .sort((a, b) => b.price - a.price)
                .map((level, index) => (
                  <div
                    key={index}
                    className={`flex items-center space-x-2 py-1 px-2 rounded ${
                      level.is_value_area ? 'bg-green-900/20' : ''
                    }`}
                  >
                    {/* Price */}
                    <div className="w-20 text-xs font-mono text-right">
                      ${formatPrice(level.price)}
                    </div>
                    
                    {/* Volume Bar */}
                    <div className="flex-1 relative h-6">
                      {/* Background bar */}
                      <div className="absolute inset-0 bg-slate-700 rounded"></div>
                      
                      {/* Volume bar */}
                      <div
                        className={`absolute left-0 top-0 h-full rounded ${getNodeColor(level)}`}
                        style={{ width: `${getVolumeBarWidth(level.volume, maxVolume)}%` }}
                      ></div>
                      
                      {/* Buy/Sell split */}
                      <div className="absolute inset-0 flex">
                        <div
                          className="bg-green-500/60 h-full"
                          style={{ 
                            width: `${(level.buy_volume / level.volume) * getVolumeBarWidth(level.volume, maxVolume)}%` 
                          }}
                        ></div>
                        <div
                          className="bg-red-500/60 h-full"
                          style={{ 
                            width: `${(level.sell_volume / level.volume) * getVolumeBarWidth(level.volume, maxVolume)}%` 
                          }}
                        ></div>
                      </div>
                      
                      {/* Volume text */}
                      <div className="absolute inset-0 flex items-center justify-center">
                        <span className="text-xs font-mono text-white">
                          {formatVolume(level.volume)}
                        </span>
                      </div>
                    </div>
                    
                    {/* Percentage */}
                    <div className="w-12 text-xs font-mono text-right">
                      {level.percentage.toFixed(1)}%
                    </div>
                    
                    {/* Delta */}
                    <div className={`w-16 text-xs font-mono text-right ${
                      level.delta > 0 ? 'text-green-400' : 
                      level.delta < 0 ? 'text-red-400' : 'text-gray-400'
                    }`}>
                      {level.delta > 0 ? '+' : ''}{formatVolume(level.delta)}
                    </div>
                    
                    {/* Node Label */}
                    <div className="w-8">
                      {getNodeLabel(level) && (
                        <span className={`px-1 py-0.5 rounded text-xs ${
                          level.is_poc ? 'bg-purple-600 text-white' :
                          level.is_hvn ? 'bg-blue-600 text-white' :
                          level.is_lvn ? 'bg-gray-600 text-white' : ''
                        }`}>
                          {getNodeLabel(level)}
                        </span>
                      )}
                    </div>
                  </div>
                ))}
            </div>

            {/* Legend */}
            <div className="flex flex-wrap gap-4 text-xs">
              <div className="flex items-center space-x-1">
                <div className="w-3 h-3 bg-purple-500 rounded"></div>
                <span>Point of Control (POC)</span>
              </div>
              <div className="flex items-center space-x-1">
                <div className="w-3 h-3 bg-blue-500 rounded"></div>
                <span>High Volume Node (HVN)</span>
              </div>
              <div className="flex items-center space-x-1">
                <div className="w-3 h-3 bg-gray-600 rounded"></div>
                <span>Low Volume Node (LVN)</span>
              </div>
              <div className="flex items-center space-x-1">
                <div className="w-3 h-3 bg-green-500/30 rounded"></div>
                <span>Value Area (70%)</span>
              </div>
            </div>

            {/* Statistics */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm pt-4 border-t border-gray-700">
              <div>
                <span className="text-gray-400">HVN Count:</span>
                <span className="ml-2 font-mono text-blue-400">
                  {profileData.price_levels.filter(l => l.is_hvn).length}
                </span>
              </div>
              <div>
                <span className="text-gray-400">LVN Count:</span>
                <span className="ml-2 font-mono text-gray-400">
                  {profileData.price_levels.filter(l => l.is_lvn).length}
                </span>
              </div>
              <div>
                <span className="text-gray-400">Value Area:</span>
                <span className="ml-2 font-mono text-green-400">
                  {profileData.price_levels.filter(l => l.is_value_area).length} levels
                </span>
              </div>
              <div>
                <span className="text-gray-400">Price Range:</span>
                <span className="ml-2 font-mono">
                  ${formatPrice(Math.max(...profileData.price_levels.map(l => l.price)) - 
                    Math.min(...profileData.price_levels.map(l => l.price)))}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default VolumeProfile;
