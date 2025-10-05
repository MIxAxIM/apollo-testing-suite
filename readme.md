# Apollo Testing Suite

This repository contains a suite of tools and examples for testing and benchmarking the Apollo Cardano transaction-building library.

## Project Structure

- [`utxo-rpc-based-tx-builder/`](utxo-rpc-based-tx-builder/): Contains an example Go program demonstrating how to interact with the Cardano blockchain using the Apollo library and Utxorpc.
- [`plutus-v3-sc-tx-builder/`](plutus-v3-sc-tx-builder/): Provides a transaction builder for Plutus V3 smart contracts.
- [`apollo-bench/`](apollo-bench/): Houses the `Apollo-Bench` benchmarking tool for stress-testing the Apollo library's performance.

## UTXO RPC Example

The `utxo-rpc-based-tx-builder` directory provides a simple Go application that showcases the use of the Apollo library to perform basic Cardano blockchain operations, such as:

- Wallet setup using mnemonics.
- Fetching UTXOs for a given address.
- Building and signing transactions.
- Submitting transactions to the Cardano network via Utxorpc.

## Plutus V3 Smart Contract Transaction Builder

The `plutus-v3-sc-tx-builder` directory contains a Go project that demonstrates building, signing, evaluating, and submitting transactions interacting with a Plutus V3 escrow contract on the Cardano blockchain. It leverages the Apollo library for transaction building and Blockfrost for chain context interaction and transaction submission.

## Apollo-Bench: Cardano Transaction Benchmark Tool

The `apollo-bench` directory contains `Apollo-Bench`, a dedicated benchmarking tool for the Apollo Cardano transaction-building library. This tool is designed to stress-test key parts of the library, particularly UTXO selection, and measure performance metrics like transactions per second (TPS) and average latency.

For detailed information on features, requirements, installation, configuration, usage, available flags, examples, benchmark metrics, and the `compare_versions.sh` script, please refer to the dedicated README:

- [`apollo-bench/readme.md`](apollo-bench/readme.md)
