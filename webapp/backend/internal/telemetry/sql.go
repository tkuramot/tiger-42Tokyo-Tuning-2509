package telemetry

import (
	"database/sql"
	"log"

	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func WrapSQLDriver(baseDriver string) string {
	if !enabled() {
		return baseDriver
	}
	name, err := otelsql.Register(baseDriver,
		otelsql.WithAttributes(semconv.DBSystemKey.String(baseDriver)),
		otelsql.WithSQLCommenter(true),
		otelsql.WithSpanOptions(otelsql.SpanOptions{DisableErrSkip: true}),
	)
	if err != nil {
		log.Printf("otelsql.Register failed, fallback to base driver: %v", err)
		return baseDriver
	}
	return name
}

var _ = sql.ErrNoRows
