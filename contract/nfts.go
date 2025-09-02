package main

import (
	"fmt"
	"strconv"
	"vsc_nft_mgmt/sdk"
)

const (
	nftVersion         = 1   // for poentailly upcoming versioning of the nft contract
	maxMetaKeys        = 25  // maximum count of metadata keys for an nft
	maxMetaKeyLength   = 50  // maaximum length of a key within the metadata
	maxMetaValueLength = 512 // maximum length of a value within the metadata
)

// the basic nft object
type NFT struct {
	ID             string      `json:"id"`
	Creator        sdk.Address `json:"creator"`
	Owner          sdk.Address `json:"owner"`
	Version        int         `json:"version"`
	CreationTxID   string      `json:"txID"`
	Collection     string      `json:"col"`
	SingleTransfer bool        `json:"singleTransfer"` // true if the nft can only be transferred once (from creator to another user)
	NFTPrefs       *NFTPrefs   `json:"prefs,omitempty"`
	Edition        *NFTEdition `json:"edition,omitempty"`

	// later more "NFT types" are possible like mutables and others
}

// non-unique nfts have additional edition data
type NFTEdition struct {
	EditionNumber  int64  `json:"no"`
	EditionsTotal  int64  `json:"total"`
	GenesisEdition string `json:"genesis"` // only the genesis edition will have nftPrefs to save space
}

// more complex NFT preferences
// only get stored on unique nfts and genesis editions to reduce redundance within contract state
type NFTPrefs struct {
	Description string            `json:"desc"`
	Metadata    map[string]string `json:"meta,omitempty"`
}

type TransferNFTArgs struct {
	NftID      string      `json:"id"`
	Collection string      `json:"col"`
	Owner      sdk.Address `json:"owner"`
}

type MintNFTArgs struct {
	Collection     string            `json:"col"`   // mandatory: collectionId
	Name           string            `json:"name"`  // mandatory: name of the nft
	Description    string            `json:"desc"`  // optional: description of the nft
	SingleTransfer bool              `json:"bound"` // optional: true if the nft is non-transferrable
	Metadata       map[string]string `json:"meta"`  // optional: additional metadata for the nft
}

type MintNFTEditionsArgs struct {
	Collection     string            `json:"col"`      // mandatory: collectionId
	Name           string            `json:"name"`     // mandatory: name of the nft
	Description    string            `json:"desc"`     // optional: description of the nft
	SingleTransfer bool              `json:"bound"`    // optional: true if the nft is non-transferrable
	Metadata       map[string]string `json:"meta"`     // optional: additional metadata for the nft
	EditionsTotal  int64             `json:"editions"` // mandatory: total count of editions
}

//go:wasmexport nft_mint_edition
func MintNFTEditions(payload *string) *string {
	return mintNFTEditionsImpl(payload, RealSDK{})
}

//go:wasmexport nft_mint_unique
func MintNFTUnique(payload *string) *string {
	return mintNFTUniqueImpl(payload, RealSDK{})
}

//go:wasmexport nft_transfer
func TransferNFT(payload *string) *string {
	return transferNFTImpl(payload, RealSDK{})
}

//go:wasmexport nft_get
func GetNFT(id *string) *string {
	return getNFTImpl(id, RealSDK{})
}

//go:wasmexport nft_get_collection
func GetNFTsForCollection(collectionId *string) *string {
	return getNFTsForCollectionImpl(collectionId, RealSDK{})
}

//go:wasmexport nft_get_owner
func GetNFTsForOwner(owner *string) *string {
	return getNFTsForOwnerImpl(owner, RealSDK{})
}

//go:wasmexport nft_get_creator
func GetNFTsForCreator(creator *string) *string {
	return getNFTsForCreatorImpl(creator, RealSDK{})
}

//go:wasmexport nft_get_editions
func GetEditionsForNFT(id *string) *string {
	return getEditionsForNFTImpl(id, RealSDK{})
}

//go:wasmexport nft_get_available
func GetAvailableEditionsForNFT(id *string) *string {
	return getAvailableEditionsForNFTImpl(id, RealSDK{})
}

// MINT FUNCTIONS

// creation of an unique nft
func mintNFTUniqueImpl(payload *string, chain SDKInterface) *string {
	input := FromJSON[MintNFTArgs](*payload, "minting args")
	collection := loadCollection(input.Collection, chain)
	creator := chain.GetEnv().Sender.Address
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, creator, chain)

	createAndSaveNFT(
		creator,
		creator,
		input.Collection,
		input.Description,
		input.SingleTransfer,
		input.Metadata,
		0, 0, "", // editionNumber, editionsTotal, genesisEditionID
		chain,
	)

	return nil
}

// creation of an edition nft
// the genesis (first) nft of this series will holds potential "big" but redundant data
// all the following nfts have a property called "genesis" pointing to that genesis nft
func mintNFTEditionsImpl(payload *string, chain SDKInterface) *string {
	input := FromJSON[MintNFTEditionsArgs](*payload, "minting args")
	collection := loadCollection(input.Collection, chain)
	creator := chain.GetEnv().Sender.Address
	validateMintArgs(input.Name, input.Description, input.Metadata, collection.Owner, creator, chain)

	if input.EditionsTotal <= 0 {
		chain.Abort("editions total <= 0")
	}

	genesisEditionID := strconv.FormatInt(newNFTID(chain), 10)
	for editionNumber := 1; editionNumber <= int(input.EditionsTotal); editionNumber++ {

		createAndSaveNFT(
			creator,
			creator,
			input.Collection,
			input.Description,
			input.SingleTransfer,
			input.Metadata,
			int64(editionNumber),
			input.EditionsTotal,
			genesisEditionID,
			chain,
		)

	}
	return nil
}

// TRANSFER FUNCTIONS

// transfers a nft between users and/or collections
// owner can move between collections
// market contract or owner can move to new owner
func transferNFTImpl(payload *string, chain SDKInterface) *string {

	input := FromJSON[TransferNFTArgs](*payload, "transfer args")
	nft := loadNFT(input.NftID, chain)

	// avoid unnecessary non-transfer
	if nft.Owner == input.Owner && nft.Collection == input.Collection {
		chain.Abort("source and target are the same")
	}

	caller := chain.GetEnv().Caller // caller instead of sender to enable other contracts
	marketContract := getMarketContract(chain)

	// if the nft should move from one user to another
	if input.Owner != nft.Owner {
		if caller != marketContract && caller != nft.Owner {
			chain.Abort("only market or owner can transfer")
		}
		if nft.Creator != nft.Owner && nft.SingleTransfer {
			// nft moved once after mint and is non-transferrable
			chain.Abort("nft bound to owner")
		}
	} else {
		// it is a move between collections of the owner
		if caller != nft.Owner {
			chain.Abort("only owner can move")
		}
	}
	// check if the target owner owns the target collection
	collection := loadCollection(input.Collection, chain)
	if collection.Owner != input.Owner {
		chain.Abort("collection not owned by new owner")
	}

	originalCollection := nft.Collection
	nft.Collection = input.Collection
	nft.Owner = input.Owner
	saveNFT(nft, chain)

	// remove the nft from the source collection index
	RemoveIDFromIndex(NFTsCollection+originalCollection, nft.ID, chain)
	// add the nft to the target collection index
	AddIDToIndex(NFTsCollection+nft.Collection, nft.ID, chain)
	// if it is an edition: remove the nft from the list of available editions (if still in there)
	if nft.Edition != nil {
		RemoveIDFromIndex(AvailEditionsOfGenesis+nft.Edition.GenesisEdition, nft.ID, chain)
	}
	return nil

}

// GET FUNCTIONS
// returns an nft for a given nft id
func getNFTImpl(id *string, chain SDKInterface) *string {
	// get one NFT by Id
	nft := loadNFT(*id, chain)
	jsonStr := ToJSON(nft, "nft")
	return &jsonStr
}

// returns a list of all nfts within a give collection id
func getNFTsForCollectionImpl(collectionId *string, chain SDKInterface) *string {
	// get all NFTs in a collection
	nftIds := GetIDsFromIndex(NFTsCollection+*collectionId, chain)
	jsonStr := getNFTsByIds(nftIds, false, chain)
	return &jsonStr
}

// returns a list of nfts currently owned by a give user
func getNFTsForOwnerImpl(owner *string, chain SDKInterface) *string {
	// get all NFTs owned by a user
	collectionIds := GetIDsFromIndex(CollectionsOwner+*owner, chain)
	nfts := make([]*NFT, 0, len(collectionIds))
	// iterate over all collections of the user
	for _, n := range collectionIds {
		nftIds := GetIDsFromIndex(NFTsCollection+n, chain)

		// iterate over all nfts in the collection
		for _, n := range nftIds {
			currentNFT := loadNFT(n, chain)

			nfts = append(nfts, currentNFT)
		}
	}
	jsonStr := ToJSON(nfts, "nfts")
	return &jsonStr
}

// returns a list of all nfts minted by a give user
func getNFTsForCreatorImpl(creator *string, chain SDKInterface) *string {
	// get all NFTs created by a user
	nftIds := GetIDsFromIndex(NFTsCreator+*creator, chain)
	jsonStr := getNFTsByIds(nftIds, true, chain)
	return &jsonStr
}

// returns all editions for a given genesis nft id
//

func getEditionsForNFTImpl(id *string, chain SDKInterface) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AllEditionsOfGenesis+*id, chain)
	jsonStr := getNFTsByIds(nftIds, false, chain)
	return &jsonStr
}

// returns a list of nft editions still in creators collection for a given genesis nft id
func getAvailableEditionsForNFTImpl(id *string, chain SDKInterface) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AvailEditionsOfGenesis+*id, chain)
	jsonStr := getNFTsByIds(nftIds, false, chain)
	return &jsonStr
}

func getNFTsByIds(nftIds []string, withRelatedEditions bool, chain SDKInterface) string {
	nfts := make([]*NFT, 0, len(nftIds))
	// iterate over all nfts
	for _, n := range nftIds {
		currentNFT := loadNFT(n, chain)
		nfts = append(nfts, currentNFT)
		if withRelatedEditions {
			// iterate over all editions of genesis edition
			editionIds := GetIDsFromIndex(AllEditionsOfGenesis+n, chain)
			for _, e := range editionIds {
				currentEdition := loadNFT(e, chain)
				nfts = append(nfts, currentEdition)
			}
		}
	}

	jsonStr := ToJSON(nfts, "nfts")
	return jsonStr
}

// Contract State Interactions

// stores an nft (minded / updated)
func saveNFT(nft *NFT, chain SDKInterface) {
	key := nftKey(nft.ID)
	b := ToJSON(nft, "nft")
	chain.StateSetObject(key, string(b))

}

// returns a single nft by id
func loadNFT(id string, chain SDKInterface) *NFT {
	key := nftKey(id)
	ptr := chain.StateGetObject(key)
	if ptr == nil {
		chain.Abort(fmt.Sprintf("nft %s not found", id))
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
	caller sdk.Address,
	chain SDKInterface,
) {
	if name == "" {
		chain.Abort("name is mandatory")
	}
	if len(name) > maxNameLength {
		chain.Abort(fmt.Sprintf("name must between 1 - %d chars", maxNameLength))
		return
	}
	if len(description) > maxDescLength {
		chain.Abort(fmt.Sprintf("desc max. %d chars", maxDescLength))
	}
	if collectionOwner != caller {
		chain.Abort("not the owner")
	}

	// check size of the metadata to avoid bloat of the state storage
	if len(metadata) > maxMetaKeys {
		chain.Abort(fmt.Sprintf("meta max. %d keys", maxMetaKeys))
	}
	for k, v := range metadata {
		if k == "" {
			chain.Abort("meta keys empty")
		}
		if len(k) > maxMetaKeyLength {
			chain.Abort(fmt.Sprintf("meta key '%s' > %d chars", k, maxMetaKeyLength))
		}
		if len(v) > maxMetaValueLength {
			chain.Abort(fmt.Sprintf("meta value for '%s' > %d chars", k, maxMetaValueLength))
		}
	}
}

func createAndSaveNFT(
	creator sdk.Address,
	owner sdk.Address,
	collection string,
	description string,
	singleTransfer bool,
	metadata map[string]string,
	editionNumber int64,
	editionsTotal int64,
	genesisEditionID string,
	chain SDKInterface,
) *NFT {
	env := chain.GetEnv()
	nftID := strconv.FormatInt(newNFTID(chain), 10)

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
	saveNFT(nft, chain)
	// add nft to collection index
	AddIDToIndex(NFTsCollection+nft.Collection, nftID, chain)
	if nft.NFTPrefs != nil {
		// only store unique nfts or genesis editions to creator index
		AddIDToIndex(NFTsCreator+nft.Creator.String(), nftID, chain)
	}
	if nft.Edition != nil {
		// add editions to genesis nft index
		AddIDToIndex(AllEditionsOfGenesis+nft.Edition.GenesisEdition, nftID, chain)
		AddIDToIndex(AvailEditionsOfGenesis+nft.Edition.GenesisEdition, nftID, chain)
	}
	intId, err := strconv.ParseInt(nftID, 10, 64) // base 10, up to 64-bit
	if err != nil {
		sdk.Abort("failed to convert nftId back to int")
	}
	setNFTCount(intId+1, chain)
	return nft
}

func nftKey(nftId string) string {
	return fmt.Sprintf("nft:%s", nftId)
}

func newNFTID(chain SDKInterface) int64 {
	return getCount(NFTsCount, chain)
}

func setNFTCount(nextId int64, chain SDKInterface) {
	setCount(NFTsCount, nextId, chain)
}
