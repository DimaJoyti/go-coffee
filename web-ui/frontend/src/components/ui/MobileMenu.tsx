'use client'

import { motion, AnimatePresence } from 'framer-motion'
import { useState, useEffect } from 'react'

interface MobileMenuProps {
  isOpen: boolean
  onClose: () => void
  children: React.ReactNode
}

export default function MobileMenu({ isOpen, onClose, children }: MobileMenuProps) {
  // Prevent body scroll when menu is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = 'unset'
    }

    return () => {
      document.body.style.overflow = 'unset'
    }
  }, [isOpen])

  // Close menu on escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose()
      }
    }

    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [isOpen, onClose])

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 md:hidden"
            onClick={onClose}
          />

          {/* Menu Panel */}
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ 
              type: 'spring', 
              damping: 25, 
              stiffness: 200,
              duration: 0.3 
            }}
            className="fixed top-0 right-0 h-full w-80 max-w-[85vw] glass-card border-l border-coffee-500/20 z-50 md:hidden overflow-y-auto"
          >
            {/* Close Button */}
            <div className="flex justify-end p-4">
              <motion.button
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
                onClick={onClose}
                className="p-2 rounded-xl text-slate-400 hover:text-coffee-300 hover:bg-coffee-500/10 transition-all duration-300"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </motion.button>
            </div>

            {/* Menu Content */}
            <div className="px-4 pb-4">
              {children}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

// Mobile-optimized menu item
interface MobileMenuItemProps {
  icon?: string
  label: string
  onClick?: () => void
  active?: boolean
  badge?: string | number
  className?: string
}

export function MobileMenuItem({ 
  icon, 
  label, 
  onClick, 
  active = false, 
  badge,
  className = '' 
}: MobileMenuItemProps) {
  return (
    <motion.button
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
      onClick={onClick}
      className={`
        w-full flex items-center justify-between p-4 rounded-xl transition-all duration-200
        ${active
          ? 'bg-gradient-to-r from-coffee-500/20 to-brand-amber/20 border border-coffee-500/30 text-coffee-300 shadow-glow'
          : 'text-slate-300 hover:text-coffee-300 hover:bg-coffee-500/10'
        }
        ${className}
      `}
    >
      <div className="flex items-center space-x-3">
        {icon && <span className="text-xl">{icon}</span>}
        <span className="font-medium">{label}</span>
      </div>
      
      {badge && (
        <span className="px-2 py-1 bg-coffee-500/20 text-coffee-400 rounded-full text-xs font-medium">
          {badge}
        </span>
      )}
    </motion.button>
  )
}

// Mobile menu section
interface MobileMenuSectionProps {
  title: string
  children: React.ReactNode
}

export function MobileMenuSection({ title, children }: MobileMenuSectionProps) {
  return (
    <div className="mb-6">
      <h3 className="text-slate-400 text-sm font-medium uppercase tracking-wider mb-3 px-2">
        {title}
      </h3>
      <div className="space-y-2">
        {children}
      </div>
    </div>
  )
}

// Mobile-optimized bottom sheet
interface BottomSheetProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
  height?: 'auto' | 'half' | 'full'
}

export function BottomSheet({ 
  isOpen, 
  onClose, 
  title, 
  children, 
  height = 'auto' 
}: BottomSheetProps) {
  const getHeightClass = () => {
    switch (height) {
      case 'half':
        return 'h-1/2'
      case 'full':
        return 'h-full'
      default:
        return 'h-auto max-h-[80vh]'
    }
  }

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40"
            onClick={onClose}
          />

          {/* Bottom Sheet */}
          <motion.div
            initial={{ y: '100%' }}
            animate={{ y: 0 }}
            exit={{ y: '100%' }}
            transition={{ 
              type: 'spring', 
              damping: 25, 
              stiffness: 200 
            }}
            className={`
              fixed bottom-0 left-0 right-0 glass-card
              border-t border-coffee-500/20 rounded-t-2xl z-50 overflow-hidden
              ${getHeightClass()}
            `}
          >
            {/* Handle */}
            <div className="flex justify-center py-3">
              <div className="w-12 h-1 bg-slate-600 rounded-full" />
            </div>

            {/* Header */}
            {title && (
              <div className="flex items-center justify-between px-6 pb-4 border-b border-slate-700/50">
                <h2 className="text-lg font-semibold text-white">{title}</h2>
                <motion.button
                  whileHover={{ scale: 1.1 }}
                  whileTap={{ scale: 0.9 }}
                  onClick={onClose}
                  className="p-2 rounded-xl text-slate-400 hover:text-coffee-300 hover:bg-coffee-500/10 transition-all duration-300"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </motion.button>
              </div>
            )}

            {/* Content */}
            <div className="p-6 overflow-y-auto">
              {children}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

// Responsive container that adapts to screen size
interface ResponsiveContainerProps {
  children: React.ReactNode
  className?: string
}

export function ResponsiveContainer({ children, className = '' }: ResponsiveContainerProps) {
  return (
    <div className={`
      px-4 sm:px-6 lg:px-8 
      max-w-7xl mx-auto
      ${className}
    `}>
      {children}
    </div>
  )
}

// Mobile-optimized card component
interface MobileCardProps {
  children: React.ReactNode
  className?: string
  padding?: 'sm' | 'md' | 'lg'
}

export function MobileCard({ children, className = '', padding = 'md' }: MobileCardProps) {
  const getPaddingClass = () => {
    switch (padding) {
      case 'sm':
        return 'p-3 sm:p-4'
      case 'lg':
        return 'p-6 sm:p-8'
      default:
        return 'p-4 sm:p-6'
    }
  }

  return (
    <div className={`
      glass-card
      ${getPaddingClass()}
      ${className}
    `}>
      {children}
    </div>
  )
}
