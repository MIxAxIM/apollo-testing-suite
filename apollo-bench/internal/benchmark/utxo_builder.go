package benchmark

import (
	"fmt"
	"sort"

	"github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/serialization/Asset"
	"github.com/Salvionied/apollo/serialization/AssetName"
	"github.com/Salvionied/apollo/serialization/MultiAsset"
	"github.com/Salvionied/apollo/serialization/Policy"
	"github.com/Salvionied/apollo/serialization/TransactionInput"
	"github.com/Salvionied/apollo/serialization/TransactionOutput"
	"github.com/Salvionied/apollo/serialization/UTxO"
	"github.com/Salvionied/apollo/serialization/Value"
)

func InitUtxos(utxoCount int) []UTxO.UTxO {
	utxos := make([]UTxO.UTxO, 0)
	for i := range utxoCount {
		tx_in := TransactionInput.TransactionInput{
			TransactionId: make([]byte, 32),
			Index:         i,
		}

		Addr, _ := Address.DecodeAddress(TEST_WALLET_ADDRESS_1)
		policy := Policy.PolicyId{
			Value: "00000000000000000000000000000000000000000000000000000000",
		}
		asset_name := AssetName.NewAssetNameFromString(
			fmt.Sprintf("token%d", i),
		)
		Asset := Asset.Asset[int64]{
			asset_name: int64((i + 1) * 100)}
		assets := MultiAsset.MultiAsset[int64]{policy: Asset}
		value := Value.SimpleValue(int64((i+1)*1000000),
			assets)
		tx_out := TransactionOutput.SimpleTransactionOutput(
			Addr, value)
		utxos = append(utxos, UTxO.UTxO{Input: tx_in, Output: tx_out})
	}
	return utxos
}

func InitUtxosDifferentiated(utxoCount int) []UTxO.UTxO {
	utxos := make([]UTxO.UTxO, 0)
	for i := range utxoCount {
		tx_in := TransactionInput.TransactionInput{
			TransactionId: make([]byte, 32),
			Index:         i,
		}

		Addr, _ := Address.DecodeAddress(TEST_WALLET_ADDRESS_1)
		policy := Policy.PolicyId{
			Value: "00000000000000000000000000000000000000000000000000000000",
		}
		singleasset := Asset.Asset[int64]{}
		assetNames := make([]AssetName.AssetName, 0, i)
		for j := range i {
			asset_name := AssetName.NewAssetNameFromString(
				fmt.Sprintf("token%d", j),
			)
			assetNames = append(assetNames, asset_name)
		}
		// Sort asset names to ensure deterministic order
		sort.Slice(assetNames, func(a, b int) bool {
			return assetNames[a].String() < assetNames[b].String()
		})
		for _, name := range assetNames {
			singleasset[name] = int64((i + 1) * 100)
		}

		assets := MultiAsset.MultiAsset[int64]{policy: singleasset}
		value := Value.SimpleValue(int64((i+1)*1000000),
			assets)
		tx_out := TransactionOutput.SimpleTransactionOutput(
			Addr, value)
		utxos = append(utxos, UTxO.UTxO{Input: tx_in, Output: tx_out})
	}
	return utxos
}

func InitUtxosCongested(utxoCount int) []UTxO.UTxO {
	utxos := make([]UTxO.UTxO, 0)
	for i := range utxoCount {
		tx_in := TransactionInput.TransactionInput{
			TransactionId: make([]byte, 32),
			Index:         i,
		}

		Addr, _ := Address.DecodeAddress(TEST_WALLET_ADDRESS_1)
		policy := Policy.PolicyId{
			Value: fmt.Sprintf("0000000000000000000000000000000000000000000000000000000%d", i)[:56],
		}
		singleasset := Asset.Asset[int64]{}
		assetNames := make([]AssetName.AssetName, 0, i)
		for j := range i {
			asset_name := AssetName.NewAssetNameFromString(
				fmt.Sprintf("token%d", j),
			)
			assetNames = append(assetNames, asset_name)
		}
		// Sort asset names to ensure deterministic order
		sort.Slice(assetNames, func(a, b int) bool {
			return assetNames[a].String() < assetNames[b].String()
		})
		for _, name := range assetNames {
			singleasset[name] = int64((i + 1) * 100)
		}

		assets := MultiAsset.MultiAsset[int64]{policy: singleasset}
		value := Value.SimpleValue(int64(2000000),
			assets)
		tx_out := TransactionOutput.SimpleTransactionOutput(
			Addr, value)
		utxos = append(utxos, UTxO.UTxO{Input: tx_in, Output: tx_out})
	}
	return utxos
}
