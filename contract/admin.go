package main

import (
	"vsc_nft_mgmt/sdk"
)

// =============================
// Contract Owner + Market Admin
// =============================

// We keep the storage key extremely short to minimize state I/O cost.
const marketKey = "mc"

// Contract owner is set at deployment time or hardcoded.
// For maximum efficiency, this should be a literal string with no parsing during execution.
var contractOwner = "hive:contractowner"

// =============================
// Exported Functions (ABI: string-only)
// =============================

//go:wasmexport set_market
func SetMarketplace(addr *string) *string {
	if addr == nil || *addr == "" {
		sdk.Abort("market address required")
	}

	caller := sdk.GetEnvKey("msg.sender")
	if caller == nil || *caller != contractOwner {
		sdk.Abort("only contract owner can set market")
	}

	sdk.StateSetObject(marketKey, *addr)
	return nil
}

//go:wasmexport get_market
func GetMarketplace(_ *string) *string {
	return sdk.StateGetObject(marketKey)
}

// =============================
// Internal Helper (for NFT transfers)
// =============================
func getMarketContract() *string {
	return sdk.StateGetObject(marketKey)
}
