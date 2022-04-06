package main

import (
	"context"
	"fmt"
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
	mnemonic := "turn reform life recycle tongue zero run alter trim vibrant note bulk cushion vapor awake barrel inflict pottery cup hurry link nephew chicken bubble"
	privKeyBz, err := key.DerivePrivKeyBz(mnemonic, key.CreateHDPath(0, 0))
	if err != nil {
		logger.Error("Error creating priv key", err)
		return
	}
	privKey, err := key.PrivKeyGen(privKeyBz)
	if err != nil {
		logger.Error("Error creating priv key", err)
		return
	}
	addr := msg.AccAddress(privKey.PubKey().Address())
	logger.Info(fmt.Sprintf("address: [%s]", addr.String()))

	// Create LCDClient
	LCDClient := client.NewLCDClient(
		"https://lcd.terra.dev:443",
		"columbus-5",
		msg.NewDecCoinFromDec("uusd", msg.NewDecFromIntWithPrec(msg.NewInt(20), 2)), // 0.15uusd
		msg.NewDecFromIntWithPrec(msg.NewInt(15), 1),                                // gas
		privKey,
		time.Second, // tx timeout super short
	)

	balance, err := LCDClient.GetBalance(context.Background(), addr)
	if err != nil {
		logger.Error("Cannot get balance", err)
		return
	}
	logger.Info("Balance is", balance)

	// Create tx
	toAddr, err := msg.AccAddressFromBech32("terra1t849fxw7e8ney35mxemh4h3ayea4zf77dslwna")
	if err != nil {
		logger.Error("Error creating destination address", err)
		return
	}

	account, err := LCDClient.LoadAccount(context.Background(), addr)
	if err != nil {
		logger.Error("Error loading address", err)
		return
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
		return
	}

	// Broadcast
	res, err := LCDClient.Broadcast(context.Background(), tx)
	if err != nil {
		logger.Error("Error broadcasting tx", err)
		return
	}
	logger.Info("Sucess:", res)
}
