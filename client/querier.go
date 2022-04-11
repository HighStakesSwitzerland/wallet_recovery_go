package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tendermint/tendermint/libs/log"
	"io/ioutil"
	"os"

	"golang.org/x/net/context/ctxhttp"

	"github.com/HighStakesSwitzerland/wallet_recovery_go/msg"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/tx"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	customauthtx "github.com/terra-money/core/custom/auth/tx"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "querier")
)

// QueryAccountResData response
type QueryAccountResData struct {
	Address       msg.AccAddress `json:"address"`
	AccountNumber msg.Int        `json:"account_number"`
	Sequence      msg.Int        `json:"sequence"`
}

// QueryAccountRes response
type QueryAccountRes struct {
	Account QueryAccountResData `json:"account"`
}

type QueryAccountBalance struct {
	Balance `json:"balance"`
}

func (q QueryAccountBalance) Reset() {
}

func (q QueryAccountBalance) String() string {
	return fmt.Sprintf("%s %s", q.Amount, q.Denom)
}

func (q QueryAccountBalance) ProtoMessage() {
}

type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func (lcd LCDClient) GetBalance(ctx context.Context, address msg.AccAddress, denum string) (res *QueryAccountBalance, err error) {
	// TODO: get balance for all coin at once or not? Maybe not usefull in our scenario, and probably slower
	resp, err := ctxhttp.Get(ctx, lcd.c, lcd.URL+fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", address.String(), denum))
	if err != nil {
		return nil, fmt.Errorf("LCD call failed: %s", err.Error())
	}
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("LCD read failed: %s", err.Error())
	}

	//	logger.Info(fmt.Sprintf("Call returned [%s]", out))
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 status code received: %d %s", resp.StatusCode, resp.Status)
	}

	var response QueryAccountBalance
	err = json.Unmarshal(out, &response)
	return &response, nil
}

// LoadAccount simulates gas and fee for a transaction
func (lcd LCDClient) LoadAccount(ctx context.Context, address msg.AccAddress) (res authtypes.AccountI, err error) {
	resp, err := ctxhttp.Get(ctx, lcd.c, lcd.URL+fmt.Sprintf("/cosmos/auth/v1beta1/accounts/%s", address))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to estimate")
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to read response")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response code %d: %s", resp.StatusCode, string(out))
	}

	var response authtypes.QueryAccountResponse
	err = lcd.EncodingConfig.Marshaler.UnmarshalJSON(out, &response)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to unmarshal response")
	}

	return response.Account.GetCachedValue().(authtypes.AccountI), nil
}

// Simulate tx and get response
func (lcd LCDClient) Simulate(ctx context.Context, txbuilder tx.Builder, options CreateTxOptions) (*sdktx.SimulateResponse, error) {
	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	logger.Info(fmt.Sprintf("Simulating tx with seq# %d", options.Sequence))
	sig := signing.SignatureV2{
		PubKey: &secp256k1.PubKey{},
		Data: &signing.SingleSignatureData{
			SignMode: options.SignMode,
		},
		Sequence: options.Sequence,
	}
	if err := txbuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	bz, err := txbuilder.GetTxBytes()
	if err != nil {
		return nil, err
	}

	reqBytes, err := lcd.EncodingConfig.Marshaler.MarshalJSON(&sdktx.SimulateRequest{
		TxBytes: bz,
	})
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Simulating tx with seq# %s", reqBytes))
	resp, err := ctxhttp.Post(ctx, lcd.c, lcd.URL+"/cosmos/tx/v1beta1/simulate", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to estimate")
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to read response")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response code %d: %s", resp.StatusCode, string(out))
	}

	var response sdktx.SimulateResponse
	err = lcd.EncodingConfig.Marshaler.UnmarshalJSON(out, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (lcd LCDClient) FlushMempool(ctx context.Context) {
	resp, err := ctxhttp.Get(ctx, lcd.c, lcd.RPC+"/unsafe_flush_mempool")
	defer resp.Body.Close()
	if err != nil {
		logger.Error("Failed to flush mempool")
	}
}

// protoTxProvider is a type which can provide a proto transaction. It is a
// workaround to get access to the wrapper TxBuilder's method GetProtoTx().
// Deprecated: It's only used for testing the deprecated Simulate gRPC endpoint
// using a proto Tx field.
type protoTxProvider interface {
	GetProtoTx() *sdktx.Tx
}

// ComputeTax compute tax
func (lcd LCDClient) ComputeTax(ctx context.Context, txbuilder tx.Builder) (*customauthtx.ComputeTaxResponse, error) {
	protoProvider := txbuilder.TxBuilder.(protoTxProvider)
	protoTx := protoProvider.GetProtoTx()
	reqBytes, err := lcd.EncodingConfig.Marshaler.MarshalJSON(&customauthtx.ComputeTaxRequest{
		Tx: protoTx,
	})
	if err != nil {
		return nil, err
	}

	resp, err := ctxhttp.Post(ctx, lcd.c, lcd.URL+"/terra/tx/v1beta1/compute_tax", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to estimate")
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to read response")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response code %d: %s", resp.StatusCode, string(out))
	}

	var response customauthtx.ComputeTaxResponse
	err = lcd.EncodingConfig.Marshaler.UnmarshalJSON(out, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
