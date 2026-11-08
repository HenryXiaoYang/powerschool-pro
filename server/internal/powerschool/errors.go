// Package powerschool holds what the mobile and portal clients share: the
// error kinds callers branch on.
package powerschool

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidCredentials means PowerSchool rejected the username or password.
	ErrInvalidCredentials = errors.New("powerschool: invalid credentials")
	// ErrSessionExpired means a ticket or portal cookie is no longer accepted; log in again.
	ErrSessionExpired = errors.New("powerschool: session expired")
	// ErrNotFound means the requested record does not exist for this account.
	ErrNotFound = errors.New("powerschool: not found")
	// ErrUpstream means PowerSchool was unreachable or answered with something unusable.
	ErrUpstream = errors.New("powerschool: upstream error")
)

// Upstreamf wraps ErrUpstream with a sanitized detail. Never pass response bodies,
// tickets, or credentials as arguments.
func Upstreamf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrUpstream, fmt.Sprintf(format, args...))
}
