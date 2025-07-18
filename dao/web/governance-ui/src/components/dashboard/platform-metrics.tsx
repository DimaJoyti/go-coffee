'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Badge } from '../@/shared/components/ui/badge';
import { 
  BarChart3, 
  DollarSign, 
  Users, 
  TrendingUp,
} from 'lucide-react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { useAnalyticsOverview, useTVLMetrics, useMAUMetrics } from '../@/shared/hooks/api-hooks';

// Mock data for platform metrics
const platformData = [
  { month: 'Jan', tvl: 4200000, mau: 18000, revenue: 125000 },
  { month: 'Feb', tvl: 4800000, mau: 22000, revenue: 145000 },
  { month: 'Mar', tvl: 4500000, mau: 20000, revenue: 135000 },
  { month: 'Apr', tvl: 5200000, mau: 25000, revenue: 165000 },
  { month: 'May', tvl: 5800000, mau: 28000, revenue: 185000 },
  { month: 'Jun', tvl: 6500000, mau: 32000, revenue: 210000 },
];

export function PlatformMetrics() {
  const { data: analyticsOverview } = useAnalyticsOverview();
  const { data: tvlMetrics } = useTVLMetrics();
  const { data: mauMetrics } = useMAUMetrics();

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <BarChart3 className="w-5 h-5" />
            Platform Metrics
          </CardTitle>
          <Badge variant="info">Real-time</Badge>
        </div>
      </CardHeader>
      <CardContent>
        {/* Key Metrics Summary */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="text-center p-3 bg-accent/50 rounded-lg">
            <div className="flex items-center justify-center gap-1 mb-1">
              <DollarSign className="w-4 h-4 text-green-600" />
              <span className="text-xs text-muted-foreground">TVL</span>
            </div>
            <div className="text-lg font-bold">$6.5M</div>
            <div className="text-xs text-green-600">+12.1%</div>
          </div>
          
          <div className="text-center p-3 bg-accent/50 rounded-lg">
            <div className="flex items-center justify-center gap-1 mb-1">
              <Users className="w-4 h-4 text-blue-600" />
              <span className="text-xs text-muted-foreground">MAU</span>
            </div>
            <div className="text-lg font-bold">32K</div>
            <div className="text-xs text-green-600">+14.3%</div>
          </div>
          
          <div className="text-center p-3 bg-accent/50 rounded-lg">
            <div className="flex items-center justify-center gap-1 mb-1">
              <TrendingUp className="w-4 h-4 text-purple-600" />
              <span className="text-xs text-muted-foreground">Revenue</span>
            </div>
            <div className="text-lg font-bold">$210K</div>
            <div className="text-xs text-green-600">+13.5%</div>
          </div>
        </div>

        {/* Platform Growth Chart */}
        <div className="h-[250px] mb-6">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={platformData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="month" />
              <YAxis />
              <Tooltip
                formatter={(value, name) => {
                  if (name === 'tvl') return [`$${(value as number / 1000000).toFixed(1)}M`, 'TVL'];
                  if (name === 'mau') return [`${(value as number / 1000).toFixed(0)}K`, 'MAU'];
                  if (name === 'revenue') return [`$${(value as number / 1000).toFixed(0)}K`, 'Revenue'];
                  return [value, name];
                }}
              />
              <Area
                type="monotone"
                dataKey="tvl"
                stackId="1"
                stroke="#8884d8"
                fill="#8884d8"
                fillOpacity={0.3}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Platform Health Indicators */}
        <div className="space-y-3">
          <div className="text-sm font-medium">Platform Health</div>
          
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-sm">Developer Retention</span>
              <div className="flex items-center gap-2">
                <div className="w-20 bg-gray-200 rounded-full h-2">
                  <div className="bg-green-500 h-2 rounded-full" style={{ width: '87%' }} />
                </div>
                <span className="text-sm font-medium">87%</span>
              </div>
            </div>
            
            <div className="flex justify-between items-center">
              <span className="text-sm">Solution Quality Score</span>
              <div className="flex items-center gap-2">
                <div className="w-20 bg-gray-200 rounded-full h-2">
                  <div className="bg-blue-500 h-2 rounded-full" style={{ width: '92%' }} />
                </div>
                <span className="text-sm font-medium">4.6/5</span>
              </div>
            </div>
            
            <div className="flex justify-between items-center">
              <span className="text-sm">Bounty Completion Rate</span>
              <div className="flex items-center gap-2">
                <div className="w-20 bg-gray-200 rounded-full h-2">
                  <div className="bg-purple-500 h-2 rounded-full" style={{ width: '94%' }} />
                </div>
                <span className="text-sm font-medium">94%</span>
              </div>
            </div>
            
            <div className="flex justify-between items-center">
              <span className="text-sm">Community Engagement</span>
              <div className="flex items-center gap-2">
                <div className="w-20 bg-gray-200 rounded-full h-2">
                  <div className="bg-orange-500 h-2 rounded-full" style={{ width: '78%' }} />
                </div>
                <span className="text-sm font-medium">78%</span>
              </div>
            </div>
          </div>
        </div>

        {/* Recent Milestones */}
        <div className="mt-6 pt-4 border-t">
          <div className="text-sm font-medium mb-3">Recent Milestones</div>
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 bg-green-500 rounded-full" />
              <span>Reached $6M TVL milestone</span>
              <span className="text-xs text-muted-foreground ml-auto">2 days ago</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 bg-blue-500 rounded-full" />
              <span>30K MAU achieved</span>
              <span className="text-xs text-muted-foreground ml-auto">1 week ago</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 bg-purple-500 rounded-full" />
              <span>100th solution approved</span>
              <span className="text-xs text-muted-foreground ml-auto">2 weeks ago</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
