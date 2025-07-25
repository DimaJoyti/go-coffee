import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card"
import { TradingCard } from "../ui/trading-card"
import { Badge } from "../ui/badge"
import { Button } from "../ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs"
import { Progress } from "../ui/progress"
import { cn, formatCurrency, formatPercentage } from "../../lib/utils"
import { 
  Wallet, 
  TrendingUp, 
  TrendingDown, 
  PieChart, 
  BarChart3,
  Target,
  Shield,
  AlertTriangle,
  Plus,
  Minus,
  RotateCcw
} from "lucide-react"
import { PieChart as RechartsPieChart, Cell, ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip } from "recharts"

interface PortfolioHolding {
  symbol: string
  name: string
  amount: number
  value: number
  price: number
  change24h: number
  changePercent24h: number
  allocation: number
  logo?: string
}

interface PortfolioMetrics {
  totalValue: number
  totalChange24h: number
  totalChangePercent24h: number
  totalPnL: number
  totalPnLPercent: number
  bestPerformer: PortfolioHolding
  worstPerformer: PortfolioHolding
}

interface PortfolioOverviewProps {
  holdings: PortfolioHolding[]
  metrics: PortfolioMetrics
  historicalData: Array<{ timestamp: number; value: number }>
  loading?: boolean
  className?: string
}

const PortfolioOverview: React.FC<PortfolioOverviewProps> = ({
  holdings,
  metrics,
  historicalData,
  loading = false,
  className,
}) => {
  const [timeframe, setTimeframe] = React.useState<"24h" | "7d" | "30d" | "1y">("7d")

  // Prepare data for pie chart
  const pieData = holdings.map(holding => ({
    name: holding.symbol,
    value: holding.allocation,
    amount: holding.value,
    color: `hsl(${Math.random() * 360}, 70%, 50%)`
  }))

  // Colors for pie chart
  const COLORS = [
    '#10b981', '#3b82f6', '#8b5cf6', '#f59e0b', '#ef4444',
    '#06b6d4', '#84cc16', '#f97316', '#ec4899', '#6366f1'
  ]

  const AllocationChart = () => (
    <ResponsiveContainer width="100%" height={300}>
      <RechartsPieChart>
        <RechartsPieChart
          data={pieData}
          cx="50%"
          cy="50%"
          innerRadius={60}
          outerRadius={120}
          paddingAngle={2}
          dataKey="value"
        >
          {pieData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </RechartsPieChart>
        <Tooltip
          formatter={(value: number, name: string) => [
            `${value.toFixed(2)}%`,
            name
          ]}
        />
      </RechartsPieChart>
    </ResponsiveContainer>
  )

  const PerformanceChart = () => (
    <ResponsiveContainer width="100%" height={300}>
      <AreaChart data={historicalData}>
        <defs>
          <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
            <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
          </linearGradient>
        </defs>
        <XAxis 
          dataKey="timestamp" 
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          tickFormatter={(value) => new Date(value).toLocaleDateString()}
        />
        <YAxis 
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          tickFormatter={(value) => formatCurrency(value, "USD", 0)}
        />
        <Tooltip
          formatter={(value: number) => [formatCurrency(value, "USD"), "Portfolio Value"]}
          labelFormatter={(value) => new Date(value).toLocaleString()}
        />
        <Area
          type="monotone"
          dataKey="value"
          stroke="#10b981"
          strokeWidth={2}
          fill="url(#colorValue)"
        />
      </AreaChart>
    </ResponsiveContainer>
  )

  return (
    <div className={cn("space-y-6", className)}>
      {/* Portfolio Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <TradingCard
          title="Total Portfolio Value"
          value={metrics.totalValue}
          change={metrics.totalChange24h}
          changeType="currency"
          currency="USD"
          trend={metrics.totalChange24h > 0 ? "up" : "down"}
          icon={<Wallet className="h-4 w-4" />}
          subtitle="24h change"
          loading={loading}
        />
        <TradingCard
          title="Total P&L"
          value={metrics.totalPnL}
          change={metrics.totalPnLPercent}
          changeType="percentage"
          trend={metrics.totalPnL > 0 ? "up" : "down"}
          icon={<TrendingUp className="h-4 w-4" />}
          subtitle="All time"
          loading={loading}
        />
        <TradingCard
          title="Best Performer"
          value={metrics.bestPerformer?.symbol || "N/A"}
          change={metrics.bestPerformer?.changePercent24h || 0}
          changeType="percentage"
          trend="up"
          icon={<Target className="h-4 w-4" />}
          subtitle={metrics.bestPerformer?.name || ""}
          loading={loading}
        />
        <TradingCard
          title="Worst Performer"
          value={metrics.worstPerformer?.symbol || "N/A"}
          change={metrics.worstPerformer?.changePercent24h || 0}
          changeType="percentage"
          trend="down"
          icon={<AlertTriangle className="h-4 w-4" />}
          subtitle={metrics.worstPerformer?.name || ""}
          loading={loading}
        />
      </div>

      {/* Charts and Holdings */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Portfolio Performance Chart */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center space-x-2">
                  <BarChart3 className="h-5 w-5" />
                  <span>Portfolio Performance</span>
                </CardTitle>
                <div className="flex items-center space-x-2">
                  <Tabs value={timeframe} onValueChange={(value) => setTimeframe(value as any)}>
                    <TabsList className="grid w-full grid-cols-4">
                      <TabsTrigger value="24h">24H</TabsTrigger>
                      <TabsTrigger value="7d">7D</TabsTrigger>
                      <TabsTrigger value="30d">30D</TabsTrigger>
                      <TabsTrigger value="1y">1Y</TabsTrigger>
                    </TabsList>
                  </Tabs>
                  <Button variant="outline" size="sm">
                    <RotateCcw className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="h-[300px] bg-muted rounded animate-pulse" />
              ) : (
                <PerformanceChart />
              )}
            </CardContent>
          </Card>
        </div>

        {/* Asset Allocation */}
        <div>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <PieChart className="h-5 w-5" />
                <span>Asset Allocation</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="h-[300px] bg-muted rounded animate-pulse" />
              ) : (
                <div className="space-y-4">
                  <AllocationChart />
                  <div className="space-y-2">
                    {pieData.slice(0, 5).map((item, index) => (
                      <div key={item.name} className="flex items-center justify-between text-sm">
                        <div className="flex items-center space-x-2">
                          <div 
                            className="w-3 h-3 rounded-full"
                            style={{ backgroundColor: COLORS[index % COLORS.length] }}
                          />
                          <span>{item.name}</span>
                        </div>
                        <span className="font-medium">{item.value.toFixed(1)}%</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Holdings Table */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Holdings</CardTitle>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm">
                <Plus className="h-4 w-4 mr-1" />
                Add Asset
              </Button>
              <Button variant="outline" size="sm">
                <Minus className="h-4 w-4 mr-1" />
                Remove
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3">Asset</th>
                  <th className="pb-3 text-right">Amount</th>
                  <th className="pb-3 text-right">Price</th>
                  <th className="pb-3 text-right">Value</th>
                  <th className="pb-3 text-right">24h Change</th>
                  <th className="pb-3 text-right">Allocation</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  Array.from({ length: 5 }).map((_, index) => (
                    <tr key={index} className="border-b">
                      <td className="py-3"><div className="h-4 bg-muted rounded animate-pulse w-20" /></td>
                      <td className="py-3 text-right"><div className="h-4 bg-muted rounded animate-pulse w-16 ml-auto" /></td>
                      <td className="py-3 text-right"><div className="h-4 bg-muted rounded animate-pulse w-16 ml-auto" /></td>
                      <td className="py-3 text-right"><div className="h-4 bg-muted rounded animate-pulse w-20 ml-auto" /></td>
                      <td className="py-3 text-right"><div className="h-4 bg-muted rounded animate-pulse w-16 ml-auto" /></td>
                      <td className="py-3 text-right"><div className="h-4 bg-muted rounded animate-pulse w-12 ml-auto" /></td>
                    </tr>
                  ))
                ) : (
                  holdings.map((holding) => (
                    <tr key={holding.symbol} className="border-b hover:bg-muted/50 transition-colors">
                      <td className="py-3">
                        <div className="flex items-center space-x-3">
                          {holding.logo && (
                            <img src={holding.logo} alt={holding.symbol} className="w-6 h-6 rounded-full" />
                          )}
                          <div>
                            <div className="font-medium">{holding.symbol}</div>
                            <div className="text-sm text-muted-foreground">{holding.name}</div>
                          </div>
                        </div>
                      </td>
                      <td className="py-3 text-right font-mono">{holding.amount.toFixed(6)}</td>
                      <td className="py-3 text-right font-mono">{formatCurrency(holding.price, "USD")}</td>
                      <td className="py-3 text-right font-mono">{formatCurrency(holding.value, "USD")}</td>
                      <td className="py-3 text-right">
                        <div className={cn(
                          "flex items-center justify-end space-x-1",
                          holding.changePercent24h > 0 ? "text-green-500" : "text-red-500"
                        )}>
                          {holding.changePercent24h > 0 ? (
                            <TrendingUp className="h-3 w-3" />
                          ) : (
                            <TrendingDown className="h-3 w-3" />
                          )}
                          <span className="font-medium">
                            {formatPercentage(holding.changePercent24h / 100)}
                          </span>
                        </div>
                      </td>
                      <td className="py-3 text-right">
                        <div className="flex items-center justify-end space-x-2">
                          <Progress value={holding.allocation} className="w-12 h-2" />
                          <span className="text-sm font-medium w-12">
                            {holding.allocation.toFixed(1)}%
                          </span>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export { PortfolioOverview, type PortfolioHolding, type PortfolioMetrics }
