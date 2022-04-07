package main

import (
	"context"
	"fmt"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/client"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/key"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/msg"
	tx2 "github.com/HighStakesSwitzerland/wallet_recovery_go/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	logger      = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	mnemonic    = "turn reform life recycle tongue zero run alter trim vibrant note bulk cushion vapor awake barrel inflict pottery cup hurry link nephew chicken bubble"
	dest_wallet = "terra1uymwfafhq8fruvcjq8k67a29nqzrxnv9m4hg6d"
	lcd_client  = "https://lcd.terra.dev:443"
	fees_uluna  = int64(2000)
	fees_uusd   = int64(20000)
	sleep_time  = time.Millisecond * 20
	query_denom = "uusd"
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

	toAddr, err := msg.AccAddressFromBech32(dest_wallet)
	if err != nil {
		logger.Error("Error creating destination address", err.Error())
		return
	}

	startMonitoring(lcdClient, addr, toAddr)

}

func startMonitoring(lcdClient *client.LCDClient, addr msg.AccAddress, toAddr sdk.AccAddress) {

	for true {
		time.Sleep(sleep_time)

		balance, err := lcdClient.GetBalance(context.Background(), addr, query_denom)
		if err != nil {
			logger.Error("Cannot get balance", err.Error())
			continue
		}
		logger.Info("Balance is", balance)

		if balance.Amount == "0" {
			continue
		}

		amount, err := strconv.ParseInt(balance.Amount, 10, 64)
		if err != nil {
			logger.Error("Cannot convert balance", err.Error())
			continue
		}

		amountToMove := amount
		if balance.Denom == "uluna" {
			amountToMove -= fees_uluna
		}
		if balance.Denom == "uusd" {
			amountToMove -= fees_uusd
		}

		if amountToMove < 0 {
			continue
		}

		// TODO: bouger ca hors de la boucle ? mais quid sequence number
		account, err := lcdClient.LoadAccount(context.Background(), addr)
		if err != nil {
			logger.Error("Error loading address", err.Error())
			continue
		}

		// Create tx
		tx, err := lcdClient.CreateAndSignTx(
			context.Background(),
			client.CreateTxOptions{
				Msgs: []msg.Msg{
					msg.NewMsgSend(addr, toAddr, msg.NewCoins(msg.NewInt64Coin("uusd", amountToMove))), // 1UST
				},
				Memo:          "",
				AccountNumber: account.GetAccountNumber(),
				Sequence:      account.GetSequence(),
				SignMode:      tx2.SignModeDirect,
			})

		if err != nil {
			logger.Error("Error creating transaction again", err.Error())
			if strings.Contains(err.Error(), "sequence") {
				// retry with correct sequence number
				tx, err = lcdClient.CreateAndSignTx(
					context.Background(),
					client.CreateTxOptions{
						Msgs: []msg.Msg{
							msg.NewMsgSend(addr, toAddr, msg.NewCoins(msg.NewInt64Coin("uusd", amountToMove))), // 1UST
						},
						Memo:          "",
						AccountNumber: account.GetAccountNumber(),
						Sequence:      account.GetSequence() + 1,
						SignMode:      tx2.SignModeDirect,
					})

			}
			continue
		}

		// Broadcast
		res, err := lcdClient.Broadcast(context.Background(), tx)
		if err != nil {
			logger.Error("Error broadcasting tx", err)
			continue
		}
		logger.Info("Sucess:", res)
	}

}
