'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'

export default function Analytics() {
  const [selectedMetric, setSelectedMetric] = useState<'trading' | 'coffee' | 'ai' | 'overall'>('overall')
  const [timeRange, setTimeRange] = useState<'24h' | '7d' | '30d' | '90d'>('30d')

  const metrics = {
    overall: {
      totalRevenue: 245678.90,
      totalTrades: 1247,
      successRate: 94.2,
      avgProfit: 197.34
    },
    trading: {
      volume: 1234567.89,
      trades: 892,
      winRate: 87.5,
      avgReturn: 2.34
    },
    coffee: {
      orders: 456,
      revenue: 12345.67,
      avgOrderValue: 27.05,
      satisfaction: 4.8
    },
    ai: {
      agents: 8,
      efficiency: 96.7,
      savings: 5678.90,
      uptime: 99.9
    }
  }

  const topPerformers = [
    { name: 'BTC/USDT', profit: '+$12,345', trades: 89, winRate: 94.2 },
    { name: 'ETH/USDT', profit: '+$8,901', trades: 67, winRate: 89.5 },
    { name: 'SOL/USDT', profit: '+$5,432', trades: 45, winRate: 87.8 },
    { name: 'COFFEE/USDT', profit: '+$3,210', trades: 23, winRate: 91.3 }
  ]

  const recentInsights = [
    {
      type: 'trading',
      title: 'BTC Bullish Trend Detected',
      description: 'AI analysis shows strong buying pressure in BTC with 85% confidence',
      time: '2 hours ago',
      impact: 'high',
      icon: '📈'
    },
    {
      type: 'coffee',
      title: 'Ethiopian Coffee Demand Surge',
      description: 'Orders for Ethiopian Yirgacheffe increased by 45% this week',
      time: '4 hours ago',
      impact: 'medium',
      icon: '☕'
    },
    {
      type: 'ai',
      title: 'Risk Manager Optimization',
      description: 'AI agent reduced portfolio risk by 12% while maintaining returns',
      time: '6 hours ago',
      impact: 'high',
      icon: '🤖'
    },
    {
      type: 'market',
      title: 'Arbitrage Opportunity Alert',
      description: 'Cross-exchange price difference detected for SOL (2.3% spread)',
      time: '8 hours ago',
      impact: 'medium',
      icon: '🎯'
    }
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0"
      >
        <div>
          <h2 className="text-3xl font-bold text-white">Analytics & Reports</h2>
          <p className="text-slate-400">Comprehensive insights into your trading ecosystem</p>
        </div>
        
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            {(['24h', '7d', '30d', '90d'] as const).map((range) => (
              <button
                key={range}
                onClick={() => setTimeRange(range)}
                className={`px-4 py-2 rounded-lg font-medium transition-all duration-200 ${
                  timeRange === range
                    ? 'bg-amber-500 text-white'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                }`}
              >
                {range}
              </button>
            ))}
          </div>
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="px-6 py-2 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded-lg font-semibold hover:from-blue-600 hover:to-purple-600 transition-all duration-200"
          >
            Export Report
          </motion.button>
        </div>
      </motion.div>

      {/* Metric Selector */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex space-x-1 bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-2"
      >
        {[
          { id: 'overall', label: 'Overall', icon: '📊' },
          { id: 'trading', label: 'Trading', icon: '📈' },
          { id: 'coffee', label: 'Coffee', icon: '☕' },
          { id: 'ai', label: 'AI Agents', icon: '🤖' }
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setSelectedMetric(tab.id as any)}
            className={`flex-1 flex items-center justify-center space-x-2 px-6 py-3 rounded-xl font-semibold transition-all duration-300 ${
              selectedMetric === tab.id
                ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-lg'
                : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
            }`}
          >
            <span>{tab.icon}</span>
            <span>{tab.label}</span>
          </button>
        ))}
      </motion.div>

      {/* Key Metrics */}
      <motion.div
        key={selectedMetric}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      >
        {selectedMetric === 'overall' && (
          <>
            <MetricCard title="Total Revenue" value={`$${metrics.overall.totalRevenue.toLocaleString()}`} change="+12.5%" icon="💰" />
            <MetricCard title="Total Trades" value={metrics.overall.totalTrades.toString()} change="+8.3%" icon="📊" />
            <MetricCard title="Success Rate" value={`${metrics.overall.successRate}%`} change="+2.1%" icon="🎯" />
            <MetricCard title="Avg Profit" value={`$${metrics.overall.avgProfit}`} change="+15.7%" icon="📈" />
          </>
        )}
        {selectedMetric === 'trading' && (
          <>
            <MetricCard title="Trading Volume" value={`$${(metrics.trading.volume / 1000000).toFixed(1)}M`} change="+18.2%" icon="💱" />
            <MetricCard title="Total Trades" value={metrics.trading.trades.toString()} change="+11.4%" icon="📊" />
            <MetricCard title="Win Rate" value={`${metrics.trading.winRate}%`} change="+3.2%" icon="🏆" />
            <MetricCard title="Avg Return" value={`${metrics.trading.avgReturn}%`} change="+0.8%" icon="📈" />
          </>
        )}
        {selectedMetric === 'coffee' && (
          <>
            <MetricCard title="Total Orders" value={metrics.coffee.orders.toString()} change="+22.1%" icon="📦" />
            <MetricCard title="Revenue" value={`$${metrics.coffee.revenue.toLocaleString()}`} change="+19.5%" icon="💰" />
            <MetricCard title="Avg Order Value" value={`$${metrics.coffee.avgOrderValue}`} change="+5.3%" icon="🛒" />
            <MetricCard title="Satisfaction" value={`${metrics.coffee.satisfaction}/5`} change="+0.2" icon="⭐" />
          </>
        )}
        {selectedMetric === 'ai' && (
          <>
            <MetricCard title="Active Agents" value={metrics.ai.agents.toString()} change="+2" icon="🤖" />
            <MetricCard title="Efficiency" value={`${metrics.ai.efficiency}%`} change="+1.8%" icon="⚡" />
            <MetricCard title="Cost Savings" value={`$${metrics.ai.savings.toLocaleString()}`} change="+25.4%" icon="💡" />
            <MetricCard title="Uptime" value={`${metrics.ai.uptime}%`} change="+0.1%" icon="🔄" />
          </>
        )}
      </motion.div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-xl font-semibold text-white mb-6">Performance Over Time</h3>
          <div className="h-64 flex items-center justify-center">
            <div className="text-center">
              <div className="text-4xl mb-2">📈</div>
              <p className="text-slate-400">Interactive performance chart</p>
              <p className="text-slate-500 text-sm mt-1">Revenue, trades, and profit trends</p>
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-xl font-semibold text-white mb-6">Asset Distribution</h3>
          <div className="h-64 flex items-center justify-center">
            <div className="text-center">
              <div className="text-4xl mb-2">🥧</div>
              <p className="text-slate-400">Portfolio allocation breakdown</p>
              <p className="text-slate-500 text-sm mt-1">Trading pairs and coffee assets</p>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Top Performers */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <h3 className="text-xl font-semibold text-white mb-6">Top Performing Assets</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {topPerformers.map((performer, index) => (
            <motion.div
              key={performer.name}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="bg-slate-700/30 rounded-xl p-4 hover:bg-slate-700/50 transition-colors duration-200"
            >
              <div className="flex items-center justify-between mb-3">
                <h4 className="text-white font-semibold">{performer.name}</h4>
                <span className="text-green-400 font-bold text-sm">{performer.profit}</span>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Trades</span>
                  <span className="text-white">{performer.trades}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Win Rate</span>
                  <span className="text-green-400">{performer.winRate}%</span>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </motion.div>

      {/* Recent Insights */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <h3 className="text-xl font-semibold text-white mb-6">Recent Insights</h3>
        <div className="space-y-4">
          {recentInsights.map((insight, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
              className="flex items-start space-x-4 p-4 bg-slate-700/30 rounded-lg hover:bg-slate-700/50 transition-colors duration-200"
            >
              <div className="text-2xl flex-shrink-0">{insight.icon}</div>
              <div className="flex-1">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="text-white font-semibold">{insight.title}</h4>
                  <div className="flex items-center space-x-2">
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                      insight.impact === 'high' ? 'bg-red-500/20 text-red-400' :
                      insight.impact === 'medium' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-green-500/20 text-green-400'
                    }`}>
                      {insight.impact} impact
                    </span>
                    <span className="text-slate-400 text-xs">{insight.time}</span>
                  </div>
                </div>
                <p className="text-slate-300 text-sm">{insight.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </motion.div>

      {/* Export Options */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <h3 className="text-xl font-semibold text-white mb-6">Export & Reports</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            { label: 'Trading Report', description: 'Detailed trading performance analysis', icon: '📊' },
            { label: 'Tax Report', description: 'Tax-ready transaction summary', icon: '📋' },
            { label: 'Custom Report', description: 'Build your own analytics report', icon: '🔧' }
          ].map((report, index) => (
            <motion.button
              key={report.label}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className="flex flex-col items-start space-y-2 p-4 bg-slate-700/30 rounded-xl hover:bg-slate-700/50 transition-all duration-200 text-left"
            >
              <div className="text-2xl">{report.icon}</div>
              <h4 className="text-white font-semibold">{report.label}</h4>
              <p className="text-slate-400 text-sm">{report.description}</p>
            </motion.button>
          ))}
        </div>
      </motion.div>
    </div>
  )
}

interface MetricCardProps {
  title: string
  value: string
  change: string
  icon: string
}

function MetricCard({ title, value, change, icon }: MetricCardProps) {
  const isPositive = change.startsWith('+')
  
  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 hover:border-amber-500/30 transition-all duration-300"
    >
      <div className="flex items-center justify-between mb-4">
        <div className="text-2xl">{icon}</div>
        <div className={`text-sm font-medium ${isPositive ? 'text-green-400' : 'text-red-400'}`}>
          {change}
        </div>
      </div>
      <div>
        <p className="text-2xl font-bold text-white mb-1">{value}</p>
        <p className="text-slate-400 text-sm">{title}</p>
      </div>
    </motion.div>
  )
}
