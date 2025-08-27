'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'
import { useState, useEffect } from 'react'

export default function ServiceArchitecture() {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.1
  })

  const [selectedService, setSelectedService] = useState<string | null>(null)

  const services = [
    // Core Coffee Services
    {
      id: 'coffee-api',
      name: 'Coffee API',
      category: 'core',
      port: '3000',
      status: 'healthy',
      description: 'Main coffee ordering and management API',
      endpoints: ['POST /orders', 'GET /menu', 'GET /inventory'],
      position: { x: 20, y: 30 }
    },
    {
      id: 'order-processor',
      name: 'Order Processor',
      category: 'core',
      port: '3001',
      status: 'healthy',
      description: 'Kafka-based order processing service',
      endpoints: ['Kafka Consumer', 'Order Validation', 'Payment Processing'],
      position: { x: 50, y: 20 }
    },
    {
      id: 'inventory-service',
      name: 'Inventory Service',
      category: 'core',
      port: '3002',
      status: 'healthy',
      description: 'Real-time inventory tracking and management',
      endpoints: ['GET /stock', 'POST /restock', 'GET /suppliers'],
      position: { x: 80, y: 30 }
    },

    // Web3 & DeFi Services
    {
      id: 'web3-payments',
      name: 'Web3 Payments',
      category: 'web3',
      port: '8083',
      status: 'healthy',
      description: 'Multi-chain cryptocurrency payment processing',
      endpoints: ['POST /payment/create', 'GET /payment/status', 'GET /wallet/balance'],
      position: { x: 15, y: 60 }
    },
    {
      id: 'defi-trading',
      name: 'DeFi Trading',
      category: 'web3',
      port: '8084',
      status: 'healthy',
      description: 'Automated DeFi trading and yield farming',
      endpoints: ['POST /defi/trade', 'GET /defi/yield', 'GET /defi/strategies'],
      position: { x: 45, y: 70 }
    },
    {
      id: 'token-service',
      name: 'Coffee Token',
      category: 'web3',
      port: '8085',
      status: 'healthy',
      description: 'COFFEE token management and staking',
      endpoints: ['GET /token/price', 'POST /token/stake', 'GET /token/rewards'],
      position: { x: 75, y: 60 }
    },

    // AI & Analytics Services
    {
      id: 'ai-orchestrator',
      name: 'AI Orchestrator',
      category: 'ai',
      port: '8094',
      status: 'healthy',
      description: 'Central AI agent coordination and management',
      endpoints: ['GET /ai/agents', 'POST /ai/strategy', 'GET /ai/performance'],
      position: { x: 25, y: 90 }
    },
    {
      id: 'ai-search',
      name: 'AI Search Engine',
      category: 'ai',
      port: '8095',
      status: 'healthy',
      description: 'Redis 8 powered semantic search with vector embeddings',
      endpoints: ['POST /ai-search/semantic', 'POST /ai-search/vector', 'GET /ai-search/trending'],
      position: { x: 55, y: 85 }
    },
    {
      id: 'analytics-service',
      name: 'Analytics Service',
      category: 'ai',
      port: '8096',
      status: 'healthy',
      description: 'Advanced analytics and business intelligence',
      endpoints: ['GET /analytics/dashboard', 'POST /analytics/report', 'GET /analytics/kpis'],
      position: { x: 85, y: 90 }
    },

    // Enterprise Services
    {
      id: 'api-gateway',
      name: 'API Gateway',
      category: 'enterprise',
      port: '8080',
      status: 'healthy',
      description: 'Central API gateway with rate limiting and auth',
      endpoints: ['All Routes', 'Rate Limiting', 'Authentication'],
      position: { x: 50, y: 10 }
    },
    {
      id: 'monitoring',
      name: 'Monitoring',
      category: 'enterprise',
      port: '9090',
      status: 'healthy',
      description: 'Prometheus metrics and Grafana dashboards',
      endpoints: ['GET /metrics', 'Prometheus', 'Grafana'],
      position: { x: 10, y: 10 }
    },
    {
      id: 'redis-cluster',
      name: 'Redis Cluster',
      category: 'enterprise',
      port: '6379',
      status: 'healthy',
      description: 'High-performance caching and data storage',
      endpoints: ['Redis Commands', 'Vector Search', 'Pub/Sub'],
      position: { x: 90, y: 10 }
    }
  ]

  const categoryColors = {
    core: 'from-amber-500 to-orange-500',
    web3: 'from-green-500 to-emerald-500',
    ai: 'from-blue-500 to-purple-500',
    enterprise: 'from-slate-500 to-gray-600'
  }

  const categoryLabels = {
    core: 'Coffee Services',
    web3: 'Web3 & DeFi',
    ai: 'AI & Analytics',
    enterprise: 'Enterprise'
  }

  return (
    <section className="py-24 relative overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0" style={{
          backgroundImage: `linear-gradient(90deg, #f59e0b 1px, transparent 1px),
                           linear-gradient(0deg, #f59e0b 1px, transparent 1px)`,
          backgroundSize: '40px 40px'
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
          <div className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-blue-500/20 to-purple-500/20 border border-blue-500/30 rounded-full text-blue-300 text-sm font-medium mb-6">
            <span className="w-2 h-2 bg-blue-400 rounded-full mr-2 animate-pulse" />
            Microservices Architecture
          </div>
          <h2 className="text-4xl md:text-6xl font-bold mb-6">
            <span className="bg-gradient-to-r from-blue-400 via-purple-400 to-blue-500 bg-clip-text text-transparent">
              Service Ecosystem
            </span>
          </h2>
          <p className="text-xl text-slate-300 max-w-4xl mx-auto leading-relaxed">
            Explore our cloud-native microservices architecture powering the entire Go Coffee platform
          </p>
        </motion.div>

        {/* Service Categories Legend */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
          transition={{ delay: 0.3 }}
          className="flex flex-wrap justify-center gap-4 mb-12"
        >
          {Object.entries(categoryLabels).map(([key, label]) => (
            <div key={key} className="flex items-center gap-2 px-4 py-2 bg-slate-800/50 rounded-full border border-slate-700/50">
              <div className={`w-3 h-3 bg-gradient-to-r ${categoryColors[key as keyof typeof categoryColors]} rounded-full`} />
              <span className="text-sm text-slate-300">{label}</span>
            </div>
          ))}
        </motion.div>

        {/* Interactive Service Map */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
          transition={{ delay: 0.5 }}
          className="relative bg-slate-900/50 backdrop-blur-sm border border-slate-700/50 rounded-3xl p-8 min-h-[600px]"
        >
          {/* Service Nodes */}
          {services.map((service, index) => (
            <motion.div
              key={service.id}
              initial={{ opacity: 0, scale: 0 }}
              animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0 }}
              transition={{ delay: 0.7 + index * 0.1 }}
              className="absolute cursor-pointer group"
              style={{
                left: `${service.position.x}%`,
                top: `${service.position.y}%`,
                transform: 'translate(-50%, -50%)'
              }}
              onClick={() => setSelectedService(selectedService === service.id ? null : service.id)}
            >
              <div className={`relative p-4 bg-gradient-to-r ${categoryColors[service.category as keyof typeof categoryColors]} rounded-xl shadow-lg group-hover:scale-110 transition-all duration-300 ${
                selectedService === service.id ? 'ring-2 ring-amber-400 scale-110' : ''
              }`}>
                <div className="text-white font-bold text-sm text-center min-w-[100px]">
                  {service.name}
                </div>
                <div className="text-white/80 text-xs text-center mt-1">
                  :{service.port}
                </div>
                
                {/* Status Indicator */}
                <div className="absolute -top-1 -right-1 w-4 h-4 bg-green-400 rounded-full border-2 border-white animate-pulse" />
              </div>

              {/* Connection Lines - simplified for now */}
              {service.category === 'core' && (
                <div className="absolute top-1/2 left-full w-8 h-0.5 bg-amber-400/50 transform -translate-y-1/2" />
              )}
            </motion.div>
          ))}

          {/* Service Details Panel */}
          {selectedService && (
            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              className="absolute right-4 top-4 w-80 bg-slate-800/90 backdrop-blur-sm border border-slate-600/50 rounded-2xl p-6"
            >
              {(() => {
                const service = services.find(s => s.id === selectedService)
                if (!service) return null
                
                return (
                  <>
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="text-lg font-bold text-white">{service.name}</h3>
                      <button
                        onClick={() => setSelectedService(null)}
                        className="text-slate-400 hover:text-white"
                      >
                        ✕
                      </button>
                    </div>
                    
                    <div className="space-y-3">
                      <div>
                        <span className="text-sm text-slate-400">Port:</span>
                        <span className="text-amber-400 ml-2">:{service.port}</span>
                      </div>
                      
                      <div>
                        <span className="text-sm text-slate-400">Status:</span>
                        <span className="text-green-400 ml-2 flex items-center gap-1">
                          <div className="w-2 h-2 bg-green-400 rounded-full" />
                          {service.status}
                        </span>
                      </div>
                      
                      <div>
                        <span className="text-sm text-slate-400 block mb-2">Description:</span>
                        <p className="text-slate-300 text-sm">{service.description}</p>
                      </div>
                      
                      <div>
                        <span className="text-sm text-slate-400 block mb-2">Key Endpoints:</span>
                        <div className="space-y-1">
                          {service.endpoints.map((endpoint, idx) => (
                            <div key={idx} className="text-xs text-slate-300 font-mono bg-slate-700/50 px-2 py-1 rounded">
                              {endpoint}
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  </>
                )
              })()}
            </motion.div>
          )}
        </motion.div>

        {/* Architecture Stats */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 1 }}
          className="mt-12 grid grid-cols-2 md:grid-cols-4 gap-6"
        >
          {[
            { label: 'Microservices', value: '38+', icon: '🏗️' },
            { label: 'API Endpoints', value: '200+', icon: '🔗' },
            { label: 'Health Checks', value: '100%', icon: '💚' },
            { label: 'Response Time', value: '<50ms', icon: '⚡' }
          ].map((stat, index) => (
            <motion.div
              key={stat.label}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.8 }}
              transition={{ delay: 1.2 + index * 0.1 }}
              className="text-center bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-xl p-4"
            >
              <div className="text-2xl mb-2">{stat.icon}</div>
              <div className="text-2xl font-bold text-blue-400 mb-1">{stat.value}</div>
              <div className="text-sm text-slate-400">{stat.label}</div>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
