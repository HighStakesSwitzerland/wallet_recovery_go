package addr

import (
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"os"
)

func GenerateAddresses(mnemonic string, hdPath string, dest_wallet string) ([]byte, []byte, keyring.Keyring) {
	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")
	fmt.Println("Seed: ", hex.EncodeToString(seed))
	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(master, ch, hdPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("Using Derivation Path: ", hdPath)
	privKey := &secp256k1.PrivKey{Key: priv}
	encoded := types.AccAddress(privKey.PubKey().Address()).String()

	fmt.Println("Wallet Address from mnemonic:", encoded)

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

	// save keys to memory keyring
	tmpDir, _ := os.MkdirTemp("", ".wallet_recovery_go")
	kb, err := keyring.New(types.KeyringServiceName(), keyring.BackendMemory, tmpDir, nil)

	_, err = kb.NewAccount("memory", mnemonic, "", hdPath, hd.Secp256k1)
	if err != nil {
		panic(err)
	}

	return fromAddr, toAddr, kb
}
