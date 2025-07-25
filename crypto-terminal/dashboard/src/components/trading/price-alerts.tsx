'use client'

import * as React from "react"
import { motion } from "framer-motion"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn, formatCurrency } from "@/lib/utils"
import {
  Bell,
  BellRing,
  Plus,
  Trash2,
  Check
} from "lucide-react"

interface PriceAlert {
  id: string
  symbol: string
  targetPrice: number
  currentPrice: number
  condition: "above" | "below"
  isActive: boolean
  isTriggered: boolean
  createdAt: Date
  triggeredAt?: Date
}

interface PriceAlertsProps {
  className?: string
  onAlertCreate?: (alert: Omit<PriceAlert, 'id' | 'createdAt'>) => void
  onAlertDelete?: (alertId: string) => void
}

// Mock alerts data
const mockAlerts: PriceAlert[] = [
  {
    id: "1",
    symbol: "BTC",
    targetPrice: 45000,
    currentPrice: 43250,
    condition: "above",
    isActive: true,
    isTriggered: false,
    createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
  },
  {
    id: "2",
    symbol: "ETH",
    targetPrice: 2500,
    currentPrice: 2650,
    condition: "below",
    isActive: true,
    isTriggered: false,
    createdAt: new Date(Date.now() - 1 * 60 * 60 * 1000), // 1 hour ago
  },
  {
    id: "3",
    symbol: "SOL",
    targetPrice: 100,
    currentPrice: 98,
    condition: "above",
    isActive: false,
    isTriggered: true,
    createdAt: new Date(Date.now() - 30 * 60 * 1000), // 30 minutes ago
    triggeredAt: new Date(Date.now() - 10 * 60 * 1000), // 10 minutes ago
  },
]

const PriceAlerts: React.FC<PriceAlertsProps> = ({
  className,
  onAlertDelete,
}) => {
  const [alerts, setAlerts] = React.useState<PriceAlert[]>(mockAlerts)
  const [showCreateForm, setShowCreateForm] = React.useState(false)

  // Simulate real-time price updates and alert checking
  React.useEffect(() => {
    const interval = setInterval(() => {
      setAlerts(prev => prev.map(alert => {
        if (!alert.isActive || alert.isTriggered) return alert

        // Simulate price movement
        const priceChange = (Math.random() - 0.5) * 100
        const newPrice = Math.max(0, alert.currentPrice + priceChange)
        
        // Check if alert should be triggered
        const shouldTrigger = 
          (alert.condition === "above" && newPrice >= alert.targetPrice) ||
          (alert.condition === "below" && newPrice <= alert.targetPrice)

        if (shouldTrigger) {
          return {
            ...alert,
            currentPrice: newPrice,
            isActive: false,
            isTriggered: true,
            triggeredAt: new Date()
          }
        }

        return {
          ...alert,
          currentPrice: newPrice
        }
      }))
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  const handleDeleteAlert = (alertId: string) => {
    setAlerts(prev => prev.filter(alert => alert.id !== alertId))
    onAlertDelete?.(alertId)
  }

  const getAlertStatus = (alert: PriceAlert) => {
    if (alert.isTriggered) return { color: "text-green-500", icon: Check, label: "Triggered" }
    if (!alert.isActive) return { color: "text-gray-500", icon: Bell, label: "Inactive" }
    return { color: "text-blue-500", icon: BellRing, label: "Active" }
  }

  const getProgressToTarget = (alert: PriceAlert) => {
    if (alert.condition === "above") {
      return Math.min(100, (alert.currentPrice / alert.targetPrice) * 100)
    } else {
      return Math.min(100, ((alert.targetPrice - alert.currentPrice) / alert.targetPrice) * 100 + 50)
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.5, delay: 0.1 }}
    >
      <Card variant="glass" className={cn("w-full", className)}>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Bell className="h-5 w-5 text-primary" />
              <span>Price Alerts</span>
              <Badge variant="glass">{alerts.filter(a => a.isActive).length} Active</Badge>
            </CardTitle>
            
            <Button
              variant="epic"
              size="sm"
              onClick={() => setShowCreateForm(!showCreateForm)}
            >
              <Plus className="h-4 w-4 mr-1" />
              Add Alert
            </Button>
          </div>
        </CardHeader>
        
        <CardContent className="space-y-3">
          {/* Create Alert Form */}
          {showCreateForm && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              exit={{ opacity: 0, height: 0 }}
              className="p-3 glass rounded-lg border"
            >
              <div className="text-sm font-medium mb-2">Create New Alert</div>
              <div className="grid grid-cols-2 gap-2 text-xs">
                <input
                  type="text"
                  placeholder="Symbol (e.g., BTC)"
                  className="px-2 py-1 bg-background/50 border rounded"
                />
                <input
                  type="number"
                  placeholder="Target Price"
                  className="px-2 py-1 bg-background/50 border rounded"
                />
                <select className="px-2 py-1 bg-background/50 border rounded">
                  <option value="above">Above</option>
                  <option value="below">Below</option>
                </select>
                <Button size="sm" className="h-7">
                  Create
                </Button>
              </div>
            </motion.div>
          )}

          {/* Alerts List */}
          <div className="space-y-2 max-h-64 overflow-y-auto scrollbar-thin">
            {alerts.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <Bell className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">No price alerts set</p>
                <p className="text-xs">Create your first alert to get notified</p>
              </div>
            ) : (
              alerts.map((alert, index) => {
                const status = getAlertStatus(alert)
                const progress = getProgressToTarget(alert)
                const StatusIcon = status.icon
                
                return (
                  <motion.div
                    key={alert.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.1 }}
                    className={cn(
                      "p-3 rounded-lg border glass transition-all",
                      alert.isTriggered && "border-green-500/50 bg-green-500/5",
                      alert.isActive && "border-blue-500/50"
                    )}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <Badge variant="glass" className="text-xs">
                          {alert.symbol}
                        </Badge>
                        <div className={cn("flex items-center space-x-1", status.color)}>
                          <StatusIcon className="h-3 w-3" />
                          <span className="text-xs">{status.label}</span>
                        </div>
                      </div>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
                        onClick={() => handleDeleteAlert(alert.id)}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>
                    
                    <div className="space-y-1">
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-muted-foreground">
                          {alert.condition === "above" ? "Above" : "Below"} {formatCurrency(alert.targetPrice, 'USD', 2)}
                        </span>
                        <span className="font-mono">
                          Current: {formatCurrency(alert.currentPrice, 'USD', 2)}
                        </span>
                      </div>
                      
                      {/* Progress bar */}
                      <div className="w-full bg-muted/30 rounded-full h-1">
                        <div
                          className={cn(
                            "h-1 rounded-full transition-all duration-300",
                            alert.condition === "above" ? "bg-bull" : "bg-bear"
                          )}
                          style={{ width: `${progress}%` }}
                        />
                      </div>
                      
                      <div className="flex items-center justify-between text-2xs text-muted-foreground">
                        <span>
                          Created: {alert.createdAt.toLocaleTimeString()}
                        </span>
                        {alert.triggeredAt && (
                          <span className="text-green-500">
                            Triggered: {alert.triggeredAt.toLocaleTimeString()}
                          </span>
                        )}
                      </div>
                    </div>
                  </motion.div>
                )
              })
            )}
          </div>
          
          {/* Summary */}
          <div className="pt-2 border-t glass">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>{alerts.filter(a => a.isActive).length} active alerts</span>
              <span>{alerts.filter(a => a.isTriggered).length} triggered today</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}

export { PriceAlerts }
