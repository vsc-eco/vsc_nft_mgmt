package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"vsc_nft_mgmt/sdk"
)

const (
	nftVersion         = 1   // nftVersion specifies the contract version used when minting NFTs.
	maxMetaKeys        = 25  // maximum number of metadata keys allowed per NFT.
	maxMetaKeyLength   = 50  // maximum length of a metadata key.
	maxMetaValueLength = 512 // maximum length of a metadata value.
)

// NFT represents the base structure of a non-fungible token.
type NFT struct {
	ID             uint64            `json:"id"`          // ID is the unique identifier of the NFT.
	Creator        string            `json:"cr"`          // Creator is the original creator of the NFT.
	OwnerCol       string            `json:"oc"`          // Owner/Collection of the NFT.
	CreationTxID   string            `json:"txID"`        // CreationTxID is the transaction ID of the minting.
	Name           string            `json:"n"`           // Name is the display name of the NFT.
	Description    string            `json:"d"`           // Description is a longer text description of the NFT.
	Metadata       map[string]string `json:"m,omitempty"` // Metadata holds optional key/value properties or URIs.
	SingleTransfer bool              `json:"b"`           // SingleTransfer indicates whether the NFT can only be transferred once.
	Version        uint8             `json:"v"`           // Version is the contract version this NFT was minted with.
	EditionsTotal  uint32            `json:"et"`          // total count of editions
}

// EditionOverride holds edition-specific ownership and collection overrides.
type EditionOverride struct {
	OwnerCollection string // Owner/Collection of the override
	Burned          bool   // Burned indicates whether this edition has been burned.
}

// TransferNFTArgs defines the input arguments required to transfer an NFT.
type TransferNFTArgs struct {
	NftID                 string // NftID is the NFT ID or edition ID to transfer.
	TargetOwnerCollection string //  the target owner address followed by the target collection
}

// MintNFTArgs defines the input arguments required to mint a new NFT.
type MintNFTArgs struct {
	OwnerCollection string            `json:"c"`     // Collection is the target collection ID.
	Name            string            `json:"name"`  // Name is the NFT name.
	Description     string            `json:"desc"`  // Description is optional longer text.
	SingleTransfer  bool              `json:"bound"` // SingleTransfer marks the NFT as non-transferable if true.
	Metadata        map[string]string `json:"meta"`  // Metadata holds optional key/value properties.
	EditionsTotal   uint32            `json:"et"`    // EditionsTotal specifies how many editions to mint.
}

// NFTResponse extends NFT with edition-specific data for queries.
type NFTResponse struct {
	*NFT
	EditionIndex *uint32 `json:"editionIndex,omitempty"` // EditionIndex is the index of the edition, if applicable.
	Burned       bool    `json:"burned,omitempty"`       // Burned indicates whether the NFT or edition has been burned.
}

// --------------------------------
// MUTATIONS (Mint, Transfer, Burn)
// --------------------------------

func stringToUint32(s *string) *uint32 {
	value, _ := strconv.ParseUint(*s, 10, 32)
	num := uint32(value)
	return &num
}

func StringToMap(input string) map[string]string {
	result := make(map[string]string)

	// Split by comma to get key=value pairs
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		if pair == "" {
			continue // Skip empty entries
		}

		// Split only at the first "=" to avoid issues if value contains "="
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			result[key] = value
		}
	}

	return result
}

func csvToMintNFTArgs(csv *string) *MintNFTArgs {
	csvParts := strings.Split(*csv, "|")
	if len(csvParts) != 6 {
		sdk.Abort("Failed to parse mint payload. (targetOwner/Collection|name|description|singleTransfer|EditionsTotal|AdditionalMeta)")
		return nil
	} else {
		return &MintNFTArgs{
			OwnerCollection: csvParts[0],
			Name:            csvParts[1],
			Description:     csvParts[2],
			SingleTransfer:  csvParts[3] == "true",
			EditionsTotal:   *stringToUint32(&csvParts[4]),
			Metadata:        StringToMap(csvParts[5]),
		}
	}

}

// saveNFT stores base (immutable or never load) data of NFT in state
// creationTx|version|name|desc|meta
func saveNFT(nftID uint64, name string, desc string, meta map[string]string) {
	key := nftKey(nftID)
	content := *sdk.GetEnvKey("tx.id") + "|" + strconv.FormatUint(uint64(nftVersion), 10) + "|" + name + "|" + desc + "|" + MapToCommaSeparated(meta)
	sdk.StateSetObject(key, content)
}

// loadNFT retrieves an NFT from state by ID.
func loadNFT(id uint64) *NFT {
	key := nftKey(id)
	nftPtr := sdk.StateGetObject(key)
	if nftPtr == nil || *nftPtr == "" {
		sdk.Abort(fmt.Sprintf("nft %d not found", id))
	}
	nftValues := strings.Split(*nftPtr, "|")
	version, _ := strconv.ParseUint(nftValues[1], 10, 8)
	nftCreator, singleTransfer := loadNFTCreator(id)
	nftOwnerCollection := loadNFTOwnerCollection(id)
	nftEditionsTotal := loadNFTEditionCount(id)
	nft := &NFT{
		ID:             id,
		Creator:        *nftCreator,
		OwnerCol:       *nftOwnerCollection,
		CreationTxID:   nftValues[0],
		Name:           nftValues[2],
		Description:    nftValues[3],
		Metadata:       StringToMap(nftValues[4]),
		SingleTransfer: singleTransfer,
		Version:        uint8(version),
		EditionsTotal:  *nftEditionsTotal,
	}

	return nft
}

// Mint creates and stores a new NFT.
//
//go:wasmexport nft_mint
func Mint(payload *string) *string {

	input := csvToMintNFTArgs(payload) // targetOwnerCollection|name|description|singleTransfer|EditionsTotal|AdditionalMeta
	if input.OwnerCollection == "" {
		sdk.Abort("collection is mandatory")
	}
	loadCollection(input.OwnerCollection) // check if collection exists
	validateMintArgs(input.Name, input.Description, input.Metadata)

	nftID := newNFTID()

	creator := sdk.GetEnvKey("msg.sender")
	saveNFT(nftID, input.Name, input.Description, input.Metadata)
	saveNFTCreator(nftID, *creator, input.SingleTransfer)
	saveNFTOwnerCollection(nftID, input.OwnerCollection)
	et := input.EditionsTotal
	if et > 1 {
		saveNFTEditionCount(nftID, et)
	}
	// Emit mint event for the base NFT.
	EmitMintEvent(nftID, *creator, input.OwnerCollection, et)

	setCount(NFTsCount, nftID+uint64(1))
	return nil
}

// CsvToTransferArgs parses a pipe-delimited string into TransferNFTArgs (nftId|targetOwner|targetCollection).
func CsvToTransferArgs(csv *string) TransferNFTArgs {
	if csv == nil || *csv == "" {
		sdk.Abort("input CSV is nil or empty")
	}

	parts := strings.Split(*csv, "|")
	if len(parts) != 2 {
		sdk.Abort("invalid CSV format: expected 2 fields (nftId|targetOwner/targetCollection)")
	}
	if parts[0] == "" {
		sdk.Abort("nftId is mandatory")
	}
	if parts[1] == "" {
		sdk.Abort("owner/collection is mandatory")
	}
	return TransferNFTArgs{
		NftID:                 parts[0],
		TargetOwnerCollection: parts[1],
	}
}

// Transfer transfers an NFT or one of its editions to a new owner.
//
//go:wasmexport nft_transfer
func Transfer(payload *string) *string {
	input := CsvToTransferArgs(payload)
	loadCollection(input.TargetOwnerCollection) // check if collection exists
	// Parse NFT ID and optional edition index.
	nftID, editionIndex := parseNFTCompositeID(input.NftID)
	nftEdTotal := loadNFTEditionCount(nftID)
	nftOwnerCollection := loadNFTOwnerCollection(nftID)

	// Validate edition index.
	if editionIndex != nil {
		if *nftEdTotal <= 1 {
			sdk.Abort("NFT has no editions")
		} else {
			if *editionIndex >= *nftEdTotal {
				sdk.Abort("edition index out of range")
			}
		}

	}

	ei := uint32(0)
	if editionIndex != nil {
		ei = *editionIndex
	}

	// Resolve ownership (may be overridden for editions).
	if *nftEdTotal >= 1 && editionIndex != nil {
		nftOwnerCollection = resolveEditionOwnerAndCollection(nftID, *nftOwnerCollection, ei)
	}

	// Prevent transfers that result in no change.
	if *nftOwnerCollection == input.TargetOwnerCollection {
		sdk.Abort("source and target are the same")
	}

	currentOwner := strings.Split(*nftOwnerCollection, "/")[0]
	targetOwner := strings.Split(input.TargetOwnerCollection, "/")[0]
	collectionOnlyChange := currentOwner == targetOwner

	// Load environment.
	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()

	// Validate full owner transfer.
	if !collectionOnlyChange {
		if !isAuthorized(caller, &currentOwner, marketContract) {
			sdk.Abort("only market or owner can transfer")
		}
		nftCreator, singleTransfer := loadNFTCreator(nftID)
		if singleTransfer && *nftCreator != currentOwner {
			sdk.Abort("nft bound to owner")
		}
	}

	// Validate collection-only transfer.
	if collectionOnlyChange {
		if !isAuthorized(caller, &currentOwner, marketContract) {
			sdk.Abort("only NFT owner or market can change collection")
		}
	}

	// Update ownership and emit transfer event.
	if *nftEdTotal > 1 {
		saveEditionOverride(nftID, ei, input.TargetOwnerCollection)
		emitTransfer(nftID, editionIndex, *nftOwnerCollection, input.TargetOwnerCollection)
	} else {
		saveNFTOwnerCollection(nftID, input.TargetOwnerCollection)
		emitTransfer(nftID, nil, *nftOwnerCollection, input.TargetOwnerCollection)
	}

	return nil
}

// Burn marks an NFT or one of its editions as burned and removes it from state.
//
//go:wasmexport nft_burn
func Burn(nftId *string) *string {
	nftID, editionIndex := parseNFTCompositeID(*nftId)
	nft := loadNFT(nftID)
	nftETotal := loadNFTEditionCount(nftID)

	// If NFT has editions but no index was provided, default to 0.
	if *nftETotal > 1 && editionIndex == nil {
		tmp := uint32(0)
		editionIndex = &tmp
	}

	// Burn edition.
	if editionIndex != nil {
		if *editionIndex >= *nftETotal {
			sdk.Abort("edition index out of range")
		}
		if *nftETotal == 1 {
			sdk.Abort("NFT has no editions")
		}

		ei := *editionIndex
		ownerCollection := resolveEditionOwnerAndCollection(nft.ID, nft.OwnerCol, ei)
		ownerCollectionParts := strings.Split(*ownerCollection, "/")
		owner := ownerCollectionParts[0]
		collection := ownerCollectionParts[1]

		caller := sdk.GetEnvKey("msg.caller")

		if *caller != owner {
			sdk.Abort("only owner can burn this edition")
		}

		override := loadEditionOverride(nftID, ei)
		if override == nil {
			override = &EditionOverride{
				OwnerCollection: *ownerCollection,
			}
		}
		override.Burned = true
		eoStr := editionOverrideToCsv(*override)
		sdk.StateSetObject(editionKey(nftID, ei), *eoStr)

		emitBurn(nftID, editionIndex, owner, collection)
		return nil
	}

	owner, collection := parseOwnerCollection(nft.OwnerCol)
	// Burn full NFT.
	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()

	if !isAuthorized(caller, &owner, marketContract) {
		sdk.Abort("only owner or market can burn")
	}

	sdk.StateDeleteObject(nftKey(nft.ID))
	emitBurn(nft.ID, nil, owner, collection)
	return nil
}

func editionOverrideToCsv(eo EditionOverride) *string {
	eoString := eo.OwnerCollection + "|" + strconv.FormatBool(eo.Burned)
	return &eoString
}

// --------------------------------
// GET FUNCTIONS
// --------------------------------

// GetNFT returns an NFT or one of its editions with metadata and ownership details.
//
//go:wasmexport nft_get
func GetNFT(id *string) *string {
	nftID, editionIndex := parseNFTCompositeID(*id)
	nft := loadNFT(nftID)

	resp := &NFTResponse{NFT: nft}

	// Apply edition override if requested.
	if editionIndex != nil && nft.EditionsTotal > 1 {
		ei := *editionIndex
		resp.EditionIndex = &ei

		override := loadEditionOverride(nftID, ei)
		if override != nil {
			resp.OwnerCol = override.OwnerCollection
			resp.Burned = override.Burned
		}
	}

	jsonStr := ToJSON(resp, "nft")
	return &jsonStr
}

// GetNFTOwnedEditionsArgs specifies the arguments to query editions owned by an address.
type GetNFTOwnedEditionsArgs struct {
	NftID   uint64 `json:"id"` // NftID is the base NFT ID.
	Address string `json:"a"`  // Address is the owner address to check.
}

func csvToGetNFTOwnedEditionsArgs(csv *string) *GetNFTOwnedEditionsArgs {
	if csv == nil || *csv == "" {
		sdk.Abort("missing payload nftId|address")
	}
	csvValues := strings.Split(*csv, "|")
	if len(csvValues) != 2 {
		sdk.Abort("invalid payload - expected nftId|address")
	}
	return &GetNFTOwnedEditionsArgs{
		NftID:   StringToUInt64(&csvValues[0]),
		Address: csvValues[1],
	}
}

func Uint32ListToCSV(list []uint32) string {
	if len(list) == 0 {
		return ""
	}

	parts := make([]string, len(list))
	for i, v := range list {
		parts[i] = strconv.FormatUint(uint64(v), 10)
	}
	return strings.Join(parts, ",")
}

// GetNFTOwnedEditions returns the edition indices owned by the given address.
// This enables other contracts to check for the ownership of a "member nft"
//
//go:wasmexport nft_hasNFTEdition
func GetNFTOwnedEditions(payload *string) *string {
	input := csvToGetNFTOwnedEditionsArgs(payload)
	nftEdsTotal := loadNFTEditionCount(input.NftID)
	if *nftEdsTotal <= 1 {
		sdk.Abort("no editioned nft")
	}

	ownerKey := nftEditionsOwnerKey(input.NftID, input.Address)
	ptr := sdk.StateGetObject(ownerKey)

	if ptr == nil || *ptr == "" {
		return nil
	} else {
		editionsList := Uint32ListToCSV(decodeEditionList(*ptr))

		return &editionsList
	}
}

// --------------------------------
// CONTRACT STATE INTERACTIONS
// --------------------------------

func MapToCommaSeparated(meta map[string]string) string {
	var parts []string
	for key, value := range meta {
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, ",")
}

// loadNFT retrieves an NFT from state by ID.
func loadNFTEditionCount(id uint64) *uint32 {
	key := nftEditionCountKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		zero := uint32(1)
		return &zero
	}
	return stringToUint32(ptr)
}

func saveNFTEditionCount(nftID uint64, et uint32) {
	key := nftEditionCountKey(nftID)
	sdk.StateSetObject(key, strconv.FormatUint(uint64(et), 10))
}

// saveEditionOverride stores ownership/collection overrides for an edition.
func saveEditionOverride(nftID uint64, editionIndex uint32, ownerCollection string) {
	key := editionKey(nftID, editionIndex)
	override := &EditionOverride{
		OwnerCollection: ownerCollection,
	}
	sdk.StateSetObject(key, ToJSON(override, "override"))

	owner, _ := parseOwnerCollection(ownerCollection)
	// Append to owner mapping.
	addEditionToOwnerMapping(nftID, editionIndex, owner)
}

func decodeEditionList(data string) []uint32 {
	bytes := []byte(data)
	var editions []uint32

	// Each edition index is 4 bytes
	for i := 0; i+4 <= len(bytes); i += 4 {
		value := binary.BigEndian.Uint32(bytes[i : i+4])
		editions = append(editions, value)
	}
	return editions
}

// addEditionToOwnerMapping updates the list of editions mapped to an owner.
func addEditionToOwnerMapping(nftID uint64, editionIndex uint32, owner string) {

	ownerKey := nftEditionsOwnerKey(nftID, owner)
	ptr := sdk.StateGetObject(ownerKey)

	buf := []byte{}
	if ptr != nil && *ptr != "" {
		buf = []byte(*ptr)
	}

	var b [4]byte
	binary.BigEndian.PutUint32(b[:], editionIndex)
	buf = append(buf, b[:]...)

	sdk.StateSetObject(ownerKey, string(buf))
}

// loadEditionOverride retrieves an edition override from state.
func loadEditionOverride(nftID uint64, editionIndex uint32) *EditionOverride {
	key := editionKey(nftID, editionIndex)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		return nil
	}
	parts := strings.Split(*ptr, "|")
	if len(parts) != 2 {
		sdk.Abort("invalid edition override found")
	}
	return &EditionOverride{
		OwnerCollection: parts[0],
		Burned:          parts[0] == "true",
	}
}

// resolveEditionOwnerAndCollection resolves effective ownership and collection for an edition.
func resolveEditionOwnerAndCollection(nftID uint64, nftOwnerCollection string, editionIndex uint32) *string {
	override := loadEditionOverride(nftID, editionIndex)
	if override != nil {
		if override.Burned {
			sdk.Abort("edition has been burned")
		}
		return &override.OwnerCollection
	}
	return &nftOwnerCollection
}

// --------------------------------
// VALIDATION HELPERS
// --------------------------------

// validateMintArgs validates the arguments for minting an NFT.
func validateMintArgs(
	name string,
	description string,
	metadata map[string]string,
) {
	if name == "" {
		sdk.Abort("name is mandatory")
	}
	if len(name) > maxNameLength {
		sdk.Abort("name too long")
	}
	if len(description) > maxDescLength {
		sdk.Abort("description too long")
	}

	if len(metadata) > maxMetaKeys {
		sdk.Abort("too many meta keys")
	}
	for k, v := range metadata {
		if k == "" {
			sdk.Abort("meta key is empty")
		}
		if len(k) > maxMetaKeyLength {
			sdk.Abort("metadata key too long")
		}
		if len(v) > maxMetaValueLength {
			sdk.Abort("metadata value too long")
		}
	}
}

// --------------------------------
// KEY HELPERS
// --------------------------------

// nftKey returns the state key for an NFT ID.
func nftKey(nftId uint64) string {
	return "n|b|" + strconv.FormatUint(nftId, 10)
}
func nftEditionCountKey(nftId uint64) string {
	return "n|cnt|" + strconv.FormatUint(nftId, 10)
}

// editionKey returns the state key for an NFT edition.
func editionKey(nftID uint64, editionIndex uint32) string {
	return fmt.Sprintf("o|%s", editionCompositeID(nftID, editionIndex))
}

// editionCompositeID returns a string in the format "nftID:editionIndex".
func editionCompositeID(nftID uint64, ei uint32) string {
	return fmt.Sprintf("%d:%d", nftID, ei)
}

func nftEditionsOwnerKey(nftID uint64, owner string) string {
	return fmt.Sprintf("no|%d:%s", nftID, owner)
}

func newNFTID() uint64 {
	return getCount(NFTsCount)
}

func nftOwningKey(nftID uint64) string {
	return "n|o|" + strconv.FormatUint(nftID, 10)
}
func nftCreatorKey(nftID uint64) string {
	return "n|c|" + strconv.FormatUint(nftID, 10)
}

func saveNFTOwnerCollection(nftID uint64, ownerCollection string) {
	key := nftOwningKey(nftID)
	sdk.StateSetObject(key, ownerCollection)
}
func saveNFTCreator(nftID uint64, creator string, singleTransfer bool) {
	key := nftCreatorKey(nftID)
	sdk.StateSetObject(key, creator+"|"+strconv.FormatBool(singleTransfer))
}

func loadNFTOwnerCollection(nftID uint64) *string {
	key := nftOwningKey(nftID)
	ptr := sdk.StateGetObject(key)
	return ptr
}

func loadNFTCreator(nftID uint64) (*string, bool) {
	key := nftOwningKey(nftID)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort("mission creator value")
		return nil, false
	} else {
		stateValues := strings.Split(*ptr, "|")
		return &stateValues[0], stateValues[0] == "true"
	}

}

func parseOwnerCollection(ownerCollection string) (string, string) {

	parts := strings.Split(ownerCollection, "/")
	if len(parts) != 2 {
		sdk.Abort("invalid ownerCollection format: expected 2 fields (Owner|Collection)")
	}
	return parts[0], parts[1]
}

// parseNFTCompositeID parses an ID string into an NFT ID and optional edition index.
func parseNFTCompositeID(id string) (uint64, *uint32) {
	id = strings.TrimSpace(id)

	if i := strings.IndexByte(id, '/'); i != -1 {
		base := strings.TrimSpace(id[:i])
		editionStr := strings.TrimSpace(id[i+1:])

		nftID := StringToUInt64(&base)
		editionIdx64, err := strconv.ParseUint(editionStr, 10, 32)
		if err != nil {
			sdk.Abort("invalid edition index")
		}
		ei := uint32(editionIdx64)
		return nftID, &ei
	}

	return StringToUInt64(&id), nil
}

// isAuthorized checks whether the caller is either the market contract or the NFT owner.
func isAuthorized(caller *string, owner *string, marketContract *string) bool {
	if caller == nil {
		return false
	}
	if marketContract != nil && *caller == *marketContract {
		return true
	}
	if *caller == *owner {
		return true
	}
	return false
}

// --------------------------------
// EVENT HELPERS
// --------------------------------

// emitTransfer emits a transfer event for an NFT or edition.
func emitTransfer(
	nftID uint64,
	editionIndex *uint32,
	fromOwnerCollection string,
	toOwnerCollection string,
) {
	var id string
	if editionIndex != nil {
		id = editionCompositeID(nftID, *editionIndex)
	} else {
		id = UInt64ToString(nftID)
	}
	fromValues := strings.Split(fromOwnerCollection, "/")
	toValues := strings.Split(toOwnerCollection, "/")

	EmitTransferEvent(
		id,
		fromValues[0],
		toValues[0],
		fromValues[1],
		toValues[1],
	)
}

// emitBurn emits a burn event for an NFT or edition.
func emitBurn(nftID uint64, editionIndex *uint32, owner string, collection string) {
	var id string
	if editionIndex != nil {
		id = editionCompositeID(nftID, *editionIndex)
	} else {
		id = UInt64ToString(nftID)
	}
	EmitBurnEvent(id, owner, collection)
}
