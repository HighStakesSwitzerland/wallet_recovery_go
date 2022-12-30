package main

import (
	"context"
	"fmt"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/addr"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/lcdclient"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/tx"
	"github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/go-co-op/gocron"
	"github.com/tendermint/tendermint/rpc/client/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"os"
	"time"
)

var (
	grpcConn *grpc.ClientConn
)

func init() {
	prodConfig := zap.NewProductionEncoderConfig()
	prodConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(prodConfig), os.Stdout, zap.DebugLevel)) // or zap.DebugLevel for debug logs
	config.Logger = logger
}

func main() {
	defer config.Logger.Sync()                                                                  // write remaining logs on exit if any
	types.GetConfig().SetBech32PrefixForAccount(config.Bech32Prefix, config.Bech32Prefix+"pub") // set prefix globally

	addr.GenerateAddresses()

	grpcConn = tx.SetupGrpc(config.GrpcClientUrl)
	defer grpcConn.Close() // close connection on program exit
	config.Logger.Info("Connected to gRPC " + config.GrpcClientUrl)

	setupSnipingTx()

	setupUnbondingListener()
}

// setupSnipingTx /*
// Generate the one-time TXs to send it at the specified time
func setupSnipingTx() {
	s := gocron.NewScheduler(time.UTC)

	undelegations, err := lcdclient.GetPendingUndelegations()

	for _, unbond := range undelegations.UnbondingResponses {
		for _, entry := range unbond.Entries {
			local := entry.CompletionTime.Local()
			cronTime := fmt.Sprintf("%d %d %d %d * *", local.Second(), // add 1 second after unbond just in case? Or send tx on next block?
				local.Minute(), local.Hour(), local.Day())
			balance := entry.Balance.Sub(config.FeesAmount.AmountOf(config.CoinsDenom))
			txBytes := tx.CreateSendTx(addr.FromAddr, addr.ToAddr, types.NewCoins(types.NewInt64Coin("uscrt", balance.Int64())))
			task := func(transaction []byte) {
				config.Logger.Info("Starting Cron job!")
				tx.SendTx(transaction)
			}
			_, err = s.CronWithSeconds(cronTime).Tag(balance.String()).Do(task, txBytes)
			if err != nil {
				config.Logger.Error("Cannot register Cron task for sniping TX!", zap.Error(err))
			}
		}
	}

	s.StartAsync()

	for _, job := range s.Jobs() {
		config.Logger.Info("Job for Sniping TX registered", zap.Time("time", job.NextRun()), zap.String("with amount", job.Tags()[0]))
	}

	config.Logger.Info("Sniping configuration done")
}

func setupUnbondingListener() {
	rpcConn, err := http.New(config.RpcClientUrl, "/websocket")
	if err != nil {
		panic(err)
	}

	err = rpcConn.Start()
	if err != nil {
		panic(err)
	}
	defer rpcConn.Stop()
	config.Logger.Info("Connected to RPC websocket " + config.RpcClientUrl)

	//query := fmt.Sprintf("%s.%s='%s'", types2.EventTypeCompleteUnbonding, types2.AttributeKeyDelegator, "secret19spr25jmjlvv35alu4cnxx4em598ncr326px4a")  // or NewBlockHeader
	testQuery := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyValidator, "secretvaloper1jgx4pn3acae9esq5zha5ym3kzhq6x60frjwkrp")
	out, err := rpcConn.Subscribe(context.Background(), "127.0.0.1", testQuery)
	if err != nil {
		panic(err)
	}

	config.Logger.Info(fmt.Sprintf("Listening for unbond events for wallet with query [%s]", testQuery))
	for {
		select {
		case resultEvent := <-out:
			withdrawUnlockedAmount(resultEvent.Events)
		case <-rpcConn.Quit():
			config.Logger.Info("Disconnected from websocket") // TODO: reconnect
			return
		}
	}
}

func withdrawUnlockedAmount(events map[string][]string) {

	//amountUnbonded := events["unbond.amount"]

	// Send tx
	// tx.SendTx(txBytes)

}
