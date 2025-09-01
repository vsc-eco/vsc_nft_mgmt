package main

import (
	"fmt"
	"vsc_nft_mgmt/sdk"
)

var ContractCreator = "hive:tibfox.vsc" // TODO: set contract owner here

// Set the upcoming market contract
//

//go:wasmexport admin_set_market
func SetMarketContract(address *string) *string {
	return setMarketContractImpl(address, RealSDK{})
}

func setMarketContractImpl(address *string, chain SDKInterface) *string {
	if *address == "" {
		chain.Abort("market address needed")
	}
	creator := chain.GetEnv().Sender.Address
	if creator.String() != ContractCreator {
		chain.Abort(fmt.Sprintf("market only be set by %s", ContractCreator))
	}
	chain.StateSetObject(adminKey("marketContract"), *address)
	return nil
}

func getMarketContract(chain SDKInterface) sdk.Address {
	contract := chain.StateGetObject(adminKey("marketContract"))
	if contract == nil {
		return sdk.Address("")
	}
	return sdk.Address(*contract)
}

func adminKey(keyName string) string {
	return fmt.Sprintf("admin:%s", keyName)
}
