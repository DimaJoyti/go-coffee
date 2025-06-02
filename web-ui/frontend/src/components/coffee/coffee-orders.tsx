'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Coffee, 
  Clock, 
  CheckCircle, 
  XCircle, 
  Plus,
  Filter,
  Search
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { formatCurrency, formatRelativeTime, cn } from '@/lib/utils'

interface Order {
  id: string
  customerName: string
  items: string[]
  total: number
  status: 'pending' | 'preparing' | 'ready' | 'completed' | 'cancelled'
  createdAt: string
  location: string
}

interface CoffeeOrdersProps {
  className?: string
}

export function CoffeeOrders({ className }: CoffeeOrdersProps) {
  const [orders] = useState<Order[]>([
    {
      id: 'ORD-001',
      customerName: 'John Doe',
      items: ['Large Latte', 'Croissant'],
      total: 8.50,
      status: 'preparing',
      createdAt: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
      location: 'Downtown'
    },
    {
      id: 'ORD-002',
      customerName: 'Jane Smith',
      items: ['Cappuccino', 'Blueberry Muffin'],
      total: 7.25,
      status: 'ready',
      createdAt: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
      location: 'Mall'
    },
    {
      id: 'ORD-003',
      customerName: 'Mike Johnson',
      items: ['Espresso', 'Americano'],
      total: 6.00,
      status: 'pending',
      createdAt: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
      location: 'Airport'
    },
    {
      id: 'ORD-004',
      customerName: 'Sarah Wilson',
      items: ['Mocha', 'Chocolate Chip Cookie'],
      total: 9.75,
      status: 'completed',
      createdAt: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      location: 'Downtown'
    }
  ])

  const [filter, setFilter] = useState<'all' | 'pending' | 'preparing' | 'ready' | 'completed'>('all')

  const filteredOrders = filter === 'all' 
    ? orders 
    : orders.filter(order => order.status === filter)

  const getStatusIcon = (status: Order['status']) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4" />
      case 'preparing':
        return <Coffee className="h-4 w-4" />
      case 'ready':
        return <CheckCircle className="h-4 w-4" />
      case 'completed':
        return <CheckCircle className="h-4 w-4" />
      case 'cancelled':
        return <XCircle className="h-4 w-4" />
    }
  }

  const getStatusColor = (status: Order['status']) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'preparing':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'ready':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'completed':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300'
      case 'cancelled':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
    }
  }

  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Coffee Orders</h1>
          <p className="text-muted-foreground">
            Manage and track coffee orders across all locations
          </p>
        </div>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          New Order
        </Button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4 mb-6">
        <div className="flex items-center gap-2">
          <Filter className="h-4 w-4 text-muted-foreground" />
          <span className="text-sm font-medium">Filter:</span>
        </div>
        
        {['all', 'pending', 'preparing', 'ready', 'completed'].map((status) => (
          <Button
            key={status}
            variant={filter === status ? 'default' : 'outline'}
            size="sm"
            onClick={() => setFilter(status as any)}
            className="capitalize"
          >
            {status}
          </Button>
        ))}

        <div className="ml-auto flex items-center gap-2">
          <Search className="h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search orders..."
            className="px-3 py-1 border border-border rounded-md text-sm"
          />
        </div>
      </div>

      {/* Orders Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredOrders.map((order, index) => (
          <motion.div
            key={order.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <Card className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg">{order.id}</CardTitle>
                  <Badge className={cn("flex items-center gap-1", getStatusColor(order.status))}>
                    {getStatusIcon(order.status)}
                    {order.status}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">{order.customerName}</p>
              </CardHeader>
              
              <CardContent>
                <div className="space-y-3">
                  <div>
                    <p className="text-sm font-medium mb-1">Items:</p>
                    <ul className="text-sm text-muted-foreground">
                      {order.items.map((item, i) => (
                        <li key={i}>• {item}</li>
                      ))}
                    </ul>
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Total:</span>
                    <span className="font-semibold">{formatCurrency(order.total)}</span>
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Location:</span>
                    <span className="text-sm">{order.location}</span>
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">Ordered:</span>
                    <span className="text-sm">{formatRelativeTime(order.createdAt)}</span>
                  </div>
                  
                  {order.status !== 'completed' && order.status !== 'cancelled' && (
                    <div className="flex gap-2 mt-4">
                      <Button size="sm" variant="outline" className="flex-1">
                        Update
                      </Button>
                      <Button size="sm" className="flex-1">
                        Complete
                      </Button>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>

      {filteredOrders.length === 0 && (
        <div className="text-center py-12">
          <Coffee className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-medium mb-2">No orders found</h3>
          <p className="text-muted-foreground">
            {filter === 'all' 
              ? "No orders have been placed yet." 
              : `No ${filter} orders found.`}
          </p>
        </div>
      )}
    </motion.div>
  )
}
