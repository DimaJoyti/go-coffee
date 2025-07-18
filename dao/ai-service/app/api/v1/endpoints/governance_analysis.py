"""
Governance Analysis API Endpoints
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks

from app.models.schemas import (
    ProposalAnalysisRequest,
    ProposalAnalysisResponse,
)
from app.core.logging import AIServiceLogger

router = APIRouter()
logger = AIServiceLogger("api.governance_analysis")


@router.post("/analyze-proposal", response_model=ProposalAnalysisResponse)
async def analyze_proposal(
    request: ProposalAnalysisRequest,
    background_tasks: BackgroundTasks,
) -> ProposalAnalysisResponse:
    """
    Analyze governance proposals using AI
    """
    try:
        logger.log_request("analyze_proposal", "POST", request_data=request.dict())
        
        # Mock analysis - would use actual NLP models
        analysis = {
            "summary": "This proposal aims to increase developer rewards by 25% to improve platform competitiveness and attract high-quality contributors.",
            "key_points": [
                "Increase monthly reward pool from $100K to $125K",
                "Expected to attract 30% more developers",
                "Funding from treasury reserves",
                "Implementation timeline: 30 days",
            ],
            "potential_impact": "Positive impact on developer acquisition and retention, moderate treasury impact",
            "implementation_complexity": "Low - requires smart contract parameter update",
            "resource_requirements": [
                "Additional $25K monthly budget",
                "Smart contract upgrade",
                "Community communication",
            ],
            "risks": [
                "Treasury depletion if not sustainable",
                "Potential inflation of reward expectations",
            ],
            "benefits": [
                "Increased developer participation",
                "Higher quality solutions",
                "Platform growth acceleration",
            ],
            "voting_recommendation": "Support with monitoring provisions",
            "confidence_level": "high",
            "sentiment_analysis": {
                "overall_sentiment": "positive",
                "sentiment_score": 0.72,
                "key_topics": ["rewards", "developers", "growth"],
                "community_concerns": ["sustainability", "treasury management"],
                "support_indicators": ["developer feedback", "growth metrics"],
            } if request.include_sentiment_analysis else None,
        }
        
        response = ProposalAnalysisResponse(
            analysis=analysis,
            status="success",
            message="Proposal analysis completed successfully",
        )
        
        logger.info("Proposal analysis completed",
                   proposal_id=request.proposal_id,
                   proposal_type=request.proposal_type)
        
        return response
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "analyze_proposal", "request": request.dict()})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/proposal/{proposal_id}")
async def analyze_proposal_by_id(
    proposal_id: int,
    include_sentiment: bool = True,
) -> dict:
    """Analyze a specific proposal by ID"""
    try:
        logger.log_request("analyze_proposal_by_id", "GET", proposal_id=proposal_id)
        
        # Mock analysis
        analysis = {
            "proposal_id": proposal_id,
            "summary": "Proposal analysis summary",
            "recommendation": "support",
            "confidence": 0.85,
            "key_concerns": [],
            "benefits": [],
        }
        
        return analysis
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "analyze_proposal_by_id", "proposal_id": proposal_id})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/sentiment/{proposal_id}")
async def analyze_community_sentiment(proposal_id: int) -> dict:
    """Analyze community sentiment for a proposal"""
    try:
        sentiment = {
            "proposal_id": proposal_id,
            "overall_sentiment": "positive",
            "sentiment_score": 0.72,
            "community_engagement": "high",
            "key_topics": ["rewards", "sustainability"],
            "concerns": ["treasury impact"],
            "support_level": 0.78,
        }
        
        return sentiment
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "analyze_sentiment", "proposal_id": proposal_id})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")
