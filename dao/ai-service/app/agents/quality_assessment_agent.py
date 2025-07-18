"""
Quality Assessment Agent - Automated evaluation of solution quality
"""
import re
import ast
import asyncio
from typing import Any, Dict, List, Optional, Tuple
from datetime import datetime

from langchain.tools import BaseTool
from langchain.agents import AgentExecutor, create_openai_functions_agent
from langchain.prompts import ChatPromptTemplate
from langchain_openai import ChatOpenAI

from app.agents.base_agent import BaseAgent, AgentCallback
from app.core.config import get_settings
from app.models.schemas import (
    CodeQualityMetrics,
    QualityIssue,
    ConfidenceLevel
)
from app.services.external_api import ExternalAPIService

settings = get_settings()


class CodeAnalysisTool(BaseTool):
    """Tool for static code analysis"""
    
    name = "code_analysis"
    description = "Analyze code for quality, security, and best practices"
    
    def _run(self, code: str, language: str = "python") -> Dict[str, Any]:
        """Analyze code and return quality metrics"""
        issues = []
        metrics = {
            "lines_of_code": len(code.split('\n')),
            "complexity_score": 0,
            "security_issues": 0,
            "style_issues": 0,
            "performance_issues": 0,
        }
        
        if language.lower() == "python":
            return self._analyze_python_code(code)
        elif language.lower() in ["javascript", "typescript"]:
            return self._analyze_javascript_code(code)
        elif language.lower() == "solidity":
            return self._analyze_solidity_code(code)
        else:
            return self._analyze_generic_code(code)
    
    def _analyze_python_code(self, code: str) -> Dict[str, Any]:
        """Analyze Python code"""
        issues = []
        
        try:
            # Parse AST for complexity analysis
            tree = ast.parse(code)
            complexity = self._calculate_complexity(tree)
            
            # Check for common issues
            if "eval(" in code:
                issues.append({
                    "type": "security",
                    "severity": "high",
                    "description": "Use of eval() function detected - security risk",
                    "suggestion": "Use safer alternatives like ast.literal_eval()"
                })
            
            if "exec(" in code:
                issues.append({
                    "type": "security",
                    "severity": "high",
                    "description": "Use of exec() function detected - security risk",
                    "suggestion": "Avoid dynamic code execution"
                })
            
            # Check for SQL injection patterns
            if re.search(r'execute\s*\(\s*["\'].*%.*["\']', code):
                issues.append({
                    "type": "security",
                    "severity": "critical",
                    "description": "Potential SQL injection vulnerability",
                    "suggestion": "Use parameterized queries"
                })
            
            # Check for hardcoded secrets
            if re.search(r'(password|secret|key|token)\s*=\s*["\'][^"\']+["\']', code, re.IGNORECASE):
                issues.append({
                    "type": "security",
                    "severity": "medium",
                    "description": "Potential hardcoded secret detected",
                    "suggestion": "Use environment variables or secure vaults"
                })
            
        except SyntaxError as e:
            issues.append({
                "type": "syntax",
                "severity": "critical",
                "description": f"Syntax error: {str(e)}",
                "line_number": e.lineno,
                "suggestion": "Fix syntax errors before deployment"
            })
            complexity = 10  # High complexity for broken code
        
        return {
            "complexity_score": complexity,
            "issues": issues,
            "metrics": {
                "lines_of_code": len(code.split('\n')),
                "security_issues": len([i for i in issues if i["type"] == "security"]),
                "style_issues": 0,  # Would integrate with pylint/flake8
                "performance_issues": 0,
            }
        }
    
    def _analyze_solidity_code(self, code: str) -> Dict[str, Any]:
        """Analyze Solidity smart contract code"""
        issues = []
        
        # Check for common Solidity vulnerabilities
        if "tx.origin" in code:
            issues.append({
                "type": "security",
                "severity": "high",
                "description": "Use of tx.origin detected - authentication bypass risk",
                "suggestion": "Use msg.sender instead of tx.origin"
            })
        
        if re.search(r'\.call\s*\(', code):
            issues.append({
                "type": "security",
                "severity": "medium",
                "description": "Low-level call detected - reentrancy risk",
                "suggestion": "Use checks-effects-interactions pattern"
            })
        
        if "block.timestamp" in code and "now" in code:
            issues.append({
                "type": "security",
                "severity": "low",
                "description": "Timestamp dependence detected",
                "suggestion": "Avoid relying on block.timestamp for critical logic"
            })
        
        # Check for overflow/underflow (pre-Solidity 0.8.0)
        if re.search(r'pragma solidity \^0\.[0-7]', code):
            if any(op in code for op in ['+', '-', '*', '/']):
                issues.append({
                    "type": "security",
                    "severity": "medium",
                    "description": "Potential integer overflow/underflow",
                    "suggestion": "Use SafeMath library or upgrade to Solidity 0.8+"
                })
        
        return {
            "complexity_score": min(10, len(code.split('\n')) / 50),
            "issues": issues,
            "metrics": {
                "lines_of_code": len(code.split('\n')),
                "security_issues": len([i for i in issues if i["type"] == "security"]),
                "style_issues": 0,
                "performance_issues": 0,
            }
        }
    
    def _analyze_javascript_code(self, code: str) -> Dict[str, Any]:
        """Analyze JavaScript/TypeScript code"""
        issues = []
        
        # Check for common JavaScript issues
        if "eval(" in code:
            issues.append({
                "type": "security",
                "severity": "high",
                "description": "Use of eval() detected - XSS risk",
                "suggestion": "Use JSON.parse() or safer alternatives"
            })
        
        if "innerHTML" in code:
            issues.append({
                "type": "security",
                "severity": "medium",
                "description": "Use of innerHTML detected - XSS risk",
                "suggestion": "Use textContent or sanitize input"
            })
        
        # Check for console.log in production code
        if "console.log" in code:
            issues.append({
                "type": "style",
                "severity": "low",
                "description": "Console.log statements found",
                "suggestion": "Remove debug statements before production"
            })
        
        return {
            "complexity_score": min(10, code.count('{') / 10),
            "issues": issues,
            "metrics": {
                "lines_of_code": len(code.split('\n')),
                "security_issues": len([i for i in issues if i["type"] == "security"]),
                "style_issues": len([i for i in issues if i["type"] == "style"]),
                "performance_issues": 0,
            }
        }
    
    def _analyze_generic_code(self, code: str) -> Dict[str, Any]:
        """Generic code analysis for unknown languages"""
        issues = []
        lines = code.split('\n')
        
        # Basic checks
        if len(lines) > 1000:
            issues.append({
                "type": "maintainability",
                "severity": "medium",
                "description": "Very large file detected",
                "suggestion": "Consider breaking into smaller modules"
            })
        
        return {
            "complexity_score": min(10, len(lines) / 100),
            "issues": issues,
            "metrics": {
                "lines_of_code": len(lines),
                "security_issues": 0,
                "style_issues": 0,
                "performance_issues": 0,
            }
        }
    
    def _calculate_complexity(self, tree: ast.AST) -> float:
        """Calculate cyclomatic complexity"""
        complexity = 1  # Base complexity
        
        for node in ast.walk(tree):
            if isinstance(node, (ast.If, ast.While, ast.For, ast.AsyncFor)):
                complexity += 1
            elif isinstance(node, ast.ExceptHandler):
                complexity += 1
            elif isinstance(node, (ast.And, ast.Or)):
                complexity += 1
        
        return min(10, complexity / 5)  # Normalize to 0-10 scale


class DocumentationAnalysisTool(BaseTool):
    """Tool for analyzing documentation quality"""
    
    name = "documentation_analysis"
    description = "Analyze documentation completeness and quality"
    
    def _run(self, code: str, readme_content: str = "") -> Dict[str, Any]:
        """Analyze documentation quality"""
        doc_score = 0
        issues = []
        
        # Check for docstrings/comments
        comment_ratio = self._calculate_comment_ratio(code)
        if comment_ratio < 0.1:
            issues.append({
                "type": "documentation",
                "severity": "medium",
                "description": "Low comment ratio detected",
                "suggestion": "Add more inline comments and docstrings"
            })
        
        # Check README quality
        readme_score = self._analyze_readme(readme_content)
        
        # Calculate overall documentation score
        doc_score = (comment_ratio * 50) + (readme_score * 50)
        
        return {
            "documentation_score": min(100, doc_score),
            "comment_ratio": comment_ratio,
            "readme_score": readme_score,
            "issues": issues,
        }
    
    def _calculate_comment_ratio(self, code: str) -> float:
        """Calculate ratio of comments to code"""
        lines = code.split('\n')
        comment_lines = 0
        code_lines = 0
        
        for line in lines:
            stripped = line.strip()
            if not stripped:
                continue
            elif stripped.startswith('#') or stripped.startswith('//') or stripped.startswith('/*'):
                comment_lines += 1
            elif '"""' in stripped or "'''" in stripped:
                comment_lines += 1
            else:
                code_lines += 1
        
        if code_lines == 0:
            return 0
        return comment_lines / (comment_lines + code_lines)
    
    def _analyze_readme(self, readme: str) -> float:
        """Analyze README quality"""
        if not readme:
            return 0
        
        score = 0
        readme_lower = readme.lower()
        
        # Check for essential sections
        if "installation" in readme_lower or "setup" in readme_lower:
            score += 20
        if "usage" in readme_lower or "example" in readme_lower:
            score += 20
        if "api" in readme_lower or "documentation" in readme_lower:
            score += 15
        if "contributing" in readme_lower:
            score += 10
        if "license" in readme_lower:
            score += 10
        if "test" in readme_lower:
            score += 10
        if len(readme) > 500:  # Substantial content
            score += 15
        
        return min(100, score)


class QualityAssessmentAgent(BaseAgent):
    """Agent for automated quality assessment of solutions"""
    
    def __init__(self):
        super().__init__(
            name="quality_assessment",
            description="Automated evaluation of solution quality including code, security, and documentation"
        )
        self.llm = None
        self.external_api = ExternalAPIService()
    
    async def _setup_tools(self):
        """Setup quality assessment tools"""
        self.tools = [
            CodeAnalysisTool(),
            DocumentationAnalysisTool(),
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
            ("system", """You are an expert code quality assessment agent for a Developer DAO platform.
            Your job is to comprehensively evaluate solution quality across multiple dimensions:
            
            1. Code Quality: Security, performance, maintainability, best practices
            2. Documentation: Completeness, clarity, examples, API documentation
            3. Testing: Test coverage, test quality, edge cases
            4. Architecture: Design patterns, scalability, modularity
            
            Use the available tools to perform detailed analysis and provide actionable feedback.
            Always explain your scoring rationale and provide specific improvement suggestions.
            
            Available tools:
            - code_analysis: Analyze code for quality, security, and best practices
            - documentation_analysis: Evaluate documentation completeness and quality
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
        """Process quality assessment request"""
        solution_id = request.get("solution_id")
        repository_url = request.get("repository_url")
        code_content = request.get("code_content")
        language = request.get("language", "python")
        include_suggestions = request.get("include_suggestions", True)
        
        if solution_id:
            return await self._assess_solution_quality(solution_id, include_suggestions)
        elif repository_url:
            return await self._assess_repository_quality(repository_url, language, include_suggestions)
        elif code_content:
            return await self._assess_code_quality(code_content, language, include_suggestions)
        else:
            raise ValueError("Must provide solution_id, repository_url, or code_content")
    
    async def _assess_solution_quality(self, solution_id: int, include_suggestions: bool) -> Dict[str, Any]:
        """Assess quality of a solution by ID"""
        # Get solution data from marketplace service
        solution_data = await self._get_solution_data(solution_id)
        
        # Analyze repository if available
        if solution_data.get("repository_url"):
            return await self._assess_repository_quality(
                solution_data["repository_url"],
                solution_data.get("language", "python"),
                include_suggestions
            )
        else:
            raise ValueError("Solution does not have a repository URL")
    
    async def _assess_repository_quality(self, repository_url: str, language: str, include_suggestions: bool) -> Dict[str, Any]:
        """Assess quality of a GitHub repository"""
        # Fetch repository content
        repo_data = await self._fetch_repository_data(repository_url)
        
        # Analyze main code files
        code_analysis = await self._analyze_code_files(repo_data["files"], language)
        
        # Analyze documentation
        doc_analysis = await self._analyze_documentation(repo_data.get("readme", ""))
        
        # Calculate overall metrics
        metrics = self._calculate_quality_metrics(code_analysis, doc_analysis)
        
        # Generate issues and suggestions
        issues = self._compile_issues(code_analysis, doc_analysis)
        suggestions = await self._generate_suggestions(metrics, issues) if include_suggestions else []
        
        # Determine confidence level
        confidence_level = self._determine_confidence_level(metrics, len(repo_data["files"]))
        
        # Generate analysis summary
        summary = await self._generate_analysis_summary(metrics, issues, suggestions)
        
        return {
            "metrics": metrics.dict(),
            "issues": [issue.dict() for issue in issues],
            "suggestions": suggestions,
            "confidence_level": confidence_level,
            "analysis_summary": summary,
        }
    
    async def _assess_code_quality(self, code_content: str, language: str, include_suggestions: bool) -> Dict[str, Any]:
        """Assess quality of provided code content"""
        # Analyze code
        code_tool = self.tools[0]  # CodeAnalysisTool
        code_analysis = code_tool._run(code_content, language)
        
        # Analyze documentation (inline)
        doc_tool = self.tools[1]  # DocumentationAnalysisTool
        doc_analysis = doc_tool._run(code_content, "")
        
        # Calculate metrics
        metrics = self._calculate_quality_metrics(code_analysis, doc_analysis)
        
        # Generate issues
        issues = self._compile_issues(code_analysis, doc_analysis)
        
        # Generate suggestions
        suggestions = await self._generate_suggestions(metrics, issues) if include_suggestions else []
        
        # Determine confidence
        confidence_level = self._determine_confidence_level(metrics, 1)
        
        # Generate summary
        summary = await self._generate_analysis_summary(metrics, issues, suggestions)
        
        return {
            "metrics": metrics.dict(),
            "issues": [issue.dict() for issue in issues],
            "suggestions": suggestions,
            "confidence_level": confidence_level,
            "analysis_summary": summary,
        }
    
    def _calculate_quality_metrics(self, code_analysis: Dict, doc_analysis: Dict) -> CodeQualityMetrics:
        """Calculate overall quality metrics"""
        # Security score (inverse of security issues)
        security_issues = code_analysis.get("metrics", {}).get("security_issues", 0)
        security_score = max(0, 100 - (security_issues * 20))
        
        # Performance score (inverse of complexity)
        complexity = code_analysis.get("complexity_score", 5)
        performance_score = max(0, 100 - (complexity * 10))
        
        # Maintainability score (based on complexity and style)
        style_issues = code_analysis.get("metrics", {}).get("style_issues", 0)
        maintainability_score = max(0, 100 - (complexity * 8) - (style_issues * 5))
        
        # Documentation score
        documentation_score = doc_analysis.get("documentation_score", 50)
        
        # Test coverage score (would be calculated from actual test analysis)
        test_coverage_score = 70  # Default/estimated value
        
        # Overall score (weighted average)
        overall_score = (
            security_score * 0.25 +
            performance_score * 0.20 +
            maintainability_score * 0.20 +
            documentation_score * 0.20 +
            test_coverage_score * 0.15
        )
        
        return CodeQualityMetrics(
            overall_score=overall_score,
            security_score=security_score,
            performance_score=performance_score,
            maintainability_score=maintainability_score,
            documentation_score=documentation_score,
            test_coverage_score=test_coverage_score,
        )
    
    def _compile_issues(self, code_analysis: Dict, doc_analysis: Dict) -> List[QualityIssue]:
        """Compile all issues from analysis"""
        issues = []
        
        # Add code issues
        for issue in code_analysis.get("issues", []):
            issues.append(QualityIssue(
                type=issue["type"],
                severity=issue["severity"],
                description=issue["description"],
                file_path=issue.get("file_path"),
                line_number=issue.get("line_number"),
                suggestion=issue.get("suggestion"),
            ))
        
        # Add documentation issues
        for issue in doc_analysis.get("issues", []):
            issues.append(QualityIssue(
                type=issue["type"],
                severity=issue["severity"],
                description=issue["description"],
                suggestion=issue.get("suggestion"),
            ))
        
        return issues
    
    async def _generate_suggestions(self, metrics: CodeQualityMetrics, issues: List[QualityIssue]) -> List[str]:
        """Generate improvement suggestions"""
        suggestions = []
        
        # Security suggestions
        if metrics.security_score < 80:
            suggestions.append("Implement comprehensive security review and address all security vulnerabilities")
        
        # Performance suggestions
        if metrics.performance_score < 70:
            suggestions.append("Optimize code complexity and consider performance improvements")
        
        # Documentation suggestions
        if metrics.documentation_score < 60:
            suggestions.append("Improve documentation with better README, API docs, and inline comments")
        
        # Test coverage suggestions
        if metrics.test_coverage_score < 80:
            suggestions.append("Increase test coverage with unit tests, integration tests, and edge cases")
        
        # Issue-specific suggestions
        critical_issues = [i for i in issues if i.severity == "critical"]
        if critical_issues:
            suggestions.append("Address all critical issues before deployment")
        
        return suggestions
    
    def _determine_confidence_level(self, metrics: CodeQualityMetrics, file_count: int) -> ConfidenceLevel:
        """Determine confidence level for the assessment"""
        # Base confidence on overall score and analysis depth
        score = metrics.overall_score
        
        # Adjust for analysis depth
        if file_count >= 10:
            depth_factor = 1.0
        elif file_count >= 5:
            depth_factor = 0.9
        else:
            depth_factor = 0.8
        
        adjusted_score = score * depth_factor
        
        if adjusted_score >= 85:
            return ConfidenceLevel.VERY_HIGH
        elif adjusted_score >= 70:
            return ConfidenceLevel.HIGH
        elif adjusted_score >= 55:
            return ConfidenceLevel.MEDIUM
        else:
            return ConfidenceLevel.LOW
    
    async def _generate_analysis_summary(self, 
                                       metrics: CodeQualityMetrics, 
                                       issues: List[QualityIssue], 
                                       suggestions: List[str]) -> str:
        """Generate human-readable analysis summary"""
        summary_parts = []
        
        # Overall assessment
        if metrics.overall_score >= 85:
            summary_parts.append("Excellent code quality with strong adherence to best practices.")
        elif metrics.overall_score >= 70:
            summary_parts.append("Good code quality with minor areas for improvement.")
        elif metrics.overall_score >= 55:
            summary_parts.append("Acceptable code quality but requires attention to several areas.")
        else:
            summary_parts.append("Code quality needs significant improvement before production use.")
        
        # Specific areas
        if metrics.security_score < 70:
            summary_parts.append("Security concerns require immediate attention.")
        
        if metrics.documentation_score < 60:
            summary_parts.append("Documentation needs substantial improvement.")
        
        # Issue summary
        critical_count = len([i for i in issues if i.severity == "critical"])
        high_count = len([i for i in issues if i.severity == "high"])
        
        if critical_count > 0:
            summary_parts.append(f"{critical_count} critical issue(s) must be resolved.")
        if high_count > 0:
            summary_parts.append(f"{high_count} high-priority issue(s) should be addressed.")
        
        return " ".join(summary_parts)
    
    async def _get_solution_data(self, solution_id: int) -> Dict[str, Any]:
        """Get solution data from marketplace service"""
        # Mock implementation - would call actual service
        return {
            "id": solution_id,
            "repository_url": "https://github.com/example/solution",
            "language": "python",
        }
    
    async def _fetch_repository_data(self, repository_url: str) -> Dict[str, Any]:
        """Fetch repository data from GitHub"""
        # Mock implementation - would use GitHub API
        return {
            "files": {
                "main.py": "def hello_world():\n    print('Hello, World!')\n",
                "requirements.txt": "flask==2.0.1\n",
            },
            "readme": "# Example Project\n\nThis is an example project.\n\n## Installation\n\npip install -r requirements.txt\n",
        }
    
    async def _analyze_code_files(self, files: Dict[str, str], language: str) -> Dict[str, Any]:
        """Analyze multiple code files"""
        code_tool = self.tools[0]  # CodeAnalysisTool
        
        all_issues = []
        total_complexity = 0
        total_metrics = {"security_issues": 0, "style_issues": 0, "performance_issues": 0}
        
        for filename, content in files.items():
            if self._is_code_file(filename):
                analysis = code_tool._run(content, language)
                all_issues.extend(analysis.get("issues", []))
                total_complexity += analysis.get("complexity_score", 0)
                
                for key in total_metrics:
                    total_metrics[key] += analysis.get("metrics", {}).get(key, 0)
        
        return {
            "issues": all_issues,
            "complexity_score": total_complexity / max(1, len(files)),
            "metrics": total_metrics,
        }
    
    async def _analyze_documentation(self, readme_content: str) -> Dict[str, Any]:
        """Analyze documentation quality"""
        doc_tool = self.tools[1]  # DocumentationAnalysisTool
        return doc_tool._run("", readme_content)
    
    def _is_code_file(self, filename: str) -> bool:
        """Check if file is a code file"""
        code_extensions = ['.py', '.js', '.ts', '.sol', '.go', '.rs', '.java', '.cpp', '.c']
        return any(filename.endswith(ext) for ext in code_extensions)
