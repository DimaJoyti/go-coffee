import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  RefreshCw, 
  Brain, 
  TrendingUp, 
  TrendingDown, 
  AlertTriangle,
  CheckCircle,
  Clock,
  ExternalLink,
  Target,
  Zap,
  Calendar,
  Users
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface MarketInsight {
  id: string;
  type: string;
  category: string;
  title: string;
  description: string;
  impact: string;
  confidence: number;
  symbols: string[];
  source: string;
  url?: string;
  data: Record<string, any>;
  created_at: string;
  expires_at?: string;
}

interface TrendingTopic {
  topic: string;
  mentions: number;
  sentiment: number;
  growth: number;
  symbols: string[];
  platforms: string[];
  last_updated: string;
}

interface MarketIntelligenceWidgetProps {
  className?: string;
}

const MarketIntelligenceWidget: React.FC<MarketIntelligenceWidgetProps> = ({ className }) => {
  const [insights, setInsights] = useState<MarketInsight[]>([]);
  const [trendingTopics, setTrendingTopics] = useState<TrendingTopic[]>([]);
  const [marketSentiment, setMarketSentiment] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);
  const [activeTab, setActiveTab] = useState('insights');

  const fetchData = async () => {
    try {
      setLoading(true);
      
      // Fetch market insights
      const insightsResponse = await fetch('/api/v2/intelligence/insights?limit=10');
      const insightsData = await insightsResponse.json();
      
      // Fetch trending topics
      const topicsResponse = await fetch('/api/v2/intelligence/sentiment/trending?limit=8');
      const topicsData = await topicsResponse.json();
      
      // Fetch market sentiment
      const sentimentResponse = await fetch('/api/v2/intelligence/insights/market-sentiment');
      const sentimentData = await sentimentResponse.json();
      
      if (insightsData.success) {
        setInsights(insightsData.data || []);
      }
      
      if (topicsData.success) {
        setTrendingTopics(topicsData.data || []);
      }
      
      if (sentimentData.success) {
        setMarketSentiment(sentimentData.data);
      }
      
      setLastUpdate(new Date());
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    
    // Auto-refresh every 10 minutes
    const interval = setInterval(fetchData, 10 * 60 * 1000);
    
    return () => clearInterval(interval);
  }, []);

  const getImpactColor = (impact: string) => {
    switch (impact.toLowerCase()) {
      case 'high':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'low':
        return 'bg-green-100 text-green-800 border-green-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getCategoryColor = (category: string) => {
    switch (category.toLowerCase()) {
      case 'bullish':
        return 'text-green-600';
      case 'bearish':
        return 'text-red-600';
      case 'neutral':
        return 'text-gray-600';
      default:
        return 'text-gray-600';
    }
  };

  const getCategoryIcon = (category: string) => {
    switch (category.toLowerCase()) {
      case 'bullish':
        return <TrendingUp className="h-4 w-4" />;
      case 'bearish':
        return <TrendingDown className="h-4 w-4" />;
      default:
        return <Target className="h-4 w-4" />;
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case 'news':
        return <Zap className="h-4 w-4" />;
      case 'technical':
        return <Target className="h-4 w-4" />;
      case 'fundamental':
        return <Calendar className="h-4 w-4" />;
      case 'social':
        return <Users className="h-4 w-4" />;
      default:
        return <Brain className="h-4 w-4" />;
    }
  };

  const formatTimeAgo = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));
    
    if (diffInMinutes < 60) {
      return `${diffInMinutes}m ago`;
    } else if (diffInMinutes < 1440) {
      return `${Math.floor(diffInMinutes / 60)}h ago`;
    } else {
      return `${Math.floor(diffInMinutes / 1440)}d ago`;
    }
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) {
      return `${(num / 1000000).toFixed(1)}M`;
    }
    if (num >= 1000) {
      return `${(num / 1000).toFixed(1)}K`;
    }
    return num.toString();
  };

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-xl font-bold flex items-center gap-2">
          <Brain className="h-5 w-5 text-purple-600" />
          Market Intelligence
        </CardTitle>
        <div className="flex items-center gap-2">
          {lastUpdate && (
            <span className="text-sm text-gray-500">
              {lastUpdate.toLocaleTimeString()}
            </span>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={fetchData}
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
        
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="insights">Insights</TabsTrigger>
            <TabsTrigger value="trending">Trending</TabsTrigger>
            <TabsTrigger value="sentiment">Sentiment</TabsTrigger>
          </TabsList>

          <TabsContent value="insights" className="mt-4">
            {loading && insights.length === 0 ? (
              <div className="flex items-center justify-center py-8">
                <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
                <span className="ml-2 text-gray-500">Loading insights...</span>
              </div>
            ) : insights.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Brain className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p>No market insights available</p>
              </div>
            ) : (
              <div className="space-y-4">
                {insights.map((insight) => (
                  <div
                    key={insight.id}
                    className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <div className={cn('flex items-center gap-1', getCategoryColor(insight.category))}>
                          {getCategoryIcon(insight.category)}
                        </div>
                        <Badge className={cn('text-xs', getImpactColor(insight.impact))}>
                          {insight.impact} impact
                        </Badge>
                        <Badge variant="outline" className="text-xs">
                          {insight.type}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-gray-500">
                          {(insight.confidence * 100).toFixed(0)}% confidence
                        </span>
                        {insight.url && (
                          <a
                            href={insight.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-gray-400 hover:text-gray-600"
                          >
                            <ExternalLink className="h-4 w-4" />
                          </a>
                        )}
                      </div>
                    </div>
                    
                    <h3 className="font-semibold text-lg mb-2">{insight.title}</h3>
                    <p className="text-gray-600 text-sm mb-3">{insight.description}</p>
                    
                    <div className="flex items-center justify-between text-xs text-gray-500">
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-1">
                          {getTypeIcon(insight.type)}
                          <span>{insight.source}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          <span>{formatTimeAgo(insight.created_at)}</span>
                        </div>
                      </div>
                      
                      {insight.symbols.length > 0 && (
                        <div className="flex items-center gap-1">
                          {insight.symbols.slice(0, 3).map((symbol) => (
                            <Badge key={symbol} variant="outline" className="text-xs">
                              {symbol}
                            </Badge>
                          ))}
                          {insight.symbols.length > 3 && (
                            <span className="text-gray-400">+{insight.symbols.length - 3}</span>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          <TabsContent value="trending" className="mt-4">
            {loading && trendingTopics.length === 0 ? (
              <div className="flex items-center justify-center py-8">
                <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
                <span className="ml-2 text-gray-500">Loading trending topics...</span>
              </div>
            ) : trendingTopics.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <TrendingUp className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p>No trending topics available</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {trendingTopics.map((topic, index) => (
                  <div
                    key={topic.topic}
                    className="border border-gray-200 rounded-lg p-4"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <h3 className="font-semibold">#{topic.topic}</h3>
                      <Badge className={cn(
                        'text-xs',
                        topic.growth > 0 ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                      )}>
                        {topic.growth > 0 ? '+' : ''}{topic.growth.toFixed(0)}%
                      </Badge>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-2 text-sm">
                      <div>
                        <p className="text-gray-600">Mentions</p>
                        <p className="font-medium">{formatNumber(topic.mentions)}</p>
                      </div>
                      <div>
                        <p className="text-gray-600">Sentiment</p>
                        <p className={cn(
                          'font-medium',
                          topic.sentiment > 0 ? 'text-green-600' : 
                          topic.sentiment < 0 ? 'text-red-600' : 'text-gray-600'
                        )}>
                          {topic.sentiment > 0 ? '+' : ''}{(topic.sentiment * 100).toFixed(0)}%
                        </p>
                      </div>
                    </div>
                    
                    {topic.symbols.length > 0 && (
                      <div className="mt-3">
                        <p className="text-xs text-gray-600 mb-1">Related symbols:</p>
                        <div className="flex flex-wrap gap-1">
                          {topic.symbols.slice(0, 4).map((symbol) => (
                            <Badge key={symbol} variant="outline" className="text-xs">
                              {symbol}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          <TabsContent value="sentiment" className="mt-4">
            {loading && !marketSentiment ? (
              <div className="flex items-center justify-center py-8">
                <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
                <span className="ml-2 text-gray-500">Loading market sentiment...</span>
              </div>
            ) : marketSentiment ? (
              <div className="space-y-6">
                <div className="p-4 bg-gray-50 rounded-lg">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-semibold text-lg">Overall Market Sentiment</h3>
                    <Badge className={cn(
                      'font-medium',
                      marketSentiment.sentiment_score >= 60 ? 'bg-green-100 text-green-800' :
                      marketSentiment.sentiment_score >= 40 ? 'bg-yellow-100 text-yellow-800' :
                      'bg-red-100 text-red-800'
                    )}>
                      {marketSentiment.sentiment_score}/100
                    </Badge>
                  </div>
                  
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
                    <div>
                      <p className="text-sm text-gray-600">Trend</p>
                      <p className="font-medium capitalize">{marketSentiment.trend}</p>
                    </div>
                    <div>
                      <p className="text-sm text-gray-600">Confidence</p>
                      <p className="font-medium">{(marketSentiment.confidence * 100).toFixed(0)}%</p>
                    </div>
                    <div>
                      <p className="text-sm text-gray-600">News</p>
                      <p className={cn(
                        'font-medium',
                        marketSentiment.sentiment_breakdown.news > 0 ? 'text-green-600' : 'text-red-600'
                      )}>
                        {(marketSentiment.sentiment_breakdown.news * 100).toFixed(0)}%
                      </p>
                    </div>
                    <div>
                      <p className="text-sm text-gray-600">Social</p>
                      <p className={cn(
                        'font-medium',
                        marketSentiment.sentiment_breakdown.social > 0 ? 'text-green-600' : 'text-red-600'
                      )}>
                        {(marketSentiment.sentiment_breakdown.social * 100).toFixed(0)}%
                      </p>
                    </div>
                  </div>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <h4 className="font-medium mb-2 flex items-center gap-2">
                      <CheckCircle className="h-4 w-4 text-green-600" />
                      Key Drivers
                    </h4>
                    <ul className="space-y-1">
                      {marketSentiment.key_drivers.map((driver: string, index: number) => (
                        <li key={index} className="text-sm text-gray-600 capitalize">
                          • {driver.replace(/_/g, ' ')}
                        </li>
                      ))}
                    </ul>
                  </div>
                  
                  <div>
                    <h4 className="font-medium mb-2 flex items-center gap-2">
                      <AlertTriangle className="h-4 w-4 text-yellow-600" />
                      Risk Factors
                    </h4>
                    <ul className="space-y-1">
                      {marketSentiment.risk_factors.map((risk: string, index: number) => (
                        <li key={index} className="text-sm text-gray-600 capitalize">
                          • {risk.replace(/_/g, ' ')}
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                <Brain className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p>No market sentiment data available</p>
              </div>
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
};

export default MarketIntelligenceWidget;
