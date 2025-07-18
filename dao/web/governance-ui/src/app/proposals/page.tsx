'use client';

import React, { useState } from 'react';
import { format } from 'date-fns';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Button } from '../@/shared/components/ui/button';
import { Input } from '../@/shared/components/ui/input';
import { Badge } from '../@/shared/components/ui/badge';
import { 
  Search, 
  Plus,
  Vote,
  Clock,
  Users,
  CheckCircle,
  XCircle,
  AlertCircle,
  Filter,
} from 'lucide-react';

interface Proposal {
  id: number;
  title: string;
  description: string;
  status: 'active' | 'passed' | 'rejected' | 'pending' | 'executed';
  votesFor: number;
  votesAgainst: number;
  totalVotes: number;
  quorum: number;
  startDate: Date;
  endDate: Date;
  proposer: string;
  category: 'treasury' | 'governance' | 'technical' | 'community';
  executionDelay?: number;
}

// Mock proposals data
const proposals: Proposal[] = [
  {
    id: 1,
    title: 'Increase Developer Reward Pool by 25%',
    description: 'Proposal to increase the monthly developer reward pool from $100K to $125K to attract more high-quality contributors and improve platform growth.',
    status: 'active',
    votesFor: 1250,
    votesAgainst: 340,
    totalVotes: 1590,
    quorum: 1000,
    startDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
    endDate: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000),
    proposer: '0x1234...5678',
    category: 'treasury',
  },
  {
    id: 2,
    title: 'Implement Quality Score Weighting in Bounty Assignment',
    description: 'Adjust bounty assignment algorithm to prioritize developers with higher quality scores and community reputation to ensure better outcomes.',
    status: 'active',
    votesFor: 890,
    votesAgainst: 120,
    totalVotes: 1010,
    quorum: 800,
    startDate: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
    endDate: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000),
    proposer: '0xabcd...efgh',
    category: 'governance',
  },
  {
    id: 3,
    title: 'Add Support for Polygon Network',
    description: 'Expand platform support to include Polygon network for lower transaction costs and faster processing, enabling broader developer participation.',
    status: 'passed',
    votesFor: 2100,
    votesAgainst: 450,
    totalVotes: 2550,
    quorum: 1500,
    startDate: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000),
    endDate: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000),
    proposer: '0x9876...5432',
    category: 'technical',
    executionDelay: 7,
  },
  {
    id: 4,
    title: 'Community Grant Program Launch',
    description: 'Establish a $50K quarterly grant program for community-driven initiatives, educational content, and ecosystem development projects.',
    status: 'executed',
    votesFor: 1800,
    votesAgainst: 200,
    totalVotes: 2000,
    quorum: 1200,
    startDate: new Date(Date.now() - 20 * 24 * 60 * 60 * 1000),
    endDate: new Date(Date.now() - 13 * 24 * 60 * 60 * 1000),
    proposer: '0xdef0...1234',
    category: 'community',
  },
  {
    id: 5,
    title: 'Update Governance Token Distribution',
    description: 'Modify token distribution to allocate 5% more to active developers and 3% more to community contributors based on performance metrics.',
    status: 'rejected',
    votesFor: 650,
    votesAgainst: 1200,
    totalVotes: 1850,
    quorum: 1000,
    startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
    endDate: new Date(Date.now() - 23 * 24 * 60 * 60 * 1000),
    proposer: '0x5555...7777',
    category: 'governance',
  },
];

const categoryOptions = [
  { value: 'treasury', label: 'Treasury' },
  { value: 'governance', label: 'Governance' },
  { value: 'technical', label: 'Technical' },
  { value: 'community', label: 'Community' },
];

const statusOptions = [
  { value: 'active', label: 'Active' },
  { value: 'passed', label: 'Passed' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'executed', label: 'Executed' },
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
    case 'executed':
      return <CheckCircle className="w-4 h-4 text-purple-600" />;
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
    case 'executed':
      return <Badge variant="secondary">Executed</Badge>;
  }
}

function getCategoryBadge(category: Proposal['category']) {
  const variants = {
    treasury: 'default',
    governance: 'secondary',
    technical: 'outline',
    community: 'info',
  } as const;
  
  return <Badge variant={variants[category]}>{categoryOptions.find(c => c.value === category)?.label}</Badge>;
}

function getVotePercentage(votesFor: number, totalVotes: number) {
  return totalVotes > 0 ? (votesFor / totalVotes) * 100 : 0;
}

function getQuorumPercentage(totalVotes: number, quorum: number) {
  return (totalVotes / quorum) * 100;
}

export default function ProposalsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [selectedStatus, setSelectedStatus] = useState<string>('');

  const filteredProposals = proposals.filter(proposal => {
    const matchesSearch = proposal.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         proposal.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = !selectedCategory || proposal.category === selectedCategory;
    const matchesStatus = !selectedStatus || proposal.status === selectedStatus;
    
    return matchesSearch && matchesCategory && matchesStatus;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Proposals</h1>
          <p className="text-muted-foreground">
            View and vote on governance proposals that shape the future of Developer DAO.
          </p>
        </div>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          Create Proposal
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Proposals</CardTitle>
            <Vote className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{proposals.length}</div>
            <p className="text-xs text-muted-foreground">
              +3 this month
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Proposals</CardTitle>
            <Vote className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {proposals.filter(p => p.status === 'active').length}
            </div>
            <p className="text-xs text-muted-foreground">
              Currently voting
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Passed Proposals</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {proposals.filter(p => p.status === 'passed' || p.status === 'executed').length}
            </div>
            <p className="text-xs text-muted-foreground">
              Successfully approved
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Participation Rate</CardTitle>
            <Users className="h-4 w-4 text-purple-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">78.5%</div>
            <p className="text-xs text-muted-foreground">
              Average voter turnout
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col gap-4 md:flex-row md:items-center">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
              <Input
                placeholder="Search proposals..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Category Filter */}
            <div className="flex gap-2 flex-wrap">
              <Button
                variant={selectedCategory === '' ? "default" : "outline"}
                size="sm"
                onClick={() => setSelectedCategory('')}
              >
                All Categories
              </Button>
              {categoryOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={selectedCategory === option.value ? "default" : "outline"}
                  size="sm"
                  onClick={() => setSelectedCategory(option.value)}
                >
                  {option.label}
                </Button>
              ))}
            </div>

            {/* Status Filter */}
            <div className="flex gap-2 flex-wrap">
              <Button
                variant={selectedStatus === '' ? "default" : "outline"}
                size="sm"
                onClick={() => setSelectedStatus('')}
              >
                All Status
              </Button>
              {statusOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={selectedStatus === option.value ? "default" : "outline"}
                  size="sm"
                  onClick={() => setSelectedStatus(option.value)}
                >
                  {option.label}
                </Button>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Proposals List */}
      <div className="space-y-4">
        {filteredProposals.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <div className="text-center text-muted-foreground">
                No proposals found matching your criteria.
              </div>
            </CardContent>
          </Card>
        ) : (
          filteredProposals.map((proposal) => (
            <Card key={proposal.id} className="hover:shadow-lg transition-shadow">
              <CardContent className="pt-6">
                <div className="flex items-start justify-between gap-4 mb-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      {getStatusIcon(proposal.status)}
                      <h3 className="text-lg font-semibold">{proposal.title}</h3>
                      {getStatusBadge(proposal.status)}
                      {getCategoryBadge(proposal.category)}
                    </div>
                    <p className="text-muted-foreground mb-3">
                      {proposal.description}
                    </p>
                  </div>
                </div>

                {/* Voting Progress */}
                {(proposal.status === 'active' || proposal.status === 'passed' || proposal.status === 'rejected') && (
                  <div className="space-y-3 mb-4">
                    <div className="flex justify-between text-sm">
                      <span className="text-green-600">For: {proposal.votesFor.toLocaleString()}</span>
                      <span className="text-red-600">Against: {proposal.votesAgainst.toLocaleString()}</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-3">
                      <div
                        className="bg-green-500 h-3 rounded-full relative"
                        style={{ width: `${getVotePercentage(proposal.votesFor, proposal.totalVotes)}%` }}
                      >
                        <div className="absolute inset-0 bg-red-500 rounded-full" 
                             style={{ 
                               left: `${getVotePercentage(proposal.votesFor, proposal.totalVotes)}%`,
                               width: `${getVotePercentage(proposal.votesAgainst, proposal.totalVotes)}%`
                             }} 
                        />
                      </div>
                    </div>
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>{getVotePercentage(proposal.votesFor, proposal.totalVotes).toFixed(1)}% in favor</span>
                      <span>Quorum: {getQuorumPercentage(proposal.totalVotes, proposal.quorum).toFixed(0)}%</span>
                    </div>
                  </div>
                )}

                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <Users className="w-4 h-4" />
                      <span>{proposal.totalVotes.toLocaleString()} votes</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Clock className="w-4 h-4" />
                      <span>
                        {proposal.status === 'active' 
                          ? `Ends ${format(proposal.endDate, 'MMM dd, yyyy')}`
                          : `Ended ${format(proposal.endDate, 'MMM dd, yyyy')}`
                        }
                      </span>
                    </div>
                    <span>by {proposal.proposer}</span>
                  </div>

                  <div className="flex gap-2">
                    <Button variant="outline" size="sm">
                      View Details
                    </Button>
                    {proposal.status === 'active' && (
                      <>
                        <Button variant="outline" size="sm">
                          Vote Against
                        </Button>
                        <Button size="sm">
                          Vote For
                        </Button>
                      </>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
