package ctxkit

import (
	"context"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/log"
)

func With(ctx context.Context, keyvals ...any) (context.Context, *log.Logger, oops.OopsErrorBuilder) {
	if len(keyvals) == 0 {
		return ctx, log.FromContext(ctx), oops.FromContext(ctx)
	}
	oopser := oops.FromContext(ctx).With(keyvals...)
	logger := log.FromContext(ctx).With(keyvals...)
	ctx = logger.ContextWithLogger(ctx)
	ctx = oops.WithBuilder(ctx, oopser)
	return ctx, logger, oopser
}
