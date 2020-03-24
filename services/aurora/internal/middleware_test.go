package aurora

import (
	"strconv"
	"testing"

	"github.com/hcnet/go/services/aurora/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/throttled/throttled"
)

type RateLimitMiddlewareTestSuite struct {
	suite.Suite
	ht  *HTTPT
	c   Config
	app *App
	rh  test.RequestHelper
}

func (suite *RateLimitMiddlewareTestSuite) SetupSuite() {
	suite.ht = StartHTTPTest(suite.T(), "base")
}

func (suite *RateLimitMiddlewareTestSuite) SetupTest() {
	suite.c = NewTestConfig()
	suite.c.RateQuota = &throttled.RateQuota{
		MaxRate:  throttled.PerHour(10),
		MaxBurst: 9,
	}
	suite.app = NewApp(suite.c)
	suite.rh = NewRequestHelper(suite.app)
}

func (suite *RateLimitMiddlewareTestSuite) TearDownSuite() {
	suite.ht.Finish()
}

func (suite *RateLimitMiddlewareTestSuite) TearDownTest() {
	suite.app.Close()
}

// Sets X-RateLimit-Limit headers correctly.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_LimitHeaders() {
	w := suite.rh.Get("/")
	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), "10", w.Header().Get("X-RateLimit-Limit"))
}

// Sets X-RateLimit-Remaining headers correctly.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_RemainingHeaders() {
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/")
		expected := 10 - (i + 1)
		assert.Equal(suite.T(), strconv.Itoa(expected), w.Header().Get("X-RateLimit-Remaining"))
	}

	// confirm remaining stays at 0
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/")
		assert.Equal(suite.T(), "0", w.Header().Get("X-RateLimit-Remaining"))
	}
}

// Sets X-RateLimit-Reset header correctly. Should reset after 360 seconds since it's limited to 10 requests/hour.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_ResetHeaders() {
	w := suite.rh.Get("/")
	assert.Equal(suite.T(), "360", w.Header().Get("X-RateLimit-Reset"))
}

// Restricts based on RemoteAddr IP after too many requests.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_RemoteAddr() {
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/")
		assert.Equal(suite.T(), 200, w.Code)
	}

	w := suite.rh.Get("/")
	assert.Equal(suite.T(), 429, w.Code)

	w = suite.rh.Get("/", test.RequestHelperRemoteAddr("127.0.0.2"))
	assert.Equal(suite.T(), 200, w.Code)

	// Ignores ports
	w = suite.rh.Get("/", test.RequestHelperRemoteAddr("127.0.0.1:4312"))
	assert.Equal(suite.T(), 429, w.Code)
}

// Restrict based upon X-Forwarded-For correctly.
func (suite *RateLimitMiddlewareTestSuite) TestRateLimit_XForwardedFor() {
	for i := 0; i < 10; i++ {
		w := suite.rh.Get("/", test.RequestHelperXFF("4.4.4.4"))
		assert.Equal(suite.T(), 200, w.Code)
	}

	w := suite.rh.Get("/", test.RequestHelperXFF("4.4.4.4"))
	assert.Equal(suite.T(), 429, w.Code)

	// allow other ips
	w = suite.rh.Get("/", test.RequestHelperRemoteAddr("4.4.4.3"))
	assert.Equal(suite.T(), 200, w.Code)

	// Ignores leading private ips
	w = suite.rh.Get("/", test.RequestHelperXFF("10.0.0.1, 4.4.4.4"))
	assert.Equal(suite.T(), 429, w.Code)

	// Ignores trailing ips
	w = suite.rh.Get("/", test.RequestHelperXFF("4.4.4.4, 4.4.4.5, 127.0.0.1"))
	assert.Equal(suite.T(), 429, w.Code)
}

func TestRateLimitMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimitMiddlewareTestSuite))
}

// Rate Limiting works with redis
func TestRateLimit_Redis(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()
	c := NewTestConfig()
	c.RateQuota = &throttled.RateQuota{
		MaxRate:  throttled.PerHour(10),
		MaxBurst: 9,
	}
	c.RedisURL = "redis://127.0.0.1:6379/"
	app := NewApp(c)
	defer app.Close()
	rh := NewRequestHelper(app)

	redis := app.redis.Get()
	_, err := redis.Do("FLUSHDB")
	assert.Nil(t, err)

	for i := 0; i < 10; i++ {
		w := rh.Get("/")
		assert.Equal(t, 200, w.Code)
	}

	w := rh.Get("/")
	assert.Equal(t, 429, w.Code)

	w = rh.Get("/", test.RequestHelperRemoteAddr("127.0.0.2"))
	assert.Equal(t, 200, w.Code)
}
