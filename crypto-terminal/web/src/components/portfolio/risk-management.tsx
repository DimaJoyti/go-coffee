import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card"
import { Badge } from "../ui/badge"
import { Button } from "../ui/button"
import { Progress } from "../ui/progress"
import { Alert, AlertDescription } from "../ui/alert"
import { cn, formatCurrency, formatPercentage } from "../../lib/utils"
import { 
  Shield, 
  AlertTriangle, 
  TrendingDown, 
  Target,
  Activity,
  BarChart3,
  Settings,
  CheckCircle,
  XCircle,
  Clock
} from "lucide-react"

interface RiskMetric {
  name: string
  value: number
  threshold: number
  status: "safe" | "warning" | "danger"
  description: string
}

interface RiskAlert {
  id: string
  type: "warning" | "danger" | "info"
  title: string
  message: string
  timestamp: number
  acknowledged: boolean
}

interface RiskManagementProps {
  portfolioValue: number
  maxDrawdown: number
  volatility: number
  sharpeRatio: number
  beta: number
  var95: number
  riskMetrics: RiskMetric[]
  alerts: RiskAlert[]
  loading?: boolean
  className?: string
  onAcknowledgeAlert?: (alertId: string) => void
  onUpdateRiskSettings?: () => void
}

const RiskManagement: React.FC<RiskManagementProps> = ({
  portfolioValue,
  maxDrawdown,
  volatility,
  sharpeRatio,
  beta,
  var95,
  riskMetrics,
  alerts,
  loading = false,
  className,
  onAcknowledgeAlert,
  onUpdateRiskSettings,
}) => {
  const getRiskLevel = (value: number, thresholds: { low: number; medium: number }) => {
    if (value <= thresholds.low) return { level: "Low", color: "text-green-500", bg: "bg-green-500/10" }
    if (value <= thresholds.medium) return { level: "Medium", color: "text-yellow-500", bg: "bg-yellow-500/10" }
    return { level: "High", color: "text-red-500", bg: "bg-red-500/10" }
  }

  const portfolioRisk = getRiskLevel(volatility * 100, { low: 15, medium: 25 })
  const drawdownRisk = getRiskLevel(Math.abs(maxDrawdown) * 100, { low: 10, medium: 20 })

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "safe":
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />
      case "danger":
        return <XCircle className="h-4 w-4 text-red-500" />
      default:
        return <Clock className="h-4 w-4 text-gray-500" />
    }
  }

  const getAlertIcon = (type: string) => {
    switch (type) {
      case "danger":
        return <XCircle className="h-4 w-4 text-red-500" />
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />
      default:
        return <Activity className="h-4 w-4 text-blue-500" />
    }
  }

  return (
    <div className={cn("space-y-6", className)}>
      {/* Risk Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center space-x-2">
              <Shield className="h-4 w-4" />
              <span>Portfolio Risk</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-2xl font-bold">{(volatility * 100).toFixed(1)}%</span>
                <Badge variant="outline" className={cn("text-xs", portfolioRisk.color)}>
                  {portfolioRisk.level}
                </Badge>
              </div>
              <div className="text-xs text-muted-foreground">30-day volatility</div>
              <Progress value={volatility * 100} className="h-2" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center space-x-2">
              <TrendingDown className="h-4 w-4" />
              <span>Max Drawdown</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-2xl font-bold text-red-500">
                  {formatPercentage(maxDrawdown)}
                </span>
                <Badge variant="outline" className={cn("text-xs", drawdownRisk.color)}>
                  {drawdownRisk.level}
                </Badge>
              </div>
              <div className="text-xs text-muted-foreground">Peak to trough</div>
              <Progress value={Math.abs(maxDrawdown) * 100} className="h-2" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center space-x-2">
              <Target className="h-4 w-4" />
              <span>Sharpe Ratio</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-2xl font-bold">{sharpeRatio.toFixed(2)}</span>
                <Badge variant={sharpeRatio > 1 ? "bull" : sharpeRatio > 0.5 ? "neutral" : "bear"} className="text-xs">
                  {sharpeRatio > 1 ? "Excellent" : sharpeRatio > 0.5 ? "Good" : "Poor"}
                </Badge>
              </div>
              <div className="text-xs text-muted-foreground">Risk-adjusted return</div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center space-x-2">
              <BarChart3 className="h-4 w-4" />
              <span>Portfolio Beta</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-2xl font-bold">{beta.toFixed(2)}</span>
                <Badge variant="outline" className="text-xs">
                  {beta > 1 ? "High" : beta > 0.5 ? "Medium" : "Low"}
                </Badge>
              </div>
              <div className="text-xs text-muted-foreground">Market correlation</div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Risk Alerts */}
      {alerts.length > 0 && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center space-x-2">
                <AlertTriangle className="h-5 w-5" />
                <span>Risk Alerts</span>
                <Badge variant="destructive" className="text-xs">
                  {alerts.filter(alert => !alert.acknowledged).length}
                </Badge>
              </CardTitle>
              <Button variant="outline" size="sm" onClick={onUpdateRiskSettings}>
                <Settings className="h-4 w-4 mr-1" />
                Settings
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {alerts.slice(0, 5).map((alert) => (
                <Alert key={alert.id} className={cn(
                  "border-l-4",
                  alert.type === "danger" && "border-l-red-500",
                  alert.type === "warning" && "border-l-yellow-500",
                  alert.type === "info" && "border-l-blue-500",
                  alert.acknowledged && "opacity-50"
                )}>
                  <div className="flex items-start space-x-3">
                    {getAlertIcon(alert.type)}
                    <div className="flex-1 space-y-1">
                      <div className="flex items-center justify-between">
                        <h4 className="text-sm font-medium">{alert.title}</h4>
                        <div className="flex items-center space-x-2">
                          <span className="text-xs text-muted-foreground">
                            {new Date(alert.timestamp).toLocaleTimeString()}
                          </span>
                          {!alert.acknowledged && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => onAcknowledgeAlert?.(alert.id)}
                              className="h-6 text-xs"
                            >
                              Acknowledge
                            </Button>
                          )}
                        </div>
                      </div>
                      <AlertDescription className="text-xs">
                        {alert.message}
                      </AlertDescription>
                    </div>
                  </div>
                </Alert>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Risk Metrics */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Activity className="h-5 w-5" />
            <span>Risk Metrics</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {loading ? (
              Array.from({ length: 5 }).map((_, index) => (
                <div key={index} className="space-y-2">
                  <div className="h-4 bg-muted rounded animate-pulse w-1/3" />
                  <div className="h-2 bg-muted rounded animate-pulse" />
                </div>
              ))
            ) : (
              riskMetrics.map((metric) => (
                <div key={metric.name} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(metric.status)}
                      <span className="text-sm font-medium">{metric.name}</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className="text-sm font-mono">
                        {metric.value.toFixed(2)}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        / {metric.threshold.toFixed(2)}
                      </span>
                    </div>
                  </div>
                  <div className="space-y-1">
                    <Progress 
                      value={(metric.value / metric.threshold) * 100} 
                      className={cn(
                        "h-2",
                        metric.status === "safe" && "[&>div]:bg-green-500",
                        metric.status === "warning" && "[&>div]:bg-yellow-500",
                        metric.status === "danger" && "[&>div]:bg-red-500"
                      )}
                    />
                    <p className="text-xs text-muted-foreground">{metric.description}</p>
                  </div>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>

      {/* Value at Risk */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Shield className="h-5 w-5" />
            <span>Value at Risk (VaR)</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="text-center space-y-2">
              <div className="text-sm text-muted-foreground">1-Day VaR (95%)</div>
              <div className="text-2xl font-bold text-red-500">
                {formatCurrency(var95, "USD")}
              </div>
              <div className="text-xs text-muted-foreground">
                {formatPercentage(var95 / portfolioValue)}
              </div>
            </div>
            <div className="text-center space-y-2">
              <div className="text-sm text-muted-foreground">Portfolio Value</div>
              <div className="text-2xl font-bold">
                {formatCurrency(portfolioValue, "USD")}
              </div>
              <div className="text-xs text-muted-foreground">Current</div>
            </div>
            <div className="text-center space-y-2">
              <div className="text-sm text-muted-foreground">Risk Exposure</div>
              <div className="text-2xl font-bold text-yellow-500">
                {formatPercentage((var95 / portfolioValue))}
              </div>
              <div className="text-xs text-muted-foreground">of portfolio</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export { RiskManagement, type RiskMetric, type RiskAlert }
