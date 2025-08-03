package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type CacheService struct {
	client *Client
}

func NewCacheService(client *Client) *CacheService {
	return &CacheService{
		client: client,
	}
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.SetWithExpiration(ctx, key, data, expiration)
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.GetString(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

func (c *CacheService) Delete(ctx context.Context, keys ...string) error {
	return c.client.Delete(ctx, keys...)
}

func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return c.client.Exists(ctx, key)
}

func (c *CacheService) SetUserSession(ctx context.Context, userID, sessionID string, expiration time.Duration) error {
	key := fmt.Sprintf("user_session:%s", userID)
	return c.client.SetWithExpiration(ctx, key, sessionID, expiration)
}

func (c *CacheService) GetUserSession(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("user_session:%s", userID)
	return c.client.GetString(ctx, key)
}

func (c *CacheService) DeleteUserSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user_session:%s", userID)
	return c.client.Delete(ctx, key)
}

func (c *CacheService) SetBlacklistedToken(ctx context.Context, tokenID string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklisted_token:%s", tokenID)
	return c.client.SetWithExpiration(ctx, key, "1", expiration)
}

func (c *CacheService) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklisted_token:%s", tokenID)
	return c.client.Exists(ctx, key)
}

func (c *CacheService) IncrementLoginAttempts(ctx context.Context, identifier string, expiration time.Duration) (int64, error) {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	return c.client.IncrementWithExpiration(ctx, key, expiration)
}

func (c *CacheService) GetLoginAttempts(ctx context.Context, identifier string) (int64, error) {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	result, err := c.client.GetString(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil
		}
		return 0, err
	}

	var attempts int64
	if err := json.Unmarshal([]byte(result), &attempts); err != nil {
		return 0, err
	}

	return attempts, nil
}

func (c *CacheService) ResetLoginAttempts(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	return c.client.Delete(ctx, key)
}
