'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Button } from '../@/shared/components/ui/button';
import { Input } from '../@/shared/components/ui/input';
import { Badge } from '../@/shared/components/ui/badge';
import { SolutionCard } from '../@/shared/components/solution-card';
import { 
  Search, 
  Filter, 
  Package,
  Plus,
  TrendingUp,
  Star,
  Download,
} from 'lucide-react';
import { useSolutions, usePopularSolutions, useTrendingSolutions } from '../@/shared/hooks/api-hooks';
import { SolutionCategory, SolutionStatus, SolutionFilters } from '../@/shared/types/api';

const categoryOptions = [
  { value: SolutionCategory.DEFI, label: 'DeFi' },
  { value: SolutionCategory.NFT, label: 'NFT' },
  { value: SolutionCategory.DAO, label: 'DAO' },
  { value: SolutionCategory.ANALYTICS, label: 'Analytics' },
  { value: SolutionCategory.INFRASTRUCTURE, label: 'Infrastructure' },
  { value: SolutionCategory.SECURITY, label: 'Security' },
  { value: SolutionCategory.UI_COMPONENTS, label: 'UI Components' },
  { value: SolutionCategory.INTEGRATION, label: 'Integration' },
];

const statusOptions = [
  { value: SolutionStatus.APPROVED, label: 'Approved' },
  { value: SolutionStatus.UNDER_REVIEW, label: 'Under Review' },
  { value: SolutionStatus.SUBMITTED, label: 'Submitted' },
];

export default function SolutionsPage() {
  const [filters, setFilters] = useState<SolutionFilters>({
    limit: 12,
    offset: 0,
    status: SolutionStatus.APPROVED, // Default to approved solutions
  });
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<SolutionCategory | undefined>();

  const { data: solutionsResponse, isLoading, error } = useSolutions(filters);
  const { data: popularSolutions } = usePopularSolutions(5);
  const { data: trendingSolutions } = useTrendingSolutions(5);

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setFilters(prev => ({ ...prev, search: query, offset: 0 }));
  };

  const handleCategoryFilter = (category: SolutionCategory | undefined) => {
    setSelectedCategory(category);
    setFilters(prev => ({ ...prev, category, offset: 0 }));
  };

  const handleInstallSolution = (solutionId: number) => {
    // TODO: Implement solution installation modal
    console.log('Install solution:', solutionId);
  };

  const handleViewSolution = (solutionId: number) => {
    // TODO: Navigate to solution details page
    console.log('View solution:', solutionId);
  };

  const handleReviewSolution = (solutionId: number) => {
    // TODO: Implement review modal
    console.log('Review solution:', solutionId);
  };

  const solutions = solutionsResponse?.data || [];
  const totalSolutions = solutionsResponse?.pagination?.total || 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Solutions</h1>
          <p className="text-muted-foreground">
            Discover and install high-quality solutions from the Developer DAO community.
          </p>
        </div>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          Submit Solution
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Solutions</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalSolutions}</div>
            <p className="text-xs text-muted-foreground">
              +8% from last month
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Approved Solutions</CardTitle>
            <Package className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {solutions.filter(s => s.status === SolutionStatus.APPROVED).length}
            </div>
            <p className="text-xs text-muted-foreground">
              Ready to install
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Downloads</CardTitle>
            <Download className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">45.2K</div>
            <p className="text-xs text-muted-foreground">
              Across all solutions
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Average Rating</CardTitle>
            <Star className="h-4 w-4 text-yellow-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">4.6</div>
            <p className="text-xs text-muted-foreground">
              Out of 5 stars
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Featured Sections */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Popular Solutions */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="w-5 h-5" />
              Popular Solutions
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {popularSolutions?.data?.slice(0, 3).map((solution) => (
                <div key={solution.id} className="flex items-center gap-3 p-2 rounded-lg border">
                  <Package className="w-8 h-8 text-muted-foreground" />
                  <div className="flex-1">
                    <div className="font-medium text-sm">{solution.name}</div>
                    <div className="text-xs text-muted-foreground">
                      {solution.installations.toLocaleString()} installs
                    </div>
                  </div>
                  <Badge variant="outline">{solution.category}</Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Trending Solutions */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Star className="w-5 h-5" />
              Trending Solutions
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {trendingSolutions?.data?.slice(0, 3).map((solution) => (
                <div key={solution.id} className="flex items-center gap-3 p-2 rounded-lg border">
                  <Package className="w-8 h-8 text-muted-foreground" />
                  <div className="flex-1">
                    <div className="font-medium text-sm">{solution.name}</div>
                    <div className="text-xs text-muted-foreground">
                      {solution.reviews.length > 0 && (
                        <>★ {(solution.reviews.reduce((sum, r) => sum + r.rating, 0) / solution.reviews.length).toFixed(1)}</>
                      )}
                    </div>
                  </div>
                  <Badge variant="outline">{solution.category}</Badge>
                </div>
              ))}
            </div>
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
                placeholder="Search solutions..."
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
          </div>
        </CardContent>
      </Card>

      {/* Solutions Grid */}
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
              Failed to load solutions. Please try again.
            </div>
          </CardContent>
        </Card>
      ) : solutions.length === 0 ? (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center text-muted-foreground">
              No solutions found matching your criteria.
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {solutions.map((solution) => (
            <SolutionCard
              key={solution.id}
              solution={solution}
              onInstall={handleInstallSolution}
              onView={handleViewSolution}
              onReview={handleReviewSolution}
            />
          ))}
        </div>
      )}

      {/* Pagination */}
      {solutionsResponse?.pagination && solutionsResponse.pagination.has_more && (
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
