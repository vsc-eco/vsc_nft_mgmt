package main

import (
	"fmt"
	"vsc_nft_mgmt/sdk"
)

var ContractCreator = "hive:contractowner"

// Set the upcoming market contract
//
//go:wasmexport set_market
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

//go:wasmexport get_market
func GetMarket(id *string) *string {
	return getMarketContract()
}

func getMarketContract() *string {
	contract := sdk.StateGetObject("a:mc")
	return contract
}
