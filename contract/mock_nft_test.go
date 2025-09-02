//go:build test
// +build test

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"vsc_nft_mgmt/sdk"
)

var collectionPayload string = `{"name": "MyCollection", "desc": "A test collection"}`

func createDummyCollection(f *FakeSDK) string {
	// Prepare Collection
	createCollectionImpl(&collectionPayload, f)
	createdCollectionId := getCount(CollectionCount, f) - 1
	return strconv.FormatInt(createdCollectionId, 10)
}

// --- helpers ---
func mustJSON[T any](t *testing.T, v T) string {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	return string(b)
}

func strPtr(s string) *string { return &s }

func lastNFTID(f *FakeSDK) string {
	ptr := f.StateGetObject(NFTsCount)
	if ptr == nil {
		return "0"
	}
	count, _ := strconv.Atoi(*ptr)
	return strconv.Itoa(count - 1)
}

// --- MintNFTUniqueImpl ---
func TestMintNFTUniqueImpl_Positive(t *testing.T) {
	f := NewFakeSDK("creator1", "tx001")

	// Prepare Collection
	createCollectionImpl(&collectionPayload, f)
	args := MintNFTArgs{
		Collection:     createDummyCollection(f),
		Name:           "UniqueNFT",
		Description:    "Unique NFT desc",
		SingleTransfer: true,
		Metadata:       map[string]string{"key": "value"},
	}
	payload := mustJSON(t, args)
	mintNFTUniqueImpl(&payload, f)

	nftID := lastNFTID(f)
	nft := loadNFT(nftID, f)
	if nft.Creator != "creator1" || nft.NFTPrefs.Description != "Unique NFT desc" {
		t.Errorf("unexpected NFT after mint")
	}
}

func TestMintNFTUniqueImpl_Negative_NameEmpty(t *testing.T) {
	f := NewFakeSDK("creator1", "tx002")

	args := MintNFTArgs{
		Name:       "",
		Collection: createDummyCollection(f),
	}
	payload := mustJSON(t, args)
	defer expectAbort(t, f, "name is mandatory")
	mintNFTUniqueImpl(&payload, f)
}

func TestMintNFTUniqueImpl_Negative_NameTooLong(t *testing.T) {
	f := NewFakeSDK("creator1", "tx003")
	longName := strings.Repeat("x", maxNameLength+1)
	args := MintNFTArgs{
		Name:       longName,
		Collection: createDummyCollection(f),
	}
	payload := mustJSON(t, args)
	defer expectAbort(t, f, "name must between 1 - "+strconv.Itoa(maxNameLength)+" chars")
	mintNFTUniqueImpl(&payload, f)
}

func TestMintNFTUniqueImpl_Negative_DescriptionTooLong(t *testing.T) {
	f := NewFakeSDK("creator1", "tx004")
	longDesc := strings.Repeat("x", maxDescLength+1)
	args := MintNFTArgs{
		Name:        "NFT",
		Collection:  createDummyCollection(f),
		Description: longDesc,
	}
	payload := mustJSON(t, args)
	defer expectAbort(t, f, "desc max. "+strconv.Itoa(maxDescLength)+" chars")
	mintNFTUniqueImpl(&payload, f)
}

// --- MintNFTEditionsImpl ---
func TestMintNFTEditionsImpl_Positive(t *testing.T) {
	editionsTotal := 3
	f := NewFakeSDK("creator1", "tx005")
	args := MintNFTEditionsArgs{
		Collection:     createDummyCollection(f),
		Name:           "EditionNFT",
		Description:    "Edition desc",
		SingleTransfer: false,
		Metadata:       map[string]string{"type": "art"},
		EditionsTotal:  int64(editionsTotal),
	}
	payload := mustJSON(t, args)
	mintNFTEditionsImpl(&payload, f)

	genesisID := "0"
	editionIDs := GetIDsFromIndex(AllEditionsOfGenesis+genesisID, f)
	if len(editionIDs) != editionsTotal {
		t.Errorf("expected %d editions, got %d", editionsTotal, len(editionIDs))
	}
}

func TestMintNFTEditionsImpl_Negative_EditionZero(t *testing.T) {
	f := NewFakeSDK("creator1", "tx006")
	args := MintNFTEditionsArgs{
		Name:          "BadEdition",
		Collection:    createDummyCollection(f),
		EditionsTotal: 0,
	}
	payload := mustJSON(t, args)
	defer expectAbort(t, f, "editions total <= 0")
	mintNFTEditionsImpl(&payload, f)
}

// --- Metadata validation ---
func TestMetadata_TooManyKeys(t *testing.T) {
	f := NewFakeSDK("creator1", "tx007")
	meta := make(map[string]string)
	for i := 0; i < maxMetaKeys+1; i++ {
		meta["k"+strconv.Itoa(i)] = "v"
	}
	args := MintNFTArgs{
		Name:       "NFT",
		Collection: createDummyCollection(f),
		Metadata:   meta,
	}
	payload := mustJSON(t, args)
	defer expectAbort(t, f, "meta max. 25 keys")
	mintNFTUniqueImpl(&payload, f)
}

func TestMetadata_KeyTooLong(t *testing.T) {
	f := NewFakeSDK("creator1", "tx008")
	longKey := strings.Repeat("x", maxMetaKeyLength+1)
	args := MintNFTArgs{
		Name:       "NFT",
		Collection: createDummyCollection(f),
		Metadata:   map[string]string{longKey: "v"},
	}
	payload := mustJSON(t, args)
	defer expectAbort(t, f, "meta key '"+longKey+"' > 50 chars")
	mintNFTUniqueImpl(&payload, f)
}

func TestMetadata_ValueTooLong(t *testing.T) {
	f := NewFakeSDK("creator1", "tx009")
	longVal := strings.Repeat("x", maxMetaValueLength+1)
	args := MintNFTArgs{
		Name:       "NFT",
		Collection: createDummyCollection(f),
		Metadata:   map[string]string{"key": longVal},
	}
	payload := mustJSON(t, args)
	defer expectAbort(t, f, "meta value for 'key' > 512 chars")
	mintNFTUniqueImpl(&payload, f)
}

// --- TransferNFTImpl ---
func TestTransferNFT_Positive(t *testing.T) {
	senderAddress := sdk.Address("sender")
	receiverAddress := sdk.Address("receiver")
	f := NewFakeSDK(senderAddress.String(), "tx010")
	// create a collection for the nft that needs to be transferred
	senderCollectionId := createDummyCollection(f)
	// create another collection but now for the receiver
	f.env.Sender.Address = senderAddress
	f.env.Caller = senderAddress
	receiverCollectionId := createDummyCollection(f)
	// set sender and caller back to the initator of the transfer
	f.env.Sender.Address = senderAddress
	f.env.Caller = senderAddress

	args := MintNFTArgs{
		Name:       "TransferNFT",
		Collection: senderCollectionId,
	}
	payload := mustJSON(t, args)
	mintNFTUniqueImpl(&payload, f)
	nftID := lastNFTID(f)
	transferArgs := TransferNFTArgs{
		NftID:      nftID,
		Collection: receiverCollectionId,
		Owner:      receiverAddress,
	}
	transferPayload := mustJSON(t, transferArgs)
	transferNFTImpl(&transferPayload, f)

	nft := loadNFT(nftID, f)
	if nft.Owner != receiverAddress {
		t.Errorf("expected %s, got %s", receiverAddress, nft.Owner)
	}
}

func TestTransferNFT_Negative_Bound(t *testing.T) {
	userA := sdk.Address("sender")
	userB := sdk.Address("receiver")
	f := NewFakeSDK(userA.String(), "tx010")
	// create a collection for the nft that needs to be transferred
	userACollectionId := createDummyCollection(f)
	// create another collection but now for the receiver
	f.env.Sender.Address = userB
	f.env.Caller = userB
	userBCollectionId := createDummyCollection(f)
	// set sender and caller back to the initator of the transfer
	f.env.Sender.Address = userA
	f.env.Caller = userA
	// first we mint an nft
	args := MintNFTArgs{
		Name:           "BoundNFT",
		Collection:     userACollectionId,
		SingleTransfer: true,
	}
	payload := mustJSON(t, args)
	mintNFTUniqueImpl(&payload, f)

	// then we send the nft to user2
	nftID := lastNFTID(f)

	transferAArgs := TransferNFTArgs{NftID: nftID, Collection: userBCollectionId, Owner: userB}
	transferAPayload := mustJSON(t, transferAArgs)
	transferNFTImpl(&transferAPayload, f)

	// transfer back to sender
	f.env.Sender.Address = userB
	f.env.Caller = userB
	transferBArgs := TransferNFTArgs{NftID: nftID, Collection: userACollectionId, Owner: userA}
	transferBPayload := mustJSON(t, transferBArgs)
	defer expectAbort(t, f, "nft bound to owner")
	transferNFTImpl(&transferBPayload, f)
}

// --- Get functions ---
func TestGetNFT_Positive(t *testing.T) {
	f := NewFakeSDK("creator1", "tx012")

	args := MintNFTArgs{
		Name:       "GetNFT",
		Collection: createDummyCollection(f),
	}
	payload := mustJSON(t, args)
	mintNFTUniqueImpl(&payload, f)

	nftID := lastNFTID(f)
	res := getNFTImpl(&nftID, f)
	var nft NFT
	json.Unmarshal([]byte(*res), &nft)
	if nft.ID != nftID || nft.Creator != "creator1" {
		t.Errorf("unexpected NFT returned")
	}
}

func TestGetNFT_Negative_NotFound(t *testing.T) {
	f := NewFakeSDK("creator1", "tx013")
	id := "999"
	defer expectAbort(t, f, "nft 999 not found")
	getNFTImpl(&id, f)
}

func TestGetNFTsForCollection_Positive(t *testing.T) {
	f := NewFakeSDK("creator1", "tx014")
	CollectionId := createDummyCollection(f)
	for i := 0; i < 2; i++ {
		args := MintNFTArgs{
			Name:       "NFT" + strconv.Itoa(i),
			Collection: CollectionId,
		}
		payload := mustJSON(t, args)
		mintNFTUniqueImpl(&payload, f)
	}

	res := getNFTsForCollectionImpl(&CollectionId, f)
	var nfts []*NFT
	json.Unmarshal([]byte(*res), &nfts)
	if len(nfts) != 2 {
		t.Errorf("expected 2 NFTs, got %d", len(nfts))
	}
}

func TestGetNFTsForOwner_Positive(t *testing.T) {
	f := NewFakeSDK("creator1", "tx015")
	args := MintNFTArgs{
		Name:       "OwnerNFT",
		Collection: createDummyCollection(f),
	}
	payload := mustJSON(t, args)
	mintNFTUniqueImpl(&payload, f)

	ownerID := "creator1"
	res := getNFTsForOwnerImpl(&ownerID, f)
	var nfts []*NFT
	json.Unmarshal([]byte(*res), &nfts)
	if len(nfts) != 1 || nfts[0].Owner != "creator1" {
		t.Errorf("unexpected owner NFTs")
	}
}

func TestGetNFTsForCreator_Positive(t *testing.T) {
	f := NewFakeSDK("creator1", "tx016")
	args := MintNFTArgs{
		Name:       "CreatorNFT",
		Collection: createDummyCollection(f),
	}
	payload := mustJSON(t, args)
	mintNFTUniqueImpl(&payload, f)

	creatorID := "creator1"
	res := getNFTsForCreatorImpl(&creatorID, f)
	var nfts []*NFT
	json.Unmarshal([]byte(*res), &nfts)
	if len(nfts) != 1 || nfts[0].Creator != "creator1" {
		t.Errorf("unexpected creator NFTs")
	}
}

// --- Editions & Available Editions ---
func TestGetEditionsForNFT_Positive(t *testing.T) {
	editionsTotal := int64(2)
	f := NewFakeSDK("creator1", "tx017")
	args := MintNFTEditionsArgs{
		Name:          "EditionNFT",
		Collection:    createDummyCollection(f),
		EditionsTotal: editionsTotal,
	}
	payload := mustJSON(t, args)
	mintNFTEditionsImpl(&payload, f)

	lastIDString := lastNFTID(f)
	lastID, err := strconv.ParseInt(lastIDString, 10, 64)
	genesisID := strconv.FormatInt(lastID-editionsTotal+1, 10)
	if err != nil {
		fmt.Println("Error converting string to int64", err)
		return
	}
	res := getEditionsForNFTImpl(&genesisID, f)
	var nfts []*NFT
	json.Unmarshal([]byte(*res), &nfts)
	if int64(len(nfts)) != editionsTotal {
		t.Errorf("expected %d editions, got %d", editionsTotal, len(nfts))
	}
}

func TestGetAvailableEditionsForNFT_Dynamic(t *testing.T) {
	editionsTotal := int64(3)
	userA := sdk.Address("userA")

	f := NewFakeSDK(userA.String(), "tx010")
	// create a collection for the nft that needs to be transferred
	userACollectionId := createDummyCollection(f)
	// create another collection but now for the receiver
	userB := sdk.Address("userB")
	f.env.Sender.Address = userB
	f.env.Caller = userB
	userBCollectionId := createDummyCollection(f)
	// set sender and caller back to the initator of the transfer
	f.env.Sender.Address = userA
	f.env.Caller = userA

	args := MintNFTEditionsArgs{
		Collection:    userACollectionId,
		Name:          "AvailNFT",
		EditionsTotal: editionsTotal,
	}
	payload := mustJSON(t, args)
	mintNFTEditionsImpl(&payload, f)

	lastIDString := lastNFTID(f)
	lastID, err := strconv.ParseInt(lastIDString, 10, 64)
	genesisID := strconv.FormatInt(lastID-editionsTotal+1, 10)
	if err != nil {
		fmt.Println("Error converting string to int64", err)
		return
	}
	availIDs := GetIDsFromIndex(AvailEditionsOfGenesis+genesisID, f)
	if int64(len(availIDs)) != editionsTotal {
		t.Errorf("expected %d available editions, got %d", editionsTotal, len(availIDs))
	}

	transferArgs := TransferNFTArgs{
		NftID:      availIDs[0],
		Collection: userBCollectionId,
		Owner:      userB,
	}
	transferPayload := mustJSON(t, transferArgs)
	transferNFTImpl(&transferPayload, f)

	newAvailIDs := GetIDsFromIndex(AvailEditionsOfGenesis+genesisID, f)
	if int64(len(newAvailIDs)) != editionsTotal-1 {
		t.Errorf("expected %d available editions after transfer, got %d", editionsTotal-1, len(newAvailIDs))
	}
	for _, id := range newAvailIDs {
		nft := loadNFT(id, f)
		if nft.Owner != userA {
			t.Errorf("NFT %s should still be owned by %s", id, userA)
		}
	}
}

func TestGetAvailableEditionsForNFT_Negative(t *testing.T) {
	f := NewFakeSDK("creator1", "tx019")
	nonexistentID := "999"
	res := getAvailableEditionsForNFTImpl(&nonexistentID, f)
	if res == nil {
		t.Errorf("expected empty response, got nil")
	}
}
