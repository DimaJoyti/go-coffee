// Simplified Redis Query Builder for when dependencies are not installed
// This provides a basic Redis command building interface

interface QueryBuilderProps {
  className?: string
}

interface QueryTemplate {
  name: string
  description: string
  operation: string
  key?: string
  field?: string
  value?: any
  args?: any[]
  example: string
}

export function QueryBuilder({ className }: QueryBuilderProps) {
  // Mock state for when React hooks aren't available
  const key = 'user:1001'

  // Mock data for demonstration
  const mockTemplates: QueryTemplate[] = [
    {
      name: 'Get User Data',
      description: 'Retrieve user information from hash',
      operation: 'HGET',
      key: 'user:1001',
      field: 'name',
      example: 'HGET user:1001 name'
    },
    {
      name: 'Set Cache Value',
      description: 'Store a value with expiration',
      operation: 'SET',
      key: 'cache:session',
      value: 'session_data',
      example: 'SET cache:session session_data'
    },
    {
      name: 'Add to Queue',
      description: 'Push item to processing queue',
      operation: 'LPUSH',
      key: 'queue:orders',
      value: 'order_123',
      example: 'LPUSH queue:orders order_123'
    }
  ]

  const mockSuggestions = [
    'user:*',
    'session:*',
    'cache:*',
    'queue:*',
    'orders:*'
  ]

  const queryResult = {
    success: true,
    redis_cmd: 'GET user:1001',
    result: '{"name": "John Doe", "email": "john@example.com"}',
    preview: 'This command will retrieve the value stored at key "user:1001"'
  }

  // Return HTML string since React/JSX isn't available
  return `
    <div class="redis-query-builder ${className || ''}" style="padding: 1.5rem; background: #0f172a; color: #f8fafc; min-height: 100vh;">
      <!-- Header -->
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <div>
          <h1 style="font-size: 2rem; font-weight: bold; margin-bottom: 0.5rem; color: #f8fafc;">
            🔧 Redis Query Builder
          </h1>
          <p style="color: #94a3b8; font-size: 1rem;">
            Build and execute Redis commands visually
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
            display: flex;
            align-items: center;
            gap: 0.5rem;
          ">
            📜 History
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
            💾 Save Query
          </button>
        </div>
      </div>

      <!-- Main Content Grid -->
      <div style="display: grid; grid-template-columns: 2fr 1fr; gap: 2rem;">
        <!-- Query Builder -->
        <div>
          <!-- Query Builder Card -->
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
            margin-bottom: 2rem;
          ">
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; display: flex; align-items: center; gap: 0.5rem;">
                💻 Query Builder
              </h3>
            </div>
            <div style="padding: 1.5rem;">
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
                    Visual Builder
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
                    Raw Command
                  </button>
                </div>
              </div>

              <!-- Operation Selection -->
              <div style="margin-bottom: 1rem;">
                <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Operation</label>
                <select style="
                  width: 100%;
                  padding: 0.75rem;
                  background: rgba(30, 41, 59, 0.8);
                  border: 1px solid rgba(148, 163, 184, 0.3);
                  border-radius: 0.5rem;
                  color: #f8fafc;
                  font-size: 0.875rem;
                ">
                  <option value="GET" selected>GET - Get string value</option>
                  <option value="SET">SET - Set string value</option>
                  <option value="HGET">HGET - Get hash field</option>
                  <option value="HSET">HSET - Set hash field</option>
                  <option value="LPUSH">LPUSH - Push to list head</option>
                  <option value="RPUSH">RPUSH - Push to list tail</option>
                </select>
              </div>

              <!-- Key Input -->
              <div style="margin-bottom: 1rem;">
                <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Key</label>
                <input
                  type="text"
                  placeholder="Enter Redis key"
                  value="${key}"
                  style="
                    width: 100%;
                    padding: 0.75rem;
                    background: rgba(30, 41, 59, 0.8);
                    border: 1px solid rgba(148, 163, 184, 0.3);
                    border-radius: 0.5rem;
                    color: #f8fafc;
                    font-size: 0.875rem;
                  "
                />
              </div>

              <!-- Preview Mode Toggle -->
              <div style="margin-bottom: 1.5rem; display: flex; align-items: center; gap: 0.5rem;">
                <input type="checkbox" id="preview" checked style="margin: 0;" />
                <label for="preview" style="font-size: 0.875rem; color: #f8fafc;">Preview mode (don't execute)</label>
              </div>

              <!-- Action Buttons -->
              <div style="display: flex; gap: 0.5rem;">
                <button style="
                  padding: 0.75rem 1rem;
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
                  ✅ Validate
                </button>
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
                  💻 Build Query
                </button>
                <button style="
                  padding: 0.75rem 1rem;
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
                  📋 Copy
                </button>
              </div>
            </div>
          </div>

          <!-- Query Result -->
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
          ">
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; display: flex; align-items: center; gap: 0.5rem;">
                ✅ Query Result
              </h3>
            </div>
            <div style="padding: 1.5rem;">
              <div style="margin-bottom: 1rem;">
                <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Generated Command:</label>
                <pre style="
                  background: rgba(30, 41, 59, 0.8);
                  padding: 0.75rem;
                  border-radius: 0.5rem;
                  font-size: 0.875rem;
                  font-family: monospace;
                  color: #10b981;
                  margin: 0;
                  overflow-x: auto;
                ">${queryResult.redis_cmd}</pre>
              </div>
              <div style="margin-bottom: 1rem;">
                <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Result:</label>
                <pre style="
                  background: rgba(30, 41, 59, 0.8);
                  padding: 0.75rem;
                  border-radius: 0.5rem;
                  font-size: 0.875rem;
                  color: #f8fafc;
                  margin: 0;
                  overflow-x: auto;
                ">${queryResult.result}</pre>
              </div>
              <div>
                <label style="font-size: 0.875rem; font-weight: 500; color: #f8fafc; display: block; margin-bottom: 0.5rem;">Preview:</label>
                <p style="font-size: 0.875rem; color: #94a3b8; margin: 0;">
                  ${queryResult.preview}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Templates and Suggestions -->
        <div>
          <!-- Templates -->
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
            margin-bottom: 2rem;
          ">
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; display: flex; align-items: center; gap: 0.5rem;">
                💡 Templates
              </h3>
            </div>
            <div style="padding: 1.5rem;">
              ${mockTemplates.map(template => `
                <div style="
                  padding: 1rem;
                  border: 1px solid rgba(148, 163, 184, 0.1);
                  border-radius: 0.5rem;
                  margin-bottom: 1rem;
                  cursor: pointer;
                  transition: background-color 0.2s;
                " onmouseover="this.style.background='rgba(148, 163, 184, 0.05)'" onmouseout="this.style.background='transparent'">
                  <div style="font-weight: 500; font-size: 0.875rem; color: #f8fafc; margin-bottom: 0.25rem;">
                    ${template.name}
                  </div>
                  <div style="font-size: 0.75rem; color: #94a3b8; margin-bottom: 0.5rem;">
                    ${template.description}
                  </div>
                  <div style="font-size: 0.75rem; font-family: monospace; color: #3b82f6;">
                    ${template.example}
                  </div>
                </div>
              `).join('')}
            </div>
          </div>

          <!-- Suggestions -->
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
          ">
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; display: flex; align-items: center; gap: 0.5rem;">
                ⚡ Suggestions
              </h3>
            </div>
            <div style="padding: 1.5rem;">
              ${mockSuggestions.map(suggestion => `
                <div style="
                  font-size: 0.75rem;
                  font-family: monospace;
                  padding: 0.5rem;
                  background: rgba(30, 41, 59, 0.8);
                  border-radius: 0.5rem;
                  margin-bottom: 0.5rem;
                  cursor: pointer;
                  color: #f8fafc;
                " onmouseover="this.style.background='rgba(30, 41, 59, 0.6)'" onmouseout="this.style.background='rgba(30, 41, 59, 0.8)'">
                  ${suggestion}
                </div>
              `).join('')}
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
          🚀 Redis Query Builder Ready
        </div>
        <div style="color: #94a3b8; margin-bottom: 1rem;">
          Install dependencies to enable full Redis connectivity and query execution
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
