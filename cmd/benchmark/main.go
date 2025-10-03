package main

import (
	"apollo-testing-suite/internal/benchmark"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

func main() {
	var (
		utxoInput    int
		utxoOutput   int
		iterations   int
		parallelism  int
		outputFormat string
		cpuProfile   string
		logLevel     string
		utxoLevel    int
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
			slog.Info("Starting benchmark application")
			slog.Debug("Command PersistentPreRunE finished")
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			slog.Debug("Command Run started")
			benchmark.Run(utxoInput, utxoOutput, iterations, parallelism, outputFormat, cpuProfile, utxoLevel)
			slog.Debug("Command Run finished")
		},
	}

	cmd.Flags().IntVarP(&utxoInput, "utxo-input", "u", 10, "Number of UTXOs to use as input")
	cmd.Flags().IntVarP(&utxoOutput, "utxo-output", "v", 10, "Number of UTXOs to generate as output")
	cmd.Flags().IntVar(&utxoLevel, "utxo-level", 1, "Set UTXO generation level: 1=simple, 2=differentiated, 3=congested")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 1000, "Number of transactions to build")
	cmd.Flags().IntVarP(&parallelism, "parallelism", "p", 4, "Number of parallel goroutines")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table/json)")
	cmd.Flags().StringVarP(&cpuProfile, "cpu-profile", "c", "", "Write CPU profile to file")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "Set logging level (debug, info, warn, error)")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		slog.Debug("Command PreRunE started")
		if utxoInput <= 0 {
			slog.Warn("Invalid --utxo-input", "value", utxoInput)
			return errors.New("--utxo-input must be > 0")
		}
		if utxoOutput <= 0 {
			slog.Warn("Invalid --utxo-output", "value", utxoOutput)
			return errors.New("--utxo-output must be > 0")
		}
		if iterations <= 0 {
			slog.Warn("Invalid --iterations", "value", iterations)
			return errors.New("--iterations must be > 0")
		}
		slog.Debug("Command PreRunE finished successfully")
		return nil
	}

	if err := cmd.Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
