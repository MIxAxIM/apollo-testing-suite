package benchmark

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/Salvionied/apollo"
	"github.com/Salvionied/apollo/serialization"
	"github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/serialization/MultiAsset"
	"github.com/Salvionied/apollo/serialization/UTxO"
	"github.com/Salvionied/apollo/txBuilding/Backend/Base"
)

type Result struct {
	Duration time.Duration
	Error    error
}

func Run(utxoCount, iterations, parallelism int, backend string, outputFormat string, cpuProfile string,
	maestroNetworkID int, maestroAPIKey string,
	blockfrostAPIURL string, blockfrostNetworkID int, blockfrostAPIKey string,
	ogmiosEndpoint string, kugoEndpoint string,
	utxorpcEndpoint string, utxorpcNetworkID int, utxorpcAPIKey string) {

	slog.Info("Starting benchmark run",
		"utxoCount", utxoCount,
		"iterations", iterations,
		"parallelism", parallelism,
		"backend", backend,
		"outputFormat", outputFormat,
		"cpuProfile", cpuProfile)

	ctx, err := GetChainContext(backend,
		maestroNetworkID, maestroAPIKey,
		blockfrostAPIURL, blockfrostNetworkID, blockfrostAPIKey,
		ogmiosEndpoint, kugoEndpoint,
		utxorpcEndpoint, utxorpcNetworkID, utxorpcAPIKey)
	if err != nil {
		slog.Error("Critical error getting backend chain context", "error", err)
		os.Exit(1)
	}
	slog.Info("Chain context obtained successfully", "backend", backend)

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

	userUtxos, err := ctx.Utxos(senderWalletAddress)
	if err != nil {
		slog.Error("Failed to get UTXOs", "error", err)
		os.Exit(1)
	}
	slog.Info("Fetched user UTXOs", "count", len(userUtxos))

	// Aggregate & merge assets upfront
	var assetList []apollo.Unit
	for _, utxo := range userUtxos {
		assetList = append(assetList, MultiAssetToUnits(utxo.Output.GetValue().GetAssets())...)
	}
	assetList = MergeUnits(assetList)
	slog.Debug("Aggregated and merged assets", "count", len(assetList))

	lastSlot, err := ctx.LastBlockSlot()
	if err != nil {
		slog.Error("Failed to get last block slot", "error", err)
		os.Exit(1)
	}
	slog.Info("Last block slot obtained", "slot", lastSlot)

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
			err := buildTransaction(clonedUTxOs, &receiverWalletAddress, ctx, lastSlot, utxoCount, assetList)
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
		"utxoCount", utxoCount,
		"benchDuration", benchDuration,
		"outputFormat", outputFormat)

	PrintResults(
		actualTxPerSec,
		latencyTxPerSec,
		latencyPerTx,
		failures,
		iterations,
		parallelism,
		utxoCount,
		benchDuration,
		outputFormat,
	)
}

func buildTransaction(utxos []UTxO.UTxO, addr *Address.Address, ctx Base.ChainContext, lastSlot int, utxoCount int, assets []apollo.Unit) error {
	slog.Debug("Building transaction", "address", addr.String(), "utxoCount", utxoCount)
	apolloBE := apollo.New(ctx).
		SetWalletFromBech32(addr.String()).
		AddLoadedUTxOs(utxos...).
		SetChangeAddress(*addr).
		AddRequiredSigner(serialization.PubKeyHash(addr.PaymentPart)).
		SetTtl(int64(lastSlot) + 300)

	// Add multiple outputs
	distributedAssets := DistributeAssets(assets, utxoCount)
	slog.Debug("Distributed assets for transaction", "numOutputs", len(distributedAssets))
	for _, assetGroup := range distributedAssets {
		apolloBE = apolloBE.PayToAddress(*addr, 2_000_000, assetGroup...)
	}

	_, err := apolloBE.Complete()
	if err != nil {
		slog.Error("Transaction completion failed", "error", err)
	} else {
		slog.Debug("Transaction completed successfully")
	}

	return err
}

func MultiAssetToUnits(ma MultiAsset.MultiAsset[int64]) []apollo.Unit {
	// slog.Debug("Converting multi-asset to units", "multiAsset", ma) // Too verbose
	units := make([]apollo.Unit, 0, len(ma))
	for policyId, assets := range ma {
		if policyId.Value == "" {
			continue
		}
		for assetName, quantity := range assets {
			units = append(units, apollo.Unit{
				PolicyId: policyId.Value,
				Name:     assetName.String(),
				Quantity: int(quantity),
			})
		}
	}
	return units
}

func MergeUnits(units []apollo.Unit) []apollo.Unit {
	slog.Debug("Merging units", "initialCount", len(units))
	unitMap := make(map[string]apollo.Unit, len(units))

	for _, unit := range units {
		key := unit.PolicyId + ":" + unit.Name
		if existing, found := unitMap[key]; found {
			existing.Quantity += unit.Quantity
			unitMap[key] = existing
		} else {
			unitMap[key] = unit
		}
	}

	mergedUnits := make([]apollo.Unit, 0, len(unitMap))
	for _, unit := range unitMap {
		mergedUnits = append(mergedUnits, unit)
	}
	slog.Debug("Units merged", "finalCount", len(mergedUnits))
	return mergedUnits
}

// Distribute assets across multiple outputs
func DistributeAssets(units []apollo.Unit, utxoCount int) [][]apollo.Unit {
	slog.Debug("Distributing assets", "totalUnits", len(units), "utxoCount", utxoCount)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	outputs := make([][]apollo.Unit, utxoCount)

	for _, unit := range units { // Corrected: 'unit' is now correctly scoped
		remaining := unit.Quantity
		for remaining > 0 {
			idx := rnd.Intn(utxoCount)
			quantity := min(remaining, rnd.Intn(remaining/2+1)+1)
			outputs[idx] = append(outputs[idx], apollo.Unit{
				PolicyId: unit.PolicyId,
				Name:     unit.Name,
				Quantity: quantity,
			})
			remaining -= quantity
		}
	}
	slog.Debug("Assets distributed", "numOutputs", len(outputs))
	return outputs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
