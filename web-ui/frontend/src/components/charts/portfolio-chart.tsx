'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { TrendingUp } from 'lucide-react'

export function PortfolioChart() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TrendingUp className="h-5 w-5" />
          Portfolio Performance
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 flex items-center justify-center bg-muted/30 rounded-lg">
          <div className="text-center">
            <TrendingUp className="h-12 w-12 text-green-500 mx-auto mb-2" />
            <p className="text-muted-foreground">Portfolio chart will be implemented here</p>
            <p className="text-sm text-muted-foreground mt-1">DeFi portfolio performance over time</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
