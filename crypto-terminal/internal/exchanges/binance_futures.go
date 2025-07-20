package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// BinanceFuturesExtension extends BinanceClient with futures trading capabilities
type BinanceFuturesExtension struct {
	client *BinanceClient
}

// NewBinanceFuturesExtension creates a new futures extension
func NewBinanceFuturesExtension(client *BinanceClient) *BinanceFuturesExtension {
	return &BinanceFuturesExtension{
		client: client,
	}
}

// GetFuturesAccount gets futures account information
func (bf *BinanceFuturesExtension) GetFuturesAccount(ctx context.Context) (*FuturesAccount, error) {
	resp, err := bf.client.makeRequest(ctx, "GET", "/fapi/v2/account", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		TotalWalletBalance  string `json:"totalWalletBalance"`
		TotalUnrealizedPnL  string `json:"totalUnrealizedPnL"`
		TotalMarginBalance  string `json:"totalMarginBalance"`
		TotalPositionValue  string `json:"totalPositionValue"`
		AvailableBalance    string `json:"availableBalance"`
		MaxWithdrawAmount   string `json:"maxWithdrawAmount"`
		CanTrade            bool   `json:"canTrade"`
		CanDeposit          bool   `json:"canDeposit"`
		CanWithdraw         bool   `json:"canWithdraw"`
		UpdateTime          int64  `json:"updateTime"`
		Assets              []struct {
			Asset                  string `json:"asset"`
			WalletBalance          string `json:"walletBalance"`
			UnrealizedProfit       string `json:"unrealizedProfit"`
			MarginBalance          string `json:"marginBalance"`
			MaintMargin            string `json:"maintMargin"`
			InitialMargin          string `json:"initialMargin"`
			PositionInitialMargin  string `json:"positionInitialMargin"`
			OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
			MaxWithdrawAmount      string `json:"maxWithdrawAmount"`
			CrossWalletBalance     string `json:"crossWalletBalance"`
			CrossUnPnl             string `json:"crossUnPnl"`
			AvailableBalance       string `json:"availableBalance"`
		} `json:"assets"`
		Positions []struct {
			Symbol         string `json:"symbol"`
			PositionSide   string `json:"positionSide"`
			PositionAmt    string `json:"positionAmt"`
			EntryPrice     string `json:"entryPrice"`
			MarkPrice      string `json:"markPrice"`
			UnrealizedPnL  string `json:"unRealizedPnL"`
			Leverage       string `json:"leverage"`
			MarginType     string `json:"marginType"`
			IsolatedMargin string `json:"isolatedMargin"`
			UpdateTime     int64  `json:"updateTime"`
		} `json:"positions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to our format
	totalWalletBalance, _ := decimal.NewFromString(result.TotalWalletBalance)
	totalUnrealizedPnL, _ := decimal.NewFromString(result.TotalUnrealizedPnL)
	totalMarginBalance, _ := decimal.NewFromString(result.TotalMarginBalance)
	totalPositionValue, _ := decimal.NewFromString(result.TotalPositionValue)
	availableBalance, _ := decimal.NewFromString(result.AvailableBalance)
	maxWithdrawAmount, _ := decimal.NewFromString(result.MaxWithdrawAmount)

	assets := make([]FuturesAsset, len(result.Assets))
	for i, asset := range result.Assets {
		walletBalance, _ := decimal.NewFromString(asset.WalletBalance)
		unrealizedProfit, _ := decimal.NewFromString(asset.UnrealizedProfit)
		marginBalance, _ := decimal.NewFromString(asset.MarginBalance)
		maintMargin, _ := decimal.NewFromString(asset.MaintMargin)
		initialMargin, _ := decimal.NewFromString(asset.InitialMargin)
		positionInitialMargin, _ := decimal.NewFromString(asset.PositionInitialMargin)
		openOrderInitialMargin, _ := decimal.NewFromString(asset.OpenOrderInitialMargin)
		maxWithdrawAmount, _ := decimal.NewFromString(asset.MaxWithdrawAmount)
		crossWalletBalance, _ := decimal.NewFromString(asset.CrossWalletBalance)
		crossUnPnl, _ := decimal.NewFromString(asset.CrossUnPnl)
		availableBalance, _ := decimal.NewFromString(asset.AvailableBalance)

		assets[i] = FuturesAsset{
			Asset:                  asset.Asset,
			WalletBalance:          walletBalance,
			UnrealizedProfit:       unrealizedProfit,
			MarginBalance:          marginBalance,
			MaintMargin:            maintMargin,
			InitialMargin:          initialMargin,
			PositionInitialMargin:  positionInitialMargin,
			OpenOrderInitialMargin: openOrderInitialMargin,
			MaxWithdrawAmount:      maxWithdrawAmount,
			CrossWalletBalance:     crossWalletBalance,
			CrossUnPnl:             crossUnPnl,
			AvailableBalance:       availableBalance,
		}
	}

	positions := make([]FuturesPosition, len(result.Positions))
	for i, pos := range result.Positions {
		size, _ := decimal.NewFromString(pos.PositionAmt)
		entryPrice, _ := decimal.NewFromString(pos.EntryPrice)
		markPrice, _ := decimal.NewFromString(pos.MarkPrice)
		unrealizedPnL, _ := decimal.NewFromString(pos.UnrealizedPnL)
		leverage, _ := decimal.NewFromString(pos.Leverage)
		isolatedMargin, _ := decimal.NewFromString(pos.IsolatedMargin)

		positions[i] = FuturesPosition{
			Exchange:       ExchangeBinance,
			Symbol:         pos.Symbol,
			PositionSide:   FuturesPositionSide(pos.PositionSide),
			Size:           size,
			EntryPrice:     entryPrice,
			MarkPrice:      markPrice,
			UnrealizedPnL:  unrealizedPnL,
			Leverage:       leverage,
			MarginType:     MarginType(pos.MarginType),
			IsolatedMargin: isolatedMargin,
			Timestamp:      time.Unix(0, pos.UpdateTime*int64(time.Millisecond)),
		}
	}

	return &FuturesAccount{
		Exchange:           ExchangeBinance,
		TotalWalletBalance: totalWalletBalance,
		TotalUnrealizedPnL: totalUnrealizedPnL,
		TotalMarginBalance: totalMarginBalance,
		TotalPositionValue: totalPositionValue,
		AvailableBalance:   availableBalance,
		MaxWithdrawAmount:  maxWithdrawAmount,
		Assets:             assets,
		Positions:          positions,
		CanTrade:           result.CanTrade,
		CanDeposit:         result.CanDeposit,
		CanWithdraw:        result.CanWithdraw,
		UpdateTime:         time.Unix(0, result.UpdateTime*int64(time.Millisecond)),
	}, nil
}

// PlaceFuturesOrder places a futures order
func (bf *BinanceFuturesExtension) PlaceFuturesOrder(ctx context.Context, order *FuturesOrder) (*FuturesOrder, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(order.Symbol))
	params.Set("side", strings.ToUpper(string(order.Side)))
	params.Set("type", strings.ToUpper(string(order.Type)))
	params.Set("quantity", order.Quantity.String())

	if order.PositionSide != "" {
		params.Set("positionSide", string(order.PositionSide))
	}

	if order.Type == OrderTypeLimit || order.Type == OrderTypeStopLimit || order.Type == OrderTypeTakeProfitLimit {
		params.Set("price", order.Price.String())
		params.Set("timeInForce", "GTC")
	}

	if order.Type == OrderTypeStop || order.Type == OrderTypeStopLimit || 
	   order.Type == OrderTypeTakeProfit || order.Type == OrderTypeTakeProfitLimit {
		params.Set("stopPrice", order.StopPrice.String())
	}

	if order.Type == OrderTypeTrailingStop {
		params.Set("activationPrice", order.ActivatePrice.String())
		params.Set("callbackRate", order.CallbackRate.String())
	}

	if order.ReduceOnly {
		params.Set("reduceOnly", "true")
	}

	if order.ClosePosition {
		params.Set("closePosition", "true")
	}

	if order.ClientOrderID != "" {
		params.Set("newClientOrderId", order.ClientOrderID)
	}

	resp, err := bf.client.makeRequest(ctx, "POST", "/fapi/v1/order", params, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Symbol              string `json:"symbol"`
		OrderId             int64  `json:"orderId"`
		ClientOrderId       string `json:"clientOrderId"`
		TransactTime        int64  `json:"transactTime"`
		Price               string `json:"price"`
		OrigQty             string `json:"origQty"`
		ExecutedQty         string `json:"executedQty"`
		CumQuote            string `json:"cumQuote"`
		Status              string `json:"status"`
		TimeInForce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
		PositionSide        string `json:"positionSide"`
		StopPrice           string `json:"stopPrice"`
		ActivatePrice       string `json:"activatePrice"`
		CallbackRate        string `json:"callbackRate"`
		ReduceOnly          bool   `json:"reduceOnly"`
		ClosePosition       bool   `json:"closePosition"`
		UpdateTime          int64  `json:"updateTime"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	price, _ := decimal.NewFromString(result.Price)
	quantity, _ := decimal.NewFromString(result.OrigQty)
	filledQuantity, _ := decimal.NewFromString(result.ExecutedQty)
	stopPrice, _ := decimal.NewFromString(result.StopPrice)
	activatePrice, _ := decimal.NewFromString(result.ActivatePrice)
	callbackRate, _ := decimal.NewFromString(result.CallbackRate)

	return &FuturesOrder{
		ID:                strconv.FormatInt(result.OrderId, 10),
		ClientOrderID:     result.ClientOrderId,
		Exchange:          ExchangeBinance,
		Symbol:            result.Symbol,
		Type:              OrderType(strings.ToLower(result.Type)),
		Side:              OrderSide(strings.ToLower(result.Side)),
		PositionSide:      FuturesPositionSide(result.PositionSide),
		Status:            OrderStatus(strings.ToLower(result.Status)),
		Price:             price,
		Quantity:          quantity,
		FilledQuantity:    filledQuantity,
		RemainingQuantity: quantity.Sub(filledQuantity),
		TimeInForce:       result.TimeInForce,
		StopPrice:         stopPrice,
		ActivatePrice:     activatePrice,
		CallbackRate:      callbackRate,
		ReduceOnly:        result.ReduceOnly,
		ClosePosition:     result.ClosePosition,
		CreatedAt:         time.Unix(0, result.TransactTime*int64(time.Millisecond)),
		UpdatedAt:         time.Unix(0, result.UpdateTime*int64(time.Millisecond)),
	}, nil
}

// GetFuturesPositions gets all futures positions
func (bf *BinanceFuturesExtension) GetFuturesPositions(ctx context.Context) ([]FuturesPosition, error) {
	account, err := bf.GetFuturesAccount(ctx)
	if err != nil {
		return nil, err
	}
	return account.Positions, nil
}

// GetFuturesPosition gets a specific futures position
func (bf *BinanceFuturesExtension) GetFuturesPosition(ctx context.Context, symbol string) (*FuturesPosition, error) {
	positions, err := bf.GetFuturesPositions(ctx)
	if err != nil {
		return nil, err
	}

	for _, position := range positions {
		if position.Symbol == strings.ToUpper(symbol) && !position.Size.IsZero() {
			return &position, nil
		}
	}

	return nil, fmt.Errorf("no position found for symbol %s", symbol)
}

// ChangeLeverage changes leverage for a symbol
func (bf *BinanceFuturesExtension) ChangeLeverage(ctx context.Context, symbol string, leverage int) error {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("leverage", strconv.Itoa(leverage))

	resp, err := bf.client.makeRequest(ctx, "POST", "/fapi/v1/leverage", params, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to change leverage: status %d", resp.StatusCode)
	}

	return nil
}

// ChangeMarginType changes margin type for a symbol
func (bf *BinanceFuturesExtension) ChangeMarginType(ctx context.Context, symbol string, marginType MarginType) error {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("marginType", string(marginType))

	resp, err := bf.client.makeRequest(ctx, "POST", "/fapi/v1/marginType", params, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to change margin type: status %d", resp.StatusCode)
	}

	return nil
}
