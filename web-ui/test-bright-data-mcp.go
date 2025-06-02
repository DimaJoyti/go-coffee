package main

import (
	"fmt"
	"os"
	"time"
)

// Демонстраційний скрипт для тестування Bright Data MCP інтеграції
// Цей скрипт показує як можна використовувати Bright Data MCP для веб-скрапінгу

type BrightDataDemo struct {
	apiToken string
}

type ScrapingResult struct {
	URL       string                 `json:"url"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewBrightDataDemo() *BrightDataDemo {
	return &BrightDataDemo{
		apiToken: os.Getenv("BRIGHT_DATA_API_TOKEN"),
	}
}

func (bd *BrightDataDemo) DemoScrapeCompetitors() {
	fmt.Println("🔍 Демонстрація скрапінгу конкурентів...")

	competitors := []string{
		"https://www.starbucks.com/menu/drinks/espresso",
		"https://www.dunkindonuts.com/en/menu/espresso-and-coffee",
		"https://www.costacoffe.com/menu/coffee",
	}

	for i, url := range competitors {
		fmt.Printf("\n📊 Скрапінг %d/3: %s\n", i+1, url)

		// Симуляція скрапінгу через Bright Data MCP
		result := bd.simulateScraping(url, "competitors")

		fmt.Printf("✅ Заголовок: %s\n", result.Title)
		fmt.Printf("📝 Контент: %s\n", result.Content[:min(100, len(result.Content))]+"...")
		fmt.Printf("⏰ Час: %s\n", result.Timestamp.Format("15:04:05"))

		// Симуляція затримки між запитами
		time.Sleep(1 * time.Second)
	}
}

func (bd *BrightDataDemo) DemoSearchCoffeeNews() {
	fmt.Println("\n🔍 Демонстрація пошуку новин про каву...")

	queries := []string{
		"coffee market prices 2024",
		"arabica coffee futures",
		"sustainable coffee farming",
	}

	for i, query := range queries {
		fmt.Printf("\n🔎 Пошук %d/3: %s\n", i+1, query)

		// Симуляція пошуку через Bright Data MCP
		results := bd.simulateSearch(query)

		for j, result := range results {
			fmt.Printf("  %d. %s\n", j+1, result.Title)
			fmt.Printf("     URL: %s\n", result.URL)
		}

		time.Sleep(1 * time.Second)
	}
}

func (bd *BrightDataDemo) DemoMarketDataCollection() {
	fmt.Println("\n📈 Демонстрація збору ринкових даних...")

	dataSources := []struct {
		name     string
		url      string
		category string
	}{
		{"Coffee Futures", "https://markets.com/coffee", "prices"},
		{"Industry News", "https://coffeenews.com", "news"},
		{"Social Trends", "https://twitter.com/search?q=coffee", "social"},
	}

	for i, source := range dataSources {
		fmt.Printf("\n📊 Джерело %d/3: %s\n", i+1, source.name)

		result := bd.simulateScraping(source.url, source.category)

		fmt.Printf("✅ Категорія: %s\n", source.category)
		fmt.Printf("📝 Дані: %s\n", result.Content[:min(80, len(result.Content))]+"...")

		// Симуляція обробки даних
		bd.processMarketData(result)

		time.Sleep(1 * time.Second)
	}
}

func (bd *BrightDataDemo) simulateScraping(url, category string) ScrapingResult {
	// Симуляція реального скрапінгу через Bright Data MCP

	var title, content string
	metadata := make(map[string]interface{})

	switch category {
	case "competitors":
		title = "Menu Prices and Products"
		content = "Large Latte: $5.45 (+$0.15), Cappuccino: $4.95, Americano: $3.75, Espresso: $2.25. New seasonal drinks available. Premium coffee beans sourced from sustainable farms."
		metadata["prices_found"] = 4
		metadata["new_items"] = 2

	case "prices":
		title = "Coffee Market Futures"
		content = "Arabica coffee futures rose 2.3% to $1.85/lb. Robusta steady at $1.23/lb. Weather concerns in Brazil driving prices higher. Supply chain disruptions continue."
		metadata["arabica_price"] = 1.85
		metadata["robusta_price"] = 1.23
		metadata["change_percent"] = 2.3

	case "news":
		title = "Coffee Industry Updates"
		content = "New sustainable farming practices adopted by major producers. Climate change impact on coffee growing regions. Technology innovations in coffee processing."
		metadata["articles_found"] = 15
		metadata["sentiment"] = "positive"

	case "social":
		title = "Coffee Social Media Trends"
		content = "#CoffeeLovers trending with 50K mentions. Cold brew popularity rising 25%. Specialty coffee shops gaining social media traction."
		metadata["mentions"] = 50000
		metadata["trending_topics"] = []string{"cold brew", "specialty coffee", "sustainable"}
	}

	return ScrapingResult{
		URL:       url,
		Title:     title,
		Content:   content,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}
}

func (bd *BrightDataDemo) simulateSearch(query string) []ScrapingResult {
	// Симуляція пошукових результатів
	results := []ScrapingResult{
		{
			URL:       "https://example1.com",
			Title:     fmt.Sprintf("Latest news about %s", query),
			Content:   "Relevant content about " + query,
			Timestamp: time.Now(),
		},
		{
			URL:       "https://example2.com",
			Title:     fmt.Sprintf("Market analysis: %s", query),
			Content:   "Market insights for " + query,
			Timestamp: time.Now(),
		},
		{
			URL:       "https://example3.com",
			Title:     fmt.Sprintf("Industry report on %s", query),
			Content:   "Industry data about " + query,
			Timestamp: time.Now(),
		},
	}

	return results
}

func (bd *BrightDataDemo) processMarketData(result ScrapingResult) {
	fmt.Printf("🔄 Обробка даних з %s...\n", result.URL)

	// Симуляція обробки та аналізу даних
	if len(result.Metadata) > 0 {
		fmt.Printf("📊 Метадані: ")
		for key, value := range result.Metadata {
			fmt.Printf("%s=%v ", key, value)
		}
		fmt.Println()
	}

	// Симуляція збереження в базу даних
	fmt.Printf("💾 Дані збережено в базу\n")
}

func (bd *BrightDataDemo) ShowSessionStats() {
	fmt.Println("\n📊 Статистика сесії Bright Data MCP:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🌐 Всього запитів: %d\n", 15)
	fmt.Printf("✅ Успішних: %d\n", 14)
	fmt.Printf("❌ Помилок: %d\n", 1)
	fmt.Printf("⏱️  Середній час відповіді: %s\n", "1.2s")
	fmt.Printf("📈 Зібрано даних: %s\n", "2.3MB")
	fmt.Printf("🔄 Використано кредитів: %d\n", 45)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("🚀 Go Coffee Epic UI - Bright Data MCP Demo")
	fmt.Println("═══════════════════════════════════════════")

	demo := NewBrightDataDemo()

	// Перевірка API токена
	if demo.apiToken == "" {
		fmt.Println("⚠️  BRIGHT_DATA_API_TOKEN не встановлено")
		fmt.Println("💡 Використовуємо демонстраційний режим")
	} else {
		fmt.Printf("✅ API токен знайдено: %s...%s\n",
			demo.apiToken[:8],
			demo.apiToken[len(demo.apiToken)-8:])
	}

	fmt.Println("\n🎯 Демонстрація можливостей Bright Data MCP:")

	// Демонстрація різних типів скрапінгу
	demo.DemoScrapeCompetitors()
	demo.DemoSearchCoffeeNews()
	demo.DemoMarketDataCollection()
	demo.ShowSessionStats()

	fmt.Println("\n🎉 Демонстрація завершена!")
	fmt.Println("💡 Ці дані можуть бути інтегровані в Epic UI для:")
	fmt.Println("   • Моніторингу конкурентів")
	fmt.Println("   • Аналізу ринкових цін")
	fmt.Println("   • Відстеження новин індустрії")
	fmt.Println("   • Соціальних медіа трендів")
	fmt.Println("   • Автоматичного оновлення дашборду")
}
