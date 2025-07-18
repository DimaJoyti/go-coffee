-- Drop Developer DAO tables in reverse order to handle foreign key constraints

-- Drop indexes first
DROP INDEX IF EXISTS idx_dao_votes_voter;
DROP INDEX IF EXISTS idx_dao_votes_proposal;
DROP INDEX IF EXISTS idx_performance_metrics_solution_date;
DROP INDEX IF EXISTS idx_solutions_score;
DROP INDEX IF EXISTS idx_solutions_developer;
DROP INDEX IF EXISTS idx_solutions_category;
DROP INDEX IF EXISTS idx_solutions_status;
DROP INDEX IF EXISTS idx_bounties_deadline;
DROP INDEX IF EXISTS idx_bounties_assignee;
DROP INDEX IF EXISTS idx_bounties_creator;
DROP INDEX IF EXISTS idx_bounties_category;
DROP INDEX IF EXISTS idx_bounties_status;
DROP INDEX IF EXISTS idx_developer_profiles_active;
DROP INDEX IF EXISTS idx_developer_profiles_reputation;
DROP INDEX IF EXISTS idx_developer_profiles_address;
DROP INDEX IF EXISTS idx_dao_proposals_created_at;
DROP INDEX IF EXISTS idx_dao_proposals_proposer;
DROP INDEX IF EXISTS idx_dao_proposals_status;

-- Drop tables in reverse order
DROP TABLE IF EXISTS dao_votes;
DROP TABLE IF EXISTS performance_metrics;
DROP TABLE IF EXISTS revenue_shares;
DROP TABLE IF EXISTS solution_compatibility;
DROP TABLE IF EXISTS solution_maintainers;
DROP TABLE IF EXISTS solutions;
DROP TABLE IF EXISTS bounty_applications;
DROP TABLE IF EXISTS bounty_milestones;
DROP TABLE IF EXISTS bounties;
DROP TABLE IF EXISTS developer_profiles;
DROP TABLE IF EXISTS dao_proposals;
