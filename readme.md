# VSC NFT Management Smart Contract

This repository contains a **smart contract written in Go** for the
[VSC-Eco](https://github.com/vsc-eco/) ecosystem.
The contract is designed to integrate seamlessly with the vsc-ecosystem,
enabling various nft related functionalities.


## 📖 Overview

-   **Language:** Go (Golang) 1.23.2+
-   **Purpose:** Provides basic functions to create collections and minting nfts


## 📖 Schema

Each Adress can have multiple collections. In each collection can be included multiple NFTs. There are Edition NFTs that are a series of nfts. The first one of these series are called "genesis edition". Every NFT can be transferred no matter if they are unique, genesis or editions.

- Collection (Id 123)
    - unique NFT (Id 42)
    - Edition NFTs aka "Genesis Edition" (Id 43)
        - Edition 2 (Id 44)
        - Edition 3 (Id 45)
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
    │   └── getters_tests.go // exported getter functions for testing only
    │   └── helpers.go // various utility functions
    │   └── indexing.go // features to maintaining multiple indexes for faster reads of contract state data
    │   └── main.go // placeholder
    │   └── nfts.go // functions related to nfts like minting, transferring and getting nft data
    ├── runtime/
    │   └── gc_leaking_exported.go
    ├── sdk/ //SDK implementation. Do NOT modify
    │   └── address.go
    │   └── asset.go
    │   └── env.go        
    │   └── sdk.go    
    ├── go.mod
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
This file also uses the getters of the `contract/getters_tests.go` for visualization of the contract state.

You can build a testing build (including the various getter functions) with the following command contrary to the official documentation:
`tinygo build -gc=custom -scheduler=none -panic=trap -no-debug -target=wasm-unknown -tags=test -o test/artifacts/main.wasm ./contract`

The tests are designed to run sequencially because the mocking database layer is a single-use-file only.
For running the tests you imply run `go test -p 1 ./test -v` from within the root. 

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
    "name": "Trasure Chest", // mandatory: name of the collection
    "description": "All my NFTs" // optional: description of the collection
}
```


#### Mint Unique NFT

action: `nft_mint_unique`
Creates a **unique** NFT.

```json5
{
  "col": "123", // mandatory: target collection ID
  "name": "Golden Sword", // mandatory: name of the NFT
  "desc": "A legendary one-of-a-kind sword", // optional: description
  "bound": false, // optional: true = can be transferred only once from creator
  "meta": { // optional: metadata key-value pairs
    "rarity": "legendary",
    "attack": "150",
    "durability": "unbreakable"
  }
}
```

#### Mint Editioned NFT
action: `nft_mint_edition`
Creates **editions** of NFTs.
```json5
{
  "col": "123", // mandatory: target collection ID
  "name": "Silver Shield", // mandatory: name of the NFT
  "desc": "Limited edition silver shield", // optional: description stored only on genesis
  "bound": false, // optional: true = can be transferred only once from creator 
  "meta": { // optional: metadata (applied to genesis to avoid redundancy)
    "rarity": "rare",
    "defense": "200"
  },
  "editions": 10, // mandatory: total number of editions (> 0) max 100
  "g": 10, // optional: genesis edition if the series should get extended

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
The following exported functions return json and are meant to be used by other contracts like the market contract for example. Reading data from a contract state outside of the smart contract environment is more cost-effective and faster by utilizing the vsc API. (TODO: add link to doc part about reading key/value from contract state)
For that reason most of the getters below are only incuded in the test build (see above).


#### Collections
##### Get One Collection
Returns a collection.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| col_get     | 123     | mandatory: Collection ID     |


##### Get Collections For Address (only available in test build)
Returns all collections for a give address.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| col_get_user | hive:tibfox | mandatory: owner address |

#### NFTs
##### Get One NFT
Returns an NFT
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get | 42 | mandatory: NFT ID |

##### Get Next Available Edition for an NFT
Returns all Editions for a given NFT
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_available | 42 | mandatory: genesis NFT ID |


##### Get Editions for an NFT (only available in test build)
Returns all Editions for a given NFT
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_editions | 42 | mandatory: genesis NFT ID |

##### Get Available Editions for an NFT (only available in test build)
Returns all Editions for a given NFT that are still owned by the minting address.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_availableList | 42 | mandatory: genesis NFT ID |

##### Get NFTs for Collection (only available in test build)
Returns a list of NFTs within a collection.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_collection | 123 | mandatory: collection ID |

##### Get NFTs for Owner (only available in test build)
Returns all NFTs owned by a specified address.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_owner | hive:tibfox | mandatory: owner address |

##### Get NFTs minted by Adress (only available in test build)
Returns all NFTs minted by a specified address.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_creator | hive:tibfox | mandatory: creator address |



## 📚 Documentation

-   [Go-VSC-Node](https://github.com/vsc-eco/go-vsc-node)
-   [Go-Contract-Template](https://github.com/vsc-eco/go-contract-template)


## 📜 License

This project is licensed under the [MIT License](LICENSE).