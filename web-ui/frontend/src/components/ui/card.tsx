import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const cardVariants = cva(
  "rounded-xl border bg-card text-card-foreground shadow-sm transition-all duration-300",
  {
    variants: {
      variant: {
        default: "border-border backdrop-blur-sm",
        outline: "border-2 border-border",
        ghost: "border-transparent shadow-none",
        elevated: "shadow-xl border-border/50 hover:shadow-2xl",
        glass: "glass-card",
        coffee: "bg-gradient-to-br from-coffee-50 to-coffee-100 border-coffee-200 dark:from-coffee-900/20 dark:to-coffee-800/20 dark:border-coffee-700/30",
        crypto: "crypto-card",
        feature: "feature-card",
        stats: "stats-card",
        metric: "metric-card",
        glow: "shadow-glow hover:shadow-glow-lg border-coffee-500/30",
        premium: "bg-gradient-to-br from-brand-gold/10 to-brand-amber/10 border-brand-gold/30 shadow-xl",
      },
      padding: {
        none: "p-0",
        sm: "p-3",
        default: "p-6",
        lg: "p-8",
      },
    },
    defaultVariants: {
      variant: "default",
      padding: "default",
    },
  }
)

export interface CardProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof cardVariants> {
  hover?: boolean
  interactive?: boolean
}

const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant, padding, hover = false, interactive = false, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        cardVariants({ variant, padding }),
        hover && "hover-lift hover:shadow-xl",
        interactive && "cursor-pointer hover:shadow-xl hover:border-coffee-500/50 hover:scale-[1.02]",
        className
      )}
      {...props}
    />
  )
)
Card.displayName = "Card"

const CardHeader = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex flex-col space-y-1.5 p-6", className)}
    {...props}
  />
))
CardHeader.displayName = "CardHeader"

const CardTitle = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h3
    ref={ref}
    className={cn(
      "text-2xl font-semibold leading-none tracking-tight",
      className
    )}
    {...props}
  />
))
CardTitle.displayName = "CardTitle"

const CardDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn("text-sm text-muted-foreground", className)}
    {...props}
  />
))
CardDescription.displayName = "CardDescription"

const CardContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn("p-6 pt-0", className)} {...props} />
))
CardContent.displayName = "CardContent"

const CardFooter = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex items-center p-6 pt-0", className)}
    {...props}
  />
))
CardFooter.displayName = "CardFooter"

// Metric Card - специальний компонент для метрик
interface MetricCardProps extends CardProps {
  title: string
  value: string | number
  change?: number
  changeLabel?: string
  icon?: React.ReactNode
  loading?: boolean
}

const MetricCard = React.forwardRef<HTMLDivElement, MetricCardProps>(
  ({ 
    title, 
    value, 
    change, 
    changeLabel, 
    icon, 
    loading = false, 
    className,
    ...props 
  }, ref) => (
    <Card
      ref={ref}
      className={cn("metric-card", className)}
      hover
      {...props}
    >
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <p className="text-sm font-medium text-muted-foreground">{title}</p>
            {loading ? (
              <div className="h-8 w-24 bg-muted animate-pulse rounded mt-2" />
            ) : (
              <p className="text-2xl font-bold mt-2">{value}</p>
            )}
            {change !== undefined && !loading && (
              <p className={cn(
                "text-xs mt-1 flex items-center gap-1",
                change >= 0 ? "text-green-600" : "text-red-600"
              )}>
                <span>{change >= 0 ? "↗" : "↘"}</span>
                {Math.abs(change).toFixed(2)}%
                {changeLabel && <span className="text-muted-foreground">({changeLabel})</span>}
              </p>
            )}
          </div>
          {icon && (
            <div className="text-muted-foreground">
              {icon}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
)
MetricCard.displayName = "MetricCard"

export { 
  Card, 
  CardHeader, 
  CardFooter, 
  CardTitle, 
  CardDescription, 
  CardContent,
  MetricCard,
  cardVariants 
}
