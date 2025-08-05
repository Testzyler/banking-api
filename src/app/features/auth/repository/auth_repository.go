package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Testzyler/banking-api/app/entities"
	"github.com/Testzyler/banking-api/app/models"
	"github.com/Testzyler/banking-api/database"
	"github.com/Testzyler/banking-api/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type authRepository struct {
	db          *gorm.DB
	redisClient redis.Cmdable
}

type AuthRepository interface {
	// Redis
	GetPinAttemptData(ctx context.Context, userID string) (*entities.PinAttemptData, error)
	IncrementFailedAttempts(ctx context.Context, userID string) (*entities.PinAttemptData, error)
	SetPinLock(ctx context.Context, userID string, lockedUntil time.Time, failedAttempts int, lastAttemptAt *time.Time) error
	ResetPinAttempts(ctx context.Context, userID string) error
	ListUserTokens(ctx context.Context, userID string) ([]entities.TokenResponse, error)
	StoreToken(ctx context.Context, userID string, tokenResponse *entities.TokenResponse) error
	BanAllUserTokens(ctx context.Context, userID, reason string) error
	IsTokenBanned(ctx context.Context, tokenID string) (bool, error)
	IsInBlacklist(ctx context.Context, userID string, tokenVersion int64) (bool, error)
	ValidateTokenVersion(ctx context.Context, tokenVersion int64) (*entities.TokenValidationResult, error)
	CleanupExpiredBans(ctx context.Context) error

	// database
	GetUserWithPin(username string) (*models.User, error)
	UpdateUserPinFailedAttempts(userID string, failedAttempts int) error
	UpdateUserPinLockedUntil(userID string, lockedUntil *time.Time) error
	UpdateUserPinLastAttemptAt(userID string, lastAttemptAt *time.Time) error
}

func NewAuthRepository(db *gorm.DB, redisDB *database.RedisDatabase) AuthRepository {
	var redisClient redis.Cmdable
	if redisDB != nil {
		redisClient = redisDB.GetClient()
	}

	return &authRepository{
		db:          db,
		redisClient: redisClient,
	}
}

// Redis helper methods
func (r *authRepository) pinAttemptKey(userID string) string {
	return fmt.Sprintf("pin_attempt:%s", userID)
}

func (r *authRepository) bannedTokenKey(tokenID string) string {
	return fmt.Sprintf("banned_token:%s", tokenID)
}

func (r *authRepository) bannedBlacklistKey(userID string) string {
	return fmt.Sprintf("banned_blacklist:%s", userID)
}

func (r *authRepository) userTokensKey(userID string) string {
	return fmt.Sprintf("user_tokens:%s", userID)
}

func (r *authRepository) invalidateUserCacheByID(userID string) {
	if r.redisClient == nil {
		return
	}
	userTokensKey := r.userTokensKey(userID)
	_ = r.redisClient.Del(context.Background(), userTokensKey).Err()
}

func (r *authRepository) GetUserWithPin(username string) (*models.User, error) {
	ctx := context.Background()

	if r.redisClient != nil {
		cacheKey := fmt.Sprintf("user_with_pin:%s", username)
		result, err := r.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var user models.User
			if err := json.Unmarshal([]byte(result), &user); err == nil {
				return &user, nil
			}
		}
	}

	// Fetch from database
	var user models.User
	err := r.db.
		Preload("UserPin").
		Where("name = ?", username).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	// Store in Redis cache for future requests (async)
	if r.redisClient != nil {
		go func() {
			cacheKey := fmt.Sprintf("user_with_pin:%s", username)
			userJSON, err := json.Marshal(user)
			if err == nil {
				_ = r.redisClient.Set(ctx, cacheKey, string(userJSON), 30*time.Minute).Err()
			}
		}()
	}

	return &user, nil
}

func (r *authRepository) GetPinAttemptData(ctx context.Context, userID string) (*entities.PinAttemptData, error) {
	if r.redisClient == nil {
		return &entities.PinAttemptData{UserID: userID, FailedAttempts: 0}, nil
	}

	key := r.pinAttemptKey(userID)
	result, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return &entities.PinAttemptData{UserID: userID, FailedAttempts: 0}, nil
		}
		return nil, fmt.Errorf("failed to get pin attempt data from Redis: %w", err)
	}

	var data entities.PinAttemptData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pin attempt data: %w", err)
	}

	return &data, nil
}

func (r *authRepository) setPinAttemptData(ctx context.Context, userID string, data *entities.PinAttemptData, ttl time.Duration) error {
	if r.redisClient == nil {
		return nil
	}

	key := r.pinAttemptKey(userID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal pin attempt data: %w", err)
	}

	return r.redisClient.Set(ctx, key, string(jsonData), ttl).Err()
}

func (r *authRepository) IncrementFailedAttempts(ctx context.Context, userID string) (*entities.PinAttemptData, error) {
	data, err := r.GetPinAttemptData(ctx, userID)
	if err != nil {
		logger.Warn("Failed to get pin attempt data from Redis, falling back to database", "error", err)

		// Fallback to database-only approach
		user, errInner := r.GetUserWithPin(userID)
		if errInner != nil {
			return nil, errInner
		}
		data = &entities.PinAttemptData{
			UserID:         userID,
			FailedAttempts: user.UserPin.FailedPinAttempts,
			PinLockedUntil: user.UserPin.PinLockedUntil,
			LastAttemptAt:  user.UserPin.LastPinAttemptAt,
		}
	}

	data.FailedAttempts++
	now := time.Now()
	data.LastAttemptAt = &now

	// Set Redis immediately
	ttl := 24 * time.Hour
	if err := r.setPinAttemptData(ctx, userID, data, ttl); err != nil {
		return nil, err
	}

	// Async database sync
	go func() {
		if err := r.UpdateUserPinFailedAttempts(userID, data.FailedAttempts); err != nil {
			logger.Errorf("Failed to sync failed attempts to database for user %s: %v", userID, err)
		} else if err := r.UpdateUserPinLastAttemptAt(userID, data.LastAttemptAt); err != nil {
			logger.Errorf("Failed to sync last attempt time to database for user %s: %v", userID, err)
		} else {
			logger.Infof("Successfully synced failed attempts to database for user %s", userID)
		}
	}()

	return data, nil
}

func (r *authRepository) SetPinLock(ctx context.Context, userID string, lockedUntil time.Time, failedAttempts int, lastAttemptAt *time.Time) error {
	data, err := r.GetPinAttemptData(ctx, userID)
	if err != nil {
		return err
	}

	data.PinLockedUntil = &lockedUntil
	data.FailedAttempts = failedAttempts
	data.LastAttemptAt = lastAttemptAt

	// update database async
	go func() {
		data, err := r.GetPinAttemptData(ctx, userID)
		if err != nil {
			logger.Warnf("Failed to get pin attempt data for user %s: %v", userID, err)
		}
		if data.FailedAttempts > 0 || data.PinLockedUntil != nil || data.LastAttemptAt != nil {
			if err := r.UpdateUserPinFailedAttempts(userID, 0); err != nil {
				logger.Errorf("Failed to reset attempts in database for user %s: %v", userID, err)
			} else if err := r.UpdateUserPinLockedUntil(userID, nil); err != nil {
				logger.Errorf("Failed to reset lock in database for user %s: %v", userID, err)
			} else if err := r.UpdateUserPinLastAttemptAt(userID, nil); err != nil {
				logger.Errorf("Failed to reset last attempt time in database for user %s: %v", userID, err)
			} else {
				logger.Infof("Successfully reset PIN attempts in database for user %s", userID)
			}
		}
	}()

	// Set Redis immediately
	ttl := time.Until(lockedUntil) + time.Hour
	if err := r.setPinAttemptData(ctx, userID, data, ttl); err != nil {
		return err
	}

	return nil
}

func (r *authRepository) ResetPinAttempts(ctx context.Context, userID string) error {
	data := &entities.PinAttemptData{
		UserID:         userID,
		FailedAttempts: 0,
		PinLockedUntil: nil,
		LastAttemptAt:  nil,
	}

	// Async database sync
	go func() {
		data, err := r.GetPinAttemptData(ctx, userID)
		if err != nil {
			logger.Warnf("Failed to get pin attempt data for user %s: %v", userID, err)
		}
		if data.FailedAttempts > 0 || data.PinLockedUntil != nil || data.LastAttemptAt != nil {
			if err := r.UpdateUserPinFailedAttempts(userID, 0); err != nil {
				logger.Errorf("Failed to reset attempts in database for user %s: %v", userID, err)
			} else if err := r.UpdateUserPinLockedUntil(userID, nil); err != nil {
				logger.Errorf("Failed to reset lock in database for user %s: %v", userID, err)
			} else if err := r.UpdateUserPinLastAttemptAt(userID, nil); err != nil {
				logger.Errorf("Failed to reset last attempt time in database for user %s: %v", userID, err)
			} else {
				logger.Infof("Successfully reset PIN attempts in database for user %s", userID)
			}
		}
	}()

	// Set Redis immediately
	ttl := time.Hour
	if err := r.setPinAttemptData(ctx, userID, data, ttl); err != nil {
		return err
	}

	return nil
}

func (r *authRepository) BanAllUserTokens(ctx context.Context, userID, reason string) error {
	if r.redisClient == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	banTimestamp := time.Now().Unix()
	key := r.bannedBlacklistKey(userID)

	blacklist := entities.BlacklistBan{
		UserID:       userID,
		BannedAt:     time.Now(),
		Reason:       reason,
		BanTimestamp: banTimestamp,
	}

	blacklistData, err := json.Marshal(blacklist)
	if err != nil {
		return fmt.Errorf("failed to marshal user ban data: %w", err)
	}

	// Store user ban for 24 hours
	if err := r.redisClient.Set(ctx, key, string(blacklistData), 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store user ban in Redis: %w", err)
	}

	// Also ban individual tokens for immediate effect (optional)
	userTokensKey := r.userTokensKey(userID)
	tokens, err := r.redisClient.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		logger.Warnf("Failed to get user tokens for individual banning: %v", err)
	} else {
		// Ban each token individually
		for _, tokenStr := range tokens {
			var tokenResponse entities.TokenResponse
			if err := json.Unmarshal([]byte(tokenStr), &tokenResponse); err != nil {
				logger.Errorf("Failed to unmarshal token for banning: %v", err)
				continue
			}

			if err := r.banToken(ctx, userID, tokenResponse.TokenID, reason); err != nil {
				logger.Errorf("Failed to ban token %s: %v", tokenResponse.TokenID, err)
			}
		}
	}

	// Invalidate user cache
	r.invalidateUserCacheByID(userID)

	logger.Infof("All tokens banned for user %s: %s", userID, reason)
	return nil
}

func (r *authRepository) banToken(ctx context.Context, userID, tokenID, reason string) error {
	if r.redisClient == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	bannedToken := entities.BannedToken{
		TokenID:      tokenID,
		UserID:       userID,
		BannedAt:     time.Now(),
		Reason:       reason,
		TokenVersion: time.Now().Unix(), // Use current timestamp as version
	}
	r.invalidateUserCacheByID(userID)
	bannedKey := r.bannedTokenKey(tokenID)
	bannedData, err := json.Marshal(bannedToken)
	if err != nil {
		return fmt.Errorf("failed to marshal banned token data: %w", err)
	}

	if err := r.redisClient.Set(ctx, bannedKey, string(bannedData), 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store banned token in Redis: %w", err)
	}

	return nil
}

func (r *authRepository) IsTokenBanned(ctx context.Context, tokenID string) (bool, error) {
	if r.redisClient == nil {
		return true, nil // If Redis is not available, assume token is banned user input pin
	}

	bannedKey := r.bannedTokenKey(tokenID)
	result, err := r.redisClient.Get(ctx, bannedKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("failed to check banned token: %w", err)
	}

	// Token exists in banned list
	return result != "", nil
}

func (r *authRepository) IsInBlacklist(ctx context.Context, userID string, tokenVersion int64) (bool, error) {
	if r.redisClient == nil {
		return false, nil // If Redis is not available, don't assume user is banned
	}

	blacklistKey := r.bannedBlacklistKey(userID)
	result, err := r.redisClient.Get(ctx, blacklistKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("failed to check token ban status: %w", err)
	}

	// Parse user ban data
	var blacklist entities.BlacklistBan
	if err := json.Unmarshal([]byte(result), &blacklist); err != nil {
		logger.Errorf("Failed to unmarshal user ban data: %v", err)
		return false, nil
	}

	return tokenVersion < blacklist.BanTimestamp, nil
}

func (r *authRepository) ValidateTokenVersion(ctx context.Context, tokenVersion int64) (*entities.TokenValidationResult, error) {
	currentTime := time.Now().Unix()
	maxAge := int64(24 * 60 * 60) // 24 hours in seconds

	if tokenVersion < (currentTime - maxAge) {
		return &entities.TokenValidationResult{
			Valid:        false,
			Reason:       "Token is too old",
			TokenVersion: currentTime,
		}, nil
	}

	return &entities.TokenValidationResult{
		Valid:        true,
		TokenVersion: currentTime,
	}, nil
}

func (r *authRepository) CleanupExpiredBans(ctx context.Context) error {
	if r.redisClient == nil {
		return nil
	}

	var cursor uint64
	for {
		keys, nextCursor, err := r.redisClient.Scan(ctx, cursor, "banned_token:*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan banned tokens: %w", err)
		}

		for _, key := range keys {
			// Check if key exists (Redis TTL will auto-expire keys)
			exists, err := r.redisClient.Exists(ctx, key).Result()
			if err != nil {
				logger.Errorf("Failed to check key existence: %v", err)
				continue
			}
			if exists == 0 {
				logger.Debugf("Banned token key %s has expired", key)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (r *authRepository) StoreToken(ctx context.Context, userID string, tokenResponse *entities.TokenResponse) error {
	if r.redisClient == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	key := r.userTokensKey(userID)
	tokenJSON, err := json.Marshal(tokenResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal token response: %w", err)
	}

	// Store token in the user's token set
	if err := r.redisClient.SAdd(ctx, key, string(tokenJSON)).Err(); err != nil {
		return fmt.Errorf("failed to store token in Redis: %w", err)
	}

	// Set expiry for the key based on token expiry
	if !tokenResponse.Expiry.IsZero() {
		ttl := time.Until(tokenResponse.Expiry)
		if ttl > 0 {
			if err := r.redisClient.Expire(ctx, key, ttl).Err(); err != nil {
				logger.Warnf("Failed to set expiry for token key %s: %v", key, err)
			}
		}
	}

	return nil
}

func (r *authRepository) ListUserTokens(ctx context.Context, userID string) ([]entities.TokenResponse, error) {
	if r.redisClient == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	var allTokenResponses []entities.TokenResponse
	var cursor uint64

	for {
		key := r.userTokensKey(userID)
		keys, nextCursor, err := r.redisClient.Scan(ctx, cursor, key, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan token keys from Redis: %w", err)
		}

		// Process found keys
		for _, key := range keys {
			tokens, err := r.redisClient.SMembers(ctx, key).Result()
			if err != nil {
				logger.Errorf("Failed to get tokens for user %s: %v", userID, err)
				continue
			}

			for _, tokenStr := range tokens {
				var tokenResponse entities.TokenResponse
				if err := json.Unmarshal([]byte(tokenStr), &tokenResponse); err != nil {
					logger.Errorf("Failed to unmarshal token for user %s: %v", userID, err)
					continue
				}

				isTokenBanned, err := r.IsTokenBanned(ctx, tokenResponse.TokenID)
				if err != nil {
					logger.Errorf("Failed to check if token %s is banned: %v", tokenResponse.TokenID, err)
					isTokenBanned = false
				}

				isUserBanned, err := r.IsInBlacklist(ctx, userID, tokenResponse.TokenVersion)
				if err != nil {
					logger.Errorf("Failed to check user ban status for user %s: %v", userID, err)
					isUserBanned = false
				}
				isBanned := isUserBanned || isTokenBanned
				tokenResponse.IsBanned = &isBanned
				allTokenResponses = append(allTokenResponses, tokenResponse)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return allTokenResponses, nil
}

// Legacy database-only methods for fallback
func (r *authRepository) UpdateUserPinFailedAttempts(userID string, failedAttempts int) error {
	return r.db.Model(&models.UserPin{}).Where("user_id = ?", userID).Update("failed_pin_attempts", failedAttempts).Error
}

func (r *authRepository) UpdateUserPinLockedUntil(userID string, lockedUntil *time.Time) error {
	return r.db.Model(&models.UserPin{}).Where("user_id = ?", userID).Update("pin_locked_until", lockedUntil).Error
}

func (r *authRepository) UpdateUserPinLastAttemptAt(userID string, lastAttemptAt *time.Time) error {
	return r.db.Model(&models.UserPin{}).Where("user_id = ?", userID).Update("last_pin_attempt_at", lastAttemptAt).Error
}
