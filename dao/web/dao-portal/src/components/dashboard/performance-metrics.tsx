'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Badge } from '../@/shared/components/ui/badge';
import { TrendingUp, DollarSign, Users } from 'lucide-react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
} from 'recharts';
import { useTVLHistory, useMAUHistory } from '../@/shared/hooks/api-hooks';

// Mock data for demonstration
const tvlData = [
  { month: 'Jan', tvl: 2400, contribution: 240 },
  { month: 'Feb', tvl: 3200, contribution: 320 },
  { month: 'Mar', tvl: 2800, contribution: 280 },
  { month: 'Apr', tvl: 4100, contribution: 410 },
  { month: 'May', tvl: 3900, contribution: 390 },
  { month: 'Jun', tvl: 5200, contribution: 520 },
];

const mauData = [
  { month: 'Jan', mau: 1200, impact: 120 },
  { month: 'Feb', mau: 1800, impact: 180 },
  { month: 'Mar', mau: 1600, impact: 160 },
  { month: 'Apr', mau: 2200, impact: 220 },
  { month: 'May', mau: 2800, impact: 280 },
  { month: 'Jun', mau: 3200, impact: 320 },
];

export function PerformanceMetrics() {
  const { data: tvlHistory } = useTVLHistory('go-coffee-defi', 180);
  const { data: mauHistory } = useMAUHistory('defi_trading', 6);

  return (
    <div className="space-y-6">
      {/* TVL Impact */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="w-5 h-5 text-green-600" />
              TVL Impact
            </CardTitle>
            <Badge variant="success">+15.2% this month</Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={tvlData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis />
                <Tooltip
                  formatter={(value, name) => [
                    `$${value.toLocaleString()}`,
                    name === 'tvl' ? 'Total TVL' : 'Your Contribution'
                  ]}
                />
                <Area
                  type="monotone"
                  dataKey="tvl"
                  stackId="1"
                  stroke="#8884d8"
                  fill="#8884d8"
                  fillOpacity={0.3}
                />
                <Area
                  type="monotone"
                  dataKey="contribution"
                  stackId="2"
                  stroke="#82ca9d"
                  fill="#82ca9d"
                  fillOpacity={0.8}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* MAU Impact */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Users className="w-5 h-5 text-blue-600" />
              MAU Impact
            </CardTitle>
            <Badge variant="info">+8.5% this month</Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={mauData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis />
                <Tooltip
                  formatter={(value, name) => [
                    value.toLocaleString(),
                    name === 'mau' ? 'Total MAU' : 'Your Impact'
                  ]}
                />
                <Line
                  type="monotone"
                  dataKey="mau"
                  stroke="#8884d8"
                  strokeWidth={2}
                  dot={{ fill: '#8884d8' }}
                />
                <Line
                  type="monotone"
                  dataKey="impact"
                  stroke="#82ca9d"
                  strokeWidth={2}
                  dot={{ fill: '#82ca9d' }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
