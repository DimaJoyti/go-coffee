# ☕ Coffee Trading Dashboard - Next.js

A modern, professional cryptocurrency trading dashboard built with Next.js 14, featuring coffee-themed trading strategies and real-time market data.

## 🚀 Features

### 📊 **Dashboard Overview**
- Real-time portfolio performance tracking
- Market overview with global statistics
- Coffee strategy performance metrics
- Live price updates via WebSocket

### ☕ **Coffee Trading Strategies**
- **Espresso**: Quick & strong momentum trading
- **Latte**: Smooth & balanced portfolio approach
- **Cold Brew**: Patient & steady long-term strategy
- **Cappuccino**: Rich & frothy high-frequency trading

### 💼 **Portfolio Management**
- Multi-portfolio support
- Real-time P&L tracking
- Asset allocation visualization
- Performance analytics with charts

### 📈 **Market Data**
- Live cryptocurrency prices
- Market heatmaps
- Top gainers/losers
- Fear & Greed Index
- Multi-exchange price aggregation

### ⚡ **Real-time Features**
- WebSocket integration for live updates
- Real-time notifications
- Signal alerts with confidence levels
- Trade execution notifications
- Risk management alerts

## 🛠 Tech Stack

- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript
- **Styling**: TailwindCSS + Shadcn/ui
- **State Management**: Zustand
- **Data Fetching**: TanStack Query (React Query)
- **Charts**: Recharts
- **WebSocket**: Socket.io Client
- **Notifications**: React Hot Toast
- **Icons**: Lucide React

## 🏗 Project Structure

```
src/
├── app/                    # Next.js App Router
│   ├── globals.css        # Global styles
│   ├── layout.tsx         # Root layout
│   └── page.tsx           # Home page
├── components/            # React components
│   ├── ui/                # Base UI components (Shadcn/ui)
│   ├── layout/            # Layout components
│   ├── dashboard/         # Dashboard components
│   ├── portfolio/         # Portfolio components
│   ├── trading/           # Trading components
│   ├── market/            # Market data components
│   └── realtime/          # Real-time update components
├── hooks/                 # Custom React hooks
│   └── use-websocket.ts   # WebSocket management
├── lib/                   # Utilities and configurations
│   ├── api.ts             # API client
│   └── utils.ts           # Utility functions
├── stores/                # Zustand stores
│   └── trading-store.ts   # Main trading state
└── types/                 # TypeScript definitions
    └── trading.ts         # Trading-related types
```

## 🚀 Getting Started

### Prerequisites
- Node.js 18+ 
- npm or yarn
- Go Coffee Trading Backend running on port 8090

### Installation

1. **Navigate to the dashboard directory**:
```bash
cd crypto-terminal/nextjs-dashboard
```

2. **Install dependencies**:
```bash
npm install
```

3. **Configure environment variables**:
```bash
cp .env.local.example .env.local
# Edit .env.local with your configuration
```

4. **Start the development server**:
```bash
npm run dev
```

5. **Open your browser**:
Navigate to [http://localhost:3001](http://localhost:3001)

## 🔧 Configuration

### Environment Variables

```env
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8090
NEXT_PUBLIC_WS_URL=ws://localhost:8090

# Application Configuration
NEXT_PUBLIC_APP_NAME="Coffee Trading Dashboard"
NEXT_PUBLIC_APP_VERSION="1.0.0"

# Feature Flags
NEXT_PUBLIC_ENABLE_DEVTOOLS=true
NEXT_PUBLIC_ENABLE_WEBSOCKET=true
NEXT_PUBLIC_ENABLE_NOTIFICATIONS=true
```

### Backend Integration

The dashboard integrates with the Go Coffee Trading backend via:

- **REST API**: `http://localhost:8090/api/v1/`
- **WebSocket**: `ws://localhost:8090/ws/coffee-trading`

Ensure the Go backend is running before starting the dashboard.

## 📱 Features in Detail

### Real-time Updates
- Live price feeds for major cryptocurrencies
- WebSocket connection with automatic reconnection
- Real-time portfolio value updates
- Signal alerts with coffee-themed notifications

### Coffee Strategy Management
- Visual strategy selector with coffee themes
- One-click strategy activation/deactivation
- Performance tracking per strategy
- Risk management settings

### Portfolio Analytics
- Interactive pie charts for allocation
- Performance bar charts
- P&L tracking with color coding
- Holdings table with real-time updates

### Market Overview
- Global market statistics
- Major cryptocurrency tracking
- Top gainers/losers lists
- Fear & Greed Index visualization

## 🎨 UI/UX Features

- **Dark/Light Theme**: Automatic theme switching
- **Responsive Design**: Works on desktop and mobile
- **Real-time Animations**: Smooth price update animations
- **Coffee Branding**: Consistent coffee theme throughout
- **Professional Layout**: Clean, modern trading interface

## 🔌 API Integration

The dashboard connects to these backend endpoints:

```typescript
// Portfolio Management
GET /api/v1/coffee-trading/portfolio
GET /api/v1/coffee-trading/portfolio/{id}/performance

// Coffee Strategies
GET /api/v1/coffee-trading/strategies
POST /api/v1/coffee-trading/coffee/{type}/start
POST /api/v1/coffee-trading/strategies/{id}/start

// Market Data
GET /api/v1/market/prices
GET /api/v1/market/overview
GET /api/v1/tradingview/market-data

// WebSocket Channels
- price_updates
- signal_alerts
- trade_executions
- portfolio_updates
- risk_alerts
```

## 🧪 Development

### Available Scripts

```bash
npm run dev          # Start development server
npm run build        # Build for production
npm run start        # Start production server
npm run lint         # Run ESLint
npm run type-check   # Run TypeScript checks
```

### Code Quality
- TypeScript for type safety
- ESLint for code linting
- Prettier for code formatting
- Tailwind CSS for consistent styling

## 🚀 Deployment

### Production Build
```bash
npm run build
npm run start
```

### Docker (Optional)
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 3001
CMD ["npm", "start"]
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is part of the Go Coffee Trading Platform.

---

**Happy Trading! ☕📈**
