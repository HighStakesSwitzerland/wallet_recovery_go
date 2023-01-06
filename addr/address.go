package addr

import (
	"encoding/hex"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
)

var (
	FromAddr     []byte
	ToAddr       []byte
	Bech32wallet string
	PrivKey      *secp256k1.PrivKey
)

// GenerateAddresses /*
// Generates destination wallet bech32 address from mnemonic and checks everything
// Sets the address as package global variable
func GenerateAddresses() {
	// seed, _ := bip39.NewSeedWithErrorChecking(config.Mnemonic, "")
	// master, ch := hd.ComputeMastersFromSeed(seed)
	// mnemonicPriv, err := hd.DerivePrivateKeyForPath(master, ch, config.HdPath)
	// if err != nil {
	//	panic(err)
	//}
	config.Logger.Info("Using Derivation Path: " + config.HdPath)
	//	PrivKey = &secp256k1.PrivKey{Key: mnemonicPriv}

	decodeString, err := hex.DecodeString("1daac0ba8a73b9ea36ab70aca2ce43ec06c3ffa9c45e159ac781484d84a5d9ef") // TODO: à AirV
	if err != nil {
		panic(err)
	}
	PrivKey = &secp256k1.PrivKey{Key: decodeString}

	Bech32wallet = types.AccAddress(PrivKey.PubKey().Address()).String()

	config.Logger.Info("Wallet Address decoded: " + Bech32wallet)

	ToAddr, err = types.GetFromBech32(config.DestinationWalletBech32, config.Bech32Prefix)
	if err != nil {
		panic(err)
	}
	err = types.VerifyAddressFormat(ToAddr)
	if err != nil {
		panic(err)
	}

	FromAddr, err = types.GetFromBech32(Bech32wallet, config.Bech32Prefix)
	if err != nil {
		panic(err)
	}
	err = types.VerifyAddressFormat(FromAddr)
	if err != nil {
		panic(err)
	}
}
