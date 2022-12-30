package config

import (
	"github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

var (
	Logger *zap.Logger

	Mnemonic                = "grant rice replace explain federal release fix clever romance raise often wild taxi quarter soccer fiber love must tape steak together observe swap guitar"
	DestinationWalletBech32 = "secret19spr25jmjlvv35alu4cnxx4em598ncr326px4a"
	LcdClientUrl            = "https://secret-4.api.trivium.network:1317"
	RpcClientUrl            = "https://secret-4.api.trivium.network:26657" // for websocket queries
	GrpcClientUrl           = "https://secret-4.api.trivium.network:9091"  // to post tx
	Bech32Prefix            = "secret"
	ChainId                 = "secret-4"
	HdPath                  = "m/44'/118'/0'/0/0" // cf cosmos.directory (118 = cosmos, 330 = terra, 529 = secret...)
	CoinsDenom              = "uscrt"
	FeesAmount              = types.NewCoins(types.NewInt64Coin("uscrt", 2_000))
	GasLimit                = uint64(20000) // important! check tx on mintscan to see was is the usual gas
	Memo                    = "\\o/"
)
