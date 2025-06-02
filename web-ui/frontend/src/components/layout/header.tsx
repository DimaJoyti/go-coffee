'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Bell, 
  Settings, 
  User, 
  Moon, 
  Sun, 
  Menu,
  Wifi,
  WifiOff
} from 'lucide-react'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

type ActiveSection = 'dashboard' | 'coffee' | 'defi' | 'agents' | 'scraping' | 'analytics'

interface HeaderProps {
  activeSection: ActiveSection
  isConnected: boolean
  onToggleSidebar: () => void
}

const sectionTitles = {
  dashboard: 'Dashboard Overview',
  coffee: 'Coffee Orders & Inventory',
  defi: 'DeFi Portfolio & Trading',
  agents: 'AI Agents Monitoring',
  scraping: 'Market Data & Analytics',
  analytics: 'Reports & Analytics'
}

export function Header({ activeSection, isConnected, onToggleSidebar }: HeaderProps) {
  const { theme, setTheme } = useTheme()
  const [notificationsCount] = useState(3)

  return (
    <header className="h-16 bg-background border-b border-border px-6 flex items-center justify-between">
      {/* Left Section */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={onToggleSidebar}
          className="md:hidden"
        >
          <Menu className="h-5 w-5" />
        </Button>
        
        <div>
          <h1 className="text-xl font-semibold">
            {sectionTitles[activeSection]}
          </h1>
          <p className="text-sm text-muted-foreground">
            {new Date().toLocaleDateString('en-US', { 
              weekday: 'long', 
              year: 'numeric', 
              month: 'long', 
              day: 'numeric' 
            })}
          </p>
        </div>
      </div>

      {/* Right Section */}
      <div className="flex items-center gap-2">
        {/* Connection Status */}
        <motion.div
          initial={{ scale: 0.9 }}
          animate={{ scale: 1 }}
          className={cn(
            "flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium",
            isConnected 
              ? "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300" 
              : "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300"
          )}
        >
          {isConnected ? (
            <Wifi className="h-3 w-3" />
          ) : (
            <WifiOff className="h-3 w-3" />
          )}
          {isConnected ? 'Connected' : 'Disconnected'}
        </motion.div>

        {/* Theme Toggle */}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          className="h-9 w-9"
        >
          <Sun className="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
          <span className="sr-only">Toggle theme</span>
        </Button>

        {/* Notifications */}
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9 relative"
        >
          <Bell className="h-4 w-4" />
          {notificationsCount > 0 && (
            <motion.span
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              className="absolute -top-1 -right-1 h-5 w-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center"
            >
              {notificationsCount}
            </motion.span>
          )}
        </Button>

        {/* Settings */}
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9"
        >
          <Settings className="h-4 w-4" />
        </Button>

        {/* User Profile */}
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9"
        >
          <User className="h-4 w-4" />
        </Button>
      </div>
    </header>
  )
}
