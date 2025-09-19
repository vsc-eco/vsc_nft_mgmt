package main

import (
	"fmt"
	"vsc_nft_mgmt/sdk"
)

const (
	nftVersion         = 1   // for a possible versioning of the nft contract
	maxMetaKeys        = 25  // maximum count of metadata keys for an nft
	maxMetaKeyLength   = 50  // maaximum length of a key within the metadata
	maxMetaValueLength = 512 // maximum length of a value within the metadata
	maxEditions        = 99  // maximum editions mintable at once
)

// the basic nft object
type NFT struct {
	ID             uint64      `json:"id"`
	Creator        sdk.Address `json:"creator"`
	Owner          sdk.Address `json:"owner"`
	Version        int         `json:"v"`
	CreationTxID   string      `json:"txID"`
	Collection     uint64      `json:"c"`
	SingleTransfer bool        `json:"bound"` // true if the nft can only be transferred once (from creator to another user)
	NFTPrefs       *NFTPrefs   `json:"pref,omitempty"`
	Edition        *NFTEdition `json:"eds,omitempty"`

	// later more "NFT types" are possible like mutables and others
}

// non-unique nfts have additional edition data
type NFTEdition struct {
	EditionNumber  uint64 `json:"no"`
	EditionsTotal  uint64 `json:"t"`
	GenesisEdition uint64 `json:"g"` // only the genesis edition will have nftPrefs to save space
}

// more complex NFT preferences
// only get stored on unique nfts and genesis editions to reduce redundance within contract state
type NFTPrefs struct {
	Description string            `json:"desc"`
	Metadata    map[string]string `json:"meta,omitempty"`
}

type TransferNFTArgs struct {
	NftID      uint64      `json:"id"`
	Collection uint64      `json:"c"`
	Owner      sdk.Address `json:"owner"`
}

type MintNFTArgs struct {
	Collection     uint64            `json:"c"`     // mandatory: collectionId
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // optional: description of the nft
	SingleTransfer bool              `json:"bound"` // optional: true if the nft is non-transferrable
	Metadata       map[string]string `json:"meta"`  // optional: additional metadata for the nft
}

type MintNFTEditionsArgs struct {
	Collection     uint64            `json:"c"`     // mandatory: collectionId
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // optional: description of the nft
	SingleTransfer bool              `json:"bound"` // optional: true if the nft is non-transferrable
	Metadata       map[string]string `json:"meta"`  // optional: additional metadata for the nft
	EditionsTotal  uint64            `json:"et"`    // mandatory: total count of editions
}

// MINT FUNCTIONS

// creation of an unique nft
//
//go:wasmexport nft_mint_unique
func MintNFTUnique(payload *string) *string {
	input := FromJSON[MintNFTArgs](*payload, "minting args")
	collection := loadCollection(input.Collection)
	creator := sdk.GetEnvKey("msg.sender")
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, *creator)

	createAndSaveNFT(
		sdk.Address(*creator),
		sdk.Address(*creator),
		input.Collection,
		input.Description,
		input.SingleTransfer,
		input.Metadata,
		0, 0, 0, // editionNumber, editionsTotal, genesisEditionID

	)

	return nil
}

// creation of an edition nft
// the genesis (first) nft of this series will holds potential "big" but redundant data
// all the following nfts have a property called "genesis" pointing to that genesis nft
//
//go:wasmexport nft_mint_edition
func MintNFTEditions(payload *string) *string {
	input := FromJSON[MintNFTEditionsArgs](*payload, "minting args")
	if input.EditionsTotal > maxEditions {
		sdk.Abort(fmt.Sprintf("%d can be minted at once", maxEditions))
	}
	collection := loadCollection(input.Collection)
	creator := sdk.GetEnvKey("msg.sender")
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, *creator)

	if input.EditionsTotal <= 0 {
		sdk.Abort("editions total <= 0")
	}

	genesisEditionID := newNFTID()
	for editionNumber := 1; editionNumber <= int(input.EditionsTotal); editionNumber++ {

		createAndSaveNFT(
			sdk.Address(*creator),
			sdk.Address(*creator),
			input.Collection,
			input.Description,
			input.SingleTransfer,
			input.Metadata,
			uint64(editionNumber),
			input.EditionsTotal,
			genesisEditionID,
		)

	}
	return nil
}

// TRANSFER FUNCTIONS

// transfers a nft between users and/or collections
// owner can move between collections
// market contract or owner can move to new owner
//
//go:wasmexport nft_transfer
func TransferNFT(payload *string) *string {

	input := FromJSON[TransferNFTArgs](*payload, "transfer args")
	nft := loadNFT(input.NftID)

	// avoid unnecessary non-transfer
	if nft.Owner == input.Owner && nft.Collection == input.Collection {
		sdk.Abort("source and target are the same")
	}

	caller := sdk.GetEnv().Caller // caller instead of sender to enable other contracts
	marketContract := getMarketContract()

	// if the nft should move from one user to another
	if input.Owner != nft.Owner {
		if caller != marketContract && caller != nft.Owner {
			sdk.Abort("only market or owner can transfer")
		}
		if nft.Creator != nft.Owner && nft.SingleTransfer {
			// nft moved once after mint and is non-transferrable
			sdk.Abort("nft bound to owner")
		}
	} else {
		// it is a move between collections of the owner
		if caller != nft.Owner {
			sdk.Abort("only owner can move")
		}
	}
	// check if the target owner owns the target collection
	collection := loadCollection(input.Collection)
	if collection.Owner != input.Owner {
		sdk.Abort("collection not owned by new owner")
	}

	originalCollection := nft.Collection
	nft.Collection = input.Collection
	nft.Owner = input.Owner
	saveNFT(nft)

	// remove the nft from the source collection index
	RemoveIDFromIndex(NFTsCollection+UInt64ToString(originalCollection), nft.ID)
	// add the nft to the target collection index
	AddIDToIndex(NFTsCollection+UInt64ToString(nft.Collection), nft.ID)
	// if it is an edition: remove the nft from the list of available editions (if still in there)
	if nft.Edition != nil {
		RemoveIDFromIndex(AvailEditionsOfGenesis+UInt64ToString(nft.Edition.GenesisEdition), nft.ID)
	}
	sdk.Log("nft transferred to new owner / moved to other collection")
	return nil

}

// GET FUNCTIONS
// returns an nft for a given nft id
//
//go:wasmexport nft_get
func GetNFT(id *string) *string {
	// get one NFT by Id
	nft := loadNFT(StringToUInt64(id))
	jsonStr := ToJSON(nft, "nft")
	sdk.Log(jsonStr)
	return &jsonStr
}

// returns a list of all nfts within a give collection id
//
//go:wasmexport nft_get_collection
func GetNFTsForCollection(collectionId *string) *string {
	// get all NFTs in a collection
	nftIds := GetIDsFromIndex(NFTsCollection + *collectionId)
	jsonStr := getNFTsByIds(nftIds, false)
	sdk.Log(jsonStr)
	return &jsonStr
}

// returns a list of nfts currently owned by a give user
//
//go:wasmexport nft_get_owner
func GetNFTsForOwner(owner *string) *string {
	// get all NFTs owned by a user
	collectionIds := GetIDsFromIndex(CollectionsOwner + *owner)
	nfts := make([]*NFT, 0, len(collectionIds))
	// iterate over all collections of the user
	for _, n := range collectionIds {
		nftIds := GetIDsFromIndex(NFTsCollection + UInt64ToString(n))

		// iterate over all nfts in the collection
		for _, n := range nftIds {
			currentNFT := loadNFT(n)

			nfts = append(nfts, currentNFT)
		}
	}
	jsonStr := ToJSON(nfts, "nfts")
	sdk.Log(jsonStr)
	return &jsonStr
}

// returns a list of all nfts minted by a give user
//
//go:wasmexport nft_get_creator
func GetNFTsForCreator(creator *string) *string {
	// get all NFTs created by a user
	nftIds := GetIDsFromIndex(NFTsCreator + *creator)
	jsonStr := getNFTsByIds(nftIds, true)
	sdk.Log(jsonStr)
	return &jsonStr
}

// returns all editions for a given genesis nft id
//
//go:wasmexport nft_get_editions
func GetEditionsForNFT(id *string) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AllEditionsOfGenesis + *id)
	jsonStr := getNFTsByIds(nftIds, false)
	sdk.Log(jsonStr)
	return &jsonStr
}

// returns a list of nft editions still in creators collection for a given genesis nft id
//
//go:wasmexport nft_get_available
func GetAvailableEditionsForNFT(id *string) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AvailEditionsOfGenesis + *id)
	jsonStr := getNFTsByIds(nftIds, false)
	return &jsonStr
}

func getNFTsByIds(nftIds []uint64, withRelatedEditions bool) string {
	nfts := make([]*NFT, 0, len(nftIds))
	// iterate over all nfts
	for _, n := range nftIds {
		currentNFT := loadNFT(n)
		nfts = append(nfts, currentNFT)
		if withRelatedEditions {
			// iterate over all editions of genesis edition
			editionIds := GetIDsFromIndex(AllEditionsOfGenesis + UInt64ToString(n))
			for _, e := range editionIds {
				currentEdition := loadNFT(e)
				nfts = append(nfts, currentEdition)
			}
		}
	}

	jsonStr := ToJSON(nfts, "nfts")
	sdk.Log(jsonStr)
	return jsonStr
}

// Contract State Interactions

// stores an nft (minded / updated)
func saveNFT(nft *NFT) {
	key := nftKey(nft.ID)
	b := ToJSON(nft, "nft")
	sdk.StateSetObject(key, string(b))

}

// returns a single nft by id
func loadNFT(id uint64) *NFT {
	key := nftKey(id)
	ptr := sdk.StateGetObject(key)
	if ptr == nil || *ptr == "" {
		sdk.Abort(fmt.Sprintf("nft %d not found", id))
	}
	nft := FromJSON[NFT](*ptr, "nft")
	return nft
}

// functions arguments validation
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
		sdk.Abort(fmt.Sprintf("name must between 1 - %d chars", maxNameLength))
		return
	}
	if len(description) > maxDescLength {
		sdk.Abort(fmt.Sprintf("desc max. %d chars", maxDescLength))
	}
	if collectionOwner.String() != caller {
		sdk.Abort("not the owner")
	}

	// check size of the metadata to avoid bloat of the state storage
	if len(metadata) > maxMetaKeys {
		sdk.Abort(fmt.Sprintf("meta max. %d keys", maxMetaKeys))
	}
	for k, v := range metadata {
		if k == "" {
			sdk.Abort("meta keys empty")
		}
		if len(k) > maxMetaKeyLength {
			sdk.Abort(fmt.Sprintf("meta key '%s' > %d chars", k, maxMetaKeyLength))
		}
		if len(v) > maxMetaValueLength {
			sdk.Abort(fmt.Sprintf("meta value for '%s' > %d chars", k, maxMetaValueLength))
		}
	}
}

func createAndSaveNFT(
	creator sdk.Address,
	owner sdk.Address,
	collection uint64,
	description string,
	singleTransfer bool,
	metadata map[string]string,
	editionNumber uint64,
	editionsTotal uint64,
	genesisEditionID uint64,
) *NFT {
	env := sdk.GetEnv()
	nftID := newNFTID()

	var nftEdition *NFTEdition
	var nftPrefs *NFTPrefs

	// if we mint an nft edition
	if editionsTotal > 0 {
		// if we mint the first edition
		if editionNumber == 1 {
			nftID = genesisEditionID
		}
		nftEdition = &NFTEdition{
			EditionNumber:  editionNumber,
			EditionsTotal:  editionsTotal,
			GenesisEdition: genesisEditionID,
		}

	}
	// if it is the first edition or a unique nft
	if editionNumber == 1 || editionsTotal == 0 {
		// Only genesis edition keeps prefs to avoid redundant state storage
		nftPrefs = &NFTPrefs{
			Description: description,
			Metadata:    metadata,
		}
	}

	nft := &NFT{
		ID:             nftID,
		Creator:        creator,
		Owner:          owner,
		Version:        nftVersion,
		CreationTxID:   env.TxId,
		Collection:     collection,
		SingleTransfer: singleTransfer,
		NFTPrefs:       nftPrefs,
		Edition:        nftEdition,
	}
	saveNFT(nft)
	// add nft to collection index
	AddIDToIndex(NFTsCollection+UInt64ToString(nft.Collection), nftID)
	if nft.NFTPrefs != nil {
		// only store unique nfts or genesis editions to creator index
		AddIDToIndex(NFTsCreator+nft.Creator.String(), nftID)
	}
	if nft.Edition != nil {
		// add editions to genesis nft index
		AddIDToIndex(AllEditionsOfGenesis+UInt64ToString(nft.Edition.GenesisEdition), nftID)
		AddIDToIndex(AvailEditionsOfGenesis+UInt64ToString(nft.Edition.GenesisEdition), nftID)
	}
	setCount(NFTsCount, nftID+1)
	return nft
}

func nftKey(nftId uint64) string {
	return fmt.Sprintf("n:%d", nftId)
}

func newNFTID() uint64 {
	return getCount(NFTsCount)
}
