package main

import (
	"vsc_nft_mgmt/sdk"
)

// =============================
// Contract Owner + Market Admin
// =============================

// Very short key for minimal storage footprint.
const marketKey = "mc"

// contractOwner is hardcoded at deployment. Using const makes it immutable
// and slightly cheaper to reference compared to a var.
const contractOwner = "hive:contractowner"

// =============================
// Exported Functions (ABI: string-only)
// =============================

// SetMarketplace registers the marketplace contract address.
// Only the contract owner may call this.
//
//go:wasmexport set_market
func SetMarketplace(addr *string) *string {
	if addr == nil || *addr == "" {
		sdk.Abort("market address required")
	}

	// Faster than re-dereferencing pointer multiple times.
	callerPtr := sdk.GetEnvKey("msg.sender")
	if callerPtr == nil || *callerPtr != contractOwner {
		sdk.Abort("only contract owner can set market")
	}

	sdk.StateSetObject(marketKey, *addr)
	return nil
}

// GetMarketplace retrieves the currently configured market contract.
//
//go:wasmexport get_market
func GetMarketplace(_ *string) *string {
	return sdk.StateGetObject(marketKey)
}

// =============================
// Internal Helper (for NFT transfers)
// =============================

// getMarketContract returns the market address or nil if unset.
func getMarketContract() *string {
	return sdk.StateGetObject(marketKey)
}
