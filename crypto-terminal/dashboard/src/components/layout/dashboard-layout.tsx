'use client'

import { cn } from '@/lib/utils'
import { Sidebar } from './sidebar'
import { Header } from './header'
import { useTradingStore } from '@/stores/trading-store'

interface DashboardLayoutProps {
  children: React.ReactNode
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  const { sidebarCollapsed } = useTradingStore()

  return (
    <div className="min-h-screen bg-background">
      {/* Sidebar */}
      <Sidebar />
      
      {/* Main Content */}
      <div
        className={cn(
          'transition-all duration-300 ease-in-out',
          sidebarCollapsed ? 'ml-16' : 'ml-64'
        )}
      >
        {/* Header */}
        <Header />
        
        {/* Page Content */}
        <main className="p-6">
          {children}
        </main>
      </div>
    </div>
  )
}
