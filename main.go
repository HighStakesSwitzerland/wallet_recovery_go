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
	// gas          = msg.NewInt(200000) Not specifying gas make an additional call to simulate the tx before => very usefull
	sleep_time   = time.Millisecond * 1
	query_denom  = "uluna" // uluna or uusd
	memo         = "gcp2"
	amountToMove = int64(1000000) // must be in same unit than query_denom
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
		msg.NewDecCoinFromDec(query_denom, msg.NewDecFromIntWithPrec(msg.NewInt(20), 2)), // GasPrice
		msg.NewDecFromIntWithPrec(msg.NewInt(15), 1),                                     // GasAdjustment 1.5
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

		// Create tx without simulating
		tx, err := createTransaction(lcdClient, addr, toAddr, account.GetAccountNumber(), account.GetSequence())
		if err != nil {
			//logger.Error(fmt.Sprintf("%s", err.Error()))
		}

		if err == nil {
			// Broadcast
			err = broadcasTx(lcdClient, tx)
		}

		if err != nil {
			error := err.Error()
			if strings.Contains(error, "sequence") {
				i := strings.Index(error, "expected") + 9
				e := strings.Index(error, ", got")
				seqNumber, err := strconv.ParseUint(error[i:e], 10, 64)
				if err != nil {
					logger.Error(fmt.Sprintf("Could not parse sequence number %s, falbacking to %d", error[i:e], account.GetSequence()+1))
					seqNumber = account.GetSequence() + 1
				}

				//logger.Info(fmt.Sprintf("Retrying with sequence %d", seqNumber))

				// retry with correct sequence number
				tx, err := createTransaction(
					lcdClient,
					addr,
					toAddr,
					account.GetAccountNumber(),
					seqNumber,
				)
				if err != nil {
					//logger.Error(fmt.Sprintf("Error creating fallback transaction %s", err.Error()))
					error = err.Error()
				} else {
					err = broadcasTx(lcdClient, tx)
					if err != nil {
						//logger.Error(fmt.Sprintf("Error broacasting fallback transaction %s", err.Error()))
						error = err.Error()
					}
				}
				time.Sleep(sleep_time)
			}
		}

	}

}

func createTransaction(lcdClient *client.LCDClient,
	addr msg.AccAddress,
	toAddr sdk.AccAddress,
	accountNumber uint64,
	seqNumber uint64) (*tx2.Builder, error) {

	correctAmount, _ := calculateCorrectAmountAndFees()
	//logger.Info(fmt.Sprintf("Creating TX with seq# %d amount %d%s fees %d", seqNumber, correctAmount, query_denom, feeAmount.AmountOf(query_denom).Int64()))
	return lcdClient.CreateAndSignTx(
		context.Background(),
		client.CreateTxOptions{
			Msgs: []msg.Msg{
				msg.NewMsgSend(addr, toAddr, msg.NewCoins(msg.NewInt64Coin(query_denom, correctAmount))),
			},
			Memo:          memo,
			AccountNumber: accountNumber,
			Sequence:      seqNumber,
			SignMode:      tx2.SignModeDirect,
		})
}
func broadcasTx(lcdClient *client.LCDClient, tx *tx2.Builder) error {
	res, err := lcdClient.Broadcast(context.Background(), tx)
	if err != nil {
		//logger.Error(fmt.Sprintf("Error broadcasting tx: %s", err.Error()))
	} else {
		logger.Info("Success:", res)
	}
	return err
}

func calculateCorrectAmountAndFees() (int64, msg.Coins) {
	var feeAmount msg.Coins
	if query_denom == "uusd" {
		feeAmount = msg.NewCoins(msg.NewInt64Coin(query_denom, fees_uusd))
	} else {
		feeAmount = msg.NewCoins(msg.NewInt64Coin(query_denom, fees_uluna))
	}
	//logger.Info(fmt.Sprintf("operation %d %d", amountToMove, feeAmount.AmountOf(query_denom).Int64()))
	correctAmount := amountToMove - feeAmount.AmountOf(query_denom).Int64()*2 // WHY * 2 ? because of gas i think, unsure
	return correctAmount, feeAmount
}
