package redisstore

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

func TestFind(t *testing.T) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		addr := os.Getenv("SESSION_REDIS_TEST_ADDR")
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 1)
	defer redisPool.Close()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Do("SET", Prefix+"session_token", "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	r := New(redisPool)

	b, found, err := r.Find("session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestSaveNew(t *testing.T) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		addr := os.Getenv("SESSION_REDIS_TEST_ADDR")
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 1)
	defer redisPool.Close()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Fatal(err)
	}

	r := New(redisPool)

	err = r.Save("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	data, err := redis.Bytes(conn.Do("GET", Prefix+"session_token"))
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(data, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		addr := os.Getenv("SESSION_REDIS_TEST_ADDR")
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 1)
	defer redisPool.Close()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Fatal(err)
	}

	r := New(redisPool)

	_, found, err := r.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveUpdated(t *testing.T) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		addr := os.Getenv("SESSION_REDIS_TEST_ADDR")
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 1)
	defer redisPool.Close()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Do("SET", Prefix+"session_token", "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	r := New(redisPool)

	err = r.Save("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	data, err := redis.Bytes(conn.Do("GET", Prefix+"session_token"))
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		addr := os.Getenv("SESSION_REDIS_TEST_ADDR")
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 1)
	defer redisPool.Close()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Fatal(err)
	}

	r := New(redisPool)

	err = r.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := r.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(200 * time.Millisecond)
	_, found, _ = r.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		addr := os.Getenv("SESSION_REDIS_TEST_ADDR")
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 1)
	defer redisPool.Close()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Do("SET", Prefix+"session_token", "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	r := New(redisPool)

	err = r.Delete("session_token")
	if err != nil {
		t.Fatal(err)
	}

	data, err := conn.Do("GET", Prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Fatalf("got %v: expected %v", data, nil)
	}
}

func TestDeleteByPattern(t *testing.T) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		addr := os.Getenv("SESSION_REDIS_TEST_ADDR")
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 1)
	defer redisPool.Close()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	if err != nil {
		t.Fatal(err)
	}

	// token 1
	firstToken := "session_token_1"
	_, err = conn.Do("SET", Prefix+firstToken, "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	// token 2
	secondToken := "session_token_2"
	_, err = conn.Do("SET", Prefix+secondToken, "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	r := New(redisPool)

	err = r.DeleteByPattern("session_token_")
	if err != nil {
		t.Fatal(err)
	}

	data, err := redis.MultiBulk(conn.Do("MGET", Prefix+firstToken, Prefix+secondToken))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 2 {
		t.Fatalf("len(data) %v: expected %v", len(data), 2)
	}

	if data[0] != nil {
		t.Fatalf("data[0] got %v: expected %v", data, nil)
	}

	if data[1] != nil {
		t.Fatalf("data[1] got %v: expected %v", data, nil)
	}
}
