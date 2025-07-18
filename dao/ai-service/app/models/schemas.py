"""
Pydantic schemas for AI Service
"""
from typing import List, Optional, Dict, Any, Union
from datetime import datetime
from enum import Enum
from pydantic import BaseModel, Field, validator


class AITaskStatus(str, Enum):
    """AI task status enumeration"""
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"


class ConfidenceLevel(str, Enum):
    """Confidence level enumeration"""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    VERY_HIGH = "very_high"


# Base Schemas
class BaseResponse(BaseModel):
    """Base response schema"""
    status: str = "success"
    message: Optional[str] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class ErrorResponse(BaseModel):
    """Error response schema"""
    status: str = "error"
    error_code: str
    error_message: str
    details: Optional[Dict[str, Any]] = None
    timestamp: datetime = Field(default_factory=datetime.utcnow)


# Bounty Matching Schemas
class DeveloperProfile(BaseModel):
    """Developer profile for matching"""
    address: str
    skills: List[str]
    experience_level: str
    reputation_score: float
    completed_bounties: int
    success_rate: float
    preferred_categories: List[str]
    availability_hours: Optional[int] = None


class BountyRequirements(BaseModel):
    """Bounty requirements for matching"""
    id: int
    title: str
    description: str
    category: str
    required_skills: List[str]
    difficulty_level: str
    estimated_hours: Optional[int] = None
    reward_amount: float
    deadline: datetime


class BountyMatch(BaseModel):
    """Bounty-developer match result"""
    developer_address: str
    bounty_id: int
    compatibility_score: float = Field(ge=0.0, le=1.0)
    confidence_level: ConfidenceLevel
    skill_match_score: float = Field(ge=0.0, le=1.0)
    experience_match_score: float = Field(ge=0.0, le=1.0)
    availability_match_score: float = Field(ge=0.0, le=1.0)
    success_probability: float = Field(ge=0.0, le=1.0)
    explanation: str
    recommended_timeline: Optional[int] = None


class BountyMatchRequest(BaseModel):
    """Request for bounty matching"""
    bounty_id: Optional[int] = None
    developer_address: Optional[str] = None
    limit: int = Field(default=10, ge=1, le=50)
    min_compatibility_score: float = Field(default=0.5, ge=0.0, le=1.0)


class BountyMatchResponse(BaseResponse):
    """Response for bounty matching"""
    matches: List[BountyMatch]
    total_matches: int


# Quality Assessment Schemas
class CodeQualityMetrics(BaseModel):
    """Code quality metrics"""
    overall_score: float = Field(ge=0.0, le=100.0)
    security_score: float = Field(ge=0.0, le=100.0)
    performance_score: float = Field(ge=0.0, le=100.0)
    maintainability_score: float = Field(ge=0.0, le=100.0)
    documentation_score: float = Field(ge=0.0, le=100.0)
    test_coverage_score: float = Field(ge=0.0, le=100.0)


class QualityIssue(BaseModel):
    """Quality issue found in code"""
    type: str  # security, performance, style, etc.
    severity: str  # low, medium, high, critical
    description: str
    file_path: Optional[str] = None
    line_number: Optional[int] = None
    suggestion: Optional[str] = None


class QualityAssessmentRequest(BaseModel):
    """Request for quality assessment"""
    solution_id: Optional[int] = None
    repository_url: Optional[str] = None
    code_content: Optional[str] = None
    language: Optional[str] = None
    include_suggestions: bool = True


class QualityAssessmentResponse(BaseResponse):
    """Response for quality assessment"""
    metrics: CodeQualityMetrics
    issues: List[QualityIssue]
    suggestions: List[str]
    confidence_level: ConfidenceLevel
    analysis_summary: str


# Performance Prediction Schemas
class PerformancePredictionRequest(BaseModel):
    """Request for performance prediction"""
    solution_id: Optional[int] = None
    bounty_id: Optional[int] = None
    prediction_horizon_days: int = Field(default=30, ge=1, le=365)
    include_market_analysis: bool = True


class MarketTrend(BaseModel):
    """Market trend data"""
    metric: str
    current_value: float
    predicted_value: float
    confidence: float
    trend_direction: str  # up, down, stable


class PerformancePrediction(BaseModel):
    """Performance prediction result"""
    tvl_impact: Dict[str, float]  # predicted TVL changes
    mau_impact: Dict[str, int]    # predicted MAU changes
    adoption_probability: float = Field(ge=0.0, le=1.0)
    risk_score: float = Field(ge=0.0, le=1.0)
    roi_estimate: float
    market_trends: List[MarketTrend]
    confidence_level: ConfidenceLevel
    explanation: str


class PerformancePredictionResponse(BaseResponse):
    """Response for performance prediction"""
    prediction: PerformancePrediction
    historical_accuracy: Optional[float] = None


# Governance Analysis Schemas
class ProposalAnalysisRequest(BaseModel):
    """Request for proposal analysis"""
    proposal_id: Optional[int] = None
    proposal_text: str
    proposal_type: str
    include_sentiment_analysis: bool = True


class SentimentAnalysis(BaseModel):
    """Sentiment analysis result"""
    overall_sentiment: str  # positive, negative, neutral
    sentiment_score: float = Field(ge=-1.0, le=1.0)
    key_topics: List[str]
    community_concerns: List[str]
    support_indicators: List[str]


class ProposalAnalysis(BaseModel):
    """Proposal analysis result"""
    summary: str
    key_points: List[str]
    potential_impact: str
    implementation_complexity: str
    resource_requirements: List[str]
    risks: List[str]
    benefits: List[str]
    voting_recommendation: str
    confidence_level: ConfidenceLevel
    sentiment_analysis: Optional[SentimentAnalysis] = None


class ProposalAnalysisResponse(BaseResponse):
    """Response for proposal analysis"""
    analysis: ProposalAnalysis


# Recommendation Schemas
class RecommendationRequest(BaseModel):
    """Request for recommendations"""
    user_address: str
    recommendation_type: str  # bounties, solutions, developers
    limit: int = Field(default=10, ge=1, le=50)
    include_explanations: bool = True


class Recommendation(BaseModel):
    """Individual recommendation"""
    item_id: str
    item_type: str
    title: str
    description: str
    relevance_score: float = Field(ge=0.0, le=1.0)
    confidence_level: ConfidenceLevel
    explanation: str
    metadata: Dict[str, Any] = {}


class RecommendationResponse(BaseResponse):
    """Response for recommendations"""
    recommendations: List[Recommendation]
    total_available: int


# Optimization Schemas
class OptimizationRequest(BaseModel):
    """Request for system optimization"""
    optimization_type: str  # performance, cost, user_experience
    target_metrics: List[str]
    constraints: Dict[str, Any] = {}


class OptimizationSuggestion(BaseModel):
    """Optimization suggestion"""
    category: str
    description: str
    expected_improvement: str
    implementation_effort: str  # low, medium, high
    priority: str  # low, medium, high, critical
    estimated_impact: float = Field(ge=0.0, le=1.0)


class OptimizationResponse(BaseResponse):
    """Response for optimization"""
    suggestions: List[OptimizationSuggestion]
    current_metrics: Dict[str, float]
    projected_improvements: Dict[str, float]


# Health Check Schema
class HealthCheckResponse(BaseModel):
    """Health check response"""
    status: str
    version: str
    timestamp: datetime
    services: Dict[str, bool]
    models_loaded: Dict[str, bool]
    uptime_seconds: float
