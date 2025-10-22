package contract_test

import (
	"testing"
)

// // // admin tests
func TestAdminMarket(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "set_market", []byte("hive:tibfox"), nil, "hive:tibfox", false, uint(10_000))
	CallContract(t, ct, "set_market", []byte(""), nil, "hive:someone", false, uint(1_000))
	CallContract(t, ct, "set_market", []byte("hive:tibfox"), nil, "hive:contractowner", true, uint(100_000_000))
	CallContract(t, ct, "get_market", []byte(""), nil, "hive:contractowner", true, uint(100_000_000))

}

// // collection tests

func TestColCreate(t *testing.T) {
	ct := SetupContractTest()
	// just create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))
	CallContract(t, ct, "col_create", []byte("collectionB||"), nil, "hive:someone", true, uint(1_000_000_000))
	CallContract(t, ct, "col_get", []byte("hive:someone_0"), nil, "hive:someone", true, uint(100_000_000))
	CallContract(t, ct, "col_get", []byte("hive:someone_1"), nil, "hive:someone", true, uint(100_000_000))
}

func TestColCreateFails(t *testing.T) {
	// just create a collection
	CallContract(t, SetupContractTest(), "col_create",
		[]byte("|my description|img=testurl"),
		nil, "hive:someone", false, uint(100_000_000))

	CallContract(t, SetupContractTest(), "col_create",
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e|my description|img=testurl"),
		nil, "hive:someone", false, uint(100_000_000))

	CallContract(t, SetupContractTest(), "col_create",
		[]byte("collectionA|Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.|img=testurl"),
		nil, "hive:someone", false, uint(100_000_000))
}

func TestMintUniqueNFTSinglew(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000))
	// mint without edition value
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true||test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))
}

// // nft tests
func TestMintUniqueNFT(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000))
	// mint without edition value
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true||test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// mint nft with edition value 0
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// mint nft with edition value 1
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// mint nft to another collection
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someoneelse_1|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// get minted nft
	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:someone", true, uint(1_000_000_000))
	CallContract(t, ct, "nft_get", PayloadToJSON("3"), nil, "hive:someone", true, uint(1_000_000_000))

}

// mint editions
func TestMintEditions(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))

	// mint nft editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|10|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// get minted nft 0
	CallContract(t, ct, "nft_get", PayloadToJSON("0"), nil, "hive:someone", true, uint(1_000_000_000))

	// get minted nft 0 - edition 9
	CallContract(t, ct, "nft_get", PayloadToJSON("0|9"), nil, "hive:someone", true, uint(1_000_000_000))
}

func TestMintUniqueNFTFails(t *testing.T) {
	ct := SetupContractTest()
	// // create a collection for minter
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))

	// // mint nft with character overflows
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore e|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", false, uint(100_000_000))

	// // // character overflows

	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|nft name|Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.|true|0|test=123,test2=abc"),
		nil, "hive:someone", false, uint(100_000_000))

	// // mint nft with wrong param count
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|asd"),
		nil, "hive:someone", false, uint(100_000_000))

	// // mint nft without name
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0||description|true|0|test=123"),
		nil, "hive:someone", false, uint(100_000_000))

	// // mint nft without collection
	CallContract(t, ct, "nft_mint",
		[]byte("|asd|description|true|0|test=123"),
		nil, "hive:someone", false, uint(100_000_000))

}

// burn
func TestBurn(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))
	// burn it
	CallContract(t, ct, "nft_burn", []byte("0"), nil, "hive:someone", true, uint(100_000_000))

	// set market
	CallContract(t, ct, "set_market", []byte("hive:marketaddress"), nil, "hive:contractowner", true, uint(1_000_000_000))
	// mint
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|0|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// burn it by market
	CallContract(t, ct, "nft_burn", []byte("1"), nil, "hive:marketaddress", true, uint(1_000_000_000))

}

func TestBurnEdition(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(1_000_000_000))
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))
	// mint 10 nft editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|10|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// burn nft 0 edition 0 (is default for editioned nfts) - should succeed
	CallContract(t, ct, "nft_burn", PayloadToJSON("0"), nil, "hive:someone", true, uint(1_000_000_000))

	// try to burn edition - should succeed
	CallContract(t, ct, "nft_burn", PayloadToJSON("0|9"), nil, "hive:someone", true, uint(1_000_000_000))

}

func TestBurnFails(t *testing.T) {
	ct := SetupContractTest()
	// create a collection
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))
	// mint 1 unique nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))
	// mint 10 nft editions
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|true|10|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))
	// burn edition for unique nft
	CallContract(t, ct, "nft_burn",
		[]byte("0|2"),
		nil, "hive:someone", false, uint(100_000_000))
	// burn it by other user
	CallContract(t, ct, "nft_burn",
		[]byte("0|"),
		nil, "hive:otheruser", false, uint(100_000_000))
	// try to burn edition 2 by other user - should fail
	CallContract(t, ct, "nft_burn",
		[]byte("1|2"),
		nil, "hive:someoneelse", false, uint(1_000_000_000))
}

func TestTransfers(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(1_000_000_000))
	// create a collection for sender
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// transfer nft by minter (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someoneelse_1"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// transfer nft by minter (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someone_0"),
		nil, "hive:marketaddress", true, uint(1_000_000_000))
}

// test transfers with excepted fails
func TestTransfersFails(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(1_000_000_000))
	// create a collection for sender
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|1|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// transfer nft by other user than minter
	CallContract(t, ct, "nft_transfer",
		[]byte("0||hive:someoneelse_1"),
		nil, "hive:someoneelse", false, uint(100_000_000))

	// transfer non existing nft
	CallContract(t, ct, "nft_transfer",
		[]byte("1||hive:someone_0"),
		nil, "hive:someone", false, uint(100_000_000))

	// transfer nft to non existing collection
	CallContract(t, ct, "nft_transfer",
		[]byte("1||hive:someoneelse_0"),
		nil, "hive:someone", false, uint(100_000_000))

	// transfer nft without collection
	CallContract(t, ct, "nft_transfer",
		[]byte("1||"),
		nil, "hive:someone", false, uint(100_000_000))

	// transfer nft with non existing nft id
	CallContract(t, ct, "nft_transfer",
		[]byte("0|2|hive:someone_0"),
		nil, "hive:someone", false, uint(100_000_000))

}

// edition transfers
func TestEditionTransfers(t *testing.T) {
	ct := SetupContractTest()
	// set market
	CallContract(t, ct, "set_market", PayloadToJSON("hive:marketaddress"), nil, "hive:contractowner", true, uint(1_000_000_000))
	// create a collection for sender
	CallContract(t, ct, "col_create", []byte("collectionA|my description|img=testurl"), nil, "hive:someone", true, uint(1_000_000_000))

	// create a collection for receiver
	CallContract(t, ct, "col_create", []byte("collectionB|my description|img=testurl"), nil, "hive:someoneelse", true, uint(1_000_000_000))

	// mint nft
	CallContract(t, ct, "nft_mint",
		[]byte("hive:someone_0|name|description|false|999999|test=123,test2=abc"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// transfer edition no 3 nft (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|3|hive:someoneelse_1"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// transfer edition no 999999 nft (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|99999|hive:someoneelse_1"),
		nil, "hive:someone", true, uint(1_000_000_000))

	// transfer back edition no 999999 nft (should success)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|99999|hive:someone_0"),
		nil, "hive:someoneelse", true, uint(1_000_000_000))

	// // maliciou trying to transfer edition no 5000 not by owner (should fail)
	CallContract(t, ct, "nft_transfer",
		[]byte("0|5000|hive:someoneelse_1"),
		nil, "hive:someoneelse", false, uint(1_000_000_000))
}
