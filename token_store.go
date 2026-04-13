package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const MaxEntries = 3

// prefix for all v2 encoded tokens, base64url-encoded string '{"entries":['
const V2EncodedTokenPrefix = "eyJlbnRyaWVzIjpb"

type TokenEntry struct {
	Token   string `json:"token"`
	AddedAt int64  `json:"added_at"`
}

type TokenStore struct {
	Entries []TokenEntry `json:"entries"`
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		Entries: make([]TokenEntry, 0, MaxEntries),
	}
}

func NewTokenStoreFromString(str string) *TokenStore {
	ts := NewTokenStore()
	ts.FromString(str)
	return ts
}

func (ts *TokenStore) AddToken(token string) {
	ts.AddTokenEntry(TokenEntry{
		Token:   token,
		AddedAt: time.Now().Unix(),
	})
}

func (ts *TokenStore) AddTokenEntry(entry TokenEntry) {
	if len(ts.Entries) >= MaxEntries {
		ts.Entries = ts.Entries[1:]
	}
	ts.Entries = append(ts.Entries, entry)
}

func (ts *TokenStore) FindToken(token string) *TokenEntry {
	for i := len(ts.Entries) - 1; i >= 0; i-- {
		if ts.Entries[i].Token == token {
			return &ts.Entries[i]
		}
	}
	return nil
}

func (ts *TokenStore) ToJSON() ([]byte, error) {
	return json.Marshal(ts)
}

func (ts *TokenStore) FromJSON(data []byte) error {
	return json.Unmarshal(data, ts)
}

// FromString creates a TokenStore from a string, either a plain PN token (v1) or a base64url-encoded JSON TS object (v2).
func (ts *TokenStore) FromString(str string) {
	// if it looks like a v2 encoded object, try to parse it
	if strings.HasPrefix(str, V2EncodedTokenPrefix) {
		decoded, err := base64.URLEncoding.DecodeString(str)
		if err == nil {
			err = ts.FromJSON(decoded)
			if err == nil {
				return
			}
		}
	}

	// treat it as a single token string
	ts.AddToken(str)
}
