# 🎨 Go Coffee Epic UI

## 🌟 Overview

Epic UI for Go Coffee is a modern web interface that unites all components of the Web3 coffee ecosystem with powerful web scraping capabilities through Bright Data MCP.

## 🏗️ Architecture

### Frontend
- **Next.js 14** - React framework with App Router
- **TypeScript** - Type safety and reliability
- **TailwindCSS** - Utility-first CSS framework
- **Shadcn/ui** - Modern UI components
- **Recharts** - Interactive charts
- **Framer Motion** - Animations and transitions

### Backend
- **Go 1.22+** - API server
- **Gin** - HTTP framework
- **WebSocket** - Real-time updates
- **Bright Data MCP** - Web scraping integration

## 📊 Main Sections

### 1. 📈 Main Dashboard
- General metrics and KPIs
- Real-time statistics
- Quick actions and notifications

### 2. ☕ Coffee Orders
- Order management
- Inventory and warehouse
- Location map

### 3. 🌐 DeFi Portfolio
- Cryptocurrency balances
- Trading strategies
- P&L analytics

### 4. 🤖 AI Agents
- Status and monitoring
- Activity logs
- Agent configuration

### 5. 🔍 Bright Data Analytics
- Market data
- Competitive analysis
- Industry news

### 6. 📊 Analytics
- Reports and metrics
- Forecasts
- Data export

## 🚀 Features

- **Real-time** - WebSocket connections for instant updates
- **Responsive design** - Optimized for all devices
- **PWA support** - Works as a native app
- **Dark/light themes** - Customizable settings
- **AI insights** - Smart recommendations and forecasts
- **Interactivity** - Drag & drop, customizable widgets

## 🛠️ Development

### Quick Start
```bash
# Clone and install
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/web-ui

# Copy configuration
cp .env.example .env

# Install dependencies
make install

# Run in development mode
make dev
```

### Installation
```bash
# Frontend
cd frontend
npm install

# Backend
cd backend
go mod tidy
```

### Running
```bash
# Development mode (recommended)
make dev

# Or separately:
# Frontend (port 3000)
cd frontend && npm run dev

# Backend (port 8090)
cd backend && go run cmd/web-ui-service/main.go
```

### Docker
```bash
# Quick start
make start

# Or manually:
docker-compose -f docker-compose.ui.yml up --build

# Stop
make stop
```

### Bright Data MCP Integration
```bash
# Test Bright Data MCP
go run test-bright-data-mcp.go

# Set API token
export BRIGHT_DATA_API_TOKEN="your_token_here"
```

## 📡 API Endpoints

- `GET /api/v1/dashboard/metrics` - General metrics
- `GET /api/v1/coffee/orders` - Coffee orders
- `GET /api/v1/defi/portfolio` - DeFi portfolio
- `GET /api/v1/agents/status` - AI agents status
- `GET /api/v1/scraping/data` - Bright Data analytics
- `WS /ws/realtime` - WebSocket for real-time

## ✨ Epic UI Features

### 🎨 Design and UX
- **Modern design** - Minimalist and elegant interface
- **Responsive** - Adaptive design for all devices
- **Dark/light themes** - Automatic switching and manual settings
- **Animations** - Smooth transitions and micro-animations
- **Accessibility** - Screen reader and keyboard navigation support

### ⚡ Performance
- **Real-time updates** - WebSocket connections for instant data
- **Optimized queries** - SWR for caching and automatic updates
- **Lazy loading** - Components load on demand
- **Code splitting** - Bundle size optimization

### 🔗 Integrations
- **Bright Data MCP** - Powerful web scraping for market data
- **WebSocket** - Real-time communication
- **REST API** - Full-featured API for all operations
- **Crypto APIs** - Integration with DeFi protocols

### 🛡️ Security
- **JWT Authentication** - Secure authentication
- **CORS configuration** - Protection from unwanted requests
- **Input validation** - Validation of all input data
- **Rate limiting** - Protection from abuse

## 🎯 Roadmap

### ✅ 1: Foundation (Completed)
- [x] Basic project structure
- [x] Next.js + TailwindCSS + Shadcn/ui setup
- [x] Go API server with Gin
- [x] WebSocket integration
- [x] Docker configuration

### ✅ 2: Core Components (Completed)
- [x] Main dashboard with metrics
- [x] Coffee orders interface
- [x] DeFi portfolio and trading
- [x] AI agents monitoring
- [x] Responsive design

### ✅ 3: Bright Data Integration (Completed)
- [x] MCP client for Bright Data
- [x] Web scraping services
- [x] Market data and competitive analysis
- [x] Automatic data updates

### 🚧 4: Enhancements (In Progress)
- [ ] Real charts with Recharts
- [ ] Extended analytics
- [ ] Push notifications
- [ ] Report export
- [ ] Mobile app (PWA)

### 🔮 5: Future Features
- [ ] AI-powered insights and recommendations
- [ ] Voice control
- [ ] AR/VR interfaces
- [ ] Blockchain integration
- [ ] Multi-language support

## 📄 License

MIT License - see [LICENSE](../LICENSE) file for details.
