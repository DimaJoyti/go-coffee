import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { DashboardOverview } from '@/components/dashboard/dashboard-overview'
import { PortfolioSummary } from '@/components/portfolio/portfolio-summary'
import { CoffeeStrategies } from '@/components/trading/coffee-strategies'
import { MarketOverview } from '@/components/market/market-overview'
import { RealtimeUpdates } from '@/components/realtime/realtime-updates'

export default function HomePage() {
  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Real-time Updates Component */}
        <RealtimeUpdates />
        
        {/* Dashboard Overview */}
        <DashboardOverview />
        
        {/* Main Grid Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Portfolio Summary - Takes 2 columns on large screens */}
          <div className="lg:col-span-2">
            <PortfolioSummary />
          </div>
          
          {/* Coffee Strategies - Takes 1 column */}
          <div className="lg:col-span-1">
            <CoffeeStrategies />
          </div>
        </div>
        
        {/* Market Overview - Full width */}
        <MarketOverview />
      </div>
    </DashboardLayout>
  )
}
