package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type BenchmarkResult struct {
	WallClockTPS  float64       `json:"wall_clock_tps"`
	LatencyTPS    float64       `json:"latency_tps"`
	AvgLatency    time.Duration `json:"avg_latency"`
	Failures      int           `json:"failures"`
	Iterations    int           `json:"iterations"`
	Parallelism   int           `json:"parallelism"`
	UTXOInput     int           `json:"utxo_input"`
	UTXOOutput    int           `json:"utxo_output"`
	SystemInfo    SystemInfo    `json:"system_info"`
	BenchDuration time.Duration `json:"bench_duration"`
}

func PrintResults(wallClockTPS, latencyTPS float64, avgLatency time.Duration,
	failures, iterations, parallelism, utxoInput, utxoOutput int, benchDuration time.Duration,
	format string) {

	result := BenchmarkResult{
		WallClockTPS:  wallClockTPS,
		LatencyTPS:    latencyTPS,
		AvgLatency:    avgLatency,
		Failures:      failures,
		Iterations:    iterations,
		Parallelism:   parallelism,
		UTXOInput:     utxoInput,
		UTXOOutput:    utxoOutput,
		SystemInfo:    GetSystemInfo(),
		BenchDuration: benchDuration,
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			color.Red("Failed to encode JSON: %v", err)
			os.Exit(1)
		}
	default:
		printColorfulTable(result)
	}
}

func printColorfulTable(result BenchmarkResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Metric", "Value", "Description"})
	table.SetBorder(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetColWidth(1000)

	// Custom function for section headers
	addSectionHeader := func(title string) {
		table.Append([]string{color.HiMagentaString(title), "", ""})
	}

	// Throughput Metrics Section
	addSectionHeader("THROUGHPUT METRICS")
	addRow(table, "Wall-clock Tx/s", fmt.Sprintf("%.2f", result.WallClockTPS),
		"Actual transactions processed per second")
	addRow(table, "Latency-based Tx/s", fmt.Sprintf("%.2f", result.LatencyTPS),
		"Theoretical maximum based on average latency")
	addRow(table, "Avg Latency/Transaction", result.AvgLatency.Round(time.Microsecond).String(),
		"Mean time to build and validate one transaction")

	// Failure Analysis Section
	addSectionHeader("FAILURE ANALYSIS")
	failureStatus := fmt.Sprintf("%d/%d", result.Failures, result.Iterations)
	if result.Failures > 0 {
		failureStatus = color.HiRedString(failureStatus)
	} else {
		failureStatus = color.HiGreenString(failureStatus)
	}
	addRow(table, "Failed Transactions", failureStatus,
		"Total failed transaction constructions")

	// Configuration Section
	addSectionHeader("BENCHMARK CONFIGURATION")
	addRow(table, "Iterations", strconv.Itoa(result.Iterations), "")
	addRow(table, "Parallel Workers", strconv.Itoa(result.Parallelism), "")
	addRow(table, "Inputs per TX", strconv.Itoa(result.UTXOInput), "")
	addRow(table, "Outputs per TX", strconv.Itoa(result.UTXOOutput), "")
	addRow(table, "Total Duration", result.BenchDuration.Round(time.Millisecond).String(), "")

	// System Info Section
	addSectionHeader("SYSTEM INFORMATION")
	addRow(table, "CPU Model", result.SystemInfo.CPUModel, "")
	addRow(table, "Total Memory", fmt.Sprintf("%d GB", result.SystemInfo.TotalMemory/1e9), "")
	addRow(table, "Available Memory", fmt.Sprintf("%d GB", result.SystemInfo.AvailableMem/1e9), "")
	addRow(table, "Go Version", result.SystemInfo.GoVersion, "")
	addRow(table, "OS/Arch", fmt.Sprintf("%s/%s", result.SystemInfo.OS, runtime.GOARCH), "")

	// Efficiency Section
	addSectionHeader("EFFICIENCY ANALYSIS")
	efficiency := (result.WallClockTPS / result.LatencyTPS) * 100
	addRow(table, "Throughput Efficiency",
		fmt.Sprintf("%.1f%%", efficiency),
		"Ratio of actual vs theoretical maximum throughput")

	table.Render()
}

func addRow(table *tablewriter.Table, metric, value, description string) {
	table.Append([]string{metric, value, description})
}
