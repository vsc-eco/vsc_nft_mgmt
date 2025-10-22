package main

// ========================================
// Global Contract Constants & Base Helpers
// ========================================
//
// This file holds core constants and lightweight utility functions shared
// across the contract modules. The goal is to keep this file dependency-free
// (no heavy imports) to minimize WASM size and compile time.
//

// ----------------------------------------
// State Keys - Global Counters
// ----------------------------------------
//
// These counters are stored in contract state and incremented automatically
// via setCount(). They define the next ID to assign for collections and NFTs.
const (
	// NFTsCount tracks the next NFT ID to be assigned.
	NFTsCount = "nft_count"

	// CollectionCount tracks the next collection ID.
	CollectionCount = "col_count"
)

// ----------------------------------------
// User Input Limits
// ----------------------------------------
//
// These constants prevent abuse of storage size by limiting
// user-supplied fields. Adjust carefully, as increasing these
// values increases worst-case gas.
const (
	maxNameLength = 48  // upper bound for NFT or collection names
	maxDescLength = 128 // max allowed chars for description fields
)

// ----------------------------------------
// Authorization Logic
// ----------------------------------------
//
// isAuthorized ensures the caller is either the NFT owner
// or a trusted marketplace contract (if provided).
// Caller and market values are string pointers to save
// unnecessary allocations.
//

// isAuthorized returns true if caller == owner OR caller == market.
// If both validations fail, returns false.
//
// It’s intentionally small because this is hit in hot code paths like transfer/burn.
func isAuthorized(caller, owner, market *string) bool {
	if caller == nil {
		return false
	}
	c := *caller
	// Direct owner match (cheapest possible path)
	if owner != nil && c == *owner {
		return true
	}
	// Market match only if market is set and not empty
	if market != nil {
		m := *market
		if m != "" && c == m {
			return true
		}
	}
	return false
}

// ----------------------------------------
// Lightweight Panic Wrapper
// ----------------------------------------
//
// abort is a minimal helper used to avoid importing the SDK in this file.
// Other modules call sdk.Abort directly where appropriate.
// Here we use panic as a signal that execution should be halted.
func abort(msg string) {
	// panic is used locally. Runtime will be intercepted by sdk.Abort
	// when integrated in full wasm execution env.
	panic(msg)
}
