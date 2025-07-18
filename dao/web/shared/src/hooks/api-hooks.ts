import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../lib/api-client';
import {
  BountyFilters,
  SolutionFilters,
  CreateBountyRequest,
  CreateSolutionRequest,
  CreateReviewRequest,
  ApplyForBountyRequest,
} from '../types/api';

// Query Keys
export const queryKeys = {
  bounties: (filters?: BountyFilters) => ['bounties', filters],
  bounty: (id: number) => ['bounty', id],
  myBounties: () => ['bounties', 'my'],
  solutions: (filters?: SolutionFilters) => ['solutions', filters],
  solution: (id: number) => ['solution', id],
  mySolutions: () => ['solutions', 'my'],
  categories: () => ['categories'],
  popularSolutions: (limit?: number) => ['solutions', 'popular', limit],
  trendingSolutions: (limit?: number) => ['solutions', 'trending', limit],
  tvlMetrics: (protocol?: string, chain?: string) => ['tvl', protocol, chain],
  mauMetrics: (feature?: string) => ['mau', feature],
  tvlHistory: (protocol: string, days?: number) => ['tvl', 'history', protocol, days],
  mauHistory: (feature: string, months?: number) => ['mau', 'history', feature, months],
  performanceDashboard: () => ['performance', 'dashboard'],
  leaderboard: (limit?: number) => ['performance', 'leaderboard', limit],
  analyticsOverview: () => ['analytics', 'overview'],
  alerts: () => ['alerts'],
  dailyReport: (date?: string) => ['reports', 'daily', date],
  weeklyReport: (week?: string) => ['reports', 'weekly', week],
  monthlyReport: (month?: string) => ['reports', 'monthly', month],
  currentUser: () => ['auth', 'me'],
  developerProfile: (address: string) => ['developer', address],
} as const;

// Bounty Hooks
export const useBounties = (filters?: BountyFilters) => {
  return useQuery({
    queryKey: queryKeys.bounties(filters),
    queryFn: () => apiClient.getBounties(filters),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useBounty = (id: number) => {
  return useQuery({
    queryKey: queryKeys.bounty(id),
    queryFn: () => apiClient.getBounty(id),
    enabled: !!id,
  });
};

export const useMyBounties = () => {
  return useQuery({
    queryKey: queryKeys.myBounties(),
    queryFn: () => apiClient.getMyBounties(),
  });
};

export const useCreateBounty = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateBountyRequest) => apiClient.createBounty(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bounties'] });
    },
  });
};

export const useApplyForBounty = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: ApplyForBountyRequest) => apiClient.applyForBounty(data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.bounty(variables.bounty_id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.myBounties() });
    },
  });
};

// Solution Hooks
export const useSolutions = (filters?: SolutionFilters) => {
  return useQuery({
    queryKey: queryKeys.solutions(filters),
    queryFn: () => apiClient.getSolutions(filters),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useSolution = (id: number) => {
  return useQuery({
    queryKey: queryKeys.solution(id),
    queryFn: () => apiClient.getSolution(id),
    enabled: !!id,
  });
};

export const useMySolutions = () => {
  return useQuery({
    queryKey: queryKeys.mySolutions(),
    queryFn: () => apiClient.getMySolutions(),
  });
};

export const useCreateSolution = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateSolutionRequest) => apiClient.createSolution(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['solutions'] });
    },
  });
};

export const useReviewSolution = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateReviewRequest) => apiClient.reviewSolution(data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.solution(variables.solution_id) });
    },
  });
};

export const useInstallSolution = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ solutionId, environment }: { solutionId: number; environment: string }) =>
      apiClient.installSolution(solutionId, environment),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.solution(variables.solutionId) });
    },
  });
};

export const useCategories = () => {
  return useQuery({
    queryKey: queryKeys.categories(),
    queryFn: () => apiClient.getCategories(),
    staleTime: 30 * 60 * 1000, // 30 minutes
  });
};

export const usePopularSolutions = (limit = 10) => {
  return useQuery({
    queryKey: queryKeys.popularSolutions(limit),
    queryFn: () => apiClient.getPopularSolutions(limit),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
};

export const useTrendingSolutions = (limit = 10) => {
  return useQuery({
    queryKey: queryKeys.trendingSolutions(limit),
    queryFn: () => apiClient.getTrendingSolutions(limit),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
};

// Metrics Hooks
export const useTVLMetrics = (protocol?: string, chain?: string) => {
  return useQuery({
    queryKey: queryKeys.tvlMetrics(protocol, chain),
    queryFn: () => apiClient.getTVLMetrics(protocol, chain),
    refetchInterval: 30 * 1000, // 30 seconds
  });
};

export const useMAUMetrics = (feature?: string) => {
  return useQuery({
    queryKey: queryKeys.mauMetrics(feature),
    queryFn: () => apiClient.getMAUMetrics(feature),
    refetchInterval: 60 * 1000, // 1 minute
  });
};

export const useTVLHistory = (protocol: string, days = 30) => {
  return useQuery({
    queryKey: queryKeys.tvlHistory(protocol, days),
    queryFn: () => apiClient.getTVLHistory(protocol, days),
    enabled: !!protocol,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useMAUHistory = (feature: string, months = 12) => {
  return useQuery({
    queryKey: queryKeys.mauHistory(feature, months),
    queryFn: () => apiClient.getMAUHistory(feature, months),
    enabled: !!feature,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const usePerformanceDashboard = () => {
  return useQuery({
    queryKey: queryKeys.performanceDashboard(),
    queryFn: () => apiClient.getPerformanceDashboard(),
    refetchInterval: 60 * 1000, // 1 minute
  });
};

export const useLeaderboard = (limit = 10) => {
  return useQuery({
    queryKey: queryKeys.leaderboard(limit),
    queryFn: () => apiClient.getDeveloperLeaderboard(limit),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useAnalyticsOverview = () => {
  return useQuery({
    queryKey: queryKeys.analyticsOverview(),
    queryFn: () => apiClient.getAnalyticsOverview(),
    refetchInterval: 60 * 1000, // 1 minute
  });
};

export const useAlerts = () => {
  return useQuery({
    queryKey: queryKeys.alerts(),
    queryFn: () => apiClient.getAlerts(),
    refetchInterval: 30 * 1000, // 30 seconds
  });
};

// Report Hooks
export const useDailyReport = (date?: string) => {
  return useQuery({
    queryKey: queryKeys.dailyReport(date),
    queryFn: () => apiClient.getDailyReport(date),
    staleTime: 60 * 60 * 1000, // 1 hour
  });
};

export const useWeeklyReport = (week?: string) => {
  return useQuery({
    queryKey: queryKeys.weeklyReport(week),
    queryFn: () => apiClient.getWeeklyReport(week),
    staleTime: 60 * 60 * 1000, // 1 hour
  });
};

export const useMonthlyReport = (month?: string) => {
  return useQuery({
    queryKey: queryKeys.monthlyReport(month),
    queryFn: () => apiClient.getMonthlyReport(month),
    staleTime: 60 * 60 * 1000, // 1 hour
  });
};

// Auth Hooks
export const useCurrentUser = () => {
  return useQuery({
    queryKey: queryKeys.currentUser(),
    queryFn: () => apiClient.getCurrentUser(),
    retry: false,
  });
};

export const useLogin = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ address, signature }: { address: string; signature: string }) =>
      apiClient.login(address, signature),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.currentUser() });
    },
  });
};

export const useLogout = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: () => apiClient.logout(),
    onSuccess: () => {
      queryClient.clear();
    },
  });
};

export const useDeveloperProfile = (address: string) => {
  return useQuery({
    queryKey: queryKeys.developerProfile(address),
    queryFn: () => apiClient.getDeveloperProfile(address),
    enabled: !!address,
  });
};

export const useUpdateProfile = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: any) => apiClient.updateProfile(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.currentUser() });
    },
  });
};

// AI Service API Hooks

// Bounty Matching
export const useBountyMatching = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/bounty-matching/match', data).then(res => res.data),
  });
};

export const useDeveloperRecommendations = (bountyId: number) => {
  return useQuery({
    queryKey: ['developer-recommendations', bountyId],
    queryFn: () => apiClient.get(`/ai/bounty-matching/developers/${bountyId}`).then(res => res.data),
    enabled: !!bountyId,
  });
};

export const useBountyRecommendations = (developerAddress: string) => {
  return useQuery({
    queryKey: ['bounty-recommendations', developerAddress],
    queryFn: () => apiClient.get(`/ai/bounty-matching/bounties/${developerAddress}`).then(res => res.data),
    enabled: !!developerAddress,
  });
};

// Quality Assessment
export const useQualityAssessment = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/quality-assessment/assess', data).then(res => res.data),
  });
};

export const useCodeQualityAssessment = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/quality-assessment/assess-code', data).then(res => res.data),
  });
};

export const useRepositoryQualityAssessment = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/quality-assessment/assess-repository', data).then(res => res.data),
  });
};

// Performance Prediction
export const usePerformancePrediction = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/performance-prediction/predict', data).then(res => res.data),
  });
};

export const useSolutionPerformancePrediction = (solutionId: number) => {
  return useQuery({
    queryKey: ['solution-performance-prediction', solutionId],
    queryFn: () => apiClient.get(`/ai/performance-prediction/solution/${solutionId}`).then(res => res.data),
    enabled: !!solutionId,
  });
};

export const useMarketTrends = () => {
  return useQuery({
    queryKey: ['market-trends'],
    queryFn: () => apiClient.get('/ai/performance-prediction/market-trends').then(res => res.data),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

// Governance Analysis
export const useProposalAnalysis = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/governance/analyze-proposal', data).then(res => res.data),
  });
};

export const useProposalAnalysisById = (proposalId: number) => {
  return useQuery({
    queryKey: ['proposal-analysis', proposalId],
    queryFn: () => apiClient.get(`/ai/governance/proposal/${proposalId}`).then(res => res.data),
    enabled: !!proposalId,
  });
};

export const useCommunitySentiment = (proposalId: number) => {
  return useQuery({
    queryKey: ['community-sentiment', proposalId],
    queryFn: () => apiClient.get(`/ai/governance/sentiment/${proposalId}`).then(res => res.data),
    enabled: !!proposalId,
  });
};

// Recommendations
export const usePersonalizedRecommendations = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/recommendations/get', data).then(res => res.data),
  });
};

export const useUserBountyRecommendations = (userAddress: string) => {
  return useQuery({
    queryKey: ['user-bounty-recommendations', userAddress],
    queryFn: () => apiClient.get(`/ai/recommendations/bounties/${userAddress}`).then(res => res.data),
    enabled: !!userAddress,
  });
};

export const useUserSolutionRecommendations = (userAddress: string) => {
  return useQuery({
    queryKey: ['user-solution-recommendations', userAddress],
    queryFn: () => apiClient.get(`/ai/recommendations/solutions/${userAddress}`).then(res => res.data),
    enabled: !!userAddress,
  });
};

// Optimization
export const useOptimizationAnalysis = () => {
  return useMutation({
    mutationFn: (data: any) => apiClient.post('/ai/optimization/analyze', data).then(res => res.data),
  });
};

export const usePerformanceOptimization = () => {
  return useQuery({
    queryKey: ['performance-optimization'],
    queryFn: () => apiClient.get('/ai/optimization/performance').then(res => res.data),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
};

export const useCostOptimization = () => {
  return useQuery({
    queryKey: ['cost-optimization'],
    queryFn: () => apiClient.get('/ai/optimization/cost').then(res => res.data),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
};

export const useUXOptimization = () => {
  return useQuery({
    queryKey: ['ux-optimization'],
    queryFn: () => apiClient.get('/ai/optimization/user-experience').then(res => res.data),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
};
