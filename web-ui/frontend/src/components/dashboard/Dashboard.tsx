'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'
import Sidebar from './Sidebar'
import Header from './Header'
import DashboardOverview from './DashboardOverview'
import TradingView from './TradingView'
import CoffeeOrders from './CoffeeOrders'
import AIAgents from './AIAgents'
import Portfolio from './Portfolio'
import Analytics from './Analytics'
import NotificationSystem from '@/components/ui/NotificationSystem'

interface DashboardProps {
  onBackToLanding: () => void
}

export type DashboardView = 'overview' | 'trading' | 'coffee' | 'ai' | 'portfolio' | 'analytics'

export default function Dashboard({ onBackToLanding }: DashboardProps) {
  const [currentView, setCurrentView] = useState<DashboardView>('overview')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  const renderCurrentView = () => {
    switch (currentView) {
      case 'overview':
        return <DashboardOverview />
      case 'trading':
        return <TradingView />
      case 'coffee':
        return <CoffeeOrders />
      case 'ai':
        return <AIAgents />
      case 'portfolio':
        return <Portfolio />
      case 'analytics':
        return <Analytics />
      default:
        return <DashboardOverview />
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white flex">
      {/* Sidebar */}
      <Sidebar
        currentView={currentView}
        onViewChange={setCurrentView}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        onBackToLanding={onBackToLanding}
      />

      {/* Main Content */}
      <div className={`flex-1 flex flex-col transition-all duration-300 ${
        sidebarCollapsed ? 'md:ml-16' : 'md:ml-64'
      } ml-0`}>
        {/* Header */}
        <Header
          currentView={currentView}
          onToggleSidebar={() => setSidebarCollapsed(!sidebarCollapsed)}
        />

        {/* Content Area */}
        <main className="flex-1 p-6 overflow-auto">
          <motion.div
            key={currentView}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
            className="h-full"
          >
            {renderCurrentView()}
          </motion.div>
        </main>
      </div>

      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none z-0">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-amber-500/5 rounded-full blur-3xl animate-pulse" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-orange-500/5 rounded-full blur-3xl animate-pulse delay-1000" />
      </div>

      {/* Notification System */}
      <NotificationSystem position="top-right" maxNotifications={5} />
    </div>
  )
}
