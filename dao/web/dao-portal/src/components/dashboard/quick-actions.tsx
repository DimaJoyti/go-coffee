'use client';

import React from 'react';
import Link from 'next/link';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Button } from '../@/shared/components/ui/button';
import { 
  Plus, 
  Search, 
  Upload, 
  BarChart3,
  Target,
  Package,
} from 'lucide-react';

const quickActions = [
  {
    title: 'Browse Bounties',
    description: 'Find new bounties to work on',
    icon: <Search className="w-5 h-5" />,
    href: '/bounties',
    variant: 'default' as const,
  },
  {
    title: 'Submit Solution',
    description: 'Upload a new solution',
    icon: <Upload className="w-5 h-5" />,
    href: '/solutions/create',
    variant: 'outline' as const,
  },
  {
    title: 'View Performance',
    description: 'Check your metrics',
    icon: <BarChart3 className="w-5 h-5" />,
    href: '/performance',
    variant: 'outline' as const,
  },
  {
    title: 'Apply for Bounty',
    description: 'Quick apply to featured bounty',
    icon: <Target className="w-5 h-5" />,
    href: '/bounties?featured=true',
    variant: 'outline' as const,
  },
];

export function QuickActions() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Plus className="w-5 h-5" />
          Quick Actions
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {quickActions.map((action, index) => (
          <Link key={index} href={action.href}>
            <Button
              variant={action.variant}
              className="w-full justify-start h-auto p-4"
            >
              <div className="flex items-center gap-3">
                {action.icon}
                <div className="text-left">
                  <div className="font-medium">{action.title}</div>
                  <div className="text-xs text-muted-foreground">
                    {action.description}
                  </div>
                </div>
              </div>
            </Button>
          </Link>
        ))}
      </CardContent>
    </Card>
  );
}
