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
	caller := sdk.GetEnvKey("msg.sender")
	if caller == nil || *caller != contractOwner {
		sdk.Abort("only contract owner can set market")
	}

	existing := sdk.StateGetObject(marketKey)
	if existing != nil && *existing != "" {
		// prevent duplicate
		if containsInCSV(*existing, *addr) {
			return nil
		}
		newVal := *existing + "," + *addr
		sdk.StateSetObject(marketKey, newVal)
	} else {
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
	caller := sdk.GetEnvKey("msg.sender")
	if caller == nil || *caller != contractOwner {
		sdk.Abort("only contract owner can remove market")
	}

	existing := sdk.StateGetObject(marketKey)
	if existing == nil || *existing == "" {
		sdk.Abort("no marketplaces found")
	}

	newCSV := removeFromCSV(*existing, *addr)
	sdk.StateSetObject(marketKey, newCSV)
	return nil
}

// GetMarketContractsCSV retrieves the currently configured market contracts as comma separated string.
//
//go:wasmexport get_market
func GetMarketContractsCSV(_ *string) *string {
	return sdk.StateGetObject(marketKey)
}
