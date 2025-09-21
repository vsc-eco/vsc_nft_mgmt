package contract_test

import (
	"testing"
)

// admin tests
func TestSetAdminAdressFromWrongAddress(t *testing.T) {
	CallContract(t, SetupContractTest(), "admin_set_market", PayloadToJSON("hive:tibfox"), nil, "hive:tibfox", false, uint(10_000))
}

func TestSetAdminAdressFEmptyAddress(t *testing.T) {
	CallContract(t, SetupContractTest(), "admin_set_market", PayloadToJSON(""), nil, "hive:firstUser", false, uint(1_000))
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
	}), nil, "hive:firstUser", true, uint(100_000_000))
}

func TestCreateAndGetCollection(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))
	// get all collections of user
	CallContract(t, ct, "col_get_user", PayloadToJSON("hive:firstUser"), nil, "hive:firstUser", true, uint(100_000_000))
}

// nft tests
func TestMintUniqueNFT(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))
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
	}), nil, "hive:firstUser", true, uint(100_000_000))
	// get minted nft
	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(10_000_000))
	// CallContract(t, "nft_get_creator", PayloadToJSON("hive:firstUser"), nil, "hive:firstUser", false, true)

}

func TestMintUniqueNFTFils(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for minter
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:minter", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:receiver", true, uint(100_000_000))
	// mint nft (should fail) - collection not owned my minter
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
	}), nil, "hive:firstUser", false, uint(100_000_000))

}

func TestBasicMintNFTs(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))
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
	}), nil, "hive:firstUser", true, uint(100_000_000))
	// try to mint 101 nft editions (should fail)
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		"et": 101,
	}), nil, "hive:firstUser", false, uint(10_000_000_000_000))

	// try to mint 100 nft editions (should succeed)
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
	}), nil, "hive:firstUser", true, uint(10_000_000_000_000))
	// get minted nfts by collection id
}

func TestExtendEditionNFTs(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))
	// mint 10 nft editions
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		"et": 10,
	}), nil, "hive:firstUser", true, uint(10_000_000_000_000))
	CallContract(t, ct, "nft_get_editions", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(1_000_000))
	// extend by 100 more nft editions
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"g":     0,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		"et": 100,
	}), nil, "hive:firstUser", true, uint(10_000_000_000_000))

	CallContract(t, ct, "nft_get_creator", PayloadToJSON("hive:firstUser"), nil, "hive:firstUser", true, uint(10_000_000))
	CallContract(t, ct, "nft_get_collection", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(10_000_000))
	CallContract(t, ct, "nft_get_editions", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(10_000_000))
	CallContract(t, ct, "nft_get_available", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(10_000_000))
}

func TestBasicTranfers(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:secondUser", true, uint(100_000_000))

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
	}), nil, "hive:firstUser", true, uint(100_000_000))
	// transfer 1st nft (should success)
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    0,
		"owner": "hive:secondUser",
	}), nil, "hive:firstUser", true, uint(100_000_000))
}

func TestEditionTransfers(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:secondUser", true, uint(100_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "my first nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
		"et": 10,
	}), nil, "hive:firstUser", true, uint(10_000_000_000_000))
	// transfer edition no 3 nft (should success)
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    3,
		"owner": "hive:secondUser",
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// CallContract(t, ct, "nft_get_creator", PayloadToJSON("hive:firstUser"), nil, "hive:firstUser", true, uint(10_000_000))
	// t.Log("minter collection")
	// CallContract(t, ct, "nft_get_collection", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(10_000_000))
	// t.Log("receiver collection")
	// CallContract(t, ct, "nft_get_collection", PayloadToJSON("1"), nil, "hive:firstUser", true, uint(10_000_000))
	// t.Log("all editions")
	// CallContract(t, ct, "nft_get_editions", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(10_000_000))
	// t.Log("available editions")
	// CallContract(t, ct, "nft_get_availableList", PayloadToJSON("0"), nil, "hive:firstUser", true, uint(10_000_000))
}

func TestTranfersWithFails1(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:secondUser", true, uint(100_000_000))

	// create a 2nd collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my other cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))

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
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// mint nft (should fail) - owner of collection != caller
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "my 3rd nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
	}), nil, "hive:firstUser", false, uint(100_000_000))

	// move 1st nft (should fail) - collection and owner is same
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     0,
		"id":    0,
		"owner": "hive:firstUser",
	}), nil, "hive:firstUser", false, uint(100_000_000))

	// move 1st nft (should fail) - collection owned by other user
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    0,
		"owner": "hive:firstUser",
	}), nil, "hive:firstUser", false, uint(100_000_000))

}

func TestTranfersWithFails2(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:secondUser", true, uint(100_000_000))

	// create a 2nd collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my other cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))

	CallContract(t, ct, "col_get", PayloadToJSON("2"), nil, "hive:firstUser", true, uint(10_000_000))

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
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// transfer 1st nft (should success)
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    0,
		"owner": "hive:secondUser",
	}), nil, "hive:firstUser", true, uint(100_000_000))
	// transfer 1st nft back (should fail) - nft is bound to user
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     0,
		"id":    0,
		"owner": "hive:firstUser",
	}), nil, "hive:secondUser", false, uint(100_000_000))

}

func TestTranfersWithFails3(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:firstUser", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my cool collection",
		"desc": "description of my cool collection",
	}), nil, "hive:secondUser", true, uint(100_000_000))

	// mint nft (should fail) - owner of collection != caller
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "my 3rd nft",
		"desc":  "some description",
		"bound": true,
		"meta": map[string]string{
			"foo":   "bar",
			"hello": "world",
		},
	}), nil, "hive:firstUser", false, uint(100_000_000))
}
