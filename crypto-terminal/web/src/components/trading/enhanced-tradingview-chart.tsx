import React, { useEffect, useRef, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { 
  BarChart3, 
  TrendingUp, 
  Settings, 
  Maximize2,
  Volume2,
  Activity,
  LineChart,
  CandlestickChart,
  AreaChart,
  Layers,
  PenTool,
  Target,
  Zap,
  RefreshCw,
  TrendingDown,
  DollarSign
} from "lucide-react";

interface EnhancedTradingViewChartProps {
  symbol?: string;
  interval?: string;
  theme?: "light" | "dark";
  height?: number;
  showToolbar?: boolean;
  showVolumeProfile?: boolean;
  showDrawingTools?: boolean;
  showIndicators?: boolean;
  className?: string;
  onSymbolChange?: (symbol: string) => void;
  onIntervalChange?: (interval: string) => void;
  onPriceAlert?: (price: number) => void;
}

interface TechnicalIndicator {
  id: string;
  name: string;
  enabled: boolean;
  params?: Record<string, any>;
}

interface DrawingTool {
  id: string;
  name: string;
  icon: React.ReactNode;
  active: boolean;
}

export const EnhancedTradingViewChart: React.FC<EnhancedTradingViewChartProps> = ({
  symbol = "BTCUSD",
  interval = "1D",
  theme = "dark",
  height = 600,
  showToolbar = true,
  showVolumeProfile = false,
  showDrawingTools = true,
  showIndicators = true,
  className,
  onSymbolChange,
  onIntervalChange,
  onPriceAlert,
}) => {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [chartType, setChartType] = useState<"candlestick" | "line" | "area">("candlestick");
  const [activeInterval, setActiveInterval] = useState(interval);
  const [isLoading, setIsLoading] = useState(true);
  const [currentPrice, setCurrentPrice] = useState(43245.75);
  const [priceChange, setPriceChange] = useState(1250.75);
  const [priceChangePercent, setPriceChangePercent] = useState(2.98);

  // Technical indicators state
  const [indicators, setIndicators] = useState<TechnicalIndicator[]>([
    { id: "sma", name: "SMA (20)", enabled: false },
    { id: "ema", name: "EMA (12)", enabled: false },
    { id: "rsi", name: "RSI (14)", enabled: false },
    { id: "macd", name: "MACD", enabled: false },
    { id: "bb", name: "Bollinger Bands", enabled: false },
    { id: "volume", name: "Volume", enabled: true },
  ]);

  // Drawing tools state
  const [drawingTools, setDrawingTools] = useState<DrawingTool[]>([
    { id: "trend", name: "Trend Line", icon: <LineChart className="w-4 h-4" />, active: false },
    { id: "horizontal", name: "Horizontal Line", icon: <BarChart3 className="w-4 h-4" />, active: false },
    { id: "rectangle", name: "Rectangle", icon: <Layers className="w-4 h-4" />, active: false },
    { id: "fibonacci", name: "Fibonacci", icon: <Activity className="w-4 h-4" />, active: false },
  ]);

  const intervals = [
    { label: "1m", value: "1" },
    { label: "5m", value: "5" },
    { label: "15m", value: "15" },
    { label: "1h", value: "60" },
    { label: "4h", value: "240" },
    { label: "1D", value: "1D" },
    { label: "1W", value: "1W" },
    { label: "1M", value: "1M" },
  ];

  // Initialize TradingView widget
  useEffect(() => {
    if (!chartContainerRef.current) return;

    setIsLoading(true);

    // Clear previous widget
    chartContainerRef.current.innerHTML = "";

    // Create TradingView widget script
    const script = document.createElement("script");
    script.src = "https://s3.tradingview.com/external-embedding/embed-widget-advanced-chart.js";
    script.type = "text/javascript";
    script.async = true;
    script.onload = () => setIsLoading(false);
    script.innerHTML = JSON.stringify({
      autosize: true,
      symbol: symbol,
      interval: activeInterval,
      timezone: "Etc/UTC",
      theme: theme,
      style: chartType === "candlestick" ? "1" : chartType === "line" ? "2" : "3",
      locale: "en",
      toolbar_bg: theme === "dark" ? "#1a1a1a" : "#f1f3f6",
      enable_publishing: false,
      allow_symbol_change: true,
      container_id: "enhanced_tradingview_chart",
      hide_top_toolbar: !showToolbar,
      hide_legend: false,
      save_image: false,
      studies: [
        ...(showVolumeProfile ? ["Volume Profile@tv-volumebyprice"] : []),
        ...indicators.filter(i => i.enabled).map(i => {
          switch (i.id) {
            case "sma": return "Moving Average@tv-basicstudies";
            case "ema": return "Moving Average Exponential@tv-basicstudies";
            case "rsi": return "Relative Strength Index@tv-basicstudies";
            case "macd": return "MACD@tv-basicstudies";
            case "bb": return "Bollinger Bands@tv-basicstudies";
            case "volume": return "Volume@tv-basicstudies";
            default: return "";
          }
        }).filter(Boolean)
      ],
      overrides: {
        "paneProperties.background": theme === "dark" ? "#0a0a0a" : "#ffffff",
        "paneProperties.vertGridProperties.color": theme === "dark" ? "#1a1a1a" : "#e1e1e1",
        "paneProperties.horzGridProperties.color": theme === "dark" ? "#1a1a1a" : "#e1e1e1",
        "symbolWatermarkProperties.transparency": 90,
        "scalesProperties.textColor": theme === "dark" ? "#d1d5db" : "#374151",
        "mainSeriesProperties.candleStyle.upColor": "#10b981",
        "mainSeriesProperties.candleStyle.downColor": "#ef4444",
        "mainSeriesProperties.candleStyle.borderUpColor": "#10b981",
        "mainSeriesProperties.candleStyle.borderDownColor": "#ef4444",
        "mainSeriesProperties.candleStyle.wickUpColor": "#10b981",
        "mainSeriesProperties.candleStyle.wickDownColor": "#ef4444",
      },
      studies_overrides: {
        "volume.volume.color.0": "#ef4444",
        "volume.volume.color.1": "#10b981",
        "volume.volume.transparency": 65,
      },
      loading_screen: {
        backgroundColor: theme === "dark" ? "#0a0a0a" : "#ffffff",
        foregroundColor: theme === "dark" ? "#d1d5db" : "#374151",
      },
      disabled_features: [
        "use_localstorage_for_settings",
        "volume_force_overlay",
        "create_volume_indicator_by_default"
      ],
      enabled_features: [
        "study_templates",
        "side_toolbar_in_fullscreen_mode",
        "header_in_fullscreen_mode"
      ]
    });

    chartContainerRef.current.appendChild(script);

    return () => {
      if (chartContainerRef.current) {
        chartContainerRef.current.innerHTML = "";
      }
    };
  }, [symbol, activeInterval, theme, chartType, showToolbar, showVolumeProfile, indicators]);

  const handleIntervalChange = useCallback((newInterval: string) => {
    setActiveInterval(newInterval);
    onIntervalChange?.(newInterval);
  }, [onIntervalChange]);

  const handleChartTypeChange = useCallback((type: "candlestick" | "line" | "area") => {
    setChartType(type);
  }, []);

  const toggleIndicator = useCallback((indicatorId: string) => {
    setIndicators(prev => prev.map(indicator => 
      indicator.id === indicatorId 
        ? { ...indicator, enabled: !indicator.enabled }
        : indicator
    ));
  }, []);

  const toggleDrawingTool = useCallback((toolId: string) => {
    setDrawingTools(prev => prev.map(tool => ({
      ...tool,
      active: tool.id === toolId ? !tool.active : false
    })));
  }, []);

  const handleFullscreen = useCallback(() => {
    setIsFullscreen(!isFullscreen);
  }, [isFullscreen]);

  return (
    <Card className={`trading-card-glass ${className} ${isFullscreen ? 'fixed inset-0 z-50' : ''}`}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <CardTitle className="text-lg">{symbol}</CardTitle>
            <div className="flex items-center space-x-2">
              <Badge variant="outline" className="text-sm font-mono">
                ${currentPrice.toLocaleString()}
              </Badge>
              <Badge 
                variant={priceChange >= 0 ? "default" : "destructive"}
                className={`text-sm ${priceChange >= 0 ? 'bg-green-600' : 'bg-red-600'}`}
              >
                {priceChange >= 0 ? '+' : ''}{priceChangePercent.toFixed(2)}%
              </Badge>
            </div>
          </div>
          
          <div className="flex items-center space-x-2">
            <Button variant="ghost" size="sm" onClick={handleFullscreen}>
              <Maximize2 className="w-4 h-4" />
            </Button>
            <Button variant="ghost" size="sm">
              <Settings className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Chart Controls */}
        <div className="flex items-center justify-between">
          {/* Timeframe Selection */}
          <div className="flex items-center space-x-1">
            {intervals.map((int) => (
              <Button
                key={int.value}
                variant={activeInterval === int.value ? "default" : "ghost"}
                size="sm"
                onClick={() => handleIntervalChange(int.value)}
                className="h-7 px-2 text-xs"
              >
                {int.label}
              </Button>
            ))}
          </div>

          {/* Chart Type Selection */}
          <div className="flex items-center space-x-1">
            <Button
              variant={chartType === "candlestick" ? "default" : "ghost"}
              size="sm"
              onClick={() => handleChartTypeChange("candlestick")}
              className="h-7 px-2"
            >
              <CandlestickChart className="w-4 h-4" />
            </Button>
            <Button
              variant={chartType === "line" ? "default" : "ghost"}
              size="sm"
              onClick={() => handleChartTypeChange("line")}
              className="h-7 px-2"
            >
              <LineChart className="w-4 h-4" />
            </Button>
            <Button
              variant={chartType === "area" ? "default" : "ghost"}
              size="sm"
              onClick={() => handleChartTypeChange("area")}
              className="h-7 px-2"
            >
              <AreaChart className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Advanced Tools */}
        {(showDrawingTools || showIndicators) && (
          <Tabs defaultValue="indicators" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              {showIndicators && <TabsTrigger value="indicators">Indicators</TabsTrigger>}
              {showDrawingTools && <TabsTrigger value="drawing">Drawing</TabsTrigger>}
            </TabsList>
            
            {showIndicators && (
              <TabsContent value="indicators" className="mt-2">
                <div className="flex flex-wrap gap-1">
                  {indicators.map((indicator) => (
                    <Button
                      key={indicator.id}
                      variant={indicator.enabled ? "default" : "outline"}
                      size="sm"
                      onClick={() => toggleIndicator(indicator.id)}
                      className="h-6 px-2 text-xs"
                    >
                      {indicator.name}
                    </Button>
                  ))}
                </div>
              </TabsContent>
            )}
            
            {showDrawingTools && (
              <TabsContent value="drawing" className="mt-2">
                <div className="flex flex-wrap gap-1">
                  {drawingTools.map((tool) => (
                    <Button
                      key={tool.id}
                      variant={tool.active ? "default" : "outline"}
                      size="sm"
                      onClick={() => toggleDrawingTool(tool.id)}
                      className="h-6 px-2 text-xs"
                    >
                      {tool.icon}
                      <span className="ml-1">{tool.name}</span>
                    </Button>
                  ))}
                </div>
              </TabsContent>
            )}
          </Tabs>
        )}
      </CardHeader>
      
      <CardContent className="p-0">
        <div 
          ref={chartContainerRef}
          className="tradingview-widget-container relative"
          style={{ height: `${height}px` }}
        >
          <div 
            id="enhanced_tradingview_chart"
            className="tradingview-widget-container__widget"
            style={{ height: "100%", width: "100%" }}
          />
          
          {/* Loading State */}
          {isLoading && (
            <div className="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm">
              <div className="text-center space-y-2">
                <RefreshCw className="animate-spin h-8 w-8 mx-auto text-primary" />
                <p className="text-sm text-muted-foreground">Loading advanced chart...</p>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};
