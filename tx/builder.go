package tx

import (
	"fmt"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func CreateSendTx(fromAddr []byte, toAddr []byte, coins types.Coins, kb keyring.Keyring) []byte {

	sendMsg := banktypes.NewMsgSend(fromAddr, toAddr, coins)

	// Create a new TxBuilder.
	encCfg := simapp.MakeTestEncodingConfig()
	txBuilder := encCfg.TxConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(sendMsg)
	if err != nil {
		panic(err)
	}

	txFactory := clienttx.Factory{}
	txFactory = txFactory.
		WithChainID("pulsar-1").
		WithMemo("my memo").
		WithKeybase(kb).
		WithTxConfig(encCfg.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	err = clienttx.Sign(txFactory, keyring.BackendMemory, txBuilder, false)
	if err != nil {
		panic(err)
	}

	txJSONBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}
	txJSON := string(txJSONBytes)
	fmt.Print("TX: ", txJSON)

	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	return txBytes
}
