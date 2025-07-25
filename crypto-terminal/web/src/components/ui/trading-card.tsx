import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "./card"
import { Badge } from "./badge"
import { cn, formatCurrency, formatPercentage, getChangeColor, getChangeIcon } from "../../lib/utils"

interface TradingCardProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string
  value: number | string
  change?: number
  changeType?: "percentage" | "currency" | "number"
  currency?: string
  subtitle?: string
  icon?: React.ReactNode
  trend?: "up" | "down" | "neutral"
  loading?: boolean
  size?: "sm" | "md" | "lg"
}

const TradingCard = React.forwardRef<HTMLDivElement, TradingCardProps>(
  ({
    className,
    title,
    value,
    change,
    changeType = "percentage",
    currency = "USD",
    subtitle,
    icon,
    trend,
    loading = false,
    size = "md",
    ...props
  }, ref) => {
    const formatValue = (val: number | string) => {
      if (typeof val === "string") return val
      if (changeType === "currency") return formatCurrency(val, currency)
      if (changeType === "percentage") return formatPercentage(val / 100)
      return val.toLocaleString()
    }

    const formatChange = (val: number) => {
      if (changeType === "currency") return formatCurrency(val, currency)
      if (changeType === "percentage") return formatPercentage(val / 100)
      return val.toLocaleString()
    }

    const getTrendColor = () => {
      if (trend) {
        switch (trend) {
          case "up": return "text-bull"
          case "down": return "text-bear"
          default: return "text-neutral"
        }
      }
      if (change !== undefined) {
        return getChangeColor(change)
      }
      return "text-foreground"
    }

    const sizeClasses = {
      sm: "p-3",
      md: "p-4",
      lg: "p-6"
    }

    const titleSizeClasses = {
      sm: "text-sm",
      md: "text-base",
      lg: "text-lg"
    }

    const valueSizeClasses = {
      sm: "text-lg",
      md: "text-2xl",
      lg: "text-3xl"
    }

    if (loading) {
      return (
        <Card ref={ref} className={cn("animate-pulse", className)} {...props}>
          <CardHeader className={cn("pb-2", sizeClasses[size])}>
            <div className="h-4 bg-muted rounded w-3/4"></div>
          </CardHeader>
          <CardContent className={cn("pt-0", sizeClasses[size])}>
            <div className="h-8 bg-muted rounded w-1/2 mb-2"></div>
            <div className="h-3 bg-muted rounded w-1/4"></div>
          </CardContent>
        </Card>
      )
    }

    return (
      <Card 
        ref={ref} 
        className={cn(
          "transition-all duration-200 hover:shadow-lg hover:scale-[1.02]",
          className
        )} 
        {...props}
      >
        <CardHeader className={cn("pb-2", sizeClasses[size])}>
          <CardTitle className={cn(
            "flex items-center justify-between",
            titleSizeClasses[size]
          )}>
            <span className="text-muted-foreground">{title}</span>
            {icon && <span className="text-muted-foreground">{icon}</span>}
          </CardTitle>
        </CardHeader>
        <CardContent className={cn("pt-0", sizeClasses[size])}>
          <div className="space-y-1">
            <div className={cn(
              "font-bold tracking-tight",
              valueSizeClasses[size],
              getTrendColor()
            )}>
              {formatValue(value)}
            </div>
            
            {change !== undefined && (
              <div className="flex items-center space-x-1">
                <span className={cn("text-sm", getChangeColor(change))}>
                  {getChangeIcon(change)} {formatChange(Math.abs(change))}
                </span>
                {trend && (
                  <Badge 
                    variant={trend === "up" ? "bull" : trend === "down" ? "bear" : "neutral"}
                    className="text-xs"
                  >
                    {trend.toUpperCase()}
                  </Badge>
                )}
              </div>
            )}
            
            {subtitle && (
              <p className="text-xs text-muted-foreground">{subtitle}</p>
            )}
          </div>
        </CardContent>
      </Card>
    )
  }
)

TradingCard.displayName = "TradingCard"

export { TradingCard }
