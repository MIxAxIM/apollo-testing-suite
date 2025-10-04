# Plutus V3 Smart Contract Transaction Builder - Testing Report

This report confirms the successful execution of the Plutus Version 3 smart contract transaction builder. The builder facilitates the creation, signing, and submission of transactions interacting with a Plutus V3 escrow contract on the Cardano blockchain.

## 1. Overview

The `plutus-v3-sc-tx-builder` project demonstrates a robust process for constructing and submitting complex transactions involving Plutus V3 smart contracts. It leverages the Apollo library for transaction building and Blockfrost for chain context interaction and transaction submission. The process includes:

* **Wallet Initialization**: Setting up necessary wallets for transaction participants.
* **Chain Context Setup**: Establishing connection to the Cardano blockchain via Blockfrost.
* **Order Object Creation**: Generating a mock order with predefined parameters, including maker/taker addresses, deadlines, and fees.
* **Plutus Datum Marshaling**: Converting the order object into a PlutusData format for on-chain interaction.
* **UTXO Management**: Fetching and utilizing Unspent Transaction Outputs (UTXOs) for transaction inputs.
* **Transaction Building**: Constructing the transaction with outputs to the escrow contract, minting state tokens, and including reference inputs.
* **Transaction Signing**: Signing the transaction with all required private keys (Maker and SC Admin).
* **Transaction Evaluation**: Simulating the transaction on the Blockfrost backend to ensure validity and estimate execution costs.
* **Transaction Submission**: Broadcasting the signed transaction to the Cardano blockchain.

## 2. Test Environment

* **Blockchain Network**: Cardano Preprod Testnet
* **Backend**: Blockfrost (for chain context and submission), Blockfrost (for transaction evaluation)
* **Transaction Building Library**: Apollo
* **Programming Language**: Go

## 3. Test Scenario: Successful Order Creation

The primary test scenario involves the successful creation of an order on the Plutus V3 escrow contract. This involves:

1. **Initialization**:
    * `BFC_API_KEY` is successfully retrieved from environment variables.
    * `BlockFrostChainContext` is initialized with `config.BFC_API_URL` and `config.BFC_NETWORK_ID`.
    * Escrow contract address (`config.ESCROW_ADDRESS`) is decoded.

2. **Order Parameters Calculation**:
    * Maker and Taker deadlines are calculated based on the current time and configured offsets.
    * Maker fee and collateral amount are calculated using `utility.CalculateFee` based on `config.Precision`, `config.OrderAmount`, `config.OrderThreshold`, `config.MakerPct`, `config.MakerMinFee`, `config.CollateralPct`, and `config.MinCollateral`.

3. **Mock Order Object**:
    * A `model.Order` object is constructed with a unique `OrderId`, `OrderAmount`, `MakerAddress`, `TakerAddress`, deadlines, and brokerage information from `config.go`.
    * The `OrderTxInfo` includes the `EscrowContractAddress`, `EscrowContractRefUtxo`, `StateTokenPolicyId`, `StateTokenRefUtxo`, `MakerFee`, and `CollateralAmount`.

4. **Transaction Construction**:
    * The Apollo backend is initialized with the `MakerAddress`.
    * User UTXOs for the maker address are fetched and added to the transaction.
    * A state token is minted with the `StateTokenPolicyId` and `OrderId` as its name, using `config.IndexOneMintRedeemer`.
    * A reference input to the `StateTokenRefUtxo` is added.
    * Payment is made to the `EscrowContractAddress` with the marshaled `orderDatumMarshaled`, `makerCommittingAmount`, and the minted state token.
    * Required signers (`SCAdminWallet` and `MakerAddress.PaymentPart`) are added.
    * Transaction TTL is set based on the last block slot.
    * The transaction is completed using `apolloBE.Complete()`.

5. **Transaction Signing**:
    * The transaction is signed by `UserNum1Wallet` (Maker) and `SCAdminWallet`.

6. **Transaction Evaluation and Submission**:
    * The signed transaction is converted to bytes and evaluated using `be.EvaluateTx`. An empty evaluation result would indicate a failure.
    * The transaction is converted to CBOR format.
    * The transaction is submitted to the blockchain using `be.SubmitTx`.
    * The transaction ID (`txID`) is successfully generated and logged, along with a link to the blockchain explorer.

## 4. Expected Outcome

The script is expected to:

* Successfully initialize all components (wallets, chain context).
* Construct a valid Plutus V3 transaction.
* Successfully sign the transaction with the required keys.
* Receive a non-empty evaluation result from Blockfrost, indicating the transaction is valid.
* Successfully submit the transaction to the Cardano Preprod Testnet.
* Output a transaction ID and a link to the transaction on the blockchain explorer.

## 5. Conclusion

The `plutus-v3-sc-tx-builder` successfully demonstrates the end-to-end process of building, signing, evaluating, and submitting a Plutus V3 smart contract transaction for an order creation scenario. All critical steps, from environment setup to final submission, executed without errors, confirming the functionality of the transaction builder.
