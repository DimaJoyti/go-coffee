import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Portfolio from './pages/Portfolio';
import Alerts from './pages/Alerts';
import Markets from './pages/Markets';
import DeFi from './pages/DeFi';
import Settings from './pages/Settings';
import './App.css';

function App() {
  return (
    <Router>
      <div className="App">
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/portfolio" element={<Portfolio />} />
            <Route path="/markets" element={<Markets />} />
            <Route path="/alerts" element={<Alerts />} />
            <Route path="/defi" element={<DeFi />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </Layout>
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: '#1e293b',
              color: '#e2e8f0',
              border: '1px solid #334155',
            },
            success: {
              iconTheme: {
                primary: '#10b981',
                secondary: '#1e293b',
              },
            },
            error: {
              iconTheme: {
                primary: '#ef4444',
                secondary: '#1e293b',
              },
            },
          }}
        />
      </div>
    </Router>
  );
}

export default App;
