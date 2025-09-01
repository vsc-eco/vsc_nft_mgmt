//go:build test
// +build test

package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// positive tests
func TestCreateAndGetCollection_Positive(t *testing.T) {
	// Arrange
	sdk := NewFakeSDK("user1", "tx123")
	payload := `{"name": "MyCollection", "desc": "A test collection"}`

	// Act
	createCollectionImpl(&payload, sdk)

	// Assert: Check collection count updated
	if getCount(CollectionCount, sdk) != 1 {
		t.Fatalf("expected collection count = 1, got %d", getCount(CollectionCount, sdk))
	}

	// Assert: Can load collection by ID
	id := "0"
	col := loadCollection(id, sdk)
	if col.Name != "MyCollection" {
		t.Errorf("expected name MyCollection, got %s", col.Name)
	}
	if col.Owner.String() != "user1" {
		t.Errorf("expected owner user1, got %s", col.Owner.String())
	}
}

func TestGetCollectionsForOwner_Positive(t *testing.T) {
	// Arrange
	sdk := NewFakeSDK("userA", "tx999")

	payload1 := `{"name": "ColA", "desc": "Desc A"}`
	payload2 := `{"name": "ColB", "desc": "Desc B"}`

	createCollectionImpl(&payload1, sdk)
	createCollectionImpl(&payload2, sdk)

	owner := "userA"
	// Act
	jsonStr := getCollectionsForOwnerImpl(&owner, sdk)

	// Assert
	var collections []Collection
	err := json.Unmarshal([]byte(*jsonStr), &collections)
	if err != nil {
		t.Fatalf("failed to unmarshal collections: %v", err)
	}

	if len(collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(collections))
	}
	if collections[0].Owner.String() != "userA" {
		t.Errorf("expected owner userA, got %s", collections[0].Owner.String())
	}
}

func TestCreateCollection_MaxLengths_Positive(t *testing.T) {
	sdk := NewFakeSDK("userY", "txBound")
	name := strings.Repeat("n", maxNameLength)
	desc := strings.Repeat("d", maxDescLength)
	payload := `{"name": "` + name + `", "desc": "` + desc + `"}`

	createCollectionImpl(&payload, sdk)
	id := "0"
	col := loadCollection(id, sdk)

	if col.Name != name || col.Description != desc {
		t.Errorf("unexpected collection values at max length")
	}
}

func TestCreateCollection_EmptyDesc_Positive(t *testing.T) {
	sdk := NewFakeSDK("userZ", "txEmptyDesc")
	payload := `{"name": "ColWithNoDesc", "desc": ""}`

	createCollectionImpl(&payload, sdk)

	id := "0"
	col := loadCollection(id, sdk)
	if col.Description != "" {
		t.Errorf("expected empty description, got %s", col.Description)
	}
}

func TestGetCollectionsForDifferentOwners_Positive(t *testing.T) {
	sdk := NewFakeSDK("user1", "txMulti")

	payload1 := `{"name": "A", "desc": "first"}`
	payload2 := `{"name": "B", "desc": "second"}`
	createCollectionImpl(&payload1, sdk)

	// switch owner
	sdk.env.Sender.Address = "user2"
	sdk.env.Caller = "user2"
	createCollectionImpl(&payload2, sdk)

	owner1 := "user1"
	res1 := getCollectionsForOwnerImpl(&owner1, sdk)
	var col1 []Collection
	json.Unmarshal([]byte(*res1), &col1)
	if len(col1) != 1 || col1[0].Owner.String() != "user1" {
		t.Errorf("expected 1 collection for user1, got %+v", col1)
	}

	owner2 := "user2"
	res2 := getCollectionsForOwnerImpl(&owner2, sdk)
	var col2 []Collection
	json.Unmarshal([]byte(*res2), &col2)
	if len(col2) != 1 || col2[0].Owner.String() != "user2" {
		t.Errorf("expected 1 collection for user2, got %+v", col2)
	}
}

// negative tests
func TestCreateCollectionFailsOnEmptyName_Negative(t *testing.T) {
	sdk := NewFakeSDK("userX", "txAbort")
	payload := `{"name": "", "desc": "oops"}`

	defer expectAbort(t, sdk, "name is mandatory")

	createCollectionImpl(&payload, sdk)
}

func TestCreateCollectionFailsOnTooLongName_Negative(t *testing.T) {
	sdk := NewFakeSDK("userX", "txAbort")

	// name longer than maxNameLength
	longName := strings.Repeat("a", maxNameLength+1)
	payload := `{"name": "` + longName + `", "desc": "fine"}`

	expected := "name: max " + strconv.Itoa(maxNameLength) + " chars"
	defer expectAbort(t, sdk, expected)

	createCollectionImpl(&payload, sdk)
}

func TestCreateCollectionFailsOnTooLongDescription_Negative(t *testing.T) {
	sdk := NewFakeSDK("userX", "txAbort")

	// description longer than maxDescLength
	longDesc := strings.Repeat("d", maxDescLength+1)
	payload := `{"name": "okay", "desc": "` + longDesc + `"}`

	expected := "desc: max " + strconv.Itoa(maxDescLength) + " chars"
	defer expectAbort(t, sdk, expected)

	createCollectionImpl(&payload, sdk)
}
