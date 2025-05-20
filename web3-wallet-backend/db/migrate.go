package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/yourusername/web3-wallet-backend/pkg/config"
)

func main() {
	// Define command line flags
	upFlag := flag.Bool("up", false, "Migrate up")
	downFlag := flag.Bool("down", false, "Migrate down")
	versionFlag := flag.Int("version", 0, "Migrate to specific version")
	configFlag := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFlag)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Construct database URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	// Create a new migrate instance
	m, err := migrate.New("file://db/migrations", dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	// Execute migration based on flags
	if *upFlag {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate up: %v", err)
		}
		log.Println("Migration up completed successfully")
	} else if *downFlag {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate down: %v", err)
		}
		log.Println("Migration down completed successfully")
	} else if *versionFlag > 0 {
		if err := m.Migrate(uint(*versionFlag)); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate to version %d: %v", *versionFlag, err)
		}
		log.Printf("Migration to version %d completed successfully", *versionFlag)
	} else {
		log.Println("No migration action specified. Use -up, -down, or -version flags.")
		os.Exit(1)
	}
}
