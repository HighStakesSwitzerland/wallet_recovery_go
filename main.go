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
	mnemonic    = "barrel excite trap abandon banana file dress comic pepper exercise rural place frequent nation castle cool steak barely liquid lonely moment gather victory horse"
	dest_wallet = "terra1yym9g75nvkzyyxwcajljh8r788h8u90t8urp89"
	lcd_client  = "http://127.0.0.1:1317"
	rpc_client  = "http://127.0.0.1:26657"
	fees_uluna  = int64(2000)
	fees_uusd   = int64(20000)
	sleep_time  = time.Millisecond * 2000
	query_denom = "uluna"
	memo        = ""
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
		rpc_client,
		"columbus-5",
		msg.NewDecCoinFromDec(query_denom, msg.NewDecFromIntWithPrec(msg.NewInt(20), 2)), // 0.15uusd
		msg.NewDecFromIntWithPrec(msg.NewInt(15), 1),                                     // gas
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
	logger.Info("Started")
	for true {
		time.Sleep(sleep_time)

		balance, err := lcdClient.GetBalance(context.Background(), addr, query_denom)
		if err != nil {
			logger.Error("Cannot get balance", err.Error())
			continue
		}

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

		logger.Info(fmt.Sprintf("Detected valid balance: %s moving %d", balance, amountToMove))

		account, err := lcdClient.LoadAccount(context.Background(), addr)
		if err != nil {
			logger.Error("Error loading address", err.Error())
			continue
		}

		// Create tx
		var tx *tx2.Builder
		tx, err = createTransaction(lcdClient, addr, toAddr, amountToMove, account.GetAccountNumber(), account.GetSequence(), balance)

		if err != nil {
			var errorMsg = err.Error()
			logger.Error("Error creating transaction", err.Error())
			if strings.Contains(errorMsg, "sequence") {
				i := strings.Index(errorMsg, "expected") + 9
				e := strings.Index(errorMsg, ", got")
				seqNumber, err := strconv.ParseUint(errorMsg[i:e], 10, 64)
				if err != nil {
					logger.Error(fmt.Sprintf("Could not parse sequence number %s, falbacking to %d", errorMsg[i:e], account.GetSequence()+1))
					seqNumber = account.GetSequence() + 1
				}

				logger.Info(fmt.Sprintf("Retrying with sequence %d", seqNumber))
				// retry with correct sequence number
				tx, err = createTransaction(lcdClient, addr, toAddr, amountToMove, account.GetAccountNumber(), seqNumber, balance)

				if err != nil {
					logger.Error("Error creating transaction", err.Error())
				}
			}
		}

		if err == nil {
			// Broadcast
			res, err := lcdClient.Broadcast(context.Background(), tx)
			if err != nil {
				logger.Error("Error broadcasting tx", err)
				continue
			}
			logger.Info("Success:", res)
		}
	}

}

func createTransaction(lcdClient *client.LCDClient, addr msg.AccAddress, toAddr sdk.AccAddress, amountToMove int64,
	accountNumber uint64, seqNumber uint64, balance *client.QueryAccountBalance) (*tx2.Builder, error) {

	logger.Info(fmt.Sprintf("Creating TX with seq# %d", seqNumber))
	return lcdClient.CreateAndSignTx(
		context.Background(),
		client.CreateTxOptions{
			Denom:         query_denom,
			Addr:          addr,
			ToAddr:        toAddr,
			Amount:        amountToMove,
			Memo:          memo,
			AccountNumber: accountNumber,
			Sequence:      seqNumber,
			SignMode:      tx2.SignModeDirect,
			Balance:       balance,
		})
}
