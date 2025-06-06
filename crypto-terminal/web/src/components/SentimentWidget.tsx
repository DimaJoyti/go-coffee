import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { 
  RefreshCw, 
  Heart, 
  TrendingUp, 
  TrendingDown, 
  MessageCircle,
  Users,
  Hash,
  Star,
  Twitter,
  MessageSquare
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface SentimentAnalysis {
  symbol: string;
  overall_sentiment: number;
  sentiment_score: number;
  confidence: number;
  total_mentions: number;
  positive_mentions: number;
  negative_mentions: number;
  neutral_mentions: number;
  platform_breakdown: Record<string, PlatformSentiment>;
  trending_topics: string[];
  influencer_posts: SocialPost[];
  time_range: string;
  last_updated: string;
}

interface PlatformSentiment {
  platform: string;
  sentiment: number;
  mentions: number;
  positive_mentions: number;
  negative_mentions: number;
  neutral_mentions: number;
  top_posts: SocialPost[];
}

interface SocialPost {
  id: string;
  platform: string;
  content: string;
  author: string;
  posted_at: string;
  sentiment: number;
  engagement: number;
  reach: number;
  symbols: string[];
  hashtags: string[];
  is_influencer: boolean;
}

interface SentimentWidgetProps {
  className?: string;
  symbol?: string;
}

const SentimentWidget: React.FC<SentimentWidgetProps> = ({ 
  className, 
  symbol = 'BTC' 
}) => {
  const [sentiment, setSentiment] = useState<SentimentAnalysis | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSymbol, setSelectedSymbol] = useState(symbol);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const availableSymbols = ['BTC', 'ETH', 'BNB', 'ADA', 'SOL', 'XRP', 'DOT', 'DOGE'];

  const fetchSentiment = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v2/intelligence/sentiment/${selectedSymbol}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch sentiment data');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setSentiment(data.data);
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
    fetchSentiment();
    
    // Auto-refresh every 2 minutes
    const interval = setInterval(fetchSentiment, 2 * 60 * 1000);
    
    return () => clearInterval(interval);
  }, [selectedSymbol]);

  const getSentimentColor = (sentimentValue: number) => {
    if (sentimentValue > 0.2) return 'text-green-600';
    if (sentimentValue < -0.2) return 'text-red-600';
    return 'text-gray-600';
  };

  const getSentimentBadgeColor = (score: number) => {
    if (score >= 70) return 'bg-green-100 text-green-800 border-green-200';
    if (score >= 40) return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    return 'bg-red-100 text-red-800 border-red-200';
  };

  const getSentimentIcon = (sentimentValue: number) => {
    if (sentimentValue > 0.2) return <TrendingUp className="h-4 w-4" />;
    if (sentimentValue < -0.2) return <TrendingDown className="h-4 w-4" />;
    return <MessageCircle className="h-4 w-4" />;
  };

  const getPlatformIcon = (platform: string) => {
    switch (platform.toLowerCase()) {
      case 'twitter':
      case 'x':
        return <Twitter className="h-4 w-4" />;
      case 'reddit':
        return <MessageSquare className="h-4 w-4" />;
      default:
        return <MessageCircle className="h-4 w-4" />;
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

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-xl font-bold flex items-center gap-2">
          <Heart className="h-5 w-5 text-pink-600" />
          Social Sentiment
        </CardTitle>
        <div className="flex items-center gap-3">
          <Select value={selectedSymbol} onValueChange={setSelectedSymbol}>
            <SelectTrigger className="w-24">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {availableSymbols.map((sym) => (
                <SelectItem key={sym} value={sym}>
                  {sym}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          
          {lastUpdate && (
            <span className="text-sm text-gray-500">
              {lastUpdate.toLocaleTimeString()}
            </span>
          )}
          
          <Button
            variant="outline"
            size="sm"
            onClick={fetchSentiment}
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
            <span className="text-red-700">{error}</span>
          </div>
        )}
        
        {loading && !sentiment ? (
          <div className="flex items-center justify-center py-8">
            <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
            <span className="ml-2 text-gray-500">Loading sentiment data...</span>
          </div>
        ) : sentiment ? (
          <div className="space-y-6">
            {/* Overall Sentiment */}
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-semibold text-lg flex items-center gap-2">
                  {getSentimentIcon(sentiment.overall_sentiment)}
                  Overall Sentiment
                </h3>
                <Badge className={cn('font-medium', getSentimentBadgeColor(sentiment.sentiment_score))}>
                  {sentiment.sentiment_score}/100
                </Badge>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
                <div className="text-center">
                  <p className="text-sm text-gray-600">Total Mentions</p>
                  <p className="text-lg font-bold">{formatNumber(sentiment.total_mentions)}</p>
                </div>
                <div className="text-center">
                  <p className="text-sm text-gray-600">Positive</p>
                  <p className="text-lg font-bold text-green-600">{formatNumber(sentiment.positive_mentions)}</p>
                </div>
                <div className="text-center">
                  <p className="text-sm text-gray-600">Negative</p>
                  <p className="text-lg font-bold text-red-600">{formatNumber(sentiment.negative_mentions)}</p>
                </div>
                <div className="text-center">
                  <p className="text-sm text-gray-600">Confidence</p>
                  <p className="text-lg font-bold">{(sentiment.confidence * 100).toFixed(0)}%</p>
                </div>
              </div>
              
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Positive</span>
                  <span>{((sentiment.positive_mentions / sentiment.total_mentions) * 100).toFixed(1)}%</span>
                </div>
                <Progress 
                  value={(sentiment.positive_mentions / sentiment.total_mentions) * 100} 
                  className="h-2 bg-gray-200"
                />
                
                <div className="flex justify-between text-sm">
                  <span>Negative</span>
                  <span>{((sentiment.negative_mentions / sentiment.total_mentions) * 100).toFixed(1)}%</span>
                </div>
                <Progress 
                  value={(sentiment.negative_mentions / sentiment.total_mentions) * 100} 
                  className="h-2 bg-gray-200"
                />
              </div>
            </div>
            
            {/* Platform Breakdown */}
            <div className="space-y-3">
              <h3 className="font-semibold text-lg">Platform Breakdown</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {Object.entries(sentiment.platform_breakdown).map(([platform, data]) => (
                  <div key={platform} className="border border-gray-200 rounded-lg p-3">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        {getPlatformIcon(platform)}
                        <span className="font-medium capitalize">{platform}</span>
                      </div>
                      <Badge 
                        className={cn('text-xs', getSentimentColor(data.sentiment))}
                        variant="outline"
                      >
                        {data.sentiment > 0 ? '+' : ''}{(data.sentiment * 100).toFixed(0)}%
                      </Badge>
                    </div>
                    
                    <div className="grid grid-cols-3 gap-2 text-xs text-gray-600">
                      <div className="text-center">
                        <p className="font-medium">{formatNumber(data.mentions)}</p>
                        <p>Total</p>
                      </div>
                      <div className="text-center">
                        <p className="font-medium text-green-600">{formatNumber(data.positive_mentions)}</p>
                        <p>Positive</p>
                      </div>
                      <div className="text-center">
                        <p className="font-medium text-red-600">{formatNumber(data.negative_mentions)}</p>
                        <p>Negative</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            
            {/* Trending Topics */}
            {sentiment.trending_topics.length > 0 && (
              <div className="space-y-3">
                <h3 className="font-semibold text-lg flex items-center gap-2">
                  <Hash className="h-4 w-4" />
                  Trending Topics
                </h3>
                <div className="flex flex-wrap gap-2">
                  {sentiment.trending_topics.map((topic) => (
                    <Badge key={topic} variant="outline" className="text-sm">
                      #{topic}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
            
            {/* Influencer Posts */}
            {sentiment.influencer_posts.length > 0 && (
              <div className="space-y-3">
                <h3 className="font-semibold text-lg flex items-center gap-2">
                  <Star className="h-4 w-4" />
                  Influencer Posts
                </h3>
                <div className="space-y-3">
                  {sentiment.influencer_posts.slice(0, 3).map((post) => (
                    <div key={post.id} className="border border-gray-200 rounded-lg p-3">
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          {getPlatformIcon(post.platform)}
                          <span className="font-medium">@{post.author}</span>
                          <Badge variant="outline" className="text-xs">
                            Influencer
                          </Badge>
                        </div>
                        <div className="flex items-center gap-2 text-xs text-gray-500">
                          <Users className="h-3 w-3" />
                          <span>{formatNumber(post.reach)}</span>
                          <MessageCircle className="h-3 w-3" />
                          <span>{formatNumber(post.engagement)}</span>
                        </div>
                      </div>
                      
                      <p className="text-sm text-gray-700 mb-2 line-clamp-2">
                        {post.content}
                      </p>
                      
                      <div className="flex items-center justify-between text-xs text-gray-500">
                        <span>{formatTimeAgo(post.posted_at)}</span>
                        <div className={cn('flex items-center gap-1', getSentimentColor(post.sentiment))}>
                          {getSentimentIcon(post.sentiment)}
                          <span>
                            {post.sentiment > 0 ? 'Positive' : post.sentiment < 0 ? 'Negative' : 'Neutral'}
                          </span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <Heart className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p>No sentiment data available</p>
            <p className="text-sm">Try selecting a different symbol</p>
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default SentimentWidget;
