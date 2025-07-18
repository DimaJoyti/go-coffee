'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Badge } from '@/shared/components/ui/badge';
import { 
  Target, 
  Package, 
  DollarSign, 
  TrendingUp,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react';
import { useCurrentUser, usePerformanceDashboard } from '@/shared/hooks/api-hooks';

interface MetricCardProps {
  title: string;
  value: string;
  change?: string;
  changeType?: 'positive' | 'negative';
  icon: React.ReactNode;
}

function MetricCard({ title, value, change, changeType, icon }: MetricCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {change && (
          <div className="flex items-center text-xs text-muted-foreground">
            {changeType === 'positive' ? (
              <ArrowUpRight className="w-4 h-4 text-green-500 mr-1" />
            ) : (
              <ArrowDownRight className="w-4 h-4 text-red-500 mr-1" />
            )}
            <span className={changeType === 'positive' ? 'text-green-500' : 'text-red-500'}>
              {change}
            </span>
            <span className="ml-1">from last month</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function DashboardOverview() {
  const { data: user } = useCurrentUser();
  const { data: dashboard } = usePerformanceDashboard();

  // Mock data for demonstration
  const metrics = [
    {
      title: 'Active Bounties',
      value: user?.data?.active_bounties?.toString() || '3',
      change: '+2',
      changeType: 'positive' as const,
      icon: <Target className="h-4 w-4 text-muted-foreground" />,
    },
    {
      title: 'Solutions Created',
      value: user?.data?.solutions_count?.toString() || '12',
      change: '+3',
      changeType: 'positive' as const,
      icon: <Package className="h-4 w-4 text-muted-foreground" />,
    },
    {
      title: 'Total Earnings',
      value: user?.data?.total_earnings || '$8,450',
      change: '+12%',
      changeType: 'positive' as const,
      icon: <DollarSign className="h-4 w-4 text-muted-foreground" />,
    },
    {
      title: 'Reputation Score',
      value: user?.data?.reputation_score?.toFixed(1) || '8.7',
      change: '+0.3',
      changeType: 'positive' as const,
      icon: <TrendingUp className="h-4 w-4 text-muted-foreground" />,
    },
  ];

  return (
    <>
      {metrics.map((metric, index) => (
        <MetricCard key={index} {...metric} />
      ))}
    </>
  );
}
