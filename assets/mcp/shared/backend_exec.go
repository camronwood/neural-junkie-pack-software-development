package shared

import (
	"context"
)

type backendKey struct{}
type implementationSessionKey struct{}

// ContextWithImplementationSession marks tool calls that may use broader run_command allowlists.
func ContextWithImplementationSession(ctx context.Context, enabled bool) context.Context {
	if !enabled {
		return ctx
	}
	return context.WithValue(ctx, implementationSessionKey{}, true)
}

// ImplementationSessionFromContext reports whether implementation-session commands are allowed.
func ImplementationSessionFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v, _ := ctx.Value(implementationSessionKey{}).(bool)
	return v
}

// ContextWithBackend is a no-op in the pack sidecar (local exec only).
func ContextWithBackend(ctx context.Context, _ any) context.Context {
	return ctx
}

// BackendFromContext always returns nil in the pack sidecar.
func BackendFromContext(ctx context.Context) any {
	return nil
}

// RunCommandMaybeRemote runs commands locally in the pack sidecar.
func RunCommandMaybeRemote(ctx context.Context, dir, name string, args ...string) (string, error) {
	return RunCommand(ctx, dir, name, args...)
}
