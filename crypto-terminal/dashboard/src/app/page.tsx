export default function HomePage() {
  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-6xl mx-auto">
        <header className="text-center mb-12">
          <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-blue-400 to-purple-600 bg-clip-text text-transparent">
            Epic Crypto Terminal
          </h1>
          <p className="text-xl text-gray-300">
            Professional cryptocurrency trading dashboard
          </p>
        </header>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h3 className="text-lg font-semibold mb-2 text-blue-400">Market Overview</h3>
            <p className="text-gray-300">Real-time cryptocurrency market data and analysis</p>
          </div>
          
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h3 className="text-lg font-semibold mb-2 text-green-400">Portfolio Manager</h3>
            <p className="text-gray-300">Track and manage your cryptocurrency investments</p>
          </div>
          
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h3 className="text-lg font-semibold mb-2 text-purple-400">Trading Strategies</h3>
            <p className="text-gray-300">Advanced trading algorithms and automation</p>
          </div>
          
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h3 className="text-lg font-semibold mb-2 text-yellow-400">Arbitrage Tools</h3>
            <p className="text-gray-300">Cross-exchange arbitrage opportunities</p>
          </div>
          
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h3 className="text-lg font-semibold mb-2 text-red-400">Risk Management</h3>
            <p className="text-gray-300">Portfolio risk analysis and protection</p>
          </div>
          
          <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <h3 className="text-lg font-semibold mb-2 text-indigo-400">Market Analysis</h3>
            <p className="text-gray-300">Technical and fundamental market insights</p>
          </div>
        </div>

        <div className="text-center">
          <p className="text-gray-400">
            Part of the Go Coffee Web3 Ecosystem - Combining traditional coffee commerce with advanced cryptocurrency trading
          </p>
        </div>
      </div>
    </div>
  )
}
