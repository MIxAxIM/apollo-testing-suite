package benchmark

import (
	"fmt"
	"log/slog"

	"github.com/Salvionied/apollo/txBuilding/Backend/Base"
	"github.com/Salvionied/apollo/txBuilding/Backend/BlockFrostChainContext"
	"github.com/Salvionied/apollo/txBuilding/Backend/MaestroChainContext"
	"github.com/Salvionied/apollo/txBuilding/Backend/OgmiosChainContext"
	"github.com/Salvionied/apollo/txBuilding/Backend/UtxorpcChainContext"
	"github.com/SundaeSwap-finance/kugo"
	"github.com/SundaeSwap-finance/ogmigo/v6"
)

type ChainContext interface{}

func OgmiosCTXSetup(ogmiosEndpoint string, kugoEndpoint string) OgmiosChainContext.OgmiosChainContext {
	slog.Debug("Setting up Ogmios context", "ogmiosEndpoint", ogmiosEndpoint, "kugoEndpoint", kugoEndpoint)
	return OgmiosChainContext.NewOgmiosChainContext(*ogmigo.New(ogmigo.WithEndpoint(ogmiosEndpoint)), *kugo.New(kugo.WithEndpoint(kugoEndpoint)))
}

func UtxorpcCTXSetup(utxorpcEndpoint string, utxorpcNetworkID int, utxorpcAPIKey string) (UtxorpcChainContext.UtxorpcChainContext, error) {
	slog.Debug("Setting up Utxorpc context", "utxorpcEndpoint", utxorpcEndpoint, "utxorpcNetworkID", utxorpcNetworkID)
	return UtxorpcChainContext.NewUtxorpcChainContext(utxorpcEndpoint, utxorpcNetworkID, utxorpcAPIKey)
}

func BlockfrostCTXSetup(blockfrostAPIURL string, blockfrostNetworkID int, blockfrostAPIKey string) (bfc BlockFrostChainContext.BlockFrostChainContext, err error) {
	slog.Debug("Setting up Blockfrost context", "blockfrostAPIURL", blockfrostAPIURL, "blockfrostNetworkID", blockfrostNetworkID)
	bfc, err = BlockFrostChainContext.NewBlockfrostChainContext(
		blockfrostAPIURL,
		blockfrostNetworkID,
		blockfrostAPIKey,
	)

	if err != nil {
		slog.Error("Failed to create Blockfrost chain context", "error", err)
		return BlockFrostChainContext.BlockFrostChainContext{}, err
	}
	slog.Debug("Blockfrost chain context created successfully")
	return bfc, nil

}

func MaestroCTXSetup(maestroNetworkID int, maestroAPIKey string) (mc MaestroChainContext.MaestroChainContext, err error) {
	slog.Debug("Setting up Maestro context", "maestroNetworkID", maestroNetworkID, "maestroAPIKey", maestroAPIKey)
	mc, err = MaestroChainContext.NewMaestroChainContext(
		maestroNetworkID,
		maestroAPIKey,
	)

	if err != nil {
		slog.Error("MaestroChainContext.NewMaestroChainContext failed", "error", err, "maestroNetworkID", maestroNetworkID, "maestroAPIKey", maestroAPIKey)
		return MaestroChainContext.MaestroChainContext{}, fmt.Errorf("MaestroChainContext.NewMaestroChainContext failed: %w", err)
	}
	slog.Debug("Maestro chain context created successfully")
	return mc, nil

}

func GetChainContext(backend string,
	maestroNetworkID int, maestroAPIKey string,
	blockfrostAPIURL string, blockfrostNetworkID int, blockfrostAPIKey string,
	ogmiosEndpoint string, kugoEndpoint string,
	utxorpcEndpoint string, utxorpcNetworkID int, utxorpcAPIKey string) (Base.ChainContext, error) {

	slog.Debug("Getting chain context", "backend", backend)
	switch backend {
	case "maestro":
		slog.Debug("Attempting to create Maestro chain context", "maestroNetworkID", maestroNetworkID, "maestroAPIKey", maestroAPIKey)
		mc, err := MaestroCTXSetup(maestroNetworkID, maestroAPIKey)
		if err != nil {
			slog.Error("Failed to create Maestro chain context", "error", err)
			return nil, fmt.Errorf("failed to create Maestro chain context: %w", err)
		}
		slog.Info("Created Maestro chain context")
		return &mc, nil
	case "blockfrost":
		slog.Debug("Attempting to create Blockfrost chain context", "blockfrostAPIURL", blockfrostAPIURL, "blockfrostNetworkID", blockfrostNetworkID, "blockfrostAPIKey", blockfrostAPIKey)
		bfc, err := BlockfrostCTXSetup(blockfrostAPIURL, blockfrostNetworkID, blockfrostAPIKey)
		if err != nil {
			slog.Error("Failed to create Blockfrost chain context", "error", err)
			return nil, fmt.Errorf("failed to create Blockfrost chain context: %w", err)
		}
		slog.Info("Created Blockfrost chain context")
		return &bfc, nil
	case "ogmios":
		ctx := OgmiosCTXSetup(ogmiosEndpoint, kugoEndpoint)
		slog.Info("Created Ogmios chain context")
		return &ctx, nil
	case "utxorpc":
		ctx, err := UtxorpcCTXSetup(utxorpcEndpoint, utxorpcNetworkID, utxorpcAPIKey)
		if err != nil {
			slog.Error("Failed to create Utxorpc chain context", "error", err)
			return nil, fmt.Errorf("failed to create Utxorpc chain context: %w", err)
		}
		slog.Info("Created Utxorpc chain context")
		return &ctx, nil
	default:
		slog.Error("Unknown backend", "backend", backend)
		return nil, fmt.Errorf("unknown backend: %s", backend)
	}
}
