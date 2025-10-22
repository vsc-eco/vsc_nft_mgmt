package main

import (
	"encoding/binary"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

//
// ==================================
// NFT Data Structures (Getter Only)
// ==================================
//
// These structs are not persisted directly in state. They represent
// logical "views" returned by nft_get and are marshaled to JSON manually.
// Keeping them minimal reduces gas and improves clarity.
//

// NFT represents the immutable core + current owner state of a base NFT.
// It does *not* hold edition overrides.
type NFT struct {
	ID           uint64 // unique numeric id of the NFT
	Creator      string // address that created the NFT
	OwnerCol     string // "<owner>_<collection>" for current owner
	CreationTxID string // tx id when the NFT was minted
	Name         string // display name
	Desc         string // short description
	Meta         string // metadata URI or payload (opaque)
	EdTotal      uint32 // total number of editions (1 = no editions)
}

// NFTResponse is a higher-level view exposed via nft_get. It adds
// edition-specific context such as index and burned flag.
type NFTResponse struct {
	NFT
	EditionIndex *uint32 // nil for base, or edition number if specified
	Burned       bool    // indicates if this edition is considered destroyed
}

//
// ==================================
// Public Contract Entry Points (WASM)
// ==================================
// These methods define the NFT lifecycle and are callable by external users.
//

// Mint issues a new NFT under an existing collection.
// Payload format: "<owner>_<collection>|<name>|<desc>|<singleTransfer>|<editions>|<metadata>"
// - singleTransfer="true" means NFT is non-transferable away from creator (soulbound-like)
// - editions defaults to 1 if the field is empty
// After state writes, a mint event is emmited.
//
// Gas tweaks applied:
// - Parse editions directly from parts[4] (no idxStart/idxEnd walk)
// - Read NFT counter once (reuse for setCount)
// - Cache env lookups (creator) in locals
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

	// Create NFT (read counter once; reuse for setCount)
	nftID := newNFTID()
	creatorPtr := sdk.GetEnvKey("msg.sender")
	creator := *creatorPtr

	saveNFTCore(nftID, name, desc, meta)
	saveNFTCreator(nftID, creator, single)
	saveNFTOwnerCollection(nftID, ownerCol)
	if editions > 1 {
		saveNFTEditionCount(nftID, editions)
	}

	EmitMintEvent(nftID, creator, ownerCol, editions)
	setCount(NFTsCount, nftID+1) // reuse same counter value
	return nil
}

// ==================================
// NFT Transfer Logic
// ==================================
//
// Transfer moves ownership of either an entire NFT or a specific edition.
// Payload format: "<nftID>|<editionIndex>|<owner>_<collection>"
// - If editionIndex is empty, it defaults to 0 (base).
// - Moving editions uses override storage; base NFTs update main owner.
// - Edition override is only used if EdTotal > 1.
//
// Micro-optimizations:
// - Cached caller and market values
// - Avoid string mutations on nft.OwnerCol
// - Use direct parse on parts[1] for edition
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
		// Use override state if exists
		resolved := resolveEditionOwnerAndCollection(id, *nftOwnerCol, effectiveEd)
		nftOwnerCol = &resolved
	} else {
		effectiveEd = 0
	}

	// Prevent sending to same owner/collection
	if *nftOwnerCol == target {
		sdk.Abort("source and target are the same")
	}

	// Identify owners (just strings)
	currentOwner, _ := splitOwnerCollection(*nftOwnerCol, "nftownercol")
	targetOwner, _ := splitOwnerCollection(target, "targetownercol")
	collectionOnly := currentOwner == targetOwner

	// Cached env keys
	caller := sdk.GetEnvKey("msg.caller")
	market := getMarketContract()

	// Authorization logic
	if !collectionOnly {
		// actual owner change
		if !isAuthorized(caller, &currentOwner, market) {
			sdk.Abort("only market or owner can transfer")
		}
		creator, single := loadNFTCreator(id)
		if single && *creator != currentOwner {
			sdk.Abort("nft bound to owner")
		}
	} else {
		// collection change only
		if !isAuthorized(caller, &currentOwner, market) {
			sdk.Abort("only owner/market can change collection")
		}
	}

	// Perform state write
	if *nftEdTotal > 1 {
		saveEditionOverride(id, effectiveEd, target)
		emitTransfer(id, &effectiveEd, *nftOwnerCol, target)
	} else {
		saveNFTOwnerCollection(id, target)
		emitTransfer(id, nil, *nftOwnerCol, target)
	}

	// Update owner edition index only on true owner transfer
	if *nftEdTotal > 1 && !collectionOnly {
		addEditionToOwnerMapping(id, effectiveEd, targetOwner)
	}

	return nil
}

// ==================================
// NFT Burn Logic
// ==================================
//
// Payload formats:
//
//	"<nftID>"           → burn data for full NFT
//	"<nftID>|<edition>" → burn a specific edition
//
// When burning full NFT, we reset state to reflect owner, and record burn.
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

	// Parse ID and optional edition index
	if idx == -1 {
		nftID = parseUint64Field(p, 0, len(p))
	} else {
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			ed := parseUint32Field(p, idx+1, len(p))
			edPtr = &ed
		}
	}

	// Load base NFT
	nft := loadNFT(nftID)
	owner, collection := splitOwnerCollection(nft.OwnerCol, "burnOwnerCheck")

	// Cached env
	caller := sdk.GetEnvKey("msg.caller")
	market := getMarketContract()

	// Authorization: only owner or market
	if !isAuthorized(caller, &owner, market) {
		sdk.Abort("only owner or market can burn")
	}

	edCount := loadNFTEditionCount(nftID)

	if edPtr != nil {
		// burn single edition
		if *edCount <= 1 {
			sdk.Abort("NFT has no editions")
		}
		if *edPtr >= *edCount {
			sdk.Abort("edition index out of range")
		}
		markEditionBurned(nftID, *edPtr)
	} else {
		// burn entire NFT → we keep same owner/collection (no concat to prevent alloc)
		saveNFTOwnerCollection(nftID, nft.OwnerCol)
	}

	emitBurn(nftID, edPtr, owner, collection)
	return nil
}

//
// ==========================================
// NFT Query Functions (Public ABI)
// ==========================================

// GetNFT returns a minimal JSON description of either the base NFT
// or a specific edition. Edition overrides (owner/burn) are resolved
// automatically if present.
//
// Payload formats:
//
//	"<id>"
//	"<id>|<editionIndex>"
//
//go:wasmexport nft_get
func GetNFT(payload *string) *string {
	if payload == nil || *payload == "" {
		sdk.Abort("empty id")
	}
	p := *payload
	idx := indexByte(p, '|')

	var nftID uint64
	var edPtr *uint32
	if idx == -1 {
		nftID = parseUint64Field(p, 0, len(p))
	} else {
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			ed := parseUint32Field(p, idx+1, len(p))
			edPtr = &ed
		}
	}

	// Load NFT core data
	nft := loadNFT(nftID)
	ownerCol := nft.OwnerCol // work with a local variable

	// Resolve edition override if applicable
	if edPtr != nil && nft.EdTotal > 1 {
		if eo := loadEditionOverride(nftID, *edPtr); eo != nil {
			ownerCol = eo.OwnerCollection
		}
	}

	// Build minimal JSON manually
	b := make([]byte, 0, 128+len(nft.Creator)+len(ownerCol)+len(nft.Name)+len(nft.Desc)+len(nft.Meta)+len(nft.CreationTxID))
	b = append(b, '{', '"', 'i', 'd', '"', ':')
	b = strconv.AppendUint(b, nft.ID, 10)

	if edPtr != nil && nft.EdTotal > 1 {
		b = append(b, ',', '"', 'e', 'd', '"', ':')
		b = strconv.AppendUint(b, uint64(*edPtr), 10)
	}

	// "c": creator
	b = append(b, ',', '"', 'c', '"', ':', '"')
	b = append(b, nft.Creator...)
	b = append(b, '"', ',', '"', 'o', 'c', '"', ':', '"')
	b = append(b, ownerCol...)
	b = append(b, '"', ',', '"', 't', 'x', '"', ':', '"')
	b = append(b, nft.CreationTxID...)
	b = append(b, '"', ',', '"', 'n', '"', ':', '"')
	b = append(b, nft.Name...)
	b = append(b, '"', ',', '"', 'd', '"', ':', '"')
	b = append(b, nft.Desc...)
	b = append(b, '"', ',', '"', 'm', '"', ':', '"')
	b = append(b, nft.Meta...)
	b = append(b, '"', ',', '"', 'e', '"', ':')
	b = strconv.AppendUint(b, uint64(nft.EdTotal), 10)
	b = append(b, '}')

	json := string(b)
	return &json
}

// GetNFTOwnedEditions returns a CSV list of edition indices for a given owner.
// Payload: "<nftID>,<owner>"
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

//
// =======================================
// Internal State I/O for NFT Core
// =======================================

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

// loadNFT reads core NFT data + creator, ownership, edition count.
func loadNFT(nftID uint64) *NFT {
	core := sdk.StateGetObject(nftCoreKey(nftID))
	if core == nil || *core == "" {
		sdk.Abort("nft not found")
	}
	tx, name, desc, meta := parse4(*core)

	creator, _ := loadNFTCreator(nftID)
	ownerCol := loadNFTOwnerCollection(nftID)
	ed := loadNFTEditionCount(nftID)

	return &NFT{
		ID:           nftID,
		Creator:      *creator,
		OwnerCol:     *ownerCol,
		CreationTxID: tx,
		Name:         name,
		Desc:         desc,
		Meta:         meta,
		EdTotal:      *ed,
	}
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

//
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

//
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

//
// ==========================
// Tiny Parsing Utilities
// ==========================
//
// These are on the hot path and written with zero-alloc intent.

func idxStart(parts []string, i int) int {
	pos := 0
	for k := 0; k < i; k++ {
		pos += len(parts[k]) + 1
	}
	return pos
}

func idxEnd(parts []string, i int) int { return idxStart(parts, i) + len(parts[i]) }

func split2Str(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '|' {
			return s[:i], s[i+1:]
		}
	}
	sdk.Abort("invalid 2-field state")
	return "", ""
}

// ----------------------------------------
// Mint Argument Validation
// ----------------------------------------
//
// validateMintArgs performs structural checks when creating new NFTs
// such as length constraints. Metadata is considered opaque and therefore
// not validated here.
func validateMintArgs(name, desc string) {
	if len(name) == 0 || len(name) > maxNameLength {
		abort("invalid name length")
	}
	if len(desc) > maxDescLength {
		abort("description too long")
	}
}
