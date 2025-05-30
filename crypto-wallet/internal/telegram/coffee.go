package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/ai"
)

// CoffeeMenu represents the coffee menu
type CoffeeMenu struct {
	Categories []CoffeeCategory `json:"categories"`
}

// CoffeeCategory represents a category of coffee items
type CoffeeCategory struct {
	Name  string       `json:"name"`
	Items []CoffeeItem `json:"items"`
}

// CoffeeItem represents a coffee menu item
type CoffeeItem struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Prices      map[string]float64 `json:"prices"` // size -> price
	Available   bool              `json:"available"`
	Extras      []CoffeeExtra     `json:"extras"`
}

// CoffeeExtra represents an extra/addon for coffee
type CoffeeExtra struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// CoffeeOrder represents a coffee order
type CoffeeOrder struct {
	ID          string        `json:"id"`
	UserID      int64         `json:"user_id"`
	Items       []OrderItem   `json:"items"`
	TotalPrice  float64       `json:"total_price"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	PaymentInfo *PaymentRequest `json:"payment_info,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	CoffeeID string   `json:"coffee_id"`
	Name     string   `json:"name"`
	Size     string   `json:"size"`
	Quantity int      `json:"quantity"`
	Extras   []string `json:"extras"`
	Price    float64  `json:"price"`
}

// getCoffeeMenu returns the coffee menu
func (b *Bot) getCoffeeMenu() *CoffeeMenu {
	return &CoffeeMenu{
		Categories: []CoffeeCategory{
			{
				Name: "☕️ Еспресо напої",
				Items: []CoffeeItem{
					{
						ID:          "espresso",
						Name:        "Еспресо",
						Description: "Класичний італійський еспресо",
						Prices:      map[string]float64{"small": 3.50, "medium": 4.00, "large": 4.50},
						Available:   true,
						Extras:      []CoffeeExtra{{ID: "extra_shot", Name: "Додатковий шот", Price: 1.00}},
					},
					{
						ID:          "americano",
						Name:        "Американо",
						Description: "Еспресо з гарячою водою",
						Prices:      map[string]float64{"small": 4.00, "medium": 4.50, "large": 5.00},
						Available:   true,
						Extras:      []CoffeeExtra{{ID: "extra_shot", Name: "Додатковий шот", Price: 1.00}},
					},
					{
						ID:          "latte",
						Name:        "Латте",
						Description: "Еспресо з молоком та молочною піною",
						Prices:      map[string]float64{"small": 5.50, "medium": 6.00, "large": 6.50},
						Available:   true,
						Extras: []CoffeeExtra{
							{ID: "extra_milk", Name: "Додаткове молоко", Price: 0.50},
							{ID: "vanilla_syrup", Name: "Ванільний сироп", Price: 0.75},
							{ID: "caramel_syrup", Name: "Карамельний сироп", Price: 0.75},
						},
					},
					{
						ID:          "cappuccino",
						Name:        "Капучино",
						Description: "Еспресо з молоком та густою піною",
						Prices:      map[string]float64{"small": 5.00, "medium": 5.50, "large": 6.00},
						Available:   true,
						Extras: []CoffeeExtra{
							{ID: "extra_milk", Name: "Додаткове молоко", Price: 0.50},
							{ID: "cinnamon", Name: "Кориця", Price: 0.25},
						},
					},
				},
			},
			{
				Name: "🥛 Молочні напої",
				Items: []CoffeeItem{
					{
						ID:          "flat_white",
						Name:        "Флет Вайт",
						Description: "Еспресо з мікропіною молока",
						Prices:      map[string]float64{"small": 5.50, "medium": 6.00, "large": 6.50},
						Available:   true,
					},
					{
						ID:          "mocha",
						Name:        "Мокка",
						Description: "Еспресо з шоколадом та молоком",
						Prices:      map[string]float64{"small": 6.00, "medium": 6.50, "large": 7.00},
						Available:   true,
						Extras: []CoffeeExtra{
							{ID: "whipped_cream", Name: "Збиті вершки", Price: 0.75},
							{ID: "extra_chocolate", Name: "Додатковий шоколад", Price: 0.50},
						},
					},
				},
			},
			{
				Name: "❄️ Холодні напої",
				Items: []CoffeeItem{
					{
						ID:          "iced_latte",
						Name:        "Айс Латте",
						Description: "Холодний латте з льодом",
						Prices:      map[string]float64{"small": 5.50, "medium": 6.00, "large": 6.50},
						Available:   true,
					},
					{
						ID:          "cold_brew",
						Name:        "Колд Брю",
						Description: "Кава холодного заварювання",
						Prices:      map[string]float64{"small": 4.50, "medium": 5.00, "large": 5.50},
						Available:   true,
					},
				},
			},
		},
	}
}

// processCoffeeOrderWithAI processes coffee order using AI
func (b *Bot) processCoffeeOrderWithAI(ctx context.Context, message string, session *UserSession) (*ai.CoffeeOrderResponse, error) {
	// Create AI request for coffee order processing
	aiReq := &ai.CoffeeOrderRequest{
		UserID:    fmt.Sprintf("%d", session.UserID),
		Message:   message,
		ChatID:    session.ChatID,
		MessageID: 0, // Will be set by the caller
	}

	// Use AI service to process the order
	response, err := b.aiService.ProcessCoffeeOrder(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI processing failed: %w", err)
	}

	// If AI successfully parsed order details, validate and price them
	if response.OrderDetails != nil {
		validatedOrder, err := b.validateAndPriceOrder(response.OrderDetails)
		if err != nil {
			b.logger.Warn(fmt.Sprintf("Order validation failed: %v", err))
			// Continue with AI response but without order details
			response.OrderDetails = nil
		} else {
			response.OrderDetails = validatedOrder
		}
	}

	return response, nil
}

// validateAndPriceOrder validates and prices a coffee order
func (b *Bot) validateAndPriceOrder(order *ai.ParsedCoffeeOrder) (*ai.ParsedCoffeeOrder, error) {
	menu := b.getCoffeeMenu()
	
	// Find the coffee item
	var coffeeItem *CoffeeItem
	for _, category := range menu.Categories {
		for _, item := range category.Items {
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(order.CoffeeType)) ||
			   strings.Contains(strings.ToLower(order.CoffeeType), strings.ToLower(item.Name)) {
				coffeeItem = &item
				break
			}
		}
		if coffeeItem != nil {
			break
		}
	}

	if coffeeItem == nil {
		return nil, fmt.Errorf("coffee type not found: %s", order.CoffeeType)
	}

	if !coffeeItem.Available {
		return nil, fmt.Errorf("coffee not available: %s", order.CoffeeType)
	}

	// Validate size and get price
	size := strings.ToLower(order.Size)
	if size == "" {
		size = "medium" // Default size
	}

	basePrice, exists := coffeeItem.Prices[size]
	if !exists {
		// Try to find closest size
		if len(coffeeItem.Prices) > 0 {
			for s, p := range coffeeItem.Prices {
				size = s
				basePrice = p
				break
			}
		} else {
			return nil, fmt.Errorf("no pricing available for %s", coffeeItem.Name)
		}
	}

	// Calculate total price with extras
	totalPrice := basePrice * float64(order.Quantity)

	// Add extras pricing (simplified)
	for _, extra := range order.Extras {
		for _, availableExtra := range coffeeItem.Extras {
			if strings.Contains(strings.ToLower(availableExtra.Name), strings.ToLower(extra)) {
				totalPrice += availableExtra.Price * float64(order.Quantity)
				break
			}
		}
	}

	// Update the order with validated information
	order.CoffeeType = coffeeItem.Name
	order.Size = size
	order.EstimatedPriceUSD = totalPrice

	// Suggest payment method based on amount
	if totalPrice < 10 {
		order.PaymentSuggestion = "USDC або USDT (низькі комісії)"
	} else if totalPrice < 25 {
		order.PaymentSuggestion = "ETH або USDC"
	} else {
		order.PaymentSuggestion = "BTC або ETH"
	}

	return order, nil
}

// createOrderFromParsed creates a CoffeeOrder from parsed AI order
func (b *Bot) createOrderFromParsed(session *UserSession, parsedOrder *ai.ParsedCoffeeOrder) *CoffeeOrder {
	orderID := fmt.Sprintf("order_%d_%d", session.UserID, time.Now().Unix())
	
	orderItem := OrderItem{
		CoffeeID: strings.ToLower(strings.ReplaceAll(parsedOrder.CoffeeType, " ", "_")),
		Name:     parsedOrder.CoffeeType,
		Size:     parsedOrder.Size,
		Quantity: parsedOrder.Quantity,
		Extras:   parsedOrder.Extras,
		Price:    parsedOrder.EstimatedPriceUSD,
	}

	return &CoffeeOrder{
		ID:         orderID,
		UserID:     session.UserID,
		Items:      []OrderItem{orderItem},
		TotalPrice: parsedOrder.EstimatedPriceUSD,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
}

// generateCoffeeRecommendations generates AI-powered coffee recommendations
func (b *Bot) generateCoffeeRecommendations(ctx context.Context, session *UserSession) (string, error) {
	// Get user preferences from session or history
	preferences := b.getUserPreferences(session)
	
	// Create context-aware prompt
	prompt := fmt.Sprintf(`Надай персоналізовані рекомендації кави для користувача. 

Контекст:
- Час дня: %s
- Попередні замовлення: %s
- Улюблені напої: %s

Меню:
%s

Надай 3-4 рекомендації з коротким описом кожної українською мовою. 
Враховуй час дня, погоду та популярні вибори.`,
		b.getTimeOfDay(),
		preferences["previous_orders"],
		preferences["favorite_drinks"],
		b.getMenuText(),
	)

	generateReq := &ai.GenerateRequest{
		UserID:      fmt.Sprintf("%d", session.UserID),
		Message:     prompt,
		Context:     "coffee_recommendations",
		Temperature: 0.8,
	}

	response, err := b.aiService.GenerateResponse(ctx, generateReq)
	if err != nil {
		return "", fmt.Errorf("failed to generate recommendations: %w", err)
	}

	return response.Text, nil
}

// getUserPreferences gets user preferences from session
func (b *Bot) getUserPreferences(session *UserSession) map[string]string {
	// In a real implementation, this would fetch from database
	return map[string]string{
		"previous_orders": "Латте, Капучино",
		"favorite_drinks": "Молочні напої",
		"preferred_size":  "Medium",
	}
}

// getTimeOfDay returns current time of day description
func (b *Bot) getTimeOfDay() string {
	hour := time.Now().Hour()
	switch {
	case hour < 6:
		return "Рання ніч"
	case hour < 12:
		return "Ранок"
	case hour < 17:
		return "День"
	case hour < 21:
		return "Вечір"
	default:
		return "Ніч"
	}
}

// getMenuText returns formatted menu text for AI context
func (b *Bot) getMenuText() string {
	menu := b.getCoffeeMenu()
	var menuText strings.Builder
	
	for _, category := range menu.Categories {
		menuText.WriteString(fmt.Sprintf("%s:\n", category.Name))
		for _, item := range category.Items {
			if item.Available {
				prices := make([]string, 0, len(item.Prices))
				for size, price := range item.Prices {
					prices = append(prices, fmt.Sprintf("%s: $%.2f", size, price))
				}
				menuText.WriteString(fmt.Sprintf("- %s (%s)\n", item.Name, strings.Join(prices, ", ")))
			}
		}
		menuText.WriteString("\n")
	}
	
	return menuText.String()
}
