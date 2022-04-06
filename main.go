package main

import (
	"context"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/client"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/key"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/msg"
	tx2 "github.com/HighStakesSwitzerland/wallet_recovery_go/tx"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"time"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "config")
)

func main() {
	mnemonic := "essence gallery exit illegal nasty luxury sport trouble measure benefit busy almost bulb fat shed today produce glide meadow require impact fruit omit weasel"
	privKeyBz, err := key.DerivePrivKeyBz(mnemonic, key.CreateHDPath(0, 0))
	if err != nil {
		logger.Error("Error creating priv key", err)
	}
	privKey, err := key.PrivKeyGen(privKeyBz)
	if err != nil {
		logger.Error("Error creating priv key", err)
	}

	// Create LCDClient
	LCDClient := client.NewLCDClient(
		"http://127.0.0.1:1317",
		"columbus-5",
		msg.NewDecCoinFromDec("uusd", msg.NewDecFromIntWithPrec(msg.NewInt(20), 2)), // 0.15uusd
		msg.NewDecFromIntWithPrec(msg.NewInt(15), 1),                                // gas
		privKey,
		time.Millisecond, // tx timeout super short
	)

	// Create tx
	addr := msg.AccAddress(privKey.PubKey().Address())
	toAddr, err := msg.AccAddressFromBech32("terra1t849fxw7e8ney35mxemh4h3ayea4zf77dslwna")
	if err != nil {
		logger.Error("Error creating destination address", err)
	}

	account, err := LCDClient.LoadAccount(context.Background(), addr)
	if err != nil {
		logger.Error("Error loading address", err)
	}

	tx, err := LCDClient.CreateAndSignTx(
		context.Background(),
		client.CreateTxOptions{
			Msgs: []msg.Msg{
				msg.NewMsgSend(addr, toAddr, msg.NewCoins(msg.NewInt64Coin("uusd", 100000000))), // 100UST
			},
			Memo:          "",
			AccountNumber: account.GetAccountNumber(),
			Sequence:      account.GetSequence(),
			SignMode:      tx2.SignModeDirect,
		})

	if err != nil {
		logger.Error("Error creating transaction", err)
	}

	// Broadcast
	res, err := LCDClient.Broadcast(context.Background(), tx)
	if err != nil {
		logger.Error("Error broadcasting tx", err)
	}
	logger.Info("Sucess:", res)
}
