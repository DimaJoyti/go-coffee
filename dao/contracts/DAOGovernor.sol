// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/governance/Governor.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorCountingSimple.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotes.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotesQuorumFraction.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorTimelockControl.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title DAOGovernor
 * @dev Governor contract for the Developer DAO with Coffee Token voting
 */
contract DAOGovernor is 
    Governor,
    GovernorSettings,
    GovernorCountingSimple,
    GovernorVotes,
    GovernorVotesQuorumFraction,
    GovernorTimelockControl,
    Ownable
{
    // Proposal categories for better organization
    enum ProposalCategory {
        GENERAL,
        BOUNTY,
        TREASURY,
        TECHNICAL,
        PARTNERSHIP
    }

    // Enhanced proposal structure
    struct ProposalInfo {
        uint256 id;
        ProposalCategory category;
        string title;
        string description;
        address proposer;
        uint256 createdAt;
        uint256 executionDeadline;
        bool executed;
        mapping(address => bool) hasVoted;
    }

    // State variables
    mapping(uint256 => ProposalInfo) public proposalInfo;
    mapping(address => uint256) public proposalCount;
    uint256 public totalProposals;

    // Minimum Coffee Token balance required to create proposals
    uint256 public constant MIN_PROPOSAL_THRESHOLD = 10000 * 10**18; // 10,000 COFFEE tokens

    // Events
    event ProposalCreatedWithInfo(
        uint256 indexed proposalId,
        address indexed proposer,
        ProposalCategory category,
        string title,
        string description
    );

    event ProposalExecuted(uint256 indexed proposalId, bool success);

    constructor(
        IVotes _token,
        TimelockController _timelock,
        uint256 _quorumPercentage
    )
        Governor("Developer DAO Governor")
        GovernorSettings(
            7200, /* 1 day voting delay */
            50400, /* 1 week voting period */
            MIN_PROPOSAL_THRESHOLD /* proposal threshold */
        )
        GovernorVotes(_token)
        GovernorVotesQuorumFraction(_quorumPercentage)
        GovernorTimelockControl(_timelock)
    {}

    /**
     * @dev Create a proposal with enhanced metadata
     */
    function proposeWithInfo(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description,
        ProposalCategory category,
        string memory title,
        uint256 executionDeadline
    ) public returns (uint256) {
        // Check if proposer has minimum token balance
        require(
            getVotes(msg.sender, block.number - 1) >= MIN_PROPOSAL_THRESHOLD,
            "Insufficient tokens to create proposal"
        );

        // Create the proposal
        uint256 proposalId = propose(targets, values, calldatas, description);

        // Store additional proposal information
        ProposalInfo storage info = proposalInfo[proposalId];
        info.id = proposalId;
        info.category = category;
        info.title = title;
        info.description = description;
        info.proposer = msg.sender;
        info.createdAt = block.timestamp;
        info.executionDeadline = executionDeadline;
        info.executed = false;

        // Update counters
        proposalCount[msg.sender]++;
        totalProposals++;

        emit ProposalCreatedWithInfo(proposalId, msg.sender, category, title, description);

        return proposalId;
    }

    /**
     * @dev Get proposal information
     */
    function getProposalInfo(uint256 proposalId) 
        external 
        view 
        returns (
            ProposalCategory category,
            string memory title,
            string memory description,
            address proposer,
            uint256 createdAt,
            uint256 executionDeadline,
            bool executed
        ) 
    {
        ProposalInfo storage info = proposalInfo[proposalId];
        return (
            info.category,
            info.title,
            info.description,
            info.proposer,
            info.createdAt,
            info.executionDeadline,
            info.executed
        );
    }

    /**
     * @dev Check if an address has voted on a proposal
     */
    function hasVoted(uint256 proposalId, address account) external view returns (bool) {
        return proposalInfo[proposalId].hasVoted[account];
    }

    /**
     * @dev Override castVote to track voting
     */
    function castVote(uint256 proposalId, uint8 support) 
        public 
        override 
        returns (uint256) 
    {
        proposalInfo[proposalId].hasVoted[msg.sender] = true;
        return super.castVote(proposalId, support);
    }

    /**
     * @dev Override castVoteWithReason to track voting
     */
    function castVoteWithReason(
        uint256 proposalId,
        uint8 support,
        string calldata reason
    ) public override returns (uint256) {
        proposalInfo[proposalId].hasVoted[msg.sender] = true;
        return super.castVoteWithReason(proposalId, support, reason);
    }

    /**
     * @dev Override execute to mark proposal as executed
     */
    function execute(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) public payable override(Governor, GovernorTimelockControl) returns (uint256) {
        uint256 proposalId = hashProposal(targets, values, calldatas, descriptionHash);
        
        // Execute the proposal
        uint256 result = super.execute(targets, values, calldatas, descriptionHash);
        
        // Mark as executed
        proposalInfo[proposalId].executed = true;
        
        emit ProposalExecuted(proposalId, true);
        
        return result;
    }

    // Required overrides
    function votingDelay() public view override(IGovernor, GovernorSettings) returns (uint256) {
        return super.votingDelay();
    }

    function votingPeriod() public view override(IGovernor, GovernorSettings) returns (uint256) {
        return super.votingPeriod();
    }

    function quorum(uint256 blockNumber)
        public
        view
        override(IGovernor, GovernorVotesQuorumFraction)
        returns (uint256)
    {
        return super.quorum(blockNumber);
    }

    function proposalThreshold()
        public
        view
        override(Governor, GovernorSettings)
        returns (uint256)
    {
        return super.proposalThreshold();
    }

    function state(uint256 proposalId)
        public
        view
        override(Governor, GovernorTimelockControl)
        returns (ProposalState)
    {
        return super.state(proposalId);
    }

    function _execute(
        uint256 proposalId,
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) internal override(Governor, GovernorTimelockControl) {
        super._execute(proposalId, targets, values, calldatas, descriptionHash);
    }

    function _cancel(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        bytes32 descriptionHash
    ) internal override(Governor, GovernorTimelockControl) returns (uint256) {
        return super._cancel(targets, values, calldatas, descriptionHash);
    }

    function _executor()
        internal
        view
        override(Governor, GovernorTimelockControl)
        returns (address)
    {
        return super._executor();
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(Governor, GovernorTimelockControl)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }
}
