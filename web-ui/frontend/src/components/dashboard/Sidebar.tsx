'use client'

import { motion } from 'framer-motion'
import { DashboardView } from './Dashboard'

interface SidebarProps {
  currentView: DashboardView
  onViewChange: (view: DashboardView) => void
  collapsed: boolean
  onToggleCollapse: () => void
  onBackToLanding: () => void
}

export default function Sidebar({
  currentView,
  onViewChange,
  collapsed,
  onToggleCollapse,
  onBackToLanding
}: SidebarProps) {
  const menuItems = [
    { id: 'overview' as DashboardView, label: 'Dashboard', icon: '📊' },
    { id: 'trading' as DashboardView, label: 'Trading', icon: '📈' },
    { id: 'coffee' as DashboardView, label: 'Coffee Orders', icon: '☕' },
    { id: 'portfolio' as DashboardView, label: 'Portfolio', icon: '💰' },
    { id: 'ai' as DashboardView, label: 'AI Agents', icon: '🤖' },
    { id: 'analytics' as DashboardView, label: 'Analytics', icon: '📊' }
  ]

  return (
    <motion.div
      initial={{ x: -280 }}
      animate={{ x: 0, width: collapsed ? 64 : 256 }}
      transition={{ duration: 0.3 }}
      className="fixed left-0 top-0 h-full glass-card border-r border-coffee-500/20 z-40 hidden md:block"
    >
      <div className="flex flex-col h-full">
        {/* Logo */}
        <div className="p-4 border-b border-coffee-500/20">
          <motion.div
            animate={{ justifyContent: collapsed ? 'center' : 'flex-start' }}
            className="flex items-center space-x-3"
          >
            <div className="w-10 h-10 bg-gradient-to-r from-coffee-500 to-brand-amber rounded-xl flex items-center justify-center text-xl font-bold flex-shrink-0 shadow-coffee coffee-pulse">
              ☕
            </div>
            {!collapsed && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="text-xl font-bold gradient-text font-display"
              >
                Go Coffee
              </motion.div>
            )}
          </motion.div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-2">
          {menuItems.map((item) => (
            <motion.button
              key={item.id}
              onClick={() => onViewChange(item.id)}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className={`sidebar-item ${
                currentView === item.id ? 'active' : ''
              }`}
            >
              <span className="text-xl flex-shrink-0">{item.icon}</span>
              {!collapsed && (
                <motion.span
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="font-medium"
                >
                  {item.label}
                </motion.span>
              )}
            </motion.button>
          ))}
        </nav>

        {/* Bottom Actions */}
        <div className="p-4 border-t border-coffee-500/20 space-y-2">
          {/* Settings */}
          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            className="sidebar-item"
          >
            <span className="text-xl flex-shrink-0">⚙️</span>
            {!collapsed && (
              <motion.span
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="font-medium"
              >
                Settings
              </motion.span>
            )}
          </motion.button>

          {/* Back to Landing */}
          <motion.button
            onClick={onBackToLanding}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            className="w-full flex items-center space-x-3 px-4 py-3 rounded-xl text-slate-300 hover:text-white hover:bg-slate-800/50 transition-all duration-200"
          >
            <span className="text-xl flex-shrink-0">🏠</span>
            {!collapsed && (
              <motion.span
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="font-medium"
              >
                Landing Page
              </motion.span>
            )}
          </motion.button>

          {/* Collapse Toggle */}
          <motion.button
            onClick={onToggleCollapse}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            className="w-full flex items-center justify-center px-4 py-3 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/50 transition-all duration-200"
          >
            <motion.span
              animate={{ rotate: collapsed ? 180 : 0 }}
              transition={{ duration: 0.3 }}
              className="text-xl"
            >
              ◀
            </motion.span>
          </motion.button>
        </div>
      </div>
    </motion.div>
  )
}
