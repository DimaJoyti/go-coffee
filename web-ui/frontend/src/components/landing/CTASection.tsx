'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'

interface CTASectionProps {
  onEnterDashboard: () => void
}

export default function CTASection({ onEnterDashboard }: CTASectionProps) {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.3
  })

  return (
    <section className="py-24 relative overflow-hidden">
      {/* Background Effects */}
      <div className="absolute inset-0">
        <div className="absolute top-0 left-1/4 w-96 h-96 bg-amber-500/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-orange-500/10 rounded-full blur-3xl" />
      </div>

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center relative z-10">
        <motion.div
          ref={ref}
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ duration: 0.8 }}
          className="space-y-8"
        >
          {/* Main CTA */}
          <div className="space-y-6">
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.9 }}
              transition={{ delay: 0.2, duration: 0.8 }}
              className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-amber-500/20 to-orange-500/20 border border-amber-500/30 rounded-full text-amber-300 text-sm font-medium mb-6"
            >
              <span className="w-2 h-2 bg-amber-400 rounded-full mr-2 animate-pulse" />
              Join the Future of Coffee Trading
            </motion.div>

            <motion.h2
              initial={{ opacity: 0, y: 20 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
              transition={{ delay: 0.3, duration: 0.8 }}
              className="text-4xl md:text-6xl font-bold leading-tight"
            >
              <span className="bg-gradient-to-r from-white via-amber-100 to-orange-200 bg-clip-text text-transparent">
                Ready to Start Trading?
              </span>
            </motion.h2>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
              transition={{ delay: 0.4, duration: 0.8 }}
              className="text-xl text-slate-300 max-w-2xl mx-auto leading-relaxed"
            >
              Join thousands of traders and coffee enthusiasts who are already benefiting from our 
              revolutionary Web3 ecosystem. Start your journey today!
            </motion.p>
          </div>

          {/* Action Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
            transition={{ delay: 0.6, duration: 0.8 }}
            className="flex flex-col sm:flex-row gap-4 justify-center items-center"
          >
            <motion.button
              whileHover={{ 
                scale: 1.05, 
                boxShadow: "0 25px 50px rgba(245, 158, 11, 0.4)" 
              }}
              whileTap={{ scale: 0.95 }}
              onClick={onEnterDashboard}
              className="px-8 py-4 bg-gradient-to-r from-amber-500 to-orange-500 text-white text-lg font-bold rounded-xl shadow-2xl hover:from-amber-600 hover:to-orange-600 transition-all duration-300 min-w-[250px] group"
            >
              <span className="flex items-center justify-center">
                🚀 Launch Trading Platform
                <motion.span
                  animate={{ x: [0, 5, 0] }}
                  transition={{ duration: 1.5, repeat: Infinity }}
                  className="ml-2"
                >
                  →
                </motion.span>
              </span>
            </motion.button>
            
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="px-8 py-4 bg-slate-800/50 border border-slate-600/50 text-white text-lg font-semibold rounded-xl backdrop-blur-sm hover:bg-slate-700/50 transition-all duration-300 min-w-[250px]"
            >
              📖 Read Documentation
            </motion.button>
          </motion.div>

          {/* Features Grid */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
            transition={{ delay: 0.8, duration: 0.8 }}
            className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-16"
          >
            {[
              {
                icon: '⚡',
                title: 'Instant Setup',
                description: 'Get started in minutes with our streamlined onboarding process'
              },
              {
                icon: '🔒',
                title: 'Secure & Safe',
                description: 'Bank-grade security with multi-layer protection for your assets'
              },
              {
                icon: '🌍',
                title: 'Global Access',
                description: 'Trade from anywhere with 24/7 support and global infrastructure'
              }
            ].map((feature, index) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 20 }}
                animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
                transition={{ delay: 1 + index * 0.1, duration: 0.6 }}
                className="bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 text-center hover:border-amber-500/30 transition-all duration-300"
              >
                <div className="text-3xl mb-4">{feature.icon}</div>
                <h3 className="text-lg font-semibold text-white mb-2">{feature.title}</h3>
                <p className="text-slate-400 text-sm">{feature.description}</p>
              </motion.div>
            ))}
          </motion.div>

          {/* Social Proof */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
            transition={{ delay: 1.4, duration: 0.8 }}
            className="pt-12 border-t border-slate-700/50"
          >
            <p className="text-slate-400 mb-6">Trusted by traders worldwide</p>
            <div className="flex flex-wrap justify-center items-center gap-8 opacity-60">
              {[
                'Binance', 'Coinbase', 'Ethereum', 'Solana', 'Polygon'
              ].map((partner, index) => (
                <motion.div
                  key={partner}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.8 }}
                  transition={{ delay: 1.6 + index * 0.1 }}
                  className="px-4 py-2 bg-slate-800/50 border border-slate-700/50 rounded-lg text-slate-300 text-sm font-medium"
                >
                  {partner}
                </motion.div>
              ))}
            </div>
          </motion.div>
        </motion.div>
      </div>

      {/* Footer */}
      <motion.footer
        initial={{ opacity: 0 }}
        animate={inView ? { opacity: 1 } : { opacity: 0 }}
        transition={{ delay: 2, duration: 0.8 }}
        className="mt-24 pt-12 border-t border-slate-700/50 text-center"
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row justify-between items-center space-y-4 md:space-y-0">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-gradient-to-r from-amber-500 to-orange-500 rounded-lg flex items-center justify-center text-lg font-bold">
                ☕
              </div>
              <div className="text-lg font-bold bg-gradient-to-r from-amber-400 to-orange-400 bg-clip-text text-transparent">
                Go Coffee
              </div>
            </div>
            
            <div className="flex items-center space-x-6 text-slate-400 text-sm">
              <a href="#" className="hover:text-amber-400 transition-colors">Privacy Policy</a>
              <a href="#" className="hover:text-amber-400 transition-colors">Terms of Service</a>
              <a href="#" className="hover:text-amber-400 transition-colors">Support</a>
            </div>
            
            <div className="text-slate-400 text-sm">
              © 2024 Go Coffee. All rights reserved.
            </div>
          </div>
        </div>
      </motion.footer>
    </section>
  )
}
