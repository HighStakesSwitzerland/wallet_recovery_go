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
	logger      = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	mnemonic    = "turn reform life recycle tongue zero run alter trim vibrant note bulk cushion vapor awake barrel inflict pottery cup hurry link nephew chicken bubble"
	dest_wallet = "terra1uymwfafhq8fruvcjq8k67a29nqzrxnv9m4hg6d"
	lcd_client  = "https://lcd.terra.dev:443"
)

func main() {
	privKeyBz, err := key.DerivePrivKeyBz(mnemonic, key.CreateHDPath(0, 0))
	if err != nil {
		logger.Error("Error creating priv key", err.Error())
		return
	}
	privKey, err := key.PrivKeyGen(privKeyBz)
	if err != nil {
		logger.Error("Error creating priv key", err.Error())
		return
	}
	addr := msg.AccAddress(privKey.PubKey().Address())
	logger.Info(fmt.Sprintf("address: [%s]", addr.String()))

	// Create LCDClient
	lcdClient := client.NewLCDClient(
		lcd_client,
		"columbus-5",
		msg.NewDecCoinFromDec("uusd", msg.NewDecFromIntWithPrec(msg.NewInt(20), 2)), // 0.15uusd
		msg.NewDecFromIntWithPrec(msg.NewInt(15), 1),                                // gas
		privKey,
		time.Second, // tx timeout super short
	)

	startMonitoring(lcdClient, addr)

}

func startMonitoring(lcdClient *client.LCDClient, addr msg.AccAddress) {
	balance, err := lcdClient.GetBalance(context.Background(), addr)
	if err != nil {
		logger.Error("Cannot get balance", err.Error())
		return
	}
	logger.Info("Balance is", balance)

	// Create tx
	toAddr, err := msg.AccAddressFromBech32(dest_wallet)
	if err != nil {
		logger.Error("Error creating destination address", err.Error())
		return
	}

	account, err := lcdClient.LoadAccount(context.Background(), addr)
	if err != nil {
		logger.Error("Error loading address", err.Error())
		return
	}

	tx, err := lcdClient.CreateAndSignTx(
		context.Background(),
		client.CreateTxOptions{
			Msgs: []msg.Msg{
				msg.NewMsgSend(addr, toAddr, msg.NewCoins(msg.NewInt64Coin("uusd", 1000000))), // 1UST
			},
			Memo:          "",
			AccountNumber: account.GetAccountNumber(),
			Sequence:      account.GetSequence(),
			SignMode:      tx2.SignModeDirect,
		})

	if err != nil {
		logger.Error("Error creating transaction", err.Error())
		return
	}

	// Broadcast
	res, err := lcdClient.Broadcast(context.Background(), tx)
	if err != nil {
		logger.Error("Error broadcasting tx", err)
		return
	}
	logger.Info("Sucess:", res)
}
