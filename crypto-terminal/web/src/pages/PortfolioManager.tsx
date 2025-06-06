import React, { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  PieChart, 
  BarChart3, 
  TrendingUp, 
  Shield, 
  Activity,
  Eye,
  Target,
  Zap
} from 'lucide-react';

import PortfolioManagerDashboard from '../components/PortfolioManagerDashboard';
import MarketHeatmap from '../components/MarketHeatmap';
import TradingViewWidget from '../components/TradingViewWidget';
import DataQualityDashboard from '../components/DataQualityDashboard';

const PortfolioManager: React.FC = () => {
  const [selectedPortfolio, setSelectedPortfolio] = useState('portfolio-1');
  const [activeTab, setActiveTab] = useState('dashboard');

  // Sample portfolio data
  const portfolios = [
    {
      id: 'portfolio-1',
      name: 'Institutional Portfolio',
      value: 1250000.50,
      return: 11.11,
      risk: 6.5,
      assets: 12
    },
    {
      id: 'portfolio-2',
      name: 'DeFi Growth Fund',
      value: 850000.25,
      return: 18.75,
      risk: 8.2,
      assets: 8
    },
    {
      id: 'portfolio-3',
      name: 'Conservative Holdings',
      value: 2100000.00,
      return: 7.25,
      risk: 4.1,
      assets: 5
    }
  ];

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const getRiskColor = (score: number) => {
    if (score <= 3) return 'text-green-600';
    if (score <= 6) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getRiskBadgeColor = (score: number) => {
    if (score <= 3) return 'bg-green-100 text-green-800 border-green-200';
    if (score <= 6) return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    return 'bg-red-100 text-red-800 border-red-200';
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Portfolio Manager</h1>
            <p className="text-gray-600 mt-1">
              Professional crypto portfolio management with real-time analytics and risk assessment
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Badge variant="outline" className="text-blue-600 border-blue-200">
              <Zap className="h-3 w-3 mr-1" />
              Real-time Data
            </Badge>
            <Badge variant="outline" className="text-green-600 border-green-200">
              <Shield className="h-3 w-3 mr-1" />
              Risk Monitored
            </Badge>
          </div>
        </div>

        {/* Portfolio Selection */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PieChart className="h-5 w-5 text-blue-600" />
              Portfolio Selection
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {portfolios.map((portfolio) => (
                <div
                  key={portfolio.id}
                  className={`p-4 border-2 rounded-lg cursor-pointer transition-all ${
                    selectedPortfolio === portfolio.id
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                  onClick={() => setSelectedPortfolio(portfolio.id)}
                >
                  <div className="flex items-center justify-between mb-3">
                    <h3 className="font-semibold">{portfolio.name}</h3>
                    {selectedPortfolio === portfolio.id && (
                      <Badge className="bg-blue-100 text-blue-800">Active</Badge>
                    )}
                  </div>
                  
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Total Value</span>
                      <span className="font-medium">{formatCurrency(portfolio.value)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Return</span>
                      <span className="font-medium text-green-600">+{portfolio.return}%</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Risk Score</span>
                      <span className={`font-medium ${getRiskColor(portfolio.risk)}`}>
                        {portfolio.risk}/10
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Assets</span>
                      <span className="font-medium">{portfolio.assets}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Main Dashboard Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-5">
            <TabsTrigger value="dashboard" className="flex items-center gap-2">
              <PieChart className="h-4 w-4" />
              Dashboard
            </TabsTrigger>
            <TabsTrigger value="market-data" className="flex items-center gap-2">
              <BarChart3 className="h-4 w-4" />
              Market Data
            </TabsTrigger>
            <TabsTrigger value="heatmap" className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              Heatmap
            </TabsTrigger>
            <TabsTrigger value="analytics" className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              Analytics
            </TabsTrigger>
            <TabsTrigger value="quality" className="flex items-center gap-2">
              <Shield className="h-4 w-4" />
              Data Quality
            </TabsTrigger>
          </TabsList>

          <TabsContent value="dashboard" className="space-y-6">
            <PortfolioManagerDashboard 
              portfolioId={selectedPortfolio}
              className="w-full"
            />
          </TabsContent>

          <TabsContent value="market-data" className="space-y-6">
            <TradingViewWidget className="w-full" />
          </TabsContent>

          <TabsContent value="heatmap" className="space-y-6">
            <MarketHeatmap className="w-full" />
          </TabsContent>

          <TabsContent value="analytics" className="space-y-6">
            {/* Advanced Analytics Section */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Target className="h-5 w-5 text-purple-600" />
                    Risk Analysis
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="p-4 bg-gradient-to-r from-purple-50 to-blue-50 rounded-lg">
                      <h4 className="font-semibold mb-2">Portfolio Risk Assessment</h4>
                      <p className="text-sm text-gray-600 mb-3">
                        Comprehensive risk analysis using Value at Risk (VaR), stress testing, and correlation analysis.
                      </p>
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <p className="text-xs text-gray-500">VaR (95%)</p>
                          <p className="font-bold text-red-600">-$45,000</p>
                        </div>
                        <div>
                          <p className="text-xs text-gray-500">Max Drawdown</p>
                          <p className="font-bold text-red-600">-15.25%</p>
                        </div>
                      </div>
                    </div>
                    
                    <div className="space-y-2">
                      <h5 className="font-medium">Stress Test Scenarios</h5>
                      <div className="space-y-2">
                        <div className="flex justify-between text-sm">
                          <span>Market Crash 2008</span>
                          <span className="text-red-600">-34.0%</span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span>Regulatory Crackdown</span>
                          <span className="text-red-600">-25.0%</span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span>Exchange Hack</span>
                          <span className="text-red-600">-15.0%</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Eye className="h-5 w-5 text-green-600" />
                    Performance Metrics
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="p-4 bg-gradient-to-r from-green-50 to-emerald-50 rounded-lg">
                      <h4 className="font-semibold mb-2">Risk-Adjusted Returns</h4>
                      <p className="text-sm text-gray-600 mb-3">
                        Advanced performance metrics including Sharpe ratio, Sortino ratio, and alpha generation.
                      </p>
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <p className="text-xs text-gray-500">Sharpe Ratio</p>
                          <p className="font-bold text-green-600">1.25</p>
                        </div>
                        <div>
                          <p className="text-xs text-gray-500">Alpha</p>
                          <p className="font-bold text-green-600">+2.15%</p>
                        </div>
                      </div>
                    </div>
                    
                    <div className="space-y-2">
                      <h5 className="font-medium">Performance Breakdown</h5>
                      <div className="space-y-2">
                        <div className="flex justify-between text-sm">
                          <span>Win Rate</span>
                          <span className="text-green-600">62.5%</span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span>Profit Factor</span>
                          <span className="text-green-600">1.67</span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span>Volatility</span>
                          <span className="text-yellow-600">35.5%</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Correlation Matrix */}
            <Card>
              <CardHeader>
                <CardTitle>Asset Correlation Matrix</CardTitle>
                <p className="text-sm text-gray-600">
                  Correlation analysis helps identify diversification opportunities and concentration risks.
                </p>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-5 gap-2 text-sm">
                  <div className="font-medium">Asset</div>
                  <div className="font-medium text-center">BTC</div>
                  <div className="font-medium text-center">ETH</div>
                  <div className="font-medium text-center">BNB</div>
                  <div className="font-medium text-center">SOL</div>
                  
                  <div className="font-medium">BTC</div>
                  <div className="text-center bg-blue-100 rounded p-1">1.00</div>
                  <div className="text-center bg-green-100 rounded p-1">0.75</div>
                  <div className="text-center bg-green-100 rounded p-1">0.68</div>
                  <div className="text-center bg-green-100 rounded p-1">0.72</div>
                  
                  <div className="font-medium">ETH</div>
                  <div className="text-center bg-green-100 rounded p-1">0.75</div>
                  <div className="text-center bg-blue-100 rounded p-1">1.00</div>
                  <div className="text-center bg-yellow-100 rounded p-1">0.82</div>
                  <div className="text-center bg-green-100 rounded p-1">0.78</div>
                  
                  <div className="font-medium">BNB</div>
                  <div className="text-center bg-green-100 rounded p-1">0.68</div>
                  <div className="text-center bg-yellow-100 rounded p-1">0.82</div>
                  <div className="text-center bg-blue-100 rounded p-1">1.00</div>
                  <div className="text-center bg-green-100 rounded p-1">0.69</div>
                  
                  <div className="font-medium">SOL</div>
                  <div className="text-center bg-green-100 rounded p-1">0.72</div>
                  <div className="text-center bg-green-100 rounded p-1">0.78</div>
                  <div className="text-center bg-green-100 rounded p-1">0.69</div>
                  <div className="text-center bg-blue-100 rounded p-1">1.00</div>
                </div>
                
                <div className="mt-4 flex items-center gap-4 text-xs">
                  <div className="flex items-center gap-1">
                    <div className="w-3 h-3 bg-blue-100 rounded"></div>
                    <span>Perfect (1.0)</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <div className="w-3 h-3 bg-yellow-100 rounded"></div>
                    <span>High (0.8+)</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <div className="w-3 h-3 bg-green-100 rounded"></div>
                    <span>Moderate (0.5-0.8)</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="quality" className="space-y-6">
            <DataQualityDashboard className="w-full" />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
};

export default PortfolioManager;
