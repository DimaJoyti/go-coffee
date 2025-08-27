'use client'

import { useRealTimeData, usePrices, useConnectionStatus } from '@/contexts/RealTimeDataContext'

export default function DataDebug() {
  const { prices, trades, portfolio } = useRealTimeData()
  const { prices: priceArray } = usePrices()
  const { isConnected, connectionStatus } = useConnectionStatus()

  if (process.env.NODE_ENV !== 'development') {
    return null
  }

  return (
    <div className="fixed bottom-4 left-4 bg-black/80 text-white p-4 rounded-lg text-xs max-w-sm z-50">
      <h3 className="font-bold mb-2">Debug Info</h3>
      <div className="space-y-1">
        <div>Connection: {connectionStatus} ({isConnected ? 'Connected' : 'Disconnected'})</div>
        <div>Prices Map Size: {prices.size}</div>
        <div>Prices Array Length: {priceArray.length}</div>
        <div>Trades: {trades.length}</div>
        <div>Portfolio: {portfolio ? 'Loaded' : 'Not loaded'}</div>
        <div>Sample Price: {priceArray[0] ? `${priceArray[0].symbol}: $${priceArray[0].price}` : 'None'}</div>
      </div>
    </div>
  )
}
