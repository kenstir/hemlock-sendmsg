package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAddToken(t *testing.T) {
	ts := NewTokenStore()
	token := "test-token-1"
	ts.AddToken(token)

	want := 1
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestAddTooManyTokens(t *testing.T) {
	ts := NewTokenStore()
	for i := 0; i <= MaxEntries+1; i++ {
		ts.AddToken(fmt.Sprintf("token-%d", i))
	}

	want := MaxEntries
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}

	firstToken := ts.Entries[0].Token
	wantFirst := "token-2"
	if diff := cmp.Diff(wantFirst, firstToken); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}

	lastToken := ts.Entries[len(ts.Entries)-1].Token
	wantLast := fmt.Sprintf("token-%d", MaxEntries+1)
	if diff := cmp.Diff(wantLast, lastToken); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFindToken(t *testing.T) {
	ts := NewTokenStore()
	ts.AddToken("token-1")
	ts.AddToken("token-2")

	found := ts.FindToken("token-2")
	if diff := cmp.Diff("token-2", found.Token); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFindTokenNotFound(t *testing.T) {
	ts := NewTokenStore()
	ts.AddToken("token-1")

	found := ts.FindToken("nonexistent")
	if found != nil {
		t.Errorf("expected nil, got %v", found)
	}
}

func TestToJSON(t *testing.T) {
	ts := NewTokenStore()
	ts.AddTokenEntry(TokenEntry{
		Token:   "token-1",
		AddedAt: 1712664900,
	})

	data, err := ts.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	want := `{"entries":[{"token":"token-1","added_at":1712664900}]}`
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFromJSON(t *testing.T) {
	original := NewTokenStore()
	original.AddToken("token-1")
	original.AddToken("token-2")

	data, _ := original.ToJSON()

	ts := NewTokenStore()
	err := ts.FromJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(ts.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(ts.Entries))
	}
}

func TestFromJSONInvalid(t *testing.T) {
	ts := NewTokenStore()
	err := ts.FromJSON([]byte("invalid json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFromStringSingleToken(t *testing.T) {
	ts := NewTokenStoreFromString("token-1")

	want := 1
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestEncodingIsCompatible(t *testing.T) {
	// Check that the implementation we are using is compatible with other implementations,
	// that is, base64-url-encoding with no padding.
	json := `{"a":"??~"}`
	want := "eyJhIjoiPz9-In0" // plain base64 would be "eyJhIjoiPz9+In0="

	encoded := V2EncodeString([]byte(json))
	if diff := cmp.Diff(want, encoded); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}

	decoded, err := V2DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(json, string(decoded)); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFromStringV2SingleToken(t *testing.T) {
	json := `{"entries":[{"token":"token-1","added_at":1712664900}]}`
	encoded := V2EncodeString([]byte(json))
	ts := NewTokenStoreFromString(encoded)

	want := 1
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFromStringV2MultipleTokens(t *testing.T) {
	json := `{"entries":[{"token":"token-1","added_at":1712578500},{"token":"token-2","added_at":1712664960}]}`
	encoded := V2EncodeString([]byte(json))
	ts := NewTokenStoreFromString(encoded)

	want := 2
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFromStringThatLooksLikeV2(t *testing.T) {
	str := V2EncodedTokenPrefix + "xyzzy"
	ts := NewTokenStoreFromString(str)

	want := 1
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
	token := ts.Entries[0].Token
	if diff := cmp.Diff(str, token); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}
