import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { cn } from '@/lib/utils';
import {
    Activity,
    BarChart3,
    DollarSign,
    Globe,
    PieChart,
    Shield,
    TrendingUp,
    Zap
} from 'lucide-react';
import React, { useState } from 'react';
import ArbitrageOpportunities from './ArbitrageOpportunities';
import DataQualityDashboard from './DataQualityDashboard';
import MultiExchangePrices from './MultiExchangePrices';

interface EnhancedDashboardProps {
  className?: string;
}

const EnhancedDashboard: React.FC<EnhancedDashboardProps> = ({ className }) => {
  const [activeTab, setActiveTab] = useState('overview');

  const stats = [
    {
      title: 'Active Exchanges',
      value: '3',
      change: '+1',
      changeType: 'positive' as const,
      icon: Globe,
      description: 'Connected exchanges'
    },
    {
      title: 'News Articles',
      value: '247',
      change: '+23',
      changeType: 'positive' as const,
      icon: Globe,
      description: 'Latest crypto news'
    },
    {
      title: 'Sentiment Score',
      value: '72/100',
      change: '+8',
      changeType: 'positive' as const,
      icon: BarChart3,
      description: 'Market sentiment'
    },
    {
      title: 'Arbitrage Opportunities',
      value: '12',
      change: '+5',
      changeType: 'positive' as const,
      icon: TrendingUp,
      description: 'Profitable opportunities'
    },
    {
      title: 'Data Quality',
      value: '94%',
      change: '+2%',
      changeType: 'positive' as const,
      icon: Shield,
      description: 'Overall data quality'
    },
    {
      title: 'Total Volume',
      value: '$2.4B',
      change: '+12%',
      changeType: 'positive' as const,
      icon: DollarSign,
      description: '24h trading volume'
    }
  ];

  const getChangeColor = (type: 'positive' | 'negative' | 'neutral') => {
    switch (type) {
      case 'positive':
        return 'text-green-600 bg-green-50';
      case 'negative':
        return 'text-red-600 bg-red-50';
      default:
        return 'text-gray-600 bg-gray-50';
    }
  };

  return (
    <div className={cn('w-full space-y-6', className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Crypto Data Terminal</h1>
          <p className="text-gray-600 mt-1">
            Real-time multi-exchange crypto data aggregation with AI-powered news, sentiment analysis, and market intelligence
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge className="bg-green-100 text-green-800 border-green-200">
            <Zap className="h-3 w-3 mr-1" />
            Live Data
          </Badge>
          <Badge className="bg-blue-100 text-blue-800 border-blue-200">
            <Activity className="h-3 w-3 mr-1" />
            Multi-Exchange
          </Badge>
          <Badge className="bg-purple-100 text-purple-800 border-purple-200">
            <Globe className="h-3 w-3 mr-1" />
            AI Intelligence
          </Badge>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-6">
        {stats.map((stat, index) => (
          <Card key={index} className="hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">{stat.title}</p>
                  <p className="text-2xl font-bold text-gray-900 mt-1">{stat.value}</p>
                  <div className="flex items-center gap-1 mt-2">
                    <Badge className={cn('text-xs font-medium', getChangeColor(stat.changeType))}>
                      {stat.change}
                    </Badge>
                    <span className="text-xs text-gray-500">{stat.description}</span>
                  </div>
                </div>
                <div className="p-3 bg-blue-50 rounded-lg">
                  <stat.icon className="h-6 w-6 text-blue-600" />
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Main Content Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-7">
          <TabsTrigger value="overview" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="prices" className="flex items-center gap-2">
            <DollarSign className="h-4 w-4" />
            Prices
          </TabsTrigger>
          <TabsTrigger value="arbitrage" className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            Arbitrage
          </TabsTrigger>
          <TabsTrigger value="news" className="flex items-center gap-2">
            <Globe className="h-4 w-4" />
            News
          </TabsTrigger>
          <TabsTrigger value="sentiment" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Sentiment
          </TabsTrigger>
          <TabsTrigger value="analytics" className="flex items-center gap-2">
            <PieChart className="h-4 w-4" />
            Analytics
          </TabsTrigger>
          <TabsTrigger value="quality" className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            Quality
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6 mt-6">
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
            <MultiExchangePrices className="xl:col-span-1" />
            <ArbitrageOpportunities className="xl:col-span-1" />
          </div>
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
            <NewsWidget className="xl:col-span-2" />
            <SentimentWidget className="xl:col-span-1" />
          </div>
          <DataQualityDashboard />
        </TabsContent>

        <TabsContent value="prices" className="mt-6">
          <MultiExchangePrices />
        </TabsContent>

        <TabsContent value="arbitrage" className="mt-6">
          <ArbitrageOpportunities />
        </TabsContent>

        <TabsContent value="news" className="mt-6">
          <NewsWidget />
        </TabsContent>

        <TabsContent value="sentiment" className="mt-6">
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
            <SentimentWidget />
            <MarketIntelligenceWidget />
          </div>
        </TabsContent>

        <TabsContent value="analytics" className="mt-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BarChart3 className="h-5 w-5" />
                  Volume Analytics
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-center py-8 text-gray-500">
                  <BarChart3 className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p>Volume analytics coming soon</p>
                  <p className="text-sm">Real-time volume distribution across exchanges</p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <PieChart className="h-5 w-5" />
                  Spread Analytics
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-center py-8 text-gray-500">
                  <PieChart className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p>Spread analytics coming soon</p>
                  <p className="text-sm">Price spread analysis across exchanges</p>
                </div>
              </CardContent>
            </Card>

            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-5 w-5" />
                  Liquidity Analytics
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-center py-8 text-gray-500">
                  <Activity className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p>Liquidity analytics coming soon</p>
                  <p className="text-sm">Order book depth and liquidity analysis</p>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="quality" className="mt-6">
          <DataQualityDashboard />
        </TabsContent>
      </Tabs>

      {/* Footer */}
      <div className="border-t border-gray-200 pt-6">
        <div className="flex items-center justify-between text-sm text-gray-500">
          <div className="flex items-center gap-4">
            <span>© 2024 Crypto Terminal</span>
            <span>•</span>
            <span>Powered by Go Coffee</span>
          </div>
          <div className="flex items-center gap-4">
            <span>Real-time multi-exchange data</span>
            <span>•</span>
            <span>AI-powered news & sentiment</span>
            <span>•</span>
            <span>Advanced arbitrage detection</span>
            <span>•</span>
            <span>Market intelligence</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EnhancedDashboard;
