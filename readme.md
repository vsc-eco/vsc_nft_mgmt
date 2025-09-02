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
    │   └── collections.go // functions for creating and getting collections
    │   └── helpers.go // various handy functions
    │   └── indexing.go // features to maintaining multiple indexes for faster reads of contract state data
    │   └── main.go // placeholder
    │   └── mock_collection_test.go // unit tests all about *collections.go*
    │   └── mock_nft_tests.go // unit tests all about *nfts.go*
    │   └── nfts.go // functions related to nfts like minting, transferring and various getters
    │   └── sdkInterface.go // an interface for enabling the developer to do unit tests without production environment
    ├── runtime/
    │   └── gc_leaking_exported.go
    │   └── mock_placeholder.go // dummy for building mock version
    ├── sdk/ //SDK implementation. Do NOT modify
    │   └── address.go
    │   └── asset.go
    │   └── env.go        
    │   └── sdk_mock.go // sdk definition to enable local unit tests  
    │   └── sdk.go    
    ├── go.mod
    ├── readme.md


## ⚙️ Requirements

-   [Go](https://golang.org/dl/) **1.23.2+**
-   [TinyGo](https://tinygo.org/getting-started/install/)
-   [Wasm Edge](https://wasmedge.org/docs/start/install/)
-   [Wasm Tools](https://github.com/bytecodealliance/wasm-tools)


## 🚀 Build & Deploy

For instructions related to building and deploying see the [official docs - TODO add link](link).


## ✅ Testing

There are unit tests defined for all the important exported function implementations:

``` bash
cd contract
go test -tags=test -v
```
You should see a PASS at the end. If not there is at least one unit tests that failed:
```
...
--- PASS: TestGetAvailableEditionsForNFT_Dynamic (0.00s)
=== RUN   TestGetAvailableEditionsForNFT_Negative
--- PASS: TestGetAvailableEditionsForNFT_Negative (0.00s)
PASS
ok      vsc_nft_mgmt/contract   0.003s

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
  "editions": 10 // mandatory: total number of editions (> 0)
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


#### Collections
##### Get One Collection
Returns a collection.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| col_get     | 123     | mandatory: Collection ID     |


##### Get Collections For Address
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

##### Get Editions for an NFT
Returns all Editions for a given NFT
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_editions | 42 | mandatory: genesis NFT ID |

##### Get Available Editions for an NFT
Returns all Editions for a given NFT that are still owned by the minting address.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_available | 42 | mandatory: genesis NFT ID |

##### Get NFTs for Collection
Returns a list of NFTs within a collection.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_collection | 123 | mandatory: collection ID |

##### Get NFTs for Owner
Returns all NFTs owned by a specified address.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_owner | hive:tibfox | mandatory: owner address |

##### Get NFTs minted by Adress
Returns all NFTs minted by a specified address.
| action | payload  | payload description |
| ------ | -------- | ------------------- |
| nft_get_creator | hive:tibfox | mandatory: creator address |



## 📚 Documentation

-   [Go-VSC-Node](https://github.com/vsc-eco/go-vsc-node)
-   [Go-Contract-Template](https://github.com/vsc-eco/go-contract-template)


## 📜 License

This project is licensed under the [MIT License](LICENSE).