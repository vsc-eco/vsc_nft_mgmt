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
	// mint without edition value (defaults to 1)
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true||test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// mint nft with edition value 1
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// mint nft to another collection (singleTransfer=false)
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someoneelse_0|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000), "")

	// get minted nft
	CallContract(t, ct, "nft_get", []byte("0"), nil, "hive:someone", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_get", []byte("2"), nil, "hive:someone", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_isOwner", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_ownerColOf", []byte("2"), nil, "hive:someone", true, uint(100_000_000), "hive:someoneelse_0")
	CallContract(t, ct, "nft_creator", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "hive:someone")
	CallContract(t, ct, "nft_supply", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "1")
	CallContract(t, ct, "nft_meta", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "test=123,test2=abc")
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "false")
	CallContract(t, ct, "nft_isSingleTransfer", []byte("0"), nil, "hive:someone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_isSingleTransfer", []byte("2"), nil, "hive:someone", true, uint(100_000_000), "false")

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
		[]byte("hive:someone_0|name|description|true|1|test=123,test2=abc"),
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
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:someone"), nil, "hive:someone", true, uint(100_000_000), "99999")
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

// =====================
// SOULBOUND NFT TESTS
// =====================

func TestSoulboundTransferByCreator(t *testing.T) {
	ct := SetupContractTest()
	// create collections
	CallContract(t, ct, "col_create", []byte("collectionA|my description|"), nil, "hive:creator", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|my description|"), nil, "hive:recipient", true, uint(1_000_000_000), "")

	// mint soulbound NFT (singleTransfer=true)
	CallContract(t, ct, "nft_mint",
		[]byte("hive:creator_0|soulbound|desc|true|1|meta"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// verify it's soulbound
	CallContract(t, ct, "nft_isSingleTransfer", []byte("0"), nil, "hive:creator", true, uint(100_000_000), "true")

	// creator can transfer once (should succeed)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:recipient_0"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:creator", true, uint(100_000_000), "hive:recipient_0")
}

func TestSoulboundTransferByRecipientFails(t *testing.T) {
	ct := SetupContractTest()
	// create collections
	CallContract(t, ct, "col_create", []byte("collectionA|my description|"), nil, "hive:creator", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|my description|"), nil, "hive:recipient", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionC|my description|"), nil, "hive:thirdparty", true, uint(1_000_000_000), "")

	// mint soulbound NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:creator_0|soulbound|desc|true|1|meta"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// creator transfers to recipient
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:recipient_0"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// recipient tries to transfer to third party (should FAIL - soulbound)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:thirdparty_0"),
		nil, "hive:recipient", false, uint(1_000_000_000), "msg: nft bound to owner")
}

func TestSoulboundMarketCannotTransfer(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "add_market", []byte("vscmarket"), nil, "hive:contractowner", true, uint(1_000_000_000), "")

	// create collections
	CallContract(t, ct, "col_create", []byte("collectionA|my description|"), nil, "hive:creator", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|my description|"), nil, "hive:recipient", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionC|my description|"), nil, "hive:thirdparty", true, uint(1_000_000_000), "")

	// mint soulbound NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:creator_0|soulbound|desc|true|1|meta"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// creator transfers to recipient
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:recipient_0"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// market tries to transfer soulbound NFT (should FAIL - soulbound even for markets)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:thirdparty_0"),
		nil, "vscmarket", false, uint(1_000_000_000), "msg: nft bound to owner")
}

// =====================
// EDITION MAPPING TESTS
// =====================

func TestEditionMappingMultipleEditions(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:buyer", true, uint(1_000_000_000), "")

	// mint NFT with 10 editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|10|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// transfer editions 2, 5, 7 to buyer
	CallContract(t, ct, "nft_transfer", []byte("0|2|hive:buyer_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|5|hive:buyer_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|7|hive:buyer_0"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// verify buyer owns these editions
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:buyer"), nil, "hive:buyer", true, uint(100_000_000), "2,5,7")
}

func TestEditionMappingTransferBackAndForth(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:buyer", true, uint(1_000_000_000), "")

	// mint NFT with 5 editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// transfer edition 2 to buyer
	CallContract(t, ct, "nft_transfer", []byte("0|2|hive:buyer_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:buyer"), nil, "hive:buyer", true, uint(100_000_000), "2")

	// transfer edition 2 back to owner
	CallContract(t, ct, "nft_transfer", []byte("0|2|hive:owner_0"), nil, "hive:buyer", true, uint(1_000_000_000), "")

	// buyer should have no editions now (empty result)
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:buyer"), nil, "hive:buyer", true, uint(100_000_000), "")

	// owner should have edition 2 back (no duplicates)
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:owner"), nil, "hive:owner", true, uint(100_000_000), "2")

	// transfer back to buyer again - should not cause duplicates
	CallContract(t, ct, "nft_transfer", []byte("0|2|hive:buyer_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:buyer"), nil, "hive:buyer", true, uint(100_000_000), "2")
}

func TestEditionMappingEmptyResult(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT with editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// check for non-owner - should return empty
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:nonowner"), nil, "hive:someone", true, uint(100_000_000), "")
}

// ==============================
// COLLECTION-ONLY TRANSFER TEST
// ==============================

func TestCollectionOnlyTransfer(t *testing.T) {
	ct := SetupContractTest()
	// create two collections for same owner
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT in collection A
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// verify NFT is in collection 0
	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:owner", true, uint(100_000_000), "hive:owner_0")

	// transfer NFT to collection B (same owner, different collection)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:owner_1"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// verify NFT moved to collection 1
	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:owner", true, uint(100_000_000), "hive:owner_1")

	// owner should still own it
	CallContract(t, ct, "nft_isOwner", []byte("0"), nil, "hive:owner", true, uint(100_000_000), "true")
}

// ====================
// EDGE CASE TESTS
// ====================

func TestTransferToSameLocationFails(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// transfer to same location should fail
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:owner_0"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: source and target are the same")
}

func TestBurnAlreadyBurnedEdition(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// burn edition 2
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|2"), nil, "hive:owner", true, uint(100_000_000), "true")

	// burn same edition again - this should succeed (idempotent operation)
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|2"), nil, "hive:owner", true, uint(100_000_000), "true")
}

func TestMintToNonExistentCollectionFails(t *testing.T) {
	ct := SetupContractTest()
	// don't create collection, try to mint
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: collection not found")
}

func TestEditionIndexOutOfRange(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT with 5 editions (indices 0-4)
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// try to get edition 5 (out of range)
	CallContract(t, ct, "nft_get", []byte("0|5"), nil, "hive:owner", false, uint(100_000_000), "msg: edition index out of range")

	// try to transfer edition 5
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:other", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|5|hive:other_0"), nil, "hive:owner", false, uint(100_000_000), "msg: edition index out of range")

	// try to burn edition 5
	CallContract(t, ct, "nft_burn", []byte("0|5"), nil, "hive:owner", false, uint(100_000_000), "msg: edition index out of range")
}

// =====================
// METADATA TESTS
// =====================

func TestMetadataWithSpecialCharacters(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT with JSON-like metadata containing special chars (but no pipe in metadata field)
	metadata := `{"name":"test","url":"https://example.com?a=1&b=2","emoji":"🎨"}`
	payload := []byte("hive:owner_0|name|desc|false|1|" + metadata)
	CallContract(t, ct, "nft_mint", payload, nil, "hive:owner", true, uint(1_000_000_000), "")

	// verify metadata is stored correctly
	CallContract(t, ct, "nft_meta", []byte("0"), nil, "hive:owner", true, uint(100_000_000), metadata)
}

func TestMetadataLargeButValid(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// create metadata just under 32KB limit (32000 bytes to be safe)
	largeMeta := make([]byte, 30000)
	for i := range largeMeta {
		largeMeta[i] = 'x'
	}

	payload := append([]byte("hive:owner_0|name|desc|false|1|"), largeMeta...)
	// Large metadata requires more gas for state storage
	CallContract(t, ct, "nft_mint", payload, nil, "hive:owner", true, uint(100_000_000_000), "")

	// verify NFT was created
	CallContract(t, ct, "nft_supply", []byte("0"), nil, "hive:owner", true, uint(100_000_000), "1")
}

func TestMetadataTooLarge(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// create metadata over 32KB limit
	hugeMeta := make([]byte, 33000)
	for i := range hugeMeta {
		hugeMeta[i] = 'y'
	}

	payload := append([]byte("hive:owner_0|name|desc|false|1|"), hugeMeta...)
	CallContract(t, ct, "nft_mint", payload, nil, "hive:owner", false, uint(1_000_000_000), "msg: metadata too large")
}

func TestCollectionMetadataTooLarge(t *testing.T) {
	ct := SetupContractTest()

	// create metadata over 32KB limit for collection
	hugeMeta := make([]byte, 33000)
	for i := range hugeMeta {
		hugeMeta[i] = 'z'
	}

	payload := append([]byte("collectionA|desc|"), hugeMeta...)
	CallContract(t, ct, "col_create", payload, nil, "hive:owner", false, uint(1_000_000_000), "msg: metadata too large")
}

// ================================
// CONCURRENT EDITION TRANSFER TEST
// ================================

func TestConcurrentEditionTransfers(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:alice", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionC|desc|"), nil, "hive:bob", true, uint(1_000_000_000), "")

	// mint NFT with 10 editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|10|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// transfer different editions to different users
	CallContract(t, ct, "nft_transfer", []byte("0|0|hive:alice_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|1|hive:bob_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|2|hive:alice_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|3|hive:bob_0"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// verify ownership
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:alice"), nil, "hive:alice", true, uint(100_000_000), "0,2")
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:bob"), nil, "hive:bob", true, uint(100_000_000), "1,3")

	// alice and bob can each transfer their own editions
	CallContract(t, ct, "nft_transfer", []byte("0|0|hive:bob_0"), nil, "hive:alice", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|1|hive:alice_0"), nil, "hive:bob", true, uint(1_000_000_000), "")

	// verify swapped ownership
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:alice"), nil, "hive:alice", true, uint(100_000_000), "2,1")
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:bob"), nil, "hive:bob", true, uint(100_000_000), "3,0")

	// alice cannot transfer bob's edition
	CallContract(t, ct, "nft_transfer", []byte("0|3|hive:owner_0"), nil, "hive:alice", false, uint(1_000_000_000), "msg: only market or owner can transfer")
}

// ==========================
// INVALID PAYLOAD TESTS
// ==========================

func TestEmptyPayloads(t *testing.T) {
	ct := SetupContractTest()

	// empty collection create
	CallContract(t, ct, "col_create", []byte(""), nil, "hive:owner", false, uint(1_000_000_000), "msg: empty payload")

	// empty mint
	CallContract(t, ct, "nft_mint", []byte(""), nil, "hive:owner", false, uint(1_000_000_000), "msg: empty payload")

	// empty transfer
	CallContract(t, ct, "nft_transfer", []byte(""), nil, "hive:owner", false, uint(1_000_000_000), "msg: empty payload")

	// empty burn
	CallContract(t, ct, "nft_burn", []byte(""), nil, "hive:owner", false, uint(1_000_000_000), "msg: empty id")

	// empty getters
	CallContract(t, ct, "nft_get", []byte(""), nil, "hive:owner", false, uint(100_000_000), "msg: empty id")
	CallContract(t, ct, "col_get", []byte(""), nil, "hive:owner", false, uint(100_000_000), "msg: empty payload")
}

func TestInvalidNumericPayloads(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint with invalid edition count (non-numeric)
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|abc|meta"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid uint64")

	// get NFT with invalid ID
	CallContract(t, ct, "nft_get", []byte("abc"), nil, "hive:owner", false, uint(100_000_000), "msg: invalid uint64")

	// transfer with invalid NFT ID
	CallContract(t, ct, "nft_transfer", []byte("abc||hive:owner_0"), nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid uint64")
}

func TestMalformedPayloads(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint with missing fields (only 3 pipes instead of 5)
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid payload")

	// collection with missing fields
	CallContract(t, ct, "col_create", []byte("collectionA"), nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid payload")

	// transfer with missing target
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|"), nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid payload")
}

func TestInvalidOwnerCollectionFormat(t *testing.T) {
	ct := SetupContractTest()

	// collection get with no underscore
	CallContract(t, ct, "col_get", []byte("invalidformat"), nil, "hive:owner", false, uint(100_000_000), "msg: invalid owner_collection")

	// collection get with underscore at start
	CallContract(t, ct, "col_get", []byte("_collection"), nil, "hive:owner", false, uint(100_000_000), "msg: invalid owner_collection")

	// collection get with underscore at end
	CallContract(t, ct, "col_get", []byte("owner_"), nil, "hive:owner", false, uint(100_000_000), "msg: invalid owner_collection")
}

// ==========================
// MARKET CONTRACT TESTS
// ==========================

func TestMarketContractValidation(t *testing.T) {
	ct := SetupContractTest()

	// too short market address
	CallContract(t, ct, "add_market", []byte("ab"), nil, "hive:contractowner", false, uint(1_000_000_000), "msg: invalid market address: too short")

	// market address with pipe delimiter
	CallContract(t, ct, "add_market", []byte("market|contract"), nil, "hive:contractowner", false, uint(1_000_000_000), "msg: invalid market address: contains delimiter")

	// valid market address
	CallContract(t, ct, "add_market", []byte("validmarket"), nil, "hive:contractowner", true, uint(1_000_000_000), "")

	// non-owner cannot add market
	CallContract(t, ct, "add_market", []byte("anothermarket"), nil, "hive:randomuser", false, uint(1_000_000_000), "msg: only contract owner can set market")

	// duplicate market (should silently succeed/no-op)
	CallContract(t, ct, "add_market", []byte("validmarket"), nil, "hive:contractowner", true, uint(1_000_000_000), "")

	// verify only one market
	CallContract(t, ct, "get_markets", []byte(""), nil, "hive:anyone", true, uint(100_000_000), "validmarket")
}

func TestMarketContractRemoval(t *testing.T) {
	ct := SetupContractTest()

	// add markets
	CallContract(t, ct, "add_market", []byte("market1"), nil, "hive:contractowner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "add_market", []byte("market2"), nil, "hive:contractowner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "add_market", []byte("market3"), nil, "hive:contractowner", true, uint(1_000_000_000), "")

	CallContract(t, ct, "get_markets", []byte(""), nil, "hive:anyone", true, uint(100_000_000), "market1|market2|market3")

	// remove middle market
	CallContract(t, ct, "remove_market", []byte("market2"), nil, "hive:contractowner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "get_markets", []byte(""), nil, "hive:anyone", true, uint(100_000_000), "market1|market3")

	// non-owner cannot remove
	CallContract(t, ct, "remove_market", []byte("market1"), nil, "hive:randomuser", false, uint(1_000_000_000), "msg: only contract owner can remove market")

	// remove non-existent market
	CallContract(t, ct, "remove_market", []byte("nonexistent"), nil, "hive:contractowner", false, uint(1_000_000_000), "msg: market not found")
}

// ==========================
// SOULBOUND EDITION TESTS
// ==========================

func TestSoulboundMultiEdition(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:creator", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:recipient", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionC|desc|"), nil, "hive:thirdparty", true, uint(1_000_000_000), "")

	// mint soulbound NFT with multiple editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:creator_0|soulbound|desc|true|5|meta"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// creator can transfer edition 0 once
	CallContract(t, ct, "nft_transfer",
		[]byte("0|0|hive:recipient_0"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// recipient cannot transfer that edition further
	CallContract(t, ct, "nft_transfer",
		[]byte("0|0|hive:thirdparty_0"),
		nil, "hive:recipient", false, uint(1_000_000_000), "msg: nft bound to owner")

	// creator can still transfer other editions
	CallContract(t, ct, "nft_transfer",
		[]byte("0|1|hive:recipient_0"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// verify ownership
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:recipient"), nil, "hive:recipient", true, uint(100_000_000), "0,1")
}

// ==========================
// BURN EDGE CASES
// ==========================

func TestBurnSingleEditionNFT(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint single edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// burn without edition index (should work for single edition)
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:owner", true, uint(100_000_000), "true")

	// cannot transfer burned NFT
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:other", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0||hive:other_0"), nil, "hive:owner", false, uint(1_000_000_000), "msg: nft is burned")
}

func TestBurnMultiEditionRequiresIndex(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint multi edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// burn without edition index should fail
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:owner", false, uint(1_000_000_000), "msg: edition required to burn multi-edition NFT")

	// burn with valid edition index works
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|2"), nil, "hive:owner", true, uint(100_000_000), "true")
}

// ====================================
// EDITION BURN AUTHORIZATION TESTS
// ====================================

func TestEditionBurnOnlyByEditionOwner(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:alice", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:bob", true, uint(1_000_000_000), "")

	// alice mints multi-edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:alice_0|name|desc|false|5|meta"),
		nil, "hive:alice", true, uint(1_000_000_000), "")

	// alice transfers edition 2 to bob
	CallContract(t, ct, "nft_transfer", []byte("0|2|hive:bob_0"), nil, "hive:alice", true, uint(1_000_000_000), "")

	// verify bob owns edition 2
	CallContract(t, ct, "nft_hasNFTEdition", []byte("0,hive:bob"), nil, "hive:bob", true, uint(100_000_000), "2")

	// alice tries to burn bob's edition 2 (should FAIL - alice no longer owns it)
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:alice", false, uint(1_000_000_000), "msg: only owner can burn")

	// bob can burn his own edition
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:bob", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|2"), nil, "hive:bob", true, uint(100_000_000), "true")

	// alice can still burn editions she owns
	CallContract(t, ct, "nft_burn", []byte("0|0"), nil, "hive:alice", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isBurned", []byte("0|0"), nil, "hive:alice", true, uint(100_000_000), "true")
}

func TestBurnAfterMultipleTransfers(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:alice", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:bob", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionC|desc|"), nil, "hive:charlie", true, uint(1_000_000_000), "")

	// alice mints multi-edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:alice_0|name|desc|false|3|meta"),
		nil, "hive:alice", true, uint(1_000_000_000), "")

	// alice -> bob -> charlie for edition 1
	CallContract(t, ct, "nft_transfer", []byte("0|1|hive:bob_0"), nil, "hive:alice", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_transfer", []byte("0|1|hive:charlie_0"), nil, "hive:bob", true, uint(1_000_000_000), "")

	// alice cannot burn (original minter)
	CallContract(t, ct, "nft_burn", []byte("0|1"), nil, "hive:alice", false, uint(1_000_000_000), "msg: only owner can burn")

	// bob cannot burn (previous owner)
	CallContract(t, ct, "nft_burn", []byte("0|1"), nil, "hive:bob", false, uint(1_000_000_000), "msg: only owner can burn")

	// charlie can burn (current owner)
	CallContract(t, ct, "nft_burn", []byte("0|1"), nil, "hive:charlie", true, uint(1_000_000_000), "")
}

// ====================================
// ZERO EDITIONS VALIDATION TEST
// ====================================

func TestMintZeroEditionsFails(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint with 0 editions should fail
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|0|meta"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: editions must be at least 1")
}

// ====================================
// NFT COUNT GETTER TEST
// ====================================

func TestNFTCount(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// initially 0 NFTs
	CallContract(t, ct, "nft_count", []byte(""), nil, "hive:anyone", true, uint(100_000_000), "0")

	// mint first NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name1|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_count", []byte(""), nil, "hive:anyone", true, uint(100_000_000), "1")

	// mint second NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name2|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_count", []byte(""), nil, "hive:anyone", true, uint(100_000_000), "2")

	// mint third NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name3|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_count", []byte(""), nil, "hive:anyone", true, uint(100_000_000), "3")
}

func TestNFTCurrentSupply(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint single edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|single|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint multi-edition NFT with 5 editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|multi|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// check initial supply equals total for single edition
	CallContract(t, ct, "nft_supply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "1")
	CallContract(t, ct, "nft_currentSupply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "1")

	// check initial supply equals total for multi-edition
	CallContract(t, ct, "nft_supply", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "5")
	CallContract(t, ct, "nft_currentSupply", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "5")

	// burn the single edition NFT
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	// supply stays same, current supply decreases
	CallContract(t, ct, "nft_supply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "1")
	CallContract(t, ct, "nft_currentSupply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "0")

	// burn edition 0 of multi-edition NFT
	CallContract(t, ct, "nft_burn", []byte("1|0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_supply", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "5")
	CallContract(t, ct, "nft_currentSupply", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "4")

	// burn another edition
	CallContract(t, ct, "nft_burn", []byte("1|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_currentSupply", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "3")

	// burn all remaining editions
	CallContract(t, ct, "nft_burn", []byte("1|1"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_burn", []byte("1|3"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_burn", []byte("1|4"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_currentSupply", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "0")
}

// ====================================
// DOUBLE BURN COUNT FIX TEST
// ====================================

func TestDoubleBurnDoesNotDoubleDecrement(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint multi-edition NFT with 5 editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// verify initial supply
	CallContract(t, ct, "nft_currentSupply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "5")

	// burn edition 2
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_currentSupply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "4")

	// burn edition 2 again (idempotent - should NOT decrement again)
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_currentSupply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "4")

	// burn again to be sure
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_currentSupply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "4")

	// burn another edition to verify counting still works
	CallContract(t, ct, "nft_burn", []byte("0|3"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_currentSupply", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "3")
}

// ====================================
// NFT EXISTS GETTER TEST
// ====================================

func TestNFTExists(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// check non-existent NFT
	CallContract(t, ct, "nft_exists", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "false")
	CallContract(t, ct, "nft_exists", []byte("999"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// mint NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// check existing NFT
	CallContract(t, ct, "nft_exists", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "true")

	// check still non-existent
	CallContract(t, ct, "nft_exists", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// mint another
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name2|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	CallContract(t, ct, "nft_exists", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "true")
}

// ====================================
// NFT IS ALIVE GETTER TEST
// ====================================

func TestNFTIsAlive(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// check non-existent NFT
	CallContract(t, ct, "nft_isAlive", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// mint single-edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|single|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint multi-edition NFT with 5 editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|multi|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// single edition alive
	CallContract(t, ct, "nft_isAlive", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "true")

	// multi-edition alive (various editions)
	CallContract(t, ct, "nft_isAlive", []byte("1|0"), nil, "hive:anyone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_isAlive", []byte("1|4"), nil, "hive:anyone", true, uint(100_000_000), "true")

	// out of range edition returns false (not abort)
	CallContract(t, ct, "nft_isAlive", []byte("1|5"), nil, "hive:anyone", true, uint(100_000_000), "false")
	CallContract(t, ct, "nft_isAlive", []byte("1|100"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// burn single edition
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isAlive", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// burn multi-edition
	CallContract(t, ct, "nft_burn", []byte("1|2"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_isAlive", []byte("1|2"), nil, "hive:anyone", true, uint(100_000_000), "false")
	// other editions still alive
	CallContract(t, ct, "nft_isAlive", []byte("1|0"), nil, "hive:anyone", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_isAlive", []byte("1|3"), nil, "hive:anyone", true, uint(100_000_000), "true")
}

// ====================================
// ADDITIONAL COVERAGE TESTS
// ====================================

func TestTransferToNonExistentCollection(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// try to transfer to non-existent collection (should fail)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:other_0"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: collection not found")
}

func TestMarketTransferRegularEdition(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "add_market", []byte("vscmarket"), nil, "hive:contractowner", true, uint(1_000_000_000), "")

	// create collections
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:seller", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:buyer", true, uint(1_000_000_000), "")

	// mint regular (non-soulbound) NFT with editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:seller_0|name|desc|false|5|meta"),
		nil, "hive:seller", true, uint(1_000_000_000), "")

	// market can transfer regular edition
	CallContract(t, ct, "nft_transfer",
		[]byte("0|2|hive:buyer_0"),
		nil, "vscmarket", true, uint(1_000_000_000), "")

	// verify buyer owns edition 2
	CallContract(t, ct, "nft_isOwner", []byte("0|2"), nil, "hive:buyer", true, uint(100_000_000), "true")
}

func TestBurnedEditionTransferMultiEdition(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:other", true, uint(1_000_000_000), "")

	// mint NFT with editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// burn edition 2
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// try to transfer burned edition (should fail)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|2|hive:other_0"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: edition is burned")
}

func TestMarketCannotBurn(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "add_market", []byte("vscmarket"), nil, "hive:contractowner", true, uint(1_000_000_000), "")

	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// market tries to burn (should fail - only owner can burn)
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "vscmarket", false, uint(1_000_000_000), "msg: only owner can burn")

	// verify not burned
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "false")
}

func TestSoulboundCollectionOnlyTransferAfterBound(t *testing.T) {
	ct := SetupContractTest()
	// create collections for recipient
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:creator", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:recipient", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionC|desc|"), nil, "hive:recipient", true, uint(1_000_000_000), "")

	// mint soulbound NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:creator_0|soulbound|desc|true|1|meta"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// creator transfers to recipient
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:recipient_0"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// recipient CAN move soulbound NFT between their own collections (same owner)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:recipient_1"),
		nil, "hive:recipient", true, uint(1_000_000_000), "")

	// verify moved to collection 1
	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:recipient", true, uint(100_000_000), "hive:recipient_1")
}

func TestSoulboundCreatorCannotBurnAfterTransfer(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:creator", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:recipient", true, uint(1_000_000_000), "")

	// mint soulbound NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:creator_0|soulbound|desc|true|1|meta"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// creator transfers to recipient
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:recipient_0"),
		nil, "hive:creator", true, uint(1_000_000_000), "")

	// creator tries to burn (should fail - no longer owner)
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:creator", false, uint(1_000_000_000), "msg: only owner can burn")

	// recipient can burn their own soulbound NFT
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:recipient", true, uint(1_000_000_000), "")
}

func TestNameAtExactBoundary(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// name exactly 48 chars (should succeed)
	name48 := "123456789012345678901234567890123456789012345678"
	payload := []byte("hive:owner_0|" + name48 + "|desc|false|1|meta")
	CallContract(t, ct, "nft_mint", payload, nil, "hive:owner", true, uint(1_000_000_000), "")

	// name 49 chars (should fail)
	name49 := "1234567890123456789012345678901234567890123456789"
	payload2 := []byte("hive:owner_0|" + name49 + "|desc|false|1|meta")
	CallContract(t, ct, "nft_mint", payload2, nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid name length")
}

func TestDescriptionAtExactBoundary(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// description exactly 128 chars (should succeed)
	desc128 := "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678"
	payload := []byte("hive:owner_0|name|" + desc128 + "|false|1|meta")
	CallContract(t, ct, "nft_mint", payload, nil, "hive:owner", true, uint(1_000_000_000), "")

	// description 129 chars (should fail)
	desc129 := "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	payload2 := []byte("hive:owner_0|name2|" + desc129 + "|false|1|meta")
	CallContract(t, ct, "nft_mint", payload2, nil, "hive:owner", false, uint(1_000_000_000), "msg: description too long")
}

func TestEmptyNameFails(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// empty name (should fail)
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0||desc|false|1|meta"),
		nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid name length")
}

func TestGetNonExistentCollection(t *testing.T) {
	ct := SetupContractTest()

	// try to get collection that doesn't exist
	CallContract(t, ct, "col_get", []byte("hive:owner_0"), nil, "hive:anyone", false, uint(100_000_000), "msg: collection not found")

	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|desc|meta"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// now it exists
	CallContract(t, ct, "col_get", []byte("hive:owner_0"), nil, "hive:anyone", true, uint(100_000_000), "")

	// but collection 1 doesn't
	CallContract(t, ct, "col_get", []byte("hive:owner_1"), nil, "hive:anyone", false, uint(100_000_000), "msg: collection not found")
}

func TestTwoUsersWithSameCollectionName(t *testing.T) {
	ct := SetupContractTest()

	// alice creates collection named "MyCollection"
	CallContract(t, ct, "col_create", []byte("MyCollection|desc|"), nil, "hive:alice", true, uint(1_000_000_000), "")

	// bob creates collection with same name - should work (different owner)
	CallContract(t, ct, "col_create", []byte("MyCollection|desc|"), nil, "hive:bob", true, uint(1_000_000_000), "")

	// both collections exist and are independent
	CallContract(t, ct, "col_get", []byte("hive:alice_0"), nil, "hive:anyone", true, uint(100_000_000), "")
	CallContract(t, ct, "col_get", []byte("hive:bob_0"), nil, "hive:anyone", true, uint(100_000_000), "")

	// alice can mint to her collection
	CallContract(t, ct, "nft_mint",
		[]byte("hive:alice_0|name|desc|false|1|meta"),
		nil, "hive:alice", true, uint(1_000_000_000), "")

	// bob can mint to his collection
	CallContract(t, ct, "nft_mint",
		[]byte("hive:bob_0|name|desc|false|1|meta"),
		nil, "hive:bob", true, uint(1_000_000_000), "")

	// verify ownership
	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "hive:alice_0")
	CallContract(t, ct, "nft_ownerColOf", []byte("1"), nil, "hive:anyone", true, uint(100_000_000), "hive:bob_0")
}

func TestGetBurnedNFT(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|myname|mydesc|false|1|mymeta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// burn it
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// nft_get should still work (returns core data, burn state is separate)
	CallContract(t, ct, "nft_get", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "")

	// nft_meta should still return metadata
	CallContract(t, ct, "nft_meta", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "mymeta")

	// nft_exists should return true (NFT exists, just burned)
	CallContract(t, ct, "nft_exists", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "true")

	// nft_isAlive should return false
	CallContract(t, ct, "nft_isAlive", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// nft_isBurned should return true
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "true")
}

func TestEdition0ExplicitVsImplicit(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "col_create", []byte("collectionB|desc|"), nil, "hive:other", true, uint(1_000_000_000), "")

	// mint single-edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|single|desc|false|1|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint multi-edition NFT
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|multi|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// Single-edition: implicit edition 0
	CallContract(t, ct, "nft_isOwner", []byte("0"), nil, "hive:owner", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "hive:owner_0")
	CallContract(t, ct, "nft_isBurned", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// Multi-edition: explicit edition 0 should work
	CallContract(t, ct, "nft_isOwner", []byte("1|0"), nil, "hive:owner", true, uint(100_000_000), "true")
	CallContract(t, ct, "nft_ownerColOf", []byte("1|0"), nil, "hive:anyone", true, uint(100_000_000), "hive:owner_0")
	CallContract(t, ct, "nft_isBurned", []byte("1|0"), nil, "hive:anyone", true, uint(100_000_000), "false")

	// Multi-edition: implicit (no edition) for nft_get should fail (edition required)
	CallContract(t, ct, "nft_get", []byte("1"), nil, "hive:anyone", false, uint(100_000_000), "msg: edition required for multi-edition NFT")
	// Multi-edition: explicit edition 0 for nft_get should work
	CallContract(t, ct, "nft_get", []byte("1|0"), nil, "hive:anyone", true, uint(100_000_000), "")

	// Transfer single edition (implicit)
	CallContract(t, ct, "nft_transfer", []byte("0||hive:other_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_ownerColOf", []byte("0"), nil, "hive:anyone", true, uint(100_000_000), "hive:other_0")

	// Transfer multi-edition edition 0 (explicit)
	CallContract(t, ct, "nft_transfer", []byte("1|0|hive:other_0"), nil, "hive:owner", true, uint(1_000_000_000), "")
	CallContract(t, ct, "nft_ownerColOf", []byte("1|0"), nil, "hive:anyone", true, uint(100_000_000), "hive:other_0")
}

func TestCollectionNameBoundary(t *testing.T) {
	ct := SetupContractTest()

	// name exactly 48 chars for collection (should succeed)
	name48 := "123456789012345678901234567890123456789012345678"
	payload := []byte(name48 + "|desc|meta")
	CallContract(t, ct, "col_create", payload, nil, "hive:owner", true, uint(1_000_000_000), "")

	// name 49 chars for collection (should fail)
	name49 := "1234567890123456789012345678901234567890123456789"
	payload2 := []byte(name49 + "|desc|meta")
	CallContract(t, ct, "col_create", payload2, nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid name length")
}

func TestCollectionDescriptionBoundary(t *testing.T) {
	ct := SetupContractTest()

	// description exactly 128 chars for collection (should succeed)
	desc128 := "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678"
	payload := []byte("colname|" + desc128 + "|meta")
	CallContract(t, ct, "col_create", payload, nil, "hive:owner", true, uint(1_000_000_000), "")

	// description 129 chars for collection (should fail)
	desc129 := "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"
	payload2 := []byte("colname2|" + desc129 + "|meta")
	CallContract(t, ct, "col_create", payload2, nil, "hive:owner", false, uint(1_000_000_000), "msg: description too long")
}

func TestEmptyCollectionName(t *testing.T) {
	ct := SetupContractTest()

	// empty collection name (should fail)
	CallContract(t, ct, "col_create", []byte("|desc|meta"), nil, "hive:owner", false, uint(1_000_000_000), "msg: invalid name length")
}

func TestMarketCannotBurnEdition(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "add_market", []byte("vscmarket"), nil, "hive:contractowner", true, uint(1_000_000_000), "")

	CallContract(t, ct, "col_create", []byte("collectionA|desc|"), nil, "hive:owner", true, uint(1_000_000_000), "")

	// mint NFT with editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:owner_0|name|desc|false|5|meta"),
		nil, "hive:owner", true, uint(1_000_000_000), "")

	// market tries to burn edition (should fail - only owner can burn)
	CallContract(t, ct, "nft_burn", []byte("0|2"), nil, "vscmarket", false, uint(1_000_000_000), "msg: only owner can burn")

	// verify not burned
	CallContract(t, ct, "nft_isBurned", []byte("0|2"), nil, "hive:anyone", true, uint(100_000_000), "false")
}
