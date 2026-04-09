package main

import (
	"encoding/json"
	"time"
)

const MaxEntries = 3

type TokenEntry struct {
	Token   string    `json:"tok"`
	AddedAt time.Time `json:"added_at"`
}

type TokenStore struct {
	Entries []TokenEntry `json:"tokens"`
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		Entries: make([]TokenEntry, 0, MaxEntries),
	}
}

func (cm *TokenStore) AddToken(token string) {
	cm.AddTokenEntry(TokenEntry{
		Token:   token,
		AddedAt: time.Now().UTC().Truncate(time.Second),
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
