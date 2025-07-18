'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { 
  Vote, 
  FileText, 
  Users, 
  TrendingUp,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react';

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

export function GovernanceOverview() {
  const metrics = [
    {
      title: 'Active Proposals',
      value: '12',
      change: '+3',
      changeType: 'positive' as const,
      icon: <Vote className="h-4 w-4 text-blue-600" />,
    },
    {
      title: 'Total Proposals',
      value: '87',
      change: '+8',
      changeType: 'positive' as const,
      icon: <FileText className="h-4 w-4 text-green-600" />,
    },
    {
      title: 'Active Voters',
      value: '1,234',
      change: '+156',
      changeType: 'positive' as const,
      icon: <Users className="h-4 w-4 text-purple-600" />,
    },
    {
      title: 'Participation Rate',
      value: '78.5%',
      change: '+5.2%',
      changeType: 'positive' as const,
      icon: <TrendingUp className="h-4 w-4 text-orange-600" />,
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
