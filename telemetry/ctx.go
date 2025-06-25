package telemetry

import "context"

// contextKey is a custom type for context keys in telemetry package.
type contextKey string

const msgIDKey contextKey = "msgID"

// WithMsgID stores the message ID in the context.
func WithMsgID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, msgIDKey, id)
}

// MsgID retrieves the message ID from the context.
func MsgID(ctx context.Context) string {
	v := ctx.Value(msgIDKey)
	if id, ok := v.(string); ok {
		return id
	}
	return ""
}
