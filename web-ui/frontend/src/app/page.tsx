'use client'

import { AIAgents } from '@/components/ai/ai-agents'
import { Analytics } from '@/components/analytics/analytics'
import { CoffeeOrders } from '@/components/coffee/coffee-orders'
import { DashboardOverview } from '@/components/dashboard/dashboard-overview'
import { DefiPortfolio } from '@/components/defi/defi-portfolio'
import { Header } from '@/components/layout/header'
import { Sidebar } from '@/components/layout/sidebar'
import { BrightDataAnalytics } from '@/components/scraping/bright-data-analytics'
import { useWebSocket } from '@/hooks/use-websocket'
import { cn } from '@/lib/utils'
import { motion } from 'framer-motion'
import { useEffect, useState } from 'react'

type ActiveSection = 'dashboard' | 'coffee' | 'defi' | 'agents' | 'scraping' | 'analytics' | 'redis'

export default function HomePage() {
  const [activeSection, setActiveSection] = useState<ActiveSection>('dashboard')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const { isConnected, lastMessage } = useWebSocket()

  // Handle real-time updates
  useEffect(() => {
    if (lastMessage) {
      console.log('Real-time update:', lastMessage)
      // Handle different types of real-time updates
      // This will be expanded based on the WebSocket message types
    }
  }, [lastMessage])

  const renderActiveSection = () => {
    const sectionProps = {
      className: "animate-fade-in"
    }

    switch (activeSection) {
      case 'dashboard':
        return <DashboardOverview {...sectionProps} />
      case 'coffee':
        return <CoffeeOrders {...sectionProps} />
      case 'defi':
        return <DefiPortfolio {...sectionProps} />
      case 'agents':
        return <AIAgents {...sectionProps} />
      case 'scraping':
        return <BrightDataAnalytics {...sectionProps} />
      case 'analytics':
        return <Analytics {...sectionProps} />
      case 'redis':
        return <div className="text-center p-8">Redis Dashboard Coming Soon</div>
      default:
        return <DashboardOverview {...sectionProps} />
    }
  }

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <Sidebar
        activeSection={activeSection}
        onSectionChange={setActiveSection}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
      />

      {/* Main Content */}
      <div className={cn(
        "flex-1 flex flex-col transition-all duration-300",
        sidebarCollapsed ? "ml-16" : "ml-64"
      )}>
        {/* Header */}
        <Header
          activeSection={activeSection}
          isConnected={isConnected}
          onToggleSidebar={() => setSidebarCollapsed(!sidebarCollapsed)}
        />

        {/* Content Area */}
        <main className="flex-1 overflow-auto p-6 bg-muted/30">
          <motion.div
            key={activeSection}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
            className="h-full"
          >
            {renderActiveSection()}
          </motion.div>
        </main>
      </div>

      {/* Connection Status Indicator */}
      <div className={cn(
        "fixed bottom-4 right-4 z-50 px-3 py-2 rounded-full text-xs font-medium transition-all duration-300",
        isConnected
          ? "bg-green-500 text-white"
          : "bg-red-500 text-white animate-pulse"
      )}>
        {isConnected ? "🟢 Connected" : "🔴 Disconnected"}
      </div>
    </div>
  )
}
