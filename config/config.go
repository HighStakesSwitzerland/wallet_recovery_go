package config

import (
	"github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

var (
	Logger *zap.Logger

	Mnemonic                = "ripple cinnamon spread police dance auto gentle inflict gossip solve merge clog"
	DestinationWalletBech32 = "secret1fc3fzy78ttp0lwuujw7e52rhspxn8uj52zfyne"
	LcdClientUrl            = "http://212.95.51.215:1317"  // for web queries; http:// required even for localhost
	RpcClientUrl            = "http://212.95.51.215:26657" // for websocket queries; http:// required even for localhost
	GrpcClientUrl           = "212.95.51.215:26090"        // to post tx
	Bech32Prefix            = "secret"
	ChainId                 = "secret-4"
	HdPath                  = "m/44'/529'/0'/0/0" // cf cosmos.directory (118 = cosmos, 330 = terra, 529 = secret...)
	CoinsDenom              = "uscrt"
	FeesAmount              = types.NewInt64Coin(CoinsDenom, 1250)
	GasLimit                = uint64(100000) // important! check tx on mintscan to see was is the usual gas
	Memo                    = ""
)
