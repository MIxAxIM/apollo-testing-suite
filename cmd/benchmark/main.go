package main

import (
	"apollo-testing-suite/internal/benchmark"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

func main() {
	var (
		utxoCount           int
		iterations          int
		parallelism         int
		backend             string
		outputFormat        string
		cpuProfile          string
		maestroNetworkID    int
		maestroAPIKey       string
		blockfrostAPIURL    string
		blockfrostNetworkID int
		blockfrostAPIKey    string
		ogmiosEndpoint      string
		kugoEndpoint        string
		utxorpcEndpoint     string
		utxorpcNetworkID    int
		utxorpcAPIKey       string
		proxy               string
		logLevel            string
	)

	cmd := &cobra.Command{
		Use: "apollo-bench",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize slog
			var level slog.Level
			switch logLevel {
			case "debug":
				level = slog.LevelDebug
			case "info":
				level = slog.LevelInfo
			case "warn":
				level = slog.LevelWarn
			case "error":
				level = slog.LevelError
			default:
				level = slog.LevelInfo
			}

			slog.SetDefault(slog.New(
				tint.NewHandler(os.Stderr, &tint.Options{
					Level:      level,
					AddSource:  true,
					NoColor:    false,
					TimeFormat: time.Kitchen,
				}),
			))

			slog.Debug("Command PersistentPreRunE started")
			if proxy != "" {
				os.Setenv("HTTP_PROXY", proxy)
				os.Setenv("HTTPS_PROXY", proxy)
				slog.Info("Proxy enabled", "address", proxy, "HTTP_PROXY", os.Getenv("HTTP_PROXY"), "HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
			}
			slog.Info("Starting benchmark application")
			slog.Debug("Command PersistentPreRunE finished")
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			slog.Debug("Command Run started")
			benchmark.Run(utxoCount, iterations, parallelism, backend, outputFormat, cpuProfile,
				maestroNetworkID, maestroAPIKey,
				blockfrostAPIURL, blockfrostNetworkID, blockfrostAPIKey,
				ogmiosEndpoint, kugoEndpoint,
				utxorpcEndpoint, utxorpcNetworkID, utxorpcAPIKey)
			slog.Debug("Command Run finished")
		},
	}

	cmd.Flags().IntVarP(&utxoCount, "utxo-count", "u", 10, "Number of UTXOs to test")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 1000, "Number of transactions to build")
	cmd.Flags().IntVarP(&parallelism, "parallelism", "p", 4, "Number of parallel goroutines")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table/json)")
	cmd.Flags().StringVarP(&backend, "backend", "b", "maestro", "Backend Chain Indexer (maestro, blockfrost, ogmios, utxorpc)")
	cmd.Flags().StringVarP(&cpuProfile, "cpu-profile", "c", "", "Write CPU profile to file")

	// Proxy flag
	cmd.Flags().StringVar(&proxy, "proxy", "", "Proxy address (e.g., http://127.0.0.1:2080 or socks5://127.0.0.1:1080)")
	// Log level flag
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "Set logging level (debug, info, warn, error)")

	// Maestro flags
	cmd.Flags().IntVar(&maestroNetworkID, "maestro-network-id", 0, "Maestro Network ID")
	cmd.Flags().StringVar(&maestroAPIKey, "maestro-api-key", "", "Maestro API Key")
	cmd.Flags().MarkHidden("maestro-network-id")
	cmd.Flags().MarkHidden("maestro-api-key")

	// Blockfrost flags
	cmd.Flags().StringVar(&blockfrostAPIURL, "blockfrost-api-url", "", "Blockfrost API URL")
	cmd.Flags().IntVar(&blockfrostNetworkID, "blockfrost-network-id", 0, "Blockfrost Network ID")
	cmd.Flags().StringVar(&blockfrostAPIKey, "blockfrost-api-key", "", "Blockfrost API Key")
	cmd.Flags().MarkHidden("blockfrost-api-url")
	cmd.Flags().MarkHidden("blockfrost-network-id")
	cmd.Flags().MarkHidden("blockfrost-api-key")

	// Ogmios flags
	cmd.Flags().StringVar(&ogmiosEndpoint, "ogmios-endpoint", "", "Ogmios Endpoint")
	cmd.Flags().StringVar(&kugoEndpoint, "kugo-endpoint", "", "Kugo Endpoint")
	cmd.Flags().MarkHidden("ogmios-endpoint")
	cmd.Flags().MarkHidden("kugo-endpoint")

	// Utxorpc flags
	cmd.Flags().StringVar(&utxorpcEndpoint, "utxorpc-endpoint", "", "Utxorpc Endpoint")
	cmd.Flags().IntVar(&utxorpcNetworkID, "utxorpc-network-id", 0, "Utxorpc Network ID")
	cmd.Flags().StringVar(&utxorpcAPIKey, "utxorpc-api-key", "", "Utxorpc API Key")
	cmd.Flags().MarkHidden("utxorpc-endpoint")
	cmd.Flags().MarkHidden("utxorpc-network-id")
	cmd.Flags().MarkHidden("utxorpc-api-key")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		slog.Debug("Command PreRunE started")
		if utxoCount <= 0 {
			slog.Warn("Invalid --utxo-count", "value", utxoCount)
			return errors.New("--utxo-count must be > 0")
		}
		if iterations <= 0 {
			slog.Warn("Invalid --iterations", "value", iterations)
			return errors.New("--iterations must be > 0")
		}

		switch backend {
		case "maestro":
			slog.Debug("Backend selected: maestro")
			cmd.Flags().MarkHidden("maestro-network-id")
			cmd.Flags().MarkHidden("maestro-api-key")
			if maestroAPIKey == "" {
				slog.Warn("--maestro-api-key is required for maestro backend")
				return errors.New("--maestro-api-key is required for maestro backend")
			}
		case "blockfrost":
			slog.Debug("Backend selected: blockfrost")
			cmd.Flags().MarkHidden("blockfrost-api-url")
			cmd.Flags().MarkHidden("blockfrost-network-id")
			cmd.Flags().MarkHidden("blockfrost-api-key")
			if blockfrostAPIKey == "" || blockfrostAPIURL == "" {
				slog.Warn("--blockfrost-api-key and --blockfrost-api-url are required for blockfrost backend")
				return errors.New("--blockfrost-api-key and --blockfrost-api-url are required for blockfrost backend")
			}
		case "ogmios":
			slog.Debug("Backend selected: ogmios")
			cmd.Flags().MarkHidden("ogmios-endpoint")
			cmd.Flags().MarkHidden("kugo-endpoint")
			if ogmiosEndpoint == "" || kugoEndpoint == "" {
				slog.Warn("--ogmios-endpoint and --kugo-endpoint are required for ogmios backend")
				return errors.New("--ogmios-endpoint and --kugo-endpoint are required for ogmios backend")
			}
		case "utxorpc":
			slog.Debug("Backend selected: utxorpc")
			cmd.Flags().MarkHidden("utxorpc-endpoint")
			cmd.Flags().MarkHidden("utxorpc-network-id")
			cmd.Flags().MarkHidden("utxorpc-api-key")
			if utxorpcEndpoint == "" {
				slog.Warn("--utxorpc-endpoint is required for utxorpc backend")
				return errors.New("--utxorpc-endpoint is required for utxorpc backend")
			}
		default:
			slog.Error("Unknown backend", "backend", backend)
			return fmt.Errorf("unknown backend: %s", backend)
		}
		slog.Debug("Command PreRunE finished successfully")
		return nil
	}

	if err := cmd.Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
