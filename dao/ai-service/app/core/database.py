"""
Database configuration and connection management
"""
import asyncio
from typing import AsyncGenerator
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import DeclarativeBase
import redis.asyncio as redis
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams

from app.core.config import get_settings

settings = get_settings()


class Base(DeclarativeBase):
    """Base class for SQLAlchemy models"""
    pass


# Database Engine
engine = create_async_engine(
    settings.database_url,
    echo=settings.debug,
    pool_pre_ping=True,
    pool_recycle=300,
)

# Session Factory
AsyncSessionLocal = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    """Get database session"""
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()


# Redis Connection
redis_client = None


async def get_redis() -> redis.Redis:
    """Get Redis client"""
    global redis_client
    if redis_client is None:
        redis_client = redis.from_url(
            settings.redis_url,
            encoding="utf-8",
            decode_responses=True,
        )
    return redis_client


async def close_redis():
    """Close Redis connection"""
    global redis_client
    if redis_client:
        await redis_client.close()


# Qdrant Vector Database
qdrant_client = None


def get_qdrant() -> QdrantClient:
    """Get Qdrant client"""
    global qdrant_client
    if qdrant_client is None:
        qdrant_client = QdrantClient(
            host=settings.qdrant_host,
            port=settings.qdrant_port,
            api_key=settings.qdrant_api_key,
        )
    return qdrant_client


async def init_vector_collections():
    """Initialize Qdrant collections"""
    client = get_qdrant()
    
    collections = [
        {
            "name": "developer_profiles",
            "vector_size": 384,  # all-MiniLM-L6-v2 embedding size
            "distance": Distance.COSINE,
        },
        {
            "name": "bounty_requirements",
            "vector_size": 384,
            "distance": Distance.COSINE,
        },
        {
            "name": "solution_descriptions",
            "vector_size": 384,
            "distance": Distance.COSINE,
        },
        {
            "name": "code_embeddings",
            "vector_size": 768,  # CodeBERT embedding size
            "distance": Distance.COSINE,
        },
        {
            "name": "proposal_content",
            "vector_size": 384,
            "distance": Distance.COSINE,
        },
    ]
    
    for collection in collections:
        try:
            # Check if collection exists
            client.get_collection(collection["name"])
        except Exception:
            # Create collection if it doesn't exist
            client.create_collection(
                collection_name=collection["name"],
                vectors_config=VectorParams(
                    size=collection["vector_size"],
                    distance=collection["distance"],
                ),
            )


# Database initialization
async def init_database():
    """Initialize database tables"""
    async with engine.begin() as conn:
        # Create all tables
        await conn.run_sync(Base.metadata.create_all)


# Cleanup
async def close_database():
    """Close database connections"""
    await engine.dispose()
    await close_redis()


# Health check
async def check_database_health() -> dict:
    """Check database connectivity"""
    health = {
        "postgres": False,
        "redis": False,
        "qdrant": False,
    }
    
    # Check PostgreSQL
    try:
        async with AsyncSessionLocal() as session:
            await session.execute("SELECT 1")
            health["postgres"] = True
    except Exception:
        pass
    
    # Check Redis
    try:
        redis_conn = await get_redis()
        await redis_conn.ping()
        health["redis"] = True
    except Exception:
        pass
    
    # Check Qdrant
    try:
        client = get_qdrant()
        client.get_collections()
        health["qdrant"] = True
    except Exception:
        pass
    
    return health
