package contract_test

import (
	"testing"
)

// // // admin tests
func TestAdminMarket(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "add_market", []byte("vscxyz"), nil, "hive:nocontractowner", false, uint(1_000_000_000), "")
	CallContract(t, ct, "add_market", []byte(""), nil, "hive:contractowner", false, uint(1_000), "")
	CallContract(t, ct, "add_market", []byte("vscxyz"), nil, "hive:contractowner", true, uint(100_000_000), "")
	CallContract(t, ct, "get_markets", []byte(""), nil, "hive:contractowner", true, uint(100_000_000), "vscxyz")
	CallContract(t, ct, "add_market", []byte("vscabc"), nil, "hive:contractowner", true, uint(100_000_000), "")
	CallContract(t, ct, "get_markets", []byte(""), nil, "hive:contractowner", true, uint(100_000_000), "vscxyz|vscabc")
	CallContract(t, ct, "remove_market", []byte("vscxyz"), nil, "hive:contractowner", true, uint(100_000_000), "")
	CallContract(t, ct, "get_markets", []byte(""), nil, "hive:contractowner", true, uint(100_000_000), "vscabc")

}

// // collection tests

func TestColCreateSuccess(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB||"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_get", []byte("hive:someone_0"), nil, "hive:someone", true, uint(100_000_000), "")
	CallContract(t, ct, "col_get", []byte("hive:someone_1"), nil, "hive:someone", true, uint(100_000_000), "")
	CallContract(t, ct, "col_count", []byte("hive:someone"), nil, "hive:someone", true, uint(100_000_000), "2")
	CallContract(t, ct, "col_exists", []byte("hive:someone_1"), nil, "hive:someone", true, uint(100_000_000), "true")
}

func TestColCreateFails(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "col_create",
		[]byte("|my description|img=testurl"),
		nil, "hive:someone", false, uint(100_000_000), "")

	CallContract(t, ct, "col_create",
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e|my description|img=testurl"),
		nil, "hive:someone", false, uint(100_000_000), "")

	CallContract(t, ct, "col_create",
		[]byte("collectionA|Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.|img=testurl"),
		nil, "hive:someone", false, uint(100_000_000), "")

	CallContract(t, ct, "col_exists", []byte("hive:someone_0"), nil, "hive:someone", true, uint(100_000_000), "false")
	CallContract(t, ct, "col_get", []byte("hive:someone_0"), nil, "hive:someone", false, uint(100_000_000), "")
	CallContract(t, ct, "col_count", []byte("hive:someone"), nil, "hive:someone", true, uint(100_000_000), "0")

}

// // nft tests
func TestMintUniqueNFTSingleSuccess(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|"), nil, "hive:someone", true, uint(1_000_000_000), "")
	// mint without edition value
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|n||true||img=testurl"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_get",
		[]byte("0"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

}

func TestMintUniqueNFTSuccess(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000), "")
	// mint without edition value
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true||test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// mint nft with edition value 0
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// mint nft with edition value 1
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// mint nft to another collection
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someoneelse_0|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// get minted nft
	CallContract(t, ct, "nft_get", []byte("0"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_get", []byte("3"), nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_isOwner", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_ownerColOf", []byte("3"), nil, "hive:someone", true, uint(100_000_000), "hive:someoneelse_0")
	CallContract(t, ct, "nft_creator", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "hive:someone")
	CallContract(t, ct, "nft_supply", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "1")
	CallContract(t, ct, "nft_meta", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "test=123,test2=abc")
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "false")
	CallContract(t, ct, "nft_isSingleTransfer", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_isSingleTransfer", []byte("3"), nil, "hive:someone", true, uint(100_000_000), "false")

}

// mint editions
func TestMintEditionsSuccess(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// mint nft editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|10|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// get minted nft 0
	CallContract(t, ct, "nft_get", PayloadToJSON("0|0"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// // get minted nft 0 - edition 9
	CallContract(t, ct, "nft_get", PayloadToJSON("0|9"), nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_isOwner", []byte("0|0"), nil, "hive:someone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_isOwner", []byte("0|9"), nil, "hive:someone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_ownerColOf", []byte("0|3"), nil, "hive:someone", true, uint(100_000_000), "hive:someone_0")

	CallContract(t, ct, "nft_supply", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "10")

	CallContract(t, ct, "nft_meta", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "test=123,test2=abc")
	CallContract(t, ct, "nft_meta", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "test=123,test2=abc")

}

func TestMintUniqueNFTFails(t *testing.T) {
	ct := SetupContractTest()
	// // create a collection for minter
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// // mint nft with character overflows
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", false, uint(100_000_000), "")

	// // // character overflows

	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|nft name|Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.|true|0|test=123,test2=abc"),
		nil, "hive:someone", false, uint(100_000_000), "")

	// // mint nft with wrong param count
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|asd"),
		nil, "hive:someone", false, uint(100_000_000), "")

	// // mint nft without name
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0||description|true|0|test=123"),
		nil, "hive:someone", false, uint(100_000_000), "")

	// // mint nft without collection
	CallContract(t, ct, "nft_mint",
		[]byte("|asd|description|true|0|test=123"),
		nil, "hive:someone", false, uint(100_000_000), "")

}

// burn
func TestBurnSuccess(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// mint
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// burn 1st one
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "false")
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "true")
}

func TestBurnEditionSuccess(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")
	// mint 10 nft editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|10|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// burn nft 0 edition 0 (is default for editioned nfts) - should succeed
	CallContract(t, ct, "nft_burn", PayloadToJSON("0|0"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|0"), nil, "hive:someone", true, uint(100_000_000), "true")

	// try to burn edition - should succeed
	CallContract(t, ct, "nft_burn", PayloadToJSON("0|9"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|9"), nil, "hive:someone", true, uint(100_000_000), "true")

}

func TestBurnFails(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")
	// mint 1 unique nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")
	// mint 10 nft editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|10|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")
	// try to burn editioned nft without specifying edition index
	CallContract(t, ct, "nft_burn", PayloadToJSON("1"), nil, "hive:someone", false, uint(1_000_000_000), "")
	// burn edition for unique nft
	CallContract(t, ct, "nft_burn",
		[]byte("0|2"),
		nil, "hive:someone", false, uint(100_000_000), "")
	// check burn state for editioned nft without specifying edition index
	CallContract(t, ct, "nft_isBurned", []byte("1"), nil, "hive:someone", false, uint(100_000_000), "")
	// burn it by other user
	CallContract(t, ct, "nft_burn",
		[]byte("0|0"),
		nil, "hive:otheruser", false, uint(100_000_000), "")
	// try to burn edition 2 by other user - should fail
	CallContract(t, ct, "nft_burn",
		[]byte("1|2"),
		nil, "hive:someoneelse", false, uint(1_000_000_000), "")

}

func TestTransfersSuccess(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "add_market", []byte("vscxyz"), nil, "hive:contractowner", true, uint(1_000_000_000), "")
	// create a collection for sender
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000), "")

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// transfer nft by minter (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someoneelse_0"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "hive:someoneelse_0")

	// transfer nft by minter (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someone_0"),
		nil, "vscxyz", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isOwner", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "true")
}

func TestTransfersSuccessBurnProtection(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "add_market", []byte("vscxyz"), nil, "hive:contractowner", true, uint(1_000_000_000), "")
	// create a collection for sender
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000), "")

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// transfer nft by minter (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someoneelse_0"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:someoneelse", true, uint(100_000_000), "")

	// transfer nft by market (should fail bc it i burned)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someone_0"),
		nil, "vscxyz", false, uint(1_000_000_000), "msg: nft is burned")
}

// test transfers with excepted fails
func TestTransfersFails(t *testing.T) {
	ct := SetupContractTest()

	// create a collection for sender
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000), "")

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// transfer nft by other user than minter
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someoneelse_0"),
		nil, "hive:someoneelse", false, uint(100_000_000), "")

	// transfer non existing nft
	CallContract(t, ct, "nft_transfer",
		[]byte("1||hive:someone_0"),
		nil, "hive:someone", false, uint(100_000_000), "")

	// transfer nft to non existing collection
	CallContract(t, ct, "nft_transfer",
		[]byte("1||hive:someoneelse_0"),
		nil, "hive:someone", false, uint(100_000_000), "")

	// transfer nft without collection
	CallContract(t, ct, "nft_transfer",
		[]byte("1||"),
		nil, "hive:someone", false, uint(100_000_000), "")

	// transfer nft with non existing nft id
	CallContract(t, ct, "nft_transfer",
		[]byte("0|2|hive:someone_0"),
		nil, "hive:someone", false, uint(100_000_000), "")

}

// edition transfers
func TestEditionTransfers(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "add_market", []byte("vscxyz"), nil, "hive:contractowner", true, uint(1_000_000_000), "")
	// create a collection for sender
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000), "")

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|999999|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_isOwner", []byte("0|23"), nil, "hive:someone", true, uint(100_000_000), "true")

	CallContract(t, ct, "nft_supply", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "999999")
	CallContract(t, ct, "nft_meta", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "test=123,test2=abc")

	// transfer edition no 3 nft (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|3|hive:someoneelse_0"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_isOwner", []byte("0|3"), nil, "hive:someoneelse", true, uint(100_000_000), "true")

	// transfer edition no 999999 nft (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|99999|hive:someoneelse_0"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// transfer back edition no 999999 nft (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|99999|hive:someone_0"),
		nil, "hive:someoneelse", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isOwner", []byte("0|99999"), nil, "hive:someone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_ownerColOf", []byte("0|99999"), nil, "hive:someone", true, uint(100_000_000), "hive:someone_0")

	// // // maliciou trying to transfer edition no 5000 not by owner (should fail)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|5000|hive:someoneelse_0"),
		nil, "hive:someoneelse", false, uint(1_000_000_000), "msg: only market or owner can transfer")

	CallContract(t, ct, "nft_isOwner", []byte("0|5000"), nil, "hive:someone", true, uint(100_000_000), "true")

}

// edition transfers burn protection
func TestEditionTransfersBurnProtection(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000), "")

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000), "")

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|10|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_isOwner", []byte("0|7"), nil, "hive:someone", true, uint(100_000_000), "true")
	// burn edition 7
	CallContract(t, ct, "nft_burn", PayloadToJSON("0|7"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|7"), nil, "hive:someone", true, uint(100_000_000), "true")
	// try to transfer edition 23
	CallContract(t, ct, "nft_transfer",
		[]byte("0|7|hive:someoneelse_0"),
		nil, "hive:someone", false, uint(1_000_000_000), "")
	// burn edition 8 by someone else (should fail)
	CallContract(t, ct, "nft_burn", PayloadToJSON("0|8"), nil, "hive:someoneelse", false, uint(1_000_000_000), "")

}
