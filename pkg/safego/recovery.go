package safego

import (
	"context"
	"github.com/xiehqing/infra/pkg/logs"
	"runtime/debug"
)

// Recovery recover panic
func Recovery(ctx context.Context) {
	e := recover()
	if e == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	logs.Errorf("safego.Recover: panic error = %v \n stack = \n%s", e, string(debug.Stack()))
}
