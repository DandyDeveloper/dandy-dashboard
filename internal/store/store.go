// Package store provides a backend-agnostic key-value store used by dashboard widgets.
//
// Two backends are available:
//   - bbolt (default) — embedded, file-based, no external service required.
//   - Redis           — external, set STORE_URL=redis://... to enable.
//
// The active backend is selected by New().
package store

// Store is the persistence interface every widget uses.
// Keys are namespaced by a bucket string (analogous to a Redis key prefix or
// a bbolt bucket), so widgets cannot accidentally collide with each other.
type Store interface {
	// Get returns the value for key within bucket, or nil if not found.
	Get(bucket, key string) ([]byte, error)

	// Set stores value under key in bucket.
	Set(bucket, key string, value []byte) error

	// Delete removes key from bucket (no-op if either doesn't exist).
	Delete(bucket, key string) error

	// Keys returns every key that belongs to bucket.
	Keys(bucket string) ([]string, error)

	// DeleteBucket removes all keys in bucket (used to reset a word cycle).
	DeleteBucket(bucket string) error

	// Close releases resources held by the backend.
	Close() error
}
