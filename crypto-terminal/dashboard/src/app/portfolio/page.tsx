import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { PortfolioManager } from '@/components/portfolio/portfolio-manager'
import { RealtimeUpdates } from '@/components/realtime/realtime-updates'

export default function PortfolioPage() {
  return (
    <DashboardLayout>
      <RealtimeUpdates />
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Portfolio Management</h1>
            <p className="text-muted-foreground">
              Manage your cryptocurrency portfolios and track performance
            </p>
          </div>
        </div>
        
        <PortfolioManager />
      </div>
    </DashboardLayout>
  )
}
