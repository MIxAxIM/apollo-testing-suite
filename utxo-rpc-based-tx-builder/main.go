package main

import (
	"encoding/hex"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"

	"github.com/Salvionied/apollo"
	"github.com/Salvionied/apollo/constants"
	"github.com/Salvionied/apollo/txBuilding/Backend/UtxorpcChainContext"
	"github.com/Salvionied/apollo/txBuilding/Utils"
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      LOG_LEVEL,
			TimeFormat: time.StampMilli,
		}),
	))
}

func main() {
	slog.Info("Starting UTXO RPC based tx building script.")
	WalletSetup()

	slog.Info("Initializing UtxorpcChainContext...")
	be, err := UtxorpcChainContext.NewUtxorpcChainContext(UtxorpcBaseUrl, int(constants.PREVIEW))
	if err != nil {
		slog.Error("Error creating UtxorpcChainContext", "error", err)
		os.Exit(1)
	}
	slog.Info("UtxorpcChainContext initialized successfully.")

	apolloBE := apollo.New(&be)
	slog.Debug("Apollo backend initialized.")

	slog.Info("Fetching UTXOs for target address...")
	userUtxos, err := be.Utxos(GetUser1Wallet().Address)
	if err != nil {
		slog.Error("Error getting UTXOs", "error", err, "address", GetUser1Wallet().Address.String())
		os.Exit(1)
	}
	slog.Info("UTXOs fetched successfully.", "count", len(userUtxos))
	slog.Debug("User UTXOs", "utxos", userUtxos)

	slog.Info("Building transaction...")
	apolloBE, err = apolloBE.
		SetWalletFromBech32(GetUser1Wallet().Address.String()).
		SetChangeAddress(GetUser1Wallet().Address).
		AddLoadedUTxOs(userUtxos...).
		PayToAddress(GetUser2Wallet().Address, AMOUNT_TO_SEND).
		AddRequiredSigner(GetUser1Wallet().PKH).
		Complete()

	if err != nil {
		slog.Error("Error completing transaction", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction built successfully.")
	slog.Debug("Transaction details", "apolloBE", apolloBE)

	slog.Info("Signing transaction...")
	apolloBE, err = apolloBE.SignWithSkey(GetUser1Wallet().Vkey, GetUser1Wallet().Skey)
	if err != nil {
		slog.Error("Error signing transaction", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction signed successfully.")

	tx := apolloBE.GetTx()
	slog.Debug("Transaction object retrieved.")

	slog.Info("Converting transaction to CBOR...")
	cbor, err := Utils.ToCbor(tx)
	if err != nil {
		slog.Error("Error converting transaction to CBOR", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction converted to CBOR.")
	slog.Info("Tx CBOR:", "cbor", cbor)

	slog.Info("Submitting transaction...")
	txHash, err := apolloBE.Submit()
	if err != nil {
		slog.Error("Error submitting transaction", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction submitted successfully.")

	txID := hex.EncodeToString(txHash.Payload)
	slog.Info("TxID:", "txid", txID)
	slog.Info("You can check the transaction on the blockchain explorer:", "link", "https://preview.cexplorer.io/tx/"+txID)
	slog.Info("Script finished.")
}
