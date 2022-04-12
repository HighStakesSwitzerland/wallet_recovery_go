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
	logger       = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	mnemonic     = "barrel excite trap abandon banana file dress comic pepper exercise rural place frequent nation castle cool steak barely liquid lonely moment gather victory horse"
	dest_wallet  = "terra1yym9g75nvkzyyxwcajljh8r788h8u90t8urp89"
	lcd_client   = "http://127.0.0.1:1317"
	rpc_client   = "http://127.0.0.1:26657"
	fees         = msg.NewDecFromIntWithPrec(msg.NewInt(20), 2)
	sleep_time   = time.Millisecond * 1
	query_denom  = "uluna"
	memo         = "go_aws"
	amountToMove = int64(990000)
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
		msg.NewDecCoinFromDec(query_denom, fees),
		msg.NewDecFromIntWithPrec(msg.NewInt(15), 1),
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

		account, err := lcdClient.LoadAccount(context.Background(), addr)
		if err != nil {
			logger.Error("Error loading address", err.Error())
			continue
		}

		// Create tx
		tx, err := createTransaction(lcdClient, addr, toAddr, amountToMove, account.GetAccountNumber(), account.GetSequence())
		var error = ""

		if err != nil {
			error = err.Error()
		}

		var errCount = 0
		for err != nil {
			//logger.Error("Error creating transaction", err.Error())
			if strings.Contains(error, "sequence") {
				i := strings.Index(error, "expected") + 9
				e := strings.Index(error, ", got")
				seqNumber, err := strconv.ParseUint(error[i:e], 10, 64)
				if err != nil {
					logger.Error(fmt.Sprintf("Could not parse sequence number %s, falbacking to %d", error[i:e], account.GetSequence()+1))
					seqNumber = account.GetSequence() + 1
				}

				if errCount > 3 {
					break // stop trying
				}

				if errCount > 1 {
					logger.Info("Sequence seems stuck, increasing")
					seqNumber++
				}

				//				logger.Info(fmt.Sprintf("Retrying with sequence %d", seqNumber))

				// retry with correct sequence number
				tx, err = createTransaction(
					lcdClient,
					addr,
					toAddr,
					amountToMove,
					account.GetAccountNumber(),
					seqNumber,
				)
				if err != nil {
					//					logger.Error("Error creating transaction", err.Error())
					error = err.Error()
					errCount++
				}
				time.Sleep(sleep_time)
			} else {
				break
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

func createTransaction(lcdClient *client.LCDClient, addr msg.AccAddress, toAddr sdk.AccAddress, amountToMove int64, accountNumber uint64, seqNumber uint64) (*tx2.Builder, error) {
	//	logger.Info(fmt.Sprintf("Creating TX with seq# %d", seqNumber))
	return lcdClient.CreateAndSignTx(
		context.Background(),
		client.CreateTxOptions{
			Msgs: []msg.Msg{
				msg.NewMsgSend(addr, toAddr, msg.NewCoins(msg.NewInt64Coin(query_denom, amountToMove))),
			},
			Memo:          memo,
			AccountNumber: accountNumber,
			Sequence:      seqNumber,
			SignMode:      tx2.SignModeDirect,
		})
}
