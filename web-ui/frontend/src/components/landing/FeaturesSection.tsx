'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'
import { useState } from 'react'

export default function FeaturesSection() {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.1
  })

  const [activeCategory, setActiveCategory] = useState('all')

  const categories = [
    { id: 'all', label: 'All Features', icon: '🌟' },
    { id: 'coffee', label: 'Coffee', icon: '☕' },
    { id: 'web3', label: 'Web3 & DeFi', icon: '🌐' },
    { id: 'ai', label: 'AI & Automation', icon: '🤖' },
    { id: 'enterprise', label: 'Enterprise', icon: '🏢' }
  ]

  const features = [
    // Coffee Commerce Features
    {
      icon: '☕',
      title: 'Coffee Commerce',
      category: 'coffee',
      description: 'Complete coffee ordering ecosystem with global suppliers, real-time inventory, and quality assurance.',
      details: ['Real-time inventory', 'Quality tracking', 'Global suppliers', 'Smart contracts', 'Order management'],
      gradient: 'from-amber-500 to-orange-500',
      endpoints: ['POST /orders', 'GET /inventory', 'GET /suppliers'],
      status: 'active'
    },
    {
      icon: '🏪',
      title: 'Coffee Shop Management',
      category: 'coffee',
      description: 'Comprehensive shop operations with staff management, equipment monitoring, and customer analytics.',
      details: ['Staff scheduling', 'Equipment IoT', 'Customer insights', 'Revenue tracking', 'Loyalty programs'],
      gradient: 'from-coffee-500 to-amber-600',
      endpoints: ['GET /shops', 'POST /staff', 'GET /analytics'],
      status: 'active'
    },
    {
      icon: '📦',
      title: 'Supply Chain',
      category: 'coffee',
      description: 'End-to-end supply chain visibility from farm to cup with blockchain traceability.',
      details: ['Farm tracking', 'Logistics optimization', 'Quality assurance', 'Blockchain records', 'Sustainability metrics'],
      gradient: 'from-green-600 to-emerald-500',
      endpoints: ['GET /supply-chain', 'POST /shipments', 'GET /traceability'],
      status: 'active'
    },

    // Web3 & DeFi Features
    {
      icon: '💰',
      title: 'Multi-Chain Payments',
      category: 'web3',
      description: 'Accept cryptocurrency payments across Ethereum, BSC, Polygon, and Solana networks.',
      details: ['4 blockchain networks', 'ETH, BNB, MATIC, SOL', 'USDC, USDT support', 'QR code payments', 'Real-time verification'],
      gradient: 'from-green-500 to-emerald-500',
      endpoints: ['POST /web3/payment', 'GET /web3/status', 'GET /web3/supported'],
      status: 'active'
    },
    {
      icon: '🔄',
      title: 'DeFi Trading Bots',
      category: 'web3',
      description: 'Automated trading strategies with yield farming, arbitrage, and grid trading across DEXs.',
      details: ['Uniswap V3 integration', 'Aave lending', '1inch aggregation', 'Automated strategies', 'Risk management'],
      gradient: 'from-blue-500 to-cyan-500',
      endpoints: ['POST /defi/trade', 'GET /defi/yield', 'GET /defi/strategies'],
      status: 'active'
    },
    {
      icon: '🪙',
      title: 'Coffee Token (COFFEE)',
      category: 'web3',
      description: 'Native platform token with staking rewards, governance rights, and utility benefits.',
      details: ['Staking rewards', 'Governance voting', 'Payment discounts', 'Liquidity mining', 'Burn mechanisms'],
      gradient: 'from-amber-500 to-yellow-500',
      endpoints: ['GET /token/price', 'POST /token/stake', 'GET /token/rewards'],
      status: 'active'
    },

    // AI & Automation Features
    {
      icon: '🤖',
      title: 'AI Trading Agents',
      category: 'ai',
      description: '25+ intelligent agents handling automated trading, market analysis, and portfolio optimization.',
      details: ['25+ active agents', 'Market sentiment analysis', 'Portfolio optimization', 'Risk assessment', '24/7 monitoring'],
      gradient: 'from-blue-500 to-purple-500',
      endpoints: ['GET /ai/agents', 'POST /ai/strategy', 'GET /ai/performance'],
      status: 'active'
    },
    {
      icon: '🧠',
      title: 'AI Search Engine',
      category: 'ai',
      description: 'Redis 8 powered semantic search with vector embeddings and hybrid search capabilities.',
      details: ['Semantic search', 'Vector embeddings', 'Hybrid queries', 'Real-time indexing', 'Personalized results'],
      gradient: 'from-purple-500 to-pink-500',
      endpoints: ['POST /ai-search/semantic', 'POST /ai-search/vector', 'GET /ai-search/trending'],
      status: 'active'
    },
    {
      icon: '📈',
      title: 'Predictive Analytics',
      category: 'ai',
      description: 'Machine learning models for demand forecasting, price prediction, and market trend analysis.',
      details: ['Demand forecasting', 'Price prediction', 'Trend analysis', 'Customer behavior', 'Inventory optimization'],
      gradient: 'from-indigo-500 to-purple-500',
      endpoints: ['GET /analytics/forecast', 'POST /analytics/predict', 'GET /analytics/trends'],
      status: 'active'
    },

    // Enterprise Features
    {
      icon: '🏗️',
      title: 'Microservices Architecture',
      category: 'enterprise',
      description: 'Cloud-native microservices with Kubernetes orchestration and service mesh.',
      details: ['38+ microservices', 'Kubernetes deployment', 'Service mesh', 'Auto-scaling', 'Health monitoring'],
      gradient: 'from-slate-500 to-gray-600',
      endpoints: ['GET /health', 'GET /metrics', 'GET /services'],
      status: 'active'
    },
    {
      icon: '🔒',
      title: 'Enterprise Security',
      category: 'enterprise',
      description: 'Bank-grade security with multi-factor authentication, encryption, and compliance monitoring.',
      details: ['Multi-factor auth', 'End-to-end encryption', 'SOC2 compliance', 'Audit trails', 'Threat detection'],
      gradient: 'from-red-500 to-orange-500',
      endpoints: ['POST /auth/login', 'GET /security/audit', 'POST /security/scan'],
      status: 'active'
    },
    {
      icon: '📊',
      title: 'Business Intelligence',
      category: 'enterprise',
      description: 'Comprehensive dashboards with real-time KPIs, custom reports, and executive insights.',
      details: ['Real-time dashboards', 'Custom reports', 'KPI tracking', 'Executive insights', 'Data visualization'],
      gradient: 'from-cyan-500 to-blue-500',
      endpoints: ['GET /bi/dashboard', 'POST /bi/report', 'GET /bi/kpis'],
      status: 'active'
    }
  ]

  const filteredFeatures = activeCategory === 'all'
    ? features
    : features.filter(feature => feature.category === activeCategory)

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.08,
        delayChildren: 0.2
      }
    }
  }

  const itemVariants = {
    hidden: { opacity: 0, y: 50 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.8, ease: "easeOut" }
    }
  }

  return (
    <section id="features" className="py-24 relative overflow-hidden">
      {/* Enhanced Background */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0" style={{
          backgroundImage: `radial-gradient(circle at 25% 25%, #f59e0b 1px, transparent 1px),
                           radial-gradient(circle at 75% 75%, #d97706 1px, transparent 1px)`,
          backgroundSize: '60px 60px'
        }} />
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <motion.div
          ref={ref}
          variants={containerVariants}
          initial="hidden"
          animate={inView ? "visible" : "hidden"}
          className="text-center mb-16"
        >
          <motion.div variants={itemVariants}>
            <div className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-amber-500/20 to-orange-500/20 border border-amber-500/30 rounded-full text-amber-300 text-sm font-medium mb-6">
              <span className="w-2 h-2 bg-amber-400 rounded-full mr-2 animate-pulse" />
              Comprehensive Platform Features
            </div>
            <h2 className="text-4xl md:text-6xl font-bold mb-6">
              <span className="bg-gradient-to-r from-amber-400 via-orange-400 to-amber-500 bg-clip-text text-transparent">
                Everything You Need
              </span>
            </h2>
            <p className="text-xl text-slate-300 max-w-4xl mx-auto leading-relaxed">
              From coffee commerce to DeFi trading, AI automation to enterprise infrastructure -
              discover the complete ecosystem that powers the future of coffee business
            </p>
          </motion.div>
        </motion.div>

        {/* Category Filter */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate={inView ? "visible" : "hidden"}
          className="flex flex-wrap justify-center gap-4 mb-16"
        >
          {categories.map((category, index) => (
            <motion.button
              key={category.id}
              variants={itemVariants}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setActiveCategory(category.id)}
              className={`px-6 py-3 rounded-full border transition-all duration-300 ${
                activeCategory === category.id
                  ? 'bg-gradient-to-r from-amber-500/30 to-orange-500/30 border-amber-500/50 text-amber-300'
                  : 'bg-slate-800/50 border-slate-700/50 text-slate-300 hover:border-amber-500/30 hover:text-amber-400'
              }`}
            >
              <span className="flex items-center gap-2">
                {category.icon} {category.label}
              </span>
            </motion.button>
          ))}
        </motion.div>

        {/* Features Grid */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate={inView ? "visible" : "hidden"}
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
        >
          {filteredFeatures.map((feature, index) => (
            <motion.div
              key={feature.title}
              variants={itemVariants}
              whileHover={{
                scale: 1.02,
                y: -5,
                transition: { duration: 0.2 }
              }}
              className="group relative"
            >
              <div className="h-full bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 hover:border-amber-500/30 hover:bg-slate-800/60 transition-all duration-300">
                {/* Status Indicator */}
                <div className="flex items-center justify-between mb-4">
                  <div className={`w-14 h-14 bg-gradient-to-r ${feature.gradient} rounded-xl flex items-center justify-center text-xl group-hover:scale-110 transition-transform duration-300`}>
                    {feature.icon}
                  </div>
                  <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${
                      feature.status === 'active' ? 'bg-green-400' : 'bg-yellow-400'
                    } animate-pulse`} />
                    <span className="text-xs text-slate-400 uppercase tracking-wide">
                      {feature.status}
                    </span>
                  </div>
                </div>

                {/* Title & Category */}
                <div className="mb-4">
                  <h3 className="text-lg font-bold text-white mb-1 group-hover:text-amber-400 transition-colors duration-300">
                    {feature.title}
                  </h3>
                  <span className="text-xs text-amber-400/70 uppercase tracking-wide font-medium">
                    {feature.category}
                  </span>
                </div>

                {/* Description */}
                <p className="text-slate-300 text-sm mb-4 leading-relaxed line-clamp-3">
                  {feature.description}
                </p>

                {/* Key Features */}
                <div className="mb-4">
                  <div className="flex flex-wrap gap-1">
                    {feature.details.slice(0, 3).map((detail, detailIndex) => (
                      <motion.span
                        key={detail}
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.8 }}
                        transition={{ delay: 0.3 + index * 0.05 + detailIndex * 0.02 }}
                        className="px-2 py-1 bg-slate-700/50 text-xs text-slate-300 rounded-md"
                      >
                        {detail}
                      </motion.span>
                    ))}
                    {feature.details.length > 3 && (
                      <span className="px-2 py-1 bg-amber-500/20 text-xs text-amber-300 rounded-md">
                        +{feature.details.length - 3} more
                      </span>
                    )}
                  </div>
                </div>

                {/* API Endpoints */}
                <div className="border-t border-slate-700/50 pt-3">
                  <div className="text-xs text-slate-400 mb-2">API Endpoints:</div>
                  <div className="space-y-1">
                    {feature.endpoints.slice(0, 2).map((endpoint, endpointIndex) => (
                      <div key={endpoint} className="flex items-center gap-2">
                        <div className="w-1 h-1 bg-amber-400 rounded-full" />
                        <code className="text-xs text-slate-300 font-mono">{endpoint}</code>
                      </div>
                    ))}
                    {feature.endpoints.length > 2 && (
                      <div className="text-xs text-amber-400">
                        +{feature.endpoints.length - 2} more endpoints
                      </div>
                    )}
                  </div>
                </div>

                {/* Hover Glow Effect */}
                <div className={`absolute inset-0 bg-gradient-to-r ${feature.gradient} opacity-0 group-hover:opacity-5 rounded-2xl transition-opacity duration-300`} />
              </div>
            </motion.div>
          ))}
        </motion.div>

        {/* Platform Statistics */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 0.8 }}
          className="mt-20 bg-gradient-to-r from-slate-800/30 to-slate-700/30 backdrop-blur-sm border border-slate-600/30 rounded-3xl p-8"
        >
          <div className="text-center mb-8">
            <h3 className="text-2xl font-bold text-white mb-2">
              🚀 Platform Overview
            </h3>
            <p className="text-slate-300">
              Comprehensive ecosystem powering the future of coffee commerce
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-6">
            {[
              { label: 'Microservices', value: '38+', icon: '🏗️' },
              { label: 'AI Agents', value: '25+', icon: '🤖' },
              { label: 'Blockchains', value: '4', icon: '🌐' },
              { label: 'Countries', value: '45+', icon: '🌍' },
              { label: 'Uptime', value: '99.9%', icon: '⚡' },
              { label: 'APIs', value: '200+', icon: '🔗' }
            ].map((stat, index) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.8 }}
                transition={{ delay: 1 + index * 0.1 }}
                className="text-center"
              >
                <div className="text-2xl mb-2">{stat.icon}</div>
                <div className="text-2xl font-bold text-amber-400 mb-1">{stat.value}</div>
                <div className="text-sm text-slate-400">{stat.label}</div>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Enhanced CTA */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 1.2 }}
          className="text-center mt-16"
        >
          <div className="bg-gradient-to-r from-amber-500/10 to-orange-500/10 border border-amber-500/20 rounded-2xl p-8 max-w-4xl mx-auto">
            <h3 className="text-3xl font-bold text-white mb-4">
              Ready to Experience the Future?
            </h3>
            <p className="text-slate-300 mb-6 text-lg">
              Join thousands of users already benefiting from our integrated coffee and crypto ecosystem
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
              <div className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-amber-500/20 to-orange-500/20 border border-amber-500/30 rounded-full text-amber-300 text-sm font-medium">
                <span className="w-2 h-2 bg-amber-400 rounded-full mr-2 animate-pulse" />
                All features ready • Launch in seconds
              </div>

              <div className="flex items-center gap-4 text-sm text-slate-400">
                <span className="flex items-center gap-1">
                  <div className="w-2 h-2 bg-green-400 rounded-full" />
                  Live Services
                </span>
                <span className="flex items-center gap-1">
                  <div className="w-2 h-2 bg-blue-400 rounded-full" />
                  Real-time Data
                </span>
                <span className="flex items-center gap-1">
                  <div className="w-2 h-2 bg-purple-400 rounded-full" />
                  AI Powered
                </span>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}
