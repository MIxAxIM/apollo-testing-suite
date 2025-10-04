package benchmark

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/Salvionied/apollo"
	"github.com/Salvionied/apollo/serialization"
	"github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/serialization/UTxO"
	"github.com/Salvionied/apollo/txBuilding/Backend/Base"
	"github.com/Salvionied/apollo/txBuilding/Backend/FixedChainContext"
)

type Result struct {
	Duration time.Duration
	Error    error
}

func Run(utxoInput, utxoOutput, iterations, parallelism int, outputFormat string, cpuProfile string, utxoLevel int) {

	slog.Info("Starting benchmark run",
		"utxoInput", utxoInput,
		"utxoOutput", utxoOutput,
		"iterations", iterations,
		"parallelism", parallelism,
		"outputFormat", outputFormat,
		"cpuProfile", cpuProfile,
		"utxoLevel", utxoLevel)

	ctx := FixedChainContext.InitFixedChainContext()

	senderWalletAddress, err := Address.DecodeAddress(TEST_WALLET_ADDRESS_1)
	if err != nil {
		slog.Error("Error decoding wallet address", "error", err)
		os.Exit(1)
	}
	slog.Debug("Sender wallet address decoded", "address", senderWalletAddress.String())

	receiverWalletAddress, err := Address.DecodeAddress(TEST_WALLET_ADDRESS_2)
	if err != nil {
		slog.Error("Error decoding wallet address", "error", err)
		os.Exit(1)
	}
	slog.Debug("Receiver wallet address decoded", "address", receiverWalletAddress.String())

	var userUtxos []UTxO.UTxO

	switch utxoLevel {
	case 1: // Simple
		userUtxos = InitUtxos(utxoInput)
	case 2: // Differentiated
		userUtxos = InitUtxosDifferentiated(utxoInput)
	case 3: // Congested
		userUtxos = InitUtxosCongested(utxoInput)
	default:
		slog.Error("Invalid UTXO level", "level", utxoLevel)
		os.Exit(1)
	}

	slog.Info("Fetched user UTXOs", "count", len(userUtxos))

	// Warm-up phase before any measurements
	runtime.GC()
	time.Sleep(2 * time.Second)
	slog.Info("Warm-up phase completed")

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			slog.Error("Failed to create cpu profile file", "error", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			slog.Error("Could not start CPU profile", "error", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
		slog.Info("CPU profiling started", "file", cpuProfile)
	}

	var (
		wg           sync.WaitGroup
		results      = make(chan Result, iterations)
		totalLatency time.Duration
		mu           sync.Mutex
	)

	sem := make(chan struct{}, parallelism)

	// Actual benchmark start time
	benchStart := time.Now()
	slog.Info("Benchmark iterations starting", "iterations", iterations, "parallelism", parallelism)

	for i := range iterations {
		wg.Add(1)
		sem <- struct{}{}

		go func(iter int) {
			defer func() {
				<-sem
				wg.Done()
				if r := recover(); r != nil {
					mu.Lock()
					results <- Result{
						Error: fmt.Errorf("panic in iteration %d: %v", iter, r),
					}
					mu.Unlock()
					slog.Error("Panic during iteration", "iteration", iter, "panic", r)
				}
			}()

			// For thread safety
			clonedUTxOs := make([]UTxO.UTxO, len(userUtxos))
			copy(clonedUTxOs, userUtxos)

			start := time.Now()
			err := buildTransaction(clonedUTxOs, &receiverWalletAddress, ctx, utxoOutput)
			elapsed := time.Since(start)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results <- Result{Error: fmt.Errorf("iteration %d: %w", iter, err)}
				slog.Warn("Transaction build failed", "iteration", iter, "error", err)
			} else {
				results <- Result{Duration: elapsed}
				totalLatency += elapsed
				slog.Debug("Transaction built successfully", "iteration", iter, "duration", elapsed)
			}
		}(i)
	}

	wg.Wait()
	close(results)
	slog.Info("All benchmark iterations completed")

	// Calculate metrics
	benchDuration := time.Since(benchStart)
	var failures int
	successes := 0

	for res := range results {
		if res.Error != nil {
			slog.Error("Error during iteration", "error", res.Error)
			failures++
		} else {
			successes++
		}
	}

	if successes == 0 {
		slog.Error("All iterations failed! Check logs for errors.")
		os.Exit(1)
	}

	// Calculate accurate Tx/s metrics
	actualTxPerSec := float64(successes) / benchDuration.Seconds()
	latencyPerTx := totalLatency / time.Duration(successes)

	// For comparison: latency-based Tx/s
	latencyTxPerSec := float64(time.Second) / float64(latencyPerTx)
	slog.Info("Benchmark results",
		"actualTxPerSec", actualTxPerSec,
		"latencyTxPerSec", latencyTxPerSec,
		"latencyPerTx", latencyPerTx,
		"failures", failures,
		"iterations", iterations,
		"parallelism", parallelism,
		"utxoInput", utxoInput,
		"utxoOutput", utxoOutput,
		"benchDuration", benchDuration,
		"outputFormat", outputFormat)

	PrintResults(
		actualTxPerSec,
		latencyTxPerSec,
		latencyPerTx,
		failures,
		iterations,
		parallelism,
		utxoInput,
		utxoOutput,
		benchDuration,
		outputFormat,
	)
}

func buildTransaction(utxos []UTxO.UTxO, addr *Address.Address, ctx Base.ChainContext, utxoOutput int) error {
	slog.Debug("Building transaction", "address", addr.String(), "utxoOutput", utxoOutput)

	apolloBE := apollo.New(ctx).
		SetWalletFromBech32(addr.String()).
		AddLoadedUTxOs(utxos...).
		SetChangeAddress(*addr).
		AddRequiredSigner(serialization.PubKeyHash(addr.PaymentPart))

	// Add multiple outputs
	for i := 0; i < utxoOutput; i++ {
		apolloBE = apolloBE.PayToAddress(*addr, 2_000_000)
	}
	_, err := apolloBE.Complete()
	if err != nil {
		slog.Error("Transaction completion failed", "error", err)
	} else {
		slog.Debug("Transaction completed successfully")
	}

	return err
}
