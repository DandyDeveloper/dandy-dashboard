package store

import (
	"fmt"
	"path/filepath"
	"strings"
)

// New opens a Store using the backend selected by storeURL:
//
//   - ""                      → embedded bolt DB at <dataDir>/dashboard.db
//   - "redis://..."           → external Redis
//   - "rediss://..."          → external Redis over TLS
//
// dataDir is only used when falling back to the embedded bolt backend.
func New(storeURL, dataDir string) (Store, error) {
	if strings.HasPrefix(storeURL, "redis://") || strings.HasPrefix(storeURL, "rediss://") {
		return openRedis(storeURL)
	}
	if storeURL != "" {
		return nil, fmt.Errorf("unrecognised STORE_URL scheme %q (supported: redis://, rediss://)", storeURL)
	}
	return openBolt(filepath.Join(dataDir, "dashboard.db"))
}
