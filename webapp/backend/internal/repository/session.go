package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionRepository struct {
	db    DBTX
	cache map[string]*sessionCacheEntry
	mu    sync.RWMutex
}

type sessionCacheEntry struct {
	userID    int
	expiresAt time.Time
}

func NewSessionRepository(db DBTX) *SessionRepository {
	repo := &SessionRepository{
		db:    db,
		cache: make(map[string]*sessionCacheEntry),
	}
	go repo.cleanupCache()
	return repo
}

func (r *SessionRepository) cleanupCache() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		now := time.Now()
		for sessionID, entry := range r.cache {
			if now.After(entry.expiresAt) {
				delete(r.cache, sessionID)
			}
		}
		r.mu.Unlock()
	}
}

// セッションを作成し、セッションIDと有効期限を返す
func (r *SessionRepository) Create(ctx context.Context, userBusinessID int, duration time.Duration) (string, time.Time, error) {
	sessionUUID, err := uuid.NewRandom()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(duration)
	sessionIDStr := sessionUUID.String()

	query := "INSERT INTO user_sessions (session_uuid, user_id, expires_at) VALUES (?, ?, ?)"
	_, err = r.db.ExecContext(ctx, query, sessionIDStr, userBusinessID, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}
	return sessionIDStr, expiresAt, nil
}

// セッションIDからユーザーIDを取得
func (r *SessionRepository) FindUserBySessionID(ctx context.Context, sessionID string) (int, error) {
	r.mu.RLock()
	entry, exists := r.cache[sessionID]
	r.mu.RUnlock()

	if exists && time.Now().Before(entry.expiresAt) {
		return entry.userID, nil
	}

	type result struct {
		UserID    int       `db:"user_id"`
		ExpiresAt time.Time `db:"expires_at"`
	}
	var res result
	query := `
		SELECT
			u.user_id, s.expires_at
		FROM users u
		JOIN user_sessions s ON u.user_id = s.user_id
		WHERE s.session_uuid = ? AND s.expires_at > ?`
	err := r.db.GetContext(ctx, &res, query, sessionID, time.Now())
	if err != nil {
		return 0, err
	}

	r.mu.Lock()
	r.cache[sessionID] = &sessionCacheEntry{
		userID:    res.UserID,
		expiresAt: res.ExpiresAt,
	}
	r.mu.Unlock()

	return res.UserID, nil
}
