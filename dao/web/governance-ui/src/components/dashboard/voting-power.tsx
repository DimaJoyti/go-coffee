'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Badge } from '../@/shared/components/ui/badge';
import { Button } from '../@/shared/components/ui/button';
import { 
  Coins, 
  TrendingUp, 
  Users, 
  Award,
  ArrowUpRight,
} from 'lucide-react';

export function VotingPower() {
  // Mock voting power data
  const votingData = {
    totalTokens: 15000,
    votingPower: 12500,
    delegatedTo: 0,
    delegatedFrom: 2500,
    participationRate: 85.5,
    rank: 47,
    totalHolders: 2341,
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Coins className="w-5 h-5" />
          Voting Power
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Token Balance */}
        <div className="space-y-3">
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Total Tokens</span>
            <span className="font-semibold">{votingData.totalTokens.toLocaleString()}</span>
          </div>
          
          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Voting Power</span>
            <span className="font-semibold text-primary">{votingData.votingPower.toLocaleString()}</span>
          </div>

          <div className="flex justify-between items-center">
            <span className="text-sm text-muted-foreground">Delegated From Others</span>
            <span className="font-semibold text-green-600">+{votingData.delegatedFrom.toLocaleString()}</span>
          </div>
        </div>

        {/* Voting Power Breakdown */}
        <div className="space-y-2">
          <div className="text-sm font-medium">Power Distribution</div>
          <div className="w-full bg-gray-200 rounded-full h-3">
            <div className="bg-primary h-3 rounded-full relative" style={{ width: '83.3%' }}>
              <div className="bg-green-500 h-3 rounded-r-full absolute right-0" style={{ width: '20%' }} />
            </div>
          </div>
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>Own Tokens</span>
            <span>Delegated</span>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-4">
          <div className="text-center p-3 bg-accent/50 rounded-lg">
            <div className="text-lg font-bold text-primary">#{votingData.rank}</div>
            <div className="text-xs text-muted-foreground">Voting Rank</div>
          </div>
          
          <div className="text-center p-3 bg-accent/50 rounded-lg">
            <div className="text-lg font-bold text-green-600">{votingData.participationRate}%</div>
            <div className="text-xs text-muted-foreground">Participation</div>
          </div>
        </div>

        {/* Recent Voting Activity */}
        <div className="space-y-3">
          <div className="text-sm font-medium">Recent Votes</div>
          <div className="space-y-2">
            <div className="flex items-center justify-between p-2 bg-accent/30 rounded">
              <div className="text-sm">Reward Pool Increase</div>
              <Badge variant="success" className="text-xs">For</Badge>
            </div>
            <div className="flex items-center justify-between p-2 bg-accent/30 rounded">
              <div className="text-sm">Quality Score Weighting</div>
              <Badge variant="success" className="text-xs">For</Badge>
            </div>
            <div className="flex items-center justify-between p-2 bg-accent/30 rounded">
              <div className="text-sm">Polygon Network Support</div>
              <Badge variant="success" className="text-xs">For</Badge>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="space-y-2">
          <Button className="w-full" size="sm">
            <Vote className="w-4 h-4 mr-2" />
            Delegate Voting Power
          </Button>
          <Button variant="outline" className="w-full" size="sm">
            <TrendingUp className="w-4 h-4 mr-2" />
            View Voting History
          </Button>
        </div>

        {/* Achievements */}
        <div className="pt-4 border-t">
          <div className="text-sm font-medium mb-2">Governance Achievements</div>
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <Award className="w-4 h-4 text-yellow-500" />
              <span>Active Voter</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Users className="w-4 h-4 text-blue-500" />
              <span>Community Delegate</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <TrendingUp className="w-4 h-4 text-green-500" />
              <span>Early Adopter</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
