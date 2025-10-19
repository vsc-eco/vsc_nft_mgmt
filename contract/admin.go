package main

import (
	"fmt"
	"vsc_nft_mgmt/sdk"
)

// ContractCreator is the address that deployed the contract and has admin rights.
var ContractCreator = "hive:contractowner"

// SetMarketContract sets the address of the market contract.
// Only the ContractCreator is authorized to call this function.
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
	sdk.StateSetObject("mc", *address)
	return nil
}

// GetMarket returns the address of the market contract currently set.
//
//go:wasmexport get_market
func GetMarket(id *string) *string {
	return getMarketContract()
}

// getMarketContract retrieves the stored market contract address from state.
func getMarketContract() *string {
	contract := sdk.StateGetObject("mc")
	return contract
}
