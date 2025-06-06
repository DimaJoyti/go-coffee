import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { 
  RefreshCw, 
  Newspaper, 
  ExternalLink, 
  Search, 
  TrendingUp, 
  TrendingDown,
  Clock,
  User,
  Tag
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface NewsArticle {
  id: string;
  title: string;
  summary: string;
  url: string;
  source: string;
  author: string;
  published_at: string;
  sentiment: number;
  relevance: number;
  symbols: string[];
  tags: string[];
  image_url?: string;
}

interface NewsWidgetProps {
  className?: string;
  symbols?: string[];
  limit?: number;
}

const NewsWidget: React.FC<NewsWidgetProps> = ({ 
  className, 
  symbols = [], 
  limit = 10 
}) => {
  const [news, setNews] = useState<NewsArticle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSymbol, setSelectedSymbol] = useState<string>('all');
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const availableSymbols = ['all', 'BTC', 'ETH', 'BNB', 'ADA', 'SOL', 'XRP', 'DOT'];

  const fetchNews = async () => {
    try {
      setLoading(true);
      
      let url = `/api/v2/intelligence/news?limit=${limit}`;
      
      if (selectedSymbol !== 'all') {
        url += `&symbols=${selectedSymbol}`;
      } else if (symbols.length > 0) {
        url += `&symbols=${symbols.join(',')}`;
      }
      
      const response = await fetch(url);
      
      if (!response.ok) {
        throw new Error('Failed to fetch news');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setNews(data.data || []);
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

  const searchNews = async () => {
    if (!searchQuery.trim()) {
      fetchNews();
      return;
    }
    
    try {
      setLoading(true);
      const response = await fetch(`/api/v2/intelligence/news/search?q=${encodeURIComponent(searchQuery)}`);
      
      if (!response.ok) {
        throw new Error('Failed to search news');
      }
      
      const data = await response.json();
      
      if (data.success) {
        // Convert search results to news articles format
        const searchResults = data.data.map((result: any) => ({
          id: result.url,
          title: result.title,
          summary: result.description,
          url: result.url,
          source: result.source,
          author: 'Unknown',
          published_at: result.timestamp,
          sentiment: 0,
          relevance: result.relevance,
          symbols: [],
          tags: [],
        }));
        setNews(searchResults);
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
    fetchNews();
    
    // Auto-refresh every 5 minutes
    const interval = setInterval(fetchNews, 5 * 60 * 1000);
    
    return () => clearInterval(interval);
  }, [selectedSymbol, limit]);

  const getSentimentColor = (sentiment: number) => {
    if (sentiment > 0.1) return 'text-green-600';
    if (sentiment < -0.1) return 'text-red-600';
    return 'text-gray-600';
  };

  const getSentimentIcon = (sentiment: number) => {
    if (sentiment > 0.1) return <TrendingUp className="h-3 w-3" />;
    if (sentiment < -0.1) return <TrendingDown className="h-3 w-3" />;
    return null;
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

  const getRelevanceColor = (relevance: number) => {
    if (relevance >= 0.8) return 'bg-green-100 text-green-800';
    if (relevance >= 0.6) return 'bg-yellow-100 text-yellow-800';
    return 'bg-gray-100 text-gray-800';
  };

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-xl font-bold flex items-center gap-2">
          <Newspaper className="h-5 w-5 text-blue-600" />
          Crypto News
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
            onClick={fetchNews}
            disabled={loading}
            className="flex items-center gap-1"
          >
            <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
            Refresh
          </Button>
        </div>
      </CardHeader>
      
      <CardContent>
        {/* Search and Filter Controls */}
        <div className="flex flex-col sm:flex-row gap-3 mb-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              placeholder="Search crypto news..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && searchNews()}
              className="pl-10"
            />
          </div>
          <Select value={selectedSymbol} onValueChange={setSelectedSymbol}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {availableSymbols.map((symbol) => (
                <SelectItem key={symbol} value={symbol}>
                  {symbol === 'all' ? 'All' : symbol}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button onClick={searchNews} disabled={loading}>
            Search
          </Button>
        </div>

        {error && (
          <div className="flex items-center gap-2 p-4 bg-red-50 border border-red-200 rounded-lg mb-4">
            <span className="text-red-700">{error}</span>
          </div>
        )}
        
        {loading && news.length === 0 ? (
          <div className="flex items-center justify-center py-8">
            <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
            <span className="ml-2 text-gray-500">Loading news...</span>
          </div>
        ) : news.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <Newspaper className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p>No news articles found</p>
            <p className="text-sm">Try adjusting your search or filters</p>
          </div>
        ) : (
          <div className="space-y-4">
            {news.map((article) => (
              <div
                key={article.id}
                className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-semibold text-lg leading-tight pr-4">
                    <a
                      href={article.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="hover:text-blue-600 transition-colors"
                    >
                      {article.title}
                    </a>
                  </h3>
                  <div className="flex items-center gap-1 flex-shrink-0">
                    {article.relevance > 0 && (
                      <Badge className={cn('text-xs', getRelevanceColor(article.relevance))}>
                        {(article.relevance * 100).toFixed(0)}%
                      </Badge>
                    )}
                    <a
                      href={article.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-gray-400 hover:text-gray-600"
                    >
                      <ExternalLink className="h-4 w-4" />
                    </a>
                  </div>
                </div>
                
                <p className="text-gray-600 text-sm mb-3 line-clamp-2">
                  {article.summary}
                </p>
                
                <div className="flex items-center justify-between text-xs text-gray-500">
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-1">
                      <User className="h-3 w-3" />
                      <span>{article.source}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      <span>{formatTimeAgo(article.published_at)}</span>
                    </div>
                    {article.sentiment !== 0 && (
                      <div className={cn('flex items-center gap-1', getSentimentColor(article.sentiment))}>
                        {getSentimentIcon(article.sentiment)}
                        <span>
                          {article.sentiment > 0 ? 'Positive' : 'Negative'}
                        </span>
                      </div>
                    )}
                  </div>
                  
                  <div className="flex items-center gap-2">
                    {article.symbols.length > 0 && (
                      <div className="flex items-center gap-1">
                        {article.symbols.slice(0, 3).map((symbol) => (
                          <Badge key={symbol} variant="outline" className="text-xs">
                            {symbol}
                          </Badge>
                        ))}
                        {article.symbols.length > 3 && (
                          <span className="text-gray-400">+{article.symbols.length - 3}</span>
                        )}
                      </div>
                    )}
                    
                    {article.tags.length > 0 && (
                      <div className="flex items-center gap-1">
                        <Tag className="h-3 w-3" />
                        <span>{article.tags[0]}</span>
                        {article.tags.length > 1 && (
                          <span className="text-gray-400">+{article.tags.length - 1}</span>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default NewsWidget;
