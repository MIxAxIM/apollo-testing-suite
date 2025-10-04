# UTXO RPC Example

This directory contains a Go program that demonstrates how to interact with the Cardano blockchain using the Apollo transaction-building library and Utxorpc as the backend.

## Features

- **Wallet Setup**: Initializes two wallets using mnemonics (from environment variables).
- **Chain Context Initialization**: Sets up the `UtxorpcChainContext` to connect to the Cardano `preview` network.
- **UTXO Fetching**: Retrieves Unspent Transaction Outputs (UTXOs) for a specified wallet address.
- **Transaction Building**: Constructs a simple transaction to send a fixed amount of Ada from one wallet to another.
- **Transaction Signing**: Signs the built transaction with the appropriate signing key.
- **Transaction Submission**: Submits the signed transaction to the Cardano network via Utxorpc.
- **Transaction ID and Explorer Link**: Provides the transaction ID and a link to view the transaction on CardanoScan.

## Requirements

- Go (version 1.18 or higher recommended)
- Environment variables `USER1_MNEMONIC` and `USER2_MNEMONIC` (optional, dummy mnemonics will be used if not set).

## Setup and Run

1. **Navigate to the `utxo-rpc-based-tx-builder` directory**:

    ```bash
    cd utxo-rpc-based-tx-builder
    ```

2. **Set Environment Variables (Optional but Recommended)**:

    ```bash
    export USER1_MNEMONIC="your user 1 mnemonic phrase here"
    export USER2_MNEMONIC="your user 2 mnemonic phrase here"
    ```

    Replace `"your user 1 mnemonic phrase here"` and `"your user 2 mnemonic phrase here"` with actual 24-word mnemonic phrases for testing. If these are not set, the program will use dummy mnemonics, which are not suitable for real transactions.

3. **Run the example**:

    ```bash
    go run .
    ```

    The script will output detailed logs of the transaction process, including wallet addresses, UTXO fetching, transaction building, signing, submission, and the final transaction ID with a link to the CardanoScan explorer.

## Configuration

The following constants can be modified in [`config.go`](utxo-rpc-based-tx-builder/config.go) to change the behavior of the example:

- `LOG_LEVEL`: Adjust the logging verbosity (e.g., `slog.LevelDebug`, `slog.LevelInfo`).
- `NETWORK`: Change the Cardano network (e.g., `"preview"`, `"preprod"`).
- `AMOUNT_TO_SEND`: Modify the amount of Ada to be sent in the transaction.
- `UtxorpcBaseUrl`: Update the Utxorpc endpoint if needed.
