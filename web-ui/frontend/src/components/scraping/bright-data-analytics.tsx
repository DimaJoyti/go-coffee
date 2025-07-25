// Simplified Bright Data Analytics for when dependencies are not installed
// This provides a basic market data visualization interface

interface MarketData {
  id: string
  source: string
  title: string
  price?: number
  change?: number
  url: string
  lastUpdated: string
  category: 'coffee-prices' | 'competitors' | 'news' | 'social'
}

interface BrightDataAnalyticsProps {
  className?: string
}

export function BrightDataAnalytics({ className }: BrightDataAnalyticsProps) {
  // Mock state for when React hooks aren't available
  const filter: 'all' | 'coffee-prices' | 'competitors' | 'news' | 'social' = 'all'

  // Mock market data for demonstration
  const marketData: MarketData[] = [
    {
      id: '1',
      source: 'Starbucks',
      title: 'Premium Coffee Blend - Market Price Update',
      price: 4.95,
      change: 2.3,
      url: 'https://starbucks.com',
      lastUpdated: new Date(Date.now() - 15 * 60 * 1000).toISOString(),
      category: 'coffee-prices'
    },
    {
      id: '2',
      source: 'Dunkin Donuts',
      title: 'Medium Coffee Pricing Analysis',
      price: 3.49,
      change: -1.2,
      url: 'https://dunkindonuts.com',
      lastUpdated: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      category: 'competitors'
    },
    {
      id: '3',
      source: 'Coffee News Daily',
      title: 'Global Coffee Market Trends Q4 2024',
      url: 'https://coffeenews.com',
      lastUpdated: new Date(Date.now() - 45 * 60 * 1000).toISOString(),
      category: 'news'
    },
    {
      id: '4',
      source: 'Twitter Analytics',
      title: 'Coffee Brand Sentiment Analysis',
      url: 'https://twitter.com',
      lastUpdated: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
      category: 'social'
    }
  ]

  const dataSources = [
    'Starbucks', 'Dunkin Donuts', 'Costa Coffee', 'Tim Hortons',
    'Coffee News Daily', 'Twitter Analytics', 'Instagram Insights'
  ]

  const filteredData = filter === 'all'
    ? marketData
    : marketData.filter((item: MarketData) => item.category === filter)

  // Helper functions for display
  const getCategoryIcon = (category: MarketData['category']) => {
    switch (category) {
      case 'coffee-prices': return '☕'
      case 'competitors': return '📈'
      case 'news': return '🌐'
      case 'social': return '🔍'
    }
  }

  const getCategoryColor = (category: MarketData['category']) => {
    switch (category) {
      case 'coffee-prices': return '#d97706'
      case 'competitors': return '#3b82f6'
      case 'news': return '#10b981'
      case 'social': return '#8b5cf6'
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))

    if (diffInMinutes < 60) return `${diffInMinutes}m ago`
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h ago`
    return `${Math.floor(diffInMinutes / 1440)}d ago`
  }

  const stats = {
    totalSources: dataSources.length || 15,
    lastUpdate: marketData.length > 0
      ? marketData.reduce((latest: string, item: MarketData) =>
          new Date(item.lastUpdated) > new Date(latest) ? item.lastUpdated : latest,
          marketData[0].lastUpdated
        )
      : new Date(Date.now() - 2 * 60 * 1000).toISOString(),
    dataPoints: marketData.length * 50 || 1247, // Simulate data points
    avgCoffeePrice: marketData
      .filter((item: MarketData) => item.price && item.category === 'competitors')
      .reduce((sum: number, item: MarketData, _: number, arr: MarketData[]) => sum + (item.price! / arr.length), 0) || 4.23
  }

  // Return HTML string since React/JSX isn't available
  return `
    <div class="bright-data-analytics ${className || ''}" style="padding: 1.5rem; background: #0f172a; color: #f8fafc; min-height: 100vh;">
      <!-- Header -->
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <div>
          <h1 style="font-size: 2rem; font-weight: bold; margin-bottom: 0.5rem; color: #f8fafc;">
            📊 Market Data & Analytics
          </h1>
          <p style="color: #94a3b8; font-size: 1rem;">
            Real-time market intelligence powered by Bright Data
          </p>
        </div>
        <button style="
          padding: 0.75rem 1rem;
          background: #f59e0b;
          border: none;
          border-radius: 0.5rem;
          color: white;
          cursor: pointer;
          font-size: 0.875rem;
          display: flex;
          align-items: center;
          gap: 0.5rem;
        ">
          🔄 Refresh Data
        </button>
      </div>

      <!-- Stats Overview -->
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem; margin-bottom: 2rem;">
        <!-- Data Sources Card -->
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border: 1px solid rgba(148, 163, 184, 0.1);
          border-radius: 1rem;
          padding: 1.5rem;
        ">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
            <span style="font-size: 1.25rem;">🌐</span>
            <span style="font-weight: 500; color: #f8fafc;">Data Sources</span>
          </div>
          <div style="font-size: 2rem; font-weight: bold; color: #f8fafc;">${stats.totalSources}</div>
          <div style="font-size: 0.875rem; color: #94a3b8;">Active sources</div>
        </div>

        <!-- Data Points Card -->
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border: 1px solid rgba(148, 163, 184, 0.1);
          border-radius: 1rem;
          padding: 1.5rem;
        ">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
            <span style="font-size: 1.25rem;">🔍</span>
            <span style="font-weight: 500; color: #f8fafc;">Data Points</span>
          </div>
          <div style="font-size: 2rem; font-weight: bold; color: #f8fafc;">${stats.dataPoints.toLocaleString()}</div>
          <div style="font-size: 0.875rem; color: #94a3b8;">Collected today</div>
        </div>

        <!-- Average Coffee Price Card -->
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border: 1px solid rgba(148, 163, 184, 0.1);
          border-radius: 1rem;
          padding: 1.5rem;
        ">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
            <span style="font-size: 1.25rem;">☕</span>
            <span style="font-weight: 500; color: #f8fafc;">Avg Coffee Price</span>
          </div>
          <div style="font-size: 2rem; font-weight: bold; color: #f8fafc;">${formatCurrency(stats.avgCoffeePrice)}</div>
          <div style="font-size: 0.875rem; color: #94a3b8;">Market average</div>
        </div>

        <!-- Last Update Card -->
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border: 1px solid rgba(148, 163, 184, 0.1);
          border-radius: 1rem;
          padding: 1.5rem;
        ">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
            <span style="font-size: 1.25rem;">🔄</span>
            <span style="font-weight: 500; color: #f8fafc;">Last Update</span>
          </div>
          <div style="font-size: 1.125rem; font-weight: bold; color: #f8fafc;">${formatRelativeTime(stats.lastUpdate)}</div>
          <div style="font-size: 0.875rem; color: #94a3b8;">Auto-refresh enabled</div>
        </div>
      </div>

      <!-- Filters -->
      <div style="display: flex; align-items: center; gap: 1rem; margin-bottom: 2rem; flex-wrap: wrap;">
        <span style="font-size: 0.875rem; font-weight: 500; color: #f8fafc;">Filter by category:</span>
        ${['all', 'coffee-prices', 'competitors', 'news', 'social'].map(category => `
          <button style="
            padding: 0.5rem 1rem;
            background: ${filter === category ? '#f59e0b' : 'rgba(15, 23, 42, 0.8)'};
            border: 1px solid rgba(148, 163, 184, 0.3);
            border-radius: 0.5rem;
            color: ${filter === category ? 'white' : '#f8fafc'};
            cursor: pointer;
            font-size: 0.875rem;
            text-transform: capitalize;
          ">
            ${category.replace('-', ' ')}
          </button>
        `).join('')}
      </div>

      <!-- Data Grid -->
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); gap: 1.5rem; margin-bottom: 2rem;">
        ${filteredData.map((item: MarketData) => `
          <div style="
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(148, 163, 184, 0.1);
            border-radius: 1rem;
            overflow: hidden;
            transition: box-shadow 0.2s;
          " onmouseover="this.style.boxShadow='0 10px 25px rgba(0,0,0,0.3)'" onmouseout="this.style.boxShadow='none'">
            <!-- Card Header -->
            <div style="padding: 1.5rem; border-bottom: 1px solid rgba(148, 163, 184, 0.1);">
              <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1rem;">
                <div style="display: flex; align-items: center; gap: 0.5rem;">
                  <span style="
                    padding: 0.25rem 0.5rem;
                    background: ${getCategoryColor(item.category)};
                    color: white;
                    border-radius: 0.25rem;
                    font-size: 0.75rem;
                    font-weight: 500;
                    display: inline-flex;
                    align-items: center;
                    gap: 0.25rem;
                  ">
                    ${getCategoryIcon(item.category)} ${item.category.replace('-', ' ')}
                  </span>
                </div>
                <a href="${item.url}" target="_blank" rel="noopener noreferrer" style="
                  padding: 0.5rem;
                  color: #94a3b8;
                  text-decoration: none;
                  border-radius: 0.25rem;
                  transition: color 0.2s;
                " onmouseover="this.style.color='#f8fafc'" onmouseout="this.style.color='#94a3b8'">
                  🔗
                </a>
              </div>
              <h3 style="font-size: 1.125rem; font-weight: 600; color: #f8fafc; margin-bottom: 0.5rem;">
                ${item.title}
              </h3>
              <p style="font-size: 0.875rem; color: #94a3b8;">${item.source}</p>
            </div>

            <!-- Card Content -->
            <div style="padding: 1.5rem;">
              ${item.price ? `
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
                  <span style="font-size: 0.875rem; color: #94a3b8;">Price:</span>
                  <div style="display: flex; align-items: center; gap: 0.5rem;">
                    <span style="font-weight: 600; color: #f8fafc;">${formatCurrency(item.price)}</span>
                    ${item.change ? `
                      <span style="
                        padding: 0.25rem 0.5rem;
                        background: ${item.change >= 0 ? '#10b981' : '#ef4444'};
                        color: white;
                        border-radius: 0.25rem;
                        font-size: 0.75rem;
                        font-weight: 500;
                      ">
                        ${item.change >= 0 ? '+' : ''}${item.change}%
                      </span>
                    ` : ''}
                  </div>
                </div>
              ` : ''}

              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span style="font-size: 0.875rem; color: #94a3b8;">Last updated:</span>
                <span style="font-size: 0.875rem; color: #f8fafc;">${formatRelativeTime(item.lastUpdated)}</span>
              </div>
            </div>
          </div>
        `).join('')}
      </div>

      ${filteredData.length === 0 ? `
        <div style="text-align: center; padding: 3rem;">
          <div style="font-size: 3rem; margin-bottom: 1rem;">🔍</div>
          <h3 style="font-size: 1.125rem; font-weight: 500; color: #f8fafc; margin-bottom: 0.5rem;">No data found</h3>
          <p style="color: #94a3b8;">
            ${filter === 'all' ? 'No market data available.' : `No ${(filter as string).replace('-', ' ')} data found.`}
          </p>
        </div>
      ` : ''}

      <!-- Bright Data Attribution -->
      <div style="
        margin-top: 2rem;
        padding: 1rem;
        background: rgba(217, 119, 6, 0.1);
        border: 1px solid rgba(217, 119, 6, 0.3);
        border-radius: 0.75rem;
      ">
        <div style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; color: #94a3b8;">
          <span style="color: #f59e0b;">⚠️</span>
          <span>
            Market data powered by <strong style="color: #f59e0b;">Bright Data</strong> -
            Real-time web scraping and data collection platform
          </span>
        </div>
      </div>

      <!-- Setup Notice -->
      <div style="
        background: rgba(217, 119, 6, 0.1);
        border: 1px solid rgba(217, 119, 6, 0.3);
        border-radius: 0.75rem;
        padding: 1.5rem;
        text-align: center;
        margin-top: 2rem;
      ">
        <div style="font-size: 1.125rem; font-weight: 600; color: #f59e0b; margin-bottom: 0.5rem;">
          🚀 Bright Data Analytics Ready
        </div>
        <div style="color: #94a3b8; margin-bottom: 1rem;">
          Install dependencies to enable real-time market data collection and analysis
        </div>
        <div style="
          background: rgba(15, 23, 42, 0.8);
          border-radius: 0.5rem;
          padding: 0.75rem;
          font-family: monospace;
          font-size: 0.875rem;
          color: #10b981;
        ">
          npm install && npm run dev
        </div>
      </div>
    </div>
  `
}
