// API Types for Developer DAO Platform

export interface APIResponse<T> {
  status: 'success' | 'error';
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, any>;
  };
  timestamp: string;
}

export interface PaginatedResponse<T> {
  status: 'success';
  data: T[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
  timestamp: string;
}

// Bounty Types
export interface Bounty {
  id: number;
  title: string;
  description: string;
  category: BountyCategory;
  status: BountyStatus;
  reward_amount: string;
  currency: string;
  deadline: string;
  created_at: string;
  updated_at: string;
  creator_address: string;
  assigned_developer?: string;
  milestones: Milestone[];
  required_skills: string[];
  tags: string[];
}

export enum BountyCategory {
  TVL_GROWTH = 0,
  MAU_EXPANSION = 1,
  INNOVATION = 2,
  SECURITY = 3,
  INFRASTRUCTURE = 4,
  COMMUNITY = 5
}

export enum BountyStatus {
  OPEN = 0,
  ASSIGNED = 1,
  IN_PROGRESS = 2,
  UNDER_REVIEW = 3,
  COMPLETED = 4,
  CANCELLED = 5
}

export interface Milestone {
  id: number;
  title: string;
  description: string;
  reward_percentage: string;
  deadline: string;
  status: MilestoneStatus;
  deliverables: string[];
}

export enum MilestoneStatus {
  PENDING = 0,
  IN_PROGRESS = 1,
  SUBMITTED = 2,
  APPROVED = 3,
  REJECTED = 4
}

export interface BountyApplication {
  id: number;
  bounty_id: number;
  applicant_address: string;
  message: string;
  proposed_timeline: number;
  status: ApplicationStatus;
  applied_at: string;
}

export enum ApplicationStatus {
  PENDING = 0,
  APPROVED = 1,
  REJECTED = 2
}

// Solution Types
export interface Solution {
  id: number;
  name: string;
  description: string;
  category: SolutionCategory;
  status: SolutionStatus;
  version: string;
  developer_address: string;
  repository_url: string;
  documentation_url?: string;
  demo_url?: string;
  tags: string[];
  created_at: string;
  updated_at: string;
  quality_score: QualityScore;
  reviews: Review[];
  installations: number;
}

export enum SolutionCategory {
  DEFI = 0,
  NFT = 1,
  DAO = 2,
  ANALYTICS = 3,
  INFRASTRUCTURE = 4,
  SECURITY = 5,
  UI_COMPONENTS = 6,
  INTEGRATION = 7
}

export enum SolutionStatus {
  DRAFT = 0,
  SUBMITTED = 1,
  UNDER_REVIEW = 2,
  APPROVED = 3,
  REJECTED = 4,
  DEPRECATED = 5
}

export interface QualityScore {
  overall_score: number;
  security_score: number;
  performance_score: number;
  usability_score: number;
  documentation_score: number;
  last_calculated: string;
}

export interface Review {
  id: number;
  solution_id: number;
  reviewer_address: string;
  rating: number;
  comment: string;
  security_score: number;
  performance_score: number;
  usability_score: number;
  documentation_score: number;
  created_at: string;
}

// Metrics Types
export interface TVLMetrics {
  current_tvl: string;
  growth_24h: string;
  growth_7d: string;
  growth_30d: string;
  timestamp: string;
  breakdown: Record<string, string>;
}

export interface MAUMetrics {
  current_mau: number;
  growth_30d: string;
  growth_90d: string;
  retention: string;
  timestamp: string;
  breakdown: Record<string, number>;
}

export interface PerformanceDashboard {
  tvl_metrics: TVLMetrics;
  mau_metrics: MAUMetrics;
  top_contributors: ImpactLeaderboard[];
  recent_alerts: Alert[];
  trending_protocols: ProtocolTrend[];
  last_updated: string;
}

export interface ImpactLeaderboard {
  entity_id: string;
  entity_type: string;
  name: string;
  tvl_impact: string;
  mau_impact: number;
  total_score: string;
  rank: number;
}

export interface Alert {
  id: number;
  name: string;
  type: AlertType;
  metric_type: MetricType;
  status: AlertStatus;
  message: string;
  created_at: string;
  triggered_at?: string;
  resolved_at?: string;
}

export enum AlertType {
  THRESHOLD = 0,
  GROWTH_RATE = 1,
  ANOMALY = 2,
  DOWNTIME = 3
}

export enum MetricType {
  TVL = 0,
  MAU = 1,
  REVENUE = 2,
  TRANSACTIONS = 3,
  USERS = 4
}

export enum AlertStatus {
  ACTIVE = 0,
  RESOLVED = 1,
  SUPPRESSED = 2
}

export interface ProtocolTrend {
  protocol: string;
  chain: string;
  current_tvl: string;
  growth_24h: string;
  growth_7d: string;
  trend_score: string;
}

// User Types
export interface User {
  address: string;
  username?: string;
  email?: string;
  avatar_url?: string;
  reputation_score: number;
  total_earnings: string;
  active_bounties: number;
  completed_bounties: number;
  solutions_count: number;
  joined_at: string;
  last_active: string;
}

export interface DeveloperProfile {
  address: string;
  username: string;
  bio?: string;
  skills: string[];
  portfolio_url?: string;
  github_url?: string;
  twitter_url?: string;
  reputation: ReputationScore;
  performance: PerformanceMetrics;
  earnings: EarningsHistory;
}

export interface ReputationScore {
  overall_score: number;
  technical_skill: number;
  communication: number;
  reliability: number;
  innovation: number;
  community_contribution: number;
  last_updated: string;
}

export interface PerformanceMetrics {
  tvl_impact: string;
  mau_impact: number;
  bounties_completed: number;
  solutions_created: number;
  average_rating: number;
  completion_rate: string;
}

export interface EarningsHistory {
  total_earned: string;
  pending_payments: string;
  last_payment: string;
  payment_history: Payment[];
}

export interface Payment {
  id: number;
  amount: string;
  currency: string;
  type: 'bounty' | 'milestone' | 'bonus' | 'revenue_share';
  status: 'pending' | 'completed' | 'failed';
  transaction_hash?: string;
  created_at: string;
  processed_at?: string;
}

// Filter and Search Types
export interface BountyFilters {
  category?: BountyCategory;
  status?: BountyStatus;
  min_reward?: number;
  max_reward?: number;
  deadline_before?: string;
  deadline_after?: string;
  skills?: string[];
  search?: string;
  limit?: number;
  offset?: number;
}

export interface SolutionFilters {
  category?: SolutionCategory;
  status?: SolutionStatus;
  min_rating?: number;
  developer?: string;
  tags?: string[];
  search?: string;
  limit?: number;
  offset?: number;
}

// Request Types
export interface CreateBountyRequest {
  title: string;
  description: string;
  category: BountyCategory;
  reward_amount: string;
  currency: string;
  deadline: string;
  milestones: Omit<Milestone, 'id' | 'status'>[];
  required_skills: string[];
  tags: string[];
}

export interface CreateSolutionRequest {
  name: string;
  description: string;
  category: SolutionCategory;
  version: string;
  repository_url: string;
  documentation_url?: string;
  demo_url?: string;
  tags: string[];
}

export interface CreateReviewRequest {
  solution_id: number;
  rating: number;
  comment: string;
  security_score: number;
  performance_score: number;
  usability_score: number;
  documentation_score: number;
}

export interface ApplyForBountyRequest {
  bounty_id: number;
  message: string;
  proposed_timeline: number;
}
