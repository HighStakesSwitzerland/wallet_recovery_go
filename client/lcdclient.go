package client

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"net/http"
	"strconv"
	"time"

	"github.com/HighStakesSwitzerland/wallet_recovery_go/key"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/msg"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/tx"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	terraapp "github.com/terra-money/core/app"
	terraappparams "github.com/terra-money/core/app/params"
)

// LCDClient outer interface for building & signing & broadcasting tx
type LCDClient struct {
	URL           string
	RPC           string
	ChainID       string
	GasPrice      msg.DecCoin
	GasAdjustment msg.Dec

	PrivKey        key.PrivKey
	EncodingConfig terraappparams.EncodingConfig

	c *http.Client
}

// NewLCDClient create new LCDClient
func NewLCDClient(URL, RPC, chainID string, gasPrice msg.DecCoin, gasAdjustment msg.Dec, tmKey key.PrivKey, httpTimeout time.Duration) *LCDClient {
	return &LCDClient{
		URL:            URL,
		RPC:            RPC,
		ChainID:        chainID,
		GasPrice:       gasPrice,
		GasAdjustment:  gasAdjustment,
		PrivKey:        tmKey,
		EncodingConfig: terraapp.MakeEncodingConfig(),
		c:              &http.Client{Timeout: httpTimeout},
	}
}

// CreateTxOptions tx creation options
type CreateTxOptions struct {
	Msgs []msg.Msg
	Memo string

	// Optional parameters
	AccountNumber uint64
	Sequence      uint64
	GasLimit      uint64
	FeeAmount     msg.Coins

	SignMode      tx.SignMode
	FeeGranter    msg.AccAddress
	TimeoutHeight uint64
	Balance       *QueryAccountBalance
	ToAddr        sdk.AccAddress
	Amount        int64
	Addr          msg.AccAddress
	Denom         string
}

// CreateAndSignTx build and sign tx
func (lcd *LCDClient) CreateAndSignTx(ctx context.Context, options CreateTxOptions) (*tx.Builder, error) {
	txbuilder := tx.NewTxBuilder(lcd.EncodingConfig.TxConfig)
	txbuilder.SetFeeAmount(options.FeeAmount)
	txbuilder.SetFeeGranter(options.FeeGranter)
	txbuilder.SetGasLimit(options.GasLimit)
	txbuilder.SetMemo(options.Memo)
	txbuilder.SetTimeoutHeight(options.TimeoutHeight)
	txbuilder.SetMsgs(msg.NewMsgSend(options.Addr, options.ToAddr, msg.NewCoins(msg.NewInt64Coin(options.Denom, options.Amount))))

	// use direct sign mode as default
	if tx.SignModeUnspecified == options.SignMode {
		options.SignMode = tx.SignModeDirect
	}

	if options.AccountNumber == 0 || options.Sequence == 0 {
		account, err := lcd.LoadAccount(ctx, msg.AccAddress(lcd.PrivKey.PubKey().Address()))
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to load account")
		}

		options.AccountNumber = account.GetAccountNumber()
		options.Sequence = account.GetSequence()
	}

	gasLimit := int64(options.GasLimit)
	if options.GasLimit == 0 {
		simulateRes, err := lcd.Simulate(ctx, txbuilder, options)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to simulate")
		}

		gasLimit = lcd.GasAdjustment.MulInt64(int64(simulateRes.GasInfo.GasUsed)).TruncateInt64()
		txbuilder.SetGasLimit(uint64(gasLimit))
	}

	if options.FeeAmount.IsZero() {
		computeTaxRes, err := lcd.ComputeTax(ctx, txbuilder)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to compute tax")
		}

		gasFee := msg.NewCoin(lcd.GasPrice.Denom, lcd.GasPrice.Amount.MulInt64(gasLimit).TruncateInt())
		txbuilder.SetFeeAmount(computeTaxRes.TaxAmount.Add(gasFee))
	}

	// adjust amount if fees + current amount > balance
	fees := txbuilder.TxBuilder.GetTx().GetFee().AmountOf(options.Denom).Int64()
	total := fees + options.Amount
	balanceAmount, _ := strconv.ParseInt(options.Balance.Amount, 10, 64)
	maxAmount := balanceAmount - fees
	if total > balanceAmount {
		logger.Info(fmt.Sprintf("Amount to send is more than balance, changing to %d", maxAmount))
		if maxAmount < 0 {
			return nil, fmt.Errorf("negative balance")
		}
		txbuilder.SetMsgs(msg.NewMsgSend(options.Addr, options.ToAddr, msg.NewCoins(msg.NewInt64Coin(options.Denom, maxAmount))))
	}

	err := txbuilder.Sign(options.SignMode, tx.SignerData{
		AccountNumber: options.AccountNumber,
		ChainID:       lcd.ChainID,
		Sequence:      options.Sequence,
	}, lcd.PrivKey, true)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to sign tx")
	}

	return &txbuilder, nil
}
