package scs

import "time"

// Store is the interface for session stores.
type Store interface {
	// Delete should remove the session token and corresponding data from the
	// session store. If the token does not exist then Delete should be a no-op
	// and return nil (not an error).
	Delete(token string) (err error)

	// DeleteByPattern will delete all keys that match the provided pattern. If the
	// pattern doesn't match any keys then DeleteByPattern will no-op and return nil
	// (not an error).
	DeleteByPattern(pattern string) (err error)

	// Find should return the data for a session token from the session store.
	// If the session token is not found or is expired, the found return value
	// should be false (and the err return value should be nil). Similarly, tampered
	// or malformed tokens should result in a found return value of false and a
	// nil err value. The err return value should be used for system errors only.
	Find(token string) (b []byte, found bool, err error)

	// Save should add the session token and data to the session store, with
	// the given expiry time. If the session token already exists, then the data
	// and expiry time should be overwritten.
	Save(token string, b []byte, expiry time.Time) (err error)
}

type cookieStore interface {
	MakeToken(b []byte, expiry time.Time) (token string, err error)
}
