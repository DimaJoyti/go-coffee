import React from 'react';
import { format } from 'date-fns';
import { Star, Download, ExternalLink, Github, FileText, Eye } from 'lucide-react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Solution, SolutionCategory, SolutionStatus } from '../types/api';

interface SolutionCardProps {
  solution: Solution;
  onInstall?: (solutionId: number) => void;
  onView?: (solutionId: number) => void;
  onReview?: (solutionId: number) => void;
  showActions?: boolean;
}

const categoryLabels: Record<SolutionCategory, string> = {
  [SolutionCategory.DEFI]: 'DeFi',
  [SolutionCategory.NFT]: 'NFT',
  [SolutionCategory.DAO]: 'DAO',
  [SolutionCategory.ANALYTICS]: 'Analytics',
  [SolutionCategory.INFRASTRUCTURE]: 'Infrastructure',
  [SolutionCategory.SECURITY]: 'Security',
  [SolutionCategory.UI_COMPONENTS]: 'UI Components',
  [SolutionCategory.INTEGRATION]: 'Integration',
};

const statusLabels: Record<SolutionStatus, string> = {
  [SolutionStatus.DRAFT]: 'Draft',
  [SolutionStatus.SUBMITTED]: 'Submitted',
  [SolutionStatus.UNDER_REVIEW]: 'Under Review',
  [SolutionStatus.APPROVED]: 'Approved',
  [SolutionStatus.REJECTED]: 'Rejected',
  [SolutionStatus.DEPRECATED]: 'Deprecated',
};

const statusVariants: Record<SolutionStatus, 'default' | 'secondary' | 'success' | 'warning' | 'destructive' | 'info'> = {
  [SolutionStatus.DRAFT]: 'secondary',
  [SolutionStatus.SUBMITTED]: 'warning',
  [SolutionStatus.UNDER_REVIEW]: 'info',
  [SolutionStatus.APPROVED]: 'success',
  [SolutionStatus.REJECTED]: 'destructive',
  [SolutionStatus.DEPRECATED]: 'destructive',
};

export const SolutionCard: React.FC<SolutionCardProps> = ({
  solution,
  onInstall,
  onView,
  onReview,
  showActions = true,
}) => {
  const isInstallable = solution.status === SolutionStatus.APPROVED;
  const averageRating = solution.reviews.length > 0
    ? solution.reviews.reduce((sum, review) => sum + review.rating, 0) / solution.reviews.length
    : 0;

  return (
    <Card className="h-full flex flex-col hover:shadow-lg transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-lg line-clamp-2">{solution.name}</CardTitle>
          <Badge variant={statusVariants[solution.status]}>
            {statusLabels[solution.status]}
          </Badge>
        </div>
        <div className="flex items-center justify-between">
          <Badge variant="outline" className="text-xs">
            {categoryLabels[solution.category]}
          </Badge>
          <span className="text-xs text-muted-foreground">v{solution.version}</span>
        </div>
      </CardHeader>

      <CardContent className="flex-1 pb-3">
        <p className="text-sm text-muted-foreground line-clamp-3 mb-4">
          {solution.description}
        </p>

        <div className="space-y-2 mb-4">
          {averageRating > 0 && (
            <div className="flex items-center gap-2 text-sm">
              <div className="flex items-center gap-1">
                <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                <span className="font-medium">{averageRating.toFixed(1)}</span>
              </div>
              <span className="text-muted-foreground">
                ({solution.reviews.length} review{solution.reviews.length !== 1 ? 's' : ''})
              </span>
            </div>
          )}

          <div className="flex items-center gap-2 text-sm">
            <Download className="h-4 w-4 text-muted-foreground" />
            <span>{solution.installations.toLocaleString()} installation{solution.installations !== 1 ? 's' : ''}</span>
          </div>

          <div className="text-xs text-muted-foreground">
            by {solution.developer_address.slice(0, 6)}...{solution.developer_address.slice(-4)}
          </div>

          <div className="text-xs text-muted-foreground">
            Updated {format(new Date(solution.updated_at), 'MMM dd, yyyy')}
          </div>
        </div>

        {solution.quality_score && (
          <div className="mb-4">
            <div className="text-xs text-muted-foreground mb-1">Quality Score</div>
            <div className="flex items-center gap-2">
              <div className="flex-1 bg-gray-200 rounded-full h-2">
                <div
                  className="bg-green-500 h-2 rounded-full"
                  style={{ width: `${solution.quality_score.overall_score * 10}%` }}
                />
              </div>
              <span className="text-sm font-medium">
                {solution.quality_score.overall_score.toFixed(1)}/10
              </span>
            </div>
          </div>
        )}

        {solution.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {solution.tags.slice(0, 3).map((tag) => (
              <Badge key={tag} variant="outline" className="text-xs">
                {tag}
              </Badge>
            ))}
            {solution.tags.length > 3 && (
              <Badge variant="outline" className="text-xs">
                +{solution.tags.length - 3} more
              </Badge>
            )}
          </div>
        )}
      </CardContent>

      {showActions && (
        <CardFooter className="pt-0 gap-2 flex-col">
          <div className="flex gap-2 w-full">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onView?.(solution.id)}
              className="flex-1"
            >
              <Eye className="h-4 w-4 mr-1" />
              View
            </Button>
            {isInstallable && (
              <Button
                size="sm"
                onClick={() => onInstall?.(solution.id)}
                className="flex-1"
              >
                <Download className="h-4 w-4 mr-1" />
                Install
              </Button>
            )}
          </div>

          <div className="flex gap-1 w-full">
            {solution.repository_url && (
              <Button
                variant="ghost"
                size="sm"
                asChild
                className="flex-1"
              >
                <a
                  href={solution.repository_url}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <Github className="h-4 w-4 mr-1" />
                  Code
                </a>
              </Button>
            )}

            {solution.documentation_url && (
              <Button
                variant="ghost"
                size="sm"
                asChild
                className="flex-1"
              >
                <a
                  href={solution.documentation_url}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <FileText className="h-4 w-4 mr-1" />
                  Docs
                </a>
              </Button>
            )}

            {solution.demo_url && (
              <Button
                variant="ghost"
                size="sm"
                asChild
                className="flex-1"
              >
                <a
                  href={solution.demo_url}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <ExternalLink className="h-4 w-4 mr-1" />
                  Demo
                </a>
              </Button>
            )}
          </div>

          {isInstallable && onReview && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => onReview(solution.id)}
              className="w-full"
            >
              <Star className="h-4 w-4 mr-1" />
              Write Review
            </Button>
          )}
        </CardFooter>
      )}
    </Card>
  );
};
