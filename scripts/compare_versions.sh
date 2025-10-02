#!/bin/bash
set -e # Exit on error

# Determine directories (assuming script is in apollo/scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)" # Root is 'apollo'
RESULTS_DIR="$SCRIPT_DIR/results"
BIN_DIR="$ROOT_DIR/bin"

# Ensure the results and bin directories exist
mkdir -p "$RESULTS_DIR" "$BIN_DIR"

# Check if at least one version is passed as parameter
if [ "$#" -lt 1 ]; then
    echo "Usage: $0 version1 [version2 version3 ...]"
    exit 1
fi

# The versions array is created from all command-line arguments
VERSIONS=("$@")

# Color variables
CYAN="\033[1;36m"
GREEN="\033[1;32m"
RED="\033[1;31m"
RESET="\033[0m"

for version in "${VERSIONS[@]}"; do
    # Save original go.mod and go.sum
    cp go.mod go.mod.bak
    cp go.sum go.sum.bak

    printf "${CYAN}[%s] Setting Apollo dependency to version: %s...${RESET}\n" "$(date '+%Y-%m-%d %H:%M:%S')" "$version"
    # Use 'go mod edit -replace' to point to the specific version/commit
    if ! go mod edit -replace github.com/Salvionied/apollo=github.com/Salvionied/apollo@"$version"; then
        echo -e "${RED}Error: Failed to set Apollo dependency to $version${RESET}"
        mv go.mod.bak go.mod
        mv go.sum.bak go.sum
        exit 1
    fi
    if ! go mod tidy; then
        echo -e "${RED}Error: go mod tidy failed for $version${RESET}"
        mv go.mod.bak go.mod
        mv go.sum.bak go.sum
        exit 1
    fi

    printf "${CYAN}[%s] Building binary for version: %s...${RESET}\n" "$(date '+%Y-%m-%d %H:%M:%S')" "$version"
    if ! go build -o "$BIN_DIR/apollo-bench-$version" ./cmd/benchmark; then
        echo -e "${RED}Error: Build failed for $version${RESET}"
        mv go.mod.bak go.mod
        mv go.sum.bak go.sum
        exit 1
    fi

    printf "${CYAN}[%s] Benchmarking version: %s...${RESET}\n" "$(date '+%Y-%m-%d %H:%M:%S')" "$version"
    if ! "$BIN_DIR/apollo-bench-$version" \
        --utxo-count 100 \
        --iterations 10000 \
        --parallelism 10 \
        --backend maestro \
        --output json \
        --maestro-network-id 3 \
        --maestro-api-key "so4a45BCnj80EdcFa9OwLr8pK8um4bWE" \
        --blockfrost-api-url "https://cardano-preprod.blockfrost.io/api" \
        --blockfrost-network-id 0 \
        --blockfrost-api-key "preprod9zzl4g8Xa3faU50a1OVDZdPeQ92ZsdcT" \
        --ogmios-endpoint "ws://localhost:1337" \
        --kugo-endpoint "http://localhost:1442" \
        >"$RESULTS_DIR/$version.json"; then
        echo -e "${RED}Error: Benchmark failed for $version${RESET}"
    else
        echo -e "${GREEN}Benchmark successful for $version! Output stored in $RESULTS_DIR/$version.json${RESET}"
    fi

    # Restore original go.mod and go.sum
    mv go.mod.bak go.mod
    mv go.sum.bak go.sum
done
