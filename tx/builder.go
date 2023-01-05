package tx

import (
	"github.com/HighStakesSwitzerland/wallet_recovery_go/addr"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/lcdclient"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	signing2 "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"strconv"
)

func CreateSendTx(fromAddr []byte, toAddr []byte, coins types.Coins) []byte {

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
		panic(err)
	}

	encCfg := simapp.MakeTestEncodingConfig() // say to not use except for tests but then how to do?
	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithTxConfig(encCfg.TxConfig).
		WithChainID(config.ChainId).
		WithAccountNumber(acc).
		WithSequence(seq).
		WithSimulateAndExecute(false).
		WithTxConfig(encCfg.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	config.Logger.Info("Sequence: " + strconv.FormatUint(txFactory.Sequence(), 10))

	sendMsg := banktypes.NewMsgSend(fromAddr, toAddr, coins)
	txBuilder, err := txFactory.BuildUnsignedTx(sendMsg)
	if err != nil {
		panic(err)
	}

	// set fees
	txBuilder.SetFeeAmount(types.NewCoins(config.FeesAmount))
	txBuilder.SetGasLimit(config.GasLimit)
	txBuilder.SetMemo(config.Memo)

	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())

	// Sign

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: addr.PrivKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encCfg.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		panic(err)
	}

	// Second round: all signer infos are set, so each signer can sign.
	signerData := signing2.SignerData{
		ChainID:       config.ChainId,
		AccountNumber: acc,
		Sequence:      seq,
	}

	sigV2, err = tx.SignWithPrivKey(
		encCfg.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, addr.PrivKey, encCfg.TxConfig, 0)
	if err != nil {
		panic(err)
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		panic(err)
	}

	if err != nil {
		config.Logger.Error("Cannot set signature in TX", zap.Error(err))
	}

	txJSONBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}
	txJSON := string(txJSONBytes)

	config.Logger.Debug("TX to send" + txJSON)

	if err != nil {
		panic(err)
	}

	// re-get the tx (now signed) from the txBuilder
	txBytes, err = encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	return txBytes
}
