package main

import (
	"encoding/binary"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

// ============ Data Shapes (getter-only JSON)
// We keep this minimal and only for Get* responses.
type NFT struct {
	ID           uint64
	Creator      string
	OwnerCol     string
	CreationTxID string
	Name         string
	Desc         string
	Meta         string
	EdTotal      uint32
}

type NFTResponse struct {
	NFT
	EditionIndex *uint32
	Burned       bool
}

// ============ Exported ABI (string-only) ==========

//go:wasmexport nft_mint
func Mint(payload *string) *string {
	// Payload: "<owner>_<collection>|<name>|<desc>|<singleTransfer>|<editions>|<metadata>"
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	p := *payload

	// split into 6 parts; metadata may be empty but must be present
	parts := splitFixedPipe(p, 6)
	ownerCol := parts[0]
	name := parts[1]
	desc := parts[2]
	single := parts[3] == "true"
	// editions field
	edStart := idxStart(parts, 4)
	edEnd := idxEnd(parts, 4)
	var editions uint32 = 1 // default
	if edEnd > edStart {    // only parse if non-empty
		editions = parseUint32Field(p, edStart, edEnd)
	}
	meta := parts[5] // stored exactly as provided

	if ownerCol == "" {
		sdk.Abort("collection is mandatory")
	}
	loadCollection(ownerCol)          // validate existence via collections index
	validateMintArgs(name, desc, nil) // metadata is opaque

	nftID := newNFTID()
	creator := sdk.GetEnvKey("msg.sender")

	saveNFTCore(nftID, name, desc, meta)
	saveNFTCreator(nftID, *creator, single)
	saveNFTOwnerCollection(nftID, ownerCol)
	if editions > 1 {
		saveNFTEditionCount(nftID, editions)
	}

	EmitMintEvent(nftID, *creator, ownerCol, editions)
	setCount(NFTsCount, nftID+1)
	return nil
}

//go:wasmexport nft_transfer
//go:wasmexport nft_transfer
func Transfer(payload *string) *string {
	// Payload: "<nftID>|<editionIndex>|<owner>_<collection>"
	if payload == nil || *payload == "" {
		sdk.Abort("empty payload")
	}
	parts := splitFixedPipe(*payload, 3)

	// Parse nftID
	id := parseUint64Field(parts[0], 0, len(parts[0]))

	// Parse edition (default = 0 if empty)
	var ed uint32 = 0
	if len(parts[1]) > 0 {
		ed = parseUint32Field(parts[1], 0, len(parts[1]))
	}

	target := parts[2]
	nftEdTotal := loadNFTEditionCount(id)
	nftOwnerCol := loadNFTOwnerCollection(id)

	var effectiveEd uint32
	if *nftEdTotal > 1 {
		effectiveEd = ed
		if effectiveEd >= *nftEdTotal {
			sdk.Abort("edition index out of range")
		}
		resolved := resolveEditionOwnerAndCollection(id, *nftOwnerCol, effectiveEd)
		nftOwnerCol = &resolved
	} else {
		effectiveEd = 0
	}

	if *nftOwnerCol == target {
		sdk.Abort("source and target are the same")
	}

	currentOwner, _ := splitOwnerCollection(*nftOwnerCol, "nftownercol")
	targetOwner, _ := splitOwnerCollection(target, "targetownercol")
	collectionOnly := currentOwner == targetOwner

	caller := sdk.GetEnvKey("msg.caller")
	market := getMarketContract()

	if !collectionOnly {
		if !isAuthorized(caller, &currentOwner, market) {
			sdk.Abort("only market or owner can transfer")
		}
		nftCreator, single := loadNFTCreator(id)
		if single && *nftCreator != currentOwner {
			sdk.Abort("nft bound to owner")
		}
	} else {
		if !isAuthorized(caller, &currentOwner, market) {
			sdk.Abort("only owner/market can change collection")
		}
	}

	if *nftEdTotal > 1 {
		saveEditionOverride(id, effectiveEd, target)
		emitTransfer(id, &effectiveEd, *nftOwnerCol, target)
	} else {
		saveNFTOwnerCollection(id, target)
		emitTransfer(id, nil, *nftOwnerCol, target)
	}

	if *nftEdTotal > 1 && !collectionOnly {
		addEditionToOwnerMapping(id, effectiveEd, targetOwner)
	}
	return nil
}

//go:wasmexport nft_burn
//go:wasmexport nft_burn
func Burn(nftId *string) *string {
	// Payload: "<nftID>" or "<nftID>|<editionIndex>"
	if nftId == nil || *nftId == "" {
		sdk.Abort("empty id")
	}
	p := *nftId
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

	// Load NFT and split owner/collection
	nft := loadNFT(nftID)
	owner, collection := splitOwnerCollection(nft.OwnerCol, "owncol")

	// Authorization check: only owner or market can burn
	caller := sdk.GetEnvKey("msg.caller")
	market := getMarketContract()
	if !isAuthorized(caller, &owner, market) {
		sdk.Abort("only owner or market can burn")
	}

	// Get edition count
	edCount := loadNFTEditionCount(nftID)

	if edPtr != nil {
		if *edCount <= 1 {
			sdk.Abort("NFT has no editions")
		}
		if *edPtr >= *edCount {
			sdk.Abort("edition index out of range")
		}
		// properly mark edition as burned (creates override if missing)
		markEditionBurned(nftID, *edPtr)
	} else {
		// burn entire nft base
		saveNFTOwnerCollection(nftID, owner+"_"+collection)
	}

	emitBurn(nftID, edPtr, owner, collection)
	return nil
}

//go:wasmexport nft_get
func GetNFT(payload *string) *string {
	// Payload: "<nftID>" or "<nftID>|<editionIndex>"
	if payload == nil || *payload == "" {
		sdk.Abort("empty id")
	}
	p := *payload

	// Detect if edition specified (pipe exists)
	idx := indexByte(p, '|')
	var nftID uint64
	var edPtr *uint32

	if idx == -1 {
		// Only NFT ID
		nftID = parseUint64Field(p, 0, len(p))
	} else {
		// Split manually to avoid allocations
		nftID = parseUint64Field(p, 0, idx)
		if idx < len(p)-1 {
			ed := parseUint32Field(p, idx+1, len(p))
			edPtr = &ed
		}
	}

	nft := loadNFT(nftID)

	// Build minimal JSON manually for speed
	b := make([]byte, 0, 128+len(nft.Creator)+len(nft.OwnerCol)+len(nft.Name)+len(nft.Desc)+len(nft.Meta)+len(nft.CreationTxID))
	b = append(b, '{', '"', 'i', 'd', '"', ':')
	b = strconv.AppendUint(b, nft.ID, 10)
	if edPtr != nil && nft.EdTotal > 1 {
		b = append(b, ',', '"', 'e', 'd', '"', ':')
		b = strconv.AppendUint(b, uint64(*edPtr), 10)
	}
	b = append(b, ',', '"', 'c', '"', ':', '"')
	b = append(b, nft.Creator...)
	b = append(b, '"', ',', '"', 'o', 'c', '"', ':', '"')
	// if an edition override exists, surface it in getter
	if edPtr != nil && nft.EdTotal > 1 {
		if eo := loadEditionOverride(nftID, *edPtr); eo != nil {
			nft.OwnerCol = eo.OwnerCollection
		}
	}
	b = append(b, nft.OwnerCol...)
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

//go:wasmexport nft_hasNFTEdition
func GetNFTOwnedEditions(payload *string) *string {
	// Payload: "<nftID>,<ownerAddress>"
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

// ============ Internal state IO ============

func saveNFTCore(nftID uint64, name, desc, meta string) {
	// txID|version|name|desc|meta
	txID := sdk.GetEnvKey("tx.id")
	b := make([]byte, 0, len(*txID)+1+3+1+len(name)+1+len(desc)+1+len(meta))
	b = append(b, (*txID)...)
	b = append(b, '|')
	b = append(b, name...)
	b = append(b, '|')
	b = append(b, desc...)
	b = append(b, '|')
	b = append(b, meta...)
	sdk.StateSetObject(nftCoreKey(nftID), string(b))
}

func loadNFT(nftID uint64) *NFT {
	core := sdk.StateGetObject(nftCoreKey(nftID))
	if core == nil || *core == "" {
		sdk.Abort("nft not found")
	}
	// parse core: tx|v|name|desc|meta
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

func saveNFTCreator(nftID uint64, creator string, singleTransfer bool) {
	val := creator
	if singleTransfer {
		val += "|1"
	} else {
		val += "|0"
	}
	sdk.StateSetObject(creatorKey(nftID), val)
}

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
	sdk.StateSetObject(editionCountKey(nftID), strconv.FormatUint(uint64(cnt), 10))
}

func loadNFTEditionCount(nftID uint64) *uint32 {
	ptr := sdk.StateGetObject(editionCountKey(nftID))
	if ptr == nil || *ptr == "" {
		one := uint32(1) // default is 1 (no editions)
		return &one
	}
	v := stringToUint32(ptr)
	return v
}

type EditionOverride struct {
	OwnerCollection string
	Burned          bool
}

func editionOverrideToStr(eo EditionOverride) string {
	// owner|b (0/1)
	b := make([]byte, 0, len(eo.OwnerCollection)+3)
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
	owner, b := split2Str(s)
	return EditionOverride{OwnerCollection: owner, Burned: b == "1"}
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
		// No override exists yet, so create one and set burned
		baseOwnerCol := *loadNFTOwnerCollection(nftID)
		eo = EditionOverride{
			OwnerCollection: baseOwnerCol,
			Burned:          true,
		}
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

func addEditionToOwnerMapping(nftID uint64, editionIndex uint32, owner string) {
	key := ownedIndexKey(nftID, owner)
	ptr := sdk.StateGetObject(key)
	buf := []byte{}
	if ptr != nil && *ptr != "" {
		buf = []byte(*ptr)
	}
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], editionIndex)
	buf = append(buf, b[:]...)
	sdk.StateSetObject(key, string(buf))
}

func resolveEditionOwnerAndCollection(nftID uint64, base string, editionIndex uint32) string {
	if eo := loadEditionOverride(nftID, editionIndex); eo != nil {
		return eo.OwnerCollection
	}
	return base
}

// ============ Tiny parsing helpers (fast, no alloc) ============

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
