package store

import "sync"

// in-memory store for bot tokens

// could be extended with ExpiresAt field or ideally it should be stored in Reids with autoexpire after 24h
type BotToken struct {
	Token string
}

type BotStore struct {
	mu     sync.RWMutex
	tokens map[string]*BotToken
}

func NewBotStore() *BotStore {
	return &BotStore{
		tokens: make(map[string]*BotToken),
	}
}

func (bs *BotStore) StoreBotToken(organizationID, token string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.tokens[organizationID] = &BotToken{
		Token: token,
	}
}

func (bs *BotStore) GetBotToken(organizationID string) (string, bool) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	botToken, exists := bs.tokens[organizationID]
	if !exists {
		return "", false
	}
	return botToken.Token, true
}
