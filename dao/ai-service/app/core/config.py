"""
AI Service Configuration
"""
from typing import Optional, List
from pydantic import BaseSettings, Field
import os


class Settings(BaseSettings):
    """Application settings"""
    
    # Application
    app_name: str = "Developer DAO AI Service"
    app_version: str = "1.0.0"
    debug: bool = Field(default=False, env="DEBUG")
    
    # Server
    host: str = Field(default="0.0.0.0", env="HOST")
    port: int = Field(default=8083, env="PORT")
    workers: int = Field(default=1, env="WORKERS")
    
    # Database
    database_url: str = Field(env="DATABASE_URL")
    redis_url: str = Field(env="REDIS_URL")
    
    # AI/ML Configuration
    openai_api_key: Optional[str] = Field(default=None, env="OPENAI_API_KEY")
    anthropic_api_key: Optional[str] = Field(default=None, env="ANTHROPIC_API_KEY")
    huggingface_token: Optional[str] = Field(default=None, env="HUGGINGFACE_TOKEN")
    
    # Vector Database
    qdrant_host: str = Field(default="localhost", env="QDRANT_HOST")
    qdrant_port: int = Field(default=6333, env="QDRANT_PORT")
    qdrant_api_key: Optional[str] = Field(default=None, env="QDRANT_API_KEY")
    
    # Model Configuration
    default_embedding_model: str = Field(
        default="sentence-transformers/all-MiniLM-L6-v2",
        env="DEFAULT_EMBEDDING_MODEL"
    )
    default_llm_model: str = Field(default="gpt-4", env="DEFAULT_LLM_MODEL")
    max_tokens: int = Field(default=4096, env="MAX_TOKENS")
    temperature: float = Field(default=0.1, env="TEMPERATURE")
    
    # Service URLs
    bounty_service_url: str = Field(default="http://localhost:8080", env="BOUNTY_SERVICE_URL")
    marketplace_service_url: str = Field(default="http://localhost:8081", env="MARKETPLACE_SERVICE_URL")
    metrics_service_url: str = Field(default="http://localhost:8082", env="METRICS_SERVICE_URL")
    
    # External APIs
    github_token: Optional[str] = Field(default=None, env="GITHUB_TOKEN")
    defillama_api_url: str = Field(default="https://api.llama.fi", env="DEFILLAMA_API_URL")
    coingecko_api_url: str = Field(default="https://api.coingecko.com/api/v3", env="COINGECKO_API_URL")
    
    # Caching
    cache_ttl: int = Field(default=3600, env="CACHE_TTL")  # 1 hour
    embedding_cache_ttl: int = Field(default=86400, env="EMBEDDING_CACHE_TTL")  # 24 hours
    
    # Processing
    max_concurrent_requests: int = Field(default=10, env="MAX_CONCURRENT_REQUESTS")
    request_timeout: int = Field(default=300, env="REQUEST_TIMEOUT")  # 5 minutes
    batch_size: int = Field(default=32, env="BATCH_SIZE")
    
    # Monitoring
    enable_metrics: bool = Field(default=True, env="ENABLE_METRICS")
    metrics_port: int = Field(default=8084, env="METRICS_PORT")
    log_level: str = Field(default="INFO", env="LOG_LEVEL")
    
    # Security
    api_key: Optional[str] = Field(default=None, env="AI_SERVICE_API_KEY")
    allowed_origins: List[str] = Field(
        default=["http://localhost:3000", "http://localhost:3001"],
        env="ALLOWED_ORIGINS"
    )
    
    # Model Storage
    model_storage_path: str = Field(default="./models", env="MODEL_STORAGE_PATH")
    data_storage_path: str = Field(default="./data", env="DATA_STORAGE_PATH")
    
    class Config:
        env_file = ".env"
        case_sensitive = False


# Global settings instance
settings = Settings()


def get_settings() -> Settings:
    """Get application settings"""
    return settings
