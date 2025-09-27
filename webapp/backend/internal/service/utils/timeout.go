package utils

import "context"

// WithTimeout は呼び出し先の実装に委ねるため、親コンテキストをそのまま渡して実行する
func WithTimeout(parent context.Context, fn func(ctx context.Context) error) error {
	return fn(parent)
}
