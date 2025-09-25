package main

import (
	"fmt"
	"strconv"
	"strings"
	"vsc_nft_mgmt/sdk"
)

const (
	nftVersion         = 1   // for a potential versioning of the nft contract
	maxMetaKeys        = 25  // max count of metadata keys for an nft
	maxMetaKeyLength   = 50  // max length of a key within the metadata
	maxMetaValueLength = 512 // max length of a value within the metadata
)

// the basic nft object
type NFT struct {
	ID             uint64            `json:"id"`           // unique id of an nft
	Creator        sdk.Address       `json:"cr"`           // original creator of the nft
	Owner          sdk.Address       `json:"o"`            // current owner of the nft
	CreationTxID   string            `json:"txID"`         // tx when the nft was minted
	Collection     uint64            `json:"c"`            // current collection id the nft is part of
	Name           string            `json:"n"`            // general name of the nft
	Description    string            `json:"d"`            // long description of the nft
	Metadata       map[string]string `json:"m,omitempty"`  // additional metadata like properties, uri and more
	SingleTransfer bool              `json:"b"`            // true if the nft can only be transferred once
	EditionsTotal  uint32            `json:"et,omitempty"` // total number of editions
	Version        int               `json:"v"`            // version of the nft contract this nft was minted with
}

// EditionOverride stores ownership for a specific edition
type EditionOverride struct {
	Owner      sdk.Address `json:"owner"`
	Collection uint64      `json:"collection"`
	Burned     bool        `json:"burned"`
}

type TransferNFTArgs struct {
	NftID      string      `json:"id"` // mandatory: id of the nft or edition to get transferred
	Collection uint64      `json:"c"`  // mandatory: target collection
	Owner      sdk.Address `json:"o"`  // mandatory: target owner
}

type MintNFTArgs struct {
	Collection     *uint64           `json:"c"`     // mandatory: target collection id
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // opt: description
	SingleTransfer bool              `json:"bound"` // opt: non-transferrable (default false)
	Metadata       map[string]string `json:"meta"`  // opt: metadata
	EditionsTotal  uint32            `json:"et"`    // mandatory: number editions to mint
}

type NFTResponse struct {
	*NFT
	EditionIndex *uint32 `json:"editionIndex,omitempty"`
	Burned       bool    `json:"burned,omitempty"`
}

// MINT FUNCTIONS

//go:wasmexport nft_mint
func MintNFT(payload *string) *string {
	input := FromJSON[MintNFTArgs](*payload, "minting args")
	if input.Collection == nil {
		sdk.Abort("collection is mandatory")
	}

	collection := loadCollection(*input.Collection)
	creator := sdk.GetEnvKey("msg.sender")
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, *creator)

	// Ensure EditionsTotal is always at least 1
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
	saveNFT(nft)

	// emit mint event for base NFT
	EmitMintEvent(nftID, *creator, *creator, *input.Collection, et)

	// TODO: maybe emit events for each edition?

	setCount(NFTsCount, nftID+uint64(1))
	return nil
}

// TRANSFER FUNCTIONS

//go:wasmexport nft_transfer
func TransferNFT(payload *string) *string {
	input := FromJSON[TransferNFTArgs](*payload, "transfer args")
	if input.Owner == "" {
		sdk.Abort("owner is mandatory")
	}

	nftID, editionIndex := parseNFTCompositeID(input.NftID)
	nft := loadNFT(nftID)

	// validate edition index
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

	owner, collection := nft.Owner, nft.Collection

	// For multi-editions, resolve overrides
	if nft.EditionsTotal > 1 && ei > 0 {
		owner, collection = resolveEditionOwnerAndCollection(nft, ei)
	}

	// prevent no-op
	if owner == input.Owner && collection == input.Collection {
		sdk.Abort("source and target are the same")
	}

	collectionOnlyChange := owner == input.Owner && collection != input.Collection

	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()
	targetCollection := loadCollection(input.Collection)

	// Owner transfer validation
	if owner != input.Owner {
		if *caller != *marketContract && *caller != owner.String() {
			sdk.Abort("only market or owner can transfer")
		}
		if nft.SingleTransfer && nft.Creator != owner {
			sdk.Abort("nft bound to owner")
		}
		if targetCollection.Owner != input.Owner {
			sdk.Abort("target collection not owned by new owner " + input.Owner.String())
		}
	}

	// Collection-only transfer
	if collectionOnlyChange {
		if *caller != *marketContract && *caller != owner.String() {
			sdk.Abort("only NFT owner or market can change collection")
		}
		if targetCollection.Owner != input.Owner {
			sdk.Abort("target collection not owned by new owner ")
		}
	}

	// Update ownership
	if nft.EditionsTotal > 1 {
		// per-edition override
		saveEditionOverride(nft.ID, ei, input.Owner, input.Collection)

		EmitTransferEvent(
			UInt64ToString(nft.ID)+":"+strconv.FormatInt(int64(ei), 10),
			owner.String(),
			input.Owner.String(),
			nft.Collection,
			input.Collection,
		)

	} else {
		// unique NFT
		nft.Owner = input.Owner
		nft.Collection = input.Collection
		saveNFT(nft)

		EmitTransferEvent(
			UInt64ToString(nft.ID),
			owner.String(),
			nft.Owner.String(),
			collection,
			nft.Collection,
		)
	}

	return nil
}

// BURN FUNCTIONS

//go:wasmexport nft_burn
func Burn(nftId *string) *string {
	nftID, editionIndex := parseNFTCompositeID(*nftId)
	nft := loadNFT(nftID)

	// validate edition index
	if nft.EditionsTotal > 1 && editionIndex == nil {
		tmp := uint32(0)
		editionIndex = &tmp
	}

	// Burn edition
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

		EmitBurnEvent(
			UInt64ToString(nftID)+":"+strconv.FormatInt(int64(ei), 10),
			owner.String(),
			nft.Collection,
		)
		return nil
	}

	// Burn full NFT (unique)

	caller := sdk.GetEnvKey("msg.caller")
	marketContract := getMarketContract()
	if *caller != *marketContract && *caller != nft.Owner.String() {
		sdk.Abort("only owner or market can burn")
	}

	sdk.StateDeleteObject(nftKey(nft.ID))
	EmitBurnEvent(UInt64ToString(nft.ID), nft.Owner.String(), nft.Collection)
	return nil
}

// GET FUNCTIONS

func parseNFTCompositeID(id string) (uint64, *uint32) {
	id = strings.TrimSpace(id)
	if strings.Contains(id, ":") {
		parts := strings.SplitN(id, ":", 2)
		parts[0] = strings.TrimSpace(parts[0])
		parts[1] = strings.TrimSpace(parts[1])
		nftID := StringToUInt64(&parts[0])
		editionIdx64, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			sdk.Abort("invalid edition index")
		}
		ei := uint32(editionIdx64)
		return nftID, &ei
	}
	return StringToUInt64(&id), nil
}

//go:wasmexport nft_get
func GetNFT(id *string) *string {
	nftID, editionIndex := parseNFTCompositeID(*id)
	nft := loadNFT(nftID)

	resp := &NFTResponse{NFT: nft}

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

// CONTRACT STATE INTERACTIONS

func saveNFT(nft *NFT) {
	key := nftKey(nft.ID)
	sdk.StateSetObject(key, ToJSON(nft, "nft"))
}

func loadNFT(id uint64) *NFT {
	key := nftKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("nft %d not found", id))
	}
	return FromJSON[NFT](*ptr, "nft")
}

func saveEditionOverride(nftID uint64, editionIndex uint32, owner sdk.Address, collection uint64) {
	key := editionKey(nftID, editionIndex)
	override := &EditionOverride{
		Owner:      owner,
		Collection: collection,
	}
	sdk.StateSetObject(key, ToJSON(override, "override"))
}

func loadEditionOverride(nftID uint64, editionIndex uint32) *EditionOverride {
	key := editionKey(nftID, editionIndex)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		return nil
	}
	return FromJSON[EditionOverride](*ptr, "override")
}

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

// VALIDATION & HELPERS

func validateMintArgs(
	name string,
	description string,
	metadata map[string]string,
	collectionOwner sdk.Address,
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
	if collectionOwner.String() != caller {
		sdk.Abort("not the owner")
	}
	if len(metadata) > maxMetaKeys {
		sdk.Abort("too many meta keys")
	}
	for k, v := range metadata {
		if k == "" {
			sdk.Abort("meta keys empty")
		}
		if len(k) > maxMetaKeyLength {
			sdk.Abort("one meta key too long")
		}
		if len(v) > maxMetaValueLength {
			sdk.Abort("one meta value too long")
		}
	}
}

func nftKey(nftId uint64) string {
	return "n:" + strconv.FormatUint(nftId, 10)
}

func editionKey(nftID uint64, editionIndex uint32) string {
	return fmt.Sprintf("o:%d:%d", nftID, editionIndex)
}

func newNFTID() uint64 {
	return getCount(NFTsCount)
}
