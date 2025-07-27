package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"governance-alerts-cosmos/internal/config"
	"governance-alerts-cosmos/internal/service"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configPath string
	logLevel   string
)

var rootCmd = &cobra.Command{
	Use:   "governance-alerts-cosmos",
	Short: "A service that monitors governance proposals on Cosmos networks",
	Long: `A service that monitors governance proposals on Cosmos networks and sends 
notifications when voting is about to start or end.`,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "config/config.yaml", "Path to configuration file")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
}

func run(cmd *cobra.Command, args []string) error {
	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	logrus.SetLevel(level)

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logrus.Info("Configuration loaded successfully")
	logrus.Infof("Monitoring %d networks", len(cfg.Networks))
	for name, network := range cfg.Networks {
		logrus.Infof("  - %s (%s)", name, network.Name)
	}

	// Create service
	svc, err := service.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logrus.Info("Service started. Press Ctrl+C to stop.")

	// Start service in goroutine
	go func() {
		if err := svc.Run(ctx); err != nil {
			logrus.Errorf("Service error: %v", err)
		}
	}()

	<-sigChan

	// Stop service
	svc.Stop()

	logrus.Info("Service stopped gracefully")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
