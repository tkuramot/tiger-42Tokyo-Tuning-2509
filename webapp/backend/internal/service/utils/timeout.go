package utils

import (
	"context"
	"errors"
	"log"
	"time"

	"golang.org/x/sync/semaphore"
)

var (
	// 120 秒から短縮して遅延リクエストが滞留しないようにする
	defaultTimeout      = 10 * time.Second
	maxConcurrentTimers int64 = 64
	// 大量同時実行でゴルーチンが膨らまないように制限する
	timeoutLimiter = semaphore.NewWeighted(maxConcurrentTimers)
)

// 終わらない処理などによる無限ループを防ぐため、タイムアウト付きで処理を実行する
func WithTimeout(parent context.Context, fn func(ctx context.Context) error) error {
	timeout := defaultTimeout
	if dl, ok := parent.Deadline(); ok {
		if rem := time.Until(dl); rem > 0 && rem < timeout {
			timeout = rem
		}
	}

	if err := timeoutLimiter.Acquire(parent, 1); err != nil {
		return err
	}
	defer timeoutLimiter.Release(1)

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	if err := fn(ctx); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("処理がタイムアウトしました (timeout=%s)", timeout)
		}
		return err
	}

	return nil
}
