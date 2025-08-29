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
		abortCustom("market contract address is mandatory")
	}

	creator := getSenderAddress()
	contractOwner := "contractOwnerAddress" // TODO: set vsc administrative account here
	if creator.String() != contractOwner {
		abortCustom(fmt.Sprintf("market only be set by %s", contractOwner))

	}
	getStore().Set(adminKey("marketContract"), *address)
	return returnJsonResponse(
		true, map[string]interface{}{
			"message": fmt.Sprintf("market set to %s", address),
		},
	)
}

func getMarketContract() (sdk.Address, error) {
	contract := getStore().Get(adminKey("marketContract"))
	if contract == nil {
		return "", fmt.Errorf("market not set")
	}
	return sdk.Address(*contract), nil
}
