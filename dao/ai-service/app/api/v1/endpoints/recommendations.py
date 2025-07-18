"""
Recommendations API Endpoints
"""
from fastapi import APIRouter, HTTPException

from app.models.schemas import (
    RecommendationRequest,
    RecommendationResponse,
)
from app.core.logging import AIServiceLogger

router = APIRouter()
logger = AIServiceLogger("api.recommendations")


@router.post("/get", response_model=RecommendationResponse)
async def get_recommendations(request: RecommendationRequest) -> RecommendationResponse:
    """Get personalized recommendations for users"""
    try:
        logger.log_request("get_recommendations", "POST", request_data=request.dict())
        
        # Mock recommendations
        recommendations = [
            {
                "item_id": "bounty_123",
                "item_type": "bounty",
                "title": "DeFi Analytics Dashboard",
                "description": "Build comprehensive analytics for DeFi protocols",
                "relevance_score": 0.92,
                "confidence_level": "high",
                "explanation": "Matches your React and DeFi expertise perfectly",
                "metadata": {"reward": 5000, "deadline": "2024-02-15"},
            },
            {
                "item_id": "solution_456",
                "item_type": "solution",
                "title": "Yield Optimizer Library",
                "description": "Smart contract library for yield optimization",
                "relevance_score": 0.87,
                "confidence_level": "high",
                "explanation": "Relevant to your smart contract development skills",
                "metadata": {"rating": 4.8, "downloads": 1250},
            },
        ]
        
        response = RecommendationResponse(
            recommendations=recommendations,
            total_available=len(recommendations),
            status="success",
            message="Recommendations generated successfully",
        )
        
        logger.info("Recommendations generated",
                   user_address=request.user_address,
                   recommendation_type=request.recommendation_type,
                   count=len(recommendations))
        
        return response
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "get_recommendations", "request": request.dict()})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/bounties/{user_address}")
async def get_bounty_recommendations(
    user_address: str,
    limit: int = 10,
) -> dict:
    """Get bounty recommendations for a user"""
    try:
        recommendations = [
            {
                "bounty_id": 123,
                "title": "DeFi Analytics Dashboard",
                "relevance_score": 0.92,
                "match_reasons": ["React expertise", "DeFi experience"],
            },
            {
                "bounty_id": 124,
                "title": "NFT Marketplace Integration",
                "relevance_score": 0.85,
                "match_reasons": ["Web3 skills", "Frontend development"],
            },
        ]
        
        return {"recommendations": recommendations[:limit]}
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "bounty_recommendations", "user_address": user_address})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/solutions/{user_address}")
async def get_solution_recommendations(
    user_address: str,
    limit: int = 10,
) -> dict:
    """Get solution recommendations for a user"""
    try:
        recommendations = [
            {
                "solution_id": 456,
                "title": "Yield Optimizer Library",
                "relevance_score": 0.87,
                "why_recommended": "Matches your DeFi and smart contract interests",
            },
        ]
        
        return {"recommendations": recommendations[:limit]}
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "solution_recommendations", "user_address": user_address})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")
