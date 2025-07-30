import { useEffect, useState, useCallback, useRef } from 'react';
import { 
  enhancedWebSocketService, 
  MarketDataUpdate, 
  OrderBookUpdate, 
  TradeUpdate, 
  TickerUpdate 
} from '../services/enhanced-websocket';

export interface UseEnhancedWebSocketOptions {
  autoConnect?: boolean;
  reconnectOnMount?: boolean;
}

export interface WebSocketState {
  isConnected: boolean;
  connectionState: string;
  error: string | null;
  lastUpdate: number;
}

export const useEnhancedWebSocket = (options: UseEnhancedWebSocketOptions = {}) => {
  const { autoConnect = true, reconnectOnMount = true } = options;
  
  const [state, setState] = useState<WebSocketState>({
    isConnected: false,
    connectionState: 'disconnected',
    error: null,
    lastUpdate: 0
  });

  const [marketData, setMarketData] = useState<Record<string, MarketDataUpdate>>({});
  const [orderBooks, setOrderBooks] = useState<Record<string, OrderBookUpdate>>({});
  const [trades, setTrades] = useState<Record<string, TradeUpdate[]>>({});
  const [tickers, setTickers] = useState<Record<string, TickerUpdate>>({});

  const isInitialized = useRef(false);

  // Event handlers
  const handleConnected = useCallback(() => {
    setState(prev => ({
      ...prev,
      isConnected: true,
      connectionState: 'connected',
      error: null,
      lastUpdate: Date.now()
    }));
  }, []);

  const handleDisconnected = useCallback((data: { reason: string }) => {
    setState(prev => ({
      ...prev,
      isConnected: false,
      connectionState: 'disconnected',
      error: data.reason,
      lastUpdate: Date.now()
    }));
  }, []);

  const handleError = useCallback((data: { message: string }) => {
    setState(prev => ({
      ...prev,
      error: data.message,
      lastUpdate: Date.now()
    }));
  }, []);

  const handleMarketData = useCallback((data: MarketDataUpdate) => {
    setMarketData(prev => ({
      ...prev,
      [data.symbol]: data
    }));
    setState(prev => ({ ...prev, lastUpdate: Date.now() }));
  }, []);

  const handleOrderBook = useCallback((data: OrderBookUpdate) => {
    setOrderBooks(prev => ({
      ...prev,
      [data.symbol]: data
    }));
    setState(prev => ({ ...prev, lastUpdate: Date.now() }));
  }, []);

  const handleTrade = useCallback((data: TradeUpdate) => {
    setTrades(prev => {
      const symbolTrades = prev[data.symbol] || [];
      const newTrades = [data, ...symbolTrades].slice(0, 100); // Keep last 100 trades
      return {
        ...prev,
        [data.symbol]: newTrades
      };
    });
    setState(prev => ({ ...prev, lastUpdate: Date.now() }));
  }, []);

  const handleTicker = useCallback((data: TickerUpdate) => {
    setTickers(prev => ({
      ...prev,
      [data.symbol]: data
    }));
    setState(prev => ({ ...prev, lastUpdate: Date.now() }));
  }, []);

  // Initialize WebSocket connection and event listeners
  useEffect(() => {
    if (isInitialized.current) return;
    isInitialized.current = true;

    // Set up event listeners
    enhancedWebSocketService.on('connected', handleConnected);
    enhancedWebSocketService.on('disconnected', handleDisconnected);
    enhancedWebSocketService.on('error', handleError);
    enhancedWebSocketService.on('marketData', handleMarketData);
    enhancedWebSocketService.on('orderBook', handleOrderBook);
    enhancedWebSocketService.on('trade', handleTrade);
    enhancedWebSocketService.on('ticker', handleTicker);

    // Auto-connect if enabled
    if (autoConnect) {
      enhancedWebSocketService.connect().catch(console.error);
    }

    // Update initial state
    setState(prev => ({
      ...prev,
      isConnected: enhancedWebSocketService.isConnected(),
      connectionState: enhancedWebSocketService.getConnectionState()
    }));

    return () => {
      enhancedWebSocketService.off('connected', handleConnected);
      enhancedWebSocketService.off('disconnected', handleDisconnected);
      enhancedWebSocketService.off('error', handleError);
      enhancedWebSocketService.off('marketData', handleMarketData);
      enhancedWebSocketService.off('orderBook', handleOrderBook);
      enhancedWebSocketService.off('trade', handleTrade);
      enhancedWebSocketService.off('ticker', handleTicker);
    };
  }, [
    autoConnect,
    handleConnected,
    handleDisconnected,
    handleError,
    handleMarketData,
    handleOrderBook,
    handleTrade,
    handleTicker
  ]);

  // Reconnect on mount if needed
  useEffect(() => {
    if (reconnectOnMount && !enhancedWebSocketService.isConnected()) {
      enhancedWebSocketService.connect().catch(console.error);
    }
  }, [reconnectOnMount]);

  // API methods
  const connect = useCallback(async () => {
    try {
      await enhancedWebSocketService.connect();
    } catch (error) {
      console.error('Failed to connect:', error);
    }
  }, []);

  const disconnect = useCallback(() => {
    enhancedWebSocketService.disconnect();
  }, []);

  const subscribeToMarketData = useCallback((symbols: string[]) => {
    enhancedWebSocketService.subscribeToMarketData(symbols);
  }, []);

  const subscribeToOrderBook = useCallback((symbol: string) => {
    enhancedWebSocketService.subscribeToOrderBook(symbol);
  }, []);

  const subscribeToTrades = useCallback((symbol: string) => {
    enhancedWebSocketService.subscribeToTrades(symbol);
  }, []);

  const subscribeToTicker = useCallback((symbols: string[]) => {
    enhancedWebSocketService.subscribeToTicker(symbols);
  }, []);

  const unsubscribeFromMarketData = useCallback((symbols: string[]) => {
    enhancedWebSocketService.unsubscribeFromMarketData(symbols);
  }, []);

  const unsubscribeFromOrderBook = useCallback((symbol: string) => {
    enhancedWebSocketService.unsubscribeFromOrderBook(symbol);
  }, []);

  const unsubscribeFromTrades = useCallback((symbol: string) => {
    enhancedWebSocketService.unsubscribeFromTrades(symbol);
  }, []);

  const unsubscribeFromTicker = useCallback((symbols: string[]) => {
    enhancedWebSocketService.unsubscribeFromTicker(symbols);
  }, []);

  // Utility functions
  const getMarketData = useCallback((symbol: string): MarketDataUpdate | null => {
    return marketData[symbol] || null;
  }, [marketData]);

  const getOrderBook = useCallback((symbol: string): OrderBookUpdate | null => {
    return orderBooks[symbol] || null;
  }, [orderBooks]);

  const getTrades = useCallback((symbol: string): TradeUpdate[] => {
    return trades[symbol] || [];
  }, [trades]);

  const getTicker = useCallback((symbol: string): TickerUpdate | null => {
    return tickers[symbol] || null;
  }, [tickers]);

  const getAllMarketData = useCallback(() => marketData, [marketData]);
  const getAllOrderBooks = useCallback(() => orderBooks, [orderBooks]);
  const getAllTrades = useCallback(() => trades, [trades]);
  const getAllTickers = useCallback(() => tickers, [tickers]);

  return {
    // Connection state
    state,
    isConnected: state.isConnected,
    connectionState: state.connectionState,
    error: state.error,
    lastUpdate: state.lastUpdate,

    // Connection methods
    connect,
    disconnect,

    // Subscription methods
    subscribeToMarketData,
    subscribeToOrderBook,
    subscribeToTrades,
    subscribeToTicker,
    unsubscribeFromMarketData,
    unsubscribeFromOrderBook,
    unsubscribeFromTrades,
    unsubscribeFromTicker,

    // Data access methods
    getMarketData,
    getOrderBook,
    getTrades,
    getTicker,
    getAllMarketData,
    getAllOrderBooks,
    getAllTrades,
    getAllTickers,

    // Raw data (for advanced use cases)
    marketData,
    orderBooks,
    trades,
    tickers
  };
};

// Specialized hooks for specific data types
export const useMarketData = (symbols: string[]) => {
  const { subscribeToMarketData, unsubscribeFromMarketData, getMarketData, getAllMarketData } = useEnhancedWebSocket();

  useEffect(() => {
    if (symbols.length > 0) {
      subscribeToMarketData(symbols);
      return () => unsubscribeFromMarketData(symbols);
    }
  }, [symbols, subscribeToMarketData, unsubscribeFromMarketData]);

  return {
    getMarketData,
    getAllMarketData
  };
};

export const useOrderBook = (symbol: string) => {
  const { subscribeToOrderBook, unsubscribeFromOrderBook, getOrderBook } = useEnhancedWebSocket();

  useEffect(() => {
    if (symbol) {
      subscribeToOrderBook(symbol);
      return () => unsubscribeFromOrderBook(symbol);
    }
  }, [symbol, subscribeToOrderBook, unsubscribeFromOrderBook]);

  return {
    orderBook: getOrderBook(symbol)
  };
};

export const useTrades = (symbol: string) => {
  const { subscribeToTrades, unsubscribeFromTrades, getTrades } = useEnhancedWebSocket();

  useEffect(() => {
    if (symbol) {
      subscribeToTrades(symbol);
      return () => unsubscribeFromTrades(symbol);
    }
  }, [symbol, subscribeToTrades, unsubscribeFromTrades]);

  return {
    trades: getTrades(symbol)
  };
};

export const useTicker = (symbols: string[]) => {
  const { subscribeToTicker, unsubscribeFromTicker, getTicker, getAllTickers } = useEnhancedWebSocket();

  useEffect(() => {
    if (symbols.length > 0) {
      subscribeToTicker(symbols);
      return () => unsubscribeFromTicker(symbols);
    }
  }, [symbols, subscribeToTicker, unsubscribeFromTicker]);

  return {
    getTicker,
    getAllTickers
  };
};
