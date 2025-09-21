package contract_test

import (
	"testing"
)

// admin tests
func TestAdminMarket(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "set_market", PayloadToJSON("hive:tibfox"), nil, "hive:tibfox", false, uint(10_000))
	CallContract(t, ct, "set_market", PayloadToJSON(""), nil, "hive:userA", false, uint(1_000))
	CallContract(t, ct, "set_market", PayloadToJSON("hive:tibfox"), nil, "hive:contractowner", true, uint(10_000_000))
	CallContract(t, ct, "get_market", PayloadToJSON(""), nil, "hive:contractowner", true, uint(10_000_000))
}

// collection tests

func TestColCreate(t *testing.T) {
	// just create a collection
	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))
}

func TestColCreateFails(t *testing.T) {
	// just create a collection
	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
		"name": "",
		"desc": "collection longer description",
	}), nil, "hive:userA", false, uint(100_000_000))

	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
		"name": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e",
		"desc": "collection longer description",
	}), nil, "hive:userA", false, uint(100_000_000))

	CallContract(t, SetupContractTest(), "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
	}), nil, "hive:userA", false, uint(100_000_000))

}

func TestColCreateAndGet(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))
}

// nft tests
func TestMintUniqueNFT(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))
	// mint nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", true, uint(100_000_000))
	// get minted nft
	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:userA", true, uint(10_000_000))

}

func TestMintUniqueNFTFails(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for minter
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:minter", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:receiver", true, uint(100_000_000))
	// mint nft (should fail) - collection not owned my minter
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", false, uint(100_000_000))

	// mint nft with character overflows
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", false, uint(100_000_000))

	// character overflows
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", false, uint(100_000_000))

	// meadata overflows

	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit.": "a value", // key too long
			"b": "b value",
		},
	}), nil, "hive:userA", false, uint(100_000_000))

	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Curabitur pretium tincidunt lacus, non luctus libero.", // value too long
			"b": "b value",
		},
	}), nil, "hive:userA", false, uint(100_000_000))

	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{ // too many entries
			"a": "f3jK2p",
			"b": "L8m2Qs",
			"c": "wP9vRt",
			"d": "h7Qx4Y",
			"e": "bN3kZd",
			"f": "V6rLmT",
			"g": "yP2fWq",
			"h": "J9vXsK",
			"i": "uH4bPn",
			"j": "Q7mLcR",
			"k": "zK5vFw",
			"l": "N2rHtB",
			"m": "pY8qWx",
			"n": "S4jVzL",
			"o": "dC9mPf",
			"p": "kL3bQt",
			"q": "X7vJnY",
			"r": "fH2qZp",
			"s": "R5mLcW",
			"t": "yK8vBt",
			"u": "J3pXnV",
			"v": "wQ6rLf",
			"w": "nP9bGk",
			"x": "S2tVzH",
			"y": "cL7mWx",
			"z": "hF4qRn",
		},
	}), nil, "hive:userA", false, uint(100_000_000))
}

func TestBurn(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))
	// mint 1 unique nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", true, uint(100_000_000))
	// burn it
	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:userA", true, uint(100_000_000))

}

func TestBurnByMarket(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(10_000_000))
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))
	// mint 1 unique nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", true, uint(100_000_000))

	// burn it by market
	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:marketaddress", true, uint(100_000_000))
}

func TestBurnFails(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))
	// mint 1 unique nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", true, uint(100_000_000))
	// burn it by other user
	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:secodUser", false, uint(100_000_000))
}

func TestMintEditions(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	// try to mint max nft editions (should succeed)
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
		"et": 10,
	}), nil, "hive:userA", true, uint(10_000_000_000_000))

	// get minted nft 0
	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:userA", true, uint(10_000_000))

	// get minted nft 9
	CallContract(t, ct, "nft_get", PayloadToJSON("9"), nil, "hive:userA", true, uint(10_000_000))

	// try to mint max+1 nft editions (should fail)
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
		"et": 1001,
	}), nil, "hive:userA", false, uint(10_000_000_000_000))

}

func TestExtendEditionNFTs(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))
	// create a collection
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userB", true, uint(100_000_000))

	// mint 10 nft editions
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
		"et": 10,
	}), nil, "hive:userA", true, uint(10_000_000_000_000))

	// extend by 10 more nft editions by userB (should fail)
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"g":     0,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
		"et": 10,
	}), nil, "hive:userB", false, uint(10_000_000_000_000))

	// extend by 10 more nft editions by userA (but different transfer rules)
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": false,
		"g":     0,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
		"et": 10,
	}), nil, "hive:userA", false, uint(10_000_000_000_000))

	// extend by 100 more nft editions by userA
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"g":     0,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
		"et": 100,
	}), nil, "hive:userA", true, uint(10_000_000_000_000))

}

func TestTransfers(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(10_000_000))
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userB", true, uint(100_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": false,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", true, uint(100_000_000))
	// transfer nft by minter (should success)
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    0,
		"owner": "hive:userB",
	}), nil, "hive:userA", true, uint(100_000_000))

	// transfer nft by market (should success)
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     0,
		"id":    0,
		"owner": "hive:userA",
	}), nil, "hive:marketaddress", true, uint(100_000_000))
}

func TestEditionTransfers(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userB", true, uint(100_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint_edition", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
		"et": 10,
	}), nil, "hive:userA", true, uint(10_000_000_000_000))
	// transfer edition no 3 nft (should success)
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    3,
		"owner": "hive:userB",
	}), nil, "hive:userA", true, uint(100_000_000))
}

func TestTranfersWithFails1(t *testing.T) {
	ct := SetupContractTest()

	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userB", true, uint(100_000_000))

	// create a 2nd collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my other cool collection",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", true, uint(100_000_000))

	// mint nft (should fail) - owner of collection != caller
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "my 3rd nft",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", false, uint(100_000_000))

	// move 1st nft (should fail) - collection and owner is same
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     0,
		"id":    0,
		"owner": "hive:userA",
	}), nil, "hive:userA", false, uint(100_000_000))

	// move 1st nft (should fail) - collection owned by other user
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    0,
		"owner": "hive:userA",
	}), nil, "hive:userA", false, uint(100_000_000))

}

func TestTranfersWithFails2(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userB", true, uint(100_000_000))

	// create a 2nd collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "my other cool collection",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	CallContract(t, ct, "col_get", PayloadToJSON("2"), nil, "hive:userA", true, uint(10_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     0,
		"name":  "nft name",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", true, uint(100_000_000))

	// transfer 1st nft (should success)
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     1,
		"id":    0,
		"owner": "hive:userB",
	}), nil, "hive:userA", true, uint(100_000_000))
	// transfer 1st nft back (should fail) - nft is bound to user
	CallContract(t, ct, "nft_transfer", PayloadToJSON(map[string]any{
		"c":     0,
		"id":    0,
		"owner": "hive:userA",
	}), nil, "hive:userB", false, uint(100_000_000))

}

func TestTranfersWithFails3(t *testing.T) {
	ct := SetupContractTest()
	// create a collection for sender
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userA", true, uint(100_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", PayloadToJSON(map[string]string{
		"name": "collection name",
		"desc": "collection longer description",
	}), nil, "hive:userB", true, uint(100_000_000))

	// mint nft (should fail) - owner of collection != caller
	CallContract(t, ct, "nft_mint_unique", PayloadToJSON(map[string]any{
		"c":     1,
		"name":  "my 3rd nft",
		"desc":  "nft longer description",
		"bound": true,
		"meta": map[string]string{
			"a": "a value",
			"b": "b value",
		},
	}), nil, "hive:userA", false, uint(100_000_000))
}
