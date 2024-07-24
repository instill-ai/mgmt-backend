// package errors contains domain errors that different layers can use to add
// meaning to an error and that middleware can transform to a status code or
// retry policy. This is implemented as a separate package in order to avoid
// cycle import errors.
//
// TODO jvallesm: when transforming domain errors to response codes, the
// middleware package should eventually use these errors only.
package errors

import (
	"fmt"
)

var (
	// ErrInvalidArgument is used when the provided argument is incorrect (e.g.
	// format, reserved).
	ErrInvalidArgument = fmt.Errorf("invalid argument")
	// ErrUnauthorized is used when a request can't be performed due to
	// insufficient permissions.
	ErrUnauthorized = fmt.Errorf("unauthorized")
)
