#!/bin/bash
set -e # Exit on error

# Determine directories (assuming script is in apollo/scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)" # Root is 'apollo'
RESULTS_DIR="$SCRIPT_DIR/results"
BIN_DIR="$ROOT_DIR/bin"

# Color variables
CYAN="\033[1;36m"
GREEN="\033[1;32m"
RED="\033[1;31m"
WHITE="\033[1;37m" # Added WHITE color
RESET="\033[0m"

# Function to display a loading animation
show_progress() {
    local pid=$1
    local base_message=$2
    shift 2
    local args=("$@")
    local delay=0.1
    local spin_chars=".oOo.-\\||//-.oOo." # Cooler ASCII spin characters
    local i=0
    tput civis # Hide cursor
    while kill -0 "$pid" 2>/dev/null; do
        i=$(( (i+1) % ${#spin_chars} ))
        # Use printf -v to format the message with arguments, then print it
        printf -v formatted_message "$base_message" "${args[@]}"
        printf "\r${CYAN}[%s] %s %c${RESET}" "$(date '+%Y-%m-%d %H:%M:%S')" "$formatted_message" "${spin_chars:$i:1}"
        sleep "$delay"
    done
    tput cnorm # Show cursor
    printf "\r%s" "$(tput el)" # Clear the line without a newline
}

# Function to print a static status message with colors
print_status() {
    local color=$1
    local base_message=$2
    shift 2
    local args=("$@")
    printf -v formatted_message "$base_message" "${args[@]}"
    printf "${color}[%s] %s${RESET}\n" "$(date '+%Y-%m-%d %H:%M:%S')" "$formatted_message"
}

# Function to draw a box around text
draw_box() {
    local content="$1"
    local lines
    IFS=$'\n' read -r -d '' -a lines <<< "$content"

    local max_len=0
    for line in "${lines[@]}"; do
        # Remove ANSI escape codes for length calculation
        local clean_line
        clean_line=$(printf "%s" "$line" | sed -r "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[mGK]//g")
        if (( ${#clean_line} > max_len )); then
            max_len=${#clean_line}
        fi
    done

    local horizontal_line_length=$((max_len + 2)) # +2 for padding on each side
    local horizontal_line
    printf -v horizontal_line "%*s" "$horizontal_line_length" ""
    horizontal_line=${horizontal_line// /─}

    printf "${CYAN}┌%s┐${RESET}\n" "$horizontal_line"
    for line in "${lines[@]}"; do
        local clean_line
        clean_line=$(printf "%s" "$line" | sed -r "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[mGK]//g")
        printf "${CYAN}│ ${RESET}%s%*s${CYAN} │${RESET}\n" "$line" $((max_len - ${#clean_line})) ""
    done
    printf "${CYAN}└%s┘${RESET}\n" "$horizontal_line"
}

# Create a temporary directory for temporary files
TMP_DIR=$(mktemp -d -t apollo-benchmark-XXXXXXXXXX)
# print_status "$CYAN" "Created temporary directory: %s" "$TMP_DIR"

# Ensure the temporary directory is removed on script exit
# trap 'print_status "$CYAN" "Cleaning up temporary directory: %s" "$TMP_DIR"; rm -rf "$TMP_DIR"' EXIT

# Ensure the results and bin directories exist
mkdir -p "$RESULTS_DIR" "$BIN_DIR"

# Check if at least one version is passed as parameter
if [ "$#" -lt 1 ]; then
    print_status "$RED" "Usage: %s version1 [version2 version3 ...]" "$0"
    exit 1
fi

# The versions array is created from all command-line arguments
VERSIONS=("$@")
NUM_TRIALS=5 # Number of benchmark trials per version

# Color variables
CYAN="\033[1;36m"
GREEN="\033[1;32m"
RED="\033[1;31m"
RESET="\033[0m"

for version in "${VERSIONS[@]}"; do
    # Save original go.mod and go.sum to the temporary directory
    cp go.mod "$TMP_DIR/go.mod.bak"
    cp go.sum "$TMP_DIR/go.sum.bak"

    (
        go mod edit -replace github.com/Salvionied/apollo=github.com/Salvionied/apollo@"$version"
    ) &
    PID=$!
    show_progress $PID "Setting Apollo dependency to version: %s..." "$version"
    if ! wait $PID; then
        print_status "$RED" "Error: Failed to set Apollo dependency to %s" "$version"
        mv "$TMP_DIR/go.mod.bak" go.mod
        mv "$TMP_DIR/go.sum.bak" go.sum
        exit 1
    fi
    print_status "$GREEN" "Apollo dependency set to version: %s" "$version"

    (
        go mod tidy >/dev/null 2>&1
    ) &
    PID=$!
    show_progress $PID "Running go mod tidy for version: %s..." "$version"
    if ! wait $PID; then
        print_status "$RED" "Error: go mod tidy failed for %s (check logs for details)" "$version"
        mv "$TMP_DIR/go.mod.bak" go.mod
        mv "$TMP_DIR/go.sum.bak" go.sum
        exit 1
    fi
    print_status "$GREEN" "go mod tidy successful for version: %s" "$version"

    (
        go build -o "$BIN_DIR/apollo-bench-$version" ./cmd/benchmark
    ) &
    PID=$!
    show_progress $PID "Building binary for version: %s..." "$version"
    if ! wait $PID; then
        print_status "$RED" "Error: Build failed for %s" "$version"
        mv "$TMP_DIR/go.mod.bak" go.mod
        mv "$TMP_DIR/go.sum.bak" go.sum
        exit 1
    fi
    print_status "$GREEN" "Binary built successfully for version: %s" "$version"

    print_status "$CYAN" "Benchmarking version: %s with %d trials..." "$version" "$NUM_TRIALS"

    # Define benchmark parameters
    UTXO_INPUT=20
    UTXO_OUTPUT=20
    UTXO_LEVEL=2
    ITERATIONS=10000
    PARALLELISM=10
    OUTPUT_FORMAT="json"

    print_status "$CYAN" "Benchmark Parameters: UTXO Input: %d, UTXO Output: %d, UTXO Level: %d, Iterations: %d, Parallelism: %d, Output: %s" \
        "$UTXO_INPUT" "$UTXO_OUTPUT" "$UTXO_LEVEL" "$ITERATIONS" "$PARALLELISM" "$OUTPUT_FORMAT"

    for i in $(seq 1 $NUM_TRIALS); do
        (
            "$BIN_DIR/apollo-bench-$version" \
                --utxo-input "$UTXO_INPUT" \
                --utxo-output "$UTXO_OUTPUT" \
                --utxo-level "$UTXO_LEVEL" \
                --iterations "$ITERATIONS" \
                --parallelism "$PARALLELISM" \
                --output "$OUTPUT_FORMAT"
        ) >"$RESULTS_DIR/${version}_trial${i}.json" 2>/dev/null &
        PID=$!
        show_progress $PID "Running trial %d for version: %s..." "$i" "$version"
        if ! wait $PID; then
            print_status "$RED" "Error: Benchmark trial %d failed for %s" "$i" "$version"
        else
            print_status "$GREEN" "Benchmark trial %d successful for %s! Output stored in %s/${version}_trial%d.json" "$i" "$version" "$RESULTS_DIR" "$i"
        fi
    done

    # Restore original go.mod and go.sum
    mv "$TMP_DIR/go.mod.bak" go.mod
    mv "$TMP_DIR/go.sum.bak" go.sum
done

print_status "$CYAN" "Analyzing benchmark results..."

declare -A avg_tx_per_sec

for version in "${VERSIONS[@]}"; do
    total_tx_per_sec=0
    count=0
    for i in $(seq 1 $NUM_TRIALS); do
        result_file="$RESULTS_DIR/${version}_trial${i}.json"
        if [ -f "$result_file" ]; then
            tx_per_sec=$(jq -r '.wall_clock_tps' "$result_file")
            if [ -n "$tx_per_sec" ]; then
                total_tx_per_sec=$(awk "BEGIN {print $total_tx_per_sec + $tx_per_sec}")
                count=$((count + 1))
            fi
        fi
    done

    if [ "$count" -gt 0 ]; then
        avg_tx_per_sec[$version]=$(awk "BEGIN {printf \"%.2f\", $total_tx_per_sec / $count}")
        print_status "$GREEN" "Average Tx/s for %s: %s" "$version" "${avg_tx_per_sec[$version]}"
    else
        print_status "$RED" "No successful trials found for %s to calculate average Tx/s." "$version"
    fi
done

# Generate a timestamp for the results file
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
RESULTS_FILE="$RESULTS_DIR/comparison_results_$TIMESTAMP.md" # Changed to .md

TEMP_OUTPUT_FILE="$TMP_DIR/temp_analysis_output.txt" # Use temporary directory

(
    printf "${CYAN}# Benchmark Analysis Results\n\n${RESET}"
    printf "${CYAN}Date: %s${RESET}\n\n" "$(date '+%Y-%m-%d %H:%M:%S')"

    printf "${CYAN}## Average Transactions Per Second (Tx/s) Across Versions\n\n${RESET}"
    for version in "${VERSIONS[@]}"; do
        if [ -n "${avg_tx_per_sec[$version]}" ]; then
            printf "${GREEN}* %s: %s Tx/s${RESET}\n" "$version" "${avg_tx_per_sec[$version]}"
        else
            printf "${RED}* %s: No successful trials to calculate average Tx/s.${RESET}\n" "$version"
        fi
    done
    printf "\n"

    if [ "${#VERSIONS[@]}" -gt 1 ]; then
        printf "${CYAN}## Pairwise Comparisons\n\n${RESET}"
        for i in "${!VERSIONS[@]}"; do
            for j in "${!VERSIONS[@]}"; do
                if [ "$i" -lt "$j" ]; then # Ensure each pair is compared only once
                    version1="${VERSIONS[$i]}"
                    version2="${VERSIONS[$j]}"

                    avg1=${avg_tx_per_sec[$version1]}
                    avg2=${avg_tx_per_sec[$version2]}

                    if [ -n "$avg1" ] && [ -n "$avg2" ]; then
                        difference=$(awk "BEGIN {printf \"%.2f\", $avg2 - $avg1}")
                        percentage_diff=$(awk "BEGIN {if ($avg1 != 0) printf \"%.2f\", (($avg2 - $avg1) / $avg1) * 100; else print \"N/A\"}")

                        printf "${CYAN}* %s vs %s:${RESET}\n" "$version2" "$version1"
                        printf "${GREEN}  - Difference: %s Tx/s${RESET}\n" "$difference"

                        if (( $(awk "BEGIN {print ($difference > 0)}") )); then
                            printf "${GREEN}  - %s is faster by %s%%${RESET}\n" "$version2" "$percentage_diff"
                        elif (( $(awk "BEGIN {print ($difference < 0)}") )); then
                            printf "${RED}  - %s is slower by %s%%${RESET}\n" "$version2" "$(awk "BEGIN {print -($percentage_diff)}")"
                        else
                            printf "${WHITE}  - Both versions have similar average Tx/s.${RESET}\n"
                        fi
                    else
                        printf "${RED}* Cannot compare %s with %s: average Tx/s not available for both.${RESET}\n" "$version1" "$version2"
                    fi
                fi
            done
        done
    else
        printf "${WHITE}* Only one version provided. No comparisons to perform.${RESET}\n"
    fi
) > "$TEMP_OUTPUT_FILE" # Redirect analysis output to a temporary file

wait # Wait for the subshell to finish writing to TEMP_OUTPUT_FILE

echo ""

RESULT_OUTPUT=$(cat "$TEMP_OUTPUT_FILE") # Capture the output of the analysis block from the temp file

# Print the colored output to the console
BOXED_OUTPUT=$(draw_box "$RESULT_OUTPUT")
printf "%s\n" "$BOXED_OUTPUT"

# Strip color codes and write to file
printf "%s" "$RESULT_OUTPUT" | sed -r "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[mGK]//g" > "$RESULTS_FILE"

# Clean up the temporary file
rm "$TEMP_OUTPUT_FILE"

echo ""

print_status "$GREEN" "Comparison results saved to: %s" "$RESULTS_FILE"
