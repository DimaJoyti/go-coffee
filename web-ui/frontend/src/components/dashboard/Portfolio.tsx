'use client'

import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'

export default function Portfolio() {
  const [timeframe, setTimeframe] = useState<'24h' | '7d' | '30d' | '1y'>('24h')
  const [portfolioValue, setPortfolioValue] = useState(123456.78)

  const holdings = [
    {
      symbol: 'BTC',
      name: 'Bitcoin',
      amount: 2.5432,
      value: 115234.56,
      change24h: 2.34,
      allocation: 45.2,
      icon: '₿'
    },
    {
      symbol: 'ETH',
      name: 'Ethereum',
      amount: 45.8901,
      value: 130234.12,
      change24h: 1.87,
      allocation: 35.8,
      icon: 'Ξ'
    },
    {
      symbol: 'SOL',
      name: 'Solana',
      amount: 234.5678,
      value: 23098.45,
      change24h: -0.56,
      allocation: 12.3,
      icon: '◎'
    },
    {
      symbol: 'COFFEE',
      name: 'Coffee Token',
      amount: 1000.0,
      value: 12340.00,
      change24h: 5.67,
      allocation: 6.7,
      icon: '☕'
    }
  ]

  const transactions = [
    {
      id: 1,
      type: 'buy',
      symbol: 'BTC',
      amount: 0.5,
      price: 45234.56,
      total: 22617.28,
      date: '2024-01-15 14:32:15',
      status: 'completed'
    },
    {
      id: 2,
      type: 'sell',
      symbol: 'ETH',
      amount: 2.5,
      price: 2834.12,
      total: 7085.30,
      date: '2024-01-15 13:45:22',
      status: 'completed'
    },
    {
      id: 3,
      type: 'buy',
      symbol: 'SOL',
      amount: 50.0,
      price: 98.45,
      total: 4922.50,
      date: '2024-01-15 12:18:45',
      status: 'pending'
    }
  ]

  // Simulate real-time portfolio updates
  useEffect(() => {
    const interval = setInterval(() => {
      setPortfolioValue(prev => prev + (Math.random() - 0.5) * 1000)
    }, 3000)

    return () => clearInterval(interval)
  }, [])

  const totalChange24h = holdings.reduce((acc, holding) => {
    return acc + (holding.value * holding.change24h / 100)
  }, 0)

  const totalChangePercent = (totalChange24h / portfolioValue) * 100

  return (
    <div className="space-y-6">
      {/* Portfolio Overview */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0 mb-6">
          <div>
            <h2 className="text-3xl font-bold text-white">Portfolio Overview</h2>
            <p className="text-slate-400">Your crypto and coffee token holdings</p>
          </div>
          
          <div className="flex items-center space-x-2">
            {(['24h', '7d', '30d', '1y'] as const).map((period) => (
              <button
                key={period}
                onClick={() => setTimeframe(period)}
                className={`px-4 py-2 rounded-lg font-medium transition-all duration-200 ${
                  timeframe === period
                    ? 'bg-amber-500 text-white'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                }`}
              >
                {period}
              </button>
            ))}
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="text-center">
            <p className="text-slate-400 text-sm mb-2">Total Portfolio Value</p>
            <p className="text-3xl font-bold text-white">
              ${portfolioValue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </p>
          </div>
          <div className="text-center">
            <p className="text-slate-400 text-sm mb-2">24h Change</p>
            <p className={`text-3xl font-bold ${totalChangePercent >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              {totalChangePercent >= 0 ? '+' : ''}${Math.abs(totalChange24h).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </p>
            <p className={`text-sm ${totalChangePercent >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              ({totalChangePercent >= 0 ? '+' : ''}{totalChangePercent.toFixed(2)}%)
            </p>
          </div>
          <div className="text-center">
            <p className="text-slate-400 text-sm mb-2">Total Assets</p>
            <p className="text-3xl font-bold text-white">{holdings.length}</p>
          </div>
        </div>
      </motion.div>

      {/* Holdings */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <h3 className="text-xl font-semibold text-white mb-6">Holdings</h3>
        <div className="space-y-4">
          {holdings.map((holding, index) => (
            <motion.div
              key={holding.symbol}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
              className="flex items-center justify-between p-4 bg-slate-700/30 rounded-lg hover:bg-slate-700/50 transition-colors duration-200"
            >
              <div className="flex items-center space-x-4">
                <div className="w-12 h-12 bg-gradient-to-r from-amber-500 to-orange-500 rounded-full flex items-center justify-center text-white text-xl font-bold">
                  {holding.icon}
                </div>
                <div>
                  <p className="text-white font-semibold">{holding.name}</p>
                  <p className="text-slate-400 text-sm">{holding.symbol}</p>
                </div>
              </div>

              <div className="text-right">
                <p className="text-white font-semibold">{holding.amount.toLocaleString()} {holding.symbol}</p>
                <p className="text-slate-400 text-sm">${holding.value.toLocaleString()}</p>
              </div>

              <div className="text-right">
                <p className={`font-semibold ${holding.change24h >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                  {holding.change24h >= 0 ? '+' : ''}{holding.change24h.toFixed(2)}%
                </p>
                <p className="text-slate-400 text-sm">{holding.allocation}% allocation</p>
              </div>

              <div className="w-16">
                <div className="w-full bg-slate-600/50 rounded-full h-2">
                  <div
                    className="h-2 bg-gradient-to-r from-amber-500 to-orange-500 rounded-full"
                    style={{ width: `${holding.allocation}%` }}
                  />
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </motion.div>

      {/* Portfolio Allocation Chart */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-xl font-semibold text-white mb-6">Asset Allocation</h3>
          <div className="h-64 flex items-center justify-center">
            <div className="text-center">
              <div className="text-4xl mb-2">📊</div>
              <p className="text-slate-400">Portfolio allocation chart</p>
              <p className="text-slate-500 text-sm mt-1">Interactive pie chart visualization</p>
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-xl font-semibold text-white mb-6">Performance Chart</h3>
          <div className="h-64 flex items-center justify-center">
            <div className="text-center">
              <div className="text-4xl mb-2">📈</div>
              <p className="text-slate-400">Portfolio performance over time</p>
              <p className="text-slate-500 text-sm mt-1">Historical value tracking</p>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Recent Transactions */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-xl font-semibold text-white">Recent Transactions</h3>
          <button className="text-amber-400 hover:text-amber-300 text-sm font-medium">
            View All
          </button>
        </div>
        
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-700/50">
                <th className="text-left text-slate-400 font-medium py-3">Type</th>
                <th className="text-left text-slate-400 font-medium py-3">Asset</th>
                <th className="text-left text-slate-400 font-medium py-3">Amount</th>
                <th className="text-left text-slate-400 font-medium py-3">Price</th>
                <th className="text-left text-slate-400 font-medium py-3">Total</th>
                <th className="text-left text-slate-400 font-medium py-3">Date</th>
                <th className="text-left text-slate-400 font-medium py-3">Status</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((tx, index) => (
                <motion.tr
                  key={tx.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                  className="border-b border-slate-700/30 hover:bg-slate-700/20 transition-colors duration-200"
                >
                  <td className="py-4">
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                      tx.type === 'buy' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
                    }`}>
                      {tx.type.toUpperCase()}
                    </span>
                  </td>
                  <td className="py-4 text-white font-medium">{tx.symbol}</td>
                  <td className="py-4 text-slate-300">{tx.amount}</td>
                  <td className="py-4 text-slate-300">${tx.price.toLocaleString()}</td>
                  <td className="py-4 text-white font-semibold">${tx.total.toLocaleString()}</td>
                  <td className="py-4 text-slate-400 text-sm">{tx.date}</td>
                  <td className="py-4">
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                      tx.status === 'completed' ? 'bg-green-500/20 text-green-400' :
                      tx.status === 'pending' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-red-500/20 text-red-400'
                    }`}>
                      {tx.status}
                    </span>
                  </td>
                </motion.tr>
              ))}
            </tbody>
          </table>
        </div>
      </motion.div>

      {/* Quick Actions */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <h3 className="text-xl font-semibold text-white mb-6">Quick Actions</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {[
            { label: 'Buy Crypto', icon: '💰', color: 'from-green-500 to-emerald-500' },
            { label: 'Sell Assets', icon: '💸', color: 'from-red-500 to-pink-500' },
            { label: 'Rebalance', icon: '⚖️', color: 'from-blue-500 to-purple-500' },
            { label: 'Export Data', icon: '📊', color: 'from-amber-500 to-orange-500' }
          ].map((action, index) => (
            <motion.button
              key={action.label}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className={`flex flex-col items-center space-y-2 p-4 bg-gradient-to-r ${action.color} bg-opacity-20 border border-opacity-30 rounded-xl hover:bg-opacity-30 transition-all duration-200`}
            >
              <span className="text-2xl">{action.icon}</span>
              <span className="text-white font-medium text-sm">{action.label}</span>
            </motion.button>
          ))}
        </div>
      </motion.div>
    </div>
  )
}
