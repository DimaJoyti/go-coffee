'use client';

import React from 'react';
import { format } from 'date-fns';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Badge } from '../@/shared/components/ui/badge';
import { 
  Activity,
  Vote,
  FileText,
  Star,
  Users,
  Shield,
  Clock,
} from 'lucide-react';

interface ActivityItem {
  id: string;
  type: 'proposal_created' | 'vote_cast' | 'review_submitted' | 'member_joined' | 'admin_action';
  title: string;
  description: string;
  timestamp: Date;
  user: string;
  status?: 'success' | 'pending' | 'warning';
}

// Mock activity data
const activities: ActivityItem[] = [
  {
    id: '1',
    type: 'proposal_created',
    title: 'New Proposal Created',
    description: 'Increase Developer Reward Pool by 25%',
    timestamp: new Date(Date.now() - 30 * 60 * 1000), // 30 minutes ago
    user: '0x1234...5678',
    status: 'pending',
  },
  {
    id: '2',
    type: 'vote_cast',
    title: 'Vote Cast',
    description: 'Voted FOR on Quality Score Weighting proposal',
    timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
    user: '0xabcd...efgh',
    status: 'success',
  },
  {
    id: '3',
    type: 'review_submitted',
    title: 'Solution Review',
    description: 'Reviewed "DeFi Trading Widget" - 5 stars',
    timestamp: new Date(Date.now() - 4 * 60 * 60 * 1000), // 4 hours ago
    user: '0x9876...5432',
    status: 'success',
  },
  {
    id: '4',
    type: 'member_joined',
    title: 'New Member',
    description: 'Developer joined the DAO community',
    timestamp: new Date(Date.now() - 6 * 60 * 60 * 1000), // 6 hours ago
    user: '0xdef0...1234',
    status: 'success',
  },
  {
    id: '5',
    type: 'admin_action',
    title: 'Bounty Approved',
    description: 'NFT Marketplace Integration bounty approved',
    timestamp: new Date(Date.now() - 8 * 60 * 60 * 1000), // 8 hours ago
    user: 'Admin',
    status: 'success',
  },
  {
    id: '6',
    type: 'vote_cast',
    title: 'Vote Cast',
    description: 'Voted FOR on Polygon Network Support proposal',
    timestamp: new Date(Date.now() - 12 * 60 * 60 * 1000), // 12 hours ago
    user: '0x5555...7777',
    status: 'success',
  },
  {
    id: '7',
    type: 'proposal_created',
    title: 'New Proposal Created',
    description: 'Community Grant Program Launch',
    timestamp: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000), // 1 day ago
    user: '0x8888...9999',
    status: 'success',
  },
  {
    id: '8',
    type: 'review_submitted',
    title: 'Solution Review',
    description: 'Reviewed "Yield Optimizer" - 4 stars',
    timestamp: new Date(Date.now() - 1.5 * 24 * 60 * 60 * 1000), // 1.5 days ago
    user: '0x2222...3333',
    status: 'success',
  },
];

function getActivityIcon(type: ActivityItem['type']) {
  switch (type) {
    case 'proposal_created':
      return <FileText className="w-4 h-4" />;
    case 'vote_cast':
      return <Vote className="w-4 h-4" />;
    case 'review_submitted':
      return <Star className="w-4 h-4" />;
    case 'member_joined':
      return <Users className="w-4 h-4" />;
    case 'admin_action':
      return <Shield className="w-4 h-4" />;
    default:
      return <Activity className="w-4 h-4" />;
  }
}

function getStatusBadge(status?: ActivityItem['status']) {
  if (!status) return null;
  
  switch (status) {
    case 'success':
      return <Badge variant="success" className="text-xs">Success</Badge>;
    case 'pending':
      return <Badge variant="warning" className="text-xs">Pending</Badge>;
    case 'warning':
      return <Badge variant="warning" className="text-xs">Warning</Badge>;
    default:
      return null;
  }
}

function getTimeAgo(timestamp: Date) {
  const now = new Date();
  const diffInMinutes = Math.floor((now.getTime() - timestamp.getTime()) / (1000 * 60));
  
  if (diffInMinutes < 60) {
    return `${diffInMinutes}m ago`;
  } else if (diffInMinutes < 1440) {
    return `${Math.floor(diffInMinutes / 60)}h ago`;
  } else {
    return `${Math.floor(diffInMinutes / 1440)}d ago`;
  }
}

export function RecentActivity() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Activity className="w-5 h-5" />
          Recent Activity
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {activities.map((activity) => (
            <div
              key={activity.id}
              className="flex items-start gap-3 p-3 rounded-lg border border-border hover:bg-accent/50 transition-colors"
            >
              <div className="flex-shrink-0 w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                {getActivityIcon(activity.type)}
              </div>
              
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between gap-2 mb-1">
                  <h4 className="text-sm font-medium">{activity.title}</h4>
                  {getStatusBadge(activity.status)}
                </div>
                
                <p className="text-sm text-muted-foreground mb-2">
                  {activity.description}
                </p>
                
                <div className="flex items-center gap-3 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Clock className="w-3 h-3" />
                    {getTimeAgo(activity.timestamp)}
                  </div>
                  
                  <div className="flex items-center gap-1">
                    <Users className="w-3 h-3" />
                    {activity.user}
                  </div>
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
