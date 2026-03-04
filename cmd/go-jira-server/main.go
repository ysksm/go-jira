package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ysksm/go-jira/core/infrastructure/config"
	"github.com/ysksm/go-jira/core/infrastructure/database"
	"github.com/ysksm/go-jira/internal/api"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	corsOrigin := flag.String("cors-origin", "", "CORS allowed origin (e.g., http://localhost:5173)")
	configPath := flag.String("config", "", "config file path (default: ~/.config/go-jira/settings.json)")
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	flag.Parse()

	// Setup logger.
	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	// Load config.
	cfgPath := *configPath
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		cfgPath = home + "/.config/go-jira/settings.json"
	}

	configStore, err := config.NewFileConfigStore(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Setup database connection manager.
	settings, err := configStore.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading settings: %v\n", err)
		os.Exit(1)
	}

	dbDir := settings.Database.Path
	if dbDir == "" {
		home, _ := os.UserHomeDir()
		dbDir = home + "/.config/go-jira/db"
	}
	connMgr := database.NewConnection(dbDir)
	defer connMgr.Close()

	// Start server.
	server := api.NewServer(api.ServerConfig{
		Port:        *port,
		CORSOrigin:  *corsOrigin,
		ConfigStore: configStore,
		ConnMgr:     connMgr,
		Logger:      logger,
	})

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
