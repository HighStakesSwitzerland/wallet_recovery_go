package tx

import (
	"github.com/HighStakesSwitzerland/wallet_recovery_go/addr"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/lcdclient"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"strconv"
)

func CreateSendTx(fromAddr []byte, toAddr []byte, coins types.Coins) []byte {

	sendMsg := banktypes.NewMsgSend(fromAddr, toAddr, coins)

	// Create a new TxBuilder.
	encCfg := simapp.MakeTestEncodingConfig()
	txBuilder := encCfg.TxConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(sendMsg)
	if err != nil {
		panic(err)
	}

	// get account sequence and number
	account, err := lcdclient.LoadAccount(addr.Bech32wallet)
	if err != nil {
		return nil
	}
	seq, err := strconv.ParseUint(account.Account.Sequence, 10, 64)
	if err != nil {
		return nil
	}
	acc, err := strconv.ParseUint(account.Account.AccountNumber, 10, 64)
	if err != nil {
		return nil
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(config.ChainId).
		WithSequence(seq).
		WithAccountNumber(acc).
		WithSimulateAndExecute(true).
		WithKeybase(addr.Keyring).
		WithTxConfig(encCfg.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// set fees
	txBuilder.SetFeeAmount(types.NewCoins(config.FeesAmount))
	txBuilder.SetGasLimit(config.GasLimit)
	txBuilder.SetMemo(config.Memo)

	err = tx.Sign(txFactory, keyring.BackendMemory, txBuilder, false)
	if err != nil {
		panic(err)
	}

	txJSONBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}
	txJSON := string(txJSONBytes)

	config.Logger.Debug("TX to send" + txJSON)

	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	return txBytes
}
