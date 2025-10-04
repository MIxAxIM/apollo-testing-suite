package wallet

import (
	"encoding/hex"
	"log/slog"
	"os"
	"plutus-v3-sc-tx-builder/config"

	"github.com/Salvionied/apollo/serialization"
	"github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/serialization/Key"
	"github.com/blinklabs-io/bursa"
)

type Wallet struct {
	Address Address.Address
	PKH     serialization.PubKeyHash
	Vkey    Key.VerificationKey
	Skey    Key.SigningKey
}

var (
	userNum1Wallet = &Wallet{}
	userNum2Wallet = &Wallet{}
	scAdminWallet  = &Wallet{}
)

func WalletSetup() {
	slog.Info("Setting up wallets...")
	userNum1Mnemonic := os.Getenv("USER_NUM_1_MNEMONIC")
	if userNum1Mnemonic == "" {
		slog.Error("USER_NUM_1_MNEMONIC environment variable not set or empty. Exiting.")
		os.Exit(1)
	}
	userNum1Wallet = SetWallet(userNum1Mnemonic)
	slog.Debug("User 1 wallet set up.", "address", userNum1Wallet.Address.String())

	userNum2Mnemonic := os.Getenv("USER_NUM_2_MNEMONIC")
	if userNum2Mnemonic == "" {
		slog.Error("USER_NUM_2_MNEMONIC environment variable not set or empty. Exiting.")
		os.Exit(1)
	}
	userNum2Wallet = SetWallet(userNum2Mnemonic)
	slog.Debug("User 2 wallet set up.", "address", userNum2Wallet.Address.String())

	scAdminMnemonic := os.Getenv("SC_ADMIN_MNEMONIC")
	if scAdminMnemonic == "" {
		slog.Error("SC_ADMIN_MNEMONIC environment variable not set or empty. Exiting.")
		os.Exit(1)
	}
	scAdminWallet = SetWallet(scAdminMnemonic)
	slog.Debug("SC Admin wallet set up.", "address", scAdminWallet.Address.String())
	slog.Info("Wallets setup complete.")
}

func SetWallet(mnemonic string) *Wallet {
	slog.Debug("Setting up individual wallet.")
	rootKey, err := bursa.GetRootKeyFromMnemonic(mnemonic, "")
	if err != nil {
		slog.Error("Error getting root key from mnemonic", "error", err)
		panic(err)
	}
	slog.Debug("Root key derived.")
	accountKey := bursa.GetAccountKey(rootKey, 0)
	paymentKey := bursa.GetPaymentKey(accountKey, 0)
	value, err := bursa.GetAddress(accountKey, config.NETWORK, 0)
	if err != nil {
		slog.Error("Error getting value", "error", err)
		panic(err)
	}
	slog.Debug("Value derived.")
	address, err := Address.DecodeAddress(value.String())
	if err != nil {
		slog.Error("Error decoding address", "error", err)
		panic(err)
	}
	slog.Debug("Address decoded.", "address", address.String())
	vKeyBytes, err := hex.DecodeString(bursa.GetPaymentVKey(paymentKey).CborHex)
	if err != nil {
		slog.Error("Error decoding verification key", "error", err)
		panic(err)

	}
	slog.Debug("Verification key decoded.")
	sKeyBytes, err := hex.DecodeString(bursa.GetPaymentSKey(paymentKey).CborHex)
	if err != nil {
		slog.Error("Error decoding signing key", "error", err)
		panic(err)
	}
	slog.Debug("Signing key decoded.")
	vKeyBytes = vKeyBytes[2:]
	sKeyBytes = sKeyBytes[2:]
	slog.Debug("Keys processed and truncated.")

	return &Wallet{
		Address: address,
		PKH:     serialization.PubKeyHash(address.PaymentPart),
		Vkey:    Key.VerificationKey{Payload: vKeyBytes},
		Skey:    Key.SigningKey{Payload: sKeyBytes},
	}
}

func GetUserNum1Wallet() *Wallet {
	return userNum1Wallet
}

func GetUserNum2Wallet() *Wallet {
	return userNum2Wallet
}

func GetSCAdminWallet() *Wallet {
	return scAdminWallet
}
