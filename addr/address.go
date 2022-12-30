package addr

import (
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"os"
)

var (
	FromAddr     []byte
	ToAddr       []byte
	Bech32wallet string
	Keyring      keyring.Keyring
)

// GenerateAddresses /*
// Generates destination wallet bech32 address from mnemonic and checks everything
// Sets the address as package global variable
func GenerateAddresses() {
	seed, _ := bip39.NewSeedWithErrorChecking(config.Mnemonic, "")
	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(master, ch, config.HdPath)
	if err != nil {
		panic(err)
	}
	config.Logger.Info("Using Derivation Path: " + config.HdPath)
	privKey := &secp256k1.PrivKey{Key: priv}

	Bech32wallet = types.AccAddress(privKey.PubKey().Address()).String()

	config.Logger.Info("Wallet Address from mnemonic: " + Bech32wallet)

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
	err = types.VerifyAddressFormat(ToAddr)
	if err != nil {
		panic(err)
	}

	// save keys to temporary keyring
	tmpDir, _ := os.MkdirTemp("", ".wallet_recovery_go") // TODO: if multiple instances of this app runs, keyring can be overwritten?
	cdc := simapp.MakeTestEncodingConfig().Codec
	Keyring, err = keyring.New(types.KeyringServiceName(), keyring.BackendMemory, tmpDir, nil, cdc)

	_, err = Keyring.NewAccount("memory", config.Mnemonic, "", config.HdPath, hd.Secp256k1)
	if err != nil {
		panic(err)
	}

}
