package main

import (
	"fmt"
	"testing"
	"time"

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
		AddedAt: time.Date(2026, 4, 9, 13, 15, 0, 0, time.UTC),
	})

	data, err := ts.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
	got := string(data)
	want := `{"entries":[{"token":"token-1","added_at":"2026-04-09T13:15:00Z"}]}`
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

func TestFromStringJSONSingleToken(t *testing.T) {
	ts := NewTokenStoreFromString(`{"entries":[{"token":"token-1","added_at":"2026-04-09T13:15:00Z"}]}`)
	want := 1
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFromStringJSONMultipleTokens(t *testing.T) {
	ts := NewTokenStoreFromString(`{"entries":[{"token":"token-1","added_at":"2026-04-08T13:15:00Z"},{"token":"token-2","added_at":"2026-04-09T13:16:00Z"}]}`)
	want := 2
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}

func TestFromStringThatLooksLikeJSON(t *testing.T) {
	ts := NewTokenStoreFromString("{xyzzy}")
	want := 1
	got := len(ts.Entries)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
	token := ts.Entries[0].Token
	wantToken := "{xyzzy}"
	if diff := cmp.Diff(wantToken, token); diff != "" {
		t.Errorf("mismatch (-want +got): %s", diff)
	}
}
