package problem

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/diamnet/go/support/render/problem"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()
var testRender = func(ctx context.Context, err error) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	problem.Render(ctx, w, err)
	return w
}

func TestCommonProblems(t *testing.T) {
	testCases := []struct {
		testName     string
		p            problem.P
		expectedCode int
	}{
		{"NotFound", problem.NotFound, 404},
		{"RateLimitExceeded", RateLimitExceeded, 429},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			w := testRender(ctx, tc.p)
			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}
