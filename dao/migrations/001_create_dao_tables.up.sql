-- Developer DAO Database Schema
-- This migration creates the core tables for the Developer DAO platform

-- DAO Proposals table
CREATE TABLE IF NOT EXISTS dao_proposals (
    id SERIAL PRIMARY KEY,
    proposal_id VARCHAR(66) UNIQUE NOT NULL, -- Ethereum transaction hash
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category INTEGER NOT NULL, -- 0: GENERAL, 1: BOUNTY, 2: TREASURY, 3: TECHNICAL, 4: PARTNERSHIP
    proposer_address VARCHAR(42) NOT NULL,
    status INTEGER NOT NULL DEFAULT 0, -- 0: PENDING, 1: ACTIVE, 2: CANCELED, 3: DEFEATED, 4: SUCCEEDED, 5: QUEUED, 6: EXPIRED, 7: EXECUTED
    votes_for DECIMAL(78,0) DEFAULT 0,
    votes_against DECIMAL(78,0) DEFAULT 0,
    votes_abstain DECIMAL(78,0) DEFAULT 0,
    quorum_reached BOOLEAN DEFAULT FALSE,
    execution_deadline TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    executed_at TIMESTAMP,
    transaction_hash VARCHAR(66),
    block_number BIGINT,
    gas_used BIGINT
);

-- Developer profiles table
CREATE TABLE IF NOT EXISTS developer_profiles (
    id SERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    username VARCHAR(50),
    email VARCHAR(255),
    github_username VARCHAR(100),
    discord_username VARCHAR(100),
    bio TEXT,
    skills TEXT[], -- Array of skills
    reputation_score INTEGER DEFAULT 0,
    total_bounties_completed INTEGER DEFAULT 0,
    total_earnings DECIMAL(78,0) DEFAULT 0,
    tvl_contributed DECIMAL(78,0) DEFAULT 0,
    mau_contributed INTEGER DEFAULT 0,
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bounties table
CREATE TABLE IF NOT EXISTS bounties (
    id SERIAL PRIMARY KEY,
    bounty_id BIGINT UNIQUE NOT NULL, -- Contract bounty ID
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category INTEGER NOT NULL, -- 0: TVL_GROWTH, 1: MAU_EXPANSION, 2: INNOVATION, 3: MAINTENANCE, 4: SECURITY, 5: INTEGRATION
    status INTEGER NOT NULL DEFAULT 0, -- 0: OPEN, 1: ASSIGNED, 2: IN_PROGRESS, 3: SUBMITTED, 4: COMPLETED, 5: CANCELLED
    creator_address VARCHAR(42) NOT NULL,
    assignee_address VARCHAR(42),
    total_reward DECIMAL(78,0) NOT NULL,
    deadline TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_at TIMESTAMP,
    completed_at TIMESTAMP,
    transaction_hash VARCHAR(66),
    block_number BIGINT,
    tvl_impact DECIMAL(78,0) DEFAULT 0,
    mau_impact INTEGER DEFAULT 0,
    performance_verified BOOLEAN DEFAULT FALSE
);

-- Bounty milestones table
CREATE TABLE IF NOT EXISTS bounty_milestones (
    id SERIAL PRIMARY KEY,
    bounty_id BIGINT NOT NULL REFERENCES bounties(bounty_id),
    milestone_index INTEGER NOT NULL,
    description TEXT NOT NULL,
    reward DECIMAL(78,0) NOT NULL,
    deadline TIMESTAMP NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    paid BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP,
    transaction_hash VARCHAR(66),
    UNIQUE(bounty_id, milestone_index)
);

-- Bounty applications table
CREATE TABLE IF NOT EXISTS bounty_applications (
    id SERIAL PRIMARY KEY,
    bounty_id BIGINT NOT NULL REFERENCES bounties(bounty_id),
    applicant_address VARCHAR(42) NOT NULL,
    application_message TEXT,
    proposed_timeline INTEGER, -- Days to complete
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status INTEGER DEFAULT 0, -- 0: PENDING, 1: ACCEPTED, 2: REJECTED
    UNIQUE(bounty_id, applicant_address)
);

-- Solutions registry table
CREATE TABLE IF NOT EXISTS solutions (
    id SERIAL PRIMARY KEY,
    solution_id BIGINT UNIQUE NOT NULL, -- Contract solution ID
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    repository_url VARCHAR(500),
    documentation_url VARCHAR(500),
    category INTEGER NOT NULL, -- 0: DEX_INTEGRATION, 1: LENDING_PROTOCOL, etc.
    status INTEGER NOT NULL DEFAULT 0, -- 0: SUBMITTED, 1: UNDER_REVIEW, 2: APPROVED, 3: REJECTED, 4: DEPRECATED
    developer_address VARCHAR(42) NOT NULL,
    version VARCHAR(50) NOT NULL,
    is_third_party BOOLEAN DEFAULT FALSE,
    license_type INTEGER DEFAULT 0, -- 0: MIT, 1: GPL, 2: Commercial
    security_score INTEGER DEFAULT 0,
    performance_score INTEGER DEFAULT 0,
    usability_score INTEGER DEFAULT 0,
    documentation_score INTEGER DEFAULT 0,
    overall_score INTEGER DEFAULT 0,
    has_security_audit BOOLEAN DEFAULT FALSE,
    audit_count INTEGER DEFAULT 0,
    deployment_count INTEGER DEFAULT 0,
    tvl_generated DECIMAL(78,0) DEFAULT 0,
    mau_generated INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,
    transaction_hash VARCHAR(66)
);

-- Solution maintainers table
CREATE TABLE IF NOT EXISTS solution_maintainers (
    id SERIAL PRIMARY KEY,
    solution_id BIGINT NOT NULL REFERENCES solutions(solution_id),
    maintainer_address VARCHAR(42) NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    added_by VARCHAR(42) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(solution_id, maintainer_address)
);

-- Solution compatibility table
CREATE TABLE IF NOT EXISTS solution_compatibility (
    id SERIAL PRIMARY KEY,
    solution_id BIGINT NOT NULL REFERENCES solutions(solution_id),
    compatible_solutions BIGINT[],
    incompatible_solutions BIGINT[],
    supported_networks TEXT[],
    required_dependencies TEXT[],
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Revenue sharing table
CREATE TABLE IF NOT EXISTS revenue_shares (
    id SERIAL PRIMARY KEY,
    solution_id BIGINT UNIQUE NOT NULL REFERENCES solutions(solution_id),
    developer_address VARCHAR(42) NOT NULL,
    tvl_contribution DECIMAL(78,0) DEFAULT 0,
    mau_contribution INTEGER DEFAULT 0,
    total_revenue DECIMAL(78,0) DEFAULT 0,
    developer_share DECIMAL(78,0) DEFAULT 0,
    last_distribution TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Performance metrics table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id SERIAL PRIMARY KEY,
    solution_id BIGINT NOT NULL REFERENCES solutions(solution_id),
    metric_date DATE NOT NULL,
    tvl_amount DECIMAL(78,0) DEFAULT 0,
    mau_count INTEGER DEFAULT 0,
    revenue_generated DECIMAL(78,0) DEFAULT 0,
    transaction_count INTEGER DEFAULT 0,
    unique_users INTEGER DEFAULT 0,
    gas_used BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(solution_id, metric_date)
);

-- DAO voting records table
CREATE TABLE IF NOT EXISTS dao_votes (
    id SERIAL PRIMARY KEY,
    proposal_id VARCHAR(66) NOT NULL REFERENCES dao_proposals(proposal_id),
    voter_address VARCHAR(42) NOT NULL,
    support INTEGER NOT NULL, -- 0: Against, 1: For, 2: Abstain
    voting_power DECIMAL(78,0) NOT NULL,
    reason TEXT,
    voted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    transaction_hash VARCHAR(66),
    block_number BIGINT,
    UNIQUE(proposal_id, voter_address)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_dao_proposals_status ON dao_proposals(status);
CREATE INDEX IF NOT EXISTS idx_dao_proposals_proposer ON dao_proposals(proposer_address);
CREATE INDEX IF NOT EXISTS idx_dao_proposals_created_at ON dao_proposals(created_at);

CREATE INDEX IF NOT EXISTS idx_developer_profiles_address ON developer_profiles(wallet_address);
CREATE INDEX IF NOT EXISTS idx_developer_profiles_reputation ON developer_profiles(reputation_score DESC);
CREATE INDEX IF NOT EXISTS idx_developer_profiles_active ON developer_profiles(is_active);

CREATE INDEX IF NOT EXISTS idx_bounties_status ON bounties(status);
CREATE INDEX IF NOT EXISTS idx_bounties_category ON bounties(category);
CREATE INDEX IF NOT EXISTS idx_bounties_creator ON bounties(creator_address);
CREATE INDEX IF NOT EXISTS idx_bounties_assignee ON bounties(assignee_address);
CREATE INDEX IF NOT EXISTS idx_bounties_deadline ON bounties(deadline);

CREATE INDEX IF NOT EXISTS idx_solutions_status ON solutions(status);
CREATE INDEX IF NOT EXISTS idx_solutions_category ON solutions(category);
CREATE INDEX IF NOT EXISTS idx_solutions_developer ON solutions(developer_address);
CREATE INDEX IF NOT EXISTS idx_solutions_score ON solutions(overall_score DESC);

CREATE INDEX IF NOT EXISTS idx_performance_metrics_solution_date ON performance_metrics(solution_id, metric_date);
CREATE INDEX IF NOT EXISTS idx_dao_votes_proposal ON dao_votes(proposal_id);
CREATE INDEX IF NOT EXISTS idx_dao_votes_voter ON dao_votes(voter_address);
