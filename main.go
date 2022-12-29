package main

import (
	"context"
	"fmt"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/addr"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/tx"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/core/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"os"
)

func main() {
	defer config.Logger.Sync()
	// set prefix globally
	types.GetConfig().SetBech32PrefixForAccount(config.Bech32Prefix, config.Bech32Prefix+"pub")

	fromAddr, toAddr, kb := addr.GenerateAddresses(config.Mnemonic, config.HdPath, config.Dest_wallet)

	grpcConn := tx.SetupGrpc(config.Grpc_client)
	defer grpcConn.Close() // close connection on program exit

	config.Logger.Info("Started")

	txBytes := tx.CreateSendTx(fromAddr, toAddr, config.Amount_to_snipe, config.Fees_amount, kb)

	c, err := http.New(config.Rpc_client, "/websocket")
	if err != nil {
		panic(err)
	}

	// call Start/Stop if you're subscribing to events
	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer c.Stop()

	out, err := c.Subscribe(context.Background(), "127.0.0.1", "tm.event = 'Unlock'")
	if err != nil {
		panic(err)
	}

	config.Logger.Info("Connected to websocket, listening for Unlock events")
	for {
		select {
		case resultEvent := <-out:
			fmt.Println(resultEvent)
			checkBalanceAndWithdraw(grpcConn, txBytes, resultEvent)
		case <-c.Quit():
			config.Logger.Info("Disconnected from websocket")
			return
		}
	}

}

func checkBalanceAndWithdraw(grpcConn *grpc.ClientConn, txBytes []byte, event coretypes.ResultEvent) {

	// /cosmos/staking/v1beta1/validators/secretvaloper1xj5ykuzn0mkq9642yxgqmh4ycplzhr2pza25mk/unbonding_delegations\?pagination.limit=1000
	// Send tx
	// tx.SendTx(txBytes)

}

func init() {
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()), os.Stdout, zap.DebugLevel))
	config.Logger = logger
}
