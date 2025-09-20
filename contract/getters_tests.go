//go:build test
// +build test

package main

// COLLECTIONS

// returns all collection ids for a given owner
// TODO: remove for production code (indexes should be querried via API calls)
//
//go:wasmexport col_get_user
func GetCollectionIdsForOwner(owner *string) *string {
	collectionIds := GetIDsFromIndex(CollectionsOwner + *owner)
	return UInt64ArrayToJsonString(collectionIds, "collection ids")
}

// NFTS

// returns a list of all nfts within a give collection id
// TODO: remove for production code as this could result in huge gas costs (indexes should be querried via API calls)
//
//go:wasmexport nft_get_collection
func GetNFTIdsForCollection(collectionId *string) *string {
	// get all NFTs in a collection
	nftIds := GetIDsFromIndex(NFTsCollection + *collectionId)
	return UInt64ArrayToJsonString(nftIds, "nft ids")
}

// returns a list of all nfts minted by a give user (only unique or genesis)
// TODO: remove for production code as this could result in huge gas costs (indexes should be querried via API calls)
//
//go:wasmexport nft_get_creator
func GetNFTIdsForCreator(creator *string) *string {
	// get all NFTs created by a user
	nftIds := GetIDsFromIndex(NFTsCreator + *creator)
	return UInt64ArrayToJsonString(nftIds, "nft ids")
}

// returns all editions for a given genesis nft id
// TODO: remove for production code as this could result in huge gas costs (indexes should be querried via API calls)
//
//go:wasmexport nft_get_editions
func GetEditionsForNFT(id *string) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AllEditionsOfGenesis + *id)
	return UInt64ArrayToJsonString(nftIds, "nft ids")
}

// returns a list of nfts still in creators collection for a given genesis nft id
// TODO: remove for production code as this could result in huge gas costs (indexes should be querried via API calls)
//
//go:wasmexport nft_get_availableList
func GetAvailableEditionsForNFT(id *string) *string {
	// get all NFT editions related to the genesis NFT
	nftIds := GetIDsFromIndex(AvailEditionsOfGenesis + *id)
	return UInt64ArrayToJsonString(nftIds, "nft ids")
}
