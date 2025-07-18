// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/**
 * @title BountyManager
 * @dev Manages developer bounties with milestone-based payments and performance tracking
 */
contract BountyManager is ReentrancyGuard, Pausable, Ownable {
    using SafeERC20 for IERC20;
    using Counters for Counters.Counter;

    // Bounty categories
    enum BountyCategory {
        TVL_GROWTH,
        MAU_EXPANSION,
        INNOVATION,
        MAINTENANCE,
        SECURITY,
        INTEGRATION
    }

    // Bounty status
    enum BountyStatus {
        OPEN,
        ASSIGNED,
        IN_PROGRESS,
        SUBMITTED,
        COMPLETED,
        CANCELLED
    }

    // Milestone structure
    struct Milestone {
        string description;
        uint256 reward;
        uint256 deadline;
        bool completed;
        bool paid;
    }

    // Bounty structure
    struct Bounty {
        uint256 id;
        string title;
        string description;
        BountyCategory category;
        BountyStatus status;
        address creator;
        address assignee;
        uint256 totalReward;
        uint256 createdAt;
        uint256 deadline;
        Milestone[] milestones;
        mapping(address => bool) applicants;
        address[] applicantList;
        uint256 tvlImpact;
        uint256 mauImpact;
        bool performanceVerified;
    }

    // State variables
    Counters.Counter private _bountyIds;
    mapping(uint256 => Bounty) public bounties;
    mapping(address => uint256[]) public developerBounties;
    mapping(address => uint256) public developerReputationScore;
    
    IERC20 public coffeeToken;
    address public metricsOracle;
    uint256 public constant MIN_BOUNTY_REWARD = 100 * 10**18; // 100 COFFEE tokens
    uint256 public constant REPUTATION_MULTIPLIER = 10;

    // Events
    event BountyCreated(
        uint256 indexed bountyId,
        address indexed creator,
        BountyCategory category,
        string title,
        uint256 totalReward
    );

    event BountyAssigned(uint256 indexed bountyId, address indexed assignee);
    event BountySubmitted(uint256 indexed bountyId, address indexed assignee);
    event BountyCompleted(uint256 indexed bountyId, address indexed assignee, uint256 totalPaid);
    event MilestoneCompleted(uint256 indexed bountyId, uint256 milestoneIndex, uint256 reward);
    event PerformanceVerified(uint256 indexed bountyId, uint256 tvlImpact, uint256 mauImpact);

    constructor(address _coffeeToken, address _metricsOracle) {
        coffeeToken = IERC20(_coffeeToken);
        metricsOracle = _metricsOracle;
    }

    /**
     * @dev Create a new bounty with milestones
     */
    function createBounty(
        string memory title,
        string memory description,
        BountyCategory category,
        uint256 totalReward,
        uint256 deadline,
        string[] memory milestoneDescriptions,
        uint256[] memory milestoneRewards,
        uint256[] memory milestoneDeadlines
    ) external nonReentrant whenNotPaused returns (uint256) {
        require(totalReward >= MIN_BOUNTY_REWARD, "Reward below minimum");
        require(deadline > block.timestamp, "Invalid deadline");
        require(
            milestoneDescriptions.length == milestoneRewards.length &&
            milestoneRewards.length == milestoneDeadlines.length,
            "Milestone arrays length mismatch"
        );

        // Verify total milestone rewards equal total reward
        uint256 totalMilestoneRewards = 0;
        for (uint256 i = 0; i < milestoneRewards.length; i++) {
            totalMilestoneRewards += milestoneRewards[i];
        }
        require(totalMilestoneRewards == totalReward, "Milestone rewards don't match total");

        // Transfer tokens to contract
        coffeeToken.safeTransferFrom(msg.sender, address(this), totalReward);

        // Create bounty
        _bountyIds.increment();
        uint256 bountyId = _bountyIds.current();

        Bounty storage bounty = bounties[bountyId];
        bounty.id = bountyId;
        bounty.title = title;
        bounty.description = description;
        bounty.category = category;
        bounty.status = BountyStatus.OPEN;
        bounty.creator = msg.sender;
        bounty.totalReward = totalReward;
        bounty.createdAt = block.timestamp;
        bounty.deadline = deadline;

        // Add milestones
        for (uint256 i = 0; i < milestoneDescriptions.length; i++) {
            bounty.milestones.push(Milestone({
                description: milestoneDescriptions[i],
                reward: milestoneRewards[i],
                deadline: milestoneDeadlines[i],
                completed: false,
                paid: false
            }));
        }

        emit BountyCreated(bountyId, msg.sender, category, title, totalReward);
        return bountyId;
    }

    /**
     * @dev Apply for a bounty
     */
    function applyForBounty(uint256 bountyId) external nonReentrant whenNotPaused {
        Bounty storage bounty = bounties[bountyId];
        require(bounty.status == BountyStatus.OPEN, "Bounty not open");
        require(!bounty.applicants[msg.sender], "Already applied");
        require(msg.sender != bounty.creator, "Creator cannot apply");

        bounty.applicants[msg.sender] = true;
        bounty.applicantList.push(msg.sender);
    }

    /**
     * @dev Assign bounty to a developer
     */
    function assignBounty(uint256 bountyId, address assignee) external nonReentrant whenNotPaused {
        Bounty storage bounty = bounties[bountyId];
        require(msg.sender == bounty.creator || msg.sender == owner(), "Not authorized");
        require(bounty.status == BountyStatus.OPEN, "Bounty not open");
        require(bounty.applicants[assignee], "Developer didn't apply");

        bounty.assignee = assignee;
        bounty.status = BountyStatus.ASSIGNED;
        developerBounties[assignee].push(bountyId);

        emit BountyAssigned(bountyId, assignee);
    }

    /**
     * @dev Start working on assigned bounty
     */
    function startBounty(uint256 bountyId) external nonReentrant whenNotPaused {
        Bounty storage bounty = bounties[bountyId];
        require(msg.sender == bounty.assignee, "Not assigned to you");
        require(bounty.status == BountyStatus.ASSIGNED, "Bounty not assigned");

        bounty.status = BountyStatus.IN_PROGRESS;
    }

    /**
     * @dev Complete a milestone
     */
    function completeMilestone(uint256 bountyId, uint256 milestoneIndex) 
        external 
        nonReentrant 
        whenNotPaused 
    {
        Bounty storage bounty = bounties[bountyId];
        require(
            msg.sender == bounty.creator || msg.sender == owner(),
            "Not authorized to complete milestone"
        );
        require(bounty.status == BountyStatus.IN_PROGRESS, "Bounty not in progress");
        require(milestoneIndex < bounty.milestones.length, "Invalid milestone index");

        Milestone storage milestone = bounty.milestones[milestoneIndex];
        require(!milestone.completed, "Milestone already completed");

        milestone.completed = true;
        milestone.paid = true;

        // Pay milestone reward
        coffeeToken.safeTransfer(bounty.assignee, milestone.reward);

        // Update developer reputation
        developerReputationScore[bounty.assignee] += REPUTATION_MULTIPLIER;

        emit MilestoneCompleted(bountyId, milestoneIndex, milestone.reward);

        // Check if all milestones are completed
        bool allCompleted = true;
        for (uint256 i = 0; i < bounty.milestones.length; i++) {
            if (!bounty.milestones[i].completed) {
                allCompleted = false;
                break;
            }
        }

        if (allCompleted) {
            bounty.status = BountyStatus.COMPLETED;
            emit BountyCompleted(bountyId, bounty.assignee, bounty.totalReward);
        }
    }

    /**
     * @dev Submit bounty for review
     */
    function submitBounty(uint256 bountyId) external nonReentrant whenNotPaused {
        Bounty storage bounty = bounties[bountyId];
        require(msg.sender == bounty.assignee, "Not assigned to you");
        require(bounty.status == BountyStatus.IN_PROGRESS, "Bounty not in progress");

        bounty.status = BountyStatus.SUBMITTED;
        emit BountySubmitted(bountyId, msg.sender);
    }

    /**
     * @dev Verify performance impact (called by metrics oracle)
     */
    function verifyPerformance(
        uint256 bountyId,
        uint256 tvlImpact,
        uint256 mauImpact
    ) external {
        require(msg.sender == metricsOracle, "Only metrics oracle");
        
        Bounty storage bounty = bounties[bountyId];
        bounty.tvlImpact = tvlImpact;
        bounty.mauImpact = mauImpact;
        bounty.performanceVerified = true;

        // Bonus rewards for exceptional performance
        if (tvlImpact > 1000000 * 10**18 || mauImpact > 1000) { // $1M TVL or 1000 MAU
            uint256 bonus = bounty.totalReward / 10; // 10% bonus
            coffeeToken.safeTransfer(bounty.assignee, bonus);
            developerReputationScore[bounty.assignee] += REPUTATION_MULTIPLIER * 2;
        }

        emit PerformanceVerified(bountyId, tvlImpact, mauImpact);
    }

    /**
     * @dev Get bounty details
     */
    function getBounty(uint256 bountyId) 
        external 
        view 
        returns (
            string memory title,
            string memory description,
            BountyCategory category,
            BountyStatus status,
            address creator,
            address assignee,
            uint256 totalReward,
            uint256 createdAt,
            uint256 deadline
        ) 
    {
        Bounty storage bounty = bounties[bountyId];
        return (
            bounty.title,
            bounty.description,
            bounty.category,
            bounty.status,
            bounty.creator,
            bounty.assignee,
            bounty.totalReward,
            bounty.createdAt,
            bounty.deadline
        );
    }

    /**
     * @dev Get bounty milestones
     */
    function getBountyMilestones(uint256 bountyId) 
        external 
        view 
        returns (Milestone[] memory) 
    {
        return bounties[bountyId].milestones;
    }

    /**
     * @dev Get developer's bounties
     */
    function getDeveloperBounties(address developer) 
        external 
        view 
        returns (uint256[] memory) 
    {
        return developerBounties[developer];
    }

    /**
     * @dev Get total number of bounties
     */
    function getTotalBounties() external view returns (uint256) {
        return _bountyIds.current();
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

    function updateMetricsOracle(address newOracle) external onlyOwner {
        metricsOracle = newOracle;
    }

    function emergencyWithdraw(uint256 amount) external onlyOwner {
        coffeeToken.safeTransfer(owner(), amount);
    }
}
