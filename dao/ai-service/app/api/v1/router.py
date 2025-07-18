"""
API v1 Router for AI Service
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks, Depends
from typing import Dict, Any

from app.api.v1.endpoints import (
    bounty_matching,
    quality_assessment,
    performance_prediction,
    governance_analysis,
    recommendations,
    optimization,
)

# Create main API router
api_router = APIRouter()

# Include endpoint routers
api_router.include_router(
    bounty_matching.router,
    prefix="/bounty-matching",
    tags=["Bounty Matching"],
)

api_router.include_router(
    quality_assessment.router,
    prefix="/quality-assessment",
    tags=["Quality Assessment"],
)

api_router.include_router(
    performance_prediction.router,
    prefix="/performance-prediction",
    tags=["Performance Prediction"],
)

api_router.include_router(
    governance_analysis.router,
    prefix="/governance",
    tags=["Governance Analysis"],
)

api_router.include_router(
    recommendations.router,
    prefix="/recommendations",
    tags=["Recommendations"],
)

api_router.include_router(
    optimization.router,
    prefix="/optimization",
    tags=["Optimization"],
)


@api_router.get("/")
async def api_root():
    """API root endpoint"""
    return {
        "message": "Developer DAO AI Service API v1",
        "endpoints": {
            "bounty_matching": "/api/v1/bounty-matching",
            "quality_assessment": "/api/v1/quality-assessment",
            "performance_prediction": "/api/v1/performance-prediction",
            "governance": "/api/v1/governance",
            "recommendations": "/api/v1/recommendations",
            "optimization": "/api/v1/optimization",
        },
    }
