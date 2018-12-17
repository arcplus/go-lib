package mysql

import (
	"context"
	"math"
	"time"

	"github.com/arcplus/go-lib/log"
)

// Hooks satisfies the sqlhook.Hooks interface
type Hook struct{}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *Hook) Before(ctx context.Context, _ string, _ ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, "x-sql-begin", time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *Hook) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	startAt, ok := ctx.Value("x-sql-begin").(time.Time)
	if !ok {
		return ctx, nil
	}

	logger := log.KVPair(map[string]interface{}{
		"span": driverName,
		"took": nanoToMs(time.Since(startAt).Nanoseconds()),
	})

	if tid, ok := ctx.Value("x-request-id").(string); ok {
		logger = logger.Trace(tid)
	}

	if logger.DebugEnabled() {
		logger.Debugf("> %s. %v", query, args)
	}

	return ctx, nil
}

// convert nano to ms
func nanoToMs(ns int64) float64 {
	return math.Trunc((float64(ns)/float64(1000000)+0.5/1e2)*1e2) / 1e2
}
