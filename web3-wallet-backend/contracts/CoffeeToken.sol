// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

/**
 * @title CoffeeToken
 * @dev ERC20 token for the Coffee ecosystem with staking rewards
 */
contract CoffeeToken is ERC20, ERC20Burnable, Pausable, Ownable, ReentrancyGuard {
    
    // Staking structure
    struct StakeInfo {
        uint256 amount;
        uint256 startTime;
        uint256 lastClaimTime;
        uint256 totalRewards;
        bool active;
    }
    
    // Constants
    uint256 public constant TOTAL_SUPPLY = 1_000_000_000 * 10**18; // 1 billion tokens
    uint256 public constant REWARDS_APY = 12; // 12% APY
    uint256 public constant MIN_STAKE_AMOUNT = 100 * 10**18; // 100 COFFEE tokens
    uint256 public constant SECONDS_PER_YEAR = 365 * 24 * 60 * 60;
    
    // State variables
    mapping(address => StakeInfo[]) public userStakes;
    mapping(address => uint256) public totalStaked;
    uint256 public totalStakedSupply;
    uint256 public rewardsPool;
    
    // Events
    event Staked(address indexed user, uint256 amount, uint256 stakeIndex);
    event Unstaked(address indexed user, uint256 amount, uint256 rewards, uint256 stakeIndex);
    event RewardsClaimed(address indexed user, uint256 rewards, uint256 stakeIndex);
    event RewardsPoolUpdated(uint256 newAmount);
    
    constructor() ERC20("Coffee Token", "COFFEE") {
        // Mint total supply to contract deployer
        _mint(msg.sender, TOTAL_SUPPLY);
        
        // Initialize rewards pool (5% of total supply)
        rewardsPool = TOTAL_SUPPLY * 5 / 100;
    }
    
    /**
     * @dev Pause the contract (only owner)
     */
    function pause() public onlyOwner {
        _pause();
    }
    
    /**
     * @dev Unpause the contract (only owner)
     */
    function unpause() public onlyOwner {
        _unpause();
    }
    
    /**
     * @dev Stake tokens for rewards
     * @param amount Amount of tokens to stake
     */
    function stake(uint256 amount) external nonReentrant whenNotPaused {
        require(amount >= MIN_STAKE_AMOUNT, "Amount below minimum stake");
        require(balanceOf(msg.sender) >= amount, "Insufficient balance");
        
        // Transfer tokens to contract
        _transfer(msg.sender, address(this), amount);
        
        // Create new stake
        StakeInfo memory newStake = StakeInfo({
            amount: amount,
            startTime: block.timestamp,
            lastClaimTime: block.timestamp,
            totalRewards: 0,
            active: true
        });
        
        userStakes[msg.sender].push(newStake);
        totalStaked[msg.sender] += amount;
        totalStakedSupply += amount;
        
        emit Staked(msg.sender, amount, userStakes[msg.sender].length - 1);
    }
    
    /**
     * @dev Unstake tokens and claim rewards
     * @param stakeIndex Index of the stake to unstake
     */
    function unstake(uint256 stakeIndex) external nonReentrant whenNotPaused {
        require(stakeIndex < userStakes[msg.sender].length, "Invalid stake index");
        
        StakeInfo storage stakeInfo = userStakes[msg.sender][stakeIndex];
        require(stakeInfo.active, "Stake not active");
        
        uint256 stakedAmount = stakeInfo.amount;
        uint256 pendingRewards = calculatePendingRewards(msg.sender, stakeIndex);
        
        // Update state
        stakeInfo.active = false;
        totalStaked[msg.sender] -= stakedAmount;
        totalStakedSupply -= stakedAmount;
        
        // Transfer staked tokens back to user
        _transfer(address(this), msg.sender, stakedAmount);
        
        // Transfer rewards if available
        if (pendingRewards > 0 && rewardsPool >= pendingRewards) {
            rewardsPool -= pendingRewards;
            stakeInfo.totalRewards += pendingRewards;
            _mint(msg.sender, pendingRewards);
        }
        
        emit Unstaked(msg.sender, stakedAmount, pendingRewards, stakeIndex);
    }
    
    /**
     * @dev Claim rewards without unstaking
     * @param stakeIndex Index of the stake to claim rewards for
     */
    function claimRewards(uint256 stakeIndex) external nonReentrant whenNotPaused {
        require(stakeIndex < userStakes[msg.sender].length, "Invalid stake index");
        
        StakeInfo storage stakeInfo = userStakes[msg.sender][stakeIndex];
        require(stakeInfo.active, "Stake not active");
        
        uint256 pendingRewards = calculatePendingRewards(msg.sender, stakeIndex);
        require(pendingRewards > 0, "No rewards to claim");
        require(rewardsPool >= pendingRewards, "Insufficient rewards pool");
        
        // Update state
        rewardsPool -= pendingRewards;
        stakeInfo.lastClaimTime = block.timestamp;
        stakeInfo.totalRewards += pendingRewards;
        
        // Mint rewards to user
        _mint(msg.sender, pendingRewards);
        
        emit RewardsClaimed(msg.sender, pendingRewards, stakeIndex);
    }
    
    /**
     * @dev Calculate pending rewards for a stake
     * @param user Address of the user
     * @param stakeIndex Index of the stake
     * @return Pending rewards amount
     */
    function calculatePendingRewards(address user, uint256 stakeIndex) public view returns (uint256) {
        require(stakeIndex < userStakes[user].length, "Invalid stake index");
        
        StakeInfo memory stakeInfo = userStakes[user][stakeIndex];
        if (!stakeInfo.active) {
            return 0;
        }
        
        uint256 timeSinceLastClaim = block.timestamp - stakeInfo.lastClaimTime;
        uint256 yearlyRewards = (stakeInfo.amount * REWARDS_APY) / 100;
        uint256 pendingRewards = (yearlyRewards * timeSinceLastClaim) / SECONDS_PER_YEAR;
        
        return pendingRewards;
    }
    
    /**
     * @dev Get user's stake information
     * @param user Address of the user
     * @return Array of stake information
     */
    function getUserStakes(address user) external view returns (StakeInfo[] memory) {
        return userStakes[user];
    }
    
    /**
     * @dev Get user's active stakes count
     * @param user Address of the user
     * @return Number of active stakes
     */
    function getActiveStakesCount(address user) external view returns (uint256) {
        uint256 count = 0;
        for (uint256 i = 0; i < userStakes[user].length; i++) {
            if (userStakes[user][i].active) {
                count++;
            }
        }
        return count;
    }
    
    /**
     * @dev Get total pending rewards for a user
     * @param user Address of the user
     * @return Total pending rewards
     */
    function getTotalPendingRewards(address user) external view returns (uint256) {
        uint256 totalPending = 0;
        for (uint256 i = 0; i < userStakes[user].length; i++) {
            if (userStakes[user][i].active) {
                totalPending += calculatePendingRewards(user, i);
            }
        }
        return totalPending;
    }
    
    /**
     * @dev Add tokens to rewards pool (only owner)
     * @param amount Amount to add to rewards pool
     */
    function addToRewardsPool(uint256 amount) external onlyOwner {
        require(balanceOf(msg.sender) >= amount, "Insufficient balance");
        _transfer(msg.sender, address(this), amount);
        rewardsPool += amount;
        emit RewardsPoolUpdated(rewardsPool);
    }
    
    /**
     * @dev Emergency withdraw from rewards pool (only owner)
     * @param amount Amount to withdraw
     */
    function emergencyWithdrawRewardsPool(uint256 amount) external onlyOwner {
        require(amount <= rewardsPool, "Amount exceeds rewards pool");
        rewardsPool -= amount;
        _transfer(address(this), msg.sender, amount);
        emit RewardsPoolUpdated(rewardsPool);
    }
    
    /**
     * @dev Override transfer to add pause functionality
     */
    function _beforeTokenTransfer(address from, address to, uint256 amount)
        internal
        whenNotPaused
        override
    {
        super._beforeTokenTransfer(from, to, amount);
    }
    
    /**
     * @dev Get contract statistics
     * @return totalSupply Total token supply
     * @return totalStaked Total tokens staked
     * @return rewardsPoolBalance Current rewards pool balance
     * @return stakersCount Number of unique stakers
     */
    function getContractStats() external view returns (
        uint256 totalSupply_,
        uint256 totalStaked_,
        uint256 rewardsPoolBalance,
        uint256 stakersCount
    ) {
        totalSupply_ = totalSupply();
        totalStaked_ = totalStakedSupply;
        rewardsPoolBalance = rewardsPool;
        
        // Note: stakersCount would require additional tracking for efficiency
        // This is a simplified version
        stakersCount = 0; // Placeholder
    }
}
