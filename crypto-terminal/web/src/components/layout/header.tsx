import * as React from "react"
import { cn } from "../../lib/utils"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "../ui/dropdown-menu"
import {
  Bell,
  Settings,
  User,
  LogOut,
  Moon,
  Sun,
  Wifi,
  WifiOff,
  Activity,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
} from "lucide-react"

interface HeaderProps {
  className?: string
  onThemeToggle?: () => void
  isDark?: boolean
  connectionStatus?: "connected" | "disconnected" | "connecting"
  notifications?: number
}

interface MarketTicker {
  symbol: string
  price: number
  change: number
  changePercent: number
}

const marketTickers: MarketTicker[] = [
  { symbol: "BTC", price: 43250, change: 1250, changePercent: 2.98 },
  { symbol: "ETH", price: 2650, change: -45, changePercent: -1.67 },
  { symbol: "BNB", price: 315, change: 8.5, changePercent: 2.77 },
  { symbol: "SOL", price: 98.5, change: 3.2, changePercent: 3.36 },
  { symbol: "ADA", price: 0.485, change: -0.012, changePercent: -2.41 },
]

const Header: React.FC<HeaderProps> = ({
  className,
  onThemeToggle,
  isDark = false,
  connectionStatus = "connected",
  notifications = 0,
}) => {
  const getConnectionIcon = () => {
    switch (connectionStatus) {
      case "connected":
        return <Wifi className="h-4 w-4 text-green-500" />
      case "disconnected":
        return <WifiOff className="h-4 w-4 text-red-500" />
      case "connecting":
        return <Activity className="h-4 w-4 text-yellow-500 animate-pulse" />
      default:
        return <WifiOff className="h-4 w-4 text-gray-500" />
    }
  }

  const getConnectionStatus = () => {
    switch (connectionStatus) {
      case "connected":
        return "Connected"
      case "disconnected":
        return "Disconnected"
      case "connecting":
        return "Connecting..."
      default:
        return "Unknown"
    }
  }

  return (
    <header className={cn(
      "flex items-center justify-between px-6 py-3 bg-background border-b border-border",
      className
    )}>
      {/* Market Tickers */}
      <div className="flex items-center space-x-6 overflow-x-auto scrollbar-thin">
        {marketTickers.map((ticker) => (
          <div key={ticker.symbol} className="flex items-center space-x-2 min-w-fit">
            <span className="font-medium text-sm">{ticker.symbol}</span>
            <span className="font-mono text-sm">${ticker.price.toLocaleString()}</span>
            <div className="flex items-center space-x-1">
              {ticker.change > 0 ? (
                <TrendingUp className="h-3 w-3 text-green-500" />
              ) : (
                <TrendingDown className="h-3 w-3 text-red-500" />
              )}
              <span className={cn(
                "text-xs font-medium",
                ticker.change > 0 ? "text-green-500" : "text-red-500"
              )}>
                {ticker.changePercent > 0 ? "+" : ""}{ticker.changePercent.toFixed(2)}%
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* Right Section */}
      <div className="flex items-center space-x-4">
        {/* Connection Status */}
        <div className="flex items-center space-x-2">
          {getConnectionIcon()}
          <span className="text-sm text-muted-foreground hidden sm:block">
            {getConnectionStatus()}
          </span>
        </div>

        {/* Theme Toggle */}
        <Button
          variant="ghost"
          size="icon"
          onClick={onThemeToggle}
          className="h-9 w-9"
        >
          {isDark ? (
            <Sun className="h-4 w-4" />
          ) : (
            <Moon className="h-4 w-4" />
          )}
        </Button>

        {/* Notifications */}
        <Button variant="ghost" size="icon" className="h-9 w-9 relative">
          <Bell className="h-4 w-4" />
          {notifications > 0 && (
            <Badge 
              variant="destructive" 
              className="absolute -top-1 -right-1 h-5 w-5 text-xs p-0 flex items-center justify-center"
            >
              {notifications > 99 ? "99+" : notifications}
            </Badge>
          )}
        </Button>

        {/* User Menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="relative h-9 w-9 rounded-full">
              <Avatar className="h-9 w-9">
                <AvatarImage src="/avatars/user.png" alt="User" />
                <AvatarFallback>
                  <User className="h-4 w-4" />
                </AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56" align="end" forceMount>
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col space-y-1">
                <p className="text-sm font-medium leading-none">John Doe</p>
                <p className="text-xs leading-none text-muted-foreground">
                  john.doe@example.com
                </p>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <User className="mr-2 h-4 w-4" />
              <span>Profile</span>
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
            </DropdownMenuItem>
            <DropdownMenuItem>
              <AlertTriangle className="mr-2 h-4 w-4" />
              <span>Risk Settings</span>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="text-red-600">
              <LogOut className="mr-2 h-4 w-4" />
              <span>Log out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}

export { Header }
