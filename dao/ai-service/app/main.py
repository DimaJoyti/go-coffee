"""
AI Service Main Application
"""
import asyncio
import time
from contextlib import asynccontextmanager
from typing import Dict, Any

from fastapi import FastAPI, HTTPException, Depends, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.responses import JSONResponse
import uvicorn

from app.core.config import get_settings
from app.core.database import init_database, init_vector_collections, close_database, check_database_health
from app.core.logging import configure_logging, get_logger
from app.agents.base_agent import agent_manager
from app.agents.bounty_matching_agent import BountyMatchingAgent
from app.agents.quality_assessment_agent import QualityAssessmentAgent
from app.api.v1.router import api_router
from app.models.schemas import HealthCheckResponse

# Configure logging
configure_logging()
logger = get_logger("main")

settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    # Startup
    logger.info("Starting AI Service", version=settings.app_version)
    
    try:
        # Initialize database
        await init_database()
        await init_vector_collections()
        logger.info("Database initialized")
        
        # Register and initialize agents
        agent_manager.register_agent(BountyMatchingAgent())
        agent_manager.register_agent(QualityAssessmentAgent())
        
        await agent_manager.initialize_all_agents()
        logger.info("AI agents initialized", agents=agent_manager.list_agents())
        
        logger.info("AI Service startup complete")
        
    except Exception as e:
        logger.error("Failed to start AI Service", error=str(e))
        raise
    
    yield
    
    # Shutdown
    logger.info("Shutting down AI Service")
    try:
        await close_database()
        logger.info("AI Service shutdown complete")
    except Exception as e:
        logger.error("Error during shutdown", error=str(e))


# Create FastAPI application
app = FastAPI(
    title=settings.app_name,
    version=settings.app_version,
    description="AI-powered intelligent services for Developer DAO Platform",
    docs_url="/docs" if settings.debug else None,
    redoc_url="/redoc" if settings.debug else None,
    lifespan=lifespan,
)

# Add middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.allowed_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.add_middleware(
    TrustedHostMiddleware,
    allowed_hosts=["*"] if settings.debug else ["localhost", "127.0.0.1"],
)


# Request timing middleware
@app.middleware("http")
async def add_process_time_header(request, call_next):
    """Add processing time header"""
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    response.headers["X-Process-Time"] = str(process_time)
    return response


# Exception handlers
@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc):
    """Handle HTTP exceptions"""
    logger.error(
        "HTTP exception",
        status_code=exc.status_code,
        detail=exc.detail,
        path=request.url.path,
    )
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "status": "error",
            "error_code": f"HTTP_{exc.status_code}",
            "error_message": exc.detail,
            "timestamp": time.time(),
        },
    )


@app.exception_handler(Exception)
async def general_exception_handler(request, exc):
    """Handle general exceptions"""
    logger.error(
        "Unhandled exception",
        error=str(exc),
        error_type=type(exc).__name__,
        path=request.url.path,
    )
    return JSONResponse(
        status_code=500,
        content={
            "status": "error",
            "error_code": "INTERNAL_SERVER_ERROR",
            "error_message": "An internal server error occurred",
            "timestamp": time.time(),
        },
    )


# Health check endpoint
@app.get("/health", response_model=HealthCheckResponse)
async def health_check():
    """Health check endpoint"""
    start_time = time.time()
    
    # Check database health
    db_health = await check_database_health()
    
    # Check agent health
    agent_health = await agent_manager.get_all_health_status()
    
    # Determine overall status
    all_services_healthy = all(db_health.values())
    all_agents_healthy = all(
        agent.get("status") == "healthy" 
        for agent in agent_health.values()
    )
    
    overall_status = "healthy" if all_services_healthy and all_agents_healthy else "unhealthy"
    
    return HealthCheckResponse(
        status=overall_status,
        version=settings.app_version,
        timestamp=time.time(),
        services=db_health,
        models_loaded={
            agent_name: agent.get("is_initialized", False)
            for agent_name, agent in agent_health.items()
        },
        uptime_seconds=time.time() - start_time,
    )


# Metrics endpoint
@app.get("/metrics")
async def get_metrics():
    """Get service metrics"""
    agent_health = await agent_manager.get_all_health_status()
    
    metrics = {
        "service": {
            "name": settings.app_name,
            "version": settings.app_version,
            "uptime": time.time(),
        },
        "agents": {
            name: {
                "status": agent.get("status"),
                "metrics": agent.get("metrics", {}),
            }
            for name, agent in agent_health.items()
        },
        "database": await check_database_health(),
    }
    
    return metrics


# Root endpoint
@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": settings.app_name,
        "version": settings.app_version,
        "status": "running",
        "docs": "/docs" if settings.debug else "disabled",
        "health": "/health",
        "metrics": "/metrics",
    }


# Include API router
app.include_router(api_router, prefix="/api/v1")


# Background task for periodic maintenance
async def periodic_maintenance():
    """Periodic maintenance tasks"""
    while True:
        try:
            # Log system status
            logger.info("Periodic maintenance check")
            
            # Check agent health
            agent_health = await agent_manager.get_all_health_status()
            unhealthy_agents = [
                name for name, status in agent_health.items()
                if status.get("status") != "healthy"
            ]
            
            if unhealthy_agents:
                logger.warning("Unhealthy agents detected", agents=unhealthy_agents)
            
            # Sleep for 5 minutes
            await asyncio.sleep(300)
            
        except Exception as e:
            logger.error("Error in periodic maintenance", error=str(e))
            await asyncio.sleep(60)  # Shorter sleep on error


# Start background tasks
@app.on_event("startup")
async def start_background_tasks():
    """Start background tasks"""
    if settings.enable_metrics:
        asyncio.create_task(periodic_maintenance())


if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        workers=settings.workers,
        log_level=settings.log_level.lower(),
        reload=settings.debug,
    )
