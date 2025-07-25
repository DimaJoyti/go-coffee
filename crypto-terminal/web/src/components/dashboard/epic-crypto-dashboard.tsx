import * as React from "react"
import { DashboardLayout } from "../layout/dashboard-layout"
import { TradingCard } from "../ui/trading-card"
import { MarketDataTable, MarketDataItem } from "../ui/market-data-table"
import { PriceChart, PriceDataPoint } from "../ui/price-chart"
import { MarketHeatmap, HeatmapData } from "../market/market-heatmap"
import { OrderBook, OrderBookData } from "../trading/order-book"
import { TradingViewChart } from "../trading/tradingview-chart"
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs"
import { Badge } from "../ui/badge"
import { Button } from "../ui/button"
import { cn } from "../../lib/utils"
import { 
  Bitcoin, 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  BarChart3,
  Activity,
  Zap,
  Target,
  Shield
} from "lucide-react"

// Mock data - in a real app, this would come from APIs
const mockMarketData: MarketDataItem[] = [
  {
    id: "bitcoin",
    symbol: "BTC",
    name: "Bitcoin",
    price: 43250,
    change24h: 1250,
    changePercent24h: 2.98,
    volume24h: 28500000000,
    marketCap: 847000000000,
    rank: 1,
    logo: "/crypto-logos/btc.png"
  },
  {
    id: "ethereum",
    symbol: "ETH",
    name: "Ethereum",
    price: 2650,
    change24h: -45,
    changePercent24h: -1.67,
    volume24h: 15200000000,
    marketCap: 318000000000,
    rank: 2,
    logo: "/crypto-logos/eth.png"
  },
  // Add more mock data...
]

const mockHeatmapData: HeatmapData[] = mockMarketData.map(item => ({
  symbol: item.symbol,
  name: item.name,
  price: item.price,
  change24h: item.change24h,
  changePercent24h: item.changePercent24h,
  marketCap: item.marketCap,
  volume24h: item.volume24h,
  category: "Cryptocurrency"
}))

const mockPriceData: PriceDataPoint[] = Array.from({ length: 30 }, (_, i) => ({
  timestamp: Date.now() - (29 - i) * 24 * 60 * 60 * 1000,
  price: 43000 + Math.random() * 2000 - 1000,
  volume: Math.random() * 1000000000
}))

const mockOrderBookData: OrderBookData = {
  bids: Array.from({ length: 20 }, (_, i) => ({
    price: 43200 - i * 10,
    size: Math.random() * 5,
    total: Math.random() * 100
  })),
  asks: Array.from({ length: 20 }, (_, i) => ({
    price: 43250 + i * 10,
    size: Math.random() * 5,
    total: Math.random() * 100
  })),
  spread: 50,
  spreadPercent: 0.12,
  lastPrice: 43225,
  lastUpdate: Date.now()
}

const EpicCryptoDashboard: React.FC = () => {
  const [selectedSymbol, setSelectedSymbol] = React.useState("BTCUSD")
  const [loading, setLoading] = React.useState(false)

  // Simulate real-time updates
  React.useEffect(() => {
    const interval = setInterval(() => {
      // In a real app, you'd fetch new data here
      console.log("Updating market data...")
    }, 5000)

    return () => clearInterval(interval)
  }, [])

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Top Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <TradingCard
            title="Portfolio Value"
            value={125750.50}
            change={2850.25}
            changeType="currency"
            currency="USD"
            trend="up"
            icon={<DollarSign className="h-4 w-4" />}
            subtitle="24h change"
          />
          <TradingCard
            title="BTC Price"
            value={43250}
            change={1250}
            changeType="currency"
            currency="USD"
            trend="up"
            icon={<Bitcoin className="h-4 w-4" />}
            subtitle="Last updated: 2s ago"
          />
          <TradingCard
            title="Active Positions"
            value="12"
            change={2}
            changeType="number"
            trend="up"
            icon={<Target className="h-4 w-4" />}
            subtitle="2 new today"
          />
          <TradingCard
            title="P&L Today"
            value={3250.75}
            change={15.5}
            changeType="percentage"
            trend="up"
            icon={<TrendingUp className="h-4 w-4" />}
            subtitle="Unrealized"
          />
        </div>

        {/* Main Trading Interface */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Chart Section */}
          <div className="lg:col-span-2 space-y-6">
            <TradingViewChart
              symbol={selectedSymbol}
              height={500}
              onSymbolChange={setSelectedSymbol}
            />
            
            {/* Price Chart with Custom Data */}
            <PriceChart
              data={mockPriceData}
              title="BTC/USD Price History"
              symbol="BTC"
              currentPrice={43250}
              priceChange={1250}
              priceChangePercent={2.98}
              timeframe="30D"
              height={300}
            />
          </div>

          {/* Order Book */}
          <div>
            <OrderBook
              data={mockOrderBookData}
              symbol="BTC/USD"
              onPriceClick={(price, side) => {
                console.log(`Clicked ${side} at ${price}`)
              }}
            />
          </div>
        </div>

        {/* Market Overview Tabs */}
        <Tabs defaultValue="heatmap" className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="heatmap">Market Heatmap</TabsTrigger>
            <TabsTrigger value="table">Market Data</TabsTrigger>
            <TabsTrigger value="analytics">Analytics</TabsTrigger>
          </TabsList>
          
          <TabsContent value="heatmap" className="space-y-4">
            <MarketHeatmap
              data={mockHeatmapData}
              onCellClick={(item) => {
                setSelectedSymbol(`${item.symbol}USD`)
                console.log("Selected:", item.symbol)
              }}
            />
          </TabsContent>
          
          <TabsContent value="table" className="space-y-4">
            <MarketDataTable
              data={mockMarketData}
              loading={loading}
              onRowClick={(item) => {
                setSelectedSymbol(`${item.symbol}USD`)
              }}
              onFavorite={(symbol) => {
                console.log("Favorited:", symbol)
              }}
            />
          </TabsContent>
          
          <TabsContent value="analytics" className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <Activity className="h-4 w-4" />
                    <span>Market Sentiment</span>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex justify-between">
                      <span>Fear & Greed Index</span>
                      <Badge variant="bull">72 - Greed</Badge>
                    </div>
                    <div className="w-full bg-muted rounded-full h-2">
                      <div className="bg-green-500 h-2 rounded-full" style={{ width: "72%" }}></div>
                    </div>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <Zap className="h-4 w-4" />
                    <span>Arbitrage Opportunities</span>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span>BTC Binance vs Coinbase</span>
                      <span className="text-green-500">+0.15%</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span>ETH Kraken vs Binance</span>
                      <span className="text-green-500">+0.08%</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <Shield className="h-4 w-4" />
                    <span>Risk Metrics</span>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span>Portfolio Beta</span>
                      <span>1.25</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span>Max Drawdown</span>
                      <span className="text-red-500">-8.5%</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span>Sharpe Ratio</span>
                      <span>2.1</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </DashboardLayout>
  )
}

export { EpicCryptoDashboard }
