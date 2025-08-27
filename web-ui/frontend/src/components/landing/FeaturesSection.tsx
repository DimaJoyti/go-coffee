'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'

export default function FeaturesSection() {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.1
  })

  const features = [
    {
      icon: '☕',
      title: 'Coffee Commerce',
      description: 'Order premium coffee from global suppliers with real-time inventory tracking and quality assurance.',
      details: ['Real-time inventory', 'Quality tracking', 'Global suppliers', 'Smart contracts'],
      gradient: 'from-amber-500 to-orange-500'
    },
    {
      icon: '💰',
      title: 'DeFi Trading',
      description: 'Advanced cryptocurrency trading with multi-chain support, yield farming, and automated strategies.',
      details: ['Multi-chain support', 'Yield farming', 'Auto strategies', 'Risk management'],
      gradient: 'from-green-500 to-emerald-500'
    },
    {
      icon: '🤖',
      title: 'AI Automation',
      description: 'Intelligent agents handle trading, inventory management, and market analysis automatically.',
      details: ['Smart trading bots', 'Market analysis', 'Inventory AI', 'Risk assessment'],
      gradient: 'from-blue-500 to-purple-500'
    },
    {
      icon: '📊',
      title: 'Advanced Analytics',
      description: 'Comprehensive dashboards with real-time data, performance metrics, and predictive insights.',
      details: ['Real-time data', 'Performance metrics', 'Predictive AI', 'Custom reports'],
      gradient: 'from-purple-500 to-pink-500'
    },
    {
      icon: '🔒',
      title: 'Enterprise Security',
      description: 'Bank-grade security with multi-factor authentication, encryption, and compliance monitoring.',
      details: ['Multi-factor auth', 'End-to-end encryption', 'Compliance tools', 'Audit trails'],
      gradient: 'from-red-500 to-orange-500'
    },
    {
      icon: '🌐',
      title: 'Global Infrastructure',
      description: 'Distributed architecture with 99.9% uptime, auto-scaling, and disaster recovery.',
      details: ['99.9% uptime', 'Auto-scaling', 'Global CDN', 'Disaster recovery'],
      gradient: 'from-cyan-500 to-blue-500'
    }
  ]

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
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
    <section id="features" className="py-24 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          ref={ref}
          variants={containerVariants}
          initial="hidden"
          animate={inView ? "visible" : "hidden"}
          className="text-center mb-16"
        >
          <motion.div variants={itemVariants}>
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              <span className="bg-gradient-to-r from-amber-400 to-orange-400 bg-clip-text text-transparent">
                Powerful Features
              </span>
            </h2>
            <p className="text-xl text-slate-300 max-w-3xl mx-auto">
              Everything you need to succeed in the modern coffee and crypto trading ecosystem
            </p>
          </motion.div>
        </motion.div>

        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate={inView ? "visible" : "hidden"}
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8"
        >
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              variants={itemVariants}
              whileHover={{ 
                scale: 1.05,
                transition: { duration: 0.2 }
              }}
              className="group relative"
            >
              <div className="h-full bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-8 hover:border-slate-600/50 transition-all duration-300">
                {/* Icon */}
                <div className={`w-16 h-16 bg-gradient-to-r ${feature.gradient} rounded-2xl flex items-center justify-center text-2xl mb-6 group-hover:scale-110 transition-transform duration-300`}>
                  {feature.icon}
                </div>

                {/* Title */}
                <h3 className="text-2xl font-bold text-white mb-4 group-hover:text-amber-400 transition-colors duration-300">
                  {feature.title}
                </h3>

                {/* Description */}
                <p className="text-slate-300 mb-6 leading-relaxed">
                  {feature.description}
                </p>

                {/* Details */}
                <ul className="space-y-2">
                  {feature.details.map((detail, detailIndex) => (
                    <motion.li
                      key={detail}
                      initial={{ opacity: 0, x: -20 }}
                      animate={inView ? { opacity: 1, x: 0 } : { opacity: 0, x: -20 }}
                      transition={{ delay: 0.5 + index * 0.1 + detailIndex * 0.05 }}
                      className="flex items-center text-sm text-slate-400"
                    >
                      <div className={`w-2 h-2 bg-gradient-to-r ${feature.gradient} rounded-full mr-3 flex-shrink-0`} />
                      {detail}
                    </motion.li>
                  ))}
                </ul>

                {/* Hover Effect */}
                <div className={`absolute inset-0 bg-gradient-to-r ${feature.gradient} opacity-0 group-hover:opacity-5 rounded-2xl transition-opacity duration-300`} />
              </div>
            </motion.div>
          ))}
        </motion.div>

        {/* Bottom CTA */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ delay: 1 }}
          className="text-center mt-16"
        >
          <div className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-amber-500/20 to-orange-500/20 border border-amber-500/30 rounded-full text-amber-300 text-sm font-medium">
            <span className="w-2 h-2 bg-amber-400 rounded-full mr-2 animate-pulse" />
            Ready to explore all features? Launch the platform now!
          </div>
        </motion.div>
      </div>
    </section>
  )
}
