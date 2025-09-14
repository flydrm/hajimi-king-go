package github

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// TokenManager manages multiple GitHub tokens with rotation and failover
type TokenManager struct {
	tokens     []string
	current    int
	mu         sync.RWMutex
	blacklist  map[string]time.Time
	blacklistMu sync.RWMutex
	blacklistTTL time.Duration
}

// NewTokenManager creates a new token manager
func NewTokenManager(tokens []string) *TokenManager {
	if len(tokens) == 0 {
		return nil
	}

	// Shuffle tokens for load balancing
	shuffledTokens := make([]string, len(tokens))
	copy(shuffledTokens, tokens)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffledTokens), func(i, j int) {
		shuffledTokens[i], shuffledTokens[j] = shuffledTokens[j], shuffledTokens[i]
	})

	return &TokenManager{
		tokens:        shuffledTokens,
		current:       0,
		blacklist:     make(map[string]time.Time),
		blacklistTTL:  5 * time.Minute, // Blacklist for 5 minutes
	}
}

// GetNextToken returns the next available token
func (tm *TokenManager) GetNextToken() (string, error) {
	if tm == nil {
		return "", fmt.Errorf("token manager not initialized")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Clean expired blacklisted tokens
	tm.cleanBlacklist()

	// Try to find an available token
	attempts := 0
	for attempts < len(tm.tokens) {
		token := tm.tokens[tm.current]
		
		// Check if token is blacklisted
		if !tm.isBlacklisted(token) {
			// Move to next token for next call
			tm.current = (tm.current + 1) % len(tm.tokens)
			return token, nil
		}
		
		// Move to next token
		tm.current = (tm.current + 1) % len(tm.tokens)
		attempts++
	}

	return "", fmt.Errorf("no available tokens")
}

// BlacklistToken adds a token to the blacklist
func (tm *TokenManager) BlacklistToken(token string, reason string) {
	if tm == nil {
		return
	}

	tm.blacklistMu.Lock()
	defer tm.blacklistMu.Unlock()

	tm.blacklist[token] = time.Now()
	fmt.Printf("🚫 Token blacklisted: %s... (reason: %s)\n", token[:8], reason)
}

// isBlacklisted checks if a token is blacklisted
func (tm *TokenManager) isBlacklisted(token string) bool {
	tm.blacklistMu.RLock()
	defer tm.blacklistMu.RUnlock()

	blacklistTime, exists := tm.blacklist[token]
	if !exists {
		return false
	}

	// Check if blacklist has expired
	return time.Since(blacklistTime) < tm.blacklistTTL
}

// cleanBlacklist removes expired blacklisted tokens
func (tm *TokenManager) cleanBlacklist() {
	tm.blacklistMu.Lock()
	defer tm.blacklistMu.Unlock()

	now := time.Now()
	for token, blacklistTime := range tm.blacklist {
		if now.Sub(blacklistTime) >= tm.blacklistTTL {
			delete(tm.blacklist, token)
		}
	}
}

// GetTokenCount returns the number of available tokens
func (tm *TokenManager) GetTokenCount() int {
	if tm == nil {
		return 0
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	availableCount := 0
	for _, token := range tm.tokens {
		if !tm.isBlacklisted(token) {
			availableCount++
		}
	}

	return availableCount
}

// GetBlacklistedCount returns the number of blacklisted tokens
func (tm *TokenManager) GetBlacklistedCount() int {
	if tm == nil {
		return 0
	}

	tm.blacklistMu.RLock()
	defer tm.blacklistMu.RUnlock()

	return len(tm.blacklist)
}

// GetStatus returns the current status of the token manager
func (tm *TokenManager) GetStatus() map[string]interface{} {
	if tm == nil {
		return map[string]interface{}{
			"total_tokens":     0,
			"available_tokens": 0,
			"blacklisted":      0,
			"current_index":    0,
		}
	}

	tm.mu.RLock()
	tm.blacklistMu.RLock()
	defer tm.mu.RUnlock()
	defer tm.blacklistMu.RUnlock()

	availableCount := 0
	for _, token := range tm.tokens {
		if !tm.isBlacklisted(token) {
			availableCount++
		}
	}

	return map[string]interface{}{
		"total_tokens":     len(tm.tokens),
		"available_tokens": availableCount,
		"blacklisted":      len(tm.blacklist),
		"current_index":    tm.current,
	}
}