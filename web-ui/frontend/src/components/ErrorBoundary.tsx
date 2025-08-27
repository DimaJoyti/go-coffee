'use client'

import React from 'react'

interface ErrorBoundaryState {
  hasError: boolean
  error?: Error
  errorInfo?: React.ErrorInfo
}

interface ErrorBoundaryProps {
  children: React.ReactNode
  fallback?: React.ComponentType<{ error?: Error; reset: () => void }>
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
    this.setState({ error, errorInfo })
  }

  render() {
    if (this.state.hasError) {
      const FallbackComponent = this.props.fallback || DefaultErrorFallback
      return (
        <FallbackComponent 
          error={this.state.error} 
          reset={() => this.setState({ hasError: false, error: undefined, errorInfo: undefined })}
        />
      )
    }

    return this.props.children
  }
}

function DefaultErrorFallback({ error, reset }: { error?: Error; reset: () => void }) {
  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%)',
      color: 'white',
      padding: '2rem',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: 'system-ui, sans-serif'
    }}>
      <div style={{
        background: 'rgba(239, 68, 68, 0.1)',
        border: '1px solid rgba(239, 68, 68, 0.3)',
        borderRadius: '1rem',
        padding: '2rem',
        maxWidth: '600px',
        textAlign: 'center'
      }}>
        <h1 style={{ fontSize: '2rem', marginBottom: '1rem', color: '#ef4444' }}>
          🚨 Something went wrong
        </h1>
        <p style={{ marginBottom: '1rem', color: '#cbd5e1' }}>
          An error occurred while rendering the Go Coffee application.
        </p>
        {error && (
          <details style={{ 
            marginBottom: '1rem', 
            textAlign: 'left',
            background: 'rgba(0, 0, 0, 0.3)',
            padding: '1rem',
            borderRadius: '0.5rem',
            fontSize: '0.875rem'
          }}>
            <summary style={{ cursor: 'pointer', marginBottom: '0.5rem' }}>
              Error Details
            </summary>
            <pre style={{ 
              whiteSpace: 'pre-wrap', 
              wordBreak: 'break-word',
              color: '#fca5a5'
            }}>
              {error.name}: {error.message}
              {error.stack && `\n\nStack trace:\n${error.stack}`}
            </pre>
          </details>
        )}
        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
          <button
            onClick={reset}
            style={{
              background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
              color: 'white',
              border: 'none',
              padding: '0.75rem 1.5rem',
              borderRadius: '0.5rem',
              fontSize: '1rem',
              cursor: 'pointer',
              fontWeight: '600'
            }}
          >
            Try Again
          </button>
          <button
            onClick={() => window.location.reload()}
            style={{
              background: 'rgba(255, 255, 255, 0.1)',
              color: 'white',
              border: '1px solid rgba(255, 255, 255, 0.3)',
              padding: '0.75rem 1.5rem',
              borderRadius: '0.5rem',
              fontSize: '1rem',
              cursor: 'pointer'
            }}
          >
            Reload Page
          </button>
        </div>
      </div>
    </div>
  )
}

export default ErrorBoundary
