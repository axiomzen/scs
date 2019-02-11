// Package redisstore is a Redis-based session store for the SCS session package.
//
// Warning: The redisstore API is not finalized and may change, possibly significantly.
// The package is fine to use as-is, but it is strongly recommended that you vendor
// the package to avoid compatibility problems in the future.
//
// The redisstore package relies on the the popular Redigo Redis client
// (github.com/gomodule/redigo/redis).
package redisstore

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// Prefix controls the Redis key prefix. You should only need to change this if there is
// a naming clash.
var Prefix = "scs:session:"

// RedisStore represents the currently configured session session store. It is essentially
// a wrapper around a Redigo connection pool.
type RedisStore struct {
	pool *redis.Pool
}

// New returns a new RedisStore instance. The pool parameter should be a pointer to a
// Redigo connection pool. See https://godoc.org/github.com/garyburd/redigo/redis#Pool.
func New(pool *redis.Pool) *RedisStore {
	return &RedisStore{pool}
}

// Find returns the data for a given session token from the RedisStore instance. If the session
// token is not found or is expired, the returned exists flag will be set to false.
func (r *RedisStore) Find(token string) (b []byte, exists bool, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	b, err = redis.Bytes(conn.Do("GET", Prefix+token))
	if err == redis.ErrNil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Save adds a session token and data to the RedisStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (r *RedisStore) Save(token string, b []byte, expiry time.Time) error {
	conn := r.pool.Get()
	defer conn.Close()

	err := conn.Send("MULTI")
	if err != nil {
		return err
	}
	err = conn.Send("SET", Prefix+token, b)
	if err != nil {
		return err
	}
	err = conn.Send("PEXPIREAT", Prefix+token, makeMillisecondTimestamp(expiry))
	if err != nil {
		return err
	}
	_, err = conn.Do("EXEC")
	return err
}

// Delete removes a session token and corresponding data from the RedisStore instance.
func (r *RedisStore) Delete(token string) error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", Prefix+token)
	return err
}

// DeleteByPattern removes all tokens that match the pattern from the RedisStore instance
func (r *RedisStore) DeleteByPattern(pattern string) error {
	conn := r.pool.Get()
	defer conn.Close()

	matchPattern := Prefix + pattern + "*"

	itr := 0
	var keys []string

	for {
		// we scan with our itr offset, starting at 0
		arr, err := redis.MultiBulk(conn.Do("SCAN", itr, "MATCH", matchPattern))
		if err != nil {
			return err
		}

		// fetch itr and keys from the multi-bulk reply
		itr, _ = redis.Int(arr[0], nil)
		keys, _ = redis.Strings(arr[1], nil)

		// when SCAN returns a 0 when it is complete
		if itr == 0 {
			break
		}
	}

	// Unlink all the found keys. Unlink is more efficient than DEL
	// see: https://redis.io/commands/unlink
	for _, key := range keys {
		_, err := conn.Do("UNLINK", key)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeMillisecondTimestamp(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
