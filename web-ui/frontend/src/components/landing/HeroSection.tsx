'use client'

import { motion } from 'framer-motion'
import { useEffect, useState } from 'react'

interface HeroSectionProps {
  onEnterDashboard: () => void
  scrollY: number
}

export default function HeroSection({ onEnterDashboard, scrollY }: HeroSectionProps) {
  const [typedText, setTypedText] = useState('')
  const fullText = 'The Future of Coffee Trading'

  useEffect(() => {
    let index = 0
    const timer = setInterval(() => {
      if (index < fullText.length) {
        setTypedText(fullText.slice(0, index + 1))
        index++
      } else {
        clearInterval(timer)
      }
    }, 100)

    return () => clearInterval(timer)
  }, [])

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.2,
        delayChildren: 0.3
      }
    }
  }

  const itemVariants = {
    hidden: { opacity: 0, y: 30 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.8, ease: "easeOut" }
    }
  }

  return (
    <section className="relative min-h-screen flex items-center justify-center pt-16 overflow-hidden">
      {/* Enhanced Parallax Background */}
      <div
        className="absolute inset-0 opacity-30"
        style={{ transform: `translateY(${scrollY * 0.5}px)` }}
      >
        {/* Coffee Steam Animation */}
        <div className="absolute top-20 left-1/2 transform -translate-x-1/2">
          <div className="flex space-x-2">
            {[...Array(3)].map((_, i) => (
              <div
                key={i}
                className="w-1 h-8 bg-gradient-to-t from-coffee-400/60 to-transparent rounded-full animate-coffee-steam"
                style={{ animationDelay: `${i * 0.5}s` }}
              />
            ))}
          </div>
        </div>

        {/* Enhanced Background Orbs */}
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-gradient-to-r from-coffee-400/30 to-brand-amber/30 rounded-full blur-3xl animate-float" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-gradient-to-r from-crypto-bitcoin/20 to-crypto-ethereum/20 rounded-full blur-3xl animate-float delay-1000" />
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-gradient-radial from-coffee-500/10 to-transparent rounded-full animate-pulse" />

        {/* Floating Coffee Beans */}
        <div className="absolute top-20 left-10 w-3 h-3 bg-coffee-600 rounded-full animate-bounce-gentle" />
        <div className="absolute top-40 right-20 w-2 h-2 bg-coffee-500 rounded-full animate-bounce-gentle delay-1000" />
        <div className="absolute bottom-40 left-1/4 w-2.5 h-2.5 bg-coffee-400 rounded-full animate-bounce-gentle delay-2000" />
        <div className="absolute bottom-20 right-1/3 w-3 h-3 bg-coffee-600 rounded-full animate-bounce-gentle delay-3000" />
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center relative z-10">
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="space-y-8"
        >
          {/* Main Heading */}
          <motion.div variants={itemVariants} className="space-y-4">
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ duration: 0.8, delay: 0.2 }}
              className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-amber-500/20 to-orange-500/20 border border-amber-500/30 rounded-full text-amber-300 text-sm font-medium mb-6"
            >
              <span className="w-2 h-2 bg-amber-400 rounded-full mr-2 animate-pulse" />
              Web3 Coffee Ecosystem
            </motion.div>
            
            <div className="relative">
              <h1 className="text-5xl md:text-7xl lg:text-8xl font-bold leading-tight font-display">
                <span className="bg-gradient-to-r from-coffee-400 via-brand-amber to-coffee-600 bg-clip-text text-transparent animate-gradient-shift bg-[length:200%_200%]">
                  {typedText}
                </span>
                <motion.span
                  animate={{ opacity: [1, 0] }}
                  transition={{ duration: 0.8, repeat: Infinity, repeatType: "reverse" }}
                  className="text-coffee-400 glow-text"
                >
                  |
                </motion.span>
              </h1>
              {/* Glow effect behind text */}
              <div className="absolute inset-0 text-5xl md:text-7xl lg:text-8xl font-bold leading-tight font-display blur-2xl opacity-20 pointer-events-none">
                <span className="bg-gradient-to-r from-coffee-400 via-brand-amber to-coffee-600 bg-clip-text text-transparent">
                  {typedText}
                </span>
              </div>
            </div>
          </motion.div>

          {/* Enhanced Subtitle */}
          <motion.p
            variants={itemVariants}
            className="text-xl md:text-2xl text-slate-300 max-w-4xl mx-auto leading-relaxed font-body"
          >
            Experience the revolutionary fusion of{' '}
            <span className="gradient-text font-semibold">coffee commerce</span>,{' '}
            <span className="gradient-text-crypto font-semibold">DeFi trading</span>, and{' '}
            <span className="text-brand-amber font-semibold glow-text">AI automation</span>{' '}
            in one powerful platform
          </motion.p>

          {/* Feature Pills */}
          <motion.div
            variants={itemVariants}
            className="flex flex-wrap justify-center gap-4 text-sm"
          >
            {[
              '🚀 High-Performance Go Backend',
              '🤖 AI-Powered Trading Bots',
              '☕ Real Coffee Orders',
              '💰 Multi-Chain DeFi',
              '📊 Advanced Analytics'
            ].map((feature, index) => (
              <motion.div
                key={feature}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 1 + index * 0.1 }}
                className="px-4 py-2 glass-card text-slate-300 hover:text-coffee-300 hover:border-coffee-500/30 transition-all duration-300 hover-lift"
              >
                {feature}
              </motion.div>
            ))}
          </motion.div>

          {/* CTA Buttons */}
          <motion.div
            variants={itemVariants}
            className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-8"
          >
            <motion.button
              whileHover={{ scale: 1.05, boxShadow: "0 20px 40px rgba(212, 165, 116, 0.4)" }}
              whileTap={{ scale: 0.95 }}
              onClick={onEnterDashboard}
              className="btn-primary group relative overflow-hidden min-w-[220px]"
            >
              <span className="relative z-10 flex items-center gap-2">
                🚀 Launch Trading Platform
                <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform duration-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
              </span>
              <div className="absolute inset-0 bg-gradient-to-r from-coffee-600 to-coffee-700 transform scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" />
            </motion.button>

            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-ghost group min-w-[220px]"
            >
              <span className="flex items-center gap-2">
                📖 View Documentation
                <svg className="w-4 h-4 group-hover:translate-x-1 transition-transform duration-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                </svg>
              </span>
            </motion.button>
          </motion.div>

          {/* Stats Preview */}
          <motion.div
            variants={itemVariants}
            className="grid grid-cols-2 md:grid-cols-4 gap-8 pt-16 max-w-4xl mx-auto"
          >
            {[
              { label: 'Total Volume', value: '$2.4M+', icon: '💰' },
              { label: 'Active Traders', value: '12.5K+', icon: '👥' },
              { label: 'Coffee Orders', value: '45K+', icon: '☕' },
              { label: 'AI Agents', value: '25+', icon: '🤖' }
            ].map((stat, index) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 1.5 + index * 0.1 }}
                className="text-center stats-card hover-lift"
              >
                <div className="text-3xl mb-3 animate-bounce-gentle" style={{ animationDelay: `${index * 0.2}s` }}>{stat.icon}</div>
                <div className="text-2xl md:text-3xl font-bold gradient-text mb-1">{stat.value}</div>
                <div className="text-slate-400 text-sm font-medium">{stat.label}</div>
              </motion.div>
            ))}
          </motion.div>
        </motion.div>
      </div>

      {/* Scroll Indicator */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 2 }}
        className="absolute bottom-8 left-1/2 transform -translate-x-1/2"
      >
        <motion.div
          animate={{ y: [0, 10, 0] }}
          transition={{ duration: 2, repeat: Infinity }}
          className="w-6 h-10 border-2 border-coffee-400/50 rounded-full flex justify-center backdrop-blur-sm bg-slate-900/20 hover:border-coffee-400 transition-colors duration-300"
        >
          <motion.div
            animate={{ y: [0, 12, 0] }}
            transition={{ duration: 2, repeat: Infinity }}
            className="w-1 h-3 bg-gradient-to-b from-coffee-400 to-brand-amber rounded-full mt-2 shadow-glow"
          />
        </motion.div>
      </motion.div>
    </section>
  )
}
