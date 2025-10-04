package model

import (
	"github.com/Salvionied/apollo/serialization/Address"
)

type OrderTxInfo struct {
	EscrowContractAddress Address.Address
	EscrowContractRefUtxo EUTxO
	StateTokenPolicyId    string
	StateTokenRefUtxo     EUTxO
	MakerFee              int64
	TakerFee              int64
	DisputeFee            int64
	CancelFee             int64
	CollateralAmount      int64
	ChangeAddress         Address.Address
	UserUtxos             []EUTxO
	CollateralUtxo        EUTxO
	OrderUtxo             EUTxO
}

type OrderInfo struct {
	_                  struct{}        `plutusType:"DefList" plutusConstr:"1"`
	TradeTokenName     string          `plutusType:"StringBytes, omitempty"`
	TradeTokenPolicyId string          `plutusType:"HexString, omitempty"`
	OrderId            string          `plutusType:"StringBytes"`
	OrderAmount        int64           `plutusType:"Int"`
	MakerAddress       Address.Address `plutusType:"Address"`
	MakerRepAddress    MaybeAddress
	TakerAddress       Address.Address `plutusType:"Address"`
	TakerRepAddress    MaybeAddress
	MakerDeadline      int64 `plutusType:"Int"`
	TakerDeadline      int64 `plutusType:"Int"`
}

type BrokerageInfo struct {
	_              struct{} `plutusType:"DefList" plutusConstr:"1"`
	Precision      int64    `plutusType:"Int"`
	CollateralPct  int64    `plutusType:"Int"`
	MakerPct       int64    `plutusType:"Int"`
	TakerPct       int64    `plutusType:"Int"`
	CancelPct      int64    `plutusType:"Int"`
	MinCollateral  int64    `plutusType:"Int"`
	MakerMinFee    int64    `plutusType:"Int"`
	TakerMinFee    int64    `plutusType:"Int"`
	CancelMinFee   int64    `plutusType:"Int"`
	MinOrderAmount int64    `plutusType:"Int"`
	OrderThreshold int64    `plutusType:"Int"`
	CancelPenalty  int64    `plutusType:"Int"`
	AdaCollateral  int64    `plutusType:"Int, omitempty"`
}

type Order struct {
	_             struct{} `plutusType:"DefList" plutusConstr:"1"`
	OrderInfo     OrderInfo
	BrokerageInfo BrokerageInfo
	TradeState    string      `plutusType:"StringBytes"`
	OrderTxInfo   OrderTxInfo `plutusType:"Ignore"`
}
