package cqs

import (
	"context"

	"github.com/sptGabriel/gb-libs/xlog"
)

func ValidationCommandInterceptor[C Command]() CommandInterceptor[C] {
	return func(ctx context.Context, cmd C, next CommandNext[C]) error {
		if err := cmd.Validate(); err != nil {
			xlog.DebugContext(
				ctx,
				"command validation failed.",
				"name", typeName[C](),
				"error", err,
			)

			return err
		}

		return next(ctx, cmd)
	}
}
