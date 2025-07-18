"""
Logging configuration for AI Service
"""
import logging
import sys
from typing import Any, Dict
import structlog
from structlog.stdlib import LoggerFactory

from app.core.config import get_settings

settings = get_settings()


def configure_logging():
    """Configure structured logging"""
    
    # Configure structlog
    structlog.configure(
        processors=[
            structlog.stdlib.filter_by_level,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.stdlib.PositionalArgumentsFormatter(),
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.UnicodeDecoder(),
            structlog.processors.JSONRenderer()
        ],
        context_class=dict,
        logger_factory=LoggerFactory(),
        wrapper_class=structlog.stdlib.BoundLogger,
        cache_logger_on_first_use=True,
    )
    
    # Configure standard logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, settings.log_level.upper()),
    )
    
    # Set specific logger levels
    logging.getLogger("uvicorn").setLevel(logging.INFO)
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)
    logging.getLogger("sqlalchemy.engine").setLevel(logging.WARNING)


def get_logger(name: str = None) -> structlog.stdlib.BoundLogger:
    """Get a structured logger"""
    return structlog.get_logger(name)


class AIServiceLogger:
    """AI Service specific logger with context"""
    
    def __init__(self, component: str):
        self.logger = get_logger(component)
        self.component = component
    
    def info(self, message: str, **kwargs):
        """Log info message"""
        self.logger.info(message, component=self.component, **kwargs)
    
    def warning(self, message: str, **kwargs):
        """Log warning message"""
        self.logger.warning(message, component=self.component, **kwargs)
    
    def error(self, message: str, **kwargs):
        """Log error message"""
        self.logger.error(message, component=self.component, **kwargs)
    
    def debug(self, message: str, **kwargs):
        """Log debug message"""
        self.logger.debug(message, component=self.component, **kwargs)
    
    def log_request(self, endpoint: str, method: str, **kwargs):
        """Log API request"""
        self.logger.info(
            "API request",
            component=self.component,
            endpoint=endpoint,
            method=method,
            **kwargs
        )
    
    def log_response(self, endpoint: str, status_code: int, duration: float, **kwargs):
        """Log API response"""
        self.logger.info(
            "API response",
            component=self.component,
            endpoint=endpoint,
            status_code=status_code,
            duration_ms=round(duration * 1000, 2),
            **kwargs
        )
    
    def log_ai_operation(self, operation: str, model: str, duration: float, **kwargs):
        """Log AI operation"""
        self.logger.info(
            "AI operation",
            component=self.component,
            operation=operation,
            model=model,
            duration_ms=round(duration * 1000, 2),
            **kwargs
        )
    
    def log_error_with_context(self, error: Exception, context: Dict[str, Any]):
        """Log error with additional context"""
        self.logger.error(
            "Error occurred",
            component=self.component,
            error_type=type(error).__name__,
            error_message=str(error),
            **context
        )
