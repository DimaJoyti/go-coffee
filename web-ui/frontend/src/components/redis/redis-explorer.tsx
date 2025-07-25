// Simplified Redis Explorer for when dependencies are not installed
// This provides a basic Redis data visualization interface

interface RedisExplorerProps {
  className?: string
}

export function RedisExplorer({ className }: RedisExplorerProps) {
  // Mock state for when React hooks aren't available
  const searchPattern = '*'
  const selectedKey: string | null = 'user:1001'
  const dataType = 'all'

  // Mock data for demonstration
  const mockKeys = [
    { key: 'user:1001', type: 'hash', ttl: -1, field_count: 5 },
    { key: 'session:abc123', type: 'string', ttl: 3600 },
    { key: 'orders:queue', type: 'list', ttl: -1, length: 25 },
    { key: 'coffee:prices', type: 'zset', ttl: -1, cardinality: 12 },
    { key: 'active:users', type: 'set', ttl: 300, cardinality: 156 },
  ]

  // Helper functions for display
  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'string': return '🔑'
      case 'hash': return '🗂️'
      case 'list': return '📋'
      case 'set': return '👥'
      case 'zset': return '📊'
      default: return '💾'
    }
  }

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'string': return '#3b82f6'
      case 'hash': return '#10b981'
      case 'list': return '#8b5cf6'
      case 'set': return '#f59e0b'
      case 'zset': return '#ef4444'
      default: return '#6b7280'
    }
  }

  const formatTTL = (ttl: number) => {
    if (ttl === -1) return 'No expiry'
    if (ttl === -2) return 'Expired'
    if (ttl < 60) return `${ttl}s`
    if (ttl < 3600) return `${Math.floor(ttl / 60)}m`
    if (ttl < 86400) return `${Math.floor(ttl / 3600)}h`
    return `${Math.floor(ttl / 86400)}d`
  }

  const filteredKeys = mockKeys.filter(key =>
    dataType === 'all' || key.type === dataType
  )

  // Return HTML string since React/JSX isn't available
  return `
    <div class="redis-explorer ${className || ''}" style="padding: 1.5rem; background: #0f172a; color: #f8fafc; min-height: 100vh;">
      <!-- Header -->
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <div>
          <h1 style="font-size: 2rem; font-weight: bold; margin-bottom: 0.5rem; color: #f8fafc;">
            💾 Redis Data Explorer
          </h1>
          <p style="color: #94a3b8; font-size: 1rem;">
            Explore and visualize your Redis data structures
          </p>
        </div>
        <div style="display: flex; gap: 0.5rem;">
          <button style="
            padding: 0.5rem 1rem;
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.3);
            border-radius: 0.5rem;
            color: #f8fafc;
            cursor: pointer;
            font-size: 0.875rem;
          ">
            🔄 Refresh
          </button>
          <button style="
            padding: 0.5rem 1rem;
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.3);
            border-radius: 0.5rem;
            color: #f8fafc;
            cursor: pointer;
            font-size: 0.875rem;
          ">
            📥 Export
          </button>
        </div>
      </div>

      <!-- Search and Filters -->
      <div style="
        background: rgba(15, 23, 42, 0.8);
        border: 1px solid rgba(148, 163, 184, 0.1);
        border-radius: 1rem;
        padding: 1.5rem;
        margin-bottom: 2rem;
      ">
        <div style="display: flex; align-items: center; gap: 1rem; flex-wrap: wrap;">
          <div style="flex: 1; min-width: 300px;">
            <div style="position: relative;">
              <span style="
                position: absolute;
                left: 0.75rem;
                top: 50%;
                transform: translateY(-50%);
                color: #94a3b8;
              ">🔍</span>
              <input
                type="text"
                placeholder="Search keys (use * for wildcards)"
                value="${searchPattern}"
                style="
                  width: 100%;
                  padding: 0.75rem 0.75rem 0.75rem 2.5rem;
                  background: rgba(30, 41, 59, 0.8);
                  border: 1px solid rgba(148, 163, 184, 0.3);
                  border-radius: 0.5rem;
                  color: #f8fafc;
                  font-size: 0.875rem;
                "
              />
            </div>
          </div>

          <select style="
            padding: 0.75rem;
            background: rgba(30, 41, 59, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.3);
            border-radius: 0.5rem;
            color: #f8fafc;
            font-size: 0.875rem;
            min-width: 120px;
          ">
            <option value="all">All Types</option>
            <option value="string">String</option>
            <option value="hash">Hash</option>
            <option value="list">List</option>
            <option value="set">Set</option>
            <option value="zset">Sorted Set</option>
          </select>

          <select style="
            padding: 0.75rem;
            background: rgba(30, 41, 59, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.3);
            border-radius: 0.5rem;
            color: #f8fafc;
            font-size: 0.875rem;
            min-width: 80px;
          ">
            <option value="50">50</option>
            <option value="100" selected>100</option>
            <option value="500">500</option>
            <option value="1000">1000</option>
          </select>

          <button style="
            padding: 0.75rem;
            background: #f59e0b;
            border: none;
            border-radius: 0.5rem;
            color: white;
            cursor: pointer;
            font-size: 0.875rem;
            min-width: 80px;
          ">
            🔍 Search
          </button>
        </div>
      </div>

      <!-- Main Content -->
      <div style="display: grid; grid-template-columns: 1fr 2fr; gap: 2rem;">
        <!-- Keys List -->
        <div>
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
            overflow: hidden;
          ">
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; display: flex; align-items: center; gap: 0.5rem;">
                💾 Keys (${filteredKeys.length})
              </h3>
            </div>
            <div style="max-height: 400px; overflow-y: auto;">
              ${filteredKeys.map((key) => `
                <div style="
                  padding: 1rem;
                  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
                  cursor: pointer;
                  transition: background-color 0.2s;
                  ${selectedKey === key.key ? 'background: rgba(148, 163, 184, 0.1);' : ''}
                " onmouseover="this.style.background='rgba(148, 163, 184, 0.05)'" onmouseout="this.style.background='${selectedKey === key.key ? 'rgba(148, 163, 184, 0.1)' : 'transparent'}'">
                  <div style="display: flex; justify-content: space-between; align-items: center;">
                    <div style="display: flex; align-items: center; gap: 0.5rem; min-width: 0; flex: 1;">
                      <span style="font-size: 1rem;">${getTypeIcon(key.type)}</span>
                      <span style="font-family: monospace; font-size: 0.875rem; color: #f8fafc; overflow: hidden; text-overflow: ellipsis;">
                        ${key.key}
                      </span>
                    </div>
                    <div style="display: flex; align-items: center; gap: 0.5rem;">
                      <span style="
                        padding: 0.25rem 0.5rem;
                        background: ${getTypeColor(key.type)};
                        color: white;
                        border-radius: 0.25rem;
                        font-size: 0.75rem;
                        font-weight: 500;
                      ">
                        ${key.type}
                      </span>
                      ${key.ttl > 0 ? `
                        <span style="
                          padding: 0.25rem 0.5rem;
                          border: 1px solid rgba(148, 163, 184, 0.3);
                          border-radius: 0.25rem;
                          font-size: 0.75rem;
                          color: #94a3b8;
                        ">
                          ⏰ ${formatTTL(key.ttl)}
                        </span>
                      ` : ''}
                    </div>
                  </div>
                  ${(key.length || key.cardinality || key.field_count) ? `
                    <div style="margin-top: 0.5rem; font-size: 0.75rem; color: #94a3b8;">
                      ${key.length ? `Length: ${key.length}` : ''}
                      ${key.cardinality ? `Size: ${key.cardinality}` : ''}
                      ${key.field_count ? `Fields: ${key.field_count}` : ''}
                    </div>
                  ` : ''}
                </div>
              `).join('')}
            </div>
          </div>
        </div>

        <!-- Key Details -->
        <div>
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
            overflow: hidden;
          ">
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; display: flex; align-items: center; gap: 0.5rem;">
                👁️ ${selectedKey ? `Key Details: ${selectedKey}` : 'Select a key to view details'}
              </h3>
            </div>
            <div style="padding: 1.5rem;">
              ${selectedKey ? `
                <!-- Tabs -->
                <div style="margin-bottom: 1.5rem;">
                  <div style="display: flex; gap: 0.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1); padding-bottom: 1rem;">
                    <button style="
                      padding: 0.5rem 1rem;
                      background: #f59e0b;
                      border: none;
                      border-radius: 0.5rem;
                      color: white;
                      cursor: pointer;
                      font-size: 0.875rem;
                    ">
                      Overview
                    </button>
                    <button style="
                      padding: 0.5rem 1rem;
                      background: transparent;
                      border: none;
                      border-radius: 0.5rem;
                      color: #94a3b8;
                      cursor: pointer;
                      font-size: 0.875rem;
                    ">
                      Data
                    </button>
                    <button style="
                      padding: 0.5rem 1rem;
                      background: transparent;
                      border: none;
                      border-radius: 0.5rem;
                      color: #94a3b8;
                      cursor: pointer;
                      font-size: 0.875rem;
                    ">
                      Actions
                    </button>
                  </div>
                </div>

                <!-- Overview Tab Content -->
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                  <div>
                    <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Key</label>
                    <div style="
                      font-family: monospace;
                      font-size: 0.875rem;
                      background: rgba(30, 41, 59, 0.8);
                      padding: 0.5rem;
                      border-radius: 0.5rem;
                      color: #f8fafc;
                    ">
                      ${selectedKey}
                    </div>
                  </div>
                  <div>
                    <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Type</label>
                    <span style="
                      padding: 0.25rem 0.5rem;
                      background: ${getTypeColor('string')};
                      color: white;
                      border-radius: 0.25rem;
                      font-size: 0.75rem;
                      font-weight: 500;
                      display: inline-flex;
                      align-items: center;
                      gap: 0.25rem;
                    ">
                      ${getTypeIcon('string')} string
                    </span>
                  </div>
                  <div>
                    <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">TTL</label>
                    <div style="font-size: 0.875rem; color: #94a3b8;">No expiry</div>
                  </div>
                  <div>
                    <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Memory Usage</label>
                    <div style="font-size: 0.875rem; color: #94a3b8;">2.5 KB</div>
                  </div>
                </div>

                <!-- Action Buttons -->
                <div style="margin-top: 2rem; display: flex; gap: 0.5rem;">
                  <button style="
                    padding: 0.5rem 1rem;
                    background: rgba(15, 23, 42, 0.8);
                    border: 1px solid rgba(148, 163, 184, 0.3);
                    border-radius: 0.5rem;
                    color: #f8fafc;
                    cursor: pointer;
                    font-size: 0.875rem;
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                  ">
                    👁️ View Raw
                  </button>
                  <button style="
                    padding: 0.5rem 1rem;
                    background: rgba(15, 23, 42, 0.8);
                    border: 1px solid rgba(148, 163, 184, 0.3);
                    border-radius: 0.5rem;
                    color: #f8fafc;
                    cursor: pointer;
                    font-size: 0.875rem;
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                  ">
                    📥 Export
                  </button>
                  <button style="
                    padding: 0.5rem 1rem;
                    background: #ef4444;
                    border: none;
                    border-radius: 0.5rem;
                    color: white;
                    cursor: pointer;
                    font-size: 0.875rem;
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                  ">
                    🗑️ Delete
                  </button>
                </div>
              ` : `
                <div style="text-align: center; color: #94a3b8; padding: 2rem;">
                  Select a key from the list to view its details
                </div>
              `}
            </div>
          </div>
        </div>
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
          🚀 Redis Explorer Ready
        </div>
        <div style="color: #94a3b8; margin-bottom: 1rem;">
          Install dependencies to enable full Redis connectivity and real-time data exploration
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
