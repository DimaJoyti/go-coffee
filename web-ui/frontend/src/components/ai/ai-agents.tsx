// Simplified AI Agents for when dependencies are not installed
// This provides a basic AI agent management interface

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
  // Mock agents data for demonstration
  const agents: Agent[] = [
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
  ]

  // Helper functions for display
  const getStatusIcon = (status: Agent['status']) => {
    switch (status) {
      case 'active': return '✅'
      case 'inactive': return '⏸️'
      case 'error': return '⚠️'
      case 'maintenance': return '🔧'
    }
  }

  const getStatusColor = (status: Agent['status']) => {
    switch (status) {
      case 'active': return '#10b981'
      case 'inactive': return '#6b7280'
      case 'error': return '#ef4444'
      case 'maintenance': return '#f59e0b'
    }
  }

  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))

    if (diffInMinutes < 60) return `${diffInMinutes}m ago`
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h ago`
    return `${Math.floor(diffInMinutes / 1440)}d ago`
  }

  const activeAgents = agents.filter(agent => agent.status === 'active').length
  const totalTasks = agents.reduce((sum, agent) => sum + agent.tasksCompleted, 0)
  const avgPerformance = agents.reduce((sum, agent) => sum + agent.performance, 0) / agents.length

  // Return HTML string since React/JSX isn't available
  return `
    <div class="ai-agents ${className || ''}" style="padding: 1.5rem; background: #0f172a; color: #f8fafc; min-height: 100vh;">
      <!-- Header -->
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <div>
          <h1 style="font-size: 2rem; font-weight: bold; margin-bottom: 0.5rem; color: #f8fafc;">
            🤖 AI Agents
          </h1>
          <p style="color: #94a3b8; font-size: 1rem;">
            Monitor and manage your AI agent ecosystem
          </p>
        </div>
        <button style="
          padding: 0.75rem 1rem;
          background: #f59e0b;
          border: none;
          border-radius: 0.5rem;
          color: white;
          cursor: pointer;
          font-size: 0.875rem;
          display: flex;
          align-items: center;
          gap: 0.5rem;
        ">
          ⚙️ Agent Settings
        </button>
      </div>

      <!-- Overview Stats -->
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem; margin-bottom: 2rem;">
        <!-- Active Agents Card -->
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border: 1px solid rgba(148, 163, 184, 0.1);
          border-radius: 1rem;
          padding: 1.5rem;
        ">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
            <span style="font-size: 1.25rem; color: #10b981;">📊</span>
            <span style="font-weight: 500; color: #f8fafc;">Active Agents</span>
          </div>
          <div style="font-size: 2rem; font-weight: bold; color: #f8fafc;">${activeAgents}/${agents.length}</div>
          <div style="font-size: 0.875rem; color: #94a3b8;">
            ${((activeAgents / agents.length) * 100).toFixed(1)}% operational
          </div>
        </div>

        <!-- Tasks Completed Card -->
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border: 1px solid rgba(148, 163, 184, 0.1);
          border-radius: 1rem;
          padding: 1.5rem;
        ">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
            <span style="font-size: 1.25rem; color: #3b82f6;">⚡</span>
            <span style="font-weight: 500; color: #f8fafc;">Tasks Completed</span>
          </div>
          <div style="font-size: 2rem; font-weight: bold; color: #f8fafc;">${totalTasks.toLocaleString()}</div>
          <div style="font-size: 0.875rem; color: #94a3b8;">Total across all agents</div>
        </div>

        <!-- Average Performance Card -->
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border: 1px solid rgba(148, 163, 184, 0.1);
          border-radius: 1rem;
          padding: 1.5rem;
        ">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
            <span style="font-size: 1.25rem; color: #8b5cf6;">🤖</span>
            <span style="font-weight: 500; color: #f8fafc;">Avg Performance</span>
          </div>
          <div style="font-size: 2rem; font-weight: bold; color: #f8fafc;">${avgPerformance.toFixed(1)}%</div>
          <div style="font-size: 0.875rem; color: #94a3b8;">System-wide efficiency</div>
        </div>
      </div>

      <!-- Agents Grid -->
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(350px, 1fr)); gap: 1.5rem;">
        ${agents.map((agent: Agent) => `
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
            overflow: hidden;
            transition: box-shadow 0.2s;
          " onmouseover="this.style.boxShadow='0 10px 25px rgba(0,0,0,0.3)'" onmouseout="this.style.boxShadow='none'">
            <!-- Card Header -->
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <div style="display: flex; align-items: center; gap: 1rem;">
                <span style="font-size: 2rem;">${agent.icon}</span>
                <div style="flex: 1;">
                  <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; margin-bottom: 0.5rem;">
                    ${agent.name}
                  </h3>
                  <span style="
                    padding: 0.25rem 0.5rem;
                    background: ${getStatusColor(agent.status)};
                    color: white;
                    border-radius: 0.25rem;
                    font-size: 0.75rem;
                    font-weight: 500;
                    display: inline-flex;
                    align-items: center;
                    gap: 0.25rem;
                  ">
                    ${getStatusIcon(agent.status)} ${agent.status}
                  </span>
                </div>
              </div>
            </div>

            <!-- Card Content -->
            <div style="padding: 1.5rem;">
              <p style="font-size: 0.875rem; color: #94a3b8; margin-bottom: 1rem;">
                ${agent.description}
              </p>

              <!-- Stats Grid -->
              <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 1rem;">
                <div>
                  <div style="font-size: 0.875rem; color: #94a3b8;">Tasks</div>
                  <div style="font-weight: 500; color: #f8fafc;">${agent.tasksCompleted}</div>
                </div>
                <div>
                  <div style="font-size: 0.875rem; color: #94a3b8;">Uptime</div>
                  <div style="font-weight: 500; color: #f8fafc;">${agent.uptime}%</div>
                </div>
              </div>

              <!-- Performance Bar -->
              <div style="margin-bottom: 1rem;">
                <div style="display: flex; justify-content: space-between; font-size: 0.875rem; margin-bottom: 0.5rem;">
                  <span style="color: #94a3b8;">Performance</span>
                  <span style="font-weight: 500; color: #f8fafc;">${agent.performance}%</span>
                </div>
                <div style="width: 100%; background: rgba(30, 41, 59, 0.8); border-radius: 9999px; height: 0.5rem;">
                  <div style="
                    height: 0.5rem;
                    border-radius: 9999px;
                    transition: all 0.3s;
                    width: ${agent.performance}%;
                    background: ${agent.performance >= 90 ? '#10b981' : agent.performance >= 70 ? '#f59e0b' : '#ef4444'};
                  "></div>
                </div>
              </div>

              <!-- Last Activity -->
              <div style="font-size: 0.75rem; color: #94a3b8; margin-bottom: 1rem;">
                Last activity: ${formatRelativeTime(agent.lastActivity)}
              </div>

              <!-- Action Buttons -->
              <div style="display: flex; gap: 0.5rem;">
                <button style="
                  flex: 1;
                  padding: 0.5rem 1rem;
                  background: ${agent.status === 'maintenance' ? 'rgba(107, 114, 128, 0.5)' : 'rgba(15, 23, 42, 0.8)'};
                  border: 1px solid rgba(148, 163, 184, 0.3);
                  border-radius: 0.5rem;
                  color: ${agent.status === 'maintenance' ? '#6b7280' : '#f8fafc'};
                  cursor: ${agent.status === 'maintenance' ? 'not-allowed' : 'pointer'};
                  font-size: 0.875rem;
                  display: flex;
                  align-items: center;
                  justify-content: center;
                  gap: 0.25rem;
                " ${agent.status === 'maintenance' ? 'disabled' : ''}>
                  ${agent.status === 'active' ? '⏸️ Pause' : '▶️ Start'}
                </button>
                <button style="
                  padding: 0.5rem;
                  background: rgba(15, 23, 42, 0.8);
                  border: 1px solid rgba(148, 163, 184, 0.3);
                  border-radius: 0.5rem;
                  color: #f8fafc;
                  cursor: pointer;
                  font-size: 0.875rem;
                ">
                  ⚙️
                </button>
              </div>
            </div>
          </div>
        `).join('')}
      </div>

      <!-- Setup Notice -->
      <div style="
        background: rgba(217, 119, 6, 0.1);
        border: 1px solid rgba(217, 119, 6, 0.3);
        border-radius: 0.75rem;
        padding: 1.5rem;
        text-align: center;
        margin-top: 2rem;
      ">
        <div style="font-size: 1.125rem; font-weight: 600; color: #f59e0b; margin-bottom: 0.5rem;">
          🚀 AI Agents Ready
        </div>
        <div style="color: #94a3b8; margin-bottom: 1rem;">
          Install dependencies to enable real-time agent management and monitoring
        </div>
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border-radius: 0.5rem;
          padding: 0.75rem;
          font-family: monospace;
          font-size: 0.875rem;
          color: #10b981;
        ">
          npm install && npm run dev
        </div>
      </div>
    </div>
  `
}
