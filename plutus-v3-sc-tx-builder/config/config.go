package config

import (
	"log/slog"
	"plutus-v3-sc-tx-builder/utility"

	"github.com/Salvionied/apollo/serialization/PlutusData"
	"github.com/Salvionied/apollo/serialization/Redeemer"
)

const (
	LOG_LEVEL                                = slog.LevelInfo
	NETWORK                           string = "preprod"
	INDEX_ONE                         uint64 = 122
	UNCOMMITTED_ORDER_STATUS          string = "UNCOMMITTED_ORDER"
	ESCROW_ADDRESS                    string = "addr_test1xrgysjt2g0t4l7h54px984wehx9uhtautarly2nqq83w5622m3nsefutuywd8uqzd2n4u7w0vmymjhmmuuc46hdxy85qzlh2dc"
	ESCROW_SCRIPT_REF_UTXO_TXID       string = "87ce9218b91f09156b15bdb6fa5c3a8aa78d846184f2e1b5375df0114abd8b6d"
	ESCROW_SCRIPT_REF_UTXO_TXID_INDEX int    = 0
	APBST_POLICY_ID                   string = "aefdb5f954ea897ec536de425e62632728fa78b4638882654d0f4075"
	APBST_SCRIPT_REF_UTXO_TXID        string = "fbf2cc43233490bc7f748957711bda85a530484113c8b65381b0613c30c13ac2"
	APBST_SCRIPT_REF_UTXO_TXID_INDEX  int    = 0
	BFC_NETWORK_ID                    int    = 0
	BFC_API_URL                       string = "https://cardano-preprod.blockfrost.io/api"
)

var (
	OrderId          string = "ID_A2A_RR"
	toLovelace       int64  = 1000000
	OrderAmount      int64  = 25 * toLovelace
	Precision        int64  = 10
	CollateralPct    int64  = utility.ToFraction(10)
	MakerPct         int64  = utility.ToFraction(0.25)
	TakerPct         int64  = utility.ToFraction(0.75)
	CancelPct        int64  = utility.ToFraction(1)
	MinCollateral    int64  = 25 * toLovelace
	MakerMinFee      int64  = 1250000
	TakerMinFee      int64  = 3750000
	CancelMinFee     int64  = 3 * toLovelace
	MinOrderAmount   int64  = 10 * toLovelace
	OrderThreshold   int64  = 500 * toLovelace
	CancelPenalty    int64  = 2 * CancelMinFee
	AdaCollateral    int64  = 3 * toLovelace
	ExtendingPenalty int64  = 5 * toLovelace
	DisputePct       int64  = utility.ToFraction(2)
	DisputeMinFee    int64  = 5 * toLovelace
	DisputePenalty   int64  = 2 * DisputeMinFee

	MakerDeadline int64 = 7200
	TakerDeadline int64 = 3600

	IndexOneMintRedeemer *Redeemer.Redeemer = &Redeemer.Redeemer{
		Tag:   Redeemer.MINT,
		Index: 0,
		Data: PlutusData.PlutusData{
			PlutusDataType: PlutusData.PlutusArray,
			TagNr:          INDEX_ONE,
			Value:          PlutusData.PlutusDefArray{},
		},
	}
)
