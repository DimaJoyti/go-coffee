'use client';

import React from 'react';
import { GovernanceOverview } from '@/components/dashboard/governance-overview';
import { ActiveProposals } from '@/components/dashboard/active-proposals';
import { VotingPower } from '@/components/dashboard/voting-power';
import { PlatformMetrics } from '@/components/dashboard/platform-metrics';
import { RecentActivity } from '@/components/dashboard/recent-activity';

export default function GovernancePage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">DAO Governance</h1>
        <p className="text-muted-foreground">
          Participate in Developer DAO governance, vote on proposals, and shape the future of the platform.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <GovernanceOverview />
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <ActiveProposals />
        </div>
        <div>
          <VotingPower />
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <PlatformMetrics />
        <RecentActivity />
      </div>
    </div>
  );
}
