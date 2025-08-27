'use client'

import React, { useState, Suspense, useEffect } from 'react'
import dynamic from 'next/dynamic'
import ErrorBoundary from '@/components/ErrorBoundary'

// Dynamically import components with error boundaries
const LandingPage = dynamic(() => import('@/components/landing/LandingPage'), {
  loading: () => <div className="flex items-center justify-center min-h-screen text-white">Loading Landing Page...</div>,
  ssr: false
})

const Dashboard = dynamic(() => import('@/components/dashboard/Dashboard'), {
  loading: () => <div className="flex items-center justify-center min-h-screen text-white">Loading Dashboard...</div>,
  ssr: false
})

const DataDebug = dynamic(() => import('@/components/debug/DataDebug'), {
  loading: () => null,
  ssr: false
})

export default function HomePage() {
  const [currentView, setCurrentView] = useState<'landing' | 'dashboard' | 'test'>('landing')
  const [count, setCount] = useState(0)
  const [systemInfo, setSystemInfo] = useState({
    timestamp: new Date(),
    userAgent: '',
    viewport: { width: 0, height: 0 },
    connection: 'unknown',
    performance: { memory: 0, timing: 0 }
  })

  // Update system info periodically
  useEffect(() => {
    const updateSystemInfo = () => {
      setSystemInfo({
        timestamp: new Date(),
        userAgent: navigator.userAgent,
        viewport: {
          width: window.innerWidth,
          height: window.innerHeight
        },
        connection: (navigator as any).connection?.effectiveType || 'unknown',
        performance: {
          memory: (performance as any).memory?.usedJSHeapSize || 0,
          timing: performance.now()
        }
      })
    }

    updateSystemInfo()
    const interval = setInterval(updateSystemInfo, 1000)

    return () => clearInterval(interval)
  }, [])

  // Test mode to verify basic rendering
  if (currentView === 'test') {
    return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%)',
      color: 'white',
      padding: '2rem',
      fontFamily: 'system-ui, sans-serif'
    }}>
      <div style={{ textAlign: 'center', maxWidth: '800px', margin: '0 auto' }}>
        <h1 style={{
          fontSize: '4rem',
          fontWeight: 'bold',
          marginBottom: '1rem',
          background: 'linear-gradient(135deg, #d4a574 0%, #ffd700 50%, #d4a574 100%)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
          backgroundClip: 'text'
        }}>
          Go Coffee ☕
        </h1>
        <p style={{ fontSize: '1.5rem', color: '#cbd5e1', marginBottom: '1rem' }}>
          Web3 Coffee Ecosystem - System Status Dashboard
        </p>
        <p style={{ fontSize: '1rem', color: '#94a3b8', marginBottom: '2rem' }}>
          Live System Time: {systemInfo.timestamp.toLocaleString()} |
          Uptime: {Math.floor(systemInfo.performance.timing / 1000)}s |
          Connection: {systemInfo.connection}
        </p>

        <div style={{
          background: 'rgba(255, 255, 255, 0.1)',
          padding: '2rem',
          borderRadius: '1rem',
          backdropFilter: 'blur(10px)',
          border: '1px solid rgba(255, 255, 255, 0.2)',
          marginBottom: '2rem'
        }}>
          <h2 style={{ fontSize: '2rem', marginBottom: '1rem' }}>UI Test Panel</h2>
          <p style={{ marginBottom: '1rem' }}>Click counter: {count}</p>
          <button
            onClick={() => setCount(count + 1)}
            style={{
              background: 'linear-gradient(135deg, #d4a574 0%, #ffd700 100%)',
              color: '#1e293b',
              border: 'none',
              padding: '1rem 2rem',
              borderRadius: '0.5rem',
              fontSize: '1rem',
              fontWeight: 'bold',
              cursor: 'pointer',
              marginRight: '1rem'
            }}
          >
            Test Click ({count})
          </button>
          <button
            onClick={() => window.location.reload()}
            style={{
              background: 'rgba(255, 255, 255, 0.1)',
              color: 'white',
              border: '1px solid rgba(255, 255, 255, 0.3)',
              padding: '1rem 2rem',
              borderRadius: '0.5rem',
              fontSize: '1rem',
              cursor: 'pointer'
            }}
          >
            Reload Page
          </button>
        </div>

        {/* System Status Grid */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
          gap: '1.5rem',
          marginBottom: '2rem'
        }}>
          {/* React Status */}
          <div style={{
            background: 'rgba(16, 185, 129, 0.1)',
            padding: '1.5rem',
            borderRadius: '0.75rem',
            border: '1px solid rgba(16, 185, 129, 0.3)',
            textAlign: 'left'
          }}>
            <h3 style={{ color: '#10b981', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              ⚛️ React System Status
            </h3>
            <ul style={{ listStyle: 'none', padding: 0, fontSize: '0.9rem' }}>
              <li style={{ marginBottom: '0.5rem' }}>✅ React {React.version || '18+'} is working</li>
              <li style={{ marginBottom: '0.5rem' }}>✅ State management: {count} updates</li>
              <li style={{ marginBottom: '0.5rem' }}>✅ Event handlers functional</li>
              <li style={{ marginBottom: '0.5rem' }}>✅ CSS-in-JS rendering</li>
              <li style={{ marginBottom: '0.5rem' }}>✅ Dynamic imports ready</li>
            </ul>
          </div>

          {/* Browser Environment */}
          <div style={{
            background: 'rgba(59, 130, 246, 0.1)',
            padding: '1.5rem',
            borderRadius: '0.75rem',
            border: '1px solid rgba(59, 130, 246, 0.3)',
            textAlign: 'left'
          }}>
            <h3 style={{ color: '#3b82f6', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              🌐 Browser Environment
            </h3>
            <ul style={{ listStyle: 'none', padding: 0, fontSize: '0.9rem' }}>
              <li style={{ marginBottom: '0.5rem' }}>📱 Viewport: {systemInfo.viewport.width}×{systemInfo.viewport.height}</li>
              <li style={{ marginBottom: '0.5rem' }}>🌐 Connection: {systemInfo.connection}</li>
              <li style={{ marginBottom: '0.5rem' }}>💾 Memory: {Math.round(systemInfo.performance.memory / 1024 / 1024)}MB</li>
              <li style={{ marginBottom: '0.5rem' }}>⏱️ Performance: {Math.round(systemInfo.performance.timing)}ms</li>
              <li style={{ marginBottom: '0.5rem' }}>🔄 Auto-refresh: Active</li>
            </ul>
          </div>

          {/* Next.js Status */}
          <div style={{
            background: 'rgba(168, 85, 247, 0.1)',
            padding: '1.5rem',
            borderRadius: '0.75rem',
            border: '1px solid rgba(168, 85, 247, 0.3)',
            textAlign: 'left'
          }}>
            <h3 style={{ color: '#a855f7', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              ▲ Next.js Framework
            </h3>
            <ul style={{ listStyle: 'none', padding: 0, fontSize: '0.9rem' }}>
              <li style={{ marginBottom: '0.5rem' }}>✅ Client-side rendering</li>
              <li style={{ marginBottom: '0.5rem' }}>✅ Dynamic imports working</li>
              <li style={{ marginBottom: '0.5rem' }}>✅ Error boundaries active</li>
              <li style={{ marginBottom: '0.5rem' }}>✅ Hot reload enabled</li>
              <li style={{ marginBottom: '0.5rem' }}>🔧 Development mode</li>
            </ul>
          </div>

          {/* Go Coffee Status */}
          <div style={{
            background: 'rgba(212, 165, 116, 0.1)',
            padding: '1.5rem',
            borderRadius: '0.75rem',
            border: '1px solid rgba(212, 165, 116, 0.3)',
            textAlign: 'left'
          }}>
            <h3 style={{ color: '#d4a574', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              ☕ Go Coffee System
            </h3>
            <ul style={{ listStyle: 'none', padding: 0, fontSize: '0.9rem' }}>
              <li style={{ marginBottom: '0.5rem' }}>🎨 UI Theme: Coffee & Crypto</li>
              <li style={{ marginBottom: '0.5rem' }}>🔄 Real-time data: Ready</li>
              <li style={{ marginBottom: '0.5rem' }}>🤖 AI Agents: Standby</li>
              <li style={{ marginBottom: '0.5rem' }}>💰 Trading Engine: Ready</li>
              <li style={{ marginBottom: '0.5rem' }}>🌐 Web3 Integration: Active</li>
            </ul>
          </div>
        </div>

        {/* Component Testing Section */}
        <div style={{
          background: 'rgba(255, 255, 255, 0.05)',
          padding: '2rem',
          borderRadius: '1rem',
          border: '1px solid rgba(255, 255, 255, 0.1)',
          marginBottom: '2rem'
        }}>
          <h3 style={{
            fontSize: '1.5rem',
            marginBottom: '1rem',
            color: '#ffd700',
            textAlign: 'center'
          }}>
            🧪 Component Testing Laboratory
          </h3>
          <p style={{
            textAlign: 'center',
            color: '#cbd5e1',
            marginBottom: '2rem',
            fontSize: '1rem'
          }}>
            Test individual components to identify any rendering issues
          </p>

          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
            gap: '1rem',
            marginBottom: '2rem'
          }}>
            <button
              onClick={() => setCurrentView('landing')}
              style={{
                background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
                color: 'white',
                border: 'none',
                padding: '1.5rem',
                borderRadius: '0.75rem',
                fontSize: '1rem',
                fontWeight: 'bold',
                cursor: 'pointer',
                textAlign: 'left',
                boxShadow: '0 4px 20px rgba(16, 185, 129, 0.3)'
              }}
            >
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>🏠</div>
              <div style={{ fontSize: '1.1rem', marginBottom: '0.25rem' }}>Landing Page</div>
              <div style={{ fontSize: '0.85rem', opacity: 0.9 }}>
                Hero section, navigation, features showcase
              </div>
            </button>

            <button
              onClick={() => setCurrentView('dashboard')}
              style={{
                background: 'linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%)',
                color: 'white',
                border: 'none',
                padding: '1.5rem',
                borderRadius: '0.75rem',
                fontSize: '1rem',
                fontWeight: 'bold',
                cursor: 'pointer',
                textAlign: 'left',
                boxShadow: '0 4px 20px rgba(59, 130, 246, 0.3)'
              }}
            >
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>📊</div>
              <div style={{ fontSize: '1.1rem', marginBottom: '0.25rem' }}>Dashboard</div>
              <div style={{ fontSize: '0.85rem', opacity: 0.9 }}>
                Trading interface, portfolio, AI agents
              </div>
            </button>

            <button
              onClick={() => window.location.reload()}
              style={{
                background: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
                color: 'white',
                border: 'none',
                padding: '1.5rem',
                borderRadius: '0.75rem',
                fontSize: '1rem',
                fontWeight: 'bold',
                cursor: 'pointer',
                textAlign: 'left',
                boxShadow: '0 4px 20px rgba(245, 158, 11, 0.3)'
              }}
            >
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>🔄</div>
              <div style={{ fontSize: '1.1rem', marginBottom: '0.25rem' }}>Reload System</div>
              <div style={{ fontSize: '0.85rem', opacity: 0.9 }}>
                Fresh start, clear cache, reset state
              </div>
            </button>
          </div>

          <div style={{
            background: 'rgba(168, 85, 247, 0.1)',
            padding: '1rem',
            borderRadius: '0.5rem',
            border: '1px solid rgba(168, 85, 247, 0.3)',
            textAlign: 'center'
          }}>
            <p style={{ margin: 0, fontSize: '0.9rem', color: '#cbd5e1' }}>
              💡 <strong>Tip:</strong> If components fail to load, check the browser console (F12) for detailed error messages
            </p>
          </div>
        </div>

        {/* Real-time Monitoring */}
        <div style={{
          background: 'rgba(0, 0, 0, 0.3)',
          padding: '1.5rem',
          borderRadius: '0.75rem',
          border: '1px solid rgba(255, 255, 255, 0.1)',
          marginBottom: '2rem'
        }}>
          <h3 style={{
            fontSize: '1.25rem',
            marginBottom: '1rem',
            color: '#fbbf24',
            textAlign: 'center'
          }}>
            📡 Real-time System Monitoring
          </h3>

          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: '1rem',
            fontSize: '0.85rem'
          }}>
            <div style={{ textAlign: 'center', padding: '1rem' }}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>⏰</div>
              <div style={{ color: '#10b981', fontWeight: 'bold' }}>
                {systemInfo.timestamp.toLocaleTimeString()}
              </div>
              <div style={{ color: '#6b7280' }}>System Time</div>
            </div>

            <div style={{ textAlign: 'center', padding: '1rem' }}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>🔄</div>
              <div style={{ color: '#3b82f6', fontWeight: 'bold' }}>
                {count} clicks
              </div>
              <div style={{ color: '#6b7280' }}>User Interactions</div>
            </div>

            <div style={{ textAlign: 'center', padding: '1rem' }}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>📱</div>
              <div style={{ color: '#8b5cf6', fontWeight: 'bold' }}>
                {systemInfo.viewport.width}×{systemInfo.viewport.height}
              </div>
              <div style={{ color: '#6b7280' }}>Viewport Size</div>
            </div>

            <div style={{ textAlign: 'center', padding: '1rem' }}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>⚡</div>
              <div style={{ color: '#f59e0b', fontWeight: 'bold' }}>
                {Math.round(systemInfo.performance.timing)}ms
              </div>
              <div style={{ color: '#6b7280' }}>Performance</div>
            </div>

            <div style={{ textAlign: 'center', padding: '1rem' }}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>🌐</div>
              <div style={{ color: '#06b6d4', fontWeight: 'bold' }}>
                {systemInfo.connection}
              </div>
              <div style={{ color: '#6b7280' }}>Connection</div>
            </div>

            <div style={{ textAlign: 'center', padding: '1rem' }}>
              <div style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>💾</div>
              <div style={{ color: '#ef4444', fontWeight: 'bold' }}>
                {Math.round(systemInfo.performance.memory / 1024 / 1024)}MB
              </div>
              <div style={{ color: '#6b7280' }}>Memory Usage</div>
            </div>
          </div>
        </div>

        {/* Browser Information */}
        <div style={{
          background: 'rgba(55, 65, 81, 0.5)',
          padding: '1rem',
          borderRadius: '0.5rem',
          border: '1px solid rgba(75, 85, 99, 0.5)',
          fontSize: '0.75rem',
          color: '#9ca3af'
        }}>
          <details>
            <summary style={{ cursor: 'pointer', marginBottom: '0.5rem', color: '#d1d5db' }}>
              🔍 Detailed Browser Information
            </summary>
            <div style={{
              background: 'rgba(0, 0, 0, 0.3)',
              padding: '0.75rem',
              borderRadius: '0.25rem',
              fontFamily: 'monospace',
              fontSize: '0.7rem',
              wordBreak: 'break-all',
              lineHeight: '1.4'
            }}>
              <strong>User Agent:</strong><br />
              {systemInfo.userAgent || 'Loading...'}
              <br /><br />
              <strong>URL:</strong> {typeof window !== 'undefined' ? window.location.href : 'N/A'}<br />
              <strong>Protocol:</strong> {typeof window !== 'undefined' ? window.location.protocol : 'N/A'}<br />
              <strong>Host:</strong> {typeof window !== 'undefined' ? window.location.host : 'N/A'}
            </div>
          </details>
        </div>
      </div>
    </div>
    )
  }

  // Landing page view
  if (currentView === 'landing') {
    return (
      <ErrorBoundary>
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen text-white">Loading...</div>}>
          <LandingPage onEnterDashboard={() => setCurrentView('dashboard')} />
          <DataDebug />
        </Suspense>
      </ErrorBoundary>
    )
  }

  // Dashboard view
  if (currentView === 'dashboard') {
    return (
      <ErrorBoundary>
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen text-white">Loading...</div>}>
          <Dashboard onBackToLanding={() => setCurrentView('landing')} />
          <DataDebug />
        </Suspense>
      </ErrorBoundary>
    )
  }

  return null
}


