"""
Bounty Matching API Endpoints
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks, Depends
from typing import List, Optional

from app.agents.base_agent import agent_manager
from app.models.schemas import (
    BountyMatchRequest,
    BountyMatchResponse,
    BaseResponse,
    ErrorResponse,
)
from app.core.logging import AIServiceLogger

router = APIRouter()
logger = AIServiceLogger("api.bounty_matching")


@router.post("/match", response_model=BountyMatchResponse)
async def match_bounty_developer(
    request: BountyMatchRequest,
    background_tasks: BackgroundTasks,
) -> BountyMatchResponse:
    """
    Match developers with bounties using AI-powered analysis
    
    This endpoint provides intelligent matching between developers and bounties
    based on skill compatibility, experience level, availability, and success probability.
    """
    try:
        logger.log_request("match", "POST", request_data=request.dict())
        
        # Get bounty matching agent
        agent = await agent_manager.get_agent("bounty_matching")
        if not agent:
            raise HTTPException(status_code=503, detail="Bounty matching agent not available")
        
        # Process the matching request
        result = await agent.process_request(request.dict())
        
        # Create response
        response = BountyMatchResponse(
            matches=result["matches"],
            total_matches=result["total_matches"],
            status="success",
            message="Bounty matching completed successfully",
        )
        
        logger.info("Bounty matching completed", 
                   total_matches=result["total_matches"],
                   bounty_id=request.bounty_id,
                   developer_address=request.developer_address)
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "match",
            "request": request.dict(),
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/developers/{bounty_id}", response_model=BountyMatchResponse)
async def find_developers_for_bounty(
    bounty_id: int,
    limit: int = 10,
    min_compatibility_score: float = 0.5,
) -> BountyMatchResponse:
    """
    Find the best developers for a specific bounty
    
    Returns a ranked list of developers who are most compatible with the bounty
    requirements, including skill match, experience level, and availability.
    """
    try:
        logger.log_request("find_developers", "GET", bounty_id=bounty_id)
        
        # Create request
        request = BountyMatchRequest(
            bounty_id=bounty_id,
            limit=limit,
            min_compatibility_score=min_compatibility_score,
        )
        
        # Get bounty matching agent
        agent = await agent_manager.get_agent("bounty_matching")
        if not agent:
            raise HTTPException(status_code=503, detail="Bounty matching agent not available")
        
        # Process the request
        result = await agent.process_request(request.dict())
        
        # Create response
        response = BountyMatchResponse(
            matches=result["matches"],
            total_matches=result["total_matches"],
            status="success",
            message=f"Found {result['total_matches']} compatible developers",
        )
        
        logger.info("Developer search completed",
                   bounty_id=bounty_id,
                   matches_found=result["total_matches"])
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "find_developers",
            "bounty_id": bounty_id,
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/bounties/{developer_address}", response_model=BountyMatchResponse)
async def find_bounties_for_developer(
    developer_address: str,
    limit: int = 10,
    min_compatibility_score: float = 0.5,
) -> BountyMatchResponse:
    """
    Find the best bounties for a specific developer
    
    Returns a ranked list of bounties that match the developer's skills,
    experience level, and availability preferences.
    """
    try:
        logger.log_request("find_bounties", "GET", developer_address=developer_address)
        
        # Create request
        request = BountyMatchRequest(
            developer_address=developer_address,
            limit=limit,
            min_compatibility_score=min_compatibility_score,
        )
        
        # Get bounty matching agent
        agent = await agent_manager.get_agent("bounty_matching")
        if not agent:
            raise HTTPException(status_code=503, detail="Bounty matching agent not available")
        
        # Process the request
        result = await agent.process_request(request.dict())
        
        # Create response
        response = BountyMatchResponse(
            matches=result["matches"],
            total_matches=result["total_matches"],
            status="success",
            message=f"Found {result['total_matches']} compatible bounties",
        )
        
        logger.info("Bounty search completed",
                   developer_address=developer_address,
                   matches_found=result["total_matches"])
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "find_bounties",
            "developer_address": developer_address,
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.post("/analyze-compatibility")
async def analyze_compatibility(
    bounty_id: int,
    developer_address: str,
) -> dict:
    """
    Analyze compatibility between a specific bounty and developer
    
    Provides detailed analysis of how well a developer matches a bounty,
    including breakdown of skill match, experience compatibility, and success probability.
    """
    try:
        logger.log_request("analyze_compatibility", "POST", 
                          bounty_id=bounty_id, 
                          developer_address=developer_address)
        
        # Create request
        request = BountyMatchRequest(
            bounty_id=bounty_id,
            developer_address=developer_address,
        )
        
        # Get bounty matching agent
        agent = await agent_manager.get_agent("bounty_matching")
        if not agent:
            raise HTTPException(status_code=503, detail="Bounty matching agent not available")
        
        # Process the request
        result = await agent.process_request(request.dict())
        
        if not result["matches"]:
            raise HTTPException(status_code=404, detail="No compatibility analysis available")
        
        match = result["matches"][0]
        
        # Return detailed compatibility analysis
        analysis = {
            "compatibility_score": match["compatibility_score"],
            "confidence_level": match["confidence_level"],
            "skill_analysis": {
                "score": match["skill_match_score"],
                "description": "Skill compatibility analysis"
            },
            "experience_analysis": {
                "score": match["experience_match_score"],
                "description": "Experience level compatibility"
            },
            "availability_analysis": {
                "score": match["availability_match_score"],
                "description": "Timeline and availability compatibility"
            },
            "success_probability": match["success_probability"],
            "recommended_timeline": match["recommended_timeline"],
            "explanation": match["explanation"],
            "recommendation": "Proceed" if match["compatibility_score"] >= 0.7 else "Consider alternatives",
        }
        
        logger.info("Compatibility analysis completed",
                   bounty_id=bounty_id,
                   developer_address=developer_address,
                   compatibility_score=match["compatibility_score"])
        
        return analysis
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "analyze_compatibility",
            "bounty_id": bounty_id,
            "developer_address": developer_address,
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/agent-status")
async def get_agent_status():
    """Get bounty matching agent status and metrics"""
    try:
        agent = await agent_manager.get_agent("bounty_matching")
        if not agent:
            return {"status": "not_available", "message": "Agent not found"}
        
        status = await agent.get_health_status()
        return status
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "agent_status"})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")
