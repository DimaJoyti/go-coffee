import axios from 'axios'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090'

// Create axios instance with default config
const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized access
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API endpoints
export const dashboardAPI = {
  getMetrics: () => api.get('/dashboard/metrics'),
  getActivity: () => api.get('/dashboard/activity'),
}

export const coffeeAPI = {
  getOrders: (params?: any) => api.get('/coffee/orders', { params }),
  createOrder: (data: any) => api.post('/coffee/orders', data),
  updateOrder: (id: string, data: any) => api.put(`/coffee/orders/${id}`, data),
  getInventory: () => api.get('/coffee/inventory'),
}

export const defiAPI = {
  getPortfolio: () => api.get('/defi/portfolio'),
  getAssets: () => api.get('/defi/assets'),
  getStrategies: () => api.get('/defi/strategies'),
  toggleStrategy: (id: string) => api.post(`/defi/strategies/${id}/toggle`),
}

export const agentsAPI = {
  getStatus: () => api.get('/agents/status'),
  toggleAgent: (id: string) => api.post(`/agents/agents/${id}/toggle`),
  getAgentLogs: (id: string) => api.get(`/agents/agents/${id}/logs`),
}

export const scrapingAPI = {
  getMarketData: (params?: any) => api.get('/scraping/data', { params }),
  refreshData: () => api.post('/scraping/refresh'),
  getDataSources: () => api.get('/scraping/sources'),
}

export const analyticsAPI = {
  getSalesData: (params?: any) => api.get('/analytics/sales', { params }),
  getRevenueData: (params?: any) => api.get('/analytics/revenue', { params }),
  getTopProducts: () => api.get('/analytics/products'),
  getLocationPerformance: () => api.get('/analytics/locations'),
}

// Health check
export const healthAPI = {
  check: () => axios.get(`${API_BASE_URL}/health`),
}

export default api
