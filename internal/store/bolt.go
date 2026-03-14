package store

import (
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// boltStore is the embedded, file-backed Store implementation.
type boltStore struct {
	db *bolt.DB
}

func openBolt(path string) (Store, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("opening bolt store at %s: %w", path, err)
	}
	return &boltStore{db: db}, nil
}

func (s *boltStore) Close() error { return s.db.Close() }

func (s *boltStore) Get(bucket, key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		if v != nil {
			val = make([]byte, len(v))
			copy(val, v)
		}
		return nil
	})
	return val, err
}

func (s *boltStore) Set(bucket, key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func (s *boltStore) Delete(bucket, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}

func (s *boltStore) Keys(bucket string) ([]string, error) {
	var keys []string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}

func (s *boltStore) DeleteBucket(bucket string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(bucket))
		if err == bolt.ErrBucketNotFound {
			return nil
		}
		return err
	})
}
