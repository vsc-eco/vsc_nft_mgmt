package main

import (
	"fmt"
	"vsc_nft_mgmt/sdk"
)

// Set the upcoming market contract
//
//go:wasmexport admin_set_market
func SetMarketContract(address *string) *string {
	if *address == "" {
		sdk.Abort("market address needed")
	}

	creator := sdk.GetEnv().Sender.Address
	contractOwner := "contractOwnerAddress" // TODO: set vsc administrative account here
	if creator.String() != contractOwner {
		sdk.Abort(fmt.Sprintf("market only be set by %s", contractOwner))

	}
	sdk.StateSetObject(adminKey("marketContract"), *address)
	return nil
}

func getMarketContract() (sdk.Address, error) {
	contract := sdk.StateGetObject(adminKey("marketContract"))
	if contract == nil {
		return "", fmt.Errorf("market not set")
	}
	return sdk.Address(*contract), nil
}

func adminKey(keyName string) string {
	return fmt.Sprintf("admin:%s", keyName)
}
