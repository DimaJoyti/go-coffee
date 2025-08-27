'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'
import { DashboardView } from './Dashboard'
import { useConnectionStatus } from '@/contexts/RealTimeDataContext'
import MobileMenu, { MobileMenuItem, MobileMenuSection } from '@/components/ui/MobileMenu'
import { CompactPriceTicker } from '@/components/ui/PriceTicker'

interface HeaderProps {
  currentView: DashboardView
  onToggleSidebar: () => void
}

export default function Header({ currentView, onToggleSidebar }: HeaderProps) {
  const [showNotifications, setShowNotifications] = useState(false)
  const [showProfile, setShowProfile] = useState(false)
  const [showMobileMenu, setShowMobileMenu] = useState(false)
  const { isConnected } = useConnectionStatus()

  const getViewTitle = (view: DashboardView) => {
    const titles = {
      overview: 'Dashboard Overview',
      trading: 'Crypto Trading',
      coffee: 'Coffee Orders',
      portfolio: 'Portfolio Management',
      ai: 'AI Agents',
      analytics: 'Analytics & Reports'
    }
    return titles[view]
  }

  const notifications = [
    { id: 1, type: 'success', message: 'Trade executed successfully', time: '2m ago' },
    { id: 2, type: 'info', message: 'New coffee order received', time: '5m ago' },
    { id: 3, type: 'warning', message: 'AI agent needs attention', time: '10m ago' }
  ]

  return (
    <header className="bg-slate-900/50 backdrop-blur-xl border-b border-slate-700/50 px-6 py-4 sticky top-0 z-30">
      <div className="flex items-center justify-between">
        {/* Left Section */}
        <div className="flex items-center space-x-4">
          <motion.button
            onClick={() => setShowMobileMenu(true)}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="p-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/50 transition-all duration-200 md:hidden"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </motion.button>

          <div>
            <h1 className="text-2xl font-bold text-white">
              {getViewTitle(currentView)}
            </h1>
            <p className="text-slate-400 text-sm">
              {new Date().toLocaleDateString('en-US', { 
                weekday: 'long', 
                year: 'numeric', 
                month: 'long', 
                day: 'numeric' 
              })}
            </p>
          </div>
        </div>

        {/* Right Section */}
        <div className="flex items-center space-x-4">
          {/* Price Ticker */}
          <div className="hidden lg:block">
            <CompactPriceTicker />
          </div>

          {/* Search */}
          <div className="hidden md:block relative">
            <input
              type="text"
              placeholder="Search..."
              className="w-64 px-4 py-2 bg-slate-800/50 border border-slate-700/50 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:border-amber-500/50 transition-colors duration-200"
            />
            <div className="absolute right-3 top-1/2 transform -translate-y-1/2 text-slate-400">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>

          {/* Connection Status */}
          <div className={`flex items-center space-x-2 px-3 py-2 rounded-lg border ${
            isConnected
              ? 'bg-green-500/20 border-green-500/30'
              : 'bg-red-500/20 border-red-500/30'
          }`}>
            <div className={`w-2 h-2 rounded-full ${
              isConnected ? 'bg-green-400 animate-pulse' : 'bg-red-400'
            }`} />
            <span className={`text-sm font-medium ${
              isConnected ? 'text-green-400' : 'text-red-400'
            }`}>
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>

          {/* Notifications */}
          <div className="relative">
            <motion.button
              onClick={() => setShowNotifications(!showNotifications)}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="p-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/50 transition-all duration-200 relative"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-5 5v-5zM10.07 2.82l3.93 3.93-3.93 3.93-3.93-3.93 3.93-3.93z" />
              </svg>
              <div className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full flex items-center justify-center">
                <span className="text-xs text-white font-bold">3</span>
              </div>
            </motion.button>

            {/* Notifications Dropdown */}
            {showNotifications && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="absolute right-0 top-12 w-80 bg-slate-800/95 backdrop-blur-xl border border-slate-700/50 rounded-xl shadow-2xl z-50"
              >
                <div className="p-4 border-b border-slate-700/50">
                  <h3 className="text-lg font-semibold text-white">Notifications</h3>
                </div>
                <div className="max-h-64 overflow-y-auto">
                  {notifications.map((notification) => (
                    <div key={notification.id} className="p-4 border-b border-slate-700/30 hover:bg-slate-700/30 transition-colors duration-200">
                      <div className="flex items-start space-x-3">
                        <div className={`w-2 h-2 rounded-full mt-2 ${
                          notification.type === 'success' ? 'bg-green-400' :
                          notification.type === 'warning' ? 'bg-yellow-400' : 'bg-blue-400'
                        }`} />
                        <div className="flex-1">
                          <p className="text-white text-sm">{notification.message}</p>
                          <p className="text-slate-400 text-xs mt-1">{notification.time}</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="p-4">
                  <button className="w-full text-center text-amber-400 hover:text-amber-300 text-sm font-medium">
                    View All Notifications
                  </button>
                </div>
              </motion.div>
            )}
          </div>

          {/* Profile */}
          <div className="relative">
            <motion.button
              onClick={() => setShowProfile(!showProfile)}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="flex items-center space-x-3 p-2 rounded-lg hover:bg-slate-800/50 transition-all duration-200"
            >
              <div className="w-8 h-8 bg-gradient-to-r from-amber-500 to-orange-500 rounded-full flex items-center justify-center text-white font-bold">
                U
              </div>
              <div className="hidden md:block text-left">
                <p className="text-white text-sm font-medium">User</p>
                <p className="text-slate-400 text-xs">Premium</p>
              </div>
              <svg className="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </motion.button>

            {/* Profile Dropdown */}
            {showProfile && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="absolute right-0 top-12 w-64 bg-slate-800/95 backdrop-blur-xl border border-slate-700/50 rounded-xl shadow-2xl z-50"
              >
                <div className="p-4 border-b border-slate-700/50">
                  <div className="flex items-center space-x-3">
                    <div className="w-12 h-12 bg-gradient-to-r from-amber-500 to-orange-500 rounded-full flex items-center justify-center text-white font-bold text-lg">
                      U
                    </div>
                    <div>
                      <p className="text-white font-medium">User Account</p>
                      <p className="text-slate-400 text-sm">user@example.com</p>
                    </div>
                  </div>
                </div>
                <div className="p-2">
                  {[
                    { label: 'Profile Settings', icon: '👤' },
                    { label: 'Account Security', icon: '🔒' },
                    { label: 'Preferences', icon: '⚙️' },
                    { label: 'Help & Support', icon: '❓' }
                  ].map((item) => (
                    <button
                      key={item.label}
                      className="w-full flex items-center space-x-3 px-3 py-2 rounded-lg text-slate-300 hover:text-white hover:bg-slate-700/50 transition-all duration-200"
                    >
                      <span>{item.icon}</span>
                      <span className="text-sm">{item.label}</span>
                    </button>
                  ))}
                </div>
                <div className="p-2 border-t border-slate-700/50">
                  <button className="w-full flex items-center space-x-3 px-3 py-2 rounded-lg text-red-400 hover:text-red-300 hover:bg-red-500/10 transition-all duration-200">
                    <span>🚪</span>
                    <span className="text-sm">Sign Out</span>
                  </button>
                </div>
              </motion.div>
            )}
          </div>
        </div>
      </div>

      {/* Mobile Menu */}
      <MobileMenu isOpen={showMobileMenu} onClose={() => setShowMobileMenu(false)}>
        <MobileMenuSection title="Navigation">
          <MobileMenuItem icon="📊" label="Dashboard" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="📈" label="Trading" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="☕" label="Coffee Orders" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="💰" label="Portfolio" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="🤖" label="AI Agents" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="📊" label="Analytics" onClick={() => setShowMobileMenu(false)} />
        </MobileMenuSection>

        <MobileMenuSection title="Account">
          <MobileMenuItem icon="👤" label="Profile Settings" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="🔒" label="Security" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="⚙️" label="Preferences" onClick={() => setShowMobileMenu(false)} />
          <MobileMenuItem icon="❓" label="Help & Support" onClick={() => setShowMobileMenu(false)} />
        </MobileMenuSection>
      </MobileMenu>
    </header>
  )
}
