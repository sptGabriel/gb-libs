package xpg

import (
	"context"

	"github.com/jackc/pgx/v5"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type QuerierTracer struct {
	tracer trace.Tracer
}

func NewQuerierTracer(tracer trace.Tracer) QuerierTracer {
	return QuerierTracer{tracer: tracer}
}

func (qt QuerierTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, span := qt.tracer.Start(ctx, "database.sql.pg.call")
	span.SetAttributes(
		semconv.DBQueryText(data.SQL),
		semconv.DBSystemPostgreSQL,
		semconv.DBNamespace(conn.Config().Database),
		semconv.NetworkPeerAddress(conn.Config().Host),
	)

	return ctx
}

func (qt QuerierTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	trace.SpanFromContext(ctx).End()
}
