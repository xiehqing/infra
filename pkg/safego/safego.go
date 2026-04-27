package safego

import "context"

// Go 安全go执行
func Go(ctx context.Context, f func()) {
	go func() {
		defer Recovery(ctx)
		f()
	}()
}
