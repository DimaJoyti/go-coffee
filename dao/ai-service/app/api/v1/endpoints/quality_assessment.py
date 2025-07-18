"""
Quality Assessment API Endpoints
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks, UploadFile, File
from typing import Optional

from app.agents.base_agent import agent_manager
from app.models.schemas import (
    QualityAssessmentRequest,
    QualityAssessmentResponse,
    BaseResponse,
)
from app.core.logging import AIServiceLogger

router = APIRouter()
logger = AIServiceLogger("api.quality_assessment")


@router.post("/assess", response_model=QualityAssessmentResponse)
async def assess_quality(
    request: QualityAssessmentRequest,
    background_tasks: BackgroundTasks,
) -> QualityAssessmentResponse:
    """
    Assess the quality of a solution using AI-powered analysis
    
    Provides comprehensive quality assessment including:
    - Code quality and security analysis
    - Documentation completeness
    - Best practices compliance
    - Performance considerations
    """
    try:
        logger.log_request("assess", "POST", request_data=request.dict())
        
        # Get quality assessment agent
        agent = await agent_manager.get_agent("quality_assessment")
        if not agent:
            raise HTTPException(status_code=503, detail="Quality assessment agent not available")
        
        # Process the assessment request
        result = await agent.process_request(request.dict())
        
        # Create response
        response = QualityAssessmentResponse(
            metrics=result["metrics"],
            issues=result["issues"],
            suggestions=result["suggestions"],
            confidence_level=result["confidence_level"],
            analysis_summary=result["analysis_summary"],
            status="success",
            message="Quality assessment completed successfully",
        )
        
        logger.info("Quality assessment completed",
                   solution_id=request.solution_id,
                   overall_score=result["metrics"]["overall_score"],
                   issues_count=len(result["issues"]))
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "assess",
            "request": request.dict(),
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.post("/assess-code")
async def assess_code_quality(
    code_content: str,
    language: str = "python",
    include_suggestions: bool = True,
) -> dict:
    """
    Assess quality of provided code content
    
    Analyzes code directly without requiring a repository or solution ID.
    Useful for real-time code quality feedback during development.
    """
    try:
        logger.log_request("assess_code", "POST", language=language)
        
        # Create request
        request = QualityAssessmentRequest(
            code_content=code_content,
            language=language,
            include_suggestions=include_suggestions,
        )
        
        # Get quality assessment agent
        agent = await agent_manager.get_agent("quality_assessment")
        if not agent:
            raise HTTPException(status_code=503, detail="Quality assessment agent not available")
        
        # Process the request
        result = await agent.process_request(request.dict())
        
        logger.info("Code quality assessment completed",
                   language=language,
                   overall_score=result["metrics"]["overall_score"])
        
        return result
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "assess_code",
            "language": language,
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.post("/assess-repository")
async def assess_repository_quality(
    repository_url: str,
    language: str = "python",
    include_suggestions: bool = True,
) -> dict:
    """
    Assess quality of a GitHub repository
    
    Analyzes the entire repository including code quality, documentation,
    project structure, and best practices compliance.
    """
    try:
        logger.log_request("assess_repository", "POST", repository_url=repository_url)
        
        # Create request
        request = QualityAssessmentRequest(
            repository_url=repository_url,
            language=language,
            include_suggestions=include_suggestions,
        )
        
        # Get quality assessment agent
        agent = await agent_manager.get_agent("quality_assessment")
        if not agent:
            raise HTTPException(status_code=503, detail="Quality assessment agent not available")
        
        # Process the request
        result = await agent.process_request(request.dict())
        
        logger.info("Repository quality assessment completed",
                   repository_url=repository_url,
                   overall_score=result["metrics"]["overall_score"])
        
        return result
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "assess_repository",
            "repository_url": repository_url,
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/solution/{solution_id}")
async def assess_solution_quality(
    solution_id: int,
    include_suggestions: bool = True,
) -> dict:
    """
    Assess quality of a solution by ID
    
    Retrieves solution information and performs comprehensive quality assessment
    including code analysis, documentation review, and security evaluation.
    """
    try:
        logger.log_request("assess_solution", "GET", solution_id=solution_id)
        
        # Create request
        request = QualityAssessmentRequest(
            solution_id=solution_id,
            include_suggestions=include_suggestions,
        )
        
        # Get quality assessment agent
        agent = await agent_manager.get_agent("quality_assessment")
        if not agent:
            raise HTTPException(status_code=503, detail="Quality assessment agent not available")
        
        # Process the request
        result = await agent.process_request(request.dict())
        
        logger.info("Solution quality assessment completed",
                   solution_id=solution_id,
                   overall_score=result["metrics"]["overall_score"])
        
        return result
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "assess_solution",
            "solution_id": solution_id,
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.post("/upload-file")
async def assess_uploaded_file(
    file: UploadFile = File(...),
    language: Optional[str] = None,
    include_suggestions: bool = True,
) -> dict:
    """
    Assess quality of an uploaded code file
    
    Allows users to upload code files directly for quality assessment.
    Automatically detects language if not specified.
    """
    try:
        logger.log_request("assess_uploaded_file", "POST", filename=file.filename)
        
        # Read file content
        content = await file.read()
        code_content = content.decode('utf-8')
        
        # Auto-detect language if not provided
        if not language:
            if file.filename.endswith('.py'):
                language = 'python'
            elif file.filename.endswith(('.js', '.jsx')):
                language = 'javascript'
            elif file.filename.endswith(('.ts', '.tsx')):
                language = 'typescript'
            elif file.filename.endswith('.sol'):
                language = 'solidity'
            elif file.filename.endswith('.go'):
                language = 'go'
            else:
                language = 'python'  # Default
        
        # Create request
        request = QualityAssessmentRequest(
            code_content=code_content,
            language=language,
            include_suggestions=include_suggestions,
        )
        
        # Get quality assessment agent
        agent = await agent_manager.get_agent("quality_assessment")
        if not agent:
            raise HTTPException(status_code=503, detail="Quality assessment agent not available")
        
        # Process the request
        result = await agent.process_request(request.dict())
        
        # Add file information to result
        result["file_info"] = {
            "filename": file.filename,
            "size": len(code_content),
            "detected_language": language,
        }
        
        logger.info("File quality assessment completed",
                   filename=file.filename,
                   language=language,
                   overall_score=result["metrics"]["overall_score"])
        
        return result
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {
            "endpoint": "assess_uploaded_file",
            "filename": file.filename if file else "unknown",
        })
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/metrics/summary")
async def get_quality_metrics_summary():
    """
    Get summary of quality assessment metrics
    
    Returns aggregated statistics about quality assessments performed,
    common issues found, and improvement trends.
    """
    try:
        # Get agent status and metrics
        agent = await agent_manager.get_agent("quality_assessment")
        if not agent:
            raise HTTPException(status_code=503, detail="Quality assessment agent not available")
        
        status = await agent.get_health_status()
        
        # Mock aggregated metrics (would come from database in real implementation)
        summary = {
            "agent_status": status,
            "assessments_performed": status["metrics"]["total_requests"],
            "average_quality_score": 75.2,
            "common_issues": [
                {"type": "security", "count": 45, "percentage": 23.5},
                {"type": "documentation", "count": 38, "percentage": 19.8},
                {"type": "performance", "count": 32, "percentage": 16.7},
                {"type": "style", "count": 28, "percentage": 14.6},
            ],
            "language_distribution": {
                "python": 45,
                "javascript": 32,
                "typescript": 18,
                "solidity": 12,
                "other": 8,
            },
            "quality_trends": {
                "improving": 68,
                "stable": 25,
                "declining": 7,
            },
        }
        
        return summary
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "metrics_summary"})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")


@router.get("/agent-status")
async def get_agent_status():
    """Get quality assessment agent status and metrics"""
    try:
        agent = await agent_manager.get_agent("quality_assessment")
        if not agent:
            return {"status": "not_available", "message": "Agent not found"}
        
        status = await agent.get_health_status()
        return status
        
    except Exception as e:
        logger.error_with_context(e, {"endpoint": "agent_status"})
        raise HTTPException(status_code=500, detail=f"Internal server error: {str(e)}")
