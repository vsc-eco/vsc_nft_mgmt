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
	ID             uint64            `json:"id"`           // ID is the unique identifier of the NFT.
	Creator        sdk.Address       `json:"cr"`           // Creator is the original creator of the NFT.
	Owner          sdk.Address       `json:"o"`            // Owner is the current owner of the NFT.
	CreationTxID   string            `json:"txID"`         // CreationTxID is the transaction ID of the minting.
	Collection     uint64            `json:"c"`            // Collection is the ID of the collection this NFT belongs to.
	Name           string            `json:"n"`            // Name is the display name of the NFT.
	Description    string            `json:"d"`            // Description is a longer text description of the NFT.
	Metadata       map[string]string `json:"m,omitempty"`  // Metadata holds optional key/value properties or URIs.
	SingleTransfer bool              `json:"b"`            // SingleTransfer indicates whether the NFT can only be transferred once.
	EditionsTotal  uint32            `json:"et,omitempty"` // EditionsTotal is the number of editions minted for this NFT.
	Version        int               `json:"v"`            // Version is the contract version this NFT was minted with.
}

// EditionOverride holds edition-specific ownership and collection overrides.
type EditionOverride struct {
	Owner      sdk.Address `json:"o"`      // Owner is the current owner of this edition.
	Collection uint64      `json:"col"`    // Collection is the collection this edition belongs to.
	Burned     bool        `json:"burned"` // Burned indicates whether this edition has been burned.
}

// TransferNFTArgs defines the input arguments required to transfer an NFT.
type TransferNFTArgs struct {
	NftID      string      `json:"id"` // NftID is the NFT ID or edition ID to transfer.
	Collection uint64      `json:"c"`  // Collection is the target collection ID.
	Owner      sdk.Address `json:"o"`  // Owner is the target owner address.
}

// MintNFTArgs defines the input arguments required to mint a new NFT.
type MintNFTArgs struct {
	Collection     *uint64           `json:"c"`     // Collection is the target collection ID.
	Name           string            `json:"name"`  // Name is the NFT name.
	Description    string            `json:"desc"`  // Description is optional longer text.
	SingleTransfer bool              `json:"bound"` // SingleTransfer marks the NFT as non-transferable if true.
	Metadata       map[string]string `json:"meta"`  // Metadata holds optional key/value properties.
	EditionsTotal  uint32            `json:"et"`    // EditionsTotal specifies how many editions to mint.
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

// Mint creates and stores a new NFT.
//
//go:wasmexport nft_mint
func Mint(payload *string) *string {
	input := FromJSON[MintNFTArgs](*payload, "minting args")
	if input.Collection == nil {
		sdk.Abort("collection is mandatory")
	}
	creator := sdk.GetEnvKey("msg.sender")
	loadCollection(*creator, strconv.FormatUint(*input.Collection, 10))

	validateMintArgs(input.Name, input.Description, input.Metadata, *creator)

	// EditionsTotal must always be at least 1.
	et := input.EditionsTotal
	if et == 0 {
		et = 1
	}

	nftID := newNFTID()
	nft := &NFT{
		ID:             nftID,
		Creator:        sdk.Address(*creator),
		Owner:          sdk.Address(*creator),
		Version:        nftVersion,
		CreationTxID:   *sdk.GetEnvKey("tx.id"),
		Collection:     *input.Collection,
		Name:           input.Name,
		Description:    input.Description,
		Metadata:       input.Metadata,
		SingleTransfer: input.SingleTransfer,
		EditionsTotal:  et,
	}
	if len(nft.Metadata) == 0 {
		nft.Metadata = nil
	}
	saveNFT(nft)

	// Emit mint event for the base NFT.
	EmitMintEvent(nftID, *creator, *creator, *input.Collection, et)

	// Note: consider emitting events for each edition in the future.

	setCount(NFTsCount, nftID+uint64(1))
	return nil
}

// Transfer transfers an NFT or one of its editions to a new owner.
//
//go:wasmexport nft_transfer
func Transfer(payload *string) *string {
	input := FromJSON[TransferNFTArgs](*payload, "transfer args")
	if input.Owner == "" {
		sdk.Abort("owner is mandatory")
	}

	// Parse NFT ID and optional edition index.
	nftID, editionIndex := parseNFTCompositeID(input.NftID)
	nft := loadNFT(nftID)

	// Validate edition index.
	if editionIndex != nil {
		if nft.EditionsTotal == 1 {
			sdk.Abort("NFT has no editions")
		}
		if *editionIndex >= nft.EditionsTotal {
			sdk.Abort("edition index out of range")
		}
	}

	ei := uint32(0)
	if editionIndex != nil {
		ei = *editionIndex
	}

	// Resolve ownership (may be overridden for editions).
	owner, collection := nft.Owner, nft.Collection
	if nft.EditionsTotal > 1 && editionIndex != nil {
		owner, collection = resolveEditionOwnerAndCollection(nft, ei)
	}

	// Prevent transfers that result in no change.
	if owner == input.Owner && collection == input.Collection {
		sdk.Abort("source and target are the same")
	}

	collectionOnlyChange := owner == input.Owner && collection != input.Collection

	// Load environment.
	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()
	loadCollection(input.Owner.String(), strconv.FormatUint(input.Collection, 10))

	// Validate full owner transfer.
	if owner != input.Owner {
		if !isAuthorized(caller, owner, marketContract) {
			sdk.Abort("only market or owner can transfer")
		}
		if nft.SingleTransfer && nft.Creator != owner {
			sdk.Abort("nft bound to owner")
		}

	}

	// Validate collection-only transfer.
	if collectionOnlyChange {
		if !isAuthorized(caller, owner, marketContract) {
			sdk.Abort("only NFT owner or market can change collection")
		}
	}

	// Update ownership and emit transfer event.
	if nft.EditionsTotal > 1 {
		saveEditionOverride(nft.ID, ei, input.Owner, input.Collection)
		emitTransfer(nft.ID, editionIndex, owner, input.Owner, nft.Collection, input.Collection)
	} else {
		nft.Owner = input.Owner
		nft.Collection = input.Collection
		saveNFT(nft)
		emitTransfer(nft.ID, nil, owner, nft.Owner, collection, nft.Collection)
	}

	return nil
}

// Burn marks an NFT or one of its editions as burned and removes it from state.
//
//go:wasmexport nft_burn
func Burn(nftId *string) *string {
	nftID, editionIndex := parseNFTCompositeID(*nftId)
	nft := loadNFT(nftID)

	// If NFT has editions but no index was provided, default to 0.
	if nft.EditionsTotal > 1 && editionIndex == nil {
		tmp := uint32(0)
		editionIndex = &tmp
	}

	// Burn edition.
	if editionIndex != nil {
		if *editionIndex >= nft.EditionsTotal {
			sdk.Abort("edition index out of range")
		}
		if nft.EditionsTotal == 1 {
			sdk.Abort("NFT has no editions")
		}

		ei := *editionIndex
		owner, _ := resolveEditionOwnerAndCollection(nft, ei)
		caller := sdk.GetEnvKey("msg.caller")

		if *caller != owner.String() {
			sdk.Abort("only owner can burn this edition")
		}

		override := loadEditionOverride(nftID, ei)
		if override == nil {
			override = &EditionOverride{
				Owner:      owner,
				Collection: nft.Collection,
			}
		}
		override.Burned = true
		sdk.StateSetObject(editionKey(nftID, ei), ToJSON(override, "override"))

		emitBurn(nftID, editionIndex, owner, nft.Collection)
		return nil
	}

	// Burn full NFT.
	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()

	if !isAuthorized(caller, nft.Owner, marketContract) {
		sdk.Abort("only owner or market can burn")
	}

	sdk.StateDeleteObject(nftKey(nft.ID))
	emitBurn(nft.ID, nil, nft.Owner, nft.Collection)
	return nil
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
			resp.Owner = override.Owner
			resp.Collection = override.Collection
			resp.Burned = override.Burned
		}
	}

	jsonStr := ToJSON(resp, "nft")
	return &jsonStr
}

// GetNFTOwnedEditionsArgs specifies the arguments to query editions owned by an address.
type GetNFTOwnedEditionsArgs struct {
	NftID   uint64      `json:"id"` // NftID is the base NFT ID.
	Address sdk.Address `json:"a"`  // Address is the owner address to check.
}

// GetNFTOwnedEditions returns the edition indices owned by the given address.
// This enables other contracts to check for the ownership of a "member nft"
//
//go:wasmexport nft_get_ownedEditions
func GetNFTOwnedEditions(payload *string) *string {
	input := FromJSON[GetNFTOwnedEditionsArgs](*payload, "edition owner check args")
	nft := loadNFT(input.NftID)
	if nft.EditionsTotal <= 1 {
		sdk.Abort("no editioned nft")
	}

	ownerKey := nftOwnerKey(input.NftID, input.Address)
	ptr := sdk.StateGetObject(ownerKey)
	if ptr == nil || *ptr == "" {
		return nil
	}
	data := []byte(*ptr)
	n := len(data) / 4
	editions := make([]uint32, 0, n)

	for i := 0; i < n; i++ {
		ei := binary.BigEndian.Uint32(data[i*4 : (i+1)*4])

		// Check edition override for current ownership and burn status.
		override := loadEditionOverride(input.NftID, ei)
		if override != nil {
			if override.Burned || override.Owner != input.Address {
				continue
			}
		} else {
			// Fall back to base NFT ownership.
			if nft.Owner != input.Address {
				continue
			}
		}
		editions = append(editions, ei)
	}

	jsonStr := ToJSON(editions, "editions")
	return &jsonStr
}

// --------------------------------
// CONTRACT STATE INTERACTIONS
// --------------------------------

// saveNFT stores the NFT in state.
func saveNFT(nft *NFT) {
	key := nftKey(nft.ID)
	sdk.StateSetObject(key, ToJSON(nft, "nft"))
}

// loadNFT retrieves an NFT from state by ID.
func loadNFT(id uint64) *NFT {
	key := nftKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("nft %d not found", id))
	}
	return FromJSON[NFT](*ptr, "nft")
}

// saveEditionOverride stores ownership/collection overrides for an edition.
func saveEditionOverride(nftID uint64, editionIndex uint32, owner sdk.Address, collection uint64) {
	key := editionKey(nftID, editionIndex)
	override := &EditionOverride{
		Owner:      owner,
		Collection: collection,
	}
	sdk.StateSetObject(key, ToJSON(override, "override"))

	// Append to owner mapping.
	addEditionToOwnerMapping(nftID, editionIndex, owner)
}

// addEditionToOwnerMapping updates the list of editions mapped to an owner.
func addEditionToOwnerMapping(nftID uint64, editionIndex uint32, owner sdk.Address) {
	ownerKey := nftOwnerKey(nftID, owner)
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
	return FromJSON[EditionOverride](*ptr, "override")
}

// resolveEditionOwnerAndCollection resolves effective ownership and collection for an edition.
func resolveEditionOwnerAndCollection(nft *NFT, editionIndex uint32) (sdk.Address, uint64) {
	override := loadEditionOverride(nft.ID, editionIndex)
	if override != nil {
		if override.Burned {
			sdk.Abort("edition has been burned")
		}
		return override.Owner, override.Collection
	}
	return nft.Owner, nft.Collection
}

// --------------------------------
// VALIDATION HELPERS
// --------------------------------

// validateMintArgs validates the arguments for minting an NFT.
func validateMintArgs(
	name string,
	description string,
	metadata map[string]string,
	caller string,
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
	return "n:" + strconv.FormatUint(nftId, 10)
}

// editionKey returns the state key for an NFT edition.
func editionKey(nftID uint64, editionIndex uint32) string {
	return fmt.Sprintf("o:%s", editionCompositeID(nftID, editionIndex))
}

// editionCompositeID returns a string in the format "nftID:editionIndex".
func editionCompositeID(nftID uint64, ei uint32) string {
	return fmt.Sprintf("%d:%d", nftID, ei)
}

// nftOwnerKey returns the state key mapping an owner to editions of an NFT.
func nftOwnerKey(nftID uint64, owner sdk.Address) string {
	return fmt.Sprintf("no:%d:%s", nftID, owner.String())
}

// newNFTID returns the next NFT ID from the counter.
func newNFTID() uint64 {
	return getCount(NFTsCount)
}

// parseNFTCompositeID parses an ID string into an NFT ID and optional edition index.
func parseNFTCompositeID(id string) (uint64, *uint32) {
	id = strings.TrimSpace(id)

	if i := strings.IndexByte(id, ':'); i != -1 {
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
func isAuthorized(caller *string, owner sdk.Address, marketContract *string) bool {
	if caller == nil {
		return false
	}
	if marketContract != nil && *caller == *marketContract {
		return true
	}
	if *caller == owner.String() {
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
	from, to sdk.Address,
	fromCollection, toCollection uint64,
) {
	var id string
	if editionIndex != nil {
		id = editionCompositeID(nftID, *editionIndex)
	} else {
		id = UInt64ToString(nftID)
	}

	EmitTransferEvent(
		id,
		from.String(),
		to.String(),
		fromCollection,
		toCollection,
	)
}

// emitBurn emits a burn event for an NFT or edition.
func emitBurn(nftID uint64, editionIndex *uint32, owner sdk.Address, collection uint64) {
	var id string
	if editionIndex != nil {
		id = editionCompositeID(nftID, *editionIndex)
	} else {
		id = UInt64ToString(nftID)
	}

	EmitBurnEvent(id, owner.String(), collection)
}
