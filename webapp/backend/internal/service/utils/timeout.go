package utils

import (
	"context"
	"log"
	"time"
)

var defaultTimeout = 120 * time.Second

// 終わらない処理などによる無限ループを防ぐため、タイムアウト付きで処理を実行する
func WithTimeout(parent context.Context, fn func(ctx context.Context) error) error {
	timeout := defaultTimeout
	if dl, ok := parent.Deadline(); ok {
		if rem := time.Until(dl); rem > 0 && rem < timeout {
			timeout = rem
		}
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- fn(ctx) }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		log.Printf("処理がタイムアウトしました (timeout=%s)", timeout)
		return ctx.Err()
	}
}
