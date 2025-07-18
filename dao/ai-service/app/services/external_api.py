"""
External API Service for integrating with platform services and external APIs
"""
import asyncio
import httpx
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta

from app.core.config import get_settings
from app.core.logging import AIServiceLogger

settings = get_settings()


class ExternalAPIService:
    """Service for external API integrations"""
    
    def __init__(self):
        self.logger = AIServiceLogger("external_api")
        self.client = httpx.AsyncClient(timeout=30.0)
    
    async def close(self):
        """Close HTTP client"""
        await self.client.aclose()
    
    # Bounty Service Integration
    async def get_bounty(self, bounty_id: int) -> Optional[Dict[str, Any]]:
        """Get bounty data from bounty service"""
        try:
            url = f"{settings.bounty_service_url}/api/v1/bounties/{bounty_id}"
            response = await self.client.get(url)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch bounty", bounty_id=bounty_id, error=str(e))
            return None
    
    async def get_bounties(self, 
                          status: Optional[str] = None,
                          category: Optional[str] = None,
                          limit: int = 100) -> List[Dict[str, Any]]:
        """Get list of bounties"""
        try:
            url = f"{settings.bounty_service_url}/api/v1/bounties"
            params = {"limit": limit}
            if status:
                params["status"] = status
            if category:
                params["category"] = category
            
            response = await self.client.get(url, params=params)
            response.raise_for_status()
            return response.json().get("bounties", [])
        except Exception as e:
            self.logger.error("Failed to fetch bounties", error=str(e))
            return []
    
    async def get_developer_profile(self, address: str) -> Optional[Dict[str, Any]]:
        """Get developer profile from bounty service"""
        try:
            url = f"{settings.bounty_service_url}/api/v1/performance/{address}"
            response = await self.client.get(url)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch developer profile", address=address, error=str(e))
            return None
    
    # Marketplace Service Integration
    async def get_solution(self, solution_id: int) -> Optional[Dict[str, Any]]:
        """Get solution data from marketplace service"""
        try:
            url = f"{settings.marketplace_service_url}/api/v1/solutions/{solution_id}"
            response = await self.client.get(url)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch solution", solution_id=solution_id, error=str(e))
            return None
    
    async def get_solutions(self, 
                           category: Optional[str] = None,
                           limit: int = 100) -> List[Dict[str, Any]]:
        """Get list of solutions"""
        try:
            url = f"{settings.marketplace_service_url}/api/v1/solutions"
            params = {"limit": limit}
            if category:
                params["category"] = category
            
            response = await self.client.get(url, params=params)
            response.raise_for_status()
            return response.json().get("solutions", [])
        except Exception as e:
            self.logger.error("Failed to fetch solutions", error=str(e))
            return []
    
    # Metrics Service Integration
    async def get_tvl_metrics(self, days: int = 30) -> Optional[Dict[str, Any]]:
        """Get TVL metrics from metrics service"""
        try:
            url = f"{settings.metrics_service_url}/api/v1/tvl"
            params = {"days": days}
            response = await self.client.get(url, params=params)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch TVL metrics", error=str(e))
            return None
    
    async def get_mau_metrics(self, days: int = 30) -> Optional[Dict[str, Any]]:
        """Get MAU metrics from metrics service"""
        try:
            url = f"{settings.metrics_service_url}/api/v1/mau"
            params = {"days": days}
            response = await self.client.get(url, params=params)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch MAU metrics", error=str(e))
            return None
    
    # GitHub API Integration
    async def get_repository_info(self, repo_url: str) -> Optional[Dict[str, Any]]:
        """Get repository information from GitHub API"""
        try:
            # Extract owner and repo from URL
            parts = repo_url.replace("https://github.com/", "").split("/")
            if len(parts) < 2:
                return None
            
            owner, repo = parts[0], parts[1]
            
            headers = {}
            if settings.github_token:
                headers["Authorization"] = f"token {settings.github_token}"
            
            # Get repository info
            url = f"https://api.github.com/repos/{owner}/{repo}"
            response = await self.client.get(url, headers=headers)
            response.raise_for_status()
            repo_info = response.json()
            
            # Get repository contents
            contents_url = f"https://api.github.com/repos/{owner}/{repo}/contents"
            contents_response = await self.client.get(contents_url, headers=headers)
            contents_response.raise_for_status()
            contents = contents_response.json()
            
            return {
                "info": repo_info,
                "contents": contents,
                "languages": await self._get_repository_languages(owner, repo, headers),
                "readme": await self._get_repository_readme(owner, repo, headers),
            }
            
        except Exception as e:
            self.logger.error("Failed to fetch repository info", repo_url=repo_url, error=str(e))
            return None
    
    async def _get_repository_languages(self, owner: str, repo: str, headers: Dict) -> Dict[str, int]:
        """Get repository languages"""
        try:
            url = f"https://api.github.com/repos/{owner}/{repo}/languages"
            response = await self.client.get(url, headers=headers)
            response.raise_for_status()
            return response.json()
        except:
            return {}
    
    async def _get_repository_readme(self, owner: str, repo: str, headers: Dict) -> str:
        """Get repository README content"""
        try:
            url = f"https://api.github.com/repos/{owner}/{repo}/readme"
            response = await self.client.get(url, headers=headers)
            response.raise_for_status()
            readme_info = response.json()
            
            # Get the actual content
            content_url = readme_info["download_url"]
            content_response = await self.client.get(content_url)
            content_response.raise_for_status()
            return content_response.text
        except:
            return ""
    
    # DeFiLlama API Integration
    async def get_defi_protocols(self) -> List[Dict[str, Any]]:
        """Get DeFi protocols data from DeFiLlama"""
        try:
            url = f"{settings.defillama_api_url}/protocols"
            response = await self.client.get(url)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch DeFi protocols", error=str(e))
            return []
    
    async def get_defi_tvl_history(self, protocol: str = None) -> Optional[Dict[str, Any]]:
        """Get DeFi TVL history"""
        try:
            if protocol:
                url = f"{settings.defillama_api_url}/protocol/{protocol}"
            else:
                url = f"{settings.defillama_api_url}/charts"
            
            response = await self.client.get(url)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch DeFi TVL history", protocol=protocol, error=str(e))
            return None
    
    # CoinGecko API Integration
    async def get_market_data(self, coin_ids: List[str] = None) -> Optional[Dict[str, Any]]:
        """Get cryptocurrency market data from CoinGecko"""
        try:
            if not coin_ids:
                coin_ids = ["ethereum", "bitcoin", "binancecoin"]
            
            ids = ",".join(coin_ids)
            url = f"{settings.coingecko_api_url}/simple/price"
            params = {
                "ids": ids,
                "vs_currencies": "usd",
                "include_24hr_change": "true",
                "include_market_cap": "true",
            }
            
            response = await self.client.get(url, params=params)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch market data", error=str(e))
            return None
    
    async def get_trending_coins(self) -> Optional[Dict[str, Any]]:
        """Get trending cryptocurrencies"""
        try:
            url = f"{settings.coingecko_api_url}/search/trending"
            response = await self.client.get(url)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            self.logger.error("Failed to fetch trending coins", error=str(e))
            return None
    
    # Utility Methods
    async def health_check_external_services(self) -> Dict[str, bool]:
        """Check health of external services"""
        services = {
            "bounty_service": False,
            "marketplace_service": False,
            "metrics_service": False,
            "github_api": False,
            "defillama_api": False,
            "coingecko_api": False,
        }
        
        # Check bounty service
        try:
            response = await self.client.get(f"{settings.bounty_service_url}/health", timeout=5.0)
            services["bounty_service"] = response.status_code == 200
        except:
            pass
        
        # Check marketplace service
        try:
            response = await self.client.get(f"{settings.marketplace_service_url}/health", timeout=5.0)
            services["marketplace_service"] = response.status_code == 200
        except:
            pass
        
        # Check metrics service
        try:
            response = await self.client.get(f"{settings.metrics_service_url}/health", timeout=5.0)
            services["metrics_service"] = response.status_code == 200
        except:
            pass
        
        # Check GitHub API
        try:
            headers = {}
            if settings.github_token:
                headers["Authorization"] = f"token {settings.github_token}"
            response = await self.client.get("https://api.github.com/rate_limit", headers=headers, timeout=5.0)
            services["github_api"] = response.status_code == 200
        except:
            pass
        
        # Check DeFiLlama API
        try:
            response = await self.client.get(f"{settings.defillama_api_url}/protocols", timeout=5.0)
            services["defillama_api"] = response.status_code == 200
        except:
            pass
        
        # Check CoinGecko API
        try:
            response = await self.client.get(f"{settings.coingecko_api_url}/ping", timeout=5.0)
            services["coingecko_api"] = response.status_code == 200
        except:
            pass
        
        return services
