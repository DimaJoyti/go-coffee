// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/math/SafeMath.sol";

/**
 * @title RevenueSharing
 * @dev Manages performance-based revenue sharing for Developer DAO solutions
 */
contract RevenueSharing is ReentrancyGuard, Pausable, Ownable {
    using SafeERC20 for IERC20;
    using SafeMath for uint256;

    // Revenue sharing structure
    struct RevenueShare {
        address developer;
        uint256 solutionId;
        uint256 tvlContribution;
        uint256 mauContribution;
        uint256 totalRevenue;
        uint256 developerShare;
        uint256 lastDistribution;
        bool active;
    }

    // Performance metrics
    struct PerformanceMetrics {
        uint256 totalTVL;
        uint256 totalMAU;
        uint256 totalRevenue;
        uint256 lastUpdate;
    }

    // State variables
    mapping(uint256 => RevenueShare) public revenueShares;
    mapping(address => uint256[]) public developerSolutions;
    mapping(uint256 => PerformanceMetrics) public solutionMetrics;
    
    IERC20 public coffeeToken;
    address public metricsOracle;
    address public treasury;
    
    // Revenue sharing percentages (basis points, 10000 = 100%)
    uint256 public constant DEVELOPER_SHARE_BPS = 3000; // 30%
    uint256 public constant COMMUNITY_SHARE_BPS = 1000; // 10%
    uint256 public constant TREASURY_SHARE_BPS = 6000; // 60%
    
    uint256 public totalSolutions;
    uint256 public totalDistributed;

    // Events
    event SolutionRegistered(
        uint256 indexed solutionId,
        address indexed developer,
        uint256 initialTVL,
        uint256 initialMAU
    );
    
    event RevenueDistributed(
        uint256 indexed solutionId,
        address indexed developer,
        uint256 amount,
        uint256 timestamp
    );
    
    event PerformanceUpdated(
        uint256 indexed solutionId,
        uint256 newTVL,
        uint256 newMAU,
        uint256 revenue
    );

    constructor(
        address _coffeeToken,
        address _metricsOracle,
        address _treasury
    ) {
        coffeeToken = IERC20(_coffeeToken);
        metricsOracle = _metricsOracle;
        treasury = _treasury;
    }

    /**
     * @dev Register a new solution for revenue sharing
     */
    function registerSolution(
        address developer,
        uint256 solutionId,
        uint256 initialTVL,
        uint256 initialMAU
    ) external nonReentrant whenNotPaused {
        require(msg.sender == metricsOracle, "Only metrics oracle");
        require(revenueShares[solutionId].developer == address(0), "Solution already registered");

        revenueShares[solutionId] = RevenueShare({
            developer: developer,
            solutionId: solutionId,
            tvlContribution: initialTVL,
            mauContribution: initialMAU,
            totalRevenue: 0,
            developerShare: 0,
            lastDistribution: block.timestamp,
            active: true
        });

        solutionMetrics[solutionId] = PerformanceMetrics({
            totalTVL: initialTVL,
            totalMAU: initialMAU,
            totalRevenue: 0,
            lastUpdate: block.timestamp
        });

        developerSolutions[developer].push(solutionId);
        totalSolutions++;

        emit SolutionRegistered(solutionId, developer, initialTVL, initialMAU);
    }

    /**
     * @dev Update solution performance metrics
     */
    function updatePerformanceMetrics(
        uint256 solutionId,
        uint256 newTVL,
        uint256 newMAU,
        uint256 generatedRevenue
    ) external nonReentrant whenNotPaused {
        require(msg.sender == metricsOracle, "Only metrics oracle");
        require(revenueShares[solutionId].active, "Solution not active");

        RevenueShare storage share = revenueShares[solutionId];
        PerformanceMetrics storage metrics = solutionMetrics[solutionId];

        // Update metrics
        share.tvlContribution = newTVL;
        share.mauContribution = newMAU;
        share.totalRevenue = share.totalRevenue.add(generatedRevenue);

        metrics.totalTVL = newTVL;
        metrics.totalMAU = newMAU;
        metrics.totalRevenue = metrics.totalRevenue.add(generatedRevenue);
        metrics.lastUpdate = block.timestamp;

        emit PerformanceUpdated(solutionId, newTVL, newMAU, generatedRevenue);

        // Trigger revenue distribution if there's new revenue
        if (generatedRevenue > 0) {
            _distributeRevenue(solutionId, generatedRevenue);
        }
    }

    /**
     * @dev Internal function to distribute revenue
     */
    function _distributeRevenue(uint256 solutionId, uint256 revenue) internal {
        RevenueShare storage share = revenueShares[solutionId];
        
        // Calculate shares
        uint256 developerAmount = revenue.mul(DEVELOPER_SHARE_BPS).div(10000);
        uint256 communityAmount = revenue.mul(COMMUNITY_SHARE_BPS).div(10000);
        uint256 treasuryAmount = revenue.mul(TREASURY_SHARE_BPS).div(10000);

        // Update developer share
        share.developerShare = share.developerShare.add(developerAmount);
        share.lastDistribution = block.timestamp;

        // Transfer tokens
        if (developerAmount > 0) {
            coffeeToken.safeTransfer(share.developer, developerAmount);
        }
        
        if (treasuryAmount > 0) {
            coffeeToken.safeTransfer(treasury, treasuryAmount);
        }

        // Community amount stays in contract for community rewards
        totalDistributed = totalDistributed.add(developerAmount);

        emit RevenueDistributed(solutionId, share.developer, developerAmount, block.timestamp);
    }

    /**
     * @dev Get solution performance data
     */
    function getSolutionPerformance(uint256 solutionId) 
        external 
        view 
        returns (
            address developer,
            uint256 tvlContribution,
            uint256 mauContribution,
            uint256 totalRevenue,
            uint256 developerShare,
            bool active
        ) 
    {
        RevenueShare storage share = revenueShares[solutionId];
        return (
            share.developer,
            share.tvlContribution,
            share.mauContribution,
            share.totalRevenue,
            share.developerShare,
            share.active
        );
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
     * @dev Calculate total developer earnings
     */
    function getTotalDeveloperEarnings(address developer) 
        external 
        view 
        returns (uint256 totalEarnings) 
    {
        uint256[] memory solutions = developerSolutions[developer];
        for (uint256 i = 0; i < solutions.length; i++) {
            totalEarnings = totalEarnings.add(revenueShares[solutions[i]].developerShare);
        }
    }

    /**
     * @dev Deactivate a solution
     */
    function deactivateSolution(uint256 solutionId) external onlyOwner {
        revenueShares[solutionId].active = false;
    }

    /**
     * @dev Update metrics oracle
     */
    function updateMetricsOracle(address newOracle) external onlyOwner {
        metricsOracle = newOracle;
    }

    /**
     * @dev Update treasury address
     */
    function updateTreasury(address newTreasury) external onlyOwner {
        treasury = newTreasury;
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

    function emergencyWithdraw(uint256 amount) external onlyOwner {
        coffeeToken.safeTransfer(owner(), amount);
    }

    /**
     * @dev Get contract statistics
     */
    function getContractStats() 
        external 
        view 
        returns (
            uint256 _totalSolutions,
            uint256 _totalDistributed,
            uint256 contractBalance
        ) 
    {
        return (
            totalSolutions,
            totalDistributed,
            coffeeToken.balanceOf(address(this))
        );
    }
}
