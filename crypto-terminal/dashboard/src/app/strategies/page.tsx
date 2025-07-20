import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { CoffeeStrategiesManager } from '@/components/trading/coffee-strategies-manager'
import { RealtimeUpdates } from '@/components/realtime/realtime-updates'

export default function StrategiesPage() {
  return (
    <DashboardLayout>
      <RealtimeUpdates />
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">☕ Coffee Trading Strategies</h1>
            <p className="text-muted-foreground">
              Manage your coffee-themed trading strategies and monitor performance
            </p>
          </div>
        </div>
        
        <CoffeeStrategiesManager />
      </div>
    </DashboardLayout>
  )
}
