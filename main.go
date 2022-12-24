package main

import (
	"fmt"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/addr"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/tx"
	"github.com/cosmos/cosmos-sdk/types"
)

var (
	mnemonic        = "grant rice replace explain federal release fix clever romance raise often wild taxi quarter soccer fiber love must tape steak together observe swap guitar"
	dest_wallet     = "secret16mu3ttz3u3dj5fppvms86vm0jv59rllyza8pmq"
	lcd_client      = "https://lcd.testnet.secretsaturn.net"
	rpc_client      = "https://rpc.pulsar.scrttestnet.com:443"
	grpc_client     = "grpcbin.pulsar.scrttestnet.com:9099"
	bech32_prefix   = "secret"
	hdPath          = "m/44'/529'/0'/0/0" // cf cosmos.directory (118 = cosmos, 330 = terra, 529 = secret...)
	coins_denom     = "uscrt"
	amount_to_snipe = 1_000_000 // amount for the snipe tx (only once on next block)
	amount_to_spam  = 250_000   // amount for the spam txs (many many tx sent on next block)
	fees_amount     = 1_000
)

func main() {
	// set prefix globally
	types.GetConfig().SetBech32PrefixForAccount(bech32_prefix, bech32_prefix)

	fromAddr, toAddr, kb := addr.GenerateAddresses(mnemonic, hdPath, dest_wallet)

	fmt.Println("Started")

	grpcConn := tx.SetupGrpc(grpc_client)
	defer grpcConn.Close() // close connection on program exit

	coins := types.NewCoins(types.NewInt64Coin("scrt", 12))

	txBytes := tx.CreateSendTx(fromAddr, toAddr, coins, kb)

	tx.SendTx(txBytes)

	/*
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
		out, err := c.Subscribe(  .Background(), "127.0.0.1", query)
		if err != nil {
			panic(err)
		}

			for {
				select {
				case resultEvent := <-out:
					// We should have a switch here that performs a validation
					// depending on the event's type.
					logger.Info(resultEvent.Query)
					//checkBalanceAndWithdraw(lcdClient, toAddr, addr)
				case <-c.Quit():
					logger.Info("Disconnected")
					return
				}
			}
	*/

}

/*
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
*/
