package xlog

import (
	"context"
	"time"
)

type attributeCtxKeyType struct{}

var attributeCtxKey = attributeCtxKeyType{}

// Attribute represents a key-value pair for structured logging
type Attribute struct {
	Key   string
	Value interface{}
}

// String constructs a string-valued logging attribute
// Provides type safety for string data in structured logs
// Usage example: String("user_id", "abc123")
// Results in log output containing the field: "user_id": "abc123"
func String(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

// Float64 constructs a floating point logging attribute
// Useful for recording measurements, ratios, or decimal values
// Usage example: Float64("response_time_ms", 42.7)
// Results in log output containing: "response_time_ms": 42.7
func Float64(key string, value float64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int constructs an integer-valued logging attribute
// Appropriate for counts, quantities and whole number metrics
// Usage example: Int("status_code", 200)
// Results in log output containing: "status_code": 200
func Int(key string, value int) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int64 constructs a 64-bit integer logging attribute
// Suitable for large numeric identifiers or timestamps
// Usage example: Int64("user_count", 9876543210)
// Results in log output containing: "user_count": 9876543210
func Int64(key string, value int64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Bool constructs a boolean logging attribute
// Ideal for flags, feature toggles, or state indicators
// Usage example: Bool("request_successful", true)
// Results in log output containing: "request_successful": true
func Bool(key string, value bool) Attribute {
	return Attribute{Key: key, Value: value}
}

// Any constructs a logging attribute from any value type
// Provides flexibility for complex or custom data types
// Usage example: Any("config", settings)
// Results will vary based on the value's representation
func Any(key string, value interface{}) Attribute {
	return Attribute{Key: key, Value: value}
}

// Duration constructs a time interval logging attribute
// Formats the duration as a human-readable string
// Usage example: Duration("processing_time", 250*time.Millisecond)
// Results in log output containing: "processing_time": "250ms"
func Duration(key string, value time.Duration) Attribute {
	return Attribute{Key: key, Value: value.String()}
}

// Error constructs an error-specific logging attribute
// Automatically extracts the error message for consistent logging
// Usage example: Error(err)
// Results in log output containing: "error": "connection refused"
func Error(err error) Attribute {
	return Attribute{Key: "error", Value: err.Error()}
}

// WithAttributes enriches the context with logging metadata
// The attributes are preserved across function boundaries when the context is passed
// New attributes are appended to any existing ones in the context
func WithAttributes(ctx context.Context, attrs ...Attribute) context.Context {
	attrsInContext, _ := ctx.Value(attributeCtxKey).([]Attribute)
	return context.WithValue(ctx, attributeCtxKey, append(attrsInContext, attrs...))
}

// AttributesFromContext extracts all logging attributes stored in a context
// Returns an empty slice if no attributes are found
// This allows accessing context-bound logging metadata throughout a request lifecycle
func AttributesFromContext(ctx context.Context) []Attribute {
	if attrs, ok := ctx.Value(attributeCtxKey).([]Attribute); ok {
		return attrs
	}

	return []Attribute{}
}
