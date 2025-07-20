package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// BinanceAdvancedOrdersExtension extends BinanceClient with advanced order types
type BinanceAdvancedOrdersExtension struct {
	client *BinanceClient
}

// NewBinanceAdvancedOrdersExtension creates a new advanced orders extension
func NewBinanceAdvancedOrdersExtension(client *BinanceClient) *BinanceAdvancedOrdersExtension {
	return &BinanceAdvancedOrdersExtension{
		client: client,
	}
}

// PlaceOCOOrder places an OCO (One-Cancels-Other) order
func (ba *BinanceAdvancedOrdersExtension) PlaceOCOOrder(ctx context.Context, symbol string, side OrderSide, quantity, price, stopPrice, stopLimitPrice decimal.Decimal, stopLimitTimeInForce string) (*OCOOrder, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("side", strings.ToUpper(string(side)))
	params.Set("quantity", quantity.String())
	params.Set("price", price.String())
	params.Set("stopPrice", stopPrice.String())

	if !stopLimitPrice.IsZero() {
		params.Set("stopLimitPrice", stopLimitPrice.String())
	}

	if stopLimitTimeInForce != "" {
		params.Set("stopLimitTimeInForce", stopLimitTimeInForce)
	} else {
		params.Set("stopLimitTimeInForce", "GTC")
	}

	resp, err := ba.client.makeRequest(ctx, "POST", "/api/v3/order/oco", params, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OrderListId       int64 `json:"orderListId"`
		ContingencyType   string `json:"contingencyType"`
		ListStatusType    string `json:"listStatusType"`
		ListOrderStatus   string `json:"listOrderStatus"`
		ListClientOrderId string `json:"listClientOrderId"`
		TransactionTime   int64 `json:"transactionTime"`
		Symbol            string `json:"symbol"`
		Orders            []struct {
			Symbol        string `json:"symbol"`
			OrderId       int64  `json:"orderId"`
			ClientOrderId string `json:"clientOrderId"`
		} `json:"orders"`
		OrderReports []struct {
			Symbol              string `json:"symbol"`
			OrderId             int64  `json:"orderId"`
			ClientOrderId       string `json:"clientOrderId"`
			TransactTime        int64  `json:"transactTime"`
			Price               string `json:"price"`
			OrigQty             string `json:"origQty"`
			ExecutedQty         string `json:"executedQty"`
			CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
			Status              string `json:"status"`
			TimeInForce         string `json:"timeInForce"`
			Type                string `json:"type"`
			Side                string `json:"side"`
			StopPrice           string `json:"stopPrice"`
		} `json:"orderReports"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert orders
	orders := make([]Order, len(result.Orders))
	for i, order := range result.Orders {
		orders[i] = Order{
			ID:            strconv.FormatInt(order.OrderId, 10),
			ClientOrderID: order.ClientOrderId,
			Exchange:      ExchangeBinance,
			Symbol:        order.Symbol,
		}
	}

	// Convert order reports
	orderReports := make([]Order, len(result.OrderReports))
	for i, report := range result.OrderReports {
		price, _ := decimal.NewFromString(report.Price)
		quantity, _ := decimal.NewFromString(report.OrigQty)
		filledQuantity, _ := decimal.NewFromString(report.ExecutedQty)
		stopPrice, _ := decimal.NewFromString(report.StopPrice)

		orderReports[i] = Order{
			ID:                strconv.FormatInt(report.OrderId, 10),
			ClientOrderID:     report.ClientOrderId,
			Exchange:          ExchangeBinance,
			Symbol:            report.Symbol,
			Type:              OrderType(strings.ToLower(report.Type)),
			Side:              OrderSide(strings.ToLower(report.Side)),
			Status:            OrderStatus(strings.ToLower(report.Status)),
			Price:             price,
			Quantity:          quantity,
			FilledQuantity:    filledQuantity,
			RemainingQuantity: quantity.Sub(filledQuantity),
			TimeInForce:       report.TimeInForce,
			StopPrice:         stopPrice,
			CreatedAt:         time.Unix(0, report.TransactTime*int64(time.Millisecond)),
			UpdatedAt:         time.Now(),
		}
	}

	return &OCOOrder{
		ID:                strconv.FormatInt(result.OrderListId, 10),
		ListOrderStatus:   result.ListOrderStatus,
		ListStatusType:    result.ListStatusType,
		ListClientOrderID: result.ListClientOrderId,
		TransactionTime:   time.Unix(0, result.TransactionTime*int64(time.Millisecond)),
		Symbol:            result.Symbol,
		Orders:            orders,
		OrderReports:      orderReports,
	}, nil
}

// CancelOCOOrder cancels an OCO order
func (ba *BinanceAdvancedOrdersExtension) CancelOCOOrder(ctx context.Context, symbol, orderListId string) error {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("orderListId", orderListId)

	resp, err := ba.client.makeRequest(ctx, "DELETE", "/api/v3/orderList", params, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to cancel OCO order: status %d", resp.StatusCode)
	}

	return nil
}

// PlaceStopLossOrder places a stop-loss order
func (ba *BinanceAdvancedOrdersExtension) PlaceStopLossOrder(ctx context.Context, symbol string, side OrderSide, quantity, stopPrice decimal.Decimal) (*Order, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("side", strings.ToUpper(string(side)))
	params.Set("type", "STOP_LOSS")
	params.Set("quantity", quantity.String())
	params.Set("stopPrice", stopPrice.String())

	resp, err := ba.client.makeRequest(ctx, "POST", "/api/v3/order", params, true)
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
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
		Status              string `json:"status"`
		TimeInForce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
		StopPrice           string `json:"stopPrice"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	price, _ := decimal.NewFromString(result.Price)
	quantity, _ = decimal.NewFromString(result.OrigQty)
	filledQuantity, _ := decimal.NewFromString(result.ExecutedQty)
	stopPriceDecimal, _ := decimal.NewFromString(result.StopPrice)

	return &Order{
		ID:                strconv.FormatInt(result.OrderId, 10),
		ClientOrderID:     result.ClientOrderId,
		Exchange:          ExchangeBinance,
		Symbol:            result.Symbol,
		Type:              OrderType(strings.ToLower(result.Type)),
		Side:              OrderSide(strings.ToLower(result.Side)),
		Status:            OrderStatus(strings.ToLower(result.Status)),
		Price:             price,
		Quantity:          quantity,
		FilledQuantity:    filledQuantity,
		RemainingQuantity: quantity.Sub(filledQuantity),
		TimeInForce:       result.TimeInForce,
		StopPrice:         stopPriceDecimal,
		CreatedAt:         time.Unix(0, result.TransactTime*int64(time.Millisecond)),
		UpdatedAt:         time.Now(),
	}, nil
}

// PlaceTakeProfitOrder places a take-profit order
func (ba *BinanceAdvancedOrdersExtension) PlaceTakeProfitOrder(ctx context.Context, symbol string, side OrderSide, quantity, stopPrice decimal.Decimal) (*Order, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("side", strings.ToUpper(string(side)))
	params.Set("type", "TAKE_PROFIT")
	params.Set("quantity", quantity.String())
	params.Set("stopPrice", stopPrice.String())

	resp, err := ba.client.makeRequest(ctx, "POST", "/api/v3/order", params, true)
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
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
		Status              string `json:"status"`
		TimeInForce         string `json:"timeInForce"`
		Type                string `json:"type"`
		Side                string `json:"side"`
		StopPrice           string `json:"stopPrice"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	price, _ := decimal.NewFromString(result.Price)
	quantity, _ = decimal.NewFromString(result.OrigQty)
	filledQuantity, _ := decimal.NewFromString(result.ExecutedQty)
	stopPriceDecimal, _ := decimal.NewFromString(result.StopPrice)

	return &Order{
		ID:                strconv.FormatInt(result.OrderId, 10),
		ClientOrderID:     result.ClientOrderId,
		Exchange:          ExchangeBinance,
		Symbol:            result.Symbol,
		Type:              OrderType(strings.ToLower(result.Type)),
		Side:              OrderSide(strings.ToLower(result.Side)),
		Status:            OrderStatus(strings.ToLower(result.Status)),
		Price:             price,
		Quantity:          quantity,
		FilledQuantity:    filledQuantity,
		RemainingQuantity: quantity.Sub(filledQuantity),
		TimeInForce:       result.TimeInForce,
		StopPrice:         stopPriceDecimal,
		CreatedAt:         time.Unix(0, result.TransactTime*int64(time.Millisecond)),
		UpdatedAt:         time.Now(),
	}, nil
}

// GetMarginAccount gets margin account information
func (ba *BinanceAdvancedOrdersExtension) GetMarginAccount(ctx context.Context) (*MarginAccount, error) {
	resp, err := ba.client.makeRequest(ctx, "GET", "/sapi/v1/margin/account", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		BorrowEnabled       bool   `json:"borrowEnabled"`
		MarginLevel         string `json:"marginLevel"`
		TotalAssetOfBtc     string `json:"totalAssetOfBtc"`
		TotalLiabilityOfBtc string `json:"totalLiabilityOfBtc"`
		TotalNetAssetOfBtc  string `json:"totalNetAssetOfBtc"`
		TradeEnabled        bool   `json:"tradeEnabled"`
		TransferEnabled     bool   `json:"transferEnabled"`
		UpdateTime          int64  `json:"updateTime"`
		UserAssets          []struct {
			Asset    string `json:"asset"`
			Borrowed string `json:"borrowed"`
			Free     string `json:"free"`
			Interest string `json:"interest"`
			Locked   string `json:"locked"`
			NetAsset string `json:"netAsset"`
		} `json:"userAssets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	marginLevel, _ := decimal.NewFromString(result.MarginLevel)
	totalAssetOfBtc, _ := decimal.NewFromString(result.TotalAssetOfBtc)
	totalLiabilityOfBtc, _ := decimal.NewFromString(result.TotalLiabilityOfBtc)
	totalNetAssetOfBtc, _ := decimal.NewFromString(result.TotalNetAssetOfBtc)

	userAssets := make([]MarginAsset, len(result.UserAssets))
	for i, asset := range result.UserAssets {
		borrowed, _ := decimal.NewFromString(asset.Borrowed)
		free, _ := decimal.NewFromString(asset.Free)
		interest, _ := decimal.NewFromString(asset.Interest)
		locked, _ := decimal.NewFromString(asset.Locked)
		netAsset, _ := decimal.NewFromString(asset.NetAsset)

		userAssets[i] = MarginAsset{
			Asset:    asset.Asset,
			Borrowed: borrowed,
			Free:     free,
			Interest: interest,
			Locked:   locked,
			NetAsset: netAsset,
		}
	}

	return &MarginAccount{
		Exchange:            ExchangeBinance,
		BorrowEnabled:       result.BorrowEnabled,
		MarginLevel:         marginLevel,
		TotalAssetOfBtc:     totalAssetOfBtc,
		TotalLiabilityOfBtc: totalLiabilityOfBtc,
		TotalNetAssetOfBtc:  totalNetAssetOfBtc,
		TradeEnabled:        result.TradeEnabled,
		TransferEnabled:     result.TransferEnabled,
		UserAssets:          userAssets,
		UpdateTime:          time.Unix(0, result.UpdateTime*int64(time.Millisecond)),
	}, nil
}
