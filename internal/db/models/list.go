package models

// ListOptions provides pagination and ordering for list queries.
// Defined here so both the Store interface (db package) and repository
// implementations (db/repositories) can reference it without circular imports.
type ListOptions struct {
	Limit  int
	Offset int
	Order  string // e.g. "created_at DESC"
}
