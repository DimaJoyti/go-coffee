'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { BarChart3 } from 'lucide-react'

interface SalesChartProps {
  timeRange: string
}

export function SalesChart({ timeRange }: SalesChartProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <BarChart3 className="h-5 w-5" />
          Sales Trends ({timeRange})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 flex items-center justify-center bg-muted/30 rounded-lg">
          <div className="text-center">
            <BarChart3 className="h-12 w-12 text-blue-500 mx-auto mb-2" />
            <p className="text-muted-foreground">Sales chart will be implemented here</p>
            <p className="text-sm text-muted-foreground mt-1">Time range: {timeRange}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
