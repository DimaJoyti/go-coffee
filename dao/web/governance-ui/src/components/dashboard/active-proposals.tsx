'use client';

import React from 'react';
import { format } from 'date-fns';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Badge } from '../@/shared/components/ui/badge';
import { Button } from '../@/shared/components/ui/button';
import { 
  Vote, 
  Clock, 
  Users, 
  CheckCircle,
  XCircle,
  AlertCircle,
} from 'lucide-react';

interface Proposal {
  id: number;
  title: string;
  description: string;
  status: 'active' | 'passed' | 'rejected' | 'pending';
  votesFor: number;
  votesAgainst: number;
  totalVotes: number;
  endDate: Date;
  proposer: string;
  category: string;
}

// Mock proposal data
const proposals: Proposal[] = [
  {
    id: 1,
    title: 'Increase Developer Reward Pool by 25%',
    description: 'Proposal to increase the monthly developer reward pool from $100K to $125K to attract more high-quality contributors.',
    status: 'active',
    votesFor: 1250,
    votesAgainst: 340,
    totalVotes: 1590,
    endDate: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000), // 3 days from now
    proposer: '0x1234...5678',
    category: 'Treasury',
  },
  {
    id: 2,
    title: 'Implement Quality Score Weighting',
    description: 'Adjust bounty assignments to prioritize developers with higher quality scores and community reputation.',
    status: 'active',
    votesFor: 890,
    votesAgainst: 120,
    totalVotes: 1010,
    endDate: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000), // 5 days from now
    proposer: '0xabcd...efgh',
    category: 'Governance',
  },
  {
    id: 3,
    title: 'Add Support for Polygon Network',
    description: 'Expand platform support to include Polygon network for lower transaction costs and faster processing.',
    status: 'active',
    votesFor: 2100,
    votesAgainst: 450,
    totalVotes: 2550,
    endDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000), // 7 days from now
    proposer: '0x9876...5432',
    category: 'Technical',
  },
  {
    id: 4,
    title: 'Community Grant Program Launch',
    description: 'Establish a $50K quarterly grant program for community-driven initiatives and educational content.',
    status: 'passed',
    votesFor: 1800,
    votesAgainst: 200,
    totalVotes: 2000,
    endDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000), // 2 days ago
    proposer: '0xdef0...1234',
    category: 'Community',
  },
];

function getStatusIcon(status: Proposal['status']) {
  switch (status) {
    case 'active':
      return <Vote className="w-4 h-4 text-blue-600" />;
    case 'passed':
      return <CheckCircle className="w-4 h-4 text-green-600" />;
    case 'rejected':
      return <XCircle className="w-4 h-4 text-red-600" />;
    case 'pending':
      return <AlertCircle className="w-4 h-4 text-yellow-600" />;
  }
}

function getStatusBadge(status: Proposal['status']) {
  switch (status) {
    case 'active':
      return <Badge variant="info">Active</Badge>;
    case 'passed':
      return <Badge variant="success">Passed</Badge>;
    case 'rejected':
      return <Badge variant="destructive">Rejected</Badge>;
    case 'pending':
      return <Badge variant="warning">Pending</Badge>;
  }
}

function getVotePercentage(votesFor: number, totalVotes: number) {
  return totalVotes > 0 ? (votesFor / totalVotes) * 100 : 0;
}

export function ActiveProposals() {
  const activeProposals = proposals.filter(p => p.status === 'active');
  const recentProposals = proposals.slice(0, 4);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Vote className="w-5 h-5" />
          Active Proposals
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {recentProposals.map((proposal) => (
            <div
              key={proposal.id}
              className="p-4 rounded-lg border border-border hover:bg-accent/50 transition-colors"
            >
              <div className="flex items-start justify-between gap-3 mb-3">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    {getStatusIcon(proposal.status)}
                    <h4 className="font-medium line-clamp-1">{proposal.title}</h4>
                  </div>
                  <p className="text-sm text-muted-foreground line-clamp-2">
                    {proposal.description}
                  </p>
                </div>
                {getStatusBadge(proposal.status)}
              </div>

              <div className="flex items-center gap-4 text-xs text-muted-foreground mb-3">
                <div className="flex items-center gap-1">
                  <Users className="w-3 h-3" />
                  <span>{proposal.totalVotes.toLocaleString()} votes</span>
                </div>
                <div className="flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  <span>
                    {proposal.status === 'active' 
                      ? `Ends ${format(proposal.endDate, 'MMM dd')}`
                      : `Ended ${format(proposal.endDate, 'MMM dd')}`
                    }
                  </span>
                </div>
                <Badge variant="outline" className="text-xs">
                  {proposal.category}
                </Badge>
              </div>

              {/* Voting Progress */}
              <div className="space-y-2 mb-3">
                <div className="flex justify-between text-sm">
                  <span className="text-green-600">For: {proposal.votesFor.toLocaleString()}</span>
                  <span className="text-red-600">Against: {proposal.votesAgainst.toLocaleString()}</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-green-500 h-2 rounded-full"
                    style={{ width: `${getVotePercentage(proposal.votesFor, proposal.totalVotes)}%` }}
                  />
                </div>
                <div className="text-xs text-muted-foreground">
                  {getVotePercentage(proposal.votesFor, proposal.totalVotes).toFixed(1)}% in favor
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div className="text-xs text-muted-foreground">
                  Proposed by {proposal.proposer}
                </div>
                {proposal.status === 'active' && (
                  <div className="flex gap-2">
                    <Button size="sm" variant="outline">
                      Against
                    </Button>
                    <Button size="sm">
                      Vote For
                    </Button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        <div className="mt-4 text-center">
          <Button variant="outline">
            View All Proposals
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
