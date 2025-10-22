package main

// Central constants & helpers shared across the contract.
// Keep names short but clear; avoid any heavy imports here to keep code size down.

// -------- Counters (state keys) --------
const (
	// Global incremental counters
	NFTsCount       = "nft_count" // next NFT id to assign
	CollectionCount = "col_count" // next collection id to assign
)

// -------- Limits (tune as needed) --------
const (
	maxNameLength = 48
	maxDescLength = 128
)

// -------- Authorization helpers --------

// isAuthorized returns true if caller is the owner (for the resource in question) or the marketplace.
func isAuthorized(caller *string, owner *string, market *string) bool {
	if caller == nil || *caller == "" {
		return false
	}
	if owner != nil && *caller == *owner {
		return true
	}
	if market != nil && *market != "" && *caller == *market {
		return true
	}
	return false
}

// validateMintArgs does minimal structural checks for mint; metadata is treated as opaque elsewhere.
func validateMintArgs(name, desc string, _ interface{}) {
	if len(name) == 0 || len(name) > maxNameLength {
		abort("invalid name length")
	}
	if len(desc) > maxDescLength {
		abort("description too long")
	}
}

// Small inline abort wrapper so other files don't need to import sdk just for messages here.
func abort(msg string) {
	// Delegated to helpers/events/etc. which import sdk directly.
	// This indirection keeps main.go light; callers should use sdk.Abort directly where they already import it.
	panic(msg)
}
