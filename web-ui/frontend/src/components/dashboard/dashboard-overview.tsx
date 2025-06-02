'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { 
  Coffee, 
  DollarSign, 
  TrendingUp, 
  Users, 
  Bot,
  Activity,
  Zap,
  Globe
} from 'lucide-react'
import { MetricCard } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { RealtimeChart } from '@/components/charts/realtime-chart'
import { OrdersChart } from '@/components/charts/orders-chart'
import { PortfolioChart } from '@/components/charts/portfolio-chart'
import { formatCurrency, formatNumber } from '@/lib/utils'

interface DashboardMetrics {
  totalOrders: number
  totalRevenue: number
  portfolioValue: number
  activeAgents: number
  ordersChange: number
  revenueChange: number
  portfolioChange: number
  agentsChange: number
}

interface DashboardOverviewProps {
  className?: string
}

export function DashboardOverview({ className }: DashboardOverviewProps) {
  const [metrics, setMetrics] = useState<DashboardMetrics>({
    totalOrders: 1247,
    totalRevenue: 45678.90,
    portfolioValue: 123456.78,
    activeAgents: 9,
    ordersChange: 12.5,
    revenueChange: 8.3,
    portfolioChange: 15.7,
    agentsChange: 0
  })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Simulate loading
    const timer = setTimeout(() => {
      setLoading(false)
    }, 1000)

    return () => clearTimeout(timer)
  }, [])

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1
      }
    }
  }

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0 }
  }

  return (
    <motion.div
      className={className}
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Welcome back! ☕</h1>
            <p className="text-muted-foreground mt-2">
              Here's what's happening with your Go Coffee ecosystem today.
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline">
              <Activity className="h-4 w-4 mr-2" />
              View Details
            </Button>
            <Button>
              <Zap className="h-4 w-4 mr-2" />
              Quick Actions
            </Button>
          </div>
        </div>
      </motion.div>

      {/* Metrics Grid */}
      <motion.div 
        variants={itemVariants}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8"
      >
        <MetricCard
          title="Total Orders"
          value={formatNumber(metrics.totalOrders)}
          change={metrics.ordersChange}
          changeLabel="vs last month"
          icon={<Coffee className="h-6 w-6" />}
          loading={loading}
        />
        
        <MetricCard
          title="Revenue"
          value={formatCurrency(metrics.totalRevenue)}
          change={metrics.revenueChange}
          changeLabel="vs last month"
          icon={<DollarSign className="h-6 w-6" />}
          loading={loading}
        />
        
        <MetricCard
          title="Portfolio Value"
          value={formatCurrency(metrics.portfolioValue)}
          change={metrics.portfolioChange}
          changeLabel="24h change"
          icon={<TrendingUp className="h-6 w-6" />}
          loading={loading}
        />
        
        <MetricCard
          title="Active AI Agents"
          value={metrics.activeAgents}
          change={metrics.agentsChange}
          changeLabel="all operational"
          icon={<Bot className="h-6 w-6" />}
          loading={loading}
        />
      </motion.div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <motion.div variants={itemVariants}>
          <RealtimeChart />
        </motion.div>
        
        <motion.div variants={itemVariants}>
          <OrdersChart />
        </motion.div>
      </div>

      {/* Portfolio and Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <motion.div variants={itemVariants} className="lg:col-span-2">
          <PortfolioChart />
        </motion.div>
        
        <motion.div variants={itemVariants}>
          <ActivityFeed />
        </motion.div>
      </div>
    </motion.div>
  )
}

function ActivityFeed() {
  const activities = [
    {
      id: 1,
      type: 'order',
      message: 'New coffee order #1247',
      time: '2 minutes ago',
      icon: Coffee,
      color: 'text-coffee-500'
    },
    {
      id: 2,
      type: 'trade',
      message: 'DeFi arbitrage executed',
      time: '5 minutes ago',
      icon: TrendingUp,
      color: 'text-green-500'
    },
    {
      id: 3,
      type: 'agent',
      message: 'Inventory agent updated stock',
      time: '10 minutes ago',
      icon: Bot,
      color: 'text-blue-500'
    },
    {
      id: 4,
      type: 'user',
      message: 'New customer registered',
      time: '15 minutes ago',
      icon: Users,
      color: 'text-purple-500'
    },
    {
      id: 5,
      type: 'system',
      message: 'Market data updated',
      time: '20 minutes ago',
      icon: Globe,
      color: 'text-orange-500'
    }
  ]

  return (
    <div className="bg-card border border-border rounded-lg p-6">
      <h3 className="text-lg font-semibold mb-4">Recent Activity</h3>
      <div className="space-y-4">
        {activities.map((activity) => {
          const Icon = activity.icon
          return (
            <motion.div
              key={activity.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: activity.id * 0.1 }}
              className="flex items-center gap-3 p-3 rounded-lg hover:bg-muted/50 transition-colors"
            >
              <div className={`p-2 rounded-full bg-muted ${activity.color}`}>
                <Icon className="h-4 w-4" />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium">{activity.message}</p>
                <p className="text-xs text-muted-foreground">{activity.time}</p>
              </div>
            </motion.div>
          )
        })}
      </div>
      <Button variant="ghost" className="w-full mt-4">
        View All Activity
      </Button>
    </div>
  )
}
