'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'
import { useState } from 'react'

export default function TechStack() {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.1
  })

  const [activeCategory, setActiveCategory] = useState('all')

  const categories = [
    { id: 'all', label: 'All Technologies', icon: '🌟' },
    { id: 'backend', label: 'Backend', icon: '⚙️' },
    { id: 'frontend', label: 'Frontend', icon: '🎨' },
    { id: 'blockchain', label: 'Blockchain', icon: '🌐' },
    { id: 'infrastructure', label: 'Infrastructure', icon: '🏗️' },
    { id: 'ai', label: 'AI & ML', icon: '🤖' }
  ]

  const technologies = [
    // Backend Technologies
    {
      name: 'Go',
      category: 'backend',
      description: 'High-performance backend services with Go 1.22+',
      icon: '🐹',
      version: '1.22+',
      usage: 'Core APIs, Microservices, gRPC',
      gradient: 'from-blue-400 to-cyan-500',
      popularity: 95
    },
    {
      name: 'Kafka',
      category: 'backend',
      description: 'Event streaming and message processing',
      icon: '📨',
      version: '3.0+',
      usage: 'Order Processing, Event Streaming',
      gradient: 'from-orange-400 to-red-500',
      popularity: 90
    },
    {
      name: 'gRPC',
      category: 'backend',
      description: 'High-performance inter-service communication',
      icon: '🔗',
      version: 'Latest',
      usage: 'Service Communication, APIs',
      gradient: 'from-green-400 to-emerald-500',
      popularity: 85
    },
    {
      name: 'PostgreSQL',
      category: 'backend',
      description: 'Primary database for transactional data',
      icon: '🐘',
      version: '15+',
      usage: 'Orders, Users, Transactions',
      gradient: 'from-blue-500 to-indigo-600',
      popularity: 92
    },

    // Frontend Technologies
    {
      name: 'Next.js',
      category: 'frontend',
      description: 'React framework with App Router and SSR',
      icon: '⚛️',
      version: '14+',
      usage: 'Web UI, Dashboard, Landing Pages',
      gradient: 'from-slate-600 to-slate-800',
      popularity: 98
    },
    {
      name: 'TypeScript',
      category: 'frontend',
      description: 'Type-safe JavaScript for better development',
      icon: '📘',
      version: '5.0+',
      usage: 'Frontend Development, Type Safety',
      gradient: 'from-blue-500 to-blue-700',
      popularity: 96
    },
    {
      name: 'TailwindCSS',
      category: 'frontend',
      description: 'Utility-first CSS framework',
      icon: '🎨',
      version: '3.0+',
      usage: 'Styling, Responsive Design',
      gradient: 'from-cyan-400 to-blue-500',
      popularity: 94
    },
    {
      name: 'Framer Motion',
      category: 'frontend',
      description: 'Production-ready motion library for React',
      icon: '✨',
      version: '10+',
      usage: 'Animations, Transitions',
      gradient: 'from-purple-400 to-pink-500',
      popularity: 88
    },

    // Blockchain Technologies
    {
      name: 'Ethereum',
      category: 'blockchain',
      description: 'Primary smart contract platform',
      icon: '💎',
      version: 'Mainnet',
      usage: 'Smart Contracts, DeFi Integration',
      gradient: 'from-purple-500 to-indigo-600',
      popularity: 95
    },
    {
      name: 'Solana',
      category: 'blockchain',
      description: 'High-performance blockchain network',
      icon: '🌞',
      version: 'Mainnet',
      usage: 'Fast Payments, Low Fees',
      gradient: 'from-purple-400 to-pink-500',
      popularity: 85
    },
    {
      name: 'Uniswap V3',
      category: 'blockchain',
      description: 'Decentralized exchange protocol',
      icon: '🦄',
      version: 'V3',
      usage: 'Token Swaps, Liquidity',
      gradient: 'from-pink-400 to-rose-500',
      popularity: 90
    },
    {
      name: 'Web3.js',
      category: 'blockchain',
      description: 'Ethereum JavaScript API',
      icon: '🌐',
      version: '4.0+',
      usage: 'Blockchain Interaction',
      gradient: 'from-orange-400 to-yellow-500',
      popularity: 87
    },

    // Infrastructure Technologies
    {
      name: 'Kubernetes',
      category: 'infrastructure',
      description: 'Container orchestration platform',
      icon: '☸️',
      version: '1.28+',
      usage: 'Service Deployment, Scaling',
      gradient: 'from-blue-500 to-cyan-600',
      popularity: 93
    },
    {
      name: 'Docker',
      category: 'infrastructure',
      description: 'Containerization platform',
      icon: '🐳',
      version: '24+',
      usage: 'Service Packaging, Deployment',
      gradient: 'from-blue-400 to-blue-600',
      popularity: 96
    },
    {
      name: 'Redis',
      category: 'infrastructure',
      description: 'In-memory data structure store with vector search',
      icon: '🔴',
      version: '8.0+',
      usage: 'Caching, Vector Search, Pub/Sub',
      gradient: 'from-red-500 to-red-700',
      popularity: 91
    },
    {
      name: 'Prometheus',
      category: 'infrastructure',
      description: 'Monitoring and alerting toolkit',
      icon: '📊',
      version: '2.45+',
      usage: 'Metrics Collection, Monitoring',
      gradient: 'from-orange-500 to-red-600',
      popularity: 89
    },

    // AI & ML Technologies
    {
      name: 'OpenAI GPT',
      category: 'ai',
      description: 'Large language models for AI agents',
      icon: '🧠',
      version: 'GPT-4',
      usage: 'AI Agents, Natural Language Processing',
      gradient: 'from-green-400 to-emerald-600',
      popularity: 97
    },
    {
      name: 'Vector Embeddings',
      category: 'ai',
      description: 'Semantic search and similarity matching',
      icon: '🔍',
      version: 'Latest',
      usage: 'Search, Recommendations',
      gradient: 'from-purple-500 to-violet-600',
      popularity: 85
    },
    {
      name: 'TensorFlow',
      category: 'ai',
      description: 'Machine learning framework',
      icon: '🤖',
      version: '2.13+',
      usage: 'Predictive Analytics, ML Models',
      gradient: 'from-orange-400 to-orange-600',
      popularity: 88
    },
    {
      name: 'Scikit-learn',
      category: 'ai',
      description: 'Machine learning library for Python',
      icon: '📈',
      version: '1.3+',
      usage: 'Data Analysis, ML Algorithms',
      gradient: 'from-blue-400 to-indigo-500',
      popularity: 86
    }
  ]

  const filteredTechnologies = activeCategory === 'all' 
    ? technologies 
    : technologies.filter(tech => tech.category === activeCategory)

  return (
    <section className="py-24 relative overflow-hidden">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0" style={{
          backgroundImage: `conic-gradient(from 0deg at 50% 50%, #f59e0b, #d97706, #f59e0b)`,
          backgroundSize: '100px 100px'
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
          <div className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-purple-500/20 to-pink-500/20 border border-purple-500/30 rounded-full text-purple-300 text-sm font-medium mb-6">
            <span className="w-2 h-2 bg-purple-400 rounded-full mr-2 animate-pulse" />
            Modern Technology Stack
          </div>
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="bg-gradient-to-r from-purple-400 via-pink-400 to-purple-500 bg-clip-text text-transparent">
              Cutting-Edge Tech
            </span>
          </h2>
          <p className="text-xl text-slate-300 max-w-4xl mx-auto leading-relaxed">
            Built with the latest and most reliable technologies for maximum performance, scalability, and developer experience
          </p>
        </motion.div>

        {/* Category Filter */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
          transition={{ delay: 0.3 }}
          className="flex flex-wrap justify-center gap-4 mb-16"
        >
          {categories.map((category) => (
            <motion.button
              key={category.id}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => setActiveCategory(category.id)}
              className={`px-6 py-3 rounded-full border transition-all duration-300 ${
                activeCategory === category.id
                  ? 'bg-gradient-to-r from-purple-500/30 to-pink-500/30 border-purple-500/50 text-purple-300'
                  : 'bg-slate-800/50 border-slate-700/50 text-slate-300 hover:border-purple-500/30 hover:text-purple-400'
              }`}
            >
              <span className="flex items-center gap-2">
                {category.icon} {category.label}
              </span>
            </motion.button>
          ))}
        </motion.div>

        {/* Technology Grid */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={inView ? { opacity: 1 } : { opacity: 0 }}
          transition={{ delay: 0.5 }}
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
        >
          {filteredTechnologies.map((tech, index) => (
            <motion.div
              key={tech.name}
              initial={{ opacity: 0, y: 50 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 50 }}
              transition={{ delay: 0.6 + index * 0.1 }}
              whileHover={{ 
                scale: 1.02,
                y: -5,
                transition: { duration: 0.2 }
              }}
              className="group relative"
            >
              <div className="h-full bg-slate-800/40 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 hover:border-purple-500/30 hover:bg-slate-800/60 transition-all duration-300">
                {/* Icon and Version */}
                <div className="flex items-center justify-between mb-4">
                  <div className={`w-12 h-12 bg-gradient-to-r ${tech.gradient} rounded-xl flex items-center justify-center text-xl group-hover:scale-110 transition-transform duration-300`}>
                    {tech.icon}
                  </div>
                  <span className="text-xs text-slate-400 bg-slate-700/50 px-2 py-1 rounded-md">
                    {tech.version}
                  </span>
                </div>

                {/* Name and Category */}
                <div className="mb-3">
                  <h3 className="text-lg font-bold text-white mb-1 group-hover:text-purple-400 transition-colors duration-300">
                    {tech.name}
                  </h3>
                  <span className="text-xs text-purple-400/70 uppercase tracking-wide font-medium">
                    {tech.category}
                  </span>
                </div>

                {/* Description */}
                <p className="text-slate-300 text-sm mb-4 leading-relaxed">
                  {tech.description}
                </p>

                {/* Usage */}
                <div className="mb-4">
                  <div className="text-xs text-slate-400 mb-2">Primary Usage:</div>
                  <div className="text-sm text-slate-300">{tech.usage}</div>
                </div>

                {/* Popularity Bar */}
                <div className="border-t border-slate-700/50 pt-3">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-slate-400">Adoption</span>
                    <span className="text-xs text-purple-400">{tech.popularity}%</span>
                  </div>
                  <div className="w-full bg-slate-700/50 rounded-full h-2">
                    <motion.div
                      initial={{ width: 0 }}
                      animate={inView ? { width: `${tech.popularity}%` } : { width: 0 }}
                      transition={{ delay: 0.8 + index * 0.05, duration: 1 }}
                      className={`h-2 bg-gradient-to-r ${tech.gradient} rounded-full`}
                    />
                  </div>
                </div>

                {/* Hover Glow */}
                <div className={`absolute inset-0 bg-gradient-to-r ${tech.gradient} opacity-0 group-hover:opacity-5 rounded-2xl transition-opacity duration-300`} />
              </div>
            </motion.div>
          ))}
        </motion.div>

        {/* Tech Stack Summary */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 1 }}
          className="mt-16 bg-gradient-to-r from-slate-800/30 to-slate-700/30 backdrop-blur-sm border border-slate-600/30 rounded-3xl p-8"
        >
          <div className="text-center mb-8">
            <h3 className="text-2xl font-bold text-white mb-2">
              🚀 Technology Highlights
            </h3>
            <p className="text-slate-300">
              Industry-leading technologies powering enterprise-grade performance
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            {[
              { label: 'Languages', value: '5+', icon: '💻' },
              { label: 'Frameworks', value: '15+', icon: '🏗️' },
              { label: 'Cloud Services', value: '10+', icon: '☁️' },
              { label: 'Integrations', value: '25+', icon: '🔗' }
            ].map((stat, index) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.8 }}
                transition={{ delay: 1.2 + index * 0.1 }}
                className="text-center"
              >
                <div className="text-2xl mb-2">{stat.icon}</div>
                <div className="text-2xl font-bold text-purple-400 mb-1">{stat.value}</div>
                <div className="text-sm text-slate-400">{stat.label}</div>
              </motion.div>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  )
}
