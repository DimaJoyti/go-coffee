import React from 'react';
import { Route, BrowserRouter as Router, Routes } from 'react-router-dom';
import './App.css';
import { EpicCryptoDashboard } from './components/dashboard/epic-crypto-dashboard';
import { MarketHeatmap } from './components/market/market-heatmap';
import { OrderBook } from './components/trading/order-book';
import { TradingViewChart } from './components/trading/tradingview-chart';
import { MarketDataTable } from './components/ui/market-data-table';
import { PriceChart } from './components/ui/price-chart';

// Import CSS for shadcn-ui
import './index.css';

function App() {
  return (
    <Router>
      <div className="App min-h-screen bg-background text-foreground">
        <Routes>
          <Route path="/" element={<EpicCryptoDashboard />} />
          <Route path="/dashboard" element={<EpicCryptoDashboard />} />
          <Route path="/markets" element={<EpicCryptoDashboard />} />
          <Route path="/trading" element={<EpicCryptoDashboard />} />
          <Route path="/portfolio" element={<EpicCryptoDashboard />} />
          <Route path="/analytics" element={<EpicCryptoDashboard />} />
          <Route path="/arbitrage" element={<EpicCryptoDashboard />} />
          <Route path="/risk" element={<EpicCryptoDashboard />} />
          <Route path="/settings" element={<EpicCryptoDashboard />} />
          <Route path="/notifications" element={<EpicCryptoDashboard />} />
          <Route path="/help" element={<EpicCryptoDashboard />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
