'use client'

import { useState, useCallback } from 'react'
import useSWR from 'swr'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// Types
interface RedisKey {
  key: string
  type: string
  ttl: number
  memory_usage?: number
  length?: number
  cardinality?: number
  field_count?: number
}

interface DataExplorerRequest {
  data_type: string
  pattern?: string
  key?: string
  limit?: number
  offset?: number
  search_term?: string
}

interface QueryBuilderRequest {
  operation: string
  key: string
  field?: string
  value?: any
  args?: any[]
  preview?: boolean
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

interface QueryResult {
  success: boolean
  redis_cmd?: string
  result?: any
  preview?: string
  validation?: any
  suggestions?: string[]
}

// API functions
const redisAPI = {
  async exploreData(request: DataExplorerRequest) {
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/explore`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
    if (!response.ok) throw new Error('Failed to explore data')
    return response.json()
  },

  async getKeyDetails(key: string) {
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/key/${encodeURIComponent(key)}`)
    if (!response.ok) throw new Error('Failed to get key details')
    return response.json()
  },

  async searchData(query: string, type: string = 'keys', limit: number = 100) {
    const params = new URLSearchParams({ q: query, type, limit: limit.toString() })
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/search?${params}`)
    if (!response.ok) throw new Error('Failed to search data')
    return response.json()
  },

  async buildQuery(request: QueryBuilderRequest) {
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/query/build`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
    if (!response.ok) throw new Error('Failed to build query')
    return response.json()
  },

  async validateQuery(request: QueryBuilderRequest) {
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/query/validate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
    if (!response.ok) throw new Error('Failed to validate query')
    return response.json()
  },

  async getTemplates(operation?: string) {
    const params = operation ? `?operation=${operation}` : ''
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/query/templates${params}`)
    if (!response.ok) throw new Error('Failed to get templates')
    return response.json()
  },

  async getSuggestions(operation?: string, partial?: string) {
    const params = new URLSearchParams()
    if (operation) params.append('operation', operation)
    if (partial) params.append('partial', partial)
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/query/suggestions?${params}`)
    if (!response.ok) throw new Error('Failed to get suggestions')
    return response.json()
  },

  async visualizeData(request: any) {
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/visualize`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
    if (!response.ok) throw new Error('Failed to visualize data')
    return response.json()
  },

  async getMetrics() {
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/metrics`)
    if (!response.ok) throw new Error('Failed to get metrics')
    return response.json()
  },

  async getPerformanceMetrics() {
    const response = await fetch(`${API_BASE}/api/v1/redis-mcp/visual/performance`)
    if (!response.ok) throw new Error('Failed to get performance metrics')
    return response.json()
  },
}

// Hook for Redis data exploration
export function useRedisData() {
  const [keys, setKeys] = useState<RedisKey[]>([])
  const [keyDetails, setKeyDetails] = useState<any>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const exploreKeys = useCallback(async (options: { pattern?: string; limit?: number } = {}) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.exploreData({
        data_type: 'keys',
        pattern: options.pattern || '*',
        limit: options.limit || 100,
      })
      if (response.success) {
        setKeys(response.data || [])
      } else {
        setError('Failed to load keys')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const getKeyDetails = useCallback(async (key: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.getKeyDetails(key)
      if (response.success) {
        setKeyDetails(response.data)
      } else {
        setError('Failed to load key details')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const exploreData = useCallback(async (request: DataExplorerRequest) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.exploreData(request)
      return response
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  const searchData = useCallback(async (query: string, type: string = 'keys') => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.searchData(query, type)
      return response
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return {
    keys,
    keyDetails,
    isLoading,
    error,
    exploreKeys,
    getKeyDetails,
    exploreData,
    searchData,
  }
}

// Hook for Redis query building
export function useRedisQuery() {
  const [queryResult, setQueryResult] = useState<QueryResult | null>(null)
  const [templates, setTemplates] = useState<QueryTemplate[]>([])
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const buildQuery = useCallback(async (request: QueryBuilderRequest) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.buildQuery(request)
      setQueryResult(response)
      return response
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  const validateQuery = useCallback(async (request: QueryBuilderRequest) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.validateQuery(request)
      setQueryResult(response)
      return response
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  const executeQuery = useCallback(async (request: QueryBuilderRequest) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.buildQuery({ ...request, preview: false })
      setQueryResult(response)
      return response
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  const getTemplates = useCallback(async (operation?: string) => {
    try {
      const response = await redisAPI.getTemplates(operation)
      if (response.success) {
        setTemplates(response.templates || [])
      }
    } catch (err) {
      console.error('Failed to load templates:', err)
    }
  }, [])

  const getSuggestions = useCallback(async (operation?: string, partial?: string) => {
    try {
      const response = await redisAPI.getSuggestions(operation, partial)
      if (response.success) {
        setSuggestions(response.suggestions || [])
      }
    } catch (err) {
      console.error('Failed to load suggestions:', err)
    }
  }, [])

  return {
    queryResult,
    templates,
    suggestions,
    isLoading,
    error,
    buildQuery,
    validateQuery,
    executeQuery,
    getTemplates,
    getSuggestions,
  }
}

// Hook for Redis visualization
export function useRedisVisualization() {
  const [visualizationData, setVisualizationData] = useState<any>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { data: metrics } = useSWR('redis-metrics', redisAPI.getMetrics, {
    refreshInterval: 30000,
  })

  const { data: performanceMetrics } = useSWR('redis-performance', redisAPI.getPerformanceMetrics, {
    refreshInterval: 10000,
  })

  const visualizeData = useCallback(async (request: any) => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await redisAPI.visualizeData(request)
      if (response.success) {
        setVisualizationData(response)
      } else {
        setError('Failed to generate visualization')
      }
      return response
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return {
    visualizationData,
    metrics,
    performanceMetrics,
    isLoading,
    error,
    visualizeData,
  }
}
