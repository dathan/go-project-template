package model

import "time"

// User represents an authenticated end-user of the system. Users are
// uniquely identified by Provider and ProviderID fields which map to
// records in the upstream identity provider. Additional fields like
// Email and Name can be enriched from the provider's profile API.
type User struct {
    ID         string    // UUID primary key
    Provider   string    // e.g. "google", "slack", "linkedin"
    ProviderID string    // unique identifier returned by provider
    Email      string    // user email address (if available)
    Name       string    // user full name
    CreatedAt  time.Time // record creation timestamp
    UpdatedAt  time.Time // last update timestamp
}