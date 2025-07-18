import React from 'react';
import { format } from 'date-fns';
import { Clock, DollarSign, User, Tag } from 'lucide-react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Bounty, BountyCategory, BountyStatus } from '../types/api';

interface BountyCardProps {
  bounty: Bounty;
  onApply?: (bountyId: number) => void;
  onView?: (bountyId: number) => void;
  showActions?: boolean;
}

const categoryLabels: Record<BountyCategory, string> = {
  [BountyCategory.TVL_GROWTH]: 'TVL Growth',
  [BountyCategory.MAU_EXPANSION]: 'MAU Expansion',
  [BountyCategory.INNOVATION]: 'Innovation',
  [BountyCategory.SECURITY]: 'Security',
  [BountyCategory.INFRASTRUCTURE]: 'Infrastructure',
  [BountyCategory.COMMUNITY]: 'Community',
};

const statusLabels: Record<BountyStatus, string> = {
  [BountyStatus.OPEN]: 'Open',
  [BountyStatus.ASSIGNED]: 'Assigned',
  [BountyStatus.IN_PROGRESS]: 'In Progress',
  [BountyStatus.UNDER_REVIEW]: 'Under Review',
  [BountyStatus.COMPLETED]: 'Completed',
  [BountyStatus.CANCELLED]: 'Cancelled',
};

const statusVariants: Record<BountyStatus, 'default' | 'secondary' | 'success' | 'warning' | 'destructive' | 'info'> = {
  [BountyStatus.OPEN]: 'success',
  [BountyStatus.ASSIGNED]: 'warning',
  [BountyStatus.IN_PROGRESS]: 'info',
  [BountyStatus.UNDER_REVIEW]: 'warning',
  [BountyStatus.COMPLETED]: 'success',
  [BountyStatus.CANCELLED]: 'destructive',
};

export const BountyCard: React.FC<BountyCardProps> = ({
  bounty,
  onApply,
  onView,
  showActions = true,
}) => {
  const isApplicable = bounty.status === BountyStatus.OPEN;
  const deadline = new Date(bounty.deadline);
  const isUrgent = deadline.getTime() - Date.now() < 7 * 24 * 60 * 60 * 1000; // 7 days

  return (
    <Card className="h-full flex flex-col hover:shadow-lg transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-lg line-clamp-2">{bounty.title}</CardTitle>
          <Badge variant={statusVariants[bounty.status]}>
            {statusLabels[bounty.status]}
          </Badge>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Tag className="h-4 w-4" />
          <span>{categoryLabels[bounty.category]}</span>
        </div>
      </CardHeader>

      <CardContent className="flex-1 pb-3">
        <p className="text-sm text-muted-foreground line-clamp-3 mb-4">
          {bounty.description}
        </p>

        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm">
            <DollarSign className="h-4 w-4 text-green-600" />
            <span className="font-semibold">
              {bounty.reward_amount} {bounty.currency}
            </span>
          </div>

          <div className="flex items-center gap-2 text-sm">
            <Clock className={`h-4 w-4 ${isUrgent ? 'text-red-500' : 'text-muted-foreground'}`} />
            <span className={isUrgent ? 'text-red-500 font-medium' : ''}>
              Due {format(deadline, 'MMM dd, yyyy')}
            </span>
          </div>

          {bounty.assigned_developer && (
            <div className="flex items-center gap-2 text-sm">
              <User className="h-4 w-4 text-muted-foreground" />
              <span>
                Assigned to {bounty.assigned_developer.slice(0, 6)}...{bounty.assigned_developer.slice(-4)}
              </span>
            </div>
          )}
        </div>

        {bounty.required_skills.length > 0 && (
          <div className="mt-3">
            <div className="flex flex-wrap gap-1">
              {bounty.required_skills.slice(0, 3).map((skill) => (
                <Badge key={skill} variant="outline" className="text-xs">
                  {skill}
                </Badge>
              ))}
              {bounty.required_skills.length > 3 && (
                <Badge variant="outline" className="text-xs">
                  +{bounty.required_skills.length - 3} more
                </Badge>
              )}
            </div>
          </div>
        )}

        {bounty.milestones.length > 0 && (
          <div className="mt-3 text-xs text-muted-foreground">
            {bounty.milestones.length} milestone{bounty.milestones.length !== 1 ? 's' : ''}
          </div>
        )}
      </CardContent>

      {showActions && (
        <CardFooter className="pt-0 gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onView?.(bounty.id)}
            className="flex-1"
          >
            View Details
          </Button>
          {isApplicable && (
            <Button
              size="sm"
              onClick={() => onApply?.(bounty.id)}
              className="flex-1"
            >
              Apply
            </Button>
          )}
        </CardFooter>
      )}
    </Card>
  );
};
