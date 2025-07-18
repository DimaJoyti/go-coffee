'use client';

import React from 'react';
import { format } from 'date-fns';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Badge } from '@/shared/components/ui/badge';
import { 
  Activity,
  Target,
  Package,
  DollarSign,
  Star,
  CheckCircle,
  Clock,
} from 'lucide-react';

interface ActivityItem {
  id: string;
  type: 'bounty_applied' | 'bounty_assigned' | 'solution_submitted' | 'payment_received' | 'review_received' | 'milestone_completed';
  title: string;
  description: string;
  timestamp: Date;
  amount?: string;
  status?: 'success' | 'pending' | 'warning';
}

// Mock activity data
const activities: ActivityItem[] = [
  {
    id: '1',
    type: 'payment_received',
    title: 'Payment Received',
    description: 'Milestone payment for DeFi Analytics Dashboard',
    timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
    amount: '$2,500 USDC',
    status: 'success',
  },
  {
    id: '2',
    type: 'review_received',
    title: 'Review Received',
    description: 'Your Trading Widget solution received a 5-star review',
    timestamp: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    status: 'success',
  },
  {
    id: '3',
    type: 'bounty_assigned',
    title: 'Bounty Assigned',
    description: 'You were assigned to "NFT Marketplace Integration"',
    timestamp: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000), // 1 day ago
    amount: '$5,000 USDC',
    status: 'success',
  },
  {
    id: '4',
    type: 'milestone_completed',
    title: 'Milestone Completed',
    description: 'UI Design milestone completed for Analytics Dashboard',
    timestamp: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000), // 2 days ago
    status: 'success',
  },
  {
    id: '5',
    type: 'solution_submitted',
    title: 'Solution Submitted',
    description: 'Submitted "DeFi Yield Optimizer" to marketplace',
    timestamp: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000), // 3 days ago
    status: 'pending',
  },
  {
    id: '6',
    type: 'bounty_applied',
    title: 'Bounty Application',
    description: 'Applied for "Security Audit Tools" bounty',
    timestamp: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000), // 5 days ago
    status: 'pending',
  },
];

function getActivityIcon(type: ActivityItem['type']) {
  switch (type) {
    case 'bounty_applied':
    case 'bounty_assigned':
      return <Target className="w-4 h-4" />;
    case 'solution_submitted':
      return <Package className="w-4 h-4" />;
    case 'payment_received':
      return <DollarSign className="w-4 h-4" />;
    case 'review_received':
      return <Star className="w-4 h-4" />;
    case 'milestone_completed':
      return <CheckCircle className="w-4 h-4" />;
    default:
      return <Activity className="w-4 h-4" />;
  }
}

function getStatusBadge(status?: ActivityItem['status']) {
  if (!status) return null;
  
  switch (status) {
    case 'success':
      return <Badge variant="success">Success</Badge>;
    case 'pending':
      return <Badge variant="warning">Pending</Badge>;
    case 'warning':
      return <Badge variant="warning">Warning</Badge>;
    default:
      return null;
  }
}

export function ActivityFeed() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Activity className="w-5 h-5" />
          Recent Activity
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {activities.map((activity) => (
            <div
              key={activity.id}
              className="flex items-start gap-3 p-3 rounded-lg border border-border hover:bg-accent/50 transition-colors"
            >
              <div className="flex-shrink-0 w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                {getActivityIcon(activity.type)}
              </div>
              
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between gap-2">
                  <h4 className="text-sm font-medium">{activity.title}</h4>
                  {getStatusBadge(activity.status)}
                </div>
                
                <p className="text-sm text-muted-foreground mt-1">
                  {activity.description}
                </p>
                
                <div className="flex items-center gap-2 mt-2">
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Clock className="w-3 h-3" />
                    {format(activity.timestamp, 'MMM dd, HH:mm')}
                  </div>
                  
                  {activity.amount && (
                    <Badge variant="outline" className="text-xs">
                      {activity.amount}
                    </Badge>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
        
        <div className="mt-4 text-center">
          <button className="text-sm text-primary hover:underline">
            View all activity
          </button>
        </div>
      </CardContent>
    </Card>
  );
}
