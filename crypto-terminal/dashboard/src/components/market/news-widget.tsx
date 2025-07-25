'use client'

import * as React from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn, formatTimeAgo } from "@/lib/utils"
import { 
  Newspaper, 
  ExternalLink, 
  TrendingUp, 
  TrendingDown,
  Clock,
  RefreshCw
} from "lucide-react"

interface NewsItem {
  id: string
  title: string
  summary: string
  source: string
  publishedAt: Date
  url: string
  sentiment: "bullish" | "bearish" | "neutral"
  tags: string[]
  imageUrl?: string
}

interface NewsWidgetProps {
  className?: string
  maxItems?: number
  onNewsClick?: (newsItem: NewsItem) => void
}

// Mock news data
const mockNews: NewsItem[] = [
  {
    id: "1",
    title: "Bitcoin ETF Sees Record Inflows as Institutional Adoption Grows",
    summary: "Major institutional investors continue to pour money into Bitcoin ETFs, with record-breaking inflows reported this week.",
    source: "CryptoNews",
    publishedAt: new Date(Date.now() - 30 * 60 * 1000), // 30 minutes ago
    url: "#",
    sentiment: "bullish",
    tags: ["BTC", "ETF", "Institutional"],
  },
  {
    id: "2",
    title: "Ethereum Network Upgrade Scheduled for Next Month",
    summary: "The upcoming Ethereum upgrade promises to improve scalability and reduce transaction fees significantly.",
    source: "BlockchainDaily",
    publishedAt: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
    url: "#",
    sentiment: "bullish",
    tags: ["ETH", "Upgrade", "Scalability"],
  },
  {
    id: "3",
    title: "Regulatory Concerns Impact Crypto Market Sentiment",
    summary: "New regulatory proposals in major markets are causing uncertainty among cryptocurrency investors.",
    source: "FinanceToday",
    publishedAt: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    url: "#",
    sentiment: "bearish",
    tags: ["Regulation", "Market", "Policy"],
  },
  {
    id: "4",
    title: "DeFi Protocol Launches Revolutionary Yield Farming Feature",
    summary: "A new DeFi protocol introduces innovative yield farming mechanisms that could reshape the landscape.",
    source: "DeFiInsider",
    publishedAt: new Date(Date.now() - 6 * 60 * 60 * 1000), // 6 hours ago
    url: "#",
    sentiment: "bullish",
    tags: ["DeFi", "Yield", "Innovation"],
  },
  {
    id: "5",
    title: "Major Exchange Reports Security Breach",
    summary: "A prominent cryptocurrency exchange has reported a security incident affecting user accounts.",
    source: "SecurityWatch",
    publishedAt: new Date(Date.now() - 8 * 60 * 60 * 1000), // 8 hours ago
    url: "#",
    sentiment: "bearish",
    tags: ["Security", "Exchange", "Breach"],
  },
]

const NewsWidget: React.FC<NewsWidgetProps> = ({
  className,
  maxItems = 5,
  onNewsClick,
}) => {
  const [news] = React.useState<NewsItem[]>(mockNews)
  const [isRefreshing, setIsRefreshing] = React.useState(false)
  const [selectedSentiment, setSelectedSentiment] = React.useState<"all" | "bullish" | "bearish" | "neutral">("all")

  const filteredNews = React.useMemo(() => {
    let filtered = news
    if (selectedSentiment !== "all") {
      filtered = news.filter(item => item.sentiment === selectedSentiment)
    }
    return filtered.slice(0, maxItems)
  }, [news, selectedSentiment, maxItems])

  const handleRefresh = async () => {
    setIsRefreshing(true)
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsRefreshing(false)
  }

  const getSentimentColor = (sentiment: string) => {
    switch (sentiment) {
      case "bullish":
        return "text-bull bg-bull/10 border-bull/20"
      case "bearish":
        return "text-bear bg-bear/10 border-bear/20"
      default:
        return "text-neutral bg-neutral/10 border-neutral/20"
    }
  }

  const getSentimentIcon = (sentiment: string) => {
    switch (sentiment) {
      case "bullish":
        return <TrendingUp className="h-3 w-3" />
      case "bearish":
        return <TrendingDown className="h-3 w-3" />
      default:
        return null
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.5, delay: 0.3 }}
    >
      <Card variant="glass" className={cn("w-full", className)}>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Newspaper className="h-5 w-5 text-primary" />
              <span>Crypto News</span>
              <Badge variant="glass">{filteredNews.length}</Badge>
            </CardTitle>
            
            <div className="flex items-center space-x-2">
              <div className="flex items-center space-x-1">
                {["all", "bullish", "bearish", "neutral"].map((sentiment) => (
                  <Button
                    key={sentiment}
                    variant={selectedSentiment === sentiment ? "epic" : "glass"}
                    size="sm"
                    className="h-6 text-xs capitalize"
                    onClick={() => setSelectedSentiment(sentiment as any)}
                  >
                    {sentiment}
                  </Button>
                ))}
              </div>
              
              <Button
                variant="glass"
                size="sm"
                onClick={handleRefresh}
                disabled={isRefreshing}
              >
                <RefreshCw className={cn("h-4 w-4", isRefreshing && "animate-spin")} />
              </Button>
            </div>
          </div>
        </CardHeader>
        
        <CardContent className="space-y-3">
          <div className="space-y-3 max-h-96 overflow-y-auto scrollbar-thin">
            {filteredNews.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <Newspaper className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">No news available</p>
                <p className="text-xs">Try refreshing or changing filters</p>
              </div>
            ) : (
              filteredNews.map((item, index) => (
                <motion.div
                  key={item.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="p-3 rounded-lg border glass hover:bg-muted/20 transition-all cursor-pointer group"
                  onClick={() => onNewsClick?.(item)}
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <Badge 
                        variant="glass" 
                        className={cn("text-xs", getSentimentColor(item.sentiment))}
                      >
                        {getSentimentIcon(item.sentiment)}
                        <span className="ml-1 capitalize">{item.sentiment}</span>
                      </Badge>
                      <Badge variant="outline" className="text-xs">
                        {item.source}
                      </Badge>
                    </div>
                    
                    <ExternalLink className="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                  </div>
                  
                  <h4 className="text-sm font-medium mb-1 line-clamp-2 group-hover:text-primary transition-colors">
                    {item.title}
                  </h4>
                  
                  <p className="text-xs text-muted-foreground mb-2 line-clamp-2">
                    {item.summary}
                  </p>
                  
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-1 flex-wrap gap-1">
                      {item.tags.slice(0, 3).map((tag) => (
                        <Badge key={tag} variant="outline" className="text-2xs px-1 py-0">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                    
                    <div className="flex items-center space-x-1 text-2xs text-muted-foreground">
                      <Clock className="h-3 w-3" />
                      <span>{formatTimeAgo(item.publishedAt)}</span>
                    </div>
                  </div>
                </motion.div>
              ))
            )}
          </div>
          
          {/* Summary */}
          <div className="pt-2 border-t glass">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <div className="flex items-center space-x-4">
                <span className="flex items-center space-x-1">
                  <div className="w-2 h-2 bg-bull rounded-full"></div>
                  <span>{news.filter(n => n.sentiment === "bullish").length} Bullish</span>
                </span>
                <span className="flex items-center space-x-1">
                  <div className="w-2 h-2 bg-bear rounded-full"></div>
                  <span>{news.filter(n => n.sentiment === "bearish").length} Bearish</span>
                </span>
              </div>
              <span>Last updated: {new Date().toLocaleTimeString()}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}

export { NewsWidget }
