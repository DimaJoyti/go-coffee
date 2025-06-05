package main

import (
	"net/http"
	"strconv"

	"github.com/DimaJoyti/go-coffee/internal/defi"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// handleGetTokenPrice handles token price requests
func handleGetTokenPrice(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req defi.GetTokenPriceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := service.GetTokenPrice(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// handleGetSwapQuote handles swap quote requests
func handleGetSwapQuote(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req defi.GetSwapQuoteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := service.GetSwapQuote(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// handleExecuteSwap handles swap execution requests
func handleExecuteSwap(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req defi.ExecuteSwapRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := service.ExecuteSwap(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// handleGetLiquidityPools handles liquidity pool requests
func handleGetLiquidityPools(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &defi.GetLiquidityPoolsRequest{
			Chain:    defi.Chain(c.Query("chain")),
			Protocol: defi.ProtocolType(c.Query("protocol")),
			Token0:   c.Query("token0"),
			Token1:   c.Query("token1"),
		}

		// Parse optional parameters
		if minTVL := c.Query("min_tvl"); minTVL != "" {
			if tvl, err := decimal.NewFromString(minTVL); err == nil {
				req.MinTVL = tvl
			}
		}

		if limit := c.Query("limit"); limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				req.Limit = l
			}
		}

		if offset := c.Query("offset"); offset != "" {
			if o, err := strconv.Atoi(offset); err == nil {
				req.Offset = o
			}
		}

		resp, err := service.GetLiquidityPools(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// handleGetArbitrageOpportunities handles arbitrage opportunity requests
func handleGetArbitrageOpportunities(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		opportunities, err := service.GetArbitrageOpportunities(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"opportunities": opportunities,
			"count":         len(opportunities),
		})
	}
}

// handleGetYieldOpportunities handles yield farming opportunity requests
func handleGetYieldOpportunities(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 10 // default limit
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		opportunities, err := service.GetBestYieldOpportunities(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"opportunities": opportunities,
			"count":         len(opportunities),
		})
	}
}

// handleCreateTradingBot handles trading bot creation requests
func handleCreateTradingBot(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name     string                    `json:"name" binding:"required"`
			Strategy defi.TradingStrategyType `json:"strategy" binding:"required"`
			Config   defi.TradingBotConfig    `json:"config" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		bot, err := service.CreateTradingBot(c.Request.Context(), req.Name, req.Strategy, req.Config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, bot)
	}
}

// handleGetAllTradingBots handles requests to get all trading bots
func handleGetAllTradingBots(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		bots, err := service.GetAllTradingBots(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"bots":  bots,
			"count": len(bots),
		})
	}
}

// handleGetTradingBot handles requests to get a specific trading bot
func handleGetTradingBot(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		botID := c.Param("id")
		if botID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bot ID is required"})
			return
		}

		bot, err := service.GetTradingBot(c.Request.Context(), botID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, bot)
	}
}

// handleStartTradingBot handles requests to start a trading bot
func handleStartTradingBot(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		botID := c.Param("id")
		if botID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bot ID is required"})
			return
		}

		err := service.StartTradingBot(c.Request.Context(), botID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Trading bot started successfully"})
	}
}

// handleStopTradingBot handles requests to stop a trading bot
func handleStopTradingBot(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		botID := c.Param("id")
		if botID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bot ID is required"})
			return
		}

		err := service.StopTradingBot(c.Request.Context(), botID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Trading bot stopped successfully"})
	}
}

// handleDeleteTradingBot handles requests to delete a trading bot
func handleDeleteTradingBot(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		botID := c.Param("id")
		if botID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bot ID is required"})
			return
		}

		err := service.DeleteTradingBot(c.Request.Context(), botID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Trading bot deleted successfully"})
	}
}

// handleGetTradingBotPerformance handles requests to get trading bot performance
func handleGetTradingBotPerformance(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		botID := c.Param("id")
		if botID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bot ID is required"})
			return
		}

		performance, err := service.GetTradingBotPerformance(c.Request.Context(), botID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, performance)
	}
}

// handleGetMarketSignals handles requests to get market signals
func handleGetMarketSignals(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		signals, err := service.GetMarketSignals(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"signals": signals,
			"count":   len(signals),
		})
	}
}

// handleGetWhaleActivity handles requests to get whale activity
func handleGetWhaleActivity(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		whales, err := service.GetWhaleActivity(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"whales": whales,
			"count":  len(whales),
		})
	}
}

// handleGetTokenAnalysis handles requests to get token analysis
func handleGetTokenAnalysis(service *defi.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenAddress := c.Param("address")
		if tokenAddress == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token address is required"})
			return
		}

		analysis, err := service.GetTokenAnalysis(c.Request.Context(), tokenAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, analysis)
	}
}
