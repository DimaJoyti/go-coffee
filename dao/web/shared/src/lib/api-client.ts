import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import {
  APIResponse,
  PaginatedResponse,
  Bounty,
  BountyFilters,
  CreateBountyRequest,
  ApplyForBountyRequest,
  Solution,
  SolutionFilters,
  CreateSolutionRequest,
  CreateReviewRequest,
  TVLMetrics,
  MAUMetrics,
  PerformanceDashboard,
  User,
  DeveloperProfile,
} from '../types/api';

export class APIClient {
  private client: AxiosInstance;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor for auth
    this.client.interceptors.request.use((config) => {
      const token = this.getAuthToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          this.handleAuthError();
        }
        return Promise.reject(error);
      }
    );
  }

  private getAuthToken(): string | null {
    return localStorage.getItem('auth_token');
  }

  private setAuthToken(token: string): void {
    localStorage.setItem('auth_token', token);
  }

  private removeAuthToken(): void {
    localStorage.removeItem('auth_token');
  }

  private handleAuthError(): void {
    this.removeAuthToken();
    // Redirect to login or emit auth error event
    window.dispatchEvent(new CustomEvent('auth:error'));
  }

  // Generic HTTP methods for direct use in hooks
  async get(url: string, config?: AxiosRequestConfig) {
    return this.client.get(url, config);
  }

  async post(url: string, data?: any, config?: AxiosRequestConfig) {
    return this.client.post(url, data, config);
  }

  async put(url: string, data?: any, config?: AxiosRequestConfig) {
    return this.client.put(url, data, config);
  }

  async delete(url: string, config?: AxiosRequestConfig) {
    return this.client.delete(url, config);
  }

  // Authentication
  async login(address: string, signature: string): Promise<APIResponse<{ token: string; user: User }>> {
    const response = await this.client.post('/api/v1/auth/login', {
      address,
      signature,
    });
    
    if (response.data.status === 'success' && response.data.data?.token) {
      this.setAuthToken(response.data.data.token);
    }
    
    return response.data;
  }

  async logout(): Promise<void> {
    try {
      await this.client.post('/api/v1/auth/logout');
    } finally {
      this.removeAuthToken();
    }
  }

  async getCurrentUser(): Promise<APIResponse<User>> {
    const response = await this.client.get('/api/v1/auth/me');
    return response.data;
  }

  // Bounty Management API
  async getBounties(filters?: BountyFilters): Promise<PaginatedResponse<Bounty>> {
    const response = await this.client.get('/api/v1/bounties', {
      params: filters,
    });
    return response.data;
  }

  async getBounty(id: number): Promise<APIResponse<Bounty>> {
    const response = await this.client.get(`/api/v1/bounties/${id}`);
    return response.data;
  }

  async createBounty(data: CreateBountyRequest): Promise<APIResponse<Bounty>> {
    const response = await this.client.post('/api/v1/bounties', data);
    return response.data;
  }

  async applyForBounty(data: ApplyForBountyRequest): Promise<APIResponse<{ application_id: number }>> {
    const response = await this.client.post(`/api/v1/bounties/${data.bounty_id}/apply`, data);
    return response.data;
  }

  async getMyBounties(): Promise<PaginatedResponse<Bounty>> {
    const response = await this.client.get('/api/v1/bounties/my');
    return response.data;
  }

  async getPerformanceDashboard(): Promise<APIResponse<PerformanceDashboard>> {
    const response = await this.client.get('/api/v1/performance/dashboard');
    return response.data;
  }

  async getDeveloperLeaderboard(limit = 10): Promise<APIResponse<any[]>> {
    const response = await this.client.get('/api/v1/performance/leaderboard', {
      params: { limit },
    });
    return response.data;
  }

  // Solution Marketplace API
  async getSolutions(filters?: SolutionFilters): Promise<PaginatedResponse<Solution>> {
    const response = await this.client.get('/api/v1/solutions', {
      params: filters,
    });
    return response.data;
  }

  async getSolution(id: number): Promise<APIResponse<Solution>> {
    const response = await this.client.get(`/api/v1/solutions/${id}`);
    return response.data;
  }

  async createSolution(data: CreateSolutionRequest): Promise<APIResponse<Solution>> {
    const response = await this.client.post('/api/v1/solutions', data);
    return response.data;
  }

  async reviewSolution(data: CreateReviewRequest): Promise<APIResponse<{ review_id: number }>> {
    const response = await this.client.post(`/api/v1/solutions/${data.solution_id}/review`, data);
    return response.data;
  }

  async installSolution(solutionId: number, environment: string): Promise<APIResponse<{ installation_id: number }>> {
    const response = await this.client.post(`/api/v1/solutions/${solutionId}/install`, {
      environment,
    });
    return response.data;
  }

  async getMySolutions(): Promise<PaginatedResponse<Solution>> {
    const response = await this.client.get('/api/v1/solutions/my');
    return response.data;
  }

  async getCategories(): Promise<APIResponse<any[]>> {
    const response = await this.client.get('/api/v1/categories');
    return response.data;
  }

  async getPopularSolutions(limit = 10): Promise<APIResponse<Solution[]>> {
    const response = await this.client.get('/api/v1/analytics/popular', {
      params: { limit },
    });
    return response.data;
  }

  async getTrendingSolutions(limit = 10): Promise<APIResponse<Solution[]>> {
    const response = await this.client.get('/api/v1/analytics/trending', {
      params: { limit },
    });
    return response.data;
  }

  // TVL/MAU Metrics API
  async getTVLMetrics(protocol?: string, chain?: string): Promise<APIResponse<TVLMetrics>> {
    const response = await this.client.get('/api/v1/tvl', {
      params: { protocol, chain },
    });
    return response.data;
  }

  async getMAUMetrics(feature?: string): Promise<APIResponse<MAUMetrics>> {
    const response = await this.client.get('/api/v1/mau', {
      params: { feature },
    });
    return response.data;
  }

  async getTVLHistory(protocol: string, days = 30): Promise<APIResponse<any[]>> {
    const response = await this.client.get('/api/v1/tvl/history', {
      params: { protocol, days },
    });
    return response.data;
  }

  async getMAUHistory(feature: string, months = 12): Promise<APIResponse<any[]>> {
    const response = await this.client.get('/api/v1/mau/history', {
      params: { feature, months },
    });
    return response.data;
  }

  async getAnalyticsOverview(): Promise<APIResponse<any>> {
    const response = await this.client.get('/api/v1/analytics/overview');
    return response.data;
  }

  async getAlerts(): Promise<PaginatedResponse<any>> {
    const response = await this.client.get('/api/v1/analytics/alerts');
    return response.data;
  }

  async getDailyReport(date?: string): Promise<APIResponse<any>> {
    const response = await this.client.get('/api/v1/reports/daily', {
      params: { date },
    });
    return response.data;
  }

  async getWeeklyReport(week?: string): Promise<APIResponse<any>> {
    const response = await this.client.get('/api/v1/reports/weekly', {
      params: { week },
    });
    return response.data;
  }

  async getMonthlyReport(month?: string): Promise<APIResponse<any>> {
    const response = await this.client.get('/api/v1/reports/monthly', {
      params: { month },
    });
    return response.data;
  }

  // User Profile API
  async getDeveloperProfile(address: string): Promise<APIResponse<DeveloperProfile>> {
    const response = await this.client.get(`/api/v1/developers/${address}`);
    return response.data;
  }

  async updateProfile(data: Partial<DeveloperProfile>): Promise<APIResponse<DeveloperProfile>> {
    const response = await this.client.put('/api/v1/profile', data);
    return response.data;
  }

  // Utility methods
  async uploadFile(file: File, type: 'avatar' | 'document' | 'image'): Promise<APIResponse<{ url: string }>> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', type);

    const response = await this.client.post('/api/v1/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  async healthCheck(): Promise<APIResponse<{ status: string }>> {
    const response = await this.client.get('/health');
    return response.data;
  }
}

// Create singleton instance
export const apiClient = new APIClient();

// Export for dependency injection in tests
export default APIClient;
