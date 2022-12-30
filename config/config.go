package config

import (
	"github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

var (
	Logger *zap.Logger

	Mnemonic                = "jelly shadow frog dirt dragon use armed praise universe win jungle close inmate rain oil canvas beauty pioneer chef soccer icon dizzy thunder meadow"
	DestinationWalletBech32 = "osmo1cxvu6m0fatpdtm2286yprkzyrzskjd4zs7d8yn"
	LcdClientUrl            = "http://212.95.51.215:1319"  // for web queries; http:// required even for localhost
	RpcClientUrl            = "http://212.95.51.215:46657" // for websocket queries; http:// required even for localhost
	GrpcClientUrl           = "212.95.51.215:46090"        // to post tx
	Bech32Prefix            = "osmo"
	ChainId                 = "osmosis-1"
	HdPath                  = "m/44'/118'/0'/0/0" // cf cosmos.directory (118 = cosmos, 330 = terra, 529 = secret...)
	CoinsDenom              = "uosmo"
	FeesAmount              = types.NewInt64Coin(CoinsDenom, 1000)
	GasLimit                = uint64(100000) // important! check tx on mintscan to see was is the usual gas
	Memo                    = "https://media.giphy.com/media/XHr6LfW6SmFa0/giphy.gif"
)
