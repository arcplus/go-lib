package mysql

import (
	"context"
	"time"

	"github.com/arcplus/go-lib/log"
)

// Hooks satisfies the sqlhook.Hooks interface
type Hooks struct{}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, "x-sql-begin", time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	begin := ctx.Value("x-sql-begin").(time.Time)

	logger := log.KVPair(map[string]string{
		"hook": "sql",
		"took": time.Since(begin).String(),
	})

	if tid, ok := ctx.Value("x-request-id").(string); ok {
		logger = logger.KV("tid", tid)
	}

	logger.Debugf("> %s %q", query, args)
	return ctx, nil
}
