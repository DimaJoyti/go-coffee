'use client'

import { motion } from 'framer-motion'
import { useInView } from 'react-intersection-observer'
import { useState, useEffect } from 'react'

export default function TradingPreview() {
  const [ref, inView] = useInView({
    triggerOnce: true,
    threshold: 0.2
  })

  const [activeTab, setActiveTab] = useState('trading')
  const [cryptoPrices, setCryptoPrices] = useState({
    BTC: 45234.56,
    ETH: 2834.12,
    SOL: 98.45,
    COFFEE: 12.34
  })

  // Simulate real-time price updates
  useEffect(() => {
    const interval = setInterval(() => {
      setCryptoPrices(prev => ({
        BTC: prev.BTC + (Math.random() - 0.5) * 100,
        ETH: prev.ETH + (Math.random() - 0.5) * 50,
        SOL: prev.SOL + (Math.random() - 0.5) * 2,
        COFFEE: prev.COFFEE + (Math.random() - 0.5) * 0.5
      }))
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  const tabs = [
    { id: 'trading', label: 'Crypto Trading', icon: '📈' },
    { id: 'coffee', label: 'Coffee Orders', icon: '☕' },
    { id: 'ai', label: 'AI Agents', icon: '🤖' }
  ]

  return (
    <section id="trading" className="py-24 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          ref={ref}
          initial={{ opacity: 0, y: 30 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
          transition={{ duration: 0.8 }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-5xl font-bold mb-6">
            <span className="bg-gradient-to-r from-amber-400 to-orange-400 bg-clip-text text-transparent">
              Platform Preview
            </span>
          </h2>
          <p className="text-xl text-slate-300 max-w-3xl mx-auto">
            Get a glimpse of our powerful trading interface and ecosystem features
          </p>
        </motion.div>

        {/* Tab Navigation */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
          transition={{ delay: 0.3, duration: 0.6 }}
          className="flex justify-center mb-12"
        >
          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-2">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`px-6 py-3 rounded-xl font-semibold transition-all duration-300 ${
                  activeTab === tab.id
                    ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-lg'
                    : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
                }`}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </div>
        </motion.div>

        {/* Preview Content */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
          transition={{ delay: 0.5, duration: 0.8 }}
          className="bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-3xl p-8 max-w-6xl mx-auto"
        >
          {activeTab === 'trading' && <TradingInterface prices={cryptoPrices} />}
          {activeTab === 'coffee' && <CoffeeInterface />}
          {activeTab === 'ai' && <AIInterface />}
        </motion.div>
      </div>
    </section>
  )
}

function TradingInterface({ prices }: { prices: any }) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between mb-6">
        <h3 className="text-2xl font-bold text-white">Trading Dashboard</h3>
        <div className="flex items-center space-x-2 text-green-400">
          <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
          <span className="text-sm">Live Market Data</span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {Object.entries(prices).map(([symbol, price]) => (
          <div key={symbol} className="bg-slate-700/50 rounded-xl p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-300 font-medium">{symbol}</span>
              <span className="text-green-400 text-sm">+2.34%</span>
            </div>
            <div className="text-xl font-bold text-white">
              ${typeof price === 'number' ? price.toFixed(2) : price}
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-slate-700/30 rounded-xl p-6">
          <h4 className="text-lg font-semibold text-white mb-4">Order Book</h4>
          <div className="space-y-2">
            {[
              { price: '45,250.00', amount: '0.5432', side: 'sell' },
              { price: '45,240.00', amount: '1.2345', side: 'sell' },
              { price: '45,230.00', amount: '0.8765', side: 'buy' },
              { price: '45,220.00', amount: '2.1234', side: 'buy' }
            ].map((order, index) => (
              <div key={index} className="flex justify-between text-sm">
                <span className={order.side === 'buy' ? 'text-green-400' : 'text-red-400'}>
                  ${order.price}
                </span>
                <span className="text-slate-300">{order.amount}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-slate-700/30 rounded-xl p-6">
          <h4 className="text-lg font-semibold text-white mb-4">Recent Trades</h4>
          <div className="space-y-2">
            {[
              { price: '45,234.56', amount: '0.1234', time: '14:32:15' },
              { price: '45,230.00', amount: '0.5678', time: '14:32:10' },
              { price: '45,235.12', amount: '0.2345', time: '14:32:05' },
              { price: '45,240.00', amount: '0.8901', time: '14:32:00' }
            ].map((trade, index) => (
              <div key={index} className="flex justify-between text-sm">
                <span className="text-green-400">${trade.price}</span>
                <span className="text-slate-300">{trade.amount}</span>
                <span className="text-slate-400">{trade.time}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function CoffeeInterface() {
  return (
    <div className="space-y-6">
      <h3 className="text-2xl font-bold text-white mb-6">Coffee Marketplace</h3>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {[
          { name: 'Ethiopian Yirgacheffe', price: '$24.99', rating: 4.8, stock: 45 },
          { name: 'Colombian Supremo', price: '$19.99', rating: 4.6, stock: 32 },
          { name: 'Jamaican Blue Mountain', price: '$89.99', rating: 4.9, stock: 12 }
        ].map((coffee, index) => (
          <div key={index} className="bg-slate-700/50 rounded-xl p-4">
            <div className="w-full h-32 bg-gradient-to-br from-amber-500/20 to-orange-500/20 rounded-lg mb-4 flex items-center justify-center text-4xl">
              ☕
            </div>
            <h4 className="font-semibold text-white mb-2">{coffee.name}</h4>
            <div className="flex justify-between items-center mb-2">
              <span className="text-amber-400 font-bold">{coffee.price}</span>
              <span className="text-yellow-400">★ {coffee.rating}</span>
            </div>
            <div className="text-sm text-slate-400">Stock: {coffee.stock} bags</div>
          </div>
        ))}
      </div>
    </div>
  )
}

function AIInterface() {
  return (
    <div className="space-y-6">
      <h3 className="text-2xl font-bold text-white mb-6">AI Agent Network</h3>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {[
          { name: 'Trading Bot Alpha', status: 'Active', profit: '+$1,234', trades: 45 },
          { name: 'Market Analyzer', status: 'Active', profit: '+$892', trades: 23 },
          { name: 'Risk Manager', status: 'Monitoring', profit: '+$456', trades: 12 },
          { name: 'Arbitrage Hunter', status: 'Active', profit: '+$2,103', trades: 67 }
        ].map((agent, index) => (
          <div key={index} className="bg-slate-700/50 rounded-xl p-4">
            <div className="flex items-center justify-between mb-3">
              <h4 className="font-semibold text-white">{agent.name}</h4>
              <span className={`px-2 py-1 rounded-full text-xs ${
                agent.status === 'Active' ? 'bg-green-500/20 text-green-400' : 'bg-yellow-500/20 text-yellow-400'
              }`}>
                {agent.status}
              </span>
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-slate-400">24h Profit</div>
                <div className="text-green-400 font-semibold">{agent.profit}</div>
              </div>
              <div>
                <div className="text-slate-400">Trades</div>
                <div className="text-white font-semibold">{agent.trades}</div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
