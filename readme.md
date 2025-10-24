# VSC NFT Management Smart Contract

This repository contains a **high-performance NFT management smart contract** written in **TinyGo**, optimized for the **VSC blockchain**.

It enables efficient minting, transfer, and burning of NFTs-supporting both **unique NFTs** and **multi-edition NFTs**, with **gas-optimized storage and event indexing** and organizing owned NFTs in different user defined **collections**.



## 🚀 Key Features

| Feature | Description |
| - |- |
| **Collections**        | Each user can create multiple collections, uniquely indexed by `<owner>_<collectionIndex>` |
| **NFT Minting**        | Supports both unique NFTs (single-edition) and multi-edition NFTs |
| **Edition Logic**      | Editions do not store full NFT copies - only overrides when changed |
| **Transfers**          | Owner-to-owner transfers and intra-owner collection transfers |
| **Burning**            | burning of unique NFTs and edition NFTs without touching the NFT objects themselves |
| **Market Integration** | Multiple external marketplace contract can be authorized to execute transfers. There are various exported getter functions defined to support an easy integration. |
| **Low Gas Design**     | Fully manual state encoding, no JSON or reflection overhead. Simple, fast and gas-effective. |



## 🧠 Data Model Overview

```
Account: hive:alice
 └── Collection 0  ("hive:alice_0")
 │     ├── NFT 1001 (unique)
 │     ├── NFT 1002 (editions=3)
 │     │      ├── Edition 0 → owner: hive:alice_0
 │     │      ├── Edition 1 → owner: hive:bob_1  (override)
 │     │      └── Edition 2 → burned             (override)
 │     └── NFT 1003 (unique)
 └── Collection 1  ("hive:alice_1")
       └── NFT 1004 ...
```

### 🔑 Identifier Formats

| Type | Format | Example | Notes |
| - | - | - | - |
| Collection | `"owner_colIndex"` | `"hive:alice_0"` | Internal ID is numeric, but key is stored using owner+index |
| NFT  | `"nftID"` | `"1002"` | Always numeric string |
| Edition | `"nftID\|edition"` | `"1002\|1"` | Default edition is `0` if omitted (for single-edition NFTs) |



## 🧩 Pipeline Payload Format

All exported functions use **string payloads**, delimited by the `|` character.
This format avoids JSON overhead and minimizes gas.

✅ **Example: Mint Payload**

```
"owner_collection|name|description|singleTransferFlag|editions|metadata"
```

✅ **Example: Transfer Payload**

```
"nftID|editionIndex|targetOwner_collection"
```

✅ **Example: Burn Payload**

```
"nftID"            (burn full base NFT if single-edition)
"nftID|edition"   (burn a specific edition)
```



## ⚙️ Contract Architecture

```
contract/
├── admin.go          # marketplace authorization
├── collections.go    # create collections
├── nfts.go           # mint/transfer/burn NFTs
├── events.go         # event emission
├── getters.go         # all getters for NFT and collection specifics
├── helpers.go        # parsing, binary encoding, state key builders
├── main.go           # placeholder
```



## 🔥 Behavior Summary

| Action   | Multi-edition NFT | Unique NFT |
| - | - | - |
| Transfer | Updates edition override only | Updates base owner entry |
| Burn | Sets burned flag in edition override | Creates owner override and sets burned flag|
| Get | Resolves overrides and returns final real-time ownership state | Direct |




# ⚙️ Build & Deployment Guide

Follow the current official guidelines for smart contract deployment on VSC.



## 🧪 Testing Locally

### 🔧 Build WASM for Tests

```bash
tinygo build \
  -gc=custom \
  -scheduler=none \
  -panic=trap \
  -no-debug \
  -target=wasm-unknown \
  -tags=test \
  -o test/artifacts/main.wasm \
  ./contract
```

### ▶ Run Tests

```bash
go test ./test -v
```


### 🧪 Test Example Execution (Expected Logs)

```
gas used: 2159
gas max : 100000000
PASS: TestMintUniqueNFT
PASS: TestMintEditionNFT
PASS: TestTransferEditionBetweenUsers
```


# 📖 **Exported Functions — Public ABI**

These functions are callable via:

* **Hive Custom JSON** with id: `vsc.call`
* Smart contract–to–contract calls within the VSC ecosystem

All payloads are **plain strings**. This improves gas efficiency and parsing speed over JSON.



## 🔧 **Mutations (State-Changing)**


### 🎯 Create Collection

**Action:** `col_create`

**Payload Format:**

```
<name>|<desc>|<meta>
```

| Field | Description | Required | Notes |
| -- | -  | -- | - |
| name  | Collection name          | ✅        | Max 48 chars       |
| desc  | Description              | ✅        | Max 128 chars      |
| meta  | Metadata (opaque string) | ✅        | Can be any string  |

**Example:**

```
My Art Collection|Best of my artworks|ipfs://Qm123abc
```



### 🎨 Mint NFT

**Action:** `nft_mint`

**Payload Format:**

```
<owner>_<collection>|<name>|<desc>|<singleTransfer>|<editions>|<meta>
```

| Field            | Required | Description                                     |
| - | -- | -- |
| owner_collection | ✅        | Format: `hive:account_collectionIndex`          |
| name             | ✅        | NFT name (max 48 chars)                                      |
| desc             | ✅        | NFT description (max 128 chars)                                |
| singleTransfer   | ✅        | `"true"` or `"false"` — soulbound-like behavior |
| editions         | ✅        | Empty = `1`, or set number e.g. `"10"`          |
| meta             | ✅        | Metadata can be any string                               |

**Unique NFT Example:**

```
hive:alice_0|Golden Sword|Legendary blade|false||{"rarity":"legendary"}
```

**Editioned NFT Example:**

```
hive:alice_0|Trading Card|Limited series|false|10|ipfs://QmMetaHash
```



### 🔄 **Transfer NFT or Edition**

**Action:** `nft_transfer`

**Payload Format:**

```
<nftID>|<editionIndex>|<owner>_<collection>
```

| Field            | Required | Description                    |
| - | -- | -  |
| nftID            | ✅        | The core NFT ID                |
| editionIndex     | ✅        | Empty for base (0), or integer |
| owner_collection | ✅        | Target owner & collection      |

**Examples:**

* Transfer base NFT:

```
42||hive:bob_1
```

* Transfer edition #3:

```
43|3|hive:bob_2
```



### 🔥 **Burn NFT / Edition**

**Action:** `nft_burn`

**Payload Format:**

```
<nftID>
<nftID>|<editionIndex>
```

**Example (burn edition 2):**

```
50|2
```



### 🏛 Add Market Contract (Admin Only)

**Action:** `add_market`

**Payload:**

```
hive:marketaddr
```



### 🏛 Remove Market Contract (Admin Only)

**Action:** `remove_market`

**Payload:**

```
hive:marketaddr
```



## 🔍 **Queries (Read-Only)**

These functions are intended to be used exclusively by other contracts. 

### 📦 **Get Collection**

**Action:** `col_get`

**Payload:**

```
<owner>_<collection>
```

**Returns:**

```
<owner>|<col>|<tx>|<name>|<desc>|<meta>
```



### 📊 **Get Collection Count for Account**

**Action:** `col_count`

**Payload:**

```
hive:alice
```

**Returns:** `"3"`



### ✅ **Check Collection Exists**

**Action:** `col_exists`

**Payload:**

```
hive:alice_1
```

**Returns:** `"true"` or `"false"`



### 🧬 **Get NFT**

**Action:** `nft_get`

**Payload:**

```
<id>
<id>|<edition>
```

**Returns:**

```
<nftID>|<editionIndex or empty>|<creator>|<owner_col>|<tx>|<name>|<desc>|<meta>|<edTotal>
```



### 🧾 **Check Ownership**

**Action:** `nft_isOwner`

Payload:

```
<nftID>
<nftID>|<editionIndex>
```

Returns: `"true"` or `"false"`



### 🧮 **Get Supply**

**Action:** `nft_supply`

Payload:

```
<nftID>
```

Returns:

```
10
```



### 🔥 **Check Burn State**

**Action:** `nft_isBurned`

Payload:

```
<nftID>|<editionIndex>
```

Returns `"true"` or `"false"`



### 🧷 **Check Single-Transfer (Soulbound)**

**Action:** `nft_isSingleTransfer`

Payload:

```
<nftID>
```

Returns `"true"` / `"false"`



### 📌 **Get NFT Owner (with collection context)**

**Action:** `nft_ownerColOf`

Payload:

```
<nftID>
<nftID>|<editionIndex>
```

Returns:

```
hive:alice_0
```



### 🧬 **Get NFT Creator**

**Action:** `nft_creator`

Payload:

```
42
```

Returns:

```
hive:alice
```



### 🧬 **Owned Editions**

**Action:** `nft_hasNFTEdition`

Payload:

```
<nftID>,<ownerAddress>
```

Returns:

```
0,1,3
```




# 🔔 **On-Chain Events**

The contract uses **manual, gas-optimized JSON serialization** to log events using `sdk.Log`. These events are designed to be **consumed by off-chain indexers** to store data in relational databases and expose it by api endpoints.

Each event follows this format:

```json
{
  "type": "<event_name>",
  "attributes": { ... },
  "tx": "<transaction_id>"
}
```

* **`type`** – identifies the category
* **`attributes`** – compact key-value data (custom per event)
* **`tx`** – blockchain transaction ID from `tx.id`

> ✅ Events use **short attribute keys** (“id”, “cr”, “oc”, “ed”) to reduce gas and storage.



## 📦 **Event Summary Table**

| Event Type   | Triggered By   | Attributes (Compact JSON)                                                 |
|  | -- | - |
| `collection` | `col_create`   | `{ "id":<collectionID>, "cr":"<creator>" }`                               |
| `mint`       | `nft_mint`     | `{ "id":<nftID>, "cr":"<creator>", "oc":"<owner_col>", "ed":<editions> }` |
| `transfer`   | `nft_transfer` | `{ "id":<nftID>, "ed":<edition?>, "fr":"<from>", "to":"<to>" }`           |
| `burn`       | `nft_burn`     | `{ "id":<nftID>, "ed":<edition?>, "ow":"<owner>" }`                       |

> ⚠ `ed` attribute is only emitted if NFT has multiple editions.



## 🧱 Event Field Glossary

| Key  | Meaning                                                  |
| - | -- |
| `id` | The numeric NFT or collection ID                         |
| `cr` | Creator address (minter or collection creator)           |
| `oc` | `"owner_collection"` formatted as `<owner>_<collection>` |
| `ed` | Edition index (optional for editioned NFTs)              |
| `fr` | From address (current owner)                             |
| `to` | Target owner                                             |
| `ow` | Owner performing the burn                                |
| `tx` | Immutable transaction ID                                 |



# 📘 Detailed Event Types


### 📦 **Collection Created Event**

Occurs when a user creates a new collection.

**Example Log:**

```json
{
  "type": "collection",
  "attributes": {
    "id": 0,
    "cr": "hive:alice"
  },
  "tx": "TX123ABC"
}
```



### 🪙 **Mint Event**

Triggered when a new NFT is minted into a collection.

**Example (unique NFT):**

```json
{
  "type": "mint",
  "attributes": {
    "id": 1001,
    "cr": "hive:alice",
    "oc": "hive:alice_0",
    "ed": 1
  },
  "tx": "TX456ABC"
}
```

**Example (editioned NFT with 10 copies):**

```json
{
  "type": "mint",
  "attributes": {
    "id": 1002,
    "cr": "hive:alice",
    "oc": "hive:alice_0",
    "ed": 10
  },
  "tx": "TX457ABC"
}
```



### 🔄 **Transfer Event**

Emitted for:

* Ownership transfers
* Collection changes for the same owner
* Edition-level transfers (only the edition being moved is affected)

**Example (single-edition transfer):**

```json
{
  "type": "transfer",
  "attributes": {
    "id": 1001,
    "fr": "hive:alice",
    "to": "hive:bob"
  },
  "tx": "TX789CDE"
}
```

**Example (edition 2 transfer only):**

```json
{
  "type": "transfer",
  "attributes": {
    "id": 1002,
    "ed": 2,
    "fr": "hive:alice",
    "to": "hive:bob"
  },
  "tx": "TX790CDE"
}
```



### 🔥 **Burn Event**

Marks the NFT or specific edition as permanently burned (cannot be reactivated).

**Example (burn full base NFT):**

```json
{
  "type": "burn",
  "attributes": {
    "id": 1001,
    "ow": "hive:alice"
  },
  "tx": "TX900DEF"
}
```

**Example (burn edition #1 only):**

```json
{
  "type": "burn",
  "attributes": {
    "id": 1002,
    "ed": 1,
    "ow": "hive:alice"
  },
  "tx": "TX901DEF"
}
```



## 📡 Event Consumption Guidelines for External Indexers

| Use Case                     | Contract to Listen For | Action                                   |
| - | - | - |
| Show real-time NFT ownership | `transfer` events      | Apply edition-level ownership overrides  |
| Display available supply     | Track `burn` events    | Mark editions as burned                  |
| List user collections        | `collection` events    | Index collection IDs via `(id, creator)` |
| Enumerate NFTs in collection | `mint` events          | Use `"oc"` + `"id"` attributes           |



### 🔌 Event Best Practices

* Indexers should replay events in order of **`timestamp` of the related ContractOutput**.
* Indexers should subscribe to the event stream for upcoming events.
* Always resolve final state using events + edition overrides.




# 🔎 How Data Is Stored (for Off-Chain Reads)

### ✅ Collections (ASCII keys; easy to query directly)

* **Key:** `c_<owner>_<collectionIndex>`
  Example: `c_hive:alice_0`
* **Value:** `"tx|name|desc|meta"`

> ➕ Reason: We deliberately store collection *core* under the **ASCII index key** so that off-chain tools can fetch a collection **with one key lookup**.

**Parsing:** Use a simple split on `|`:

* `tx` – transaction id at creation
* `name`, `desc`, `meta` – strings as provided on creation



### 🧱 NFTs (binary-prefixed keys; **do not** fetch raw keys directly)

NFTs use compact binary-prefixed keys internally (e.g., `\x01<idLE>`). These are **not** friendly to general GraphQL APIs.
👉 Off-chain consumers should use **exported getters** (e.g. `nft_get`, `nft_ownerColOf`) or better and recommended: **index events**.

**Recommended approach (off-chain):**

* Subscribe to events (`mint`, `transfer`, `burn`) and **maintain your own tables**:

  * `nfts(id, creator, oc, name, desc, meta, edTotal, txCreate)`
  * `nft_editions(nft_id, ed_index, owner_collection, burned)`
  * `collections(owner, index, name, desc, meta, txCreate)`
* Reconstruct current state by replaying events **in `timestamp` order**.


# 🔐 Marketplace & UI Patterns

* **Show current owner of edition:** `nft_ownerColOf("<id>|<ed>")`
* **Show supply:** `nft_supply("<id>")` (string integer)
* **Show metadata:** `nft_meta("<id>")` (opaque)
* **Check if NFT is soulbound:** `nft_isSingleTransfer("<id>")`
* **Check if user owns edition(s):** `nft_hasNFTEdition("<id>,<owner>")`
* **Verify ownership inline:** `nft_isOwner("<id>|<ed>")` returns `"true"`/`"false"`



# 🧪 Payload Helper Cheatsheet (Delimiters)

| Use Case | Payload | Example |
| - | - | - |
| Create collection` | `"<name>\|<desc>\|<meta>"`| `"MyNFTs\|Personal NFT vault\|ipfs://Qm123"` |
| Mint NFT | `"<owner>_<col>\|<name>\|<desc>\|<single>\|<editions>\| <meta>"`| `"hive:alice_0\|Dragon Egg\|Hatchable eggs\|false\|10\|ipfs://QmBBB"` |
| Transfer | `"<nftID>\|<edition>\|<owner>_<col>"` | `"43\|3\|hive:bob_1"` |
| Burn | `"<nftID>"` or `"<nftID>\|<edition>"` | `"43\|0"` |
| Get collection | `"<owner>_<col>"` | `"hive:alice_0"` |
| Check collection exists | `"<owner>_<col>"` | `"hive:alice_0"`|
| Count collections | `"<owner>"` | `"hive:alice"`|
| Get NFT | `"<id>"` or `"<id>\|<ed>"` | `"43\|0"` | 
| Is owner | `"<id>"` or `"<id>\|<ed>"` | `"43\|0"` |
| Get ownerCol | `"<id>"` or `"<id>\|<ed>"` | `"43\|0"` |
| Get creator | `"<id>"` | `"43"` |
| Get meta | `"<id>"` | `"43"` |
| Get supply | `"<id>"` | `"43"` |
| Is burned | `"<id>"` or `"<id>\|<ed>"`  | `"43\|0"` |
| Is single-transfer | `"<id>"`| `"43"` |



# ⚠️ Gotchas & Best Practices

* **Always** include edition index for **multi-edition** NFTs when checking ownership or burn state.
* Collections are **directly readable** via ASCII key `c_<owner>_<idx>`.
* For NFTs, prefer **getters** or **events**; do not rely on raw state binary keys.
* Your dApp should treat **metadata** as opaque (URI or inline JSON).
* Payloads are **strings**, not JSON—avoid spaces and use exact delimiters.


--- 
## 📚 Documentation
-   [Go-VSC-Node](https://github.com/vsc-eco/go-vsc-node)
-   [Go-Contract-Template](https://github.com/vsc-eco/go-contract-template)

## 📜 License
This project is licensed under the [MIT License](LICENSE).  