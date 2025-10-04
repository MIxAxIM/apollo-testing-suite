package main

import (
	"encoding/hex"
	"flag"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"plutus-v3-sc-tx-builder/config"
	"plutus-v3-sc-tx-builder/model"
	plutusEncoder "plutus-v3-sc-tx-builder/plutus-encoder"
	"plutus-v3-sc-tx-builder/proxy"
	"plutus-v3-sc-tx-builder/utility"
	"plutus-v3-sc-tx-builder/wallet"
	"strconv"
	"time"

	"github.com/lmittmann/tint"

	"github.com/Salvionied/apollo"
	"github.com/Salvionied/apollo/serialization"
	"github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/txBuilding/Backend/BlockFrostChainContext"
	"github.com/Salvionied/apollo/txBuilding/Utils"
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      config.LOG_LEVEL,
			TimeFormat: time.StampMilli,
		}),
	))

	var proxyUrl string
	flag.StringVar(&proxyUrl, "proxy", "", "Proxy URL to route all outgoing connections through (e.g., http://localhost:8080)")
	flag.Parse()

	if proxyUrl != "" {
		slog.Info("Proxy URL provided, initializing proxy transport.", "proxy_url", proxyUrl)
		transport, err := proxy.GetProxyTransport(proxyUrl)
		if err != nil {
			slog.Error("Failed to initialize proxy transport", "error", err)
			os.Exit(1)
		}
		http.DefaultClient.Transport = transport

		slog.Info("Proxy transport initialized and set for http.DefaultClient and environment variables.")
	} else {
		slog.Info("No proxy URL provided. Application will run without a proxy.")
		os.Unsetenv("HTTP_PROXY")
		os.Unsetenv("HTTPS_PROXY")
	}
}

func main() {
	slog.Info("Starting Plutud V3 smart contract transaction building script.")
	slog.Debug("Initializing wallet setup.")
	wallet.WalletSetup()
	slog.Info("Wallets initialized successfully.")

	slog.Info("Attempting to initialize BlockFrostChainContext.")

	bfcApiKey := os.Getenv("BFC_API_KEY")
	if bfcApiKey == "" {
		slog.Error("BFC_API_KEY environment variable is not set or empty. Please set it to proceed.", "action", "set environment variable")
		os.Exit(1)
	}
	slog.Debug("BFC_API_KEY successfully retrieved.", "key_present", true)

	be, err := BlockFrostChainContext.NewBlockfrostChainContext(config.BFC_API_URL, config.BFC_NETWORK_ID, bfcApiKey)
	if err != nil {
		slog.Error("Failed to create BlockFrostChainContext. Ensure network ID and API key are correct.", "error", err)
		os.Exit(1)
	}
	slog.Info("BlockFrostChainContext initialized successfully.", "network_id", config.BFC_NETWORK_ID)

	slog.Info("Proceeding with successful order creation scenario.")

	apolloBE := apollo.New(&be)
	slog.Debug("Apollo backend instance created.")

	slog.Debug("Decoding escrow contract address.", "address", config.ESCROW_ADDRESS)
	escrowContractAddress, err := Address.DecodeAddress(config.ESCROW_ADDRESS)
	if err != nil {
		slog.Error("Failed to decode escrow contract address. Check the provided address format.", "error", err, "address", config.ESCROW_ADDRESS)
		os.Exit(1)
	}
	slog.Info("Escrow contract address decoded.", "decoded_address", escrowContractAddress.String())

	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	slog.Debug("Current time in milliseconds.", "timestamp", currentTime)

	md := currentTime + (config.MakerDeadline * 1000)
	td := currentTime + (config.TakerDeadline * 1000)
	slog.Debug("Calculated maker and taker deadlines.", "maker_deadline", md, "taker_deadline", td)

	makerFee := utility.CalculateFee(config.Precision, config.OrderAmount, config.OrderThreshold, config.MakerPct, config.MakerMinFee)
	collateralAmount := utility.CalculateFee(config.Precision, config.OrderAmount, config.OrderThreshold, config.CollateralPct, config.MinCollateral)
	slog.Debug("Calculated maker fee and collateral amount.", "maker_fee", makerFee, "collateral_amount", collateralAmount)

	slog.Info("Constructing mock order object.")
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randomNum := r.Intn(1000000)
	orderIdWithRandom := config.OrderId + strconv.Itoa(randomNum)

	mockOrder := &model.Order{
		OrderInfo: model.OrderInfo{
			OrderId:         orderIdWithRandom,
			OrderAmount:     config.OrderAmount,
			MakerAddress:    wallet.GetUserNum1Wallet().Address,
			MakerRepAddress: model.Nothing{},
			TakerAddress:    wallet.GetUserNum2Wallet().Address,
			TakerRepAddress: model.Nothing{},
			MakerDeadline:   md,
			TakerDeadline:   td,
		},
		BrokerageInfo: model.BrokerageInfo{
			Precision:      config.Precision,
			CollateralPct:  config.CollateralPct,
			MakerPct:       config.TakerPct,
			CancelPct:      config.CancelPct,
			MinCollateral:  config.MinCollateral,
			MakerMinFee:    config.MakerMinFee,
			TakerMinFee:    config.TakerMinFee,
			CancelMinFee:   config.CancelMinFee,
			MinOrderAmount: config.MinOrderAmount,
			OrderThreshold: config.OrderThreshold,
			CancelPenalty:  config.CancelPenalty,
		},
		TradeState: config.UNCOMMITTED_ORDER_STATUS,

		OrderTxInfo: model.OrderTxInfo{
			EscrowContractAddress: escrowContractAddress,
			EscrowContractRefUtxo: model.EUTxO{
				TxID:      config.ESCROW_SCRIPT_REF_UTXO_TXID,
				TxIDIndex: config.ESCROW_SCRIPT_REF_UTXO_TXID_INDEX,
			},
			StateTokenPolicyId: config.APBST_POLICY_ID,
			StateTokenRefUtxo: model.EUTxO{
				TxID:      config.APBST_SCRIPT_REF_UTXO_TXID,
				TxIDIndex: config.APBST_SCRIPT_REF_UTXO_TXID_INDEX,
			},
			MakerFee:         makerFee,
			CollateralAmount: collateralAmount,
			ChangeAddress:    wallet.GetUserNum1Wallet().Address,
			UserUtxos:        []model.EUTxO{},
			CollateralUtxo:   model.EUTxO{},
		},
	}
	slog.Debug("Mock order object created.", "order_id", mockOrder.OrderInfo.OrderId, "maker_address", mockOrder.OrderInfo.MakerAddress.String())

	slog.Debug("Setting Apollo wallet from Maker Address.", "maker_address", mockOrder.OrderInfo.MakerAddress.String())
	apolloBE = apolloBE.SetWalletFromBech32(mockOrder.OrderInfo.MakerAddress.String())
	slog.Info("Apollo wallet set.")

	makerCommittingAmount := mockOrder.OrderInfo.OrderAmount + mockOrder.OrderTxInfo.MakerFee + mockOrder.OrderTxInfo.CollateralAmount
	slog.Debug("Calculated maker committing amount.", "amount", makerCommittingAmount)

	slog.Debug("Marshaling Plutus order datum.")
	orderDatumMarshaled, err := plutusEncoder.MarshalPlutus(*mockOrder)
	if err != nil {
		slog.Error("Failed to marshal Plutus order datum.", "error", err)
		os.Exit(1)
	}
	slog.Info("Plutus order datum marshaled successfully.")

	slog.Debug("Fetching user UTXOs for maker address.", "maker_address", mockOrder.OrderInfo.MakerAddress.String())
	userUtxos, err := be.Utxos(mockOrder.OrderInfo.MakerAddress)
	if err != nil {
		slog.Error("Failed to fetch user UTXOs. Ensure the address has available UTXOs.", "error", err, "address", mockOrder.OrderInfo.MakerAddress.String())
		os.Exit(1)
	}
	slog.Info("User UTXOs fetched successfully.", "num_utxos", len(userUtxos))

	slog.Debug("Fetching last block slot.")
	lastSlot, err := be.LastBlockSlot()
	if err != nil {
		slog.Error("Failed to fetch last block slot. Check chain context connectivity.", "error", err)
		os.Exit(1)
	}
	slog.Info("Last block slot fetched.", "slot", lastSlot)

	slog.Info("Building transaction with Apollo backend.")
	apolloBE, err = apolloBE.
		SetChangeAddress(mockOrder.OrderTxInfo.ChangeAddress).
		AddLoadedUTxOs(userUtxos...).
		MintAssetsWithRedeemer(
			apollo.Unit{
				PolicyId: mockOrder.OrderTxInfo.StateTokenPolicyId,
				Name:     mockOrder.OrderInfo.OrderId,
				Quantity: int(1),
			},
			*config.IndexOneMintRedeemer,
		).
		AddReferenceInputV3(
			mockOrder.OrderTxInfo.StateTokenRefUtxo.TxID,
			mockOrder.OrderTxInfo.StateTokenRefUtxo.TxIDIndex,
		).
		PayToContract(
			mockOrder.OrderTxInfo.EscrowContractAddress,
			orderDatumMarshaled,
			int(makerCommittingAmount),
			true,
			apollo.Unit{
				PolicyId: mockOrder.OrderTxInfo.StateTokenPolicyId,
				Name:     mockOrder.OrderInfo.OrderId,
				Quantity: int(1),
			},
		).
		AddRequiredSigner(wallet.GetSCAdminWallet().PKH).
		AddRequiredSigner(serialization.PubKeyHash(mockOrder.OrderInfo.MakerAddress.PaymentPart)).
		SetTtl(int64(lastSlot) + 300).
		Complete()

	if err != nil {
		slog.Error("Failed to complete transaction building. Review transaction parameters and UTXOs.", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction built successfully.")
	slog.Debug("Transaction details after building.", "transaction_object", apolloBE)

	slog.Info("Initiating transaction signing process.")
	slog.Debug("Signing with UserNum1Wallet.")
	apolloBE, err = apolloBE.SignWithSkey(wallet.GetUserNum1Wallet().Vkey, wallet.GetUserNum1Wallet().Skey)
	if err != nil {
		slog.Error("Failed to sign transaction with UserNum1Wallet. Check private key and transaction validity.", "error", err)
		os.Exit(1)
	}
	slog.Debug("Signing with SCAdminWallet.")
	apolloBE, err = apolloBE.SignWithSkey(wallet.GetSCAdminWallet().Vkey, wallet.GetSCAdminWallet().Skey)
	if err != nil {
		slog.Error("Failed to sign transaction with SCAdminWallet. Check private key and transaction validity.", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction signed successfully by all required parties.")

	tx := apolloBE.GetTx()
	slog.Debug("Transaction object retrieved from Apollo backend.")

	slog.Debug("Converting transaction to bytes.")
	txByte, err := tx.Bytes()
	if err != nil {
		slog.Error("Failed to convert transaction to bytes.", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction converted to bytes successfully.")

	slog.Info("Evaluating transaction with Blockfrost backend.")
	evalTx, err := be.EvaluateTx(txByte)
	if err != nil {
		slog.Error("Failed to evaluate transaction with Blockfrost. Check transaction structure and network connectivity.", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction evaluation complete.", "evaluation_result", evalTx)

	if len(evalTx) == 0 {
		slog.Error("Transaction evaluation failed: received empty evaluation result. This indicates a potential issue with the transaction or script.", "evaluation_result", evalTx)
		os.Exit(1)
	}
	slog.Debug("Transaction evaluation returned non-empty result.")

	slog.Info("Converting transaction to CBOR format.")
	cbor, err := Utils.ToCbor(tx)
	if err != nil {
		slog.Error("Failed to convert transaction to CBOR. Check transaction object integrity.", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction successfully converted to CBOR.")
	slog.Info("Transaction CBOR representation.", "cbor_hex", cbor)

	slog.Info("Submitting transaction to the blockchain.")
	txHash, err := be.SubmitTx(*tx)
	if err != nil {
		slog.Error("Failed to submit transaction. Review network status and transaction validity.", "error", err)
		os.Exit(1)
	}
	slog.Info("Transaction submitted successfully to the blockchain.")

	txID := hex.EncodeToString(txHash.Payload)
	slog.Info("Transaction ID generated.", "tx_id", txID)
	slog.Info("Transaction can be monitored on the blockchain explorer.", "explorer_link", "https://preprod.cexplorer.io/tx/"+txID)
	slog.Info("Plutus V3 smart contract transaction building script finished execution.")
}
