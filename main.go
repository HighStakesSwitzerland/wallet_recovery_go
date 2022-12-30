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
	if err != nil {
		config.Logger.Warn("Some fields could not be read: " + err.Error())
	}

	for _, unbond := range undelegations.UnbondingResponses {
		for _, entry := range unbond.Entries {
			local := entry.CompletionTime.Local()
			cronTime := fmt.Sprintf("%d %d %d %d * *", local.Second(), // add 1 second after unbond just in case? Or send tx on next block?
				local.Minute(), local.Hour(), local.Day())
			balance := entry.Balance.Sub(config.FeesAmount.Amount)
			task := func() {
				config.Logger.Info("Starting Cron job!")
				txBytes := tx.CreateSendTx(addr.FromAddr, addr.ToAddr, types.NewCoins(types.NewInt64Coin(config.CoinsDenom, balance.Int64())))
				tx.SendTx(txBytes)
			}
			_, err = s.CronWithSeconds(cronTime).Tag(balance.String()).Do(task)
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

	go func() {
		query := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyValidator, "osmovaloper1hjct6q7npsspsg3dgvzk3sdf89spmlpf6t4agt") // or NewBlockHeader
		//testQuery := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyDelegator, "osmo1cxvu6m0fatpdtm2286yprkzyrzskjd4zs7d8yn")
		out, err := rpcConn.Subscribe(context.Background(), "127.0.0.1", query)
		if err != nil {
			panic(err)
		}

		config.Logger.Info(fmt.Sprintf("Listening for unbond events for wallet with query [%s]", query))
		for {
			select {
			case resultEvent := <-out:
				withdrawUnlockedAmount(resultEvent.Events)
			case <-rpcConn.Quit():
				config.Logger.Info("Disconnected from websocket") // TODO: reconnect
				return
			}
		}

	}()

	go func() {
		query := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyValidator, "osmovaloper1gy0nyn2hscxxayj2pdyu8axmfvv75nnvhc079s") // or NewBlockHeader
		//testQuery := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyDelegator, "osmo1cxvu6m0fatpdtm2286yprkzyrzskjd4zs7d8yn")
		out, err := rpcConn.Subscribe(context.Background(), "127.0.0.1", query)
		if err != nil {
			panic(err)
		}

		config.Logger.Info(fmt.Sprintf("Listening for unbond events for wallet with query [%s]", query))
		for {
			select {
			case resultEvent := <-out:
				withdrawUnlockedAmount(resultEvent.Events)
			case <-rpcConn.Quit():
				config.Logger.Info("Disconnected from websocket") // TODO: reconnect
				return
			}
		}

	}()

	go func() {
		query := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyValidator, "osmovaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5w37wr9t") // or NewBlockHeader
		//testQuery := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyDelegator, "osmo1cxvu6m0fatpdtm2286yprkzyrzskjd4zs7d8yn")
		out, err := rpcConn.Subscribe(context.Background(), "127.0.0.1", query)
		if err != nil {
			panic(err)
		}

		config.Logger.Info(fmt.Sprintf("Listening for unbond events for wallet with query [%s]", query))
		for {
			select {
			case resultEvent := <-out:
				withdrawUnlockedAmount(resultEvent.Events)
			case <-rpcConn.Quit():
				config.Logger.Info("Disconnected from websocket") // TODO: reconnect
				return
			}
		}

	}()

	query := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyValidator, "osmovaloper1w4x44ek799hvg97x0mfwu6gg5dww2r8fdpqql4") // or NewBlockHeader
	//testQuery := fmt.Sprintf("%s.%s='%s'", types2.EventTypeUnbond, types2.AttributeKeyDelegator, "osmo1cxvu6m0fatpdtm2286yprkzyrzskjd4zs7d8yn")
	out, err := rpcConn.Subscribe(context.Background(), "127.0.0.1", query)
	if err != nil {
		panic(err)
	}

	config.Logger.Info(fmt.Sprintf("Listening for unbond events for wallet with query [%s]", query))
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
	config.Logger.Info("Got new event", zap.Any("object", events))

	recipients := events["transfer.recipient"]
	isValid := false

	for i := 0; i < len(recipients); i++ {
		if recipients[i] == addr.Bech32wallet {
			isValid = true
			config.Logger.Info("Matched on transfer.recipient")
		}
	}

	sender := events["transfer.sender"]
	for i := 0; i < len(sender); i++ {
		if sender[i] == addr.Bech32wallet {
			isValid = true
			config.Logger.Info("Matched on transfer.sender")
		}
	}

	msgsender := events["message.sender"]
	for i := 0; i < len(msgsender); i++ {
		if msgsender[i] == addr.Bech32wallet {
			isValid = true
			config.Logger.Info("Matched on message.sender")
		}
	}

	coinspent := events["coin_spent.spender"]
	for i := 0; i < len(coinspent); i++ {
		if coinspent[i] == addr.Bech32wallet {
			isValid = true
			config.Logger.Info("Matched on coin_spent.spender")
		}
	}

	if !isValid {
		config.Logger.Info("Ignoring event, not for our wallet")
		return
	}

	amountUnbonded := events["unbond.amount"][0]
	coin, err := types.ParseCoinsNormalized(amountUnbonded)

	if err != nil {
		config.Logger.Error("Could not parse coins from undelegate event!", zap.Error(err))
	}
	txBytes := tx.CreateSendTx(addr.FromAddr, addr.ToAddr, coin.Sub(config.FeesAmount))
	tx.SendTx(txBytes)

}
