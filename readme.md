# VSC NFT Management Smart Contract

This repository contains a **smart contract written in Go** for the
[VSC-Eco](https://github.com/vsc-eco/) ecosystem.
The contract is designed to integrate seamlessly with the vsc-ecosystem,
enabling various nft related functionalities.


## 📖 Overview

-   **Language:** Go (Golang) 1.23.2+
-   **Purpose:** Provides basic functions to create collections, minting, transferring and burning nfts


## 📖 Schema

Each Adress can have multiple collections. In each collection can be included multiple NFTs.
There are editioned NFTs that are a sets of similar nfts and each edition is not stored as separate nft object.
Every NFT can be transferred and burned no matter if they are unique or editions.

- Collection (Id 123)
    - unique NFT (Id 42)
    - Editioned NFT (Id 43)
        - Edition 0 (Id 43:0)
        - Edition 1 (Id 43:1)
        - Edition 2 (Id 43:2)
        - Edition 3 (Id 43:3)
        - ...
    - unique NFT (Id 101)
    - ...
- Collection (Id 124)
    - ...
    - 


## 📂 Project Structure

    ./vsc_nft_mgmt
    ├── artifacts/  //Contains 
    ├── contract/
    │   └── admin.go // administrative functions
    │   └── collections.go // functions for creating and getting collection data
    │   └── events.go // event emitting functions for off-chain indexers
    │   └── helpers.go // various utility functions
    │   └── main.go // placeholder
    │   └── nfts.go // functions related to nfts like minting, transferring and getting nft data
    ├── runtime/
    │   └── gc_leaking_exported.go
    ├── sdk/ //SDK implementation. Do NOT modify
    │   └── address.go
    │   └── asset.go
    │   └── env.go        
    │   └── sdk.go    
    ├── test/
    │   └── basic_test.go // various tests
    │   └── helpers_test.go // helpers for the tests
    ├── go.mod
    ├── go.sum
    ├── readme.md


## ⚙️ Requirements

-   [Go](https://golang.org/dl/) **1.23.2+**
-   [TinyGo](https://tinygo.org/getting-started/install/)
-   [Wasm Edge](https://wasmedge.org/docs/start/install/) **v0.13.4**
-   [Wasm Tools](https://github.com/bytecodealliance/wasm-tools)

## 🚀 Build & Deploy

For instructions related to building and deploying see the [official docs - TODO add link](link).


## ✅ Testing

There are tests defined under `tests/basic_test.go`for all the important exported function implementations.

You can build a testing build with the following command contrary to the official documentation:
`tinygo build -gc=custom -scheduler=none -panic=trap -no-debug -target=wasm-unknown -tags=test -o test/artifacts/main.wasm ./contract`

The tests are designed to run sequencially because the mocking database layer is a single-use-file only.
For running the tests you simply run `go test -p 1 ./test -v` from within the root. 

You should see a PASS at the end. If not there is at least one tests that failed:
```
...
gas used: 8153175
gas max : 10000000
--- PASS: TestExtendEditionNFTs (0.57s)
PASS
ok      vsc_nft_mgmt/test       1.235s
```


## 📖 Exported Functions

Below you can find all exported and usable functions for this smart contract including example payloads.
**Warning:** These payloads contain comments which is invalid for json. Make sure tho remove the comments if you copy&paste payloads to test the contract.

You can use the [Hive Keychain SDK Playground](https://play.hive-keychain.com/#/request/custom) for testing these L1 transactions.
Username: `your Hive username`
Id: `vsc.call`
Method: `Active`
Json: `below payloads`

### Mutations

#### Create Collection
action: `col_create`
Creates a collection for the sending address. 
payload: 
```json5
{
    "name": "Trasure Chest", // mandatory: name of the collection (max 100 characters)
    "description": "All my NFTs" // optional: description of the collection (max 1000 characters)
}
```


#### Mint NFT

action: `nft_mint`
Creates a **unique** or **editioned** NFT.

**unique:**
```json5
{
  "c": "123", // mandatory: target collection ID
  "name": "Golden Sword", // mandatory: name of the NFT
  "desc": "A legendary one-of-a-kind sword", // optional: description
  "bound": false, // optional: true = can be transferred only once from creator (/)defaults to false)
  "meta": { // optional: metadata key-value pairs
    "rarity": "legendary",
    "attack": "150",
    "durability": "unbreakable"
  }
}
```

**editioned with 10 editions:**
```json5
{
  "c": "123", // mandatory: target collection ID
  "name": "Golden Sword", // mandatory: name of the NFT
  "desc": "A legendary one-of-a-kind sword", // optional: description
  "bound": false, // optional: true = can be transferred only once from creator (/)defaults to false)
  "meta": { // optional: metadata key-value pairs
    "rarity": "legendary",
    "attack": "150",
    "durability": "unbreakable"
  },
  "et":10
}
```

#### Transfer NFT
action: `nft_transfer`
Tranfers an **NFT** (edition or unique) to a new collection or a new owner. Only the owner or a administrative defined market contract can move an nft to a new owner. Only owners can move the an NFT to another owned collection.
```json5
{
  "id": "42", // mandatory: NFT ID (string-form ID used in state keys)
  "col": "123", // mandatory: destination collection ID (can be same as current)
  "owner": "hive:tibfox" // mandatory: destination owner address
}
```

### Queries
The following exported functions return json and are meant to be used by other contracts like the market contract for example. Reading data from a contract state outside of the smart contract environment is more cost-effective and faster by utilizing the vsc API and a future indexer.
For that reason there are only sing-object getters defined.


#### Collections
##### Get One Collection
Returns a collection.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| col_get     | 123     | mandatory: Collection ID     |


#### NFTs
##### Get One NFT
Returns an NFT
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get | 42 | mandatory: NFT ID |


## On-Chain Events

The contract outputs standardized event logs for future indexers. This way a UI can quickly show nfts within a collection, available editions for an nft, etc.


| Event Name            | Parameter        | Type     | Description                                           |
| --------------------- | ---------------- | -------- | ----------------------------------------------------- |
| Transfer              | `id`             | uint64   | NFT ID being transferred                              |
|                       | `from`           | string   | Address of the current owner                          |
|                       | `to`             | string   | Address of the recipient                              |
|                       | `fromCollection` | uint64   | Collection ID from which the NFT is moved             |
|                       | `toCollection`   | uint64   | Collection ID to which the NFT is moved               |
| Mint                  | `id`             | uint64   | NFT ID being minted                                   |
|                       | `by`             | string   | Address of the minter (creator)                       |
|                       | `to`             | string   | Address receiving the NFT                             |
|                       | `collection`     | uint64   | Collection ID the NFT belongs to                      |
|                       | `genesis`        | \*uint64 | Optional: ID of the genesis NFT if this is an edition |
| Burn                  | `id`             | uint64   | NFT ID being burned                                   |
|                       | `owner`          | string   | Address of the current owner of the NFT               |
|                       | `collection`     | uint64   | Collection ID the NFT belongs to                      |
| CollectionCreated     | `id`             | uint64   | Collection ID that was created                        |
|                       | `by`             | string   | Address of the creator of the collection              |



## 📚 Documentation
-   [Go-VSC-Node](https://github.com/vsc-eco/go-vsc-node)
-   [Go-Contract-Template](https://github.com/vsc-eco/go-contract-template)

## 📜 License
This project is licensed under the [MIT License](LICENSE).