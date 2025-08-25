package services

import (
	"sync"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
)

type RateLimiter struct {
	attempts map[string]*models.LoginAttempt
	mutex    sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string]*models.LoginAttempt),
	}
}

// CheckRateLimit はレート制限をチェックします
func (r *RateLimiter) CheckRateLimit(email string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	attempt, exists := r.attempts[email]
	if !exists {
		// 初回アクセス
		r.attempts[email] = &models.LoginAttempt{
			Email:    email,
			Attempts: 0,
			LastTry:  time.Now(),
		}
		return nil
	}

	// ブロック中かチェック
	if attempt.BlockedAt != nil {
		// 10分間のブロック
		if time.Since(*attempt.BlockedAt) < 10*time.Minute {
			return apperrors.Forbidden.Wrap(nil, "アクセスが一時的にブロックされています。10分後に再試行してください。")
		}
		// ブロック解除
		attempt.BlockedAt = nil
		attempt.Attempts = 0
	}

	// 1分以内に5回以上の試行でブロック
	if time.Since(attempt.LastTry) < time.Minute && attempt.Attempts >= 5 {
		now := time.Now()
		attempt.BlockedAt = &now
		return apperrors.Forbidden.Wrap(nil, "短時間に多数のアクセスが検出されました。10分後に再試行してください。")
	}

	// 1分経過していれば試行回数をリセット
	if time.Since(attempt.LastTry) >= time.Minute {
		attempt.Attempts = 0
	}

	return nil
}

// RecordAttempt は認証試行を記録します
func (r *RateLimiter) RecordAttempt(email string, success bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	attempt, exists := r.attempts[email]
	if !exists {
		attempt = &models.LoginAttempt{
			Email:    email,
			Attempts: 0,
		}
		r.attempts[email] = attempt
	}

	attempt.LastTry = time.Now()

	if !success {
		attempt.Attempts++
	} else {
		// 成功時は試行回数をリセット
		attempt.Attempts = 0
		attempt.BlockedAt = nil
	}
}

// CleanupOldAttempts は古い試行記録をクリーンアップします
func (r *RateLimiter) CleanupOldAttempts() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // 24時間前
	for email, attempt := range r.attempts {
		if attempt.LastTry.Before(cutoff) {
			delete(r.attempts, email)
		}
	}
}
