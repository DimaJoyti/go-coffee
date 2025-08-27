'use client'

import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'
import { usePrices, useTrades } from '@/contexts/RealTimeDataContext'

export default function TradingView() {
  const [selectedPair, setSelectedPair] = useState('BTC/USDT')
  const [orderType, setOrderType] = useState<'market' | 'limit'>('limit')
  const [orderSide, setOrderSide] = useState<'buy' | 'sell'>('buy')

  const { prices, getPrice } = usePrices()
  const recentTrades = useTrades(selectedPair, 10)

  // Get current market data for selected pair
  const currentPrice = getPrice(selectedPair)
  const marketData = {
    price: currentPrice?.price || 45234.56,
    change24h: currentPrice?.change24h || 2.34,
    volume: currentPrice?.volume || 1234567890,
    high24h: (currentPrice?.price || 45234.56) * 1.05,
    low24h: (currentPrice?.price || 45234.56) * 0.95
  }

  // Convert prices to trading pairs format
  const tradingPairs = prices.map(priceData => ({
    pair: priceData.symbol,
    price: priceData.price,
    change: priceData.change24h
  }))

  return (
    <div className="space-y-6">
      {/* Market Header */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0">
          <div className="flex items-center space-x-6">
            <div>
              <h2 className="text-2xl font-bold text-white">{selectedPair}</h2>
              <p className="text-slate-400">Crypto Trading Terminal</p>
            </div>
            <div className="flex items-center space-x-2 text-green-400">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
              <span className="text-sm">Live Data</span>
            </div>
          </div>
          
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <div>
              <p className="text-slate-400 text-sm">Price</p>
              <p className="text-white font-bold">${marketData.price.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-slate-400 text-sm">24h Change</p>
              <p className={`font-bold ${marketData.change24h >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                {marketData.change24h >= 0 ? '+' : ''}{marketData.change24h.toFixed(2)}%
              </p>
            </div>
            <div>
              <p className="text-slate-400 text-sm">24h High</p>
              <p className="text-white font-bold">${marketData.high24h.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-slate-400 text-sm">24h Low</p>
              <p className="text-white font-bold">${marketData.low24h.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-slate-400 text-sm">Volume</p>
              <p className="text-white font-bold">${(marketData.volume / 1000000).toFixed(1)}M</p>
            </div>
          </div>
        </div>
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Trading Pairs */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-lg font-semibold text-white mb-4">Trading Pairs</h3>
          <div className="space-y-2">
            {tradingPairs.map((pair) => (
              <button
                key={pair.pair}
                onClick={() => setSelectedPair(pair.pair)}
                className={`w-full flex items-center justify-between p-3 rounded-lg transition-all duration-200 ${
                  selectedPair === pair.pair
                    ? 'bg-amber-500/20 border border-amber-500/30'
                    : 'hover:bg-slate-700/50'
                }`}
              >
                <div className="text-left">
                  <p className="text-white font-medium">{pair.pair}</p>
                  <p className="text-slate-400 text-sm">${pair.price.toLocaleString()}</p>
                </div>
                <div className={`text-sm font-medium ${
                  pair.change >= 0 ? 'text-green-400' : 'text-red-400'
                }`}>
                  {pair.change >= 0 ? '+' : ''}{pair.change.toFixed(2)}%
                </div>
              </button>
            ))}
          </div>
        </motion.div>

        {/* Chart Area */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="lg:col-span-2 bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-white">Price Chart</h3>
            <div className="flex items-center space-x-2">
              {['1m', '5m', '15m', '1h', '4h', '1d'].map((timeframe) => (
                <button
                  key={timeframe}
                  className="px-3 py-1 text-slate-400 hover:text-white hover:bg-slate-700/50 rounded-lg text-sm transition-colors duration-200"
                >
                  {timeframe}
                </button>
              ))}
            </div>
          </div>
          
          {/* Chart Placeholder */}
          <div className="h-96 bg-slate-700/30 rounded-xl flex items-center justify-center relative overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-br from-green-500/10 to-red-500/10" />
            <div className="text-center">
              <div className="text-4xl mb-2">📈</div>
              <p className="text-slate-400">TradingView Chart Integration</p>
              <p className="text-slate-500 text-sm mt-1">Real-time candlestick chart with technical indicators</p>
            </div>
          </div>
        </motion.div>

        {/* Order Form */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          className="space-y-6"
        >
          {/* Order Type Selector */}
          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
            <div className="flex bg-slate-700/50 rounded-lg p-1 mb-6">
              <button
                onClick={() => setOrderSide('buy')}
                className={`flex-1 py-2 px-4 rounded-md font-medium transition-all duration-200 ${
                  orderSide === 'buy'
                    ? 'bg-green-500 text-white'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                Buy
              </button>
              <button
                onClick={() => setOrderSide('sell')}
                className={`flex-1 py-2 px-4 rounded-md font-medium transition-all duration-200 ${
                  orderSide === 'sell'
                    ? 'bg-red-500 text-white'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                Sell
              </button>
            </div>

            <div className="flex bg-slate-700/50 rounded-lg p-1 mb-6">
              <button
                onClick={() => setOrderType('limit')}
                className={`flex-1 py-2 px-4 rounded-md font-medium transition-all duration-200 ${
                  orderType === 'limit'
                    ? 'bg-amber-500 text-white'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                Limit
              </button>
              <button
                onClick={() => setOrderType('market')}
                className={`flex-1 py-2 px-4 rounded-md font-medium transition-all duration-200 ${
                  orderType === 'market'
                    ? 'bg-amber-500 text-white'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                Market
              </button>
            </div>

            <div className="space-y-4">
              {orderType === 'limit' && (
                <div>
                  <label className="block text-slate-400 text-sm mb-2">Price</label>
                  <input
                    type="number"
                    placeholder="0.00"
                    className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:border-amber-500/50"
                  />
                </div>
              )}
              
              <div>
                <label className="block text-slate-400 text-sm mb-2">Amount</label>
                <input
                  type="number"
                  placeholder="0.00"
                  className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:border-amber-500/50"
                />
              </div>

              <div>
                <label className="block text-slate-400 text-sm mb-2">Total</label>
                <input
                  type="number"
                  placeholder="0.00"
                  className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:border-amber-500/50"
                />
              </div>

              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className={`w-full py-3 rounded-lg font-semibold transition-all duration-200 ${
                  orderSide === 'buy'
                    ? 'bg-green-500 hover:bg-green-600 text-white'
                    : 'bg-red-500 hover:bg-red-600 text-white'
                }`}
              >
                {orderSide === 'buy' ? 'Buy' : 'Sell'} {selectedPair.split('/')[0]}
              </motion.button>
            </div>
          </div>

          {/* Account Balance */}
          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
            <h3 className="text-lg font-semibold text-white mb-4">Account Balance</h3>
            <div className="space-y-3">
              {[
                { asset: 'USDT', balance: '12,345.67', value: '$12,345.67' },
                { asset: 'BTC', balance: '0.5432', value: '$24,567.89' },
                { asset: 'ETH', balance: '8.9012', value: '$25,234.56' }
              ].map((balance) => (
                <div key={balance.asset} className="flex items-center justify-between p-3 bg-slate-700/30 rounded-lg">
                  <div>
                    <p className="text-white font-medium">{balance.asset}</p>
                    <p className="text-slate-400 text-sm">{balance.balance}</p>
                  </div>
                  <p className="text-white font-medium">{balance.value}</p>
                </div>
              ))}
            </div>
          </div>
        </motion.div>
      </div>

      {/* Order Book & Recent Trades */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-lg font-semibold text-white mb-4">Order Book</h3>
          <div className="space-y-2">
            <div className="grid grid-cols-3 gap-4 text-slate-400 text-sm mb-2">
              <span>Price</span>
              <span className="text-right">Amount</span>
              <span className="text-right">Total</span>
            </div>
            {/* Sell Orders */}
            {Array.from({ length: 5 }, (_, i) => (
              <div key={`sell-${i}`} className="grid grid-cols-3 gap-4 text-sm py-1">
                <span className="text-red-400">{(marketData.price + (i + 1) * 10).toLocaleString()}</span>
                <span className="text-slate-300 text-right">{(Math.random() * 2).toFixed(4)}</span>
                <span className="text-slate-300 text-right">{(Math.random() * 10000).toFixed(2)}</span>
              </div>
            ))}
            <div className="border-t border-slate-600/50 my-2 pt-2">
              <div className="text-center text-lg font-bold text-white">
                ${marketData.price.toLocaleString()}
              </div>
            </div>
            {/* Buy Orders */}
            {Array.from({ length: 5 }, (_, i) => (
              <div key={`buy-${i}`} className="grid grid-cols-3 gap-4 text-sm py-1">
                <span className="text-green-400">{(marketData.price - (i + 1) * 10).toLocaleString()}</span>
                <span className="text-slate-300 text-right">{(Math.random() * 2).toFixed(4)}</span>
                <span className="text-slate-300 text-right">{(Math.random() * 10000).toFixed(2)}</span>
              </div>
            ))}
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-lg font-semibold text-white mb-4">Recent Trades</h3>
          <div className="space-y-2">
            <div className="grid grid-cols-3 gap-4 text-slate-400 text-sm mb-2">
              <span>Price</span>
              <span className="text-right">Amount</span>
              <span className="text-right">Time</span>
            </div>
            {Array.from({ length: 10 }, (_, i) => {
              const isBuy = Math.random() > 0.5
              return (
                <div key={i} className="grid grid-cols-3 gap-4 text-sm py-1">
                  <span className={isBuy ? 'text-green-400' : 'text-red-400'}>
                    {(marketData.price + (Math.random() - 0.5) * 100).toLocaleString()}
                  </span>
                  <span className="text-slate-300 text-right">{(Math.random() * 2).toFixed(4)}</span>
                  <span className="text-slate-400 text-right">
                    {new Date(Date.now() - i * 30000).toLocaleTimeString()}
                  </span>
                </div>
              )
            })}
          </div>
        </motion.div>
      </div>
    </div>
  )
}
