"""
Bounty Matching Agent - Intelligent matching of developers to bounties
"""
import asyncio
import numpy as np
from typing import Any, Dict, List, Optional, Tuple
from datetime import datetime, timedelta

from langchain.tools import BaseTool
from langchain.agents import AgentExecutor, create_openai_functions_agent
from langchain.prompts import ChatPromptTemplate
from langchain_openai import ChatOpenAI
from sentence_transformers import SentenceTransformer
from sklearn.metrics.pairwise import cosine_similarity

from app.agents.base_agent import BaseAgent, AgentCallback
from app.core.config import get_settings
from app.core.database import get_qdrant, get_redis
from app.models.schemas import (
    BountyMatch, 
    DeveloperProfile, 
    BountyRequirements,
    ConfidenceLevel
)
from app.services.external_api import ExternalAPIService

settings = get_settings()


class SkillMatchingTool(BaseTool):
    """Tool for skill-based matching"""
    
    name = "skill_matching"
    description = "Match developer skills with bounty requirements using semantic similarity"
    
    def __init__(self, embedding_model: SentenceTransformer):
        super().__init__()
        self.embedding_model = embedding_model
    
    def _run(self, developer_skills: List[str], required_skills: List[str]) -> float:
        """Calculate skill matching score"""
        if not developer_skills or not required_skills:
            return 0.0
        
        # Generate embeddings
        dev_embeddings = self.embedding_model.encode(developer_skills)
        req_embeddings = self.embedding_model.encode(required_skills)
        
        # Calculate similarity matrix
        similarity_matrix = cosine_similarity(dev_embeddings, req_embeddings)
        
        # Calculate best matches for each required skill
        best_matches = np.max(similarity_matrix, axis=0)
        
        # Return average of best matches
        return float(np.mean(best_matches))


class ExperienceMatchingTool(BaseTool):
    """Tool for experience-based matching"""
    
    name = "experience_matching"
    description = "Match developer experience level with bounty difficulty"
    
    def _run(self, developer_experience: str, bounty_difficulty: str, reputation_score: float) -> float:
        """Calculate experience matching score"""
        experience_levels = {
            "junior": 1,
            "mid": 2,
            "senior": 3,
            "expert": 4
        }
        
        difficulty_levels = {
            "easy": 1,
            "medium": 2,
            "hard": 3,
            "expert": 4
        }
        
        dev_level = experience_levels.get(developer_experience.lower(), 2)
        bounty_level = difficulty_levels.get(bounty_difficulty.lower(), 2)
        
        # Calculate base match score
        level_diff = abs(dev_level - bounty_level)
        base_score = max(0, 1 - (level_diff * 0.25))
        
        # Adjust based on reputation
        reputation_factor = min(1.0, reputation_score / 10.0)
        
        return base_score * (0.7 + 0.3 * reputation_factor)


class AvailabilityMatchingTool(BaseTool):
    """Tool for availability-based matching"""
    
    name = "availability_matching"
    description = "Match developer availability with bounty timeline"
    
    def _run(self, 
             developer_hours: Optional[int], 
             estimated_hours: Optional[int],
             deadline: str,
             active_bounties: int) -> float:
        """Calculate availability matching score"""
        if not developer_hours or not estimated_hours:
            return 0.5  # Default score when data is missing
        
        # Parse deadline
        try:
            deadline_date = datetime.fromisoformat(deadline.replace('Z', '+00:00'))
            days_until_deadline = (deadline_date - datetime.utcnow()).days
        except:
            days_until_deadline = 30  # Default to 30 days
        
        # Calculate required hours per week
        weeks_available = max(1, days_until_deadline / 7)
        required_hours_per_week = estimated_hours / weeks_available
        
        # Account for existing workload
        workload_factor = max(0.1, 1 - (active_bounties * 0.2))
        available_hours_per_week = developer_hours * workload_factor
        
        # Calculate availability score
        if required_hours_per_week <= available_hours_per_week:
            return min(1.0, available_hours_per_week / required_hours_per_week)
        else:
            return max(0.0, available_hours_per_week / required_hours_per_week)


class BountyMatchingAgent(BaseAgent):
    """Agent for intelligent bounty-developer matching"""
    
    def __init__(self):
        super().__init__(
            name="bounty_matching",
            description="Intelligent matching of developers to bounties based on skills, experience, and availability"
        )
        self.embedding_model = None
        self.llm = None
        self.external_api = ExternalAPIService()
    
    async def _setup_tools(self):
        """Setup matching tools"""
        # Load embedding model
        self.embedding_model = SentenceTransformer(settings.default_embedding_model)
        
        # Initialize tools
        self.tools = [
            SkillMatchingTool(self.embedding_model),
            ExperienceMatchingTool(),
            AvailabilityMatchingTool(),
        ]
    
    async def _setup_executor(self):
        """Setup agent executor"""
        # Initialize LLM
        self.llm = ChatOpenAI(
            model=settings.default_llm_model,
            temperature=settings.temperature,
            openai_api_key=settings.openai_api_key,
        )
        
        # Create prompt template
        prompt = ChatPromptTemplate.from_messages([
            ("system", """You are an expert bounty matching agent for a Developer DAO platform.
            Your job is to intelligently match developers with bounties based on:
            1. Skill compatibility using semantic analysis
            2. Experience level matching with bounty difficulty
            3. Developer availability and workload
            4. Historical performance and success patterns
            
            Use the available tools to calculate matching scores and provide detailed explanations
            for your recommendations. Always consider the developer's success probability and
            provide actionable insights.
            
            Available tools:
            - skill_matching: Calculate semantic similarity between skills
            - experience_matching: Match experience level with difficulty
            - availability_matching: Assess timeline compatibility
            """),
            ("human", "{input}"),
            ("placeholder", "{agent_scratchpad}"),
        ])
        
        # Create agent
        agent = create_openai_functions_agent(self.llm, self.tools, prompt)
        
        # Create executor
        self.executor = AgentExecutor(
            agent=agent,
            tools=self.tools,
            callbacks=[AgentCallback(self.name)],
            verbose=settings.debug,
            max_iterations=5,
        )
    
    async def _process_request_internal(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Process bounty matching request"""
        bounty_id = request.get("bounty_id")
        developer_address = request.get("developer_address")
        limit = request.get("limit", 10)
        min_score = request.get("min_compatibility_score", 0.5)
        
        if bounty_id and developer_address:
            # Match specific bounty with specific developer
            return await self._match_bounty_developer(bounty_id, developer_address)
        elif bounty_id:
            # Find best developers for a bounty
            return await self._find_developers_for_bounty(bounty_id, limit, min_score)
        elif developer_address:
            # Find best bounties for a developer
            return await self._find_bounties_for_developer(developer_address, limit, min_score)
        else:
            raise ValueError("Must provide either bounty_id, developer_address, or both")
    
    async def _match_bounty_developer(self, bounty_id: int, developer_address: str) -> Dict[str, Any]:
        """Match a specific bounty with a specific developer"""
        # Get bounty and developer data
        bounty = await self._get_bounty_data(bounty_id)
        developer = await self._get_developer_data(developer_address)
        
        # Calculate match
        match = await self._calculate_match(developer, bounty)
        
        return {
            "matches": [match.dict()],
            "total_matches": 1,
        }
    
    async def _find_developers_for_bounty(self, bounty_id: int, limit: int, min_score: float) -> Dict[str, Any]:
        """Find best developers for a bounty"""
        # Get bounty data
        bounty = await self._get_bounty_data(bounty_id)
        
        # Get available developers
        developers = await self._get_available_developers()
        
        # Calculate matches
        matches = []
        for developer in developers:
            match = await self._calculate_match(developer, bounty)
            if match.compatibility_score >= min_score:
                matches.append(match)
        
        # Sort by compatibility score
        matches.sort(key=lambda x: x.compatibility_score, reverse=True)
        
        return {
            "matches": [match.dict() for match in matches[:limit]],
            "total_matches": len(matches),
        }
    
    async def _find_bounties_for_developer(self, developer_address: str, limit: int, min_score: float) -> Dict[str, Any]:
        """Find best bounties for a developer"""
        # Get developer data
        developer = await self._get_developer_data(developer_address)
        
        # Get available bounties
        bounties = await self._get_available_bounties()
        
        # Calculate matches
        matches = []
        for bounty in bounties:
            match = await self._calculate_match(developer, bounty)
            if match.compatibility_score >= min_score:
                matches.append(match)
        
        # Sort by compatibility score
        matches.sort(key=lambda x: x.compatibility_score, reverse=True)
        
        return {
            "matches": [match.dict() for match in matches[:limit]],
            "total_matches": len(matches),
        }
    
    async def _calculate_match(self, developer: DeveloperProfile, bounty: BountyRequirements) -> BountyMatch:
        """Calculate match between developer and bounty"""
        # Calculate individual scores
        skill_score = await self._calculate_skill_match(developer.skills, bounty.required_skills)
        experience_score = await self._calculate_experience_match(developer, bounty)
        availability_score = await self._calculate_availability_match(developer, bounty)
        
        # Calculate overall compatibility score
        weights = {"skill": 0.4, "experience": 0.3, "availability": 0.3}
        compatibility_score = (
            skill_score * weights["skill"] +
            experience_score * weights["experience"] +
            availability_score * weights["availability"]
        )
        
        # Calculate success probability
        success_probability = self._calculate_success_probability(
            developer, bounty, compatibility_score
        )
        
        # Determine confidence level
        confidence_level = self._determine_confidence_level(compatibility_score, success_probability)
        
        # Generate explanation
        explanation = await self._generate_explanation(
            developer, bounty, skill_score, experience_score, 
            availability_score, compatibility_score
        )
        
        # Estimate timeline
        recommended_timeline = self._estimate_timeline(bounty, developer)
        
        return BountyMatch(
            developer_address=developer.address,
            bounty_id=bounty.id,
            compatibility_score=compatibility_score,
            confidence_level=confidence_level,
            skill_match_score=skill_score,
            experience_match_score=experience_score,
            availability_match_score=availability_score,
            success_probability=success_probability,
            explanation=explanation,
            recommended_timeline=recommended_timeline,
        )
    
    async def _calculate_skill_match(self, dev_skills: List[str], req_skills: List[str]) -> float:
        """Calculate skill matching score"""
        tool = self.tools[0]  # SkillMatchingTool
        return tool._run(dev_skills, req_skills)
    
    async def _calculate_experience_match(self, developer: DeveloperProfile, bounty: BountyRequirements) -> float:
        """Calculate experience matching score"""
        tool = self.tools[1]  # ExperienceMatchingTool
        return tool._run(
            developer.experience_level,
            bounty.difficulty_level,
            developer.reputation_score
        )
    
    async def _calculate_availability_match(self, developer: DeveloperProfile, bounty: BountyRequirements) -> float:
        """Calculate availability matching score"""
        tool = self.tools[2]  # AvailabilityMatchingTool
        return tool._run(
            developer.availability_hours,
            bounty.estimated_hours,
            bounty.deadline.isoformat(),
            developer.completed_bounties  # Using as proxy for active bounties
        )
    
    def _calculate_success_probability(self, 
                                     developer: DeveloperProfile, 
                                     bounty: BountyRequirements,
                                     compatibility_score: float) -> float:
        """Calculate probability of successful completion"""
        # Base probability from compatibility
        base_prob = compatibility_score
        
        # Adjust based on developer's success rate
        success_factor = developer.success_rate
        
        # Adjust based on reputation
        reputation_factor = min(1.0, developer.reputation_score / 10.0)
        
        # Combine factors
        success_probability = base_prob * (0.5 + 0.3 * success_factor + 0.2 * reputation_factor)
        
        return min(1.0, success_probability)
    
    def _determine_confidence_level(self, compatibility_score: float, success_probability: float) -> ConfidenceLevel:
        """Determine confidence level for the match"""
        avg_score = (compatibility_score + success_probability) / 2
        
        if avg_score >= 0.8:
            return ConfidenceLevel.VERY_HIGH
        elif avg_score >= 0.65:
            return ConfidenceLevel.HIGH
        elif avg_score >= 0.5:
            return ConfidenceLevel.MEDIUM
        else:
            return ConfidenceLevel.LOW
    
    async def _generate_explanation(self, 
                                  developer: DeveloperProfile,
                                  bounty: BountyRequirements,
                                  skill_score: float,
                                  experience_score: float,
                                  availability_score: float,
                                  compatibility_score: float) -> str:
        """Generate human-readable explanation for the match"""
        explanation_parts = []
        
        # Skill analysis
        if skill_score >= 0.8:
            explanation_parts.append("Excellent skill match with strong expertise in required technologies.")
        elif skill_score >= 0.6:
            explanation_parts.append("Good skill alignment with most required technologies covered.")
        else:
            explanation_parts.append("Partial skill match - may require additional learning or collaboration.")
        
        # Experience analysis
        if experience_score >= 0.8:
            explanation_parts.append("Experience level perfectly matches bounty difficulty.")
        elif experience_score >= 0.6:
            explanation_parts.append("Suitable experience level for this bounty complexity.")
        else:
            explanation_parts.append("Experience level may not fully align with bounty requirements.")
        
        # Availability analysis
        if availability_score >= 0.8:
            explanation_parts.append("Excellent availability to meet project timeline.")
        elif availability_score >= 0.6:
            explanation_parts.append("Adequate availability with manageable timeline.")
        else:
            explanation_parts.append("Limited availability may require timeline adjustments.")
        
        # Overall assessment
        if compatibility_score >= 0.8:
            explanation_parts.append("Highly recommended match with strong success potential.")
        elif compatibility_score >= 0.6:
            explanation_parts.append("Good match with reasonable success probability.")
        else:
            explanation_parts.append("Consider alternative matches or provide additional support.")
        
        return " ".join(explanation_parts)
    
    def _estimate_timeline(self, bounty: BountyRequirements, developer: DeveloperProfile) -> Optional[int]:
        """Estimate recommended timeline in days"""
        if not bounty.estimated_hours or not developer.availability_hours:
            return None
        
        # Calculate based on estimated hours and developer availability
        weekly_hours = developer.availability_hours * 0.8  # 80% efficiency factor
        weeks_needed = bounty.estimated_hours / weekly_hours
        days_needed = int(weeks_needed * 7)
        
        # Add buffer based on complexity
        if bounty.difficulty_level == "expert":
            days_needed = int(days_needed * 1.3)
        elif bounty.difficulty_level == "hard":
            days_needed = int(days_needed * 1.2)
        
        return max(7, days_needed)  # Minimum 1 week
    
    async def _get_bounty_data(self, bounty_id: int) -> BountyRequirements:
        """Get bounty data from bounty service"""
        # This would call the actual bounty service
        # For now, return mock data
        return BountyRequirements(
            id=bounty_id,
            title="DeFi Analytics Dashboard",
            description="Build a comprehensive analytics dashboard for DeFi protocols",
            category="analytics",
            required_skills=["React", "TypeScript", "DeFi", "Web3", "Chart.js"],
            difficulty_level="medium",
            estimated_hours=120,
            reward_amount=5000.0,
            deadline=datetime.utcnow() + timedelta(days=30),
        )
    
    async def _get_developer_data(self, address: str) -> DeveloperProfile:
        """Get developer data from user service"""
        # This would call the actual user service
        # For now, return mock data
        return DeveloperProfile(
            address=address,
            skills=["React", "TypeScript", "Python", "DeFi", "Smart Contracts"],
            experience_level="senior",
            reputation_score=8.5,
            completed_bounties=12,
            success_rate=0.92,
            preferred_categories=["analytics", "defi"],
            availability_hours=30,
        )
    
    async def _get_available_developers(self) -> List[DeveloperProfile]:
        """Get list of available developers"""
        # Mock data - would come from actual service
        return [
            await self._get_developer_data("0x1234567890123456789012345678901234567890"),
            await self._get_developer_data("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
        ]
    
    async def _get_available_bounties(self) -> List[BountyRequirements]:
        """Get list of available bounties"""
        # Mock data - would come from actual service
        return [
            await self._get_bounty_data(1),
            await self._get_bounty_data(2),
        ]
