package aurora

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestInitRedis(t *testing.T) {
	c := NewTestConfig()

	// app.redis is nil when no RedisURL is set
	c.RedisURL = ""
	app := NewApp(c)
	assert.Nil(t, app.redis)
	app.Close()

	// app.redis gets set when RedisURL is set
	c.RedisURL = "redis://127.0.0.1:6379/"
	app = NewApp(c)
	assert.NotNil(t, app.redis)

	// redis connection works
	conn := app.redis.Get()
	conn.Do("SET", "hello", "World")
	world, _ := redis.String(conn.Do("GET", "hello"))

	assert.Equal(t, "World", world)

	conn.Close()
	app.Close()
}
