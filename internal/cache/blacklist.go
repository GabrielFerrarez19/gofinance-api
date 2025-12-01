package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBlacklist struct {
	client *redis.Client
}

// AddToken adiciona um token à blacklist até expirar
func (tb *TokenBlacklist) AddToken(ctx context.Context, tokenID string, expiresIn time.Duration) error {
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	return tb.client.Set(ctx, key, "1", expiresIn).Err()
}

// IsTokenBlacklisted verifica se um token está na blacklist
func (tb *TokenBlacklist) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:token:%s", tokenID)

	exists, err := tb.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	return exists > 0, nil
}

// RemoveToken remove um token da blacklist (logout antes de expirar)
func (tb *TokenBlacklist) RemoveToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	return tb.client.Del(ctx, key).Err()
}

// AddRefreshToken adiciona um refresh token à blacklist
func (tb *TokenBlacklist) AddRefreshToken(ctx context.Context, refreshTokenID string, expiresIn time.Duration) error {
	key := fmt.Sprintf("blacklist:refresh:%s", refreshTokenID)
	return tb.client.Set(ctx, key, "1", expiresIn).Err()
}

// IsRefreshBlacklisted verifica se um refresh token está na blacklist
func (tb *TokenBlacklist) IsRefreshBlacklisted(ctx context.Context, refreshTokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:refresh:%s", refreshTokenID)

	exists, err := tb.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check refresh token blacklist: %w", err)
	}
	return exists > 0, nil
}

// ClearUserTokens remove todos os tokens do um usuário (logout de todos os dispositivos)
func (tb *TokenBlacklist) ClearUserTokens(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("blacklist:*:*:%s", userID)
	keys, err := tb.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	if len(keys) > 0 {
		return tb.client.Del(ctx, keys...).Err()
	}
	return nil
}
