'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'

export default function AIAgents() {
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null)

  const agents = [
    {
      id: 'trading-alpha',
      name: 'Trading Bot Alpha',
      type: 'Trading',
      status: 'active',
      performance: '+$2,345',
      trades: 67,
      accuracy: 94.2,
      description: 'Advanced algorithmic trading bot with machine learning capabilities',
      icon: '🤖',
      color: 'from-green-500 to-emerald-500'
    },
    {
      id: 'market-analyzer',
      name: 'Market Analyzer',
      type: 'Analysis',
      status: 'active',
      performance: '+$1,892',
      trades: 23,
      accuracy: 87.5,
      description: 'Real-time market sentiment analysis and trend prediction',
      icon: '📊',
      color: 'from-blue-500 to-purple-500'
    },
    {
      id: 'risk-manager',
      name: 'Risk Manager',
      type: 'Risk Management',
      status: 'monitoring',
      performance: '+$456',
      trades: 12,
      accuracy: 96.8,
      description: 'Portfolio risk assessment and automated stop-loss management',
      icon: '🛡️',
      color: 'from-orange-500 to-red-500'
    },
    {
      id: 'arbitrage-hunter',
      name: 'Arbitrage Hunter',
      type: 'Arbitrage',
      status: 'active',
      performance: '+$3,103',
      trades: 89,
      accuracy: 91.3,
      description: 'Cross-exchange arbitrage opportunities detection and execution',
      icon: '🎯',
      color: 'from-purple-500 to-pink-500'
    },
    {
      id: 'coffee-optimizer',
      name: 'Coffee Supply Optimizer',
      type: 'Supply Chain',
      status: 'active',
      performance: '+$1,234',
      trades: 34,
      accuracy: 89.7,
      description: 'Optimizes coffee supply chain and inventory management',
      icon: '☕',
      color: 'from-amber-500 to-orange-500'
    },
    {
      id: 'sentiment-analyzer',
      name: 'Sentiment Analyzer',
      type: 'Social Media',
      status: 'active',
      performance: '+$789',
      trades: 45,
      accuracy: 85.4,
      description: 'Social media sentiment analysis for market prediction',
      icon: '💭',
      color: 'from-cyan-500 to-blue-500'
    }
  ]

  const agentLogs = [
    { time: '14:32:15', agent: 'Trading Bot Alpha', action: 'Executed buy order for 0.5 BTC', type: 'success' },
    { time: '14:31:45', agent: 'Market Analyzer', action: 'Detected bullish trend in ETH', type: 'info' },
    { time: '14:30:22', agent: 'Risk Manager', action: 'Adjusted stop-loss for SOL position', type: 'warning' },
    { time: '14:29:18', agent: 'Arbitrage Hunter', action: 'Found 2.3% arbitrage opportunity', type: 'success' },
    { time: '14:28:45', agent: 'Coffee Optimizer', action: 'Reordered Ethiopian beans inventory', type: 'info' }
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0"
      >
        <div>
          <h2 className="text-3xl font-bold text-white">AI Agent Network</h2>
          <p className="text-slate-400">Intelligent automation for trading and operations</p>
        </div>
        
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 px-4 py-2 bg-green-500/20 border border-green-500/30 rounded-lg">
            <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
            <span className="text-green-400 text-sm">6 Agents Active</span>
          </div>
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="px-6 py-2 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded-lg font-semibold hover:from-blue-600 hover:to-purple-600 transition-all duration-200"
          >
            Deploy New Agent
          </motion.button>
        </div>
      </motion.div>

      {/* Agents Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {agents.map((agent, index) => (
          <motion.div
            key={agent.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            whileHover={{ scale: 1.02 }}
            onClick={() => setSelectedAgent(selectedAgent === agent.id ? null : agent.id)}
            className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 cursor-pointer hover:border-slate-600/50 transition-all duration-300"
          >
            <div className="flex items-center justify-between mb-4">
              <div className={`w-12 h-12 bg-gradient-to-r ${agent.color} rounded-xl flex items-center justify-center text-2xl`}>
                {agent.icon}
              </div>
              <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                agent.status === 'active' ? 'bg-green-500/20 text-green-400' :
                agent.status === 'monitoring' ? 'bg-yellow-500/20 text-yellow-400' :
                'bg-red-500/20 text-red-400'
              }`}>
                {agent.status}
              </span>
            </div>

            <h3 className="text-lg font-semibold text-white mb-2">{agent.name}</h3>
            <p className="text-slate-400 text-sm mb-4">{agent.description}</p>

            <div className="grid grid-cols-2 gap-4 mb-4">
              <div>
                <p className="text-slate-400 text-xs">24h Profit</p>
                <p className="text-green-400 font-semibold">{agent.performance}</p>
              </div>
              <div>
                <p className="text-slate-400 text-xs">Trades</p>
                <p className="text-white font-semibold">{agent.trades}</p>
              </div>
            </div>

            <div className="mb-4">
              <div className="flex justify-between items-center mb-2">
                <span className="text-slate-400 text-xs">Accuracy</span>
                <span className="text-white text-xs">{agent.accuracy}%</span>
              </div>
              <div className="w-full bg-slate-700/50 rounded-full h-2">
                <motion.div
                  initial={{ width: 0 }}
                  animate={{ width: `${agent.accuracy}%` }}
                  transition={{ delay: index * 0.1 + 0.5, duration: 1 }}
                  className={`h-2 bg-gradient-to-r ${agent.color} rounded-full`}
                />
              </div>
            </div>

            {selectedAgent === agent.id && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
                className="border-t border-slate-700/50 pt-4 space-y-2"
              >
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Type:</span>
                  <span className="text-white">{agent.type}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Last Active:</span>
                  <span className="text-white">2 minutes ago</span>
                </div>
                <div className="flex space-x-2 mt-4">
                  <button className="flex-1 py-2 bg-slate-700/50 text-white rounded-lg text-sm hover:bg-slate-600/50 transition-colors">
                    Configure
                  </button>
                  <button className="flex-1 py-2 bg-red-500/20 text-red-400 rounded-lg text-sm hover:bg-red-500/30 transition-colors">
                    Stop
                  </button>
                </div>
              </motion.div>
            )}
          </motion.div>
        ))}
      </div>

      {/* Performance Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-xl font-semibold text-white mb-6">Network Performance</h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center p-4 bg-slate-700/30 rounded-lg">
              <span className="text-slate-400">Total Profit (24h)</span>
              <span className="text-green-400 font-bold text-lg">+$9,819</span>
            </div>
            <div className="flex justify-between items-center p-4 bg-slate-700/30 rounded-lg">
              <span className="text-slate-400">Total Trades</span>
              <span className="text-white font-bold text-lg">270</span>
            </div>
            <div className="flex justify-between items-center p-4 bg-slate-700/30 rounded-lg">
              <span className="text-slate-400">Success Rate</span>
              <span className="text-green-400 font-bold text-lg">92.1%</span>
            </div>
            <div className="flex justify-between items-center p-4 bg-slate-700/30 rounded-lg">
              <span className="text-slate-400">Active Agents</span>
              <span className="text-white font-bold text-lg">6/8</span>
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
        >
          <h3 className="text-xl font-semibold text-white mb-6">Recent Activity</h3>
          <div className="space-y-3 max-h-64 overflow-y-auto">
            {agentLogs.map((log, index) => (
              <motion.div
                key={index}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.1 }}
                className="flex items-start space-x-3 p-3 bg-slate-700/30 rounded-lg"
              >
                <div className={`w-2 h-2 rounded-full mt-2 flex-shrink-0 ${
                  log.type === 'success' ? 'bg-green-400' :
                  log.type === 'warning' ? 'bg-yellow-400' : 'bg-blue-400'
                }`} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-1">
                    <p className="text-white text-sm font-medium truncate">{log.agent}</p>
                    <p className="text-slate-400 text-xs flex-shrink-0">{log.time}</p>
                  </div>
                  <p className="text-slate-300 text-sm">{log.action}</p>
                </div>
              </motion.div>
            ))}
          </div>
        </motion.div>
      </div>

      {/* Agent Configuration Panel */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6"
      >
        <h3 className="text-xl font-semibold text-white mb-6">Quick Configuration</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div>
            <label className="block text-slate-400 text-sm mb-2">Risk Level</label>
            <select className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white">
              <option>Conservative</option>
              <option>Moderate</option>
              <option>Aggressive</option>
            </select>
          </div>
          <div>
            <label className="block text-slate-400 text-sm mb-2">Max Position Size</label>
            <input
              type="number"
              placeholder="1000"
              className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400"
            />
          </div>
          <div>
            <label className="block text-slate-400 text-sm mb-2">Stop Loss %</label>
            <input
              type="number"
              placeholder="5"
              className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white placeholder-slate-400"
            />
          </div>
        </div>
        <div className="flex justify-end mt-6">
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="px-6 py-2 bg-gradient-to-r from-green-500 to-emerald-500 text-white rounded-lg font-semibold hover:from-green-600 hover:to-emerald-600 transition-all duration-200"
          >
            Apply Configuration
          </motion.button>
        </div>
      </motion.div>
    </div>
  )
}
