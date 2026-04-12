package main

import (
	"encoding/json"
	"strings"
	"time"
)

const MaxEntries = 3

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

func (cm *TokenStore) AddToken(token string) {
	cm.AddTokenEntry(TokenEntry{
		Token:   token,
		AddedAt: time.Now().Unix(),
	})
}

func (cm *TokenStore) AddTokenEntry(entry TokenEntry) {
	if len(cm.Entries) >= MaxEntries {
		cm.Entries = cm.Entries[1:]
	}
	cm.Entries = append(cm.Entries, entry)
}

func (cm *TokenStore) FindToken(token string) *TokenEntry {
	for i := len(cm.Entries) - 1; i >= 0; i-- {
		if cm.Entries[i].Token == token {
			return &cm.Entries[i]
		}
	}
	return nil
}

func (cm *TokenStore) ToJSON() ([]byte, error) {
	return json.Marshal(cm)
}

func (cm *TokenStore) FromJSON(data []byte) error {
	return json.Unmarshal(data, cm)
}

// FromString creates a TokenStore from a string, which might be a single string token or a JSON object.
func (cm *TokenStore) FromString(str string) {
	// if it looks like a JSON object, try to parse it
	trimmed := strings.TrimSpace(str)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		err := cm.FromJSON([]byte(trimmed))
		if err == nil {
			return
		}
	}

	// treat it as a single token string
	cm.AddToken(str)
}
