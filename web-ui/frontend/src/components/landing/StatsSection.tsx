'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'
import { useState, useEffect } from 'react'

export default function StatsSection() {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.3
  })

  const stats = [
    {
      icon: '💰',
      label: 'Total Trading Volume',
      value: 2400000,
      suffix: '+',
      prefix: '$',
      description: 'Processed across all trading pairs',
      category: 'trading',
      trend: '+12.5%'
    },
    {
      icon: '👥',
      label: 'Active Users',
      value: 12500,
      suffix: '+',
      prefix: '',
      description: 'Traders and coffee enthusiasts',
      category: 'users',
      trend: '+8.3%'
    },
    {
      icon: '☕',
      label: 'Coffee Orders',
      value: 45000,
      suffix: '+',
      prefix: '',
      description: 'Successfully delivered worldwide',
      category: 'orders',
      trend: '+15.7%'
    },
    {
      icon: '🤖',
      label: 'AI Agents',
      value: 25,
      suffix: '+',
      prefix: '',
      description: 'Working 24/7 for optimization',
      category: 'ai',
      trend: 'Stable'
    },
    {
      icon: '🌍',
      label: 'Countries',
      value: 45,
      suffix: '+',
      prefix: '',
      description: 'Global presence and growing',
      category: 'global',
      trend: '+3 new'
    },
    {
      icon: '⚡',
      label: 'Uptime',
      value: 99.9,
      suffix: '%',
      prefix: '',
      description: 'Reliable infrastructure',
      category: 'performance',
      trend: '99.9%'
    },
    {
      icon: '🏗️',
      label: 'Microservices',
      value: 38,
      suffix: '+',
      prefix: '',
      description: 'Cloud-native architecture',
      category: 'infrastructure',
      trend: '+5 new'
    },
    {
      icon: '🔗',
      label: 'API Endpoints',
      value: 200,
      suffix: '+',
      prefix: '',
      description: 'Comprehensive API coverage',
      category: 'api',
      trend: '+25 new'
    },
    {
      icon: '🌐',
      label: 'Blockchain Networks',
      value: 4,
      suffix: '',
      prefix: '',
      description: 'Multi-chain Web3 support',
      category: 'web3',
      trend: 'Complete'
    }
  ]

  return (
    <section className="py-24 relative overflow-hidden">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0" style={{
          backgroundImage: `radial-gradient(circle at 25% 25%, #f59e0b 1px, transparent 1px),
                           radial-gradient(circle at 75% 75%, #d97706 1px, transparent 1px)`,
          backgroundSize: '50px 50px'
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
          <div className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-cyan-500/20 to-blue-500/20 border border-cyan-500/30 rounded-full text-cyan-300 text-sm font-medium mb-6">
            <span className="w-2 h-2 bg-cyan-400 rounded-full mr-2 animate-pulse" />
            Real-time Platform Metrics
          </div>
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="bg-gradient-to-r from-cyan-400 via-blue-400 to-cyan-500 bg-clip-text text-transparent">
              Platform Statistics
            </span>
          </h2>
          <p className="text-xl text-slate-300 max-w-4xl mx-auto leading-relaxed">
            Live metrics from our comprehensive ecosystem spanning coffee commerce, DeFi trading,
            AI automation, and enterprise infrastructure
          </p>
        </motion.div>

        {/* Enhanced Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-3 gap-6">
          {stats.map((stat, index) => (
            <StatCard
              key={stat.label}
              stat={stat}
              index={index}
              inView={inView}
            />
          ))}
        </div>

        {/* Performance Metrics Dashboard */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 1.5 }}
          className="mt-20 bg-gradient-to-r from-slate-800/30 to-slate-700/30 backdrop-blur-sm border border-slate-600/30 rounded-3xl p-8"
        >
          <div className="text-center mb-8">
            <h3 className="text-2xl font-bold text-white mb-2">
              🚀 System Performance
            </h3>
            <p className="text-slate-300">
              Real-time performance indicators across all platform components
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-6">
            {[
              { label: 'API Response', value: '<50ms', icon: '⚡', color: 'green' },
              { label: 'Database Query', value: '<10ms', icon: '🗄️', color: 'blue' },
              { label: 'Cache Hit Rate', value: '94.2%', icon: '🎯', color: 'purple' },
              { label: 'Error Rate', value: '0.01%', icon: '🛡️', color: 'green' },
              { label: 'Throughput', value: '10K/sec', icon: '📈', color: 'cyan' },
              { label: 'Availability', value: '99.99%', icon: '💚', color: 'emerald' }
            ].map((metric, index) => (
              <motion.div
                key={metric.label}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.8 }}
                transition={{ delay: 1.7 + index * 0.1 }}
                className="text-center bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-xl p-4 hover:border-cyan-500/30 transition-all duration-300"
              >
                <div className="text-2xl mb-2">{metric.icon}</div>
                <div className={`text-xl font-bold text-${metric.color}-400 mb-1`}>{metric.value}</div>
                <div className="text-xs text-slate-400">{metric.label}</div>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Additional Info */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 1.2, duration: 0.8 }}
          className="mt-16 text-center"
        >
          <div className="bg-gradient-to-r from-slate-800/50 to-slate-700/50 backdrop-blur-sm border border-slate-600/50 rounded-2xl p-8 max-w-4xl mx-auto">
            <h3 className="text-2xl font-bold text-white mb-4">
              🚀 Growing Every Day
            </h3>
            <p className="text-slate-300 mb-6">
              Our platform continues to expand with new features, partnerships, and innovations. 
              Join thousands of users who are already benefiting from our integrated ecosystem.
            </p>
            <div className="flex flex-wrap justify-center gap-4 text-sm">
              {[
                '24/7 Support',
                'Regular Updates',
                'Community Driven',
                'Open Source',
                'Enterprise Ready'
              ].map((feature, index) => (
                <motion.div
                  key={feature}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.8 }}
                  transition={{ delay: 1.4 + index * 0.1 }}
                  className="px-4 py-2 bg-amber-500/20 border border-amber-500/30 rounded-full text-amber-300"
                >
                  {feature}
                </motion.div>
              ))}
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

interface StatCardProps {
  stat: {
    icon: string
    label: string
    value: number
    suffix: string
    prefix: string
    description: string
    category: string
    trend: string
  }
  index: number
  inView: boolean
}

function StatCard({ stat, index, inView }: StatCardProps) {
  const [displayValue, setDisplayValue] = useState(0)

  useEffect(() => {
    if (inView) {
      const timer = setTimeout(() => {
        const duration = 2000
        const steps = 60
        const increment = stat.value / steps
        let current = 0
        
        const counter = setInterval(() => {
          current += increment
          if (current >= stat.value) {
            setDisplayValue(stat.value)
            clearInterval(counter)
          } else {
            setDisplayValue(Math.floor(current))
          }
        }, duration / steps)

        return () => clearInterval(counter)
      }, index * 200)

      return () => clearTimeout(timer)
    }
  }, [inView, stat.value, index])

  const formatValue = (value: number) => {
    if (value >= 1000000) {
      return (value / 1000000).toFixed(1) + 'M'
    } else if (value >= 1000) {
      return (value / 1000).toFixed(1) + 'K'
    }
    return value.toString()
  }

  const getCategoryColor = (category: string) => {
    const colors = {
      trading: 'from-green-400 to-emerald-500',
      users: 'from-blue-400 to-cyan-500',
      orders: 'from-amber-400 to-orange-500',
      ai: 'from-purple-400 to-pink-500',
      global: 'from-cyan-400 to-blue-500',
      performance: 'from-green-400 to-emerald-500',
      infrastructure: 'from-slate-400 to-gray-500',
      api: 'from-indigo-400 to-purple-500',
      web3: 'from-purple-500 to-pink-600'
    }
    return colors[category as keyof typeof colors] || 'from-amber-400 to-orange-500'
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 50, scale: 0.9 }}
      animate={inView ? { opacity: 1, y: 0, scale: 1 } : { opacity: 0, y: 50, scale: 0.9 }}
      transition={{ delay: index * 0.08, duration: 0.8, ease: "easeOut" }}
      whileHover={{
        scale: 1.02,
        y: -5,
        transition: { duration: 0.2 }
      }}
      className="group relative"
    >
      <div className="h-full bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 text-center hover:border-cyan-500/30 hover:bg-slate-800/60 transition-all duration-300">
        {/* Category Badge */}
        <div className="flex items-center justify-between mb-4">
          <span className="text-xs text-slate-400 uppercase tracking-wide font-medium bg-slate-700/50 px-2 py-1 rounded-md">
            {stat.category}
          </span>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
            <span className="text-xs text-green-400">Live</span>
          </div>
        </div>

        {/* Icon */}
        <div className="text-3xl mb-4 group-hover:scale-110 transition-transform duration-300">
          {stat.icon}
        </div>

        {/* Value */}
        <div className="mb-3">
          <span className={`text-3xl md:text-4xl font-bold bg-gradient-to-r ${getCategoryColor(stat.category)} bg-clip-text text-transparent`}>
            {stat.prefix}{stat.value === displayValue ? stat.value : formatValue(displayValue)}{stat.suffix}
          </span>
        </div>

        {/* Label */}
        <h3 className="text-lg font-semibold text-white mb-2 group-hover:text-cyan-400 transition-colors duration-300">
          {stat.label}
        </h3>

        {/* Description */}
        <p className="text-slate-400 text-sm mb-3 leading-relaxed">
          {stat.description}
        </p>

        {/* Trend Indicator */}
        <div className="border-t border-slate-700/50 pt-3">
          <div className="flex items-center justify-center gap-2">
            <span className="text-xs text-slate-400">Trend:</span>
            <span className={`text-xs font-medium ${
              stat.trend.includes('+') ? 'text-green-400' :
              stat.trend.includes('-') ? 'text-red-400' :
              'text-cyan-400'
            }`}>
              {stat.trend}
            </span>
          </div>
        </div>

        {/* Hover Glow */}
        <div className={`absolute inset-0 bg-gradient-to-r ${getCategoryColor(stat.category)} opacity-0 group-hover:opacity-5 rounded-2xl transition-opacity duration-300`} />
      </div>
    </motion.div>
  )
}
