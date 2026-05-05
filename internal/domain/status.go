package domain

import "errors"

// Status represents the lifecycle state of a Task.
type Status string

const (
	StatusPending    Status = "PENDING"
	StatusInProgress Status = "IN_PROGRESS"
	StatusDone       Status = "DONE"
)

// validStatuses is the set of all allowed Status values.
var validStatuses = map[Status]struct{}{
	StatusPending:    {},
	StatusInProgress: {},
	StatusDone:       {},
}

// IsValid reports whether s is a known Status value.
func (s Status) IsValid() bool {
	_, ok := validStatuses[s]
	return ok
}

// ParseStatus converts a raw string into a Status, returning an error if it
// does not match any of the known values.
func ParseStatus(raw string) (Status, error) {
	s := Status(raw)
	if !s.IsValid() {
		return "", errors.New("status must be one of PENDING, IN_PROGRESS, DONE")
	}
	return s, nil
}
