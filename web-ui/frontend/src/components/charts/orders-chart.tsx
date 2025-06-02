'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Coffee } from 'lucide-react'

export function OrdersChart() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Coffee className="h-5 w-5" />
          Orders Overview
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 flex items-center justify-center bg-muted/30 rounded-lg">
          <div className="text-center">
            <Coffee className="h-12 w-12 text-coffee-500 mx-auto mb-2" />
            <p className="text-muted-foreground">Orders chart will be implemented here</p>
            <p className="text-sm text-muted-foreground mt-1">Showing daily order trends</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
