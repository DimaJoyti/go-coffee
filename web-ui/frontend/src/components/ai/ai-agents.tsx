'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Bot, 
  Activity, 
  Zap, 
  Pause, 
  Play, 
  Settings,
  AlertCircle,
  CheckCircle,
  Clock
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { formatRelativeTime, cn } from '@/lib/utils'

interface Agent {
  id: string
  name: string
  description: string
  status: 'active' | 'inactive' | 'error' | 'maintenance'
  lastActivity: string
  tasksCompleted: number
  uptime: number
  performance: number
  icon: string
}

interface AIAgentsProps {
  className?: string
}

export function AIAgents({ className }: AIAgentsProps) {
  const [agents, setAgents] = useState<Agent[]>([
    {
      id: 'beverage-inventor',
      name: 'Beverage Inventor',
      description: 'Creates new drink recipes and flavor combinations',
      status: 'active',
      lastActivity: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
      tasksCompleted: 156,
      uptime: 99.8,
      performance: 94.2,
      icon: '🧪'
    },
    {
      id: 'inventory-manager',
      name: 'Inventory Manager',
      description: 'Tracks real-time inventory and supply chain',
      status: 'active',
      lastActivity: new Date(Date.now() - 1 * 60 * 1000).toISOString(),
      tasksCompleted: 342,
      uptime: 99.9,
      performance: 97.8,
      icon: '📦'
    },
    {
      id: 'task-manager',
      name: 'Task Manager',
      description: 'Creates and tracks operational tasks',
      status: 'active',
      lastActivity: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
      tasksCompleted: 89,
      uptime: 98.5,
      performance: 91.3,
      icon: '✅'
    },
    {
      id: 'feedback-analyst',
      name: 'Feedback Analyst',
      description: 'Analyzes customer feedback and sentiment',
      status: 'active',
      lastActivity: new Date(Date.now() - 3 * 60 * 1000).toISOString(),
      tasksCompleted: 234,
      uptime: 99.2,
      performance: 88.7,
      icon: '📊'
    },
    {
      id: 'scheduler',
      name: 'Scheduler Agent',
      description: 'Manages daily operations and staff scheduling',
      status: 'maintenance',
      lastActivity: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      tasksCompleted: 67,
      uptime: 95.4,
      performance: 85.2,
      icon: '📅'
    },
    {
      id: 'notifier',
      name: 'Notifier Agent',
      description: 'Sends alerts and notifications',
      status: 'active',
      lastActivity: new Date(Date.now() - 1 * 60 * 1000).toISOString(),
      tasksCompleted: 445,
      uptime: 99.7,
      performance: 96.1,
      icon: '🔔'
    },
    {
      id: 'social-media',
      name: 'Social Media Content',
      description: 'Generates social media content and posts',
      status: 'error',
      lastActivity: new Date(Date.now() - 45 * 60 * 1000).toISOString(),
      tasksCompleted: 78,
      uptime: 87.3,
      performance: 72.4,
      icon: '📱'
    },
    {
      id: 'tasting-coordinator',
      name: 'Tasting Coordinator',
      description: 'Schedules and manages tasting sessions',
      status: 'active',
      lastActivity: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
      tasksCompleted: 23,
      uptime: 98.9,
      performance: 89.6,
      icon: '👅'
    },
    {
      id: 'inter-location',
      name: 'Inter-Location Coordinator',
      description: 'Coordinates between different coffee shop locations',
      status: 'active',
      lastActivity: new Date(Date.now() - 7 * 60 * 1000).toISOString(),
      tasksCompleted: 134,
      uptime: 99.1,
      performance: 92.8,
      icon: '🌐'
    }
  ])

  const getStatusIcon = (status: Agent['status']) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="h-4 w-4" />
      case 'inactive':
        return <Pause className="h-4 w-4" />
      case 'error':
        return <AlertCircle className="h-4 w-4" />
      case 'maintenance':
        return <Clock className="h-4 w-4" />
    }
  }

  const getStatusColor = (status: Agent['status']) => {
    switch (status) {
      case 'active':
        return 'success'
      case 'inactive':
        return 'secondary'
      case 'error':
        return 'destructive'
      case 'maintenance':
        return 'warning'
    }
  }

  const toggleAgent = (agentId: string) => {
    setAgents(prev => prev.map(agent => 
      agent.id === agentId 
        ? { 
            ...agent, 
            status: agent.status === 'active' ? 'inactive' : 'active' 
          }
        : agent
    ))
  }

  const activeAgents = agents.filter(agent => agent.status === 'active').length
  const totalTasks = agents.reduce((sum, agent) => sum + agent.tasksCompleted, 0)
  const avgPerformance = agents.reduce((sum, agent) => sum + agent.performance, 0) / agents.length

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
          <h1 className="text-2xl font-bold">AI Agents</h1>
          <p className="text-muted-foreground">
            Monitor and manage your AI agent ecosystem
          </p>
        </div>
        <Button>
          <Settings className="h-4 w-4 mr-2" />
          Agent Settings
        </Button>
      </div>

      {/* Overview Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-2 mb-2">
              <Activity className="h-5 w-5 text-green-500" />
              <span className="font-medium">Active Agents</span>
            </div>
            <div className="text-2xl font-bold">{activeAgents}/{agents.length}</div>
            <div className="text-sm text-muted-foreground">
              {((activeAgents / agents.length) * 100).toFixed(1)}% operational
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-2 mb-2">
              <Zap className="h-5 w-5 text-blue-500" />
              <span className="font-medium">Tasks Completed</span>
            </div>
            <div className="text-2xl font-bold">{totalTasks.toLocaleString()}</div>
            <div className="text-sm text-muted-foreground">
              Total across all agents
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center gap-2 mb-2">
              <Bot className="h-5 w-5 text-purple-500" />
              <span className="font-medium">Avg Performance</span>
            </div>
            <div className="text-2xl font-bold">{avgPerformance.toFixed(1)}%</div>
            <div className="text-sm text-muted-foreground">
              System-wide efficiency
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Agents Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {agents.map((agent, index) => (
          <motion.div
            key={agent.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <Card className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-2xl">{agent.icon}</span>
                    <div>
                      <CardTitle className="text-lg">{agent.name}</CardTitle>
                      <Badge variant={getStatusColor(agent.status)} className="mt-1">
                        {getStatusIcon(agent.status)}
                        {agent.status}
                      </Badge>
                    </div>
                  </div>
                </div>
              </CardHeader>
              
              <CardContent>
                <div className="space-y-4">
                  <p className="text-sm text-muted-foreground">
                    {agent.description}
                  </p>
                  
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <div className="text-muted-foreground">Tasks</div>
                      <div className="font-medium">{agent.tasksCompleted}</div>
                    </div>
                    <div>
                      <div className="text-muted-foreground">Uptime</div>
                      <div className="font-medium">{agent.uptime}%</div>
                    </div>
                  </div>
                  
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-muted-foreground">Performance</span>
                      <span className="font-medium">{agent.performance}%</span>
                    </div>
                    <div className="w-full bg-muted rounded-full h-2">
                      <div 
                        className={cn(
                          "h-2 rounded-full transition-all duration-300",
                          agent.performance >= 90 ? "bg-green-500" :
                          agent.performance >= 70 ? "bg-yellow-500" : "bg-red-500"
                        )}
                        style={{ width: `${agent.performance}%` }}
                      />
                    </div>
                  </div>
                  
                  <div className="text-xs text-muted-foreground">
                    Last activity: {formatRelativeTime(agent.lastActivity)}
                  </div>
                  
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => toggleAgent(agent.id)}
                      className="flex-1"
                      disabled={agent.status === 'maintenance'}
                    >
                      {agent.status === 'active' ? (
                        <Pause className="h-3 w-3 mr-1" />
                      ) : (
                        <Play className="h-3 w-3 mr-1" />
                      )}
                      {agent.status === 'active' ? 'Pause' : 'Start'}
                    </Button>
                    <Button size="sm" variant="ghost">
                      <Settings className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>
    </motion.div>
  )
}
