'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '../@/shared/components/ui/card';
import { Button } from '../@/shared/components/ui/button';
import { Badge } from '../@/shared/components/ui/badge';
import { 
  TrendingUp, 
  DollarSign, 
  Users, 
  Target,
  Award,
  Calendar,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { 
  usePerformanceDashboard, 
  useLeaderboard,
  useTVLHistory,
  useMAUHistory,
} from '../@/shared/hooks/api-hooks';

// Mock data for detailed performance metrics
const performanceData = [
  { month: 'Jan', tvl: 2400, mau: 1200, earnings: 1200 },
  { month: 'Feb', tvl: 3200, mau: 1800, earnings: 1800 },
  { month: 'Mar', tvl: 2800, mau: 1600, earnings: 1400 },
  { month: 'Apr', tvl: 4100, mau: 2200, earnings: 2200 },
  { month: 'May', tvl: 3900, mau: 2800, earnings: 2600 },
  { month: 'Jun', tvl: 5200, mau: 3200, earnings: 3200 },
];

const categoryData = [
  { name: 'DeFi', value: 35, color: '#8884d8' },
  { name: 'Analytics', value: 25, color: '#82ca9d' },
  { name: 'Infrastructure', value: 20, color: '#ffc658' },
  { name: 'Security', value: 15, color: '#ff7300' },
  { name: 'Other', value: 5, color: '#00ff00' },
];

const reputationData = [
  { skill: 'Technical', score: 9.2 },
  { skill: 'Communication', score: 8.7 },
  { skill: 'Reliability', score: 9.5 },
  { skill: 'Innovation', score: 8.9 },
  { skill: 'Community', score: 8.4 },
];

export default function PerformancePage() {
  const [timeRange, setTimeRange] = useState<'1M' | '3M' | '6M' | '1Y'>('6M');
  
  const { data: dashboard } = usePerformanceDashboard();
  const { data: leaderboard } = useLeaderboard(10);
  const { data: tvlHistory } = useTVLHistory('go-coffee-defi', 180);
  const { data: mauHistory } = useMAUHistory('defi_trading', 6);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Performance</h1>
          <p className="text-muted-foreground">
            Track your impact, earnings, and reputation in the Developer DAO ecosystem.
          </p>
        </div>
        <div className="flex gap-2">
          {(['1M', '3M', '6M', '1Y'] as const).map((range) => (
            <Button
              key={range}
              variant={timeRange === range ? "default" : "outline"}
              size="sm"
              onClick={() => setTimeRange(range)}
            >
              {range}
            </Button>
          ))}
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">TVL Impact</CardTitle>
            <DollarSign className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$125,000</div>
            <div className="flex items-center text-xs text-muted-foreground">
              <ArrowUpRight className="w-4 h-4 text-green-500 mr-1" />
              <span className="text-green-500">+15.2%</span>
              <span className="ml-1">from last month</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">MAU Impact</CardTitle>
            <Users className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">2,450</div>
            <div className="flex items-center text-xs text-muted-foreground">
              <ArrowUpRight className="w-4 h-4 text-green-500 mr-1" />
              <span className="text-green-500">+8.5%</span>
              <span className="ml-1">from last month</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Earnings</CardTitle>
            <DollarSign className="h-4 w-4 text-yellow-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">$18,750</div>
            <div className="flex items-center text-xs text-muted-foreground">
              <ArrowUpRight className="w-4 h-4 text-green-500 mr-1" />
              <span className="text-green-500">+22.1%</span>
              <span className="ml-1">from last month</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Reputation Score</CardTitle>
            <Award className="h-4 w-4 text-purple-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">8.9</div>
            <div className="flex items-center text-xs text-muted-foreground">
              <ArrowUpRight className="w-4 h-4 text-green-500 mr-1" />
              <span className="text-green-500">+0.3</span>
              <span className="ml-1">from last month</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Performance Charts */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* TVL & MAU Impact */}
        <Card>
          <CardHeader>
            <CardTitle>Impact Over Time</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={performanceData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="month" />
                  <YAxis />
                  <Tooltip
                    formatter={(value, name) => [
                      name === 'tvl' ? `$${value.toLocaleString()}` : value.toLocaleString(),
                      name === 'tvl' ? 'TVL Impact' : 'MAU Impact'
                    ]}
                  />
                  <Line
                    type="monotone"
                    dataKey="tvl"
                    stroke="#8884d8"
                    strokeWidth={2}
                    dot={{ fill: '#8884d8' }}
                  />
                  <Line
                    type="monotone"
                    dataKey="mau"
                    stroke="#82ca9d"
                    strokeWidth={2}
                    dot={{ fill: '#82ca9d' }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Earnings */}
        <Card>
          <CardHeader>
            <CardTitle>Earnings Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={performanceData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="month" />
                  <YAxis />
                  <Tooltip
                    formatter={(value) => [`$${value.toLocaleString()}`, 'Earnings']}
                  />
                  <Area
                    type="monotone"
                    dataKey="earnings"
                    stroke="#ffc658"
                    fill="#ffc658"
                    fillOpacity={0.6}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Detailed Analytics */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Reputation Breakdown */}
        <Card>
          <CardHeader>
            <CardTitle>Reputation Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {reputationData.map((item) => (
                <div key={item.skill} className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>{item.skill}</span>
                    <span className="font-medium">{item.score}/10</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full"
                      style={{ width: `${item.score * 10}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Category Distribution */}
        <Card>
          <CardHeader>
            <CardTitle>Contribution by Category</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[200px]">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={categoryData}
                    cx="50%"
                    cy="50%"
                    innerRadius={40}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {categoryData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip formatter={(value) => [`${value}%`, 'Contribution']} />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className="mt-4 space-y-2">
              {categoryData.map((item) => (
                <div key={item.name} className="flex items-center gap-2 text-sm">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: item.color }}
                  />
                  <span>{item.name}</span>
                  <span className="ml-auto font-medium">{item.value}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Leaderboard Position */}
        <Card>
          <CardHeader>
            <CardTitle>Leaderboard Position</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="text-center">
                <div className="text-3xl font-bold text-primary">#7</div>
                <div className="text-sm text-muted-foreground">Overall Ranking</div>
              </div>
              
              <div className="space-y-3">
                <div className="flex justify-between items-center">
                  <span className="text-sm">TVL Impact</span>
                  <Badge variant="secondary">#5</Badge>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm">MAU Impact</span>
                  <Badge variant="secondary">#8</Badge>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm">Solutions Quality</span>
                  <Badge variant="secondary">#3</Badge>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm">Community Contribution</span>
                  <Badge variant="secondary">#12</Badge>
                </div>
              </div>

              <div className="pt-4 border-t">
                <Button variant="outline" className="w-full">
                  View Full Leaderboard
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Achievements */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Award className="w-5 h-5" />
            Recent Achievements
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <div className="flex items-center gap-3 p-3 rounded-lg border">
              <div className="w-10 h-10 bg-yellow-100 rounded-full flex items-center justify-center">
                <Award className="w-5 h-5 text-yellow-600" />
              </div>
              <div>
                <div className="font-medium text-sm">Top Contributor</div>
                <div className="text-xs text-muted-foreground">May 2024</div>
              </div>
            </div>

            <div className="flex items-center gap-3 p-3 rounded-lg border">
              <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                <Target className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <div className="font-medium text-sm">Bounty Master</div>
                <div className="text-xs text-muted-foreground">10 bounties completed</div>
              </div>
            </div>

            <div className="flex items-center gap-3 p-3 rounded-lg border">
              <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                <TrendingUp className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <div className="font-medium text-sm">Rising Star</div>
                <div className="text-xs text-muted-foreground">Fastest growing reputation</div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
