"""
Base Agent class for AI Service
"""
import asyncio
import time
from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional, Union
from datetime import datetime

from langchain.agents import AgentExecutor
from langchain.tools import BaseTool
from langchain.schema import BaseMessage
from langchain.callbacks.base import BaseCallbackHandler

from app.core.logging import AIServiceLogger
from app.core.config import get_settings
from app.models.schemas import ConfidenceLevel

settings = get_settings()


class AgentMetrics:
    """Agent performance metrics"""
    
    def __init__(self):
        self.total_requests = 0
        self.successful_requests = 0
        self.failed_requests = 0
        self.total_processing_time = 0.0
        self.average_processing_time = 0.0
        self.last_request_time: Optional[datetime] = None
    
    def record_request(self, processing_time: float, success: bool):
        """Record a request"""
        self.total_requests += 1
        self.total_processing_time += processing_time
        self.average_processing_time = self.total_processing_time / self.total_requests
        self.last_request_time = datetime.utcnow()
        
        if success:
            self.successful_requests += 1
        else:
            self.failed_requests += 1
    
    @property
    def success_rate(self) -> float:
        """Calculate success rate"""
        if self.total_requests == 0:
            return 0.0
        return self.successful_requests / self.total_requests
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert metrics to dictionary"""
        return {
            "total_requests": self.total_requests,
            "successful_requests": self.successful_requests,
            "failed_requests": self.failed_requests,
            "success_rate": self.success_rate,
            "average_processing_time": self.average_processing_time,
            "last_request_time": self.last_request_time.isoformat() if self.last_request_time else None,
        }


class BaseAgent(ABC):
    """Base class for all AI agents"""
    
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
        self.logger = AIServiceLogger(f"agent.{name}")
        self.metrics = AgentMetrics()
        self.tools: List[BaseTool] = []
        self.executor: Optional[AgentExecutor] = None
        self.is_initialized = False
    
    async def initialize(self):
        """Initialize the agent"""
        if self.is_initialized:
            return
        
        try:
            self.logger.info(f"Initializing agent: {self.name}")
            await self._setup_tools()
            await self._setup_executor()
            self.is_initialized = True
            self.logger.info(f"Agent {self.name} initialized successfully")
        except Exception as e:
            self.logger.error_with_context(
                e, 
                {"agent": self.name, "operation": "initialization"}
            )
            raise
    
    @abstractmethod
    async def _setup_tools(self):
        """Setup agent-specific tools"""
        pass
    
    @abstractmethod
    async def _setup_executor(self):
        """Setup agent executor"""
        pass
    
    async def process_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Process a request"""
        if not self.is_initialized:
            await self.initialize()
        
        start_time = time.time()
        success = False
        
        try:
            self.logger.log_request(
                endpoint=f"agent.{self.name}",
                method="process",
                request_id=request.get("request_id"),
            )
            
            # Validate request
            await self._validate_request(request)
            
            # Process the request
            result = await self._process_request_internal(request)
            
            # Add metadata
            result.update({
                "agent": self.name,
                "processing_time": time.time() - start_time,
                "timestamp": datetime.utcnow().isoformat(),
                "confidence_level": self._calculate_confidence(result),
            })
            
            success = True
            return result
            
        except Exception as e:
            self.logger.error_with_context(
                e,
                {
                    "agent": self.name,
                    "request_id": request.get("request_id"),
                    "operation": "process_request",
                }
            )
            raise
        finally:
            processing_time = time.time() - start_time
            self.metrics.record_request(processing_time, success)
            
            self.logger.log_response(
                endpoint=f"agent.{self.name}",
                status_code=200 if success else 500,
                duration=processing_time,
            )
    
    @abstractmethod
    async def _process_request_internal(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Internal request processing logic"""
        pass
    
    async def _validate_request(self, request: Dict[str, Any]):
        """Validate request format and content"""
        if not isinstance(request, dict):
            raise ValueError("Request must be a dictionary")
        
        # Add agent-specific validation in subclasses
    
    def _calculate_confidence(self, result: Dict[str, Any]) -> ConfidenceLevel:
        """Calculate confidence level for the result"""
        # Default implementation - override in subclasses
        return ConfidenceLevel.MEDIUM
    
    async def get_health_status(self) -> Dict[str, Any]:
        """Get agent health status"""
        return {
            "name": self.name,
            "description": self.description,
            "is_initialized": self.is_initialized,
            "metrics": self.metrics.to_dict(),
            "tools_count": len(self.tools),
            "status": "healthy" if self.is_initialized else "not_initialized",
        }
    
    async def reset_metrics(self):
        """Reset agent metrics"""
        self.metrics = AgentMetrics()
        self.logger.info(f"Metrics reset for agent: {self.name}")


class AgentCallback(BaseCallbackHandler):
    """Callback handler for agent operations"""
    
    def __init__(self, agent_name: str):
        self.agent_name = agent_name
        self.logger = AIServiceLogger(f"agent.{agent_name}.callback")
    
    def on_llm_start(self, serialized: Dict[str, Any], prompts: List[str], **kwargs):
        """Called when LLM starts"""
        self.logger.debug(
            "LLM started",
            agent=self.agent_name,
            model=serialized.get("name", "unknown"),
            prompt_count=len(prompts),
        )
    
    def on_llm_end(self, response, **kwargs):
        """Called when LLM ends"""
        self.logger.debug(
            "LLM completed",
            agent=self.agent_name,
            token_usage=response.llm_output.get("token_usage") if response.llm_output else None,
        )
    
    def on_llm_error(self, error: Exception, **kwargs):
        """Called when LLM errors"""
        self.logger.error(
            "LLM error",
            agent=self.agent_name,
            error=str(error),
        )
    
    def on_tool_start(self, serialized: Dict[str, Any], input_str: str, **kwargs):
        """Called when tool starts"""
        self.logger.debug(
            "Tool started",
            agent=self.agent_name,
            tool=serialized.get("name", "unknown"),
        )
    
    def on_tool_end(self, output: str, **kwargs):
        """Called when tool ends"""
        self.logger.debug(
            "Tool completed",
            agent=self.agent_name,
            output_length=len(output),
        )
    
    def on_tool_error(self, error: Exception, **kwargs):
        """Called when tool errors"""
        self.logger.error(
            "Tool error",
            agent=self.agent_name,
            error=str(error),
        )


class AgentManager:
    """Manager for all AI agents"""
    
    def __init__(self):
        self.agents: Dict[str, BaseAgent] = {}
        self.logger = AIServiceLogger("agent_manager")
    
    def register_agent(self, agent: BaseAgent):
        """Register an agent"""
        self.agents[agent.name] = agent
        self.logger.info(f"Registered agent: {agent.name}")
    
    async def initialize_all_agents(self):
        """Initialize all registered agents"""
        for name, agent in self.agents.items():
            try:
                await agent.initialize()
            except Exception as e:
                self.logger.error_with_context(
                    e,
                    {"agent": name, "operation": "bulk_initialization"}
                )
    
    async def get_agent(self, name: str) -> Optional[BaseAgent]:
        """Get an agent by name"""
        agent = self.agents.get(name)
        if agent and not agent.is_initialized:
            await agent.initialize()
        return agent
    
    async def get_all_health_status(self) -> Dict[str, Any]:
        """Get health status for all agents"""
        status = {}
        for name, agent in self.agents.items():
            try:
                status[name] = await agent.get_health_status()
            except Exception as e:
                status[name] = {
                    "name": name,
                    "status": "error",
                    "error": str(e),
                }
        return status
    
    def list_agents(self) -> List[str]:
        """List all registered agent names"""
        return list(self.agents.keys())


# Global agent manager instance
agent_manager = AgentManager()
