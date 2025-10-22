package main

import (
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// =============================
// Binary key prefixes (single byte)
// =============================
const (
	kNFTCore    byte = 0x01 // immutable NFT core data
	kOwner      byte = 0x02 // owner by nftID
	kCreator    byte = 0x03 // creator by nftID
	kEdCount    byte = 0x04 // edition count by nftID
	kEdOverride byte = 0x05 // edition override by (nftID,editionIndex)
	kOwnedIdx   byte = 0x06 // owned index by (nftID,owner)
)

// =============================
// Packing helpers (little-endian)
// =============================
func packU64LE(x uint64, b []byte) []byte {
	return append(b,
		byte(x), byte(x>>8), byte(x>>16), byte(x>>24),
		byte(x>>32), byte(x>>40), byte(x>>48), byte(x>>56),
	)
}

func packU32LE(x uint32, b []byte) []byte {
	return append(b,
		byte(x), byte(x>>8), byte(x>>16), byte(x>>24),
	)
}

// =============================
// Binary key builders (return string; strings can store raw bytes)
// =============================
func nftCoreKey(nftID uint64) string {
	b := make([]byte, 0, 1+8)
	b = append(b, kNFTCore)
	b = packU64LE(nftID, b)
	return string(b)
}

func ownerKey(nftID uint64) string {
	b := make([]byte, 0, 1+8)
	b = append(b, kOwner)
	b = packU64LE(nftID, b)
	return string(b)
}

func creatorKey(nftID uint64) string {
	b := make([]byte, 0, 1+8)
	b = append(b, kCreator)
	b = packU64LE(nftID, b)
	return string(b)
}

func editionCountKey(nftID uint64) string {
	b := make([]byte, 0, 1+8)
	b = append(b, kEdCount)
	b = packU64LE(nftID, b)
	return string(b)
}

func editionOverrideKey(nftID uint64, editionIndex uint32) string {
	b := make([]byte, 0, 1+8+4)
	b = append(b, kEdOverride)
	b = packU64LE(nftID, b)
	b = packU32LE(editionIndex, b)
	return string(b)
}

// Owned index: track which editions of an NFT are owned by address
func ownedIndexKey(nftID uint64, owner string) string {
	b := make([]byte, 0, 1+8+len(owner))
	b = append(b, kOwnedIdx)
	b = packU64LE(nftID, b)
	b = append(b, owner...)
	return string(b)
}

// =============================
// Minimal counters
// =============================
func newNFTID() uint64 { return getCount(NFTsCount) }

func getCount(key string) uint64 {
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		return 0
	}
	return mustParseUint64(*ptr)
}

func setCount(key string, n uint64) { sdk.StateSetObject(key, strconv.FormatUint(n, 10)) }

// =============================
// ABI payload numeric parsers (fast, allocation-free)
// =============================
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

func parseUint32Field(s string, a, b int) uint32 { return uint32(parseUint64Field(s, a, b)) }

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

// =============================
// Shared string helpers
// =============================
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func splitOwnerCollection(ownerCollection, spot string) (string, string) {
	idx := indexByte(ownerCollection, '_')
	if idx <= 0 || idx >= len(ownerCollection)-1 {
		sdk.Abort("[" + spot + "] invalid owner_collection")
	}
	return ownerCollection[:idx], ownerCollection[idx+1:]
}

func stringToUint32(s *string) *uint32 {
	v, _ := strconv.ParseUint(*s, 10, 32)
	x := uint32(v)
	return &x
}

func mustParseUint64(s string) uint64 {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		sdk.Abort("bad uint64")
	}
	return v
}

// splitFixedPipe splits s by '|' into exactly n parts; aborts on mismatch
func splitFixedPipe(s string, n int) []string {
	parts := make([]string, n)
	start, idx, count := 0, 0, 0
	for i := 0; i < len(s) && count < n-1; i++ {
		if s[i] == '|' {
			parts[idx] = s[start:i]
			idx++
			count++
			start = i + 1
		}
	}
	if idx != n-1 {
		sdk.Abort("invalid payload")
	}
	parts[n-1] = s[start:]
	return parts
}

// Uint32ListToCSV builds a comma-separated list with minimal allocations
func Uint32ListToCSV(list []uint32) string {
	if len(list) == 0 {
		return ""
	}
	// Pre-size buffer: max 10 digits per uint32 + 1 comma between items
	b := make([]byte, 0, len(list)*11)
	for i, v := range list {
		if i > 0 {
			b = append(b, ',')
		}
		b = strconv.AppendUint(b, uint64(v), 10)
	}
	return string(b)
}

// parse4 splits state "a|b|c|d" into four parts.
func parse4(s string) (string, string, string, string) {
	fields := [4]string{}
	start, idx := 0, 0
	for i := 0; i < len(s) && idx < 3; i++ {
		if s[i] == '|' {
			fields[idx] = s[start:i]
			idx++
			start = i + 1
		}
	}
	if idx != 3 {
		sdk.Abort("corrupt state")
	}
	fields[3] = s[start:]
	return fields[0], fields[1], fields[2], fields[3]
}
