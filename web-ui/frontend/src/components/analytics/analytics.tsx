'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  BarChart3, 
  Download, 
  Calendar,
  TrendingUp,
  Users,
  Coffee,
  DollarSign
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { SalesChart } from '@/components/charts/sales-chart'
import { RevenueChart } from '@/components/charts/revenue-chart'
import { formatCurrency, formatNumber } from '@/lib/utils'

interface AnalyticsProps {
  className?: string
}

export function Analytics({ className }: AnalyticsProps) {
  const [timeRange, setTimeRange] = useState<'7d' | '30d' | '90d' | '1y'>('30d')

  const metrics = {
    totalRevenue: 125678.90,
    totalOrders: 3456,
    avgOrderValue: 36.34,
    customerGrowth: 12.5,
    revenueGrowth: 18.7,
    orderGrowth: 15.2
  }

  const topProducts = [
    { name: 'Large Latte', sales: 1234, revenue: 6170 },
    { name: 'Cappuccino', sales: 987, revenue: 4935 },
    { name: 'Americano', sales: 856, revenue: 3424 },
    { name: 'Espresso', sales: 743, revenue: 2229 },
    { name: 'Mocha', sales: 654, revenue: 3924 }
  ]

  const locations = [
    { name: 'Downtown', revenue: 45678, orders: 1234, growth: 15.2 },
    { name: 'Mall', revenue: 38945, orders: 987, growth: 12.8 },
    { name: 'Airport', revenue: 28734, orders: 743, growth: 8.9 },
    { name: 'University', revenue: 12317, orders: 492, growth: 22.1 }
  ]

  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Analytics & Reports</h1>
          <p className="text-muted-foreground">
            Comprehensive business insights and performance metrics
          </p>
        </div>
        <div className="flex gap-2">
          <div className="flex border border-border rounded-lg">
            {['7d', '30d', '90d', '1y'].map((range) => (
              <Button
                key={range}
                variant={timeRange === range ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setTimeRange(range as any)}
                className="rounded-none first:rounded-l-lg last:rounded-r-lg"
              >
                {range}
              </Button>
            ))}
          </div>
          <Button>
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-2 mb-2">
              <DollarSign className="h-5 w-5 text-green-500" />
              <span className="font-medium">Total Revenue</span>
            </div>
            <div className="text-2xl font-bold">{formatCurrency(metrics.totalRevenue)}</div>
            <div className="text-sm text-green-600">
              +{metrics.revenueGrowth}% vs last period
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-2 mb-2">
              <Coffee className="h-5 w-5 text-coffee-500" />
              <span className="font-medium">Total Orders</span>
            </div>
            <div className="text-2xl font-bold">{formatNumber(metrics.totalOrders)}</div>
            <div className="text-sm text-green-600">
              +{metrics.orderGrowth}% vs last period
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="h-5 w-5 text-blue-500" />
              <span className="font-medium">Avg Order Value</span>
            </div>
            <div className="text-2xl font-bold">{formatCurrency(metrics.avgOrderValue)}</div>
            <div className="text-sm text-green-600">
              +{metrics.customerGrowth}% vs last period
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <SalesChart timeRange={timeRange} />
        <RevenueChart timeRange={timeRange} />
      </div>

      {/* Detailed Analytics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Products */}
        <Card>
          <CardHeader>
            <CardTitle>Top Products</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topProducts.map((product, index) => (
                <motion.div
                  key={product.name}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="flex items-center justify-between p-3 rounded-lg hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-coffee-100 rounded-full flex items-center justify-center text-coffee-600 font-bold text-sm">
                      {index + 1}
                    </div>
                    <div>
                      <div className="font-medium">{product.name}</div>
                      <div className="text-sm text-muted-foreground">
                        {product.sales} sales
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-semibold">{formatCurrency(product.revenue)}</div>
                    <div className="text-sm text-muted-foreground">revenue</div>
                  </div>
                </motion.div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Location Performance */}
        <Card>
          <CardHeader>
            <CardTitle>Location Performance</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {locations.map((location, index) => (
                <motion.div
                  key={location.name}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="p-3 rounded-lg border border-border"
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="font-medium">{location.name}</div>
                    <div className="text-sm text-green-600">
                      +{location.growth}%
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <div className="text-muted-foreground">Revenue</div>
                      <div className="font-medium">{formatCurrency(location.revenue)}</div>
                    </div>
                    <div>
                      <div className="text-muted-foreground">Orders</div>
                      <div className="font-medium">{location.orders}</div>
                    </div>
                  </div>
                  
                  <div className="mt-2">
                    <div className="w-full bg-muted rounded-full h-2">
                      <div 
                        className="h-2 bg-coffee-500 rounded-full transition-all duration-300"
                        style={{ width: `${(location.revenue / 50000) * 100}%` }}
                      />
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </motion.div>
  )
}
