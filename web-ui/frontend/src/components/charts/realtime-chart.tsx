'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Activity } from 'lucide-react'

export function RealtimeChart() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Activity className="h-5 w-5" />
          Real-time Activity
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 flex items-center justify-center bg-muted/30 rounded-lg">
          <div className="text-center">
            <Activity className="h-12 w-12 text-muted-foreground mx-auto mb-2 animate-pulse" />
            <p className="text-muted-foreground">Real-time chart will be implemented here</p>
            <p className="text-sm text-muted-foreground mt-1">Using Recharts library</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
