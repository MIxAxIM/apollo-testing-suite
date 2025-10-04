package main

import (
	"log/slog"
)

const (
	LOG_LEVEL             = slog.LevelInfo
	NETWORK        string = "preview"
	AMOUNT_TO_SEND int    = 5_000_000
	UtxorpcBaseUrl        = "https://preview-utxorpc.blinklabs.io"
)