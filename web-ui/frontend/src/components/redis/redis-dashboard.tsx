'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Database, 
  Code, 
  BarChart3, 
  Activity,
  Settings,
  Zap,
  TrendingUp,
  Users,
  Clock,
  HardDrive
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { RedisExplorer } from './redis-explorer'
import { QueryBuilder } from './query-builder'
import { useRedisVisualization } from '@/hooks/use-redis'
import { cn } from '@/lib/utils'

interface RedisDashboardProps {
  className?: string
}

export function RedisDashboard({ className }: RedisDashboardProps) {
  const [activeTab, setActiveTab] = useState('overview')
  const { metrics, performanceMetrics } = useRedisVisualization()

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    
    if (days > 0) return `${days}d ${hours}h`
    if (hours > 0) return `${hours}h ${minutes}m`
    return `${minutes}m`
  }

  return (
    <motion.div
      className={cn("space-y-6", className)}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Redis 8 Visual Interface</h1>
          <p className="text-muted-foreground">
            Explore, query, and visualize your Redis data with AI-powered search
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
            <Activity className="h-3 w-3 mr-1" />
            Connected
          </Badge>
          <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
            <Zap className="h-3 w-3 mr-1" />
            Redis 8.0
          </Badge>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="explorer" className="flex items-center gap-2">
            <Database className="h-4 w-4" />
            Data Explorer
          </TabsTrigger>
          <TabsTrigger value="query" className="flex items-center gap-2">
            <Code className="h-4 w-4" />
            Query Builder
          </TabsTrigger>
          <TabsTrigger value="monitoring" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Monitoring
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {/* Metrics Overview */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Connected Clients</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {metrics?.metrics?.connected_clients || 0}
                </div>
                <p className="text-xs text-muted-foreground">
                  Active connections
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
                <HardDrive className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {metrics?.metrics?.used_memory || 'N/A'}
                </div>
                <p className="text-xs text-muted-foreground">
                  Current memory usage
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Commands</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {metrics?.metrics?.total_commands?.toLocaleString() || 0}
                </div>
                <p className="text-xs text-muted-foreground">
                  Commands processed
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Uptime</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {metrics?.metrics?.uptime_in_seconds ? 
                    formatUptime(metrics.metrics.uptime_in_seconds) : 
                    'N/A'
                  }
                </div>
                <p className="text-xs text-muted-foreground">
                  Server uptime
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Keyspace Information */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Database className="h-5 w-5" />
                  Keyspace Overview
                </CardTitle>
              </CardHeader>
              <CardContent>
                {metrics?.metrics?.keyspace ? (
                  <div className="space-y-3">
                    {Object.entries(metrics.metrics.keyspace).map(([db, info]: [string, any]) => (
                      <div key={db} className="flex items-center justify-between p-3 bg-muted rounded-lg">
                        <div>
                          <div className="font-medium">{db.toUpperCase()}</div>
                          <div className="text-sm text-muted-foreground">
                            {info.keys} keys, {info.expires} with expiry
                          </div>
                        </div>
                        <div className="text-right">
                          <div className="text-2xl font-bold">{info.keys}</div>
                          <div className="text-xs text-muted-foreground">keys</div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center text-muted-foreground py-8">
                    No keyspace data available
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-5 w-5" />
                  Performance Metrics
                </CardTitle>
              </CardHeader>
              <CardContent>
                {performanceMetrics?.metrics ? (
                  <div className="space-y-4">
                    <div>
                      <div className="flex justify-between items-center mb-2">
                        <span className="text-sm font-medium">Memory Stats</span>
                      </div>
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span>Peak Allocated:</span>
                          <span>{performanceMetrics.metrics.memory_stats?.peak_allocated}</span>
                        </div>
                        <div className="flex justify-between">
                          <span>Total Allocated:</span>
                          <span>{performanceMetrics.metrics.memory_stats?.total_allocated}</span>
                        </div>
                      </div>
                    </div>
                    
                    <div>
                      <div className="flex justify-between items-center mb-2">
                        <span className="text-sm font-medium">Slow Log</span>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {performanceMetrics.metrics.slow_log?.length || 0} slow queries
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="text-center text-muted-foreground py-8">
                    Loading performance metrics...
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Zap className="h-5 w-5" />
                Quick Actions
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <motion.div
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  className="p-4 border rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => setActiveTab('explorer')}
                >
                  <Database className="h-8 w-8 mb-2 text-blue-500" />
                  <div className="font-medium">Explore Data</div>
                  <div className="text-sm text-muted-foreground">Browse keys and values</div>
                </motion.div>

                <motion.div
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  className="p-4 border rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => setActiveTab('query')}
                >
                  <Code className="h-8 w-8 mb-2 text-green-500" />
                  <div className="font-medium">Build Query</div>
                  <div className="text-sm text-muted-foreground">Visual query builder</div>
                </motion.div>

                <motion.div
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  className="p-4 border rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => setActiveTab('monitoring')}
                >
                  <Activity className="h-8 w-8 mb-2 text-orange-500" />
                  <div className="font-medium">Monitor</div>
                  <div className="text-sm text-muted-foreground">Real-time monitoring</div>
                </motion.div>

                <motion.div
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  className="p-4 border rounded-lg cursor-pointer hover:bg-muted/50 transition-colors"
                >
                  <Settings className="h-8 w-8 mb-2 text-purple-500" />
                  <div className="font-medium">Settings</div>
                  <div className="text-sm text-muted-foreground">Configure Redis</div>
                </motion.div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="explorer">
          <RedisExplorer />
        </TabsContent>

        <TabsContent value="query">
          <QueryBuilder />
        </TabsContent>

        <TabsContent value="monitoring">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                Real-time Monitoring
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-center text-muted-foreground py-8">
                Real-time monitoring dashboard will be implemented here
                <br />
                Features: Live metrics, command monitoring, performance graphs
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </motion.div>
  )
}
