package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// ===============
// VARIOUS HELPERS
// ===============

// =======================================
// Binary Key Prefixes (Single Byte Tags)
// =======================================
//
// These constants serve as the *first* byte in every stored key.
// They allow us to differentiate data types with extremely low overhead,
// while keeping keys compact and fully deterministic.
//

const (
	kNFTCore    byte = 0x01 // NFT core metadata: "tx|name|desc|meta"
	kOwner      byte = 0x02 // Owner+Collection mapping per NFT ID
	kCreator    byte = 0x03 // Creator address + transfer-flag
	kEdCount    byte = 0x04 // Edition count
	kEdOverride byte = 0x05 // Edition-specific overrides (owner or burned)
	kOwnedIdx   byte = 0x06 // Owned edition index for quick lookup
)

//
// ==============================
// Inline Packing (Little Endian)
// ==============================
//
// We use direct array writes instead of append() to *completely avoid heap allocation*.
// This pattern is used instead of make()+append for maximum gas efficiency.
//

// packU64LEInline writes uint64 into dst in little-endian format.
//
//go:inline
func packU64LEInline(x uint64, dst []byte) {
	dst[0] = byte(x)
	dst[1] = byte(x >> 8)
	dst[2] = byte(x >> 16)
	dst[3] = byte(x >> 24)
	dst[4] = byte(x >> 32)
	dst[5] = byte(x >> 40)
	dst[6] = byte(x >> 48)
	dst[7] = byte(x >> 56)
}

// packU32LEInline writes uint32 into dst in little-endian format.
//
//go:inline
func packU32LEInline(x uint32, dst []byte) {
	dst[0] = byte(x)
	dst[1] = byte(x >> 8)
	dst[2] = byte(x >> 16)
	dst[3] = byte(x >> 24)
}

//
// ==============================
// Key Builders (Zero Heap Alloc)
// ==============================
//
// Using fixed-size byte arrays ensures the Go compiler places them on the stack,
// fully avoiding GC and allocations.

// nftCoreKey builds the storage key for NFT core metadata.
func nftCoreKey(nftID uint64) string {
	var buf [9]byte
	buf[0] = kNFTCore
	packU64LEInline(nftID, buf[1:])
	return string(buf[:])
}

// ownerKey stores "<owner>_<collection>" as a single string.
func ownerKey(nftID uint64) string {
	var buf [9]byte
	buf[0] = kOwner
	packU64LEInline(nftID, buf[1:])
	return string(buf[:])
}

// creatorKey stores creator address and transfer-restriction flag.
func creatorKey(nftID uint64) string {
	var buf [9]byte
	buf[0] = kCreator
	packU64LEInline(nftID, buf[1:])
	return string(buf[:])
}

// editionCountKey stores edition amount.
func editionCountKey(nftID uint64) string {
	var buf [9]byte
	buf[0] = kEdCount
	packU64LEInline(nftID, buf[1:])
	return string(buf[:])
}

// editionOverrideKey stores per-edition owner or burned flag.
func editionOverrideKey(nftID uint64, editionIndex uint32) string {
	var buf [13]byte
	buf[0] = kEdOverride
	packU64LEInline(nftID, buf[1:])
	packU32LEInline(editionIndex, buf[9:])
	return string(buf[:])
}

// ownedIndexKey tracks editions owned by a specific address.
// Uses heap for the owner suffix since length is variable.
func ownedIndexKey(nftID uint64, owner string) string {
	b := make([]byte, 0, 1+8+len(owner))
	b = append(b, kOwnedIdx)
	b = packU64LE(nftID, b)
	b = append(b, owner...)
	return string(b)
}

//
// ===============
// Global Counters
// ===============
//
// These store the latest assigned ID for NFTs and Collections.
//

// getCount retrieves a numeric counter (defaults to 0).
func getNFTCount() uint64 {
	ptr := sdk.StateGetObject("nft_count")
	if ptr == nil || *ptr == "" {
		return 0
	}
	return parseUint64Field(*ptr, 0, len(*ptr))
}

// setCount writes a monotonically increasing counter.
func setNFTCount(n uint64) {
	buf := make([]byte, 0, 20)
	buf = strconv.AppendUint(buf, n, 10)
	sdk.StateSetObject("nft_count", string(buf))
}

// userColCountKey represents the state key for the collection incrementer per user
func userColCountKey(owner string) string { return "uc_" + owner }

// getNextCollectionId returns a new collectionId uint64 for a given user
func getNextCollectionId(owner string) uint64 {
	// Read caller's next collection index
	ptr := sdk.StateGetObject(userColCountKey(owner))
	var idx uint64
	if ptr == nil || *ptr == "" {
		idx = 0
	} else {
		idx = mustParseUint64(*ptr)
	}
	return idx
}

// updateUserCollectionCount increments the collectionId for a given user
func updateUserCollectionCount(count uint64, owner string) {
	nbuf := make([]byte, 0, 20)
	nbuf = strconv.AppendUint(nbuf, count+1, 10)
	sdk.StateSetObject(userColCountKey(owner), string(nbuf))
}

// ======================================
// Fast Numeric Parsing (Zero Allocation)
// ======================================
//
// parseUint64Field parses substring [a:b] into uint64, aborting on first error.
//
//go:inline
func parseUint64Field(s string, a, b int) uint64 {
	var v uint64
	if a >= b {
		sdk.Abort("empty uint64")
	}
	for i := a; i < b; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			sdk.Abort("invalid uint64")
		}
		v = v*10 + uint64(c-'0')
	}
	return v
}

//go:inline
func parseUint32Field(s string, a, b int) uint32 {
	return uint32(parseUint64Field(s, a, b))
}

// ===============================
// Delimiter-Based Field Splitting
// ===============================
//
// These functions are optimized for hot path ABI parsing.
//

// split2 finds the first comma separating two non-empty values: "A,B".
//
//go:inline
func split2(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if i == 0 || i == len(s)-1 {
				break
			}
			return i
		}
	}
	sdk.Abort("invalid payload format A,B")
	return -1
}

//
// ================
// String Utilities
// ================

// indexByte returns the index of c in s or -1 if missing.
//
//go:inline
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// splitOwnerCollection expects "<owner>_<collection>" strictly.
//
//go:inline
func splitOwnerCollection(ownerCollection string) (string, string) {
	idx := indexByte(ownerCollection, '_')
	if idx <= 0 || idx >= len(ownerCollection)-1 {
		sdk.Abort("invalid owner_collection")
	}
	return ownerCollection[:idx], ownerCollection[idx+1:]
}

// stringToUint32 safely parses a decimal string pointer into a uint32 pointer.
func stringToUint32(s *string) *uint32 {
	v := parseUint64Field(*s, 0, len(*s))
	x := uint32(v)
	return &x
}

// mustParseUint64 parses a full string into uint64.
//
//go:inline
func mustParseUint64(s string) uint64 {
	return parseUint64Field(s, 0, len(s))
}

//
// ============================
// Fixed-Size Delimited Parsing
// ============================

// splitFixedPipe splits s into exactly n '|' delimited segments.
func splitFixedPipe(s string, n int) []string {
	parts := make([]string, n)
	start := 0
	idx := 0
	for i := 0; i < len(s) && idx < n-1; i++ {
		if s[i] == '|' {
			parts[idx] = s[start:i]
			idx++
			start = i + 1
		}
	}
	if idx != n-1 {
		sdk.Abort("invalid payload")
	}
	parts[n-1] = s[start:]
	return parts
}

// parse4 splits a string of format "a|b|c|d" into 4 parts.
func parse4(s string) (string, string, string, string) {
	parts := splitFixedPipe(s, 4)
	return parts[0], parts[1], parts[2], parts[3]
}

// Uint32ListToCSV joins uint32 slice into CSV, optimized to reduce allocations.
func Uint32ListToCSV(list []uint32) string {
	if len(list) == 0 {
		return ""
	}
	b := make([]byte, 0, len(list)*11) // digits + comma
	for i, v := range list {
		if i > 0 {
			b = append(b, ',')
		}
		b = strconv.AppendUint(b, uint64(v), 10)
	}
	return string(b)
}

// packU64LE appends uint64 x to byte slice b in little-endian format.
//
//go:inline
func packU64LE(x uint64, b []byte) []byte {
	return append(b, byte(x), byte(x>>8), byte(x>>16), byte(x>>24), byte(x>>32), byte(x>>40), byte(x>>48), byte(x>>56))
}

// csv lookup and remover for market contract managment
// containsInCSV checks if target is in csv string without allocations
func containsInCSV(csv string, target string) bool {
	start := 0
	for i := 0; i <= len(csv); i++ {
		if i == len(csv) || csv[i] == '|' {
			if csv[start:i] == target {
				return true
			}
			start = i + 1
		}
	}
	return false
}

// removeFromCSV removes target from csv (returns new csv without trailing comma)
func removeFromCSV(csv string, target string) string {
	start := 0
	found := false
	b := make([]byte, 0, len(csv))
	for i := 0; i <= len(csv); i++ {
		if i == len(csv) || csv[i] == '|' {
			part := csv[start:i]
			if part == target {
				found = true
			} else {
				if len(b) > 0 {
					b = append(b, '|')
				}
				b = append(b, part...)
			}
			start = i + 1
		}
	}
	if !found {
		sdk.Abort("market not found")
	}
	return string(b)
}
