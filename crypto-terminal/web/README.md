# Epic Crypto Terminal - Frontend

A modern, professional cryptocurrency trading terminal built with React, TypeScript, and shadcn/ui components. This epic UI provides real-time market data, advanced charting, portfolio management, and risk analysis tools.

## 🚀 Features

### 🎨 Epic UI Components
- **Modern Design**: Built with shadcn/ui for a professional, consistent look
- **Dark/Light Theme**: Automatic theme switching with system preference
- **Responsive Layout**: Works perfectly on desktop, tablet, and mobile
- **Glass Effects**: Modern glassmorphism design elements
- **Smooth Animations**: Framer Motion powered transitions

### 📊 Advanced Market Data
- **Real-time Price Updates**: Live cryptocurrency prices and market data
- **Interactive Heatmap**: Visual representation of market performance
- **TradingView Integration**: Professional charting with technical indicators
- **Order Book**: Real-time bid/ask data with depth visualization
- **Market Analysis**: Comprehensive market intelligence and sentiment

### 💼 Portfolio Management
- **Portfolio Overview**: Complete portfolio tracking and analytics
- **Performance Charts**: Historical performance visualization
- **Asset Allocation**: Pie charts and allocation breakdowns
- **P&L Tracking**: Profit and loss analysis with detailed metrics
- **Holdings Management**: Add, remove, and manage crypto holdings

### ⚡ Trading Features
- **Advanced Charts**: Multiple chart types (candlestick, line, area)
- **Technical Indicators**: Full suite of trading indicators
- **Order Management**: Place and manage trading orders
- **Price Alerts**: Custom price and volume alerts
- **Trading Signals**: AI-powered trading recommendations

### 🛡️ Risk Management
- **Risk Metrics**: VaR, Sharpe ratio, beta, and volatility analysis
- **Drawdown Analysis**: Maximum drawdown tracking
- **Risk Alerts**: Real-time risk notifications
- **Portfolio Beta**: Market correlation analysis
- **Stress Testing**: Portfolio stress test scenarios

### 🔄 Real-time Features
- **WebSocket Integration**: Real-time data streaming
- **Live Updates**: Automatic price and portfolio updates
- **Notifications**: Toast notifications for important events
- **Connection Status**: Real-time connection monitoring

## 🛠️ Technology Stack

- **React 18**: Modern React with hooks and concurrent features
- **TypeScript**: Full type safety and developer experience
- **shadcn/ui**: Modern, accessible UI component library
- **Tailwind CSS**: Utility-first CSS framework
- **Radix UI**: Headless UI primitives
- **Recharts**: Powerful charting library
- **TradingView**: Professional trading charts
- **Framer Motion**: Smooth animations and transitions
- **Zustand**: Lightweight state management
- **React Query**: Server state management
- **Socket.io**: Real-time communication

## 📦 Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd crypto-terminal/web
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env.local
   # Edit .env.local with your API keys
   ```

4. **Start the development server**
   ```bash
   npm start
   ```

5. **Open your browser**
   Navigate to `http://localhost:3000`

## 🎯 Component Architecture

### Core Components

#### Layout Components
- `DashboardLayout`: Main application layout with sidebar and header
- `Sidebar`: Collapsible navigation sidebar with menu items
- `Header`: Top navigation with user menu and market tickers

#### UI Components
- `TradingCard`: Specialized cards for displaying trading metrics
- `MarketDataTable`: Advanced table for market data with sorting and filtering
- `PriceChart`: Customizable price charts with multiple data sources
- `Button`, `Card`, `Badge`: Core UI components from shadcn/ui

#### Trading Components
- `OrderBook`: Real-time order book with depth visualization
- `TradingViewChart`: Integrated TradingView charts
- `MarketHeatmap`: Interactive market performance heatmap

#### Portfolio Components
- `PortfolioOverview`: Complete portfolio dashboard
- `RiskManagement`: Risk analysis and management tools

### State Management

The application uses a combination of:
- **Zustand**: For global application state
- **React Query**: For server state and caching
- **React Context**: For theme and user preferences
- **Local State**: For component-specific state

### Styling System

- **Tailwind CSS**: Utility-first styling
- **CSS Variables**: Dynamic theming support
- **Custom Components**: Consistent design system
- **Responsive Design**: Mobile-first approach

## 🔧 Configuration

### Theme Configuration
The application supports both light and dark themes with automatic system detection:

```typescript
// tailwind.config.js
module.exports = {
  darkMode: ["class"],
  theme: {
    extend: {
      colors: {
        // Custom color palette for trading
        bull: "#10b981",
        bear: "#ef4444",
        neutral: "#6b7280",
      }
    }
  }
}
```

### API Configuration
Configure API endpoints in your environment file:

```env
REACT_APP_API_URL=http://localhost:8090
REACT_APP_WS_URL=ws://localhost:8090
REACT_APP_TRADINGVIEW_API_KEY=your_key_here
```

## 📱 Responsive Design

The application is fully responsive with breakpoints:
- **Mobile**: < 768px
- **Tablet**: 768px - 1024px
- **Desktop**: > 1024px

Key responsive features:
- Collapsible sidebar on mobile
- Adaptive grid layouts
- Touch-friendly interactions
- Optimized chart rendering

## 🎨 Design System

### Color Palette
- **Primary**: Blue (#3b82f6)
- **Bull/Green**: #10b981
- **Bear/Red**: #ef4444
- **Neutral**: #6b7280
- **Background**: Dynamic based on theme

### Typography
- **Font Family**: Inter (primary), JetBrains Mono (monospace)
- **Font Sizes**: Responsive scale from 0.75rem to 3rem
- **Font Weights**: 300, 400, 500, 600, 700

### Spacing
- **Base Unit**: 0.25rem (4px)
- **Scale**: 1, 2, 3, 4, 6, 8, 12, 16, 20, 24, 32, 40, 48, 64

## 🚀 Performance Optimizations

- **Code Splitting**: Lazy loading of route components
- **Memoization**: React.memo and useMemo for expensive calculations
- **Virtual Scrolling**: For large data tables
- **Image Optimization**: Lazy loading and WebP support
- **Bundle Analysis**: Webpack bundle analyzer integration

## 🧪 Testing

```bash
# Run unit tests
npm test

# Run tests with coverage
npm run test:coverage

# Run e2e tests
npm run test:e2e
```

## 📈 Deployment

### Production Build
```bash
npm run build
```

### Docker Deployment
```bash
docker build -t crypto-terminal-web .
docker run -p 3000:3000 crypto-terminal-web
```

### Environment Variables
Required environment variables for production:
- `REACT_APP_API_URL`
- `REACT_APP_WS_URL`
- `REACT_APP_TRADINGVIEW_API_KEY`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new features
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Join our Discord community

---

Built with ❤️ using modern web technologies for the ultimate crypto trading experience.
