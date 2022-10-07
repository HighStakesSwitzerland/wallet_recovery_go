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
	"github.com/tendermint/tendermint/rpc/client/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	logger      = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	mnemonic    = "turn reform life recycle tongue zero run alter trim vibrant note bulk cushion vapor awake barrel inflict pottery cup hurry link nephew chicken bubble"
	dest_wallet = "terra1p0v3n0t08r4mv6lup5lgthgjuvd58gvlehvxfs"
	lcd_client  = "https://columbus-lcd.terra.dev"
	rpc_client  = "http://0.0.0.0:36657"
	query_denom = "uluna"
	memo        = "yay \\o/"
)

func main() {
	logger.Info("Started")

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

	c, err := http.New(rpc_client, "/websocket")
	if err != nil {
		panic(err)
	}

	// call Start/Stop if you're subscribing to events
	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer c.Stop()

	query := "tm.event = 'NewBlockHeader'"
	out, err := c.Subscribe(context.Background(), "127.0.0.1", query)
	if err != nil {
		panic(err)
	}

	// Create LCDClient
	lcdClient := client.NewLCDClient(
		lcd_client,
		rpc_client,
		"columbus-5",
		msg.NewDecCoinFromDec(query_denom, msg.NewDecFromIntWithPrec(msg.NewInt(200), 2)), // 0.15uusd
		msg.NewDecFromIntWithPrec(msg.NewInt(30), 1),                                      // gas
		privKey,
		time.Second*1, // tx timeout super short
	)

	toAddr, err := msg.AccAddressFromBech32(dest_wallet)
	if err != nil {
		logger.Error("Error creating destination address", err.Error())
		return
	}

	for {
		select {
		case resultEvent := <-out:
			// We should have a switch here that performs a validation
			// depending on the event's type.
			logger.Info(resultEvent.Query)
			checkBalanceAndWithdraw(lcdClient, toAddr, addr)
		case <-c.Quit():
			logger.Info("Disconnected")
			return
		}
	}

}

func checkBalanceAndWithdraw(lcdClient *client.LCDClient, toAddr msg.AccAddress, addr msg.AccAddress) {
	balance, err := lcdClient.GetBalance(context.Background(), addr, query_denom)
	if err != nil {
		logger.Error("Cannot get balance", err.Error())
		return
	}

	if balance.Amount == "0" {
		logger.Error("balance = 0")
		return
	}

	amount, err := strconv.ParseInt(balance.Amount, 10, 64)
	if err != nil {
		logger.Error("Cannot convert balance", err.Error())
		return
	}

	logger.Info(fmt.Sprintf("Detected valid balance: %s moving %d", balance, amount))

	account, err := lcdClient.LoadAccount(context.Background(), addr)
	if err != nil {
		logger.Error("Error loading address", err.Error())
		return
	}

	// Create tx
	var tx *tx2.Builder
	tx, err = createTransaction(lcdClient, addr, toAddr, amount, account.GetAccountNumber(), account.GetSequence(), balance)

	if err != nil {
		var errorMsg = err.Error()
		logger.Error("Error creating transaction", errorMsg)
		if strings.Contains(errorMsg, "sequence") {
			i := strings.Index(errorMsg, "expected") + 9
			e := strings.Index(errorMsg, ", got")
			seqNumber, err := strconv.ParseUint(errorMsg[i:e], 10, 64)
			if err != nil {
				logger.Error(fmt.Sprintf("Could not parse sequence number %s, falbacking to %d", errorMsg[i:e], account.GetSequence()+1))
				seqNumber = account.GetSequence() + 1
			}

			logger.Info(fmt.Sprintf("Retrying with sequence %d", seqNumber))
			// amount - 1luna in case a failed tx consumed some fees
			tx, err = createTransaction(lcdClient, addr, toAddr, amount-1000000, account.GetAccountNumber(), seqNumber, balance)

			if err != nil {
				logger.Error("Error retrying transaction", err.Error())
			} else {
				logger.Info("Retry with new sequence Success:", tx)
			}
		}
		if strings.Contains(errorMsg, "insufficientfunds") {
			i := strings.Index(errorMsg, "messageindex:0:") + 9
			e := strings.Index(errorMsg, "ulunaissmallerthan")
			correctAmount, err := strconv.ParseUint(errorMsg[i:e], 10, 64)
			if err != nil {
				logger.Error(fmt.Sprintf("Could not parse new amount %s", errorMsg[i:e]))
			}
			logger.Info(fmt.Sprintf("Retrying correct amount %d", correctAmount))
			// amount - 1luna in case a failed tx consumed some fees
			tx, err = createTransaction(lcdClient, addr, toAddr, int64(correctAmount), account.GetAccountNumber(), account.GetSequence(), balance)

			if err != nil {
				logger.Error("Error retrying transaction", err.Error())
			} else {
				logger.Info("Retry with new amount Success:", tx)
			}
		}
	}

	if err == nil {
		// Broadcast
		res, err := lcdClient.Broadcast(context.Background(), tx)
		if err != nil {
			logger.Error("Error broadcasting tx", err)
			// panic(err) uncomment for stacktrace on exception
		}
		logger.Info("Success:", res)
		time.Sleep(2500) // wait a bit
	}

}

func createTransaction(lcdClient *client.LCDClient, addr msg.AccAddress, toAddr sdk.AccAddress, amountToMove int64,
	accountNumber uint64, seqNumber uint64, balance *client.QueryAccountBalance) (*tx2.Builder, error) {

	logger.Info(fmt.Sprintf("Creating TX with seq# %d and amount %d", seqNumber, amountToMove))
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
