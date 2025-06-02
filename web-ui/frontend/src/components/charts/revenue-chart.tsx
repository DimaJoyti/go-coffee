'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { DollarSign } from 'lucide-react'

interface RevenueChartProps {
  timeRange: string
}

export function RevenueChart({ timeRange }: RevenueChartProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <DollarSign className="h-5 w-5" />
          Revenue Analysis ({timeRange})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 flex items-center justify-center bg-muted/30 rounded-lg">
          <div className="text-center">
            <DollarSign className="h-12 w-12 text-green-500 mx-auto mb-2" />
            <p className="text-muted-foreground">Revenue chart will be implemented here</p>
            <p className="text-sm text-muted-foreground mt-1">Time range: {timeRange}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
