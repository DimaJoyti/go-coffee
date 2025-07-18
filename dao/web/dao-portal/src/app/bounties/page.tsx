'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Button } from '../@/shared/components/ui/button';
import { Input } from '../@/shared/components/ui/input';
import { Badge } from '../@/shared/components/ui/badge';
import { BountyCard } from '../@/shared/components/bounty-card';
import { 
  Search, 
  Filter, 
  SortAsc,
  Target,
  Plus,
} from 'lucide-react';
import { useBounties } from '../@/shared/hooks/api-hooks';
import { BountyCategory, BountyStatus, BountyFilters } from '../@/shared/types/api';

const categoryOptions = [
  { value: BountyCategory.TVL_GROWTH, label: 'TVL Growth' },
  { value: BountyCategory.MAU_EXPANSION, label: 'MAU Expansion' },
  { value: BountyCategory.INNOVATION, label: 'Innovation' },
  { value: BountyCategory.SECURITY, label: 'Security' },
  { value: BountyCategory.INFRASTRUCTURE, label: 'Infrastructure' },
  { value: BountyCategory.COMMUNITY, label: 'Community' },
];

const statusOptions = [
  { value: BountyStatus.OPEN, label: 'Open' },
  { value: BountyStatus.ASSIGNED, label: 'Assigned' },
  { value: BountyStatus.IN_PROGRESS, label: 'In Progress' },
  { value: BountyStatus.UNDER_REVIEW, label: 'Under Review' },
  { value: BountyStatus.COMPLETED, label: 'Completed' },
];

export default function BountiesPage() {
  const [filters, setFilters] = useState<BountyFilters>({
    limit: 12,
    offset: 0,
  });
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<BountyCategory | undefined>();
  const [selectedStatus, setSelectedStatus] = useState<BountyStatus | undefined>();

  const { data: bountiesResponse, isLoading, error } = useBounties(filters);

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setFilters(prev => ({ ...prev, search: query, offset: 0 }));
  };

  const handleCategoryFilter = (category: BountyCategory | undefined) => {
    setSelectedCategory(category);
    setFilters(prev => ({ ...prev, category, offset: 0 }));
  };

  const handleStatusFilter = (status: BountyStatus | undefined) => {
    setSelectedStatus(status);
    setFilters(prev => ({ ...prev, status, offset: 0 }));
  };

  const handleApplyForBounty = (bountyId: number) => {
    // TODO: Implement bounty application modal
    console.log('Apply for bounty:', bountyId);
  };

  const handleViewBounty = (bountyId: number) => {
    // TODO: Navigate to bounty details page
    console.log('View bounty:', bountyId);
  };

  const bounties = bountiesResponse?.data || [];
  const totalBounties = bountiesResponse?.pagination?.total || 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Bounties</h1>
          <p className="text-muted-foreground">
            Discover and apply for development bounties in the Developer DAO ecosystem.
          </p>
        </div>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          Create Bounty
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Bounties</CardTitle>
            <Target className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalBounties}</div>
            <p className="text-xs text-muted-foreground">
              +12% from last month
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Open Bounties</CardTitle>
            <Target className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {bounties.filter(b => b.status === BountyStatus.OPEN).length}
            </div>
            <p className="text-xs text-muted-foreground">
              Available to apply
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Rewards</CardTitle>
            <Target className="h-4 w-4 text-yellow-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$125K</div>
            <p className="text-xs text-muted-foreground">
              In active bounties
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">My Applications</CardTitle>
            <Target className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">5</div>
            <p className="text-xs text-muted-foreground">
              Pending review
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
                placeholder="Search bounties..."
                value={searchQuery}
                onChange={(e) => handleSearch(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Category Filter */}
            <div className="flex gap-2 flex-wrap">
              <Button
                variant={selectedCategory === undefined ? "default" : "outline"}
                size="sm"
                onClick={() => handleCategoryFilter(undefined)}
              >
                All Categories
              </Button>
              {categoryOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={selectedCategory === option.value ? "default" : "outline"}
                  size="sm"
                  onClick={() => handleCategoryFilter(option.value)}
                >
                  {option.label}
                </Button>
              ))}
            </div>

            {/* Status Filter */}
            <div className="flex gap-2 flex-wrap">
              <Button
                variant={selectedStatus === undefined ? "default" : "outline"}
                size="sm"
                onClick={() => handleStatusFilter(undefined)}
              >
                All Status
              </Button>
              {statusOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={selectedStatus === option.value ? "default" : "outline"}
                  size="sm"
                  onClick={() => handleStatusFilter(option.value)}
                >
                  {option.label}
                </Button>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Bounties Grid */}
      {isLoading ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-4 bg-muted rounded w-3/4"></div>
                <div className="h-3 bg-muted rounded w-1/2"></div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="h-3 bg-muted rounded"></div>
                  <div className="h-3 bg-muted rounded w-5/6"></div>
                  <div className="h-3 bg-muted rounded w-4/6"></div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : error ? (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center text-muted-foreground">
              Failed to load bounties. Please try again.
            </div>
          </CardContent>
        </Card>
      ) : bounties.length === 0 ? (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center text-muted-foreground">
              No bounties found matching your criteria.
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {bounties.map((bounty) => (
            <BountyCard
              key={bounty.id}
              bounty={bounty}
              onApply={handleApplyForBounty}
              onView={handleViewBounty}
            />
          ))}
        </div>
      )}

      {/* Pagination */}
      {bountiesResponse?.pagination && bountiesResponse.pagination.has_more && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            onClick={() => setFilters(prev => ({ 
              ...prev, 
              offset: (prev.offset || 0) + (prev.limit || 12) 
            }))}
          >
            Load More
          </Button>
        </div>
      )}
    </div>
  );
}
