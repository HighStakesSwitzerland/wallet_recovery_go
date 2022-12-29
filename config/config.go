package config

import (
	"github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

var (
	Logger *zap.Logger

	Mnemonic        = "grant rice replace explain federal release fix clever romance raise often wild taxi quarter soccer fiber love must tape steak together observe swap guitar"
	Dest_wallet     = "cosmos16mu3ttz3u3dj5fppvms86vm0jv59rllyqcngxu"
	Lcd_client      = "https://secret-4.api.trivium.network:1317"
	Rpc_client      = "https://secret-4.api.trivium.network:26657" // for websocket queries
	Grpc_client     = "https://secret-4.api.trivium.network:9091"  // to post tx
	Bech32Prefix    = "cosmos"
	Chain_id        = "cosmoshub-4"
	HdPath          = "m/44'/118'/0'/0/0" // cf cosmos.directory (118 = cosmos, 330 = terra, 529 = secret...)
	Coins_denom     = "uscrt"
	Amount_to_snipe = types.NewCoins(types.NewInt64Coin("uatom", 10))      // amount for the snipe tx (only once on next block)
	Amount_to_spam  = types.NewCoins(types.NewInt64Coin("uscrt", 250_000)) // amount for the spam txs (many many tx sent on next block)
	Fees_amount     = types.NewCoins(types.NewInt64Coin("uatom", 2_000))
	Memo            = "\\o/"
)
