package contract_test

import (
	"fmt"
	"testing"
)

// admin tests
func TestSetAdminAdressFromWrongAddress(t *testing.T) {
	CallContract(t, SetupContractTest(), "admin_set_market", PayloadToJSON("hive:tibfox"), nil, "hive:tibfox", false, uint(10_000))
}

func TestSetAdminAdressFEmptyAddress(t *testing.T) {
	CallContract(t, SetupContractTest(), "admin_set_market", PayloadToJSON(""), nil, "hive:tibfox.vsc", false, uint(1_000))
}

func TestSetAdminAdressFromCorrectAddress(t *testing.T) {
	CallContract(t, SetupContractTest(), "admin_set_market", PayloadToJSON("hive:tibfox"), nil, "hive:tibfox.vsc", true, uint(10_000_000))
}

// collection tests

func TestCreateCollection(t *testing.T) {
	// just create a collection
	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:tibfox.vsc", true, uint(100_000_000))
}

func TestCreateAndGetCollection(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:tibfox.vsc", true, uint(100_000_000))
	// get all collections of user
	CallContract(t, ct, "col_get_user", PayloadToJSON("hive:tibfox.vsc"), nil, "hive:tibfox.vsc", true, uint(100_000_000))
}

// nft tests
func TestCreateCollectionAndMintUniqueNFT(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:tibfox.vsc", true, uint(100_000_000))
	// mint nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
	}), nil, "hive:tibfox.vsc", true, uint(100_000_000))
	// get minted nft
	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:tibfox.vsc", true, uint(10_000_000))
	// CallContract(t, "nft_get_creator", PayloadToJSON("hive:tibfox.vsc"), nil, "hive:tibfox.vsc", false, true)

}

func TestCreateCollectionAndMintEditionNFT(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	fmt.Println("create col")
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:tibfox.vsc", true, uint(100_000_000))
	// mint 1 unique nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
	}), nil, "hive:tibfox.vsc", true, uint(100_000_000))
	// try to mint 100 nft editions
	fmt.Println("mint 100 editions (should fail)")
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		"et": 100,
	}), nil, "hive:tibfox.vsc", false, uint(10_000_000_000_000))

	// try to mint 99 nft editions
	fmt.Println("mint 99 editions (should not fail)")
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		"et": 99,
	}), nil, "hive:tibfox.vsc", true, uint(10_000_000_000_000))
	// get minted nfts by collection id
	fmt.Println("get collection nfts")
	// CallContract(t, ct, "nft_get_collection", PayloadToJSON("0"), nil, "hive:tibfox.vsc", true, uint(100_000_000))
	// CallContract(t, "nft_get_creator", PayloadToJSON("hive:tibfox.vsc"), nil, "hive:tibfox.vsc", false, true)
}
