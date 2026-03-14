package store

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// redisStore implements Store using Redis.
//
// Key scheme: "<bucket>:<key>" — bucket acts as a namespace prefix.
// DeleteBucket scans for all keys with that prefix and deletes them in one pipeline.
type redisStore struct {
	client *redis.Client
}

func openRedis(url string) (Store, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parsing Redis URL: %w", err)
	}
	c := redis.NewClient(opts)
	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connecting to Redis: %w", err)
	}
	return &redisStore{client: c}, nil
}

func (s *redisStore) Close() error { return s.client.Close() }

func redisKey(bucket, key string) string { return bucket + ":" + key }

func (s *redisStore) Get(bucket, key string) ([]byte, error) {
	val, err := s.client.Get(context.Background(), redisKey(bucket, key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (s *redisStore) Set(bucket, key string, value []byte) error {
	return s.client.Set(context.Background(), redisKey(bucket, key), value, 0).Err()
}

func (s *redisStore) Delete(bucket, key string) error {
	return s.client.Del(context.Background(), redisKey(bucket, key)).Err()
}

func (s *redisStore) Keys(bucket string) ([]string, error) {
	pattern := bucket + ":*"
	prefixLen := len(bucket) + 1 // strip "<bucket>:" prefix from returned keys

	var keys []string
	iter := s.client.Scan(context.Background(), 0, pattern, 0).Iterator()
	for iter.Next(context.Background()) {
		full := iter.Val()
		if len(full) > prefixLen {
			keys = append(keys, full[prefixLen:])
		}
	}
	return keys, iter.Err()
}

func (s *redisStore) DeleteBucket(bucket string) error {
	pattern := bucket + ":*"

	var toDelete []string
	iter := s.client.Scan(context.Background(), 0, pattern, 0).Iterator()
	for iter.Next(context.Background()) {
		toDelete = append(toDelete, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(toDelete) == 0 {
		return nil
	}
	return s.client.Del(context.Background(), toDelete...).Err()
}
