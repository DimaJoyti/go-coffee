import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { MarketAnalysis } from '@/components/market/market-analysis'
import { RealtimeUpdates } from '@/components/realtime/realtime-updates'

export default function MarketsPage() {
  return (
    <DashboardLayout>
      <RealtimeUpdates />
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Market Analysis</h1>
            <p className="text-muted-foreground">
              Real-time cryptocurrency market data and analysis
            </p>
          </div>
        </div>
        
        <MarketAnalysis />
      </div>
    </DashboardLayout>
  )
}
