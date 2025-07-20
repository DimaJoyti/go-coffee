'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { useTradingStore } from '@/stores/trading-store'
import { Button } from '@/components/ui/button'
import {
  LayoutDashboard,
  TrendingUp,
  Wallet,
  Coffee,
  BarChart3,
  Settings,
  Menu,
  Bell,
  Activity,
  ArrowLeftRight,
} from 'lucide-react'

const navigation = [
  {
    name: 'Dashboard',
    href: '/',
    icon: LayoutDashboard,
  },
  {
    name: 'Portfolio',
    href: '/portfolio',
    icon: Wallet,
  },
  {
    name: 'Coffee Strategies',
    href: '/strategies',
    icon: Coffee,
  },
  {
    name: 'Markets',
    href: '/markets',
    icon: TrendingUp,
  },
  {
    name: 'Arbitrage',
    href: '/arbitrage',
    icon: ArrowLeftRight,
  },
  {
    name: 'Analytics',
    href: '/analytics',
    icon: BarChart3,
  },
  {
    name: 'Alerts',
    href: '/alerts',
    icon: Bell,
  },
  {
    name: 'Activity',
    href: '/activity',
    icon: Activity,
  },
  {
    name: 'Settings',
    href: '/settings',
    icon: Settings,
  },
]

export function Sidebar() {
  const pathname = usePathname()
  const { sidebarCollapsed, toggleSidebar, isConnected } = useTradingStore()

  return (
    <div
      className={cn(
        'fixed inset-y-0 left-0 z-50 flex flex-col bg-card border-r border-border transition-all duration-300 ease-in-out',
        sidebarCollapsed ? 'w-16' : 'w-64'
      )}
    >
      {/* Header */}
      <div className="flex h-16 items-center justify-between px-4 border-b border-border">
        {!sidebarCollapsed && (
          <div className="flex items-center space-x-2">
            <Coffee className="h-8 w-8 text-coffee-500" />
            <span className="text-xl font-bold">Coffee Trading</span>
          </div>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleSidebar}
          className="h-8 w-8"
        >
          <Menu className="h-4 w-4" />
        </Button>
      </div>

      {/* Connection Status */}
      <div className="px-4 py-2 border-b border-border">
        <div className="flex items-center space-x-2">
          <div
            className={cn(
              'h-2 w-2 rounded-full',
              isConnected ? 'bg-green-500' : 'bg-red-500'
            )}
          />
          {!sidebarCollapsed && (
            <span className="text-xs text-muted-foreground">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 p-2">
        {navigation.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                'flex items-center rounded-lg px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground',
                isActive
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground',
                sidebarCollapsed ? 'justify-center' : 'justify-start'
              )}
            >
              <item.icon className="h-5 w-5" />
              {!sidebarCollapsed && (
                <span className="ml-3">{item.name}</span>
              )}
            </Link>
          )
        })}
      </nav>

      {/* Footer */}
      {!sidebarCollapsed && (
        <div className="p-4 border-t border-border">
          <div className="text-xs text-muted-foreground">
            <div>Coffee Trading Dashboard</div>
            <div>v1.0.0</div>
          </div>
        </div>
      )}
    </div>
  )
}
