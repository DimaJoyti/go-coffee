'use client'

import { motion } from 'framer-motion'
import { 
  LayoutDashboard, 
  Coffee, 
  Coins, 
  Bot, 
  Search, 
  BarChart3,
  ChevronLeft,
  ChevronRight
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

type ActiveSection = 'dashboard' | 'coffee' | 'defi' | 'agents' | 'scraping' | 'analytics'

interface SidebarProps {
  activeSection: ActiveSection
  onSectionChange: (section: ActiveSection) => void
  collapsed: boolean
  onToggleCollapse: () => void
}

const sidebarItems = [
  {
    id: 'dashboard' as ActiveSection,
    label: 'Dashboard',
    icon: LayoutDashboard,
    description: 'Overview & Metrics'
  },
  {
    id: 'coffee' as ActiveSection,
    label: 'Coffee Orders',
    icon: Coffee,
    description: 'Orders & Inventory'
  },
  {
    id: 'defi' as ActiveSection,
    label: 'DeFi Portfolio',
    icon: Coins,
    description: 'Crypto Trading'
  },
  {
    id: 'agents' as ActiveSection,
    label: 'AI Agents',
    icon: Bot,
    description: 'Agent Monitoring'
  },
  {
    id: 'scraping' as ActiveSection,
    label: 'Market Data',
    icon: Search,
    description: 'Bright Data Analytics'
  },
  {
    id: 'analytics' as ActiveSection,
    label: 'Analytics',
    icon: BarChart3,
    description: 'Reports & Insights'
  }
]

export function Sidebar({ 
  activeSection, 
  onSectionChange, 
  collapsed, 
  onToggleCollapse 
}: SidebarProps) {
  return (
    <motion.div
      initial={false}
      animate={{ width: collapsed ? 64 : 256 }}
      transition={{ duration: 0.3, ease: "easeInOut" }}
      className="fixed left-0 top-0 h-full bg-card border-r border-border shadow-lg z-40"
    >
      <div className="flex flex-col h-full">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          {!collapsed && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="flex items-center gap-2"
            >
              <div className="w-8 h-8 bg-coffee-500 rounded-lg flex items-center justify-center">
                <Coffee className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="font-bold text-lg">Go Coffee</h1>
                <p className="text-xs text-muted-foreground">Epic UI</p>
              </div>
            </motion.div>
          )}
          
          <Button
            variant="ghost"
            size="icon"
            onClick={onToggleCollapse}
            className="h-8 w-8"
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-2">
          {sidebarItems.map((item) => {
            const Icon = item.icon
            const isActive = activeSection === item.id
            
            return (
              <motion.button
                key={item.id}
                onClick={() => onSectionChange(item.id)}
                className={cn(
                  "w-full flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium transition-all duration-200",
                  "hover:bg-accent hover:text-accent-foreground",
                  isActive && "bg-primary text-primary-foreground shadow-md",
                  collapsed && "justify-center px-2"
                )}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
              >
                <Icon className={cn(
                  "h-5 w-5 flex-shrink-0",
                  isActive && "text-primary-foreground"
                )} />
                
                {!collapsed && (
                  <motion.div
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: -10 }}
                    transition={{ duration: 0.2 }}
                    className="flex-1 text-left"
                  >
                    <div className="font-medium">{item.label}</div>
                    <div className="text-xs opacity-70">{item.description}</div>
                  </motion.div>
                )}
              </motion.button>
            )
          })}
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-border">
          {!collapsed ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="text-center"
            >
              <p className="text-xs text-muted-foreground">
                Web3 Coffee Ecosystem
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                v1.0.0
              </p>
            </motion.div>
          ) : (
            <div className="flex justify-center">
              <div className="w-2 h-2 bg-coffee-500 rounded-full" />
            </div>
          )}
        </div>
      </div>
    </motion.div>
  )
}
