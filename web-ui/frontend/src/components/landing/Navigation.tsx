'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'

interface NavigationProps {
  onEnterDashboard: () => void
}

export default function Navigation({ onEnterDashboard }: NavigationProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  const navItems = [
    { name: 'Features', href: '#features' },
    { name: 'Trading', href: '#trading' },
    { name: 'AI Agents', href: '#ai' },
    { name: 'DeFi', href: '#defi' },
  ]

  return (
    <motion.nav
      initial={{ y: -100, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.8 }}
      className="fixed top-0 left-0 right-0 z-50 glass-card border-b border-coffee-500/20"
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <motion.div
            whileHover={{ scale: 1.05 }}
            className="flex items-center space-x-3"
          >
            <div className="w-10 h-10 bg-gradient-to-r from-coffee-500 to-brand-amber rounded-xl flex items-center justify-center text-xl font-bold shadow-coffee coffee-pulse">
              ☕
            </div>
            <div className="text-xl font-bold gradient-text font-display">
              Go Coffee
            </div>
          </motion.div>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center space-x-8">
            {navItems.map((item, index) => (
              <motion.a
                key={item.name}
                href={item.href}
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 + 0.3 }}
                whileHover={{ scale: 1.05 }}
                className="text-slate-300 hover:text-coffee-300 transition-all duration-300 font-medium hover:scale-105 hover-glow"
              >
                {item.name}
              </motion.a>
            ))}
          </div>

          {/* CTA Buttons */}
          <div className="hidden md:flex items-center space-x-4">
            <motion.button
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.6 }}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="btn-ghost px-4 py-2"
            >
              Connect Wallet
            </motion.button>
            <motion.button
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.7 }}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={onEnterDashboard}
              className="btn-primary"
            >
              Launch App
            </motion.button>
          </div>

          {/* Mobile Menu Button */}
          <div className="md:hidden">
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="text-slate-300 hover:text-white p-2"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                {isMenuOpen ? (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                ) : (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                )}
              </svg>
            </button>
          </div>
        </div>

        {/* Mobile Menu */}
        {isMenuOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="md:hidden border-t border-slate-700/50 py-4"
          >
            <div className="flex flex-col space-y-4">
              {navItems.map((item) => (
                <a
                  key={item.name}
                  href={item.href}
                  className="text-slate-300 hover:text-amber-400 transition-colors duration-200 font-medium px-2 py-1"
                  onClick={() => setIsMenuOpen(false)}
                >
                  {item.name}
                </a>
              ))}
              <div className="flex flex-col space-y-2 pt-4 border-t border-slate-700/50">
                <button className="text-left px-2 py-1 text-slate-300 hover:text-white transition-colors duration-200">
                  Connect Wallet
                </button>
                <button
                  onClick={onEnterDashboard}
                  className="text-left px-2 py-1 text-amber-400 hover:text-amber-300 transition-colors duration-200 font-semibold"
                >
                  Launch App
                </button>
              </div>
            </div>
          </motion.div>
        )}
      </div>
    </motion.nav>
  )
}
