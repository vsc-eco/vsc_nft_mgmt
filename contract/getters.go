package main

import (
	"encoding/binary"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

//
// =================================
// COLLECTION & NFT GETTER FUNCTIONS
// =================================
//
// These functions provide read-only access to on-chain collection and NFT
// metadata. They are intentionally minimal for low gas usage and
// are designed to be only used by other (market) contracts.
//

// GetCollection retrieves metadata for a user-owned collection.
//
// Payload format: "<owner>_<collectionIndex>"
// - The owner and numeric collection index together form a unique identifier.
// - The collection core data is stored directly under this computed key.
//
// Returns:
// <owner>|<col>|<tx>|<name>|<desc>|<meta>
//
//go:wasmexport col_get
func GetCollection(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	ownerCol := *payload
	owner, col := splitOwnerCollection(ownerCol)

	// Fetch stored core record (format: tx|name|desc|meta)
	colDataPtr := sdk.StateGetObject(colIndexKey(owner, col))
	if colDataPtr == nil || *colDataPtr == "" {
		sdk.Abort("collection not found")
	}
	colData := *colDataPtr

	// Build output: <owner>|<col>|<tx>|<name>|<desc>|<meta>
	b := make([]byte, 0, len(owner)+len(col)+len(colData)+2)
	b = append(b, owner...)
	b = append(b, '|')
	b = append(b, col...)
	b = append(b, '|')
	b = append(b, colData...)

	result := string(b)
	return &result
}

// GetCollectionCount returns the total number of collections created by a user.
//
// Payload: "<ownerAddress>"
// Returns: decimal string, e.g. "3"
//
//go:wasmexport col_count
func GetCollectionCount(owner *string) *string {
	if owner == nil || *owner == "" {
		sdk.Abort("empty owner")
	}
	ptr := sdk.StateGetObject(userColCountKey(*owner))
	if ptr == nil || *ptr == "" {
		empty := "0"
		return &empty
	}
	return ptr
}

// CollectionExists checks whether a given owner/index pair represents an existing collection.
//
// Payload: "<owner>_<collectionIndex>"
// Returns: "true" or "false"
//
//go:wasmexport col_exists
func CollectionExists(ownerIndex *string) *string {
	if ownerIndex == nil || *ownerIndex == "" {
		sdk.Abort("empty payload")
	}
	owner, col := splitOwnerCollection(*ownerIndex)
	ptr := sdk.StateGetObject(colIndexKey(owner, col))
	if ptr == nil || *ptr == "" {
		f := "false"
		return &f
	}
	t := "true"
	return &t
}

// GetNFT returns metadata and current ownership state for an NFT or edition.
//
// **Payload formats**
//
//	"<id>"           → base NFT
//	"<id>|<edition>" → specific edition
//
// If an edition override exists (due to transfer or burn), its data is surfaced.
// Output is minimal |-delimited for gas efficiency.
// Returns:
// <nftID>|<editionIndex or empty>|<creator>|<owner_col>|<tx>|<name>|<desc>|<meta>|<edTotal>
//
//go:wasmexport nft_get
func GetNFT(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty id")
	}
	p := *payload

	// Detect edition separator
	idx := indexByte(p, '|')
	var nftID uint64
	var edStr string
	var ed uint32
	var hasEdition bool

	if idx == -1 {
		// no edition provided
		nftID = parseUint64Field(p, 0, len(p))
	} else {
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			edStr = p[idx+1:]
			ed = parseUint32Field(edStr, 0, len(edStr))
			hasEdition = true
		} else {
			sdk.Abort("invalid edition format")
		}
	}

	// Load NFT core data
	corePtr := sdk.StateGetObject(nftCoreKey(nftID))
	if corePtr == nil || *corePtr == "" {
		sdk.Abort("nft not found")
	}
	tx, name, desc, meta := parse4(*corePtr)

	creatorPtr, _ := loadNFTCreator(nftID)
	ownerColPtr := loadNFTOwnerCollection(nftID)
	edCountPtr := loadNFTEditionCount(nftID)

	ownerCol := *ownerColPtr
	edTotal := *edCountPtr

	// Resolve edition override
	if hasEdition {
		if ed >= edTotal {
			sdk.Abort("edition index out of range")
		}
		if edTotal > 1 {
			if eo := loadEditionOverride(nftID, ed); eo != nil {
				ownerCol = eo.OwnerCollection
			}
		} else {
			sdk.Abort("NFT has no editions")
		}
	} else {
		// Default edition for 1-edition NFTs
		if edTotal > 1 {
			// Multi-edition NFT *must* provide edition
			sdk.Abort("edition required for multi-edition NFT")
		}
		edStr = "0" // implied edition
	}

	// Build response: id|edition|creator|owner|tx|name|desc|meta|edTotal
	b := make([]byte, 0, len(*creatorPtr)+len(ownerCol)+len(tx)+len(name)+len(desc)+len(meta)+32)
	b = strconv.AppendUint(b, nftID, 10)
	b = append(b, '|')
	b = append(b, edStr...)
	b = append(b, '|')
	b = append(b, (*creatorPtr)...)
	b = append(b, '|')
	b = append(b, ownerCol...)
	b = append(b, '|')
	b = append(b, tx...)
	b = append(b, '|')
	b = append(b, name...)
	b = append(b, '|')
	b = append(b, desc...)
	b = append(b, '|')
	b = append(b, meta...)
	b = append(b, '|')
	b = strconv.AppendUint(b, uint64(edTotal), 10)

	result := string(b)
	return &result
}

// GetNFTOwnedEditions returns a CSV list (e.g. "0,1,2") of editions
// owned by a specified address.
//
// Payload: "<nftID>,<ownerAddress>"
// If no editions are owned, returns an empty string.
//
//go:wasmexport nft_hasNFTEdition
func GetNFTOwnedEditions(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	p := *payload
	comma := split2(p)
	id := parseUint64Field(p, 0, comma)
	owner := p[comma+1:]

	key := ownedIndexKey(id, owner)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		empty := ""
		return &empty
	}

	buf := []byte(*ptr)
	N := len(buf) / 4
	editions := make([]uint32, 0, N)
	for i := 0; i+3 < len(buf); i += 4 {
		editions = append(editions, binary.BigEndian.Uint32(buf[i:i+4]))
	}
	csv := Uint32ListToCSV(editions)
	return &csv
}

// IsOwner returns "true" if msg.sender owns the NFT or its edition.
//
// Payload formats:
//
//	"<id>"
//	"<id>|<editionIndex>"
//
// For multi-edition NFTs, edition must be explicitly provided.
//
//go:wasmexport nft_isOwner
func IsOwner(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty id")
	}
	p := *payload
	idx := indexByte(p, '|')

	var nftID uint64
	var ed uint32
	hasEdition := false

	// Parse payload format
	if idx == -1 {
		nftID = parseUint64Field(p, 0, len(p))
	} else {
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			ed = parseUint32Field(p, idx+1, len(p))
			hasEdition = true
		} else {
			sdk.Abort("invalid edition format")
		}
	}

	// Load base ownership
	ownerColPtr := loadNFTOwnerCollection(nftID)
	ownerCol := *ownerColPtr
	edTotal := *loadNFTEditionCount(nftID)

	// Resolve edition-specific ownership
	if hasEdition {
		if ed >= edTotal {
			sdk.Abort("edition index out of range")
		}
		if edTotal > 1 {
			if eo := loadEditionOverride(nftID, ed); eo != nil {
				ownerCol = eo.OwnerCollection
			}
		} else {
			sdk.Abort("NFT has no editions")
		}
	} else if edTotal > 1 {
		sdk.Abort("edition required for multi-edition NFT")
	}

	curOwner, _ := splitOwnerCollection(ownerCol)
	caller := sdk.GetEnvKey("msg.sender")

	if caller != nil && *caller == curOwner {
		t := "true"
		return &t
	}
	f := "false"
	return &f
}

// GetNFTOwnerCollectionOf returns the full "<owner>_<collection>" string
// for either the base NFT or a specific edition.
//
// Payload formats:
//
//	"<id>"
//	"<id>|<editionIndex>"
//
// This function is used by external contracts or marketplace frontends
// to determine current routing of an NFT.
//
//go:wasmexport nft_ownerColOf
func GetNFTOwnerCollectionOf(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	p := *payload
	idx := indexByte(p, '|')

	var nftID uint64
	var ed uint32
	hasEdition := false

	// Parse payload
	if idx == -1 {
		nftID = parseUint64Field(p, 0, len(p))
	} else {
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			ed = parseUint32Field(p, idx+1, len(p))
			hasEdition = true
		} else {
			sdk.Abort("invalid edition index format")
		}
	}

	edTotal := *loadNFTEditionCount(nftID)
	ownerCol := *loadNFTOwnerCollection(nftID)

	// If specific edition requested, resolve override
	if hasEdition {
		if ed >= edTotal {
			sdk.Abort("edition index out of range")
		}
		if eo := loadEditionOverride(nftID, ed); eo != nil {
			ownerCol = eo.OwnerCollection
		}
	}

	return &ownerCol
}

// GetNFTCreatorRaw returns the creator address for an NFT.
//
// Payload: "<id>"
// Output: "<creatorAddress>"
//
//go:wasmexport nft_creator
func GetNFTCreatorRaw(id *string) *string {
	if id == nil || *id == "" {
		sdk.Abort("empty id")
	}
	nftID := parseUint64Field(*id, 0, len(*id))
	creator, _ := loadNFTCreator(nftID)
	return creator
}

// GetNFTMeta returns only the metadata string for an NFT.
//
// Payload: "<id>"
// Output: metadata string from core state
//
// Metadata is stored as-is from mint; opaque for this contract.
//
//go:wasmexport nft_meta
func GetNFTMeta(id *string) *string {
	if id == nil || *id == "" {
		sdk.Abort("empty id")
	}
	nftID := parseUint64Field(*id, 0, len(*id))

	core := sdk.StateGetObject(nftCoreKey(nftID))
	if core == nil || *core == "" {
		sdk.Abort("nft not found")
	}
	_, _, _, meta := parse4(*core)
	return &meta
}

// GetNFTSupply returns the total edition count for a given NFT.
//
// Payload: "<id>"
// Returns: "<numEditions>" as string
// - "1" means NFT has no editions.
// - Values >1 indicate multiple distinct edition copies.
// Note: This function does not calculate current supply - only the total supply specified on mint.
//
//go:wasmexport nft_supply
func GetNFTSupply(id *string) *string {
	if id == nil || *id == "" {
		sdk.Abort("empty id")
	}
	nftID := parseUint64Field(*id, 0, len(*id))
	countPtr := loadNFTEditionCount(nftID)

	b := make([]byte, 0, 10)
	b = strconv.AppendUint(b, uint64(*countPtr), 10)
	s := string(b)
	return &s
}

// GetNFTBurnState returns whether a specific edition is burned.
//
// Payload formats:
//
//	"<id>"           → for single-edition NFTs only (implicitly edition 0)
//	"<id>|<edition>" → explicitly checks a given edition
//
// Returns: "true" or "false"
//
//go:wasmexport nft_isBurned
func GetNFTBurnState(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty id")
	}
	p := *payload
	idx := indexByte(p, '|')

	var nftID uint64
	var ed uint32
	hasEdition := false

	// Parse input
	if idx == -1 {
		nftID = parseUint64Field(p, 0, len(p))
		edTotal := *loadNFTEditionCount(nftID)
		if edTotal > 1 {
			sdk.Abort("edition required to check burn state for multi-edition NFT")
		}
		ed = 0
		hasEdition = true
	} else {
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			ed = parseUint32Field(p, idx+1, len(p))
			hasEdition = true
		} else {
			sdk.Abort("invalid edition index format")
		}
	}

	if !hasEdition {
		sdk.Abort("missing edition index")
	}

	// Validate edition range
	edTotal := *loadNFTEditionCount(nftID)
	if ed >= edTotal {
		sdk.Abort("edition index out of range")
	}

	// Check burn flag from edition overrides
	if eo := loadEditionOverride(nftID, ed); eo != nil && eo.Burned {
		t := "true"
		return &t
	}

	f := "false"
	return &f
}

// IsNFTSingleTransfer returns whether an NFT is soulbound (transfer-restricted).
//
// Payload: "<id>"
// Output: "true" or "false"
// - "true" means NFT can only be held by one owner != creator (singleTransfer=true).
// Note: Single Transfer still enables to minter to transfer these nfts once.
//
//go:wasmexport nft_isSingleTransfer
func IsNFTSingleTransfer(id *string) *string {
	if id == nil || *id == "" {
		sdk.Abort("empty id")
	}
	nftID := parseUint64Field(*id, 0, len(*id))
	_, single := loadNFTCreator(nftID)
	if single {
		t := "true"
		return &t
	}
	f := "false"
	return &f
}
