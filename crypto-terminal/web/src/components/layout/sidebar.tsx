import * as React from "react"
import { cn } from "../../lib/utils"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { Separator } from "../ui/separator"
import {
  BarChart3,
  TrendingUp,
  Wallet,
  Settings,
  Bell,
  Search,
  Home,
  Activity,
  PieChart,
  Zap,
  Shield,
  HelpCircle,
  ChevronLeft,
  ChevronRight,
  Bitcoin,
  DollarSign,
  LineChart,
  Target,
} from "lucide-react"

interface SidebarProps {
  collapsed?: boolean
  onToggle?: () => void
  className?: string
}

interface NavItem {
  title: string
  icon: React.ReactNode
  href: string
  badge?: string
  active?: boolean
  children?: NavItem[]
}

const navItems: NavItem[] = [
  {
    title: "Dashboard",
    icon: <Home className="h-4 w-4" />,
    href: "/",
    active: true,
  },
  {
    title: "Markets",
    icon: <TrendingUp className="h-4 w-4" />,
    href: "/markets",
    badge: "Live",
  },
  {
    title: "Trading",
    icon: <BarChart3 className="h-4 w-4" />,
    href: "/trading",
    children: [
      {
        title: "Spot Trading",
        icon: <DollarSign className="h-3 w-3" />,
        href: "/trading/spot",
      },
      {
        title: "Futures",
        icon: <LineChart className="h-3 w-3" />,
        href: "/trading/futures",
      },
      {
        title: "Options",
        icon: <Target className="h-3 w-3" />,
        href: "/trading/options",
      },
    ],
  },
  {
    title: "Portfolio",
    icon: <Wallet className="h-4 w-4" />,
    href: "/portfolio",
  },
  {
    title: "Analytics",
    icon: <PieChart className="h-4 w-4" />,
    href: "/analytics",
  },
  {
    title: "Arbitrage",
    icon: <Zap className="h-4 w-4" />,
    href: "/arbitrage",
    badge: "Pro",
  },
  {
    title: "Risk Management",
    icon: <Shield className="h-4 w-4" />,
    href: "/risk",
  },
]

const bottomNavItems: NavItem[] = [
  {
    title: "Notifications",
    icon: <Bell className="h-4 w-4" />,
    href: "/notifications",
    badge: "3",
  },
  {
    title: "Settings",
    icon: <Settings className="h-4 w-4" />,
    href: "/settings",
  },
  {
    title: "Help",
    icon: <HelpCircle className="h-4 w-4" />,
    href: "/help",
  },
]

const Sidebar: React.FC<SidebarProps> = ({
  collapsed = false,
  onToggle,
  className,
}) => {
  const [expandedItems, setExpandedItems] = React.useState<string[]>([])

  const toggleExpanded = (title: string) => {
    setExpandedItems(prev =>
      prev.includes(title)
        ? prev.filter(item => item !== title)
        : [...prev, title]
    )
  }

  const NavItemComponent = ({ item, level = 0 }: { item: NavItem; level?: number }) => {
    const hasChildren = item.children && item.children.length > 0
    const isExpanded = expandedItems.includes(item.title)
    const paddingLeft = level * 12 + 16

    return (
      <div key={item.title}>
        <Button
          variant={item.active ? "secondary" : "ghost"}
          className={cn(
            "w-full justify-start h-10 px-3 mb-1",
            collapsed && !hasChildren && "justify-center px-2",
            item.active && "bg-primary/10 text-primary border-r-2 border-primary"
          )}
          style={{ paddingLeft: collapsed ? undefined : paddingLeft }}
          onClick={() => hasChildren ? toggleExpanded(item.title) : undefined}
        >
          <span className="flex items-center space-x-3 w-full">
            <span className="flex-shrink-0">{item.icon}</span>
            {!collapsed && (
              <>
                <span className="flex-1 text-left">{item.title}</span>
                {item.badge && (
                  <Badge 
                    variant={item.badge === "Live" ? "bull" : item.badge === "Pro" ? "info" : "secondary"}
                    className="text-xs"
                  >
                    {item.badge}
                  </Badge>
                )}
                {hasChildren && (
                  <span className="flex-shrink-0">
                    {isExpanded ? (
                      <ChevronLeft className="h-3 w-3" />
                    ) : (
                      <ChevronRight className="h-3 w-3" />
                    )}
                  </span>
                )}
              </>
            )}
          </span>
        </Button>
        
        {hasChildren && isExpanded && !collapsed && (
          <div className="ml-4 space-y-1">
            {item.children?.map(child => (
              <NavItemComponent key={child.title} item={child} level={level + 1} />
            ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className={cn(
      "flex flex-col h-full bg-card border-r border-border transition-all duration-300",
      collapsed ? "w-16" : "w-64",
      className
    )}>
      {/* Header */}
      <div className="p-4 border-b border-border">
        <div className="flex items-center justify-between">
          {!collapsed && (
            <div className="flex items-center space-x-2">
              <Bitcoin className="h-6 w-6 text-primary" />
              <span className="font-bold text-lg">CryptoTerminal</span>
            </div>
          )}
          {collapsed && (
            <Bitcoin className="h-6 w-6 text-primary mx-auto" />
          )}
          {onToggle && (
            <Button
              variant="ghost"
              size="icon"
              onClick={onToggle}
              className="h-8 w-8"
            >
              {collapsed ? (
                <ChevronRight className="h-4 w-4" />
              ) : (
                <ChevronLeft className="h-4 w-4" />
              )}
            </Button>
          )}
        </div>
      </div>

      {/* Search */}
      {!collapsed && (
        <div className="p-4">
          <Button variant="outline" className="w-full justify-start text-muted-foreground">
            <Search className="h-4 w-4 mr-2" />
            Search...
          </Button>
        </div>
      )}

      {/* Navigation */}
      <div className="flex-1 px-3 py-2 space-y-1 overflow-y-auto scrollbar-thin">
        {navItems.map(item => (
          <NavItemComponent key={item.title} item={item} />
        ))}
      </div>

      <Separator />

      {/* Bottom Navigation */}
      <div className="p-3 space-y-1">
        {bottomNavItems.map(item => (
          <NavItemComponent key={item.title} item={item} />
        ))}
      </div>

      {/* Market Status */}
      {!collapsed && (
        <div className="p-4 border-t border-border">
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Market Status</span>
              <Badge variant="bull" className="text-xs">Open</Badge>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">BTC/USD</span>
              <span className="font-mono text-bull">$43,250</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">ETH/USD</span>
              <span className="font-mono text-bull">$2,650</span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export { Sidebar }
