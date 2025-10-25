package main

import (
	"vsc_nft_mgmt/sdk"
)

// =====================
// Market Administration
// =====================

const marketKey = "mc"

// AddMarketContract registers (an additional) marketplace contract address.
// Only the contract owner may call this.
//
//go:wasmexport add_market
func AddMarketContract(addr *string) *string {
	if addr == nil || *addr == "" {
		sdk.Abort("market address required")
	}
	sender := sdk.GetEnvKey("msg.sender")
	if sender == nil || *sender != contractOwner {
		sdk.Abort("only contract owner can set market")
	}

	existing := GetMarketContractsCSV(nil) // load current market list
	if existing != nil && *existing != "" {
		// if we already have at least one market contract
		// prevent duplicate
		if containsInCSV(*existing, *addr) {
			return nil
		}
		newVal := *existing + "|" + *addr // append new contract to list
		sdk.StateSetObject(marketKey, newVal)
	} else {
		// if we do not have any market contract yet
		sdk.StateSetObject(marketKey, *addr)
	}
	return nil
}

// RemoveMarketContract removes a marketplace contract address from the supported contracts list.
// Only the contract owner may call this.
//

//go:wasmexport remove_market
func RemoveMarketContract(addr *string) *string {
	if addr == nil || *addr == "" {
		sdk.Abort("market address required")
	}
	sender := sdk.GetEnvKey("msg.sender")
	if sender == nil || *sender != contractOwner {
		sdk.Abort("only contract owner can remove market")
	}

	existing := GetMarketContractsCSV(nil)
	if existing == nil || *existing == "" {
		sdk.Abort("no marketplaces found")
	}

	newCSV := removeFromCSV(*existing, *addr)
	sdk.StateSetObject(marketKey, newCSV)
	return nil
}

// GetMarketContractsCSV retrieves the currently configured market contracts
//
//go:wasmexport get_markets
func GetMarketContractsCSV(_ *string) *string {
	return sdk.StateGetObject(marketKey)
}
