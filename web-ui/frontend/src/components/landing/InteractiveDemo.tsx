'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'
import { useState, useEffect } from 'react'

export default function InteractiveDemo() {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.1
  })

  const [activeDemo, setActiveDemo] = useState('coffee-order')
  const [isPlaying, setIsPlaying] = useState(false)

  const demos = [
    {
      id: 'coffee-order',
      title: 'Coffee Ordering',
      icon: '☕',
      description: 'Complete coffee ordering workflow with real-time updates',
      gradient: 'from-amber-500 to-orange-500'
    },
    {
      id: 'crypto-payment',
      title: 'Crypto Payments',
      icon: '💰',
      description: 'Multi-chain cryptocurrency payment processing',
      gradient: 'from-green-500 to-emerald-500'
    },
    {
      id: 'ai-trading',
      title: 'AI Trading',
      icon: '🤖',
      description: 'Automated trading strategies and portfolio management',
      gradient: 'from-blue-500 to-purple-500'
    },
    {
      id: 'analytics',
      title: 'Analytics Dashboard',
      icon: '📊',
      description: 'Real-time business intelligence and performance metrics',
      gradient: 'from-purple-500 to-pink-500'
    }
  ]

  // Simulated real-time data
  const [demoData, setDemoData] = useState({
    orders: 1247,
    revenue: 45678.90,
    activeUsers: 234,
    aiAgents: 25,
    cryptoPrice: 45234.56,
    tradingVolume: 2.4
  })

  useEffect(() => {
    const interval = setInterval(() => {
      setDemoData(prev => ({
        orders: prev.orders + Math.floor(Math.random() * 3),
        revenue: prev.revenue + (Math.random() * 50),
        activeUsers: prev.activeUsers + Math.floor(Math.random() * 5) - 2,
        aiAgents: 25,
        cryptoPrice: prev.cryptoPrice + (Math.random() - 0.5) * 100,
        tradingVolume: prev.tradingVolume + (Math.random() - 0.5) * 0.1
      }))
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  const renderDemoContent = () => {
    switch (activeDemo) {
      case 'coffee-order':
        return (
          <div className="space-y-6">
            <div className="bg-slate-800/50 rounded-xl p-6">
              <h4 className="text-lg font-bold text-white mb-4">☕ New Order Processing</h4>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 bg-amber-500/10 border border-amber-500/20 rounded-lg">
                  <span className="text-amber-300">Order #1247</span>
                  <span className="text-green-400">✓ Confirmed</span>
                </div>
                <div className="flex items-center justify-between p-3 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                  <span className="text-blue-300">Payment Processing</span>
                  <span className="text-yellow-400">⏳ Pending</span>
                </div>
                <div className="flex items-center justify-between p-3 bg-purple-500/10 border border-purple-500/20 rounded-lg">
                  <span className="text-purple-300">Inventory Check</span>
                  <span className="text-green-400">✓ Available</span>
                </div>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-slate-800/50 rounded-xl p-4 text-center">
                <div className="text-2xl font-bold text-amber-400">{demoData.orders}</div>
                <div className="text-sm text-slate-400">Total Orders</div>
              </div>
              <div className="bg-slate-800/50 rounded-xl p-4 text-center">
                <div className="text-2xl font-bold text-green-400">${demoData.revenue.toFixed(2)}</div>
                <div className="text-sm text-slate-400">Revenue</div>
              </div>
            </div>
          </div>
        )

      case 'crypto-payment':
        return (
          <div className="space-y-6">
            <div className="bg-slate-800/50 rounded-xl p-6">
              <h4 className="text-lg font-bold text-white mb-4">💰 Multi-Chain Payment</h4>
              <div className="space-y-3">
                {[
                  { chain: 'Ethereum', symbol: 'ETH', price: '$2,834.12', status: 'Active' },
                  { chain: 'Solana', symbol: 'SOL', price: '$98.45', status: 'Active' },
                  { chain: 'Polygon', symbol: 'MATIC', price: '$0.87', status: 'Active' }
                ].map((chain, idx) => (
                  <div key={idx} className="flex items-center justify-between p-3 bg-green-500/10 border border-green-500/20 rounded-lg">
                    <div>
                      <span className="text-green-300 font-medium">{chain.chain}</span>
                      <span className="text-slate-400 ml-2">({chain.symbol})</span>
                    </div>
                    <div className="text-right">
                      <div className="text-white">{chain.price}</div>
                      <div className="text-green-400 text-sm">{chain.status}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <div className="bg-slate-800/50 rounded-xl p-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-green-400">${demoData.tradingVolume.toFixed(1)}M</div>
                <div className="text-sm text-slate-400">24h Volume</div>
              </div>
            </div>
          </div>
        )

      case 'ai-trading':
        return (
          <div className="space-y-6">
            <div className="bg-slate-800/50 rounded-xl p-6">
              <h4 className="text-lg font-bold text-white mb-4">🤖 AI Trading Agents</h4>
              <div className="space-y-3">
                {[
                  { name: 'Arbitrage Bot', status: 'Active', profit: '+$234.56' },
                  { name: 'Grid Trading', status: 'Active', profit: '+$156.78' },
                  { name: 'Yield Farmer', status: 'Active', profit: '+$89.12' }
                ].map((bot, idx) => (
                  <div key={idx} className="flex items-center justify-between p-3 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                    <div>
                      <span className="text-blue-300 font-medium">{bot.name}</span>
                      <div className="text-green-400 text-sm">{bot.status}</div>
                    </div>
                    <div className="text-green-400 font-bold">{bot.profit}</div>
                  </div>
                ))}
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-slate-800/50 rounded-xl p-4 text-center">
                <div className="text-2xl font-bold text-blue-400">{demoData.aiAgents}</div>
                <div className="text-sm text-slate-400">Active Agents</div>
              </div>
              <div className="bg-slate-800/50 rounded-xl p-4 text-center">
                <div className="text-2xl font-bold text-purple-400">24/7</div>
                <div className="text-sm text-slate-400">Monitoring</div>
              </div>
            </div>
          </div>
        )

      case 'analytics':
        return (
          <div className="space-y-6">
            <div className="bg-slate-800/50 rounded-xl p-6">
              <h4 className="text-lg font-bold text-white mb-4">📊 Real-time Analytics</h4>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center">
                  <div className="text-3xl font-bold text-purple-400">{demoData.activeUsers}</div>
                  <div className="text-sm text-slate-400">Active Users</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-pink-400">98.7%</div>
                  <div className="text-sm text-slate-400">Uptime</div>
                </div>
              </div>
            </div>
            <div className="bg-slate-800/50 rounded-xl p-6">
              <div className="flex items-center justify-between mb-4">
                <span className="text-white font-medium">Performance Metrics</span>
                <span className="text-green-400 text-sm">All Systems Operational</span>
              </div>
              <div className="space-y-3">
                {[
                  { metric: 'API Response Time', value: '< 50ms', color: 'green' },
                  { metric: 'Database Queries', value: '< 10ms', color: 'green' },
                  { metric: 'Cache Hit Rate', value: '94.2%', color: 'blue' }
                ].map((item, idx) => (
                  <div key={idx} className="flex items-center justify-between">
                    <span className="text-slate-300">{item.metric}</span>
                    <span className={`text-${item.color}-400 font-medium`}>{item.value}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )

      default:
        return null
    }
  }

  return (
    <section className="py-24 relative overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0" style={{
          backgroundImage: `radial-gradient(circle at 20% 80%, #f59e0b 1px, transparent 1px),
                           radial-gradient(circle at 80% 20%, #d97706 1px, transparent 1px),
                           radial-gradient(circle at 40% 40%, #f59e0b 1px, transparent 1px)`,
          backgroundSize: '80px 80px'
        }} />
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <motion.div
          ref={ref}
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ duration: 0.8 }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-green-500/20 to-emerald-500/20 border border-green-500/30 rounded-full text-green-300 text-sm font-medium mb-6">
            <span className="w-2 h-2 bg-green-400 rounded-full mr-2 animate-pulse" />
            Interactive Platform Demo
          </div>
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="bg-gradient-to-r from-green-400 via-emerald-400 to-green-500 bg-clip-text text-transparent">
              See It In Action
            </span>
          </h2>
          <p className="text-xl text-slate-300 max-w-4xl mx-auto leading-relaxed">
            Experience live demonstrations of our platform features with real-time data and interactive examples
          </p>
        </motion.div>

        {/* Demo Selector */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
          transition={{ delay: 0.3 }}
          className="flex flex-wrap justify-center gap-4 mb-12"
        >
          {demos.map((demo) => (
            <motion.button
              key={demo.id}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setActiveDemo(demo.id)}
              className={`px-6 py-4 rounded-xl border transition-all duration-300 ${
                activeDemo === demo.id
                  ? `bg-gradient-to-r ${demo.gradient} bg-opacity-20 border-opacity-50 text-white`
                  : 'bg-slate-800/50 border-slate-700/50 text-slate-300 hover:border-green-500/30 hover:text-green-400'
              }`}
            >
              <div className="flex flex-col items-center gap-2">
                <span className="text-2xl">{demo.icon}</span>
                <span className="font-medium">{demo.title}</span>
                <span className="text-xs text-slate-400 text-center max-w-[120px]">{demo.description}</span>
              </div>
            </motion.button>
          ))}
        </motion.div>

        {/* Demo Content */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
          transition={{ delay: 0.5 }}
          className="bg-slate-900/50 backdrop-blur-sm border border-slate-700/50 rounded-3xl p-8"
        >
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center gap-4">
              <div className={`w-12 h-12 bg-gradient-to-r ${demos.find(d => d.id === activeDemo)?.gradient} rounded-xl flex items-center justify-center text-xl`}>
                {demos.find(d => d.id === activeDemo)?.icon}
              </div>
              <div>
                <h3 className="text-2xl font-bold text-white">
                  {demos.find(d => d.id === activeDemo)?.title} Demo
                </h3>
                <p className="text-slate-400">
                  {demos.find(d => d.id === activeDemo)?.description}
                </p>
              </div>
            </div>
            
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse" />
              <span className="text-green-400 text-sm font-medium">Live Data</span>
            </div>
          </div>

          {/* Demo Content Area */}
          <div className="min-h-[400px]">
            {renderDemoContent()}
          </div>
        </motion.div>

        {/* API Playground Teaser */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 0.8 }}
          className="mt-16 bg-gradient-to-r from-slate-800/30 to-slate-700/30 backdrop-blur-sm border border-slate-600/30 rounded-3xl p-8"
        >
          <div className="text-center mb-6">
            <h3 className="text-2xl font-bold text-white mb-2">
              🚀 Ready to Build?
            </h3>
            <p className="text-slate-300">
              Access our comprehensive APIs and start integrating with the Go Coffee ecosystem
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center">
              <div className="text-3xl mb-2">📚</div>
              <div className="text-lg font-bold text-white mb-1">API Documentation</div>
              <div className="text-sm text-slate-400">Complete API reference with examples</div>
            </div>
            <div className="text-center">
              <div className="text-3xl mb-2">🔧</div>
              <div className="text-lg font-bold text-white mb-1">SDK Libraries</div>
              <div className="text-sm text-slate-400">Ready-to-use SDKs for popular languages</div>
            </div>
            <div className="text-center">
              <div className="text-3xl mb-2">🎮</div>
              <div className="text-lg font-bold text-white mb-1">Interactive Playground</div>
              <div className="text-sm text-slate-400">Test APIs directly in your browser</div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
