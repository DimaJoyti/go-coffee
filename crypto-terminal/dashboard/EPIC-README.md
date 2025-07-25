# 🚀 Epic Crypto Terminal - Next.js Dashboard

An **epic**, professional-grade cryptocurrency trading terminal built with Next.js 14, featuring real-time market data, advanced charting, and modern UI/UX design.

![Epic Crypto Terminal](https://img.shields.io/badge/Epic-Crypto%20Terminal-blue?style=for-the-badge&logo=bitcoin)
![Next.js](https://img.shields.io/badge/Next.js-14-black?style=for-the-badge&logo=next.js)
![TypeScript](https://img.shields.io/badge/TypeScript-5.4-blue?style=for-the-badge&logo=typescript)
![Tailwind CSS](https://img.shields.io/badge/Tailwind-3.4-38B2AC?style=for-the-badge&logo=tailwind-css)

## ✨ Epic Features

### 🎨 **Epic UI/UX Design**
- **Glassmorphism Effects**: Modern glass-like UI elements with backdrop blur
- **Advanced Animations**: Framer Motion powered smooth transitions and micro-interactions
- **Dark/Light Theme**: Automatic system theme detection with manual override
- **Responsive Design**: Pixel-perfect on desktop, tablet, and mobile devices
- **Epic Color Palette**: Bull/Bear colors with dynamic gradients and glow effects

### 📊 **Professional Trading Tools**
- **TradingView Integration**: Full-featured charts with technical indicators
- **Real-time Order Book**: Live bid/ask data with depth visualization
- **Market Heatmap**: Interactive cryptocurrency performance visualization
- **Price Alerts**: Custom price notifications with real-time monitoring
- **Portfolio Analytics**: Comprehensive portfolio tracking and performance metrics

### ⚡ **Real-time Features**
- **WebSocket Integration**: Live market data streaming
- **Auto-refresh**: Smart data fetching with optimized intervals
- **Live Notifications**: Toast notifications for important market events
- **Connection Status**: Real-time connection monitoring with visual indicators

### 🛡️ **Enterprise-Grade Architecture**
- **Type Safety**: Full TypeScript implementation with strict typing
- **State Management**: Zustand for lightweight, scalable state management
- **Data Fetching**: TanStack Query for server state management and caching
- **Error Handling**: Comprehensive error boundaries and fallback UI
- **Performance**: Optimized rendering with React 18 concurrent features

## 🛠️ **Epic Tech Stack**

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Framework** | Next.js | 14.2+ | React framework with App Router |
| **Language** | TypeScript | 5.4+ | Type-safe development |
| **Styling** | Tailwind CSS | 3.4+ | Utility-first CSS framework |
| **UI Library** | Radix UI | Latest | Headless UI primitives |
| **Animations** | Framer Motion | 11+ | Advanced animations and gestures |
| **State** | Zustand | 4.5+ | Lightweight state management |
| **Data Fetching** | TanStack Query | 5.28+ | Server state management |
| **Charts** | Recharts | 2.12+ | Composable charting library |
| **Real-time** | Socket.io | 4.7+ | WebSocket communication |
| **Forms** | React Hook Form | 7.51+ | Performant forms with validation |
| **Validation** | Zod | 3.22+ | TypeScript-first schema validation |

## 🚀 **Quick Start**

### Prerequisites
- Node.js 18+ 
- npm/yarn/pnpm
- Go backend running on port 8090

### Installation

1. **Clone and navigate**
   ```bash
   git clone <repository-url>
   cd crypto-terminal/dashboard
   ```

2. **Install dependencies**
   ```bash
   npm install
   # or
   yarn install
   # or
   pnpm install
   ```

3. **Environment setup**
   ```bash
   cp .env.example .env.local
   ```
   
   Edit `.env.local`:
   ```env
   NEXT_PUBLIC_API_URL=http://localhost:8090
   NEXT_PUBLIC_WS_URL=ws://localhost:8090
   NEXT_PUBLIC_TRADINGVIEW_API_KEY=your_key_here
   ```

4. **Start development server**
   ```bash
   npm run dev
   ```

5. **Open your browser**
   Navigate to `http://localhost:3001`

## 📁 **Epic Project Structure**

```
src/
├── app/                    # Next.js App Router
│   ├── globals.css        # Global styles with CSS variables
│   ├── layout.tsx         # Root layout with providers
│   └── page.tsx           # Homepage with Epic Dashboard
├── components/
│   ├── dashboard/         # Epic dashboard components
│   │   └── epic-crypto-dashboard.tsx
│   ├── layout/            # Layout components
│   ├── market/            # Market data components
│   │   ├── market-heatmap.tsx
│   │   └── news-widget.tsx
│   ├── portfolio/         # Portfolio management
│   ├── trading/           # Trading components
│   │   ├── tradingview-chart.tsx
│   │   ├── order-book.tsx
│   │   └── price-alerts.tsx
│   └── ui/                # Reusable UI components
│       ├── button.tsx     # Enhanced button with epic variants
│       ├── card.tsx       # Card with glass/epic variants
│       └── badge.tsx      # Badge with trading-specific variants
├── hooks/                 # Custom React hooks
├── lib/
│   ├── api.ts            # API client with interceptors
│   └── utils.ts          # Utility functions (enhanced)
├── stores/               # Zustand stores
└── types/                # TypeScript definitions
```

## 🎨 **Epic Design System**

### Color Palette
```css
/* Trading Colors */
--bull: 142 76% 36%        /* Green for gains */
--bear: 0 84% 60%          /* Red for losses */
--neutral: 240 3.8% 46.1%  /* Gray for neutral */

/* Epic Effects */
--glass-bg: rgba(255, 255, 255, 0.1)
--glass-border: rgba(255, 255, 255, 0.2)
--glass-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37)
```

### Component Variants
- **Epic Button**: Gradient with shimmer effect
- **Glass Card**: Glassmorphism with backdrop blur
- **Trading Badge**: Bull/Bear color variants
- **Glow Effects**: Dynamic shadows and highlights

## 📊 **Epic Components**

### TradingView Chart
- Professional charting with TradingView integration
- Multiple chart types (candlestick, line, area)
- Time interval selection
- Volume profile support
- Fullscreen mode

### Order Book
- Real-time bid/ask data
- Depth visualization
- Price grouping
- Click-to-trade functionality

### Market Heatmap
- Interactive cryptocurrency grid
- Size based on market cap
- Color intensity based on price change
- Hover details with market data

### Price Alerts
- Custom price notifications
- Real-time monitoring
- Visual progress indicators
- Alert management

## 🔧 **Available Scripts**

```bash
# Development
npm run dev              # Start development server
npm run build           # Build for production
npm run start           # Start production server

# Code Quality
npm run lint            # Run ESLint
npm run lint:fix        # Fix ESLint issues
npm run type-check      # TypeScript type checking
npm run format          # Format code with Prettier

# Testing
npm run test            # Run tests
npm run test:watch      # Run tests in watch mode
```

## 🌐 **API Integration**

The dashboard connects to a Go backend API:

```typescript
// API endpoints
GET  /api/v1/market/prices     # Market data
GET  /api/v1/portfolio         # Portfolio data
WS   /ws/market               # Real-time market updates
WS   /ws/portfolio            # Real-time portfolio updates
```

## 🚀 **Deployment**

### Production Build
```bash
npm run build
npm run start
```

### Docker Deployment
```bash
docker build -t epic-crypto-terminal .
docker run -p 3001:3001 epic-crypto-terminal
```

### Environment Variables (Production)
```env
NEXT_PUBLIC_API_URL=https://your-api-domain.com
NEXT_PUBLIC_WS_URL=wss://your-api-domain.com
NEXT_PUBLIC_TRADINGVIEW_API_KEY=your_production_key
```

## 🎯 **Performance Optimizations**

- **Code Splitting**: Automatic route-based splitting
- **Image Optimization**: Next.js Image component
- **Bundle Analysis**: Webpack bundle analyzer
- **Caching**: Aggressive caching with TanStack Query
- **Lazy Loading**: Component-level lazy loading
- **Memoization**: React.memo and useMemo optimizations

## 🤝 **Contributing**

1. Fork the repository
2. Create an epic feature branch: `git checkout -b feature/epic-feature`
3. Commit your changes: `git commit -m 'Add epic feature'`
4. Push to the branch: `git push origin feature/epic-feature`
5. Open a Pull Request

## 📄 **License**

MIT License - see [LICENSE](LICENSE) file for details.

## 🆘 **Support**

- 📧 Email: support@epic-crypto-terminal.com
- 💬 Discord: [Join our community](https://discord.gg/epic-crypto)
- 📖 Documentation: [Full docs](https://docs.epic-crypto-terminal.com)
- 🐛 Issues: [GitHub Issues](https://github.com/your-repo/issues)

---

**Built with ❤️ and ⚡ for the ultimate crypto trading experience**
