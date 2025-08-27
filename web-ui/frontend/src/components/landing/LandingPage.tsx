'use client'

import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'
import HeroSection from './HeroSection'
import FeaturesSection from './FeaturesSection'
import StatsSection from './StatsSection'
import TradingPreview from './TradingPreview'
import CTASection from './CTASection'
import Navigation from './Navigation'
import PriceTicker from '@/components/ui/PriceTicker'

interface LandingPageProps {
  onEnterDashboard: () => void
}

export default function LandingPage({ onEnterDashboard }: LandingPageProps) {
  const [scrollY, setScrollY] = useState(0)

  useEffect(() => {
    const handleScroll = () => setScrollY(window.scrollY)
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white overflow-x-hidden">
      {/* Navigation */}
      <Navigation onEnterDashboard={onEnterDashboard} />

      {/* Price Ticker */}
      <div className="pt-16">
        <PriceTicker className="mx-4 mt-4" />
      </div>

      {/* Hero Section */}
      <HeroSection onEnterDashboard={onEnterDashboard} scrollY={scrollY} />
      
      {/* Features Section */}
      <FeaturesSection />
      
      {/* Stats Section */}
      <StatsSection />
      
      {/* Trading Preview */}
      <TradingPreview />
      
      {/* CTA Section */}
      <CTASection onEnterDashboard={onEnterDashboard} />
      
      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-amber-500/10 rounded-full blur-3xl animate-pulse" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-orange-500/10 rounded-full blur-3xl animate-pulse delay-1000" />
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-gradient-radial from-amber-500/5 to-transparent rounded-full" />
      </div>
    </div>
  )
}
