package observability

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
)

// FromContext retrieves the correlation_id from the context
// Returns empty string if not found
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// WithCorrelationID adds a correlation_id to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// EnsureCorrelationID ensures a correlation_id exists in the context
// If one doesn't exist, it generates a new UUID v4
func EnsureCorrelationID(ctx context.Context) context.Context {
	if FromContext(ctx) != "" {
		return ctx
	}
	
	newID, err := generateUUID()
	if err != nil {
		// Fallback to a simple random ID if UUID generation fails
		newID = fmt.Sprintf("fallback-%d", randomInt())
	}
	
	return WithCorrelationID(ctx, newID)
}

// LogWithCorrelationID creates a log wrapper that prefixes logs with correlation_id
func LogWithCorrelationID(ctx context.Context, logger *log.Logger) *CorrelationLogger {
	return &CorrelationLogger{
		logger:        logger,
		correlationID: FromContext(ctx),
	}
}

// CorrelationLogger wraps a standard logger to include correlation_id in messages
type CorrelationLogger struct {
	logger        *log.Logger
	correlationID string
}

// Printf formats and logs a message with correlation_id prefix
func (cl *CorrelationLogger) Printf(format string, v ...interface{}) {
	if cl.correlationID != "" {
		format = fmt.Sprintf("[correlation_id=%s] %s", cl.correlationID, format)
	}
	cl.logger.Printf(format, v...)
}

// Print logs a message with correlation_id prefix
func (cl *CorrelationLogger) Print(v ...interface{}) {
	if cl.correlationID != "" {
		args := make([]interface{}, 0, len(v)+1)
		args = append(args, fmt.Sprintf("[correlation_id=%s]", cl.correlationID))
		args = append(args, v...)
		cl.logger.Print(args...)
	} else {
		cl.logger.Print(v...)
	}
}

// Println logs a line with correlation_id prefix
func (cl *CorrelationLogger) Println(v ...interface{}) {
	if cl.correlationID != "" {
		args := make([]interface{}, 0, len(v)+1)
		args = append(args, fmt.Sprintf("[correlation_id=%s]", cl.correlationID))
		args = append(args, v...)
		cl.logger.Println(args...)
	} else {
		cl.logger.Println(v...)
	}
}

// generateUUID generates a simple UUID v4
func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}
	
	// Set version (4) and variant (10) bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 10
	
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}

// randomInt generates a simple random integer as fallback
func randomInt() int64 {
	b := make([]byte, 8)
	rand.Read(b)
	return int64(b[0])<<56 | int64(b[1])<<48 | int64(b[2])<<40 | int64(b[3])<<32 |
		int64(b[4])<<24 | int64(b[5])<<16 | int64(b[6])<<8 | int64(b[7])
}