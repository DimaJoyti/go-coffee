'use client'

import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'
import { useRealTimeData, usePrices, usePortfolio, useConnectionStatus } from '@/contexts/RealTimeDataContext'

export default function DashboardOverview() {
  const { prices, isConnected } = useRealTimeData()
  const portfolio = usePortfolio()
  const { prices: priceList } = usePrices()
  const { connectionStatus } = useConnectionStatus()

  const [activeOrders] = useState(12)
  const [aiAgents] = useState(8)

  // Debug logging
  useEffect(() => {
    console.log('Dashboard Debug:', {
      isConnected,
      connectionStatus,
      portfolio,
      pricesCount: prices.size,
      priceListLength: priceList.length
    })
  }, [isConnected, connectionStatus, portfolio, prices, priceList])

  // Calculate stats from real-time data
  const stats = {
    totalValue: portfolio?.totalValue || 123456.78,
    dailyChange: portfolio?.change24h || 5.67,
    activeOrders,
    aiAgents
  }

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
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.5 }
    }
  }

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="space-y-6"
    >
      {/* Enhanced Welcome Section */}
      <motion.div variants={itemVariants} className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-4xl font-bold gradient-text mb-3 font-display">
              Welcome back! 👋
            </h2>
            <p className="text-slate-400 text-lg">
              Here's what's happening with your <span className="gradient-text-crypto font-semibold">coffee trading ecosystem</span> today.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-status-success animate-pulse' : 'bg-status-error'}`} />
            <span className="text-sm text-slate-400">
              {connectionStatus === 'connected' ? 'Live Data' : connectionStatus}
            </span>
          </div>
        </div>
      </motion.div>

      {/* Debug Panel - Remove this in production */}
      {process.env.NODE_ENV === 'development' && (
        <motion.div variants={itemVariants} className="mb-6">
          <div className="glass-card p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-lg font-semibold text-coffee-400">Debug Info</h3>
              <button
                onClick={() => window.location.reload()}
                className="px-3 py-1 bg-coffee-500/20 text-coffee-400 rounded-lg text-sm hover:bg-coffee-500/30 transition-colors"
              >
                Refresh Data
              </button>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm mb-3">
              <div>
                <span className="text-slate-400">Connection:</span>
                <span className={`ml-2 ${isConnected ? 'text-status-success' : 'text-status-error'}`}>
                  {connectionStatus}
                </span>
              </div>
              <div>
                <span className="text-slate-400">Prices:</span>
                <span className="ml-2 text-white">{prices.size} symbols</span>
              </div>
              <div>
                <span className="text-slate-400">Portfolio:</span>
                <span className="ml-2 text-white">{portfolio ? `$${portfolio.totalValue.toLocaleString()}` : 'Loading...'}</span>
              </div>
              <div>
                <span className="text-slate-400">Price List:</span>
                <span className="ml-2 text-white">{priceList.length} items</span>
              </div>
            </div>
            {priceList.length > 0 && (
              <div className="text-xs text-slate-500">
                Sample prices: {priceList.slice(0, 2).map(p => `${p.symbol}: $${p.price.toFixed(2)}`).join(', ')}
              </div>
            )}
          </div>
        </motion.div>
      )}

      {/* Stats Cards */}
      <motion.div
        variants={containerVariants}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      >
        <StatCard
          title="Total Portfolio Value"
          value={`$${stats.totalValue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`}
          change={`+${stats.dailyChange.toFixed(2)}%`}
          icon="💰"
          trend="up"
        />
        <StatCard
          title="Active Coffee Orders"
          value={stats.activeOrders.toString()}
          change="+3 today"
          icon="☕"
          trend="up"
        />
        <StatCard
          title="AI Agents Running"
          value={stats.aiAgents.toString()}
          change="All systems operational"
          icon="🤖"
          trend="neutral"
        />
        <StatCard
          title="Trading Volume (24h)"
          value="$45,678"
          change="+12.3%"
          icon="📈"
          trend="up"
        />
      </motion.div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Trading Chart */}
        <motion.div
          variants={itemVariants}
          className="lg:col-span-2 feature-card"
        >
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-xl font-semibold gradient-text">Portfolio Performance</h3>
            <div className="flex items-center space-x-2">
              <button className="px-3 py-1 bg-coffee-500/20 text-coffee-400 rounded-lg text-sm hover:bg-coffee-500/30 transition-colors">24H</button>
              <button className="px-3 py-1 text-slate-400 hover:text-coffee-300 rounded-lg text-sm hover:bg-coffee-500/10 transition-colors">7D</button>
              <button className="px-3 py-1 text-slate-400 hover:text-coffee-300 rounded-lg text-sm hover:bg-coffee-500/10 transition-colors">30D</button>
            </div>
          </div>
          
          {/* Enhanced Chart Placeholder */}
          <div className="h-64 bg-gradient-to-br from-slate-700/30 to-slate-800/30 rounded-xl flex items-center justify-center relative overflow-hidden hover-lift">
            <div className="absolute inset-0 bg-gradient-to-r from-coffee-500/10 to-brand-amber/10" />
            <div className="absolute inset-0 bg-noise opacity-5" />
            <div className="text-center relative z-10">
              <div className="text-4xl mb-2 animate-bounce-gentle">📊</div>
              <p className="text-slate-300 font-medium">Interactive chart will be loaded here</p>
              <p className="text-slate-500 text-sm mt-1">Real-time trading data visualization</p>
            </div>
          </div>
        </motion.div>

        {/* Live Prices */}
        <motion.div
          variants={itemVariants}
          className="feature-card"
        >
          <h3 className="text-xl font-semibold gradient-text mb-4">Live Prices</h3>
          <div className="space-y-3">
            {priceList.slice(0, 4).map((price) => (
              <div key={price.symbol} className="flex items-center justify-between p-3 bg-slate-700/30 rounded-lg hover-lift">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-gradient-to-r from-coffee-500 to-brand-amber rounded-full flex items-center justify-center text-white text-sm font-bold">
                    {price.symbol.split('/')[0].charAt(0)}
                  </div>
                  <span className="text-white font-medium">{price.symbol}</span>
                </div>
                <div className="text-right">
                  <div className="text-white font-semibold">
                    ${price.price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </div>
                  <div className={`text-sm ${price.change24h >= 0 ? 'text-status-success' : 'text-status-error'}`}>
                    {price.change24h >= 0 ? '+' : ''}{price.change24h.toFixed(2)}%
                  </div>
                </div>
              </div>
            ))}
            {priceList.length === 0 && (
              <div className="text-center py-8 text-slate-400">
                <div className="animate-pulse">Loading price data...</div>
              </div>
            )}
          </div>
        </motion.div>

        {/* Quick Actions */}
        <motion.div
          variants={itemVariants}
          className="space-y-6"
        >
          {/* Recent Activity */}
          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
            <h3 className="text-lg font-semibold text-white mb-4">Recent Activity</h3>
            <div className="space-y-4">
              {[
                { type: 'trade', message: 'Bought 0.5 BTC', time: '2m ago', icon: '📈' },
                { type: 'coffee', message: 'New coffee order #1234', time: '5m ago', icon: '☕' },
                { type: 'ai', message: 'AI agent optimized portfolio', time: '10m ago', icon: '🤖' },
                { type: 'trade', message: 'Sold 100 SOL', time: '15m ago', icon: '📉' }
              ].map((activity, index) => (
                <div key={index} className="flex items-center space-x-3 p-3 bg-slate-700/30 rounded-lg">
                  <div className="text-xl">{activity.icon}</div>
                  <div className="flex-1">
                    <p className="text-white text-sm">{activity.message}</p>
                    <p className="text-slate-400 text-xs">{activity.time}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Quick Actions */}
          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
            <h3 className="text-lg font-semibold text-white mb-4">Quick Actions</h3>
            <div className="space-y-3">
              {[
                { label: 'New Trade', icon: '💱', color: 'from-green-500 to-emerald-500' },
                { label: 'Order Coffee', icon: '☕', color: 'from-amber-500 to-orange-500' },
                { label: 'Deploy AI Agent', icon: '🤖', color: 'from-blue-500 to-purple-500' },
                { label: 'View Analytics', icon: '📊', color: 'from-purple-500 to-pink-500' }
              ].map((action, index) => (
                <motion.button
                  key={action.label}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  className={`w-full flex items-center space-x-3 p-3 bg-gradient-to-r ${action.color} bg-opacity-20 border border-opacity-30 rounded-lg hover:bg-opacity-30 transition-all duration-200`}
                >
                  <span className="text-xl">{action.icon}</span>
                  <span className="text-white font-medium">{action.label}</span>
                </motion.button>
              ))}
            </div>
          </div>
        </motion.div>
      </div>

      {/* Bottom Section */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Market Overview */}
        <motion.div
          variants={itemVariants}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-lg font-semibold text-white mb-4">Market Overview</h3>
          <div className="space-y-3">
            {Array.from(prices.values()).map((priceData) => (
              <div key={priceData.symbol} className="flex items-center justify-between p-3 bg-slate-700/30 rounded-lg">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-gradient-to-r from-amber-500 to-orange-500 rounded-full flex items-center justify-center text-white text-sm font-bold">
                    {priceData.symbol.split('/')[0].charAt(0)}
                  </div>
                  <div>
                    <p className="text-white font-medium">{priceData.symbol.split('/')[0]}</p>
                    <p className="text-slate-400 text-sm">${priceData.price.toLocaleString()}</p>
                  </div>
                </div>
                <div className={`text-sm font-medium ${
                  priceData.change24h >= 0 ? 'text-green-400' : 'text-red-400'
                }`}>
                  {priceData.change24h >= 0 ? '+' : ''}{priceData.change24h.toFixed(2)}%
                </div>
              </div>
            ))}
          </div>
        </motion.div>

        {/* System Status */}
        <motion.div
          variants={itemVariants}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-lg font-semibold text-white mb-4">System Status</h3>
          <div className="space-y-4">
            {[
              { service: 'Trading Engine', status: 'operational', uptime: '99.9%' },
              { service: 'Coffee API', status: 'operational', uptime: '99.8%' },
              { service: 'AI Agents', status: 'operational', uptime: '99.7%' },
              { service: 'Data Feed', status: 'operational', uptime: '99.9%' }
            ].map((service) => (
              <div key={service.service} className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                  <span className="text-white">{service.service}</span>
                </div>
                <div className="text-right">
                  <p className="text-green-400 text-sm font-medium capitalize">{service.status}</p>
                  <p className="text-slate-400 text-xs">{service.uptime} uptime</p>
                </div>
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </motion.div>
  )
}

interface StatCardProps {
  title: string
  value: string
  change: string
  icon: string
  trend: 'up' | 'down' | 'neutral'
}

function StatCard({ title, value, change, icon, trend }: StatCardProps) {
  const getTrendColor = () => {
    if (trend === 'up') return 'text-status-success'
    if (trend === 'down') return 'text-status-error'
    return 'text-slate-400'
  }

  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      className="metric-card group"
    >
      <div className="flex items-center justify-between mb-4">
        <div className="text-3xl animate-bounce-gentle group-hover:scale-110 transition-transform duration-300">{icon}</div>
        <div className={`text-sm font-medium ${getTrendColor()}`}>
          {change}
        </div>
      </div>
      <div>
        <p className="text-2xl font-bold gradient-text mb-1">{value}</p>
        <p className="text-slate-400 text-sm font-medium">{title}</p>
      </div>
    </motion.div>
  )
}
