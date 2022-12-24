package tx

import (
	"context"
	"fmt"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
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
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
	)

	if err != nil {
		panic(err)
	}

	return grpcConn
}

func SendTx(txBytes []byte) {
	txClient := typestx.NewServiceClient(grpcConn)
	grpcRes, err := txClient.BroadcastTx(
		context.Background(),
		&typestx.BroadcastTxRequest{
			Mode:    typestx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
		},
	)

	if err != nil {
		fmt.Println("ERROR: can't send tx to GRPC", err)
	}

	fmt.Println("TX Result:", grpcRes.TxResponse) // Should be `0` if the tx is successful
}
