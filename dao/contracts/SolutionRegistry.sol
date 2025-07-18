// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title SolutionRegistry
 * @dev Registry for DeFi solutions and 3rd party components with quality scoring
 */
contract SolutionRegistry is Ownable, ReentrancyGuard, Pausable {
    using Counters for Counters.Counter;

    // Solution categories
    enum SolutionCategory {
        DEX_INTEGRATION,
        LENDING_PROTOCOL,
        YIELD_FARMING,
        ARBITRAGE_BOT,
        PRICE_ORACLE,
        GOVERNANCE_TOOL,
        ANALYTICS_DASHBOARD,
        SECURITY_AUDIT,
        OTHER
    }

    // Solution status
    enum SolutionStatus {
        SUBMITTED,
        UNDER_REVIEW,
        APPROVED,
        REJECTED,
        DEPRECATED
    }

    // Quality metrics
    struct QualityMetrics {
        uint256 securityScore;      // 0-100
        uint256 performanceScore;   // 0-100
        uint256 usabilityScore;     // 0-100
        uint256 documentationScore; // 0-100
        uint256 overallScore;       // Calculated average
        uint256 auditCount;
        bool hasSecurityAudit;
    }

    // Solution structure
    struct Solution {
        uint256 id;
        string name;
        string description;
        string repositoryUrl;
        string documentationUrl;
        SolutionCategory category;
        SolutionStatus status;
        address developer;
        address[] maintainers;
        string version;
        uint256 createdAt;
        uint256 lastUpdated;
        QualityMetrics quality;
        mapping(address => bool) isMaintainer;
        uint256 deploymentCount;
        uint256 tvlGenerated;
        uint256 mauGenerated;
        bool isThirdParty;
        uint256 licenseType; // 0: MIT, 1: GPL, 2: Commercial, etc.
    }

    // Compatibility tracking
    struct Compatibility {
        uint256 solutionId;
        uint256[] compatibleSolutions;
        uint256[] incompatibleSolutions;
        string[] supportedNetworks;
        string[] requiredDependencies;
    }

    // State variables
    Counters.Counter private _solutionIds;
    mapping(uint256 => Solution) public solutions;
    mapping(uint256 => Compatibility) public compatibility;
    mapping(address => uint256[]) public developerSolutions;
    mapping(SolutionCategory => uint256[]) public solutionsByCategory;
    mapping(address => bool) public authorizedReviewers;
    
    uint256 public constant MIN_QUALITY_SCORE = 70;
    uint256 public totalSolutions;
    uint256 public approvedSolutions;

    // Events
    event SolutionSubmitted(
        uint256 indexed solutionId,
        address indexed developer,
        string name,
        SolutionCategory category
    );

    event SolutionApproved(uint256 indexed solutionId, uint256 qualityScore);
    event SolutionRejected(uint256 indexed solutionId, string reason);
    event QualityScoreUpdated(uint256 indexed solutionId, uint256 newScore);
    event MaintainerAdded(uint256 indexed solutionId, address maintainer);
    event CompatibilityUpdated(uint256 indexed solutionId);

    constructor() {
        authorizedReviewers[msg.sender] = true;
    }

    /**
     * @dev Submit a new solution for review
     */
    function submitSolution(
        string memory name,
        string memory description,
        string memory repositoryUrl,
        string memory documentationUrl,
        SolutionCategory category,
        string memory version,
        bool isThirdParty,
        uint256 licenseType
    ) external nonReentrant whenNotPaused returns (uint256) {
        _solutionIds.increment();
        uint256 solutionId = _solutionIds.current();

        Solution storage solution = solutions[solutionId];
        solution.id = solutionId;
        solution.name = name;
        solution.description = description;
        solution.repositoryUrl = repositoryUrl;
        solution.documentationUrl = documentationUrl;
        solution.category = category;
        solution.status = SolutionStatus.SUBMITTED;
        solution.developer = msg.sender;
        solution.version = version;
        solution.createdAt = block.timestamp;
        solution.lastUpdated = block.timestamp;
        solution.isThirdParty = isThirdParty;
        solution.licenseType = licenseType;

        // Add developer as maintainer
        solution.maintainers.push(msg.sender);
        solution.isMaintainer[msg.sender] = true;

        // Initialize quality metrics
        solution.quality = QualityMetrics({
            securityScore: 0,
            performanceScore: 0,
            usabilityScore: 0,
            documentationScore: 0,
            overallScore: 0,
            auditCount: 0,
            hasSecurityAudit: false
        });

        developerSolutions[msg.sender].push(solutionId);
        solutionsByCategory[category].push(solutionId);
        totalSolutions++;

        emit SolutionSubmitted(solutionId, msg.sender, name, category);
        return solutionId;
    }

    /**
     * @dev Review and score a solution
     */
    function reviewSolution(
        uint256 solutionId,
        uint256 securityScore,
        uint256 performanceScore,
        uint256 usabilityScore,
        uint256 documentationScore,
        bool hasSecurityAudit
    ) external nonReentrant whenNotPaused {
        require(authorizedReviewers[msg.sender], "Not authorized reviewer");
        require(solutions[solutionId].status == SolutionStatus.SUBMITTED, "Solution not under review");

        Solution storage solution = solutions[solutionId];
        solution.status = SolutionStatus.UNDER_REVIEW;

        // Update quality metrics
        QualityMetrics storage quality = solution.quality;
        quality.securityScore = securityScore;
        quality.performanceScore = performanceScore;
        quality.usabilityScore = usabilityScore;
        quality.documentationScore = documentationScore;
        quality.hasSecurityAudit = hasSecurityAudit;

        // Calculate overall score
        quality.overallScore = (securityScore + performanceScore + usabilityScore + documentationScore) / 4;

        if (hasSecurityAudit) {
            quality.auditCount++;
            quality.overallScore += 5; // Bonus for security audit
        }

        solution.lastUpdated = block.timestamp;

        emit QualityScoreUpdated(solutionId, quality.overallScore);
    }

    /**
     * @dev Approve a solution
     */
    function approveSolution(uint256 solutionId) external nonReentrant whenNotPaused {
        require(authorizedReviewers[msg.sender], "Not authorized reviewer");
        
        Solution storage solution = solutions[solutionId];
        require(solution.status == SolutionStatus.UNDER_REVIEW, "Solution not under review");
        require(solution.quality.overallScore >= MIN_QUALITY_SCORE, "Quality score too low");

        solution.status = SolutionStatus.APPROVED;
        solution.lastUpdated = block.timestamp;
        approvedSolutions++;

        emit SolutionApproved(solutionId, solution.quality.overallScore);
    }

    /**
     * @dev Reject a solution
     */
    function rejectSolution(uint256 solutionId, string memory reason) 
        external 
        nonReentrant 
        whenNotPaused 
    {
        require(authorizedReviewers[msg.sender], "Not authorized reviewer");
        
        Solution storage solution = solutions[solutionId];
        require(
            solution.status == SolutionStatus.SUBMITTED || 
            solution.status == SolutionStatus.UNDER_REVIEW, 
            "Cannot reject solution"
        );

        solution.status = SolutionStatus.REJECTED;
        solution.lastUpdated = block.timestamp;

        emit SolutionRejected(solutionId, reason);
    }

    /**
     * @dev Add maintainer to solution
     */
    function addMaintainer(uint256 solutionId, address maintainer) 
        external 
        nonReentrant 
        whenNotPaused 
    {
        Solution storage solution = solutions[solutionId];
        require(
            msg.sender == solution.developer || solution.isMaintainer[msg.sender],
            "Not authorized"
        );
        require(!solution.isMaintainer[maintainer], "Already a maintainer");

        solution.maintainers.push(maintainer);
        solution.isMaintainer[maintainer] = true;

        emit MaintainerAdded(solutionId, maintainer);
    }

    /**
     * @dev Update solution compatibility
     */
    function updateCompatibility(
        uint256 solutionId,
        uint256[] memory compatibleSolutions,
        uint256[] memory incompatibleSolutions,
        string[] memory supportedNetworks,
        string[] memory requiredDependencies
    ) external nonReentrant whenNotPaused {
        Solution storage solution = solutions[solutionId];
        require(
            msg.sender == solution.developer || 
            solution.isMaintainer[msg.sender] ||
            authorizedReviewers[msg.sender],
            "Not authorized"
        );

        Compatibility storage comp = compatibility[solutionId];
        comp.solutionId = solutionId;
        comp.compatibleSolutions = compatibleSolutions;
        comp.incompatibleSolutions = incompatibleSolutions;
        comp.supportedNetworks = supportedNetworks;
        comp.requiredDependencies = requiredDependencies;

        emit CompatibilityUpdated(solutionId);
    }

    /**
     * @dev Update solution performance metrics
     */
    function updatePerformanceMetrics(
        uint256 solutionId,
        uint256 deploymentCount,
        uint256 tvlGenerated,
        uint256 mauGenerated
    ) external {
        require(authorizedReviewers[msg.sender], "Not authorized");
        
        Solution storage solution = solutions[solutionId];
        solution.deploymentCount = deploymentCount;
        solution.tvlGenerated = tvlGenerated;
        solution.mauGenerated = mauGenerated;
        solution.lastUpdated = block.timestamp;
    }

    /**
     * @dev Get solution details
     */
    function getSolution(uint256 solutionId) 
        external 
        view 
        returns (
            string memory name,
            string memory description,
            SolutionCategory category,
            SolutionStatus status,
            address developer,
            string memory version,
            uint256 overallScore,
            uint256 deploymentCount
        ) 
    {
        Solution storage solution = solutions[solutionId];
        return (
            solution.name,
            solution.description,
            solution.category,
            solution.status,
            solution.developer,
            solution.version,
            solution.quality.overallScore,
            solution.deploymentCount
        );
    }

    /**
     * @dev Get solutions by category
     */
    function getSolutionsByCategory(SolutionCategory category) 
        external 
        view 
        returns (uint256[] memory) 
    {
        return solutionsByCategory[category];
    }

    /**
     * @dev Get developer's solutions
     */
    function getDeveloperSolutions(address developer) 
        external 
        view 
        returns (uint256[] memory) 
    {
        return developerSolutions[developer];
    }

    /**
     * @dev Add authorized reviewer
     */
    function addReviewer(address reviewer) external onlyOwner {
        authorizedReviewers[reviewer] = true;
    }

    /**
     * @dev Remove authorized reviewer
     */
    function removeReviewer(address reviewer) external onlyOwner {
        authorizedReviewers[reviewer] = false;
    }

    /**
     * @dev Emergency functions
     */
    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    /**
     * @dev Get registry statistics
     */
    function getRegistryStats() 
        external 
        view 
        returns (
            uint256 _totalSolutions,
            uint256 _approvedSolutions,
            uint256 rejectedSolutions,
            uint256 pendingSolutions
        ) 
    {
        uint256 rejected = 0;
        uint256 pending = 0;
        
        for (uint256 i = 1; i <= _solutionIds.current(); i++) {
            if (solutions[i].status == SolutionStatus.REJECTED) {
                rejected++;
            } else if (solutions[i].status == SolutionStatus.SUBMITTED || 
                      solutions[i].status == SolutionStatus.UNDER_REVIEW) {
                pending++;
            }
        }

        return (totalSolutions, approvedSolutions, rejected, pending);
    }
}
