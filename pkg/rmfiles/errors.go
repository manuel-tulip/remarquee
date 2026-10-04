package rmfiles

import "github.com/pkg/errors"

// Sentinel errors returned by the service. The HTTP layer maps these to status
// codes; keep them coarse and stable.
var (
	// ErrNotFound means a referenced entry does not exist in the cloud tree.
	ErrNotFound = errors.New("entry not found")
	// ErrConflict means an entry with the requested name already exists.
	ErrConflict = errors.New("entry already exists")
	// ErrInvalid means the request is malformed (bad name, empty input, ...).
	ErrInvalid = errors.New("invalid request")
	// ErrNotDir means a path that should be a directory resolves to a file.
	ErrNotDir = errors.New("not a directory")
	// ErrUnsupported means an operation is not supported for this entry/file.
	ErrUnsupported = errors.New("unsupported")
)
