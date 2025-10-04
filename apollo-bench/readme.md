# Apollo-Bench: Cardano Transaction Benchmark Tool

Apollo-Bench is a benchmarking tool for the Apollo Cardano transaction–building library written in Golang. It is designed to stress–test key parts of the library, namely:

- **UTXO Selection:** Simulates the process of selecting UTXOs from a wallet.

The tool runs multiple iterations concurrently and computes performance metrics such as transactions per second (TPS), average latency per transaction, and theoretical throughput based on latency. Results can be output as a pretty table or in JSON format.

---

## Features

- **Throughput Metrics:**  
  - **Wall-clock Tx/s:** Actual transactions built per second (calculated as the number of successful iterations divided by the total elapsed time).
  - **Latency-based Tx/s:** Theoretical maximum based on average transaction latency.
  - **Average Latency:** Mean time to build and serialize a transaction.
  
- **Failure Analysis:** Reports any failed transaction builds.
- **Configurable Benchmarking:**  
  - Specify number of iterations, UTXO count, and parallel workers.
  - Choose among different UTXO generation levels (simple, differentiated, congested).
- **System Information:** Displays CPU model, total and available memory, Go version, and OS/Arch.
- **Optional CPU Profiling:** Write a CPU profile to a file for further performance analysis.

---

## Requirements

- **Go:** Version 1.18 or higher is recommended.
- **Dependencies:**  
  - [Cobra](https://github.com/spf13/cobra) – For CLI flag parsing.
  - [Tint](https://github.com/lmittmann/tint) – For colored and structured logging.
  - [gopsutil](https://github.com/shirou/gopsutil) – For gathering system metrics.
  - [jq](https://stedolan.github.io/jq/) - For parsing JSON output in the comparison script.

---

## Installation

Clone this repository and build the benchmark binary:

```bash
git clone https://github.com/Salvionied/apollo-testing-suite.git
cd apollo-testing-suite
go build -o ./bin/apollo-bench ./cmd/benchmark
```

---

## Configuration

Apollo-Bench relies on some environment constants defined in `internal/benchmark/consts.go`. You can change these values to test other wallet addresses:

```go
TEST_WALLET_ADDRESS_1 string
TEST_WALLET_ADDRESS_2 string
```

---

## Usage

Run the benchmark binary:

```bash
./bin/apollo-bench [flags]
```

### Available Flags

- `--utxo-input`, `-u` (default: **10**)  
  *Number of UTXOs to use as input.* This simulates the number of UTXO inputs for each transaction.

- `--utxo-output`, `-v` (default: **10**)  
  *Number of UTXOs to generate as output.* This simulates the number of UTXO outputs to include in each transaction.

- `--utxo-level` (default: **1**)  
  *Set UTXO generation level.* Options:
  - `1`: Simple UTXOs (all identical).
  - `2`: Differentiated UTXOs (varied values and assets).
  - `3`: Congested UTXOs (many small UTXOs, simulating a busy wallet).

- `--iterations`, `-i` (default: **1000**)  
  *Number of transactions to build.* This defines the total number of iterations for the benchmark run.

- `--parallelism`, `-p` (default: **4**)  
  *Number of parallel goroutines.* Controls the concurrency level during the benchmark.

- `--output`, `-o` (default: **"table"**)  
  *Output format for results.* Options:
  - `table`: Displays a formatted, colorful table.
  - `json`: Outputs results as formatted JSON.

- `--cpu-profile`, `-c` (default: **""**)  
  *Writes CPU profiling data to the specified file.*  
  Example:

  ```bash
  ./bin/apollo-bench -c cpu.prof
  ```

  You can then analyze the profile with:

  ```bash
  go tool pprof cpu.prof
  ```

- `--log-level` (default: **"info"**)  
  *Set logging level.* Options: `debug`, `info`, `warn`, `error`.

---

## Examples

### Basic Test Run

```bash
./bin/apollo-bench --utxo-input 50 --utxo-output 50 --iterations 5000 --parallelism 8
```

### Differentiated UTXO Benchmark with JSON Output

```bash
./bin/apollo-bench --utxo-level 2 --utxo-input 100 --utxo-output 100 -o json
```

### Congested UTXO Benchmark with CPU Profiling

```bash
./bin/apollo-bench --utxo-level 3 --utxo-input 200 --utxo-output 200 -c profile.out
```

---

## Benchmark Metrics & How It Works

### Key Metrics

1. **Wall-clock Tx/s**  
   - **Formula:**  

    ```plaintext
    TPS = Number of Successful Iterations / Total Elapsed Time (seconds)
    ```

   - Represents the actual throughput achieved during the benchmark.

2. **Latency-based Tx/s**  
   - **Formula:**  

     ```plaintext
     Latency-based TPS = 1 second / Average Latency per Transaction
     ```

   - Represents the theoretical maximum throughput based on average transaction latency.

3. **Average Latency**  
   - **Formula:**  

     ```plaintext
     Average Latency = Total Latency / Number of Successful Iterations
     ```

   - Measures the mean time to build and serialize a transaction.

### Benchmark Workflow

1. **Setup:**
   - Initialize the chain context using `FixedChainContext`.
   - Decode test wallet addresses (`TEST_WALLET_ADDRESS_1`, `TEST_WALLET_ADDRESS_2`).
   - Generate UTXOs based on the specified `utxo-level`.
   - Run a warm-up phase (GC + 2-second sleep).

2. **Transaction Building:**
   - For each iteration:
     - Clone UTXOs for thread safety.
     - Build and serialize the transaction using the Apollo library.
     - Record latency and track failures.

3. **Results Calculation:**
   - Compute wall-clock TPS, latency-based TPS, and average latency.
   - Generate system diagnostics.

4. **Output:**
   - Print results as a colorful table or JSON.

---

## Benchmarking Script: `scripts/compare_versions.sh`

The `scripts/compare_versions.sh` script allows users to benchmark multiple versions, tags, or commit hashes of the Apollo library by providing them as command-line parameters. The script automates the process of checking out different versions, building the benchmark tool, running benchmarks, and comparing results.

### How It Works

1. **Version Specification:** Users provide one or more Apollo library versions (e.g., Git commit hashes, tags like `v1.3.0`) as arguments to the script.
2. **Temporary Environment:** For each specified version:
   - The script temporarily modifies the `go.mod` file to point to the specified Apollo library version.
   - `go mod tidy` is run to resolve dependencies.
   - The `apollo-bench` binary is built for that specific version and saved with a unique name (e.g., `apollo-bench-v1.3.0`) in the `bin/` directory.
3. **Benchmark Execution:** The newly built binary is executed with a predefined set of benchmark parameters (UTXO input/output, iterations, parallelism). Each version is benchmarked for `NUM_TRIALS` (default: 5) times.
4. **Result Storage:** The JSON output from each benchmark trial is saved in the `scripts/results/` directory (e.g., `scripts/results/v1.3.0_trial1.json`).
5. **Version Restoration:** After benchmarking a version, the original `go.mod` and `go.sum` files are restored.
6. **Analysis and Comparison:** After all versions have been benchmarked, the script analyzes the collected JSON results:
   - It calculates the average transactions per second (Tx/s) for each version across all its trials.
   - If multiple versions were provided, it performs pairwise comparisons, showing the difference in Tx/s and the percentage change.
7. **Output:** The analysis results are printed to the console in a formatted box and saved to a Markdown file in `scripts/results/` (e.g., `comparison_results_YYYYMMDD_HHMMSS.md`).

### Usage Example

To benchmark multiple versions, run:

```bash
./scripts/compare_versions.sh 99d52bbc93e4a774d2f24bcabd03df7e9cd1ab12 v1.3.0
```

This will:

- Temporarily set the Apollo dependency to `99d52bbc93e4a774d2f24bcabd03df7e9cd1ab12` and `v1.3.0` respectively.
- Build the benchmark binary for each version.
- Run 5 trials of the benchmark for each version.
- Save individual trial results (e.g., `scripts/results/99d52bbc93e4a774d2f24bcabd03df7e9cd1ab12_trial1.json`).
- Generate a comparison Markdown file (e.g., `scripts/results/comparison_results_20251003_045315.md`).

Users can provide as many versions as needed, making it easy to compare different releases or commits efficiently.

---

Happy benchmarking!
