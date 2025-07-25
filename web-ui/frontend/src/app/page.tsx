// Go Coffee Epic UI - Dashboard Page
// Web3 Coffee Ecosystem Dashboard

export default function HomePage() {
  return (
    <div style={{
      display: 'flex',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
      color: '#f8fafc',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif'
    }}>
      {/* Sidebar */}
      <div style={{
        width: '280px',
        background: 'rgba(15, 23, 42, 0.95)',
        backdropFilter: 'blur(10px)',
        borderRight: '1px solid rgba(148, 163, 184, 0.1)',
        padding: '2rem 1rem'
      }}>
        {/* Logo */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          marginBottom: '3rem',
          padding: '0 1rem'
        }}>
          <div style={{
            width: '2.5rem',
            height: '2.5rem',
            background: 'linear-gradient(45deg, #d97706, #f59e0b)',
            borderRadius: '0.75rem',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '1.25rem'
          }}>
            ☕
          </div>
          <div style={{
            fontSize: '1.5rem',
            fontWeight: 'bold',
            background: 'linear-gradient(45deg, #d97706, #f59e0b)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent'
          }}>
            Go Coffee
          </div>
        </div>

        {/* Navigation */}
        <nav>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.875rem 1rem',
            marginBottom: '0.5rem',
            borderRadius: '0.5rem',
            background: 'rgba(217, 119, 6, 0.1)',
            color: '#f59e0b',
            cursor: 'pointer'
          }}>
            <span>📊</span>
            <span>Dashboard</span>
          </div>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.875rem 1rem',
            marginBottom: '0.5rem',
            borderRadius: '0.5rem',
            color: '#94a3b8',
            cursor: 'pointer'
          }}>
            <span>☕</span>
            <span>Coffee Orders</span>
          </div>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.875rem 1rem',
            marginBottom: '0.5rem',
            borderRadius: '0.5rem',
            color: '#94a3b8',
            cursor: 'pointer'
          }}>
            <span>💰</span>
            <span>DeFi Portfolio</span>
          </div>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.875rem 1rem',
            marginBottom: '0.5rem',
            borderRadius: '0.5rem',
            color: '#94a3b8',
            cursor: 'pointer'
          }}>
            <span>🤖</span>
            <span>AI Agents</span>
          </div>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.875rem 1rem',
            marginBottom: '0.5rem',
            borderRadius: '0.5rem',
            color: '#94a3b8',
            cursor: 'pointer'
          }}>
            <span>🔍</span>
            <span>Data Scraping</span>
          </div>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            padding: '0.875rem 1rem',
            marginBottom: '0.5rem',
            borderRadius: '0.5rem',
            color: '#94a3b8',
            cursor: 'pointer'
          }}>
            <span>📈</span>
            <span>Analytics</span>
          </div>
        </nav>
      </div>

      {/* Main Content */}
      <div style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column'
      }}>
        {/* Header */}
        <div style={{
          background: 'rgba(15, 23, 42, 0.95)',
          backdropFilter: 'blur(10px)',
          borderBottom: '1px solid rgba(148, 163, 184, 0.1)',
          padding: '1rem 2rem',
          display: 'flex',
          alignItems: 'center'
        }}>
          <div style={{
            fontSize: '1.5rem',
            fontWeight: '600'
          }}>
            Dashboard Overview
          </div>
        </div>

        {/* Content Area */}
        <div style={{
          flex: 1,
          padding: '2rem',
          background: 'rgba(30, 41, 59, 0.3)'
        }}>
          {/* Setup Notice */}
          <div style={{
            background: 'rgba(217, 119, 6, 0.1)',
            border: '1px solid rgba(217, 119, 6, 0.3)',
            borderRadius: '0.75rem',
            padding: '1.5rem',
            textAlign: 'center',
            marginBottom: '2rem'
          }}>
            <div style={{
              fontSize: '1.25rem',
              fontWeight: '600',
              color: '#f59e0b',
              marginBottom: '0.5rem'
            }}>
              🚀 Welcome to Go Coffee Epic UI
            </div>
            <div style={{
              color: '#94a3b8',
              marginBottom: '1rem'
            }}>
              Your Web3 Coffee Ecosystem is ready to be configured!
            </div>
            <div style={{
              background: 'rgba(15, 23, 42, 0.8)',
              borderRadius: '0.5rem',
              padding: '0.75rem',
              fontFamily: 'Monaco, Menlo, monospace',
              fontSize: '0.875rem',
              color: '#10b981'
            }}>
              npm install && npm run dev
            </div>
          </div>

          {/* Dashboard Cards */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
            gap: '1.5rem'
          }}>
            {/* Coffee Operations Card */}
            <div style={{
              background: 'rgba(15, 23, 42, 0.8)',
              backdropFilter: 'blur(10px)',
              border: '1px solid rgba(148, 163, 184, 0.1)',
              borderRadius: '1rem',
              padding: '1.5rem'
            }}>
              <div style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                marginBottom: '1rem',
                color: '#f8fafc'
              }}>
                ☕ Coffee Operations
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Total Orders</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>1,247</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Revenue Today</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>$3,456</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Active Locations</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>12</span>
              </div>
            </div>

            {/* DeFi Portfolio Card */}
            <div style={{
              background: 'rgba(15, 23, 42, 0.8)',
              backdropFilter: 'blur(10px)',
              border: '1px solid rgba(148, 163, 184, 0.1)',
              borderRadius: '1rem',
              padding: '1.5rem'
            }}>
              <div style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                marginBottom: '1rem',
                color: '#f8fafc'
              }}>
                💰 DeFi Portfolio
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Total Value</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>$123,456</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>24h Change</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>+5.67%</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Active Strategies</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>8</span>
              </div>
            </div>

            {/* AI Agents Card */}
            <div style={{
              background: 'rgba(15, 23, 42, 0.8)',
              backdropFilter: 'blur(10px)',
              border: '1px solid rgba(148, 163, 184, 0.1)',
              borderRadius: '1rem',
              padding: '1.5rem'
            }}>
              <div style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                marginBottom: '1rem',
                color: '#f8fafc'
              }}>
                🤖 AI Agents
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Active Agents</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>15</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Tasks Completed</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>2,847</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Efficiency</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>94.2%</span>
              </div>
            </div>

            {/* Data Analytics Card */}
            <div style={{
              background: 'rgba(15, 23, 42, 0.8)',
              backdropFilter: 'blur(10px)',
              border: '1px solid rgba(148, 163, 184, 0.1)',
              borderRadius: '1rem',
              padding: '1.5rem'
            }}>
              <div style={{
                fontSize: '1.125rem',
                fontWeight: '600',
                marginBottom: '1rem',
                color: '#f8fafc'
              }}>
                🔍 Data Analytics
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Data Sources</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>25</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.75rem'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Records Processed</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>1.2M</span>
              </div>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center'
              }}>
                <span style={{ color: '#94a3b8', fontSize: '0.875rem' }}>Insights Generated</span>
                <span style={{ fontWeight: '600', color: '#f59e0b' }}>156</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Connection Status */}
      <div style={{
        position: 'fixed',
        bottom: '1rem',
        right: '1rem',
        padding: '0.5rem 1rem',
        borderRadius: '9999px',
        fontSize: '0.875rem',
        fontWeight: '500',
        background: '#ef4444',
        color: 'white'
      }}>
        🔴 Disconnected - Install Dependencies
      </div>
    </div>
  )
}
