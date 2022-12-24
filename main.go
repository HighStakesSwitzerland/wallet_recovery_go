package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/key"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/decred/dcrd/bech32"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"os"
)

var (
	logger      = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	mnemonic    = "grant rice replace explain federal release fix clever romance raise often wild taxi quarter soccer fiber love must tape steak together observe swap guitar"
	dest_wallet = "secret16mu3ttz3u3dj5fppvms86vm0jv59rllyza8pmq"
	lcd_client  = "https://lcd.testnet.secretsaturn.net"
	rpc_client  = "https://rpc.pulsar.scrttestnet.com:443"
	grpc_client = "grpcbin.pulsar.scrttestnet.com:9099"
	query_denom = "uscrt"
	memo        = "yay \\o/"
)

func main() {
	types.GetConfig().SetBech32PrefixForAccount("secret", "secretpub")

	logger.Info("Started")

	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")
	fmt.Println("Seed: ", hex.EncodeToString(seed))
	master, ch := hd.ComputeMastersFromSeed(seed)
	path := "m/44'/529'/0'/0/0"
	priv, err := hd.DerivePrivateKeyForPath(master, ch, path)
	if err != nil {
		panic(err)
	}
	fmt.Println("Derivation Path: ", path)
	fmt.Println("Private Key: ", hex.EncodeToString(priv))
	privKey := &secp256k1.PrivKey{Key: priv}
	pubKey := privKey.PubKey()

	fmt.Println("Public Key: ", hex.EncodeToString(pubKey.Bytes()))
	fmt.Println("Public Address: ", pubKey.Address())
	decodeString, err := hex.DecodeString(fmt.Sprintf("04%x", pubKey.Bytes()))
	if err != nil {
		panic(err)
	}

	conv, err := bech32.ConvertBits(decodeString, 8, 5, true)
	if err != nil {
		fmt.Println("Error:", err)
	}
	encoded, err := bech32.Encode("secret", conv) //TODO: global variable
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Show the encoded data.
	fmt.Println("Wallet Address:", encoded)

	privKeyBz, err := key.DerivePrivKeyBz(mnemonic, key.CreateHDPath(0, 0))
	if err != nil {
		logger.Error("Error creating priv key", err.Error())
		return
	}
	privKey2, err := key.PrivKeyGen(privKeyBz)
	if err != nil {
		logger.Error("Error creating priv key", err.Error())
		return
	}

	fmt.Println("Pubkey Wallet Address:", privKey.PubKey().Address().String())
	encoded = types.AccAddress(privKey.PubKey().Address()).String()
	fmt.Println("Correct Wallet Address:", encoded)

	fmt.Println("Pubkey Wallet Address2:", privKey2.PubKey().Address().String())
	addr2 := types.AccAddress(privKey2.PubKey().Address())
	fmt.Println("Correct Wallet Address2:", addr2.String())

	toAddr, err := types.GetFromBech32(dest_wallet, "secret")
	if err != nil {
		panic(err)
	}
	err = types.VerifyAddressFormat(toAddr)
	if err != nil {
		panic(err)
	}

	fromAddr, err := types.GetFromBech32(encoded, "secret")
	if err != nil {
		panic(err)
	}
	err = types.VerifyAddressFormat(toAddr)
	if err != nil {
		panic(err)
	}

	msg1 := banktypes.NewMsgSend(fromAddr, toAddr, types.NewCoins(types.NewInt64Coin("scrt", 12)))

	// Create a new TxBuilder.
	encCfg := simapp.MakeTestEncodingConfig()
	txBuilder := encCfg.TxConfig.NewTxBuilder()

	err = txBuilder.SetMsgs(msg1)
	if err != nil {
		panic(err)
	}

	tmpDir, _ := os.MkdirTemp("", ".wallet_recovery_go")
	kb, err := keyring.New(types.KeyringServiceName(), keyring.BackendMemory, tmpDir, nil)

	_, err = kb.NewAccount("memory", mnemonic,
		"passphrase", path, hd.Secp256k1)
	if err != nil {
		panic(err)
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID("pulsar-1").
		WithMemo("my memo").
		WithKeybase(kb).
		WithTxConfig(encCfg.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	err = tx.Sign(txFactory, keyring.BackendMemory, txBuilder, false)
	if err != nil {
		panic(err)
	}

	txJSONBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}
	txJSON := string(txJSONBytes)
	logger.Info("TX:", txJSON)

	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	sendTx(txBytes)

	/*
		c, err := http.New(rpc_client, "/websocket")
		if err != nil {
			panic(err)
		}

		// call Start/Stop if you're subscribing to events
		err = c.Start()
		if err != nil {
			panic(err)
		}
		defer c.Stop()

		query := "tm.event = 'NewBlockHeader'"
		out, err := c.Subscribe(context.Background(), "127.0.0.1", query)
		if err != nil {
			panic(err)
		}

			for {
				select {
				case resultEvent := <-out:
					// We should have a switch here that performs a validation
					// depending on the event's type.
					logger.Info(resultEvent.Query)
					//checkBalanceAndWithdraw(lcdClient, toAddr, addr)
				case <-c.Quit():
					logger.Info("Disconnected")
					return
				}
			}
	*/

}

func sendTx(txBytes []byte) {
	// --snip--

	// Create a connection to the gRPC server.
	grpcConn, _ := grpc.Dial(
		grpc_client,         // Or your gRPC server address.
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
	)
	defer grpcConn.Close()

	// Broadcast the tx via gRPC. We create a new client for the Protobuf Tx
	// service.
	txClient := typestx.NewServiceClient(grpcConn)
	// We then call the BroadcastTx method on this client.
	grpcRes, err := txClient.BroadcastTx(
		context.Background(),
		&typestx.BroadcastTxRequest{
			Mode:    typestx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(grpcRes.TxResponse) // Should be `0` if the tx is successful
}

/*
func checkBalanceAndWithdraw(lcdClient *client.LCDClient, toAddr msg.AccAddress, addr msg.AccAddress) {
	balance, err := lcdClient.GetBalance(context.Background(), addr, query_denom)
	if err != nil {
		logger.Error("Cannot get balance", err.Error())
		return
	}

	if balance.Amount == "0" {
		logger.Error("balance = 0")
		return
	}

	amount, err := strconv.ParseInt(balance.Amount, 10, 64)
	if err != nil {
		logger.Error("Cannot convert balance", err.Error())
		return
	}

	logger.Info(fmt.Sprintf("Detected valid balance: %s moving %d", balance, amount))

	account, err := lcdClient.LoadAccount(context.Background(), addr)
	if err != nil {
		logger.Error("Error loading address", err.Error())
		return
	}

	// Create tx
	var tx *tx2.Builder
	tx, err = createTransaction(lcdClient, addr, toAddr, amount, account.GetAccountNumber(), account.GetSequence(), balance)

	if err != nil {
		var errorMsg = err.Error()
		logger.Error("Error creating transaction", errorMsg)
		if strings.Contains(errorMsg, "sequence") {
			i := strings.Index(errorMsg, "expected") + 9
			e := strings.Index(errorMsg, ", got")
			seqNumber, err := strconv.ParseUint(errorMsg[i:e], 10, 64)
			if err != nil {
				logger.Error(fmt.Sprintf("Could not parse sequence number %s, falbacking to %d", errorMsg[i:e], account.GetSequence()+1))
				seqNumber = account.GetSequence() + 1
			}

			logger.Info(fmt.Sprintf("Retrying with sequence %d", seqNumber))
			// amount - 1luna in case a failed tx consumed some fees
			tx, err = createTransaction(lcdClient, addr, toAddr, amount-1000000, account.GetAccountNumber(), seqNumber, balance)

			if err != nil {
				logger.Error("Error retrying transaction", err.Error())
			} else {
				logger.Info("Retry with new sequence Success:", tx)
			}
		}
		if strings.Contains(errorMsg, "insufficientfunds") {
			i := strings.Index(errorMsg, "messageindex:0:") + 9
			e := strings.Index(errorMsg, "ulunaissmallerthan")
			correctAmount, err := strconv.ParseUint(errorMsg[i:e], 10, 64)
			if err != nil {
				logger.Error(fmt.Sprintf("Could not parse new amount %s", errorMsg[i:e]))
			}
			logger.Info(fmt.Sprintf("Retrying correct amount %d", correctAmount))
			// amount - 1luna in case a failed tx consumed some fees
			tx, err = createTransaction(lcdClient, addr, toAddr, int64(correctAmount), account.GetAccountNumber(), account.GetSequence(), balance)

			if err != nil {
				logger.Error("Error retrying transaction", err.Error())
			} else {
				logger.Info("Retry with new amount Success:", tx)
			}
		}
	}

	if err == nil {
		// Broadcast
		res, err := lcdClient.Broadcast(context.Background(), tx)
		if err != nil {
			logger.Error("Error broadcasting tx", err)
			// panic(err) uncomment for stacktrace on exception
		}
		logger.Info("Success:", res)
		time.Sleep(2500) // wait a bit
	}

}

func createTransaction(lcdClient *client.LCDClient, addr msg.AccAddress, toAddr sdk.AccAddress, amountToMove int64,
	accountNumber uint64, seqNumber uint64, balance *client.QueryAccountBalance) (*tx2.Builder, error) {

	logger.Info(fmt.Sprintf("Creating TX with seq# %d and amount %d", seqNumber, amountToMove))
	return lcdClient.CreateAndSignTx(
		context.Background(),
		client.CreateTxOptions{
			Denom:         query_denom,
			Addr:          addr,
			ToAddr:        toAddr,
			Amount:        amountToMove,
			Memo:          memo,
			AccountNumber: accountNumber,
			Sequence:      seqNumber,
			SignMode:      tx2.SignModeDirect,
			Balance:       balance,
		})
}
*/
