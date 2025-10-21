package contract_test

import (
	"testing"
)

// // admin tests
func TestAdminMarket(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "set_market", PayloadToJSON("hive:tibfox"), nil, "hive:tibfox", false, uint(10_000))
	CallContract(t, ct, "set_market", PayloadToJSON(""), nil, "hive:someone", false, uint(1_000))
	CallContract(t, ct, "set_market", PayloadToJSON("hive:tibfox"), nil, "hive:contractowner", true, uint(100_000_000))
	CallContract(t, ct, "get_market", PayloadToJSON(""), nil, "hive:contractowner", true, uint(100_000_000))

}

// collection tests

func TestColCreate(t *testing.T) {
	ct := SetupContractTest()
	// just create a collection
	// CallContract(t, ct, "col_create", PayloadToJSON("collectionA|my description"), nil, "hive:someone", true, uint(10_000_000))
	CallContract(t, ct, "col_get", PayloadToJSON("hive:someone/0"), nil, "hive:someone", true, uint(100_000_000))
}

// func TestColCreateFails(t *testing.T) {
// 	// just create a collection
// 	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
// 		"name": "",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
// 		"name": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
// 	}), nil, "hive:someone", false, uint(100_000_000))

// }

// func TestColCreateAndGet(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))
// }

// // nft tests
// func TestMintUniqueNFT(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// mint nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// get minted nft
// 	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:someone", true, uint(10_000_000))

// }

// func TestMintUniqueNFTFails(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection for minter
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:minter", true, uint(100_000_000))

// 	// create a collection for receiver
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:receiver", true, uint(100_000_000))
// 	// mint nft (should fail) - collection not owned my minter
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     1,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	// mint nft with character overflows
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	// character overflows
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	// meadata overflows

// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"Lorem ipsum dolor sit amet, consectetur adipiscing elit.": "a value", // key too long
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Curabitur pretium tincidunt lacus, non luctus libero.", // value too long
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     1,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{ // too many entries
// 			"a": "f3jK2p",
// 			"b": "L8m2Qs",
// 			"c": "wP9vRt",
// 			"d": "h7Qx4Y",
// 			"e": "bN3kZd",
// 			"f": "V6rLmT",
// 			"g": "yP2fWq",
// 			"h": "J9vXsK",
// 			"i": "uH4bPn",
// 			"j": "Q7mLcR",
// 			"k": "zK5vFw",
// 			"l": "N2rHtB",
// 			"m": "pY8qWx",
// 			"n": "S4jVzL",
// 			"o": "dC9mPf",
// 			"p": "kL3bQt",
// 			"q": "X7vJnY",
// 			"r": "fH2qZp",
// 			"s": "R5mLcW",
// 			"t": "yK8vBt",
// 			"u": "J3pXnV",
// 			"v": "wQ6rLf",
// 			"w": "nP9bGk",
// 			"x": "S2tVzH",
// 			"y": "cL7mWx",
// 			"z": "hF4qRn",
// 		},
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	// mint nft without name
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c": 0,
// 	}), nil, "hive:someone", false, uint(100_000_000))

// 	// mint nft without collection
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"name": "nft name",
// 	}), nil, "hive:someone", false, uint(100_000_000))
// }

// func TestBurn(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// mint 1 unique nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// burn it
// 	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:someone", true, uint(100_000_000))

// }

// func TestBurnByMarket(t *testing.T) {
// 	ct := SetupContractTest()
// 	// set market
// 	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(10_000_000))
// 	// create a collection
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// mint 1 unique nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// burn it by market
// 	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:marketaddress", true, uint(100_000_000))
// }

// func TestBurnEdition(t *testing.T) {
// 	ct := SetupContractTest()
// 	// set market
// 	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(10_000_000))
// 	// create a collection
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// mint 1 unique nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 		"et": 10,
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// burn nft 0 edition 0 (is default for editioned nfts) - should succeed
// 	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:someone", true, uint(100_000_000))

// 	// try to burn edition - should succeed
// 	CallContract(t, ct, "nft_burn", PayloadToJSON("0:1"), nil, "hive:someone", true, uint(100_000_000))

// 	// try to burn edition 2 by other user - should fail
// 	CallContract(t, ct, "nft_burn", PayloadToJSON("0:2"), nil, "hive:someoneelse", false, uint(100_000_000))
// }

// func TestBurnFails(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// mint 1 unique nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", true, uint(100_000_000))
// 	// burn it by other user
// 	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:secodUser", false, uint(100_000_000))
// }

// func TestMintEditions(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// try to mint max nft editions (should succeed)
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": true,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 		"et": 10,
// 	}), nil, "hive:someone", true, uint(10_000_000_000_000))

// 	// get minted nft 0
// 	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:someone", true, uint(10_000_000))

// 	// get minted nft 0 - edition 9
// 	CallContract(t, ct, "nft_get", PayloadToJSON("0:9"), nil, "hive:someone", true, uint(10_000_000))
// }

// func TestTransfers(t *testing.T) {
// 	ct := SetupContractTest()
// 	// set market
// 	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(10_000_000))
// 	// create a collection for sender
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// create a collection for receiver
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someoneelse", true, uint(100_000_000))

// 	// mint nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": false,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// transfer nft by minter (should success)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  1,
// 		"id": "0",
// 		"o":  "hive:someoneelse",
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// transfer nft by market (should success)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  0,
// 		"id": "0",
// 		"o":  "hive:someone",
// 	}), nil, "hive:marketaddress", true, uint(100_000_000))
// }

// func TestTransfersFails(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection for sender
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// create a collection for receiver
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someoneelse", true, uint(100_000_000))

// 	// mint nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": false,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// transfer nft by minter to collection not owned by new owner (should fail)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  0,
// 		"id": "0",
// 		"o":  "hive:someoneelse",
// 	}), nil, "hive:someone", false, uint(100_000_000))

// }

// func TestEditionTransfers(t *testing.T) {
// 	ct := SetupContractTest()
// 	// create a collection for sender
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// create a collection for receiver
// 	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
// 		"name": "collection name",
// 		"desc": "collection longer description",
// 	}), nil, "hive:someoneelse", true, uint(100_000_000))

// 	// mint nft
// 	CallContract(t, ct, "nft_mint", PayloadToJSON(map[string]any{
// 		"c":     0,
// 		"name":  "nft name",
// 		"desc":  "nft longer description",
// 		"bound": false,
// 		"meta": map[string]string{
// 			"a": "a value",
// 			"b": "b value",
// 		},
// 		"et": 999_999_999,
// 	}), nil, "hive:someone", true, uint(10_000_000_000_000))
// 	// transfer edition no 3 nft (should success)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  1,
// 		"id": "0:3",
// 		"o":  "hive:someoneelse",
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// transfer edition no 99999 nft (should success)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  1,
// 		"id": "0:99999",
// 		"o":  "hive:someoneelse",
// 	}), nil, "hive:someone", true, uint(100_000_000))

// 	// transfer back edition no 99999 nft (should success)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  0,
// 		"id": "0:99999",
// 		"o":  "hive:someone",
// 	}), nil, "hive:someoneelse", true, uint(100_000_000))

// 	// maliciou trying to transfer edition no 5000 not by owner (should fail)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  1,
// 		"id": "0:5000",
// 		"o":  "hive:someoneelse",
// 	}), nil, "hive:someoneelse", false, uint(100_000_000))

// 	// trying to transfer edition no 6000 to collection not owned by new nft owner (should fail)
// 	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
// 		"c":  0,
// 		"id": "0:6000",
// 		"o":  "hive:someoneelse",
// 	}), nil, "hive:someone", false, uint(100_000_000))

// }
