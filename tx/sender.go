package tx

import (
	"context"
	"fmt"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	grpcClient *string
	grpcConn   *grpc.ClientConn
)

func SetupGrpc(grpcUrl string) *grpc.ClientConn {
	grpcClient = &grpcUrl

	// Create a connection to the gRPC server.
	var err error
	grpcConn, err = grpc.Dial(
		*grpcClient,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // The Cosmos SDK doesn't support any transport security mechanism.
	)

	if err != nil {
		panic(err)
	}

	return grpcConn
}

func SendTx(txBytes []byte) int64 {
	txClient := tx.NewServiceClient(grpcConn)
	grpcRes, err := txClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
		},
	)

	if err != nil {
		fmt.Println(err)
		return -1
	}

	config.Logger.Info("TX Result:", zap.String("json", grpcRes.TxResponse.RawLog)) // Should be `0` if the tx is successful
	if grpcRes.TxResponse == nil {
		return -1
	}
	if grpcRes.TxResponse.Code != 0 {
		var desc = ""
		switch grpcRes.TxResponse.Code {
		case errors.ErrInsufficientFee.ABCICode():
			desc += errors.ErrInsufficientFee.Error()
		case errors.ErrInsufficientFunds.ABCICode():
			desc += errors.ErrInsufficientFunds.Error()
		case errors.ErrInvalidSequence.ABCICode():
			desc += errors.ErrInvalidSequence.Error()
		case errors.ErrTxInMempoolCache.ABCICode():
			desc += errors.ErrTxInMempoolCache.Error()
		case errors.ErrMempoolIsFull.ABCICode():
			desc += errors.ErrMempoolIsFull.Error()
		case errors.ErrTxTimeoutHeight.ABCICode():
			desc += errors.ErrTxTimeoutHeight.Error()
		case errors.ErrWrongSequence.ABCICode():
			desc += errors.ErrWrongSequence.Error()
		}

		config.Logger.Error("TX FAILED!!", zap.Uint32("code", grpcRes.TxResponse.Code), zap.String("description", desc))
	}
	return int64(grpcRes.TxResponse.Code)
}
