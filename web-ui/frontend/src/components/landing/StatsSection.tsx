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
      description: 'Processed across all trading pairs'
    },
    {
      icon: '👥',
      label: 'Active Users',
      value: 12500,
      suffix: '+',
      prefix: '',
      description: 'Traders and coffee enthusiasts'
    },
    {
      icon: '☕',
      label: 'Coffee Orders',
      value: 45000,
      suffix: '+',
      prefix: '',
      description: 'Successfully delivered worldwide'
    },
    {
      icon: '🤖',
      label: 'AI Agents',
      value: 25,
      suffix: '+',
      prefix: '',
      description: 'Working 24/7 for optimization'
    },
    {
      icon: '🌍',
      label: 'Countries',
      value: 45,
      suffix: '+',
      prefix: '',
      description: 'Global presence and growing'
    },
    {
      icon: '⚡',
      label: 'Uptime',
      value: 99.9,
      suffix: '%',
      prefix: '',
      description: 'Reliable infrastructure'
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
          <h2 className="text-4xl md:text-5xl font-bold mb-6">
            <span className="bg-gradient-to-r from-amber-400 to-orange-400 bg-clip-text text-transparent">
              Platform Statistics
            </span>
          </h2>
          <p className="text-xl text-slate-300 max-w-3xl mx-auto">
            Real numbers from our growing ecosystem of traders, coffee lovers, and AI-powered automation
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {stats.map((stat, index) => (
            <StatCard
              key={stat.label}
              stat={stat}
              index={index}
              inView={inView}
            />
          ))}
        </div>

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

  return (
    <motion.div
      initial={{ opacity: 0, y: 50, scale: 0.9 }}
      animate={inView ? { opacity: 1, y: 0, scale: 1 } : { opacity: 0, y: 50, scale: 0.9 }}
      transition={{ delay: index * 0.1, duration: 0.8, ease: "easeOut" }}
      whileHover={{ 
        scale: 1.05,
        transition: { duration: 0.2 }
      }}
      className="group relative"
    >
      <div className="h-full bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-8 text-center hover:border-amber-500/30 transition-all duration-300">
        {/* Icon */}
        <div className="text-4xl mb-4 group-hover:scale-110 transition-transform duration-300">
          {stat.icon}
        </div>

        {/* Value */}
        <div className="mb-2">
          <span className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-amber-400 to-orange-400 bg-clip-text text-transparent">
            {stat.prefix}{stat.value === displayValue ? stat.value : formatValue(displayValue)}{stat.suffix}
          </span>
        </div>

        {/* Label */}
        <h3 className="text-xl font-semibold text-white mb-2 group-hover:text-amber-400 transition-colors duration-300">
          {stat.label}
        </h3>

        {/* Description */}
        <p className="text-slate-400 text-sm">
          {stat.description}
        </p>

        {/* Hover Glow */}
        <div className="absolute inset-0 bg-gradient-to-r from-amber-500/10 to-orange-500/10 opacity-0 group-hover:opacity-100 rounded-2xl transition-opacity duration-300" />
      </div>
    </motion.div>
  )
}
