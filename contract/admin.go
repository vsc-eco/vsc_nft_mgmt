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
	if address == nil || *address == "" {
		sdk.Abort("market address needed")
	}
	sender := sdk.GetEnvKey("msg.sender")
	if *sender != ContractCreator {
		sdk.Abort(fmt.Sprintf("only %s can set", ContractCreator))
	}
	sdk.StateSetObject("a:mc", *address)
	return nil
}

func getMarketContract() *string {
	contract := sdk.StateGetObject("a:mc")
	return contract
}
