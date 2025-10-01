package workflows_test

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/samber/oops"

	"github.com/drornir/better-actions/pkg/log"
)

//go:embed *.yaml
var rootFs embed.FS

func makeContext(t *testing.T, logLevel slog.Level, logAttrs ...any) context.Context {
	ctx := t.Context()
	ctx = log.New(slog.New(slog.NewTextHandler(
		t.Output(),
		&slog.HandlerOptions{
			Level: logLevel,
		})),
	).With(logAttrs...).ContextWithLogger(ctx)

	return ctx
}

func errParse(err error) string {
	if err == nil {
		return "<nil>"
	}

	var oo oops.OopsError
	if !errors.As(err, &oo) {
		return err.Error()
	}

	// var b bytes.Buffer
	// slogH := slog.NewTextHandler(&b, &slog.HandlerOptions{})
	// logValue := oo.LogValue()
	// serr := slogH.WithAttrs(logValue.Group()).Handle(context.Background(), slog.Record{})
	// if serr != nil {
	// 	panic(serr)
	// }

	// res = append(res, b.String())

	res := []any{oo.Error()}
	for k, v := range oo.Context() {
		if k == "message" {
			continue
		}
		res = append(res, " ", k, "=", fmt.Sprintf("'%v'", v))
	}
	res = append(res, "\nstacktrace:\n", oo.Stacktrace())
	return fmt.Sprint(res...)
}
