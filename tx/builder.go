package tx

import (
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func CreateSendTx(fromAddr []byte, toAddr []byte, coins types.Coins, fees types.Coins, kb keyring.Keyring) []byte {

	sendMsg := banktypes.NewMsgSend(fromAddr, toAddr, coins)

	// Create a new TxBuilder.
	encCfg := simapp.MakeTestEncodingConfig()
	txBuilder := encCfg.TxConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(sendMsg)
	if err != nil {
		panic(err)
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(config.Chain_id).
		WithSimulateAndExecute(true).
		WithMemo(config.Memo).
		WithKeybase(kb).
		WithTxConfig(encCfg.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).WithAccountNumber(1470138)

	// set fees
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetGasLimit(20000)

	err = tx.Sign(txFactory, keyring.BackendMemory, txBuilder, false)
	if err != nil {
		panic(err)
	}

	txJSONBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}
	txJSON := string(txJSONBytes)

	config.Logger.Info("TX to send" + txJSON)

	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	return txBytes
}
