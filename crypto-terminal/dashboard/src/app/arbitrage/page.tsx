import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { ArbitrageDashboard } from '@/components/arbitrage/arbitrage-dashboard'
import { RealtimeUpdates } from '@/components/realtime/realtime-updates'

export default function ArbitragePage() {
  return (
    <DashboardLayout>
      <RealtimeUpdates />
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Arbitrage Opportunities</h1>
            <p className="text-muted-foreground">
              Discover and track cross-exchange arbitrage opportunities
            </p>
          </div>
        </div>
        
        <ArbitrageDashboard />
      </div>
    </DashboardLayout>
  )
}
