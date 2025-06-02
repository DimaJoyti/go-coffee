# 🚀 Go Coffee Epic UI - Quick Start

## ⚡ Quick Launch (1 minute)

```bash
# 1. Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/web-ui

# 2. Run demonstration
./demo.sh

# 3. Open browser
# Frontend: http://localhost:3000
# Backend: http://localhost:8090
```

## 🎯 What You'll See

### 📊 Main Dashboard
- **Real-time metrics** - Orders, revenue, portfolio
- **Activity** - Live event feed
- **Charts** - Interactive diagrams

### ☕ Coffee Orders
- **Order management** - Create, update, track
- **Statuses** - Pending, Preparing, Ready, Completed
- **Filtering** - By status, location, time

### 💰 DeFi Portfolio
- **Cryptocurrency assets** - BTC, ETH, USDC, USDT
- **Trading strategies** - Arbitrage, Yield Farming, Grid Trading
- **P&L analytics** - Profits and losses

### 🤖 AI Agents
- **9 specialized agents** - From beverage invention to location coordination
- **Status monitoring** - Active, Inactive, Error, Maintenance
- **Performance metrics** - Uptime, completed tasks

### 🔍 Bright Data Analytics
- **Web scraping** - Automatic market data collection
- **Competitive analysis** - Starbucks, Dunkin', Costa prices
- **Industry news** - Trends and analytics
- **Social media** - Trend monitoring

### 📈 Analytics
- **Sales reports** - For different periods
- **Top products** - Most popular beverages
- **Location performance** - Branch comparison

## 🛠️ Management Commands

```bash
# Start services
./demo.sh --quick          # Quick start
make start                 # Docker start
make dev                   # Development mode

# Monitoring
make health                # Health check
make status                # Container status
make logs                  # View logs

# Stop
./demo.sh --stop           # Stop via demo
make stop                  # Stop via make

# Cleanup
make clean                 # Full cleanup
```

## 🌐 Endpoints

| Service | URL | Description |
|---------|-----|-------------|
| **Frontend** | http://localhost:3000 | Main interface |
| **Backend API** | http://localhost:8090 | REST API |
| **Health Check** | http://localhost:8090/health | Server status |
| **WebSocket** | ws://localhost:8090/ws/realtime | Real-time data |

## 📱 UI Features

### 🎨 Design
- ✅ **Responsive** - Works on all devices
- ✅ **Dark/Light themes** - Automatic switching
- ✅ **Animations** - Smooth Framer Motion transitions
- ✅ **Accessibility** - WCAG compliance

### ⚡ Performance
- ✅ **Real-time updates** - WebSocket connections
- ✅ **SWR caching** - Optimized queries
- ✅ **Code splitting** - Lazy loading components
- ✅ **PWA support** - Works as native app

### 🔗 Integrations
- ✅ **Bright Data MCP** - Web scraping
- ✅ **WebSocket** - Real-time communication
- ✅ **REST API** - Full-featured backend
- ✅ **Docker** - Containerization

## 🔧 Configuration

### Environment Variables
```bash
# Copy example configuration
cp .env.example .env

# Basic settings
NEXT_PUBLIC_API_URL=http://localhost:8090
NEXT_PUBLIC_WS_URL=ws://localhost:8090
BRIGHT_DATA_API_TOKEN=your_token_here
```

### Bright Data MCP
```bash
# Set API token
export BRIGHT_DATA_API_TOKEN="your_token_here"

# Test integration
go run test-bright-data-mcp.go
```

## 🚨 Troubleshooting

### Docker won't start
```bash
# Check Docker
docker --version
docker-compose --version

# Restart Docker
sudo systemctl restart docker
```

### Ports occupied
```bash
# Find processes on ports
lsof -i :3000
lsof -i :8090

# Kill processes
kill -9 <PID>
```

### Frontend won't load
```bash
# Check logs
make logs

# Rebuild containers
docker-compose -f docker-compose.ui.yml up --build --force-recreate
```

## 📞 Support

- 📧 **Email**: aws.inspiration@gmail.com
- 🐛 **Issues**: [GitHub Issues](https://github.com/DimaJoyti/go-coffee/issues)
- 📖 **Documentation**: [README.md](./README.md)

## 🎉 Next Steps

1. **Explore the interface** - Navigate through all sections
2. **Test the API** - Use Postman or curl
3. **Configure Bright Data** - Add your API token
4. **Customize** - Change themes and settings
5. **Develop** - Add new features

---

**🚀 Ready! Your Epic UI is launched and ready to use!**
