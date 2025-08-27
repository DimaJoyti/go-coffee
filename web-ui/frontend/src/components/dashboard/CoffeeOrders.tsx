'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'

export default function CoffeeOrders() {
  const [activeTab, setActiveTab] = useState<'marketplace' | 'orders' | 'inventory'>('marketplace')

  const coffeeProducts = [
    {
      id: 1,
      name: 'Ethiopian Yirgacheffe',
      origin: 'Ethiopia',
      price: 24.99,
      rating: 4.8,
      stock: 45,
      description: 'Bright, floral notes with citrus undertones',
      image: '☕',
      roast: 'Light',
      processing: 'Washed'
    },
    {
      id: 2,
      name: 'Colombian Supremo',
      origin: 'Colombia',
      price: 19.99,
      rating: 4.6,
      stock: 32,
      description: 'Rich, full-bodied with chocolate notes',
      image: '☕',
      roast: 'Medium',
      processing: 'Natural'
    },
    {
      id: 3,
      name: 'Jamaican Blue Mountain',
      origin: 'Jamaica',
      price: 89.99,
      rating: 4.9,
      stock: 12,
      description: 'Smooth, mild flavor with no bitterness',
      image: '☕',
      roast: 'Medium',
      processing: 'Washed'
    },
    {
      id: 4,
      name: 'Hawaiian Kona',
      origin: 'Hawaii',
      price: 45.99,
      rating: 4.7,
      stock: 28,
      description: 'Smooth, rich taste with low acidity',
      image: '☕',
      roast: 'Medium-Dark',
      processing: 'Washed'
    }
  ]

  const recentOrders = [
    {
      id: '#ORD-001',
      product: 'Ethiopian Yirgacheffe',
      quantity: 2,
      total: 49.98,
      status: 'delivered',
      date: '2024-01-15',
      customer: 'John Doe'
    },
    {
      id: '#ORD-002',
      product: 'Colombian Supremo',
      quantity: 1,
      total: 19.99,
      status: 'processing',
      date: '2024-01-14',
      customer: 'Jane Smith'
    },
    {
      id: '#ORD-003',
      product: 'Jamaican Blue Mountain',
      quantity: 1,
      total: 89.99,
      status: 'shipped',
      date: '2024-01-13',
      customer: 'Mike Johnson'
    }
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
          <h2 className="text-3xl font-bold text-white">Coffee Marketplace</h2>
          <p className="text-slate-400">Premium coffee beans from around the world</p>
        </div>
        
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2 px-4 py-2 bg-green-500/20 border border-green-500/30 rounded-lg">
            <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
            <span className="text-green-400 text-sm">Supply Chain Active</span>
          </div>
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="px-6 py-2 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-lg font-semibold hover:from-amber-600 hover:to-orange-600 transition-all duration-200"
          >
            New Order
          </motion.button>
        </div>
      </motion.div>

      {/* Tab Navigation */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex space-x-1 bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-2"
      >
        {[
          { id: 'marketplace', label: 'Marketplace', icon: '🛒' },
          { id: 'orders', label: 'My Orders', icon: '📦' },
          { id: 'inventory', label: 'Inventory', icon: '📊' }
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as any)}
            className={`flex-1 flex items-center justify-center space-x-2 px-6 py-3 rounded-xl font-semibold transition-all duration-300 ${
              activeTab === tab.id
                ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-lg'
                : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
            }`}
          >
            <span>{tab.icon}</span>
            <span>{tab.label}</span>
          </button>
        ))}
      </motion.div>

      {/* Content */}
      <motion.div
        key={activeTab}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        {activeTab === 'marketplace' && (
          <div className="space-y-6">
            {/* Filters */}
            <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
              <div className="flex flex-wrap items-center gap-4">
                <div className="flex items-center space-x-2">
                  <label className="text-slate-400 text-sm">Origin:</label>
                  <select className="px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white text-sm">
                    <option>All Origins</option>
                    <option>Ethiopia</option>
                    <option>Colombia</option>
                    <option>Jamaica</option>
                    <option>Hawaii</option>
                  </select>
                </div>
                <div className="flex items-center space-x-2">
                  <label className="text-slate-400 text-sm">Roast:</label>
                  <select className="px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white text-sm">
                    <option>All Roasts</option>
                    <option>Light</option>
                    <option>Medium</option>
                    <option>Medium-Dark</option>
                    <option>Dark</option>
                  </select>
                </div>
                <div className="flex items-center space-x-2">
                  <label className="text-slate-400 text-sm">Price:</label>
                  <select className="px-3 py-2 bg-slate-700/50 border border-slate-600/50 rounded-lg text-white text-sm">
                    <option>All Prices</option>
                    <option>Under $25</option>
                    <option>$25 - $50</option>
                    <option>$50 - $75</option>
                    <option>Over $75</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Products Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {coffeeProducts.map((product, index) => (
                <motion.div
                  key={product.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                  whileHover={{ scale: 1.02 }}
                  className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6 hover:border-amber-500/30 transition-all duration-300"
                >
                  <div className="text-center mb-4">
                    <div className="w-20 h-20 bg-gradient-to-br from-amber-500/20 to-orange-500/20 rounded-full flex items-center justify-center text-4xl mx-auto mb-3">
                      {product.image}
                    </div>
                    <h3 className="text-lg font-semibold text-white mb-1">{product.name}</h3>
                    <p className="text-slate-400 text-sm">{product.origin}</p>
                  </div>

                  <div className="space-y-3 mb-4">
                    <div className="flex justify-between items-center">
                      <span className="text-slate-400 text-sm">Price</span>
                      <span className="text-amber-400 font-bold">${product.price}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-slate-400 text-sm">Rating</span>
                      <div className="flex items-center space-x-1">
                        <span className="text-yellow-400">★</span>
                        <span className="text-white text-sm">{product.rating}</span>
                      </div>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-slate-400 text-sm">Stock</span>
                      <span className="text-white text-sm">{product.stock} bags</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-slate-400 text-sm">Roast</span>
                      <span className="text-white text-sm">{product.roast}</span>
                    </div>
                  </div>

                  <p className="text-slate-300 text-sm mb-4">{product.description}</p>

                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    className="w-full py-2 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-lg font-semibold hover:from-amber-600 hover:to-orange-600 transition-all duration-200"
                  >
                    Add to Cart
                  </motion.button>
                </motion.div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'orders' && (
          <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
            <h3 className="text-xl font-semibold text-white mb-6">Recent Orders</h3>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-700/50">
                    <th className="text-left text-slate-400 font-medium py-3">Order ID</th>
                    <th className="text-left text-slate-400 font-medium py-3">Product</th>
                    <th className="text-left text-slate-400 font-medium py-3">Quantity</th>
                    <th className="text-left text-slate-400 font-medium py-3">Total</th>
                    <th className="text-left text-slate-400 font-medium py-3">Status</th>
                    <th className="text-left text-slate-400 font-medium py-3">Date</th>
                  </tr>
                </thead>
                <tbody>
                  {recentOrders.map((order, index) => (
                    <motion.tr
                      key={order.id}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.1 }}
                      className="border-b border-slate-700/30 hover:bg-slate-700/20 transition-colors duration-200"
                    >
                      <td className="py-4 text-white font-medium">{order.id}</td>
                      <td className="py-4 text-white">{order.product}</td>
                      <td className="py-4 text-slate-300">{order.quantity}</td>
                      <td className="py-4 text-amber-400 font-semibold">${order.total}</td>
                      <td className="py-4">
                        <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                          order.status === 'delivered' ? 'bg-green-500/20 text-green-400' :
                          order.status === 'shipped' ? 'bg-blue-500/20 text-blue-400' :
                          'bg-yellow-500/20 text-yellow-400'
                        }`}>
                          {order.status}
                        </span>
                      </td>
                      <td className="py-4 text-slate-400">{order.date}</td>
                    </motion.tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === 'inventory' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
              <h3 className="text-xl font-semibold text-white mb-6">Inventory Overview</h3>
              <div className="space-y-4">
                {coffeeProducts.map((product) => (
                  <div key={product.id} className="flex items-center justify-between p-4 bg-slate-700/30 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <div className="text-2xl">{product.image}</div>
                      <div>
                        <p className="text-white font-medium">{product.name}</p>
                        <p className="text-slate-400 text-sm">{product.origin}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-white font-semibold">{product.stock} bags</p>
                      <p className={`text-sm ${
                        product.stock > 30 ? 'text-green-400' :
                        product.stock > 15 ? 'text-yellow-400' : 'text-red-400'
                      }`}>
                        {product.stock > 30 ? 'In Stock' :
                         product.stock > 15 ? 'Low Stock' : 'Critical'}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl p-6">
              <h3 className="text-xl font-semibold text-white mb-6">Supply Chain Status</h3>
              <div className="space-y-4">
                {[
                  { supplier: 'Ethiopian Farms Co.', status: 'active', nextDelivery: '2024-01-20' },
                  { supplier: 'Colombian Coffee Ltd.', status: 'active', nextDelivery: '2024-01-22' },
                  { supplier: 'Jamaica Blue Mountain', status: 'delayed', nextDelivery: '2024-01-25' },
                  { supplier: 'Hawaiian Kona Farms', status: 'active', nextDelivery: '2024-01-18' }
                ].map((supplier, index) => (
                  <div key={index} className="flex items-center justify-between p-4 bg-slate-700/30 rounded-lg">
                    <div>
                      <p className="text-white font-medium">{supplier.supplier}</p>
                      <p className="text-slate-400 text-sm">Next delivery: {supplier.nextDelivery}</p>
                    </div>
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                      supplier.status === 'active' ? 'bg-green-500/20 text-green-400' :
                      'bg-yellow-500/20 text-yellow-400'
                    }`}>
                      {supplier.status}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </motion.div>
    </div>
  )
}
