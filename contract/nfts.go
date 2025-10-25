package main

import (
	"encoding/binary"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// ==========================
// UNIQUE NFTS & EDITION NFTS
// ==========================

//
// ===================================
// Public Contract Entry Points (WASM)
// ===================================
// These methods define the NFT lifecycle and are callable by external users.
//

// Mint issues a new NFT under an existing collection.
// Payload format: "<owner>_<collection>|<name>|<desc>|<singleTransfer>|<editions>|<metadata>"
// - singleTransfer="true" means NFT is non-transferable away from the 2nd owner (minter=1st owner) (soulbound-like)
// - editions defaults to 1 if the field is empty
// After state writes, a mint event is emmited.
//
//go:wasmexport nft_mint
func Mint(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	p := *payload
	parts := splitFixedPipe(p, 6)

	ownerCol := parts[0]
	name := parts[1]
	desc := parts[2]
	single := parts[3] == "true"

	// editions: empty -> 1 (fast path: parse directly from parts[4])
	var editions uint32 = 1
	if edStr := parts[4]; len(edStr) > 0 {
		editions = parseUint32Field(edStr, 0, len(edStr))
	}

	meta := parts[5]
	if ownerCol == "" {
		sdk.Abort("collection is mandatory")
	}

	// Validate arguments and collection existence
	loadCollection(ownerCol) // ensures "<owner>_<collection>" exists
	validateMintArgs(name, desc)

	// Create NFT
	nftID := getNFTCount()
	creatorPtr := sdk.GetEnvKey("msg.sender")
	creator := *creatorPtr

	saveNFTCore(nftID, name, desc, meta)
	saveNFTCreator(nftID, creator, single)
	saveNFTOwnerCollection(nftID, ownerCol)
	if editions > 1 {
		saveNFTEditionCount(nftID, editions)
	}

	EmitMintEvent(nftID, creator, ownerCol, editions)
	setNFTCount(nftID + 1)
	return nil
}

// ==================================
// NFT Transfer Logic
// ==================================
//
// Transfer moves ownership of either an unique NFT or a specific edition.
// Payload format: "<nftID>|<editionIndex>|<owner>_<collection>"
// - If editionIndex is empty, it defaults to 0.
// - Moving editions uses override storage; base NFTs update main owner (extra state key)
// - Edition override is only used if EdTotal > 1.
//
//go:wasmexport nft_transfer
func Transfer(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	parts := splitFixedPipe(*payload, 3)

	// Parse NFT ID (always present)
	id := parseUint64Field(parts[0], 0, len(parts[0]))

	// Parse optional edition, default = 0
	var ed uint32
	if edStr := parts[1]; len(edStr) > 0 {
		ed = parseUint32Field(edStr, 0, len(edStr))
	} else {
		ed = 0
	}
	target := parts[2]

	// Load edition count and base ownership
	nftEdTotal := loadNFTEditionCount(id)
	nftOwnerCol := loadNFTOwnerCollection(id)

	// Determine effective edition ownership
	var effectiveEd uint32
	if *nftEdTotal > 1 {
		effectiveEd = ed
		if effectiveEd >= *nftEdTotal {
			sdk.Abort("edition index out of range")
		}
		// Resolve edition-specific owner if an override exists
		resolved := resolveEditionOwnerAndCollection(id, *nftOwnerCol, effectiveEd)
		nftOwnerCol = &resolved
	} else {
		effectiveEd = 0
	}

	// Burn protection logic
	if *nftEdTotal > 1 {
		if eo := loadEditionOverride(id, effectiveEd); eo != nil && eo.Burned {
			sdk.Abort("edition is burned")
		}
	} else {
		if eo := loadEditionOverride(id, 0); eo != nil && eo.Burned {
			sdk.Abort("nft is burned")
		}
	}

	// Prevent no-op transfer
	if *nftOwnerCol == target {
		sdk.Abort("source and target are the same")
	}

	// Identify current and target owners
	currentOwner, _ := splitOwnerCollection(*nftOwnerCol)
	targetOwner, _ := splitOwnerCollection(target)
	collectionOnly := currentOwner == targetOwner

	caller := sdk.GetEnvKey("msg.caller")
	marketContracts := GetMarketContractsCSV(nil)

	// Authorization logic
	if !collectionOnly {
		if !isAuthorized(caller, &currentOwner, marketContracts) {
			sdk.Abort("only market or owner can transfer")
		}
		creator, single := loadNFTCreator(id)
		if single && *creator != currentOwner {
			sdk.Abort("nft bound to owner")
		}
	} else {
		if !isAuthorized(caller, &currentOwner, marketContracts) {
			sdk.Abort("only owner/market can change collection")
		}
	}

	// Perform state write
	if *nftEdTotal > 1 {
		// edition transfer
		saveEditionOverride(id, effectiveEd, target)
		emitTransfer(id, &effectiveEd, *nftOwnerCol, target)
		// Update owned index only when actual owner changes
		if !collectionOnly {
			addEditionToOwnerMapping(id, effectiveEd, targetOwner)
		}
	} else {
		// single nft transfer
		saveNFTOwnerCollection(id, target)
		emitTransfer(id, nil, *nftOwnerCol, target)
	}

	return nil
}

// ==================================
// NFT Burn Logic
// ==================================
//
// Payload formats:
//
//	"<nftID>"           → burn data for unique NFT
//	"<nftID>|<edition>" → burn a specific edition
//
// When burning a uniue NFT, we reset state to reflect owner, and record burn.
// Burn is logical: does not delete base data to preserve history.
//
//go:wasmexport nft_burn
func Burn(nftId *string) *string {
	if nftId == nil || *nftId == "" {
		sdk.Abort("empty id")
	}
	p := *nftId
	idx := indexByte(p, '|')

	var nftID uint64
	var edPtr *uint32

	// Parse "<id>" or "<id>|<edition>"
	if idx == -1 {
		nftID = parseUint64Field(p, 0, len(p))
	} else {
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			ed := parseUint32Field(p, idx+1, len(p))
			edPtr = &ed
		} else {
			sdk.Abort("invalid edition format")
		}
	}

	// Load only owner+collection (fast lookup, no metadata)
	ownerColPtr := loadNFTOwnerCollection(nftID)
	ownerCol := *ownerColPtr
	owner, collection := splitOwnerCollection(ownerCol)

	// Authorization: only the owner can burn
	caller := sdk.GetEnvKey("msg.caller")
	if !isAuthorized(caller, &owner, nil) {
		sdk.Abort("only owner can burn")
	}

	// Check editions
	edCountPtr := loadNFTEditionCount(nftID)
	edCount := *edCountPtr

	if edPtr != nil {
		// Attempting to burn a specific edition
		if edCount <= 1 {
			sdk.Abort("NFT has no editions")
		}
		if *edPtr >= edCount {
			sdk.Abort("edition index out of range")
		}
		markEditionBurned(nftID, *edPtr)
	} else {
		// Base-level burn request
		if edCount > 1 {
			sdk.Abort("edition required to burn multi-edition NFT")
		}
		// Single-edition NFT; edition index implicitly 0
		markEditionBurned(nftID, 0)
	}

	emitBurn(nftID, edPtr, owner, collection)
	return nil
}

// ===============================
// Internal State I/O for NFT Core
// ===============================

// saveNFTCore stores immutable metadata in compact pipe-delimited format.
func saveNFTCore(nftID uint64, name, desc, meta string) {
	txID := sdk.GetEnvKey("tx.id")
	b := make([]byte, 0, len(*txID)+1+len(name)+1+len(desc)+1+len(meta))
	b = append(b, (*txID)...)
	b = append(b, '|')
	b = append(b, name...)
	b = append(b, '|')
	b = append(b, desc...)
	b = append(b, '|')
	b = append(b, meta...)
	sdk.StateSetObject(nftCoreKey(nftID), string(b))
}

// saveNFTCreator stores creator + singleTransfer flag with minimal allocations.
// Format: "creator|1" for restricted transfer, "creator|0" otherwise.
func saveNFTCreator(nftID uint64, creator string, singleTransfer bool) {
	// Pre-size buffer: creator length + 2 bytes ("|" + flag)
	b := make([]byte, 0, len(creator)+2)
	b = append(b, creator...)
	b = append(b, '|')
	if singleTransfer {
		b = append(b, '1')
	} else {
		b = append(b, '0')
	}
	sdk.StateSetObject(creatorKey(nftID), string(b))
}

// loadNFTCreator returns (creatorAddress, isSingleTransferRestricted).
func loadNFTCreator(nftID uint64) (*string, bool) {
	ptr := sdk.StateGetObject(creatorKey(nftID))
	if ptr == nil || *ptr == "" {
		sdk.Abort("creator missing")
	}
	creator, flag := split2Str(*ptr)
	return &creator, flag == "1"
}

func saveNFTOwnerCollection(nftID uint64, ownerCollection string) {
	sdk.StateSetObject(ownerKey(nftID), ownerCollection)
}

func loadNFTOwnerCollection(nftID uint64) *string {
	ptr := sdk.StateGetObject(ownerKey(nftID))
	if ptr == nil || *ptr == "" {
		sdk.Abort("nft owner missing")
	}
	return ptr
}

func saveNFTEditionCount(nftID uint64, cnt uint32) {
	b := make([]byte, 0, 11)
	b = strconv.AppendUint(b, uint64(cnt), 10)
	sdk.StateSetObject(editionCountKey(nftID), string(b))
}

func loadNFTEditionCount(nftID uint64) *uint32 {
	ptr := sdk.StateGetObject(editionCountKey(nftID))
	if ptr == nil || *ptr == "" {
		one := uint32(1)
		return &one
	}
	return stringToUint32(ptr)
}

// ===========================
// Edition Overrides & Burning
// ===========================

type EditionOverride struct {
	OwnerCollection string
	Burned          bool
}

func editionOverrideToStr(eo EditionOverride) string {
	b := make([]byte, 0, len(eo.OwnerCollection)+2)
	b = append(b, eo.OwnerCollection...)
	b = append(b, '|')
	if eo.Burned {
		b = append(b, '1')
	} else {
		b = append(b, '0')
	}
	return string(b)
}

func parseEditionOverride(s string) EditionOverride {
	owner, f := split2Str(s)
	return EditionOverride{OwnerCollection: owner, Burned: f == "1"}
}

func saveEditionOverride(nftID uint64, editionIndex uint32, ownerCollection string) {
	val := editionOverrideToStr(EditionOverride{OwnerCollection: ownerCollection, Burned: false})
	sdk.StateSetObject(editionOverrideKey(nftID, editionIndex), val)
	addEditionToOwnerMapping(nftID, editionIndex, ownerCollection[:indexByte(ownerCollection, '_')])
}

func markEditionBurned(nftID uint64, editionIndex uint32) {
	key := editionOverrideKey(nftID, editionIndex)
	ptr := sdk.StateGetObject(key)

	var eo EditionOverride
	if ptr == nil || *ptr == "" {
		baseOwnerCol := *loadNFTOwnerCollection(nftID)
		eo = EditionOverride{OwnerCollection: baseOwnerCol, Burned: true}
	} else {
		eo = parseEditionOverride(*ptr)
		eo.Burned = true
	}
	sdk.StateSetObject(key, editionOverrideToStr(eo))
}

func loadEditionOverride(nftID uint64, editionIndex uint32) *EditionOverride {
	ptr := sdk.StateGetObject(editionOverrideKey(nftID, editionIndex))
	if ptr == nil || *ptr == "" {
		return nil
	}
	eo := parseEditionOverride(*ptr)
	return &eo
}

// ================================
// Owner Edition Index Data Mapping
// ================================

func addEditionToOwnerMapping(nftID uint64, editionIndex uint32, owner string) {
	key := ownedIndexKey(nftID, owner)
	ptr := sdk.StateGetObject(key)
	buf := []byte{}
	if ptr != nil && *ptr != "" {
		buf = []byte(*ptr)
	}
	var tmp [4]byte
	binary.BigEndian.PutUint32(tmp[:], editionIndex)
	buf = append(buf, tmp[:]...)
	sdk.StateSetObject(key, string(buf))
}

func resolveEditionOwnerAndCollection(nftID uint64, base string, editionIndex uint32) string {
	if eo := loadEditionOverride(nftID, editionIndex); eo != nil {
		return eo.OwnerCollection
	}
	return base
}

// ======================
// Tiny Parsing Utilities
// ======================
//
// These are on the hot path and written with zero-alloc intent.

func idxStart(parts []string, i int) int {
	pos := 0
	for k := 0; k < i; k++ {
		pos += len(parts[k]) + 1
	}
	return pos
}

// func idxEnd(parts []string, i int) int { return idxStart(parts, i) + len(parts[i]) }

func split2Str(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '|' {
			return s[:i], s[i+1:]
		}
	}
	sdk.Abort("invalid 2-field state")
	return "", ""
}

// ========================
// Mint Argument Validation
// ========================
//
// validateMintArgs performs length constraints checks when creating new NFTs.
func validateMintArgs(name, desc string) {
	if len(name) == 0 || len(name) > maxNameLength {
		sdk.Abort("invalid name length")
	}
	if len(desc) > maxDescLength {
		sdk.Abort("description too long")
	}
}

// ====================
// Ownership Validation
// ====================

// isAuthorized returns true if caller == owner OR caller == one of the markets.
// If both validations fail, returns false.
//
// It’s intentionally small because this is hit in hot code paths like transfer/burn.
func isAuthorized(caller, owner, marketContracts *string) bool {
	if caller == nil {
		return false
	}
	c := *caller
	// Direct owner match (cheapest possible path)
	if owner != nil && c == *owner {
		return true
	}
	// Market matches
	if marketContracts != nil && containsInCSV(*marketContracts, *caller) {
		return true

	}
	return false
}
