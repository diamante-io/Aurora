package problem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	ge "github.com/go-errors/errors"
	"github.com/hcnet/go/support/test"
	"github.com/stretchr/testify/assert"
)

// TestRender tests the render cases
func TestRender(t *testing.T) {
	testCases := []struct {
		name     string
		p        P
		wantList []string
		wantCode int
	}{
		{
			"server error",
			ServerError,
			[]string{"500"},
			500,
		}, {
			"renders the type correctly",
			P{Type: "foo"},
			[]string{"foo"},
			0,
		}, {
			"renders the status correctly",
			P{Status: 201},
			[]string{"201"},
			201,
		}, {
			"renders the extras correctly",
			P{Extras: map[string]interface{}{"hello": "hcnet"}},
			[]string{"hello", "hcnet"},
			0,
		},
	}

	for _, kase := range testCases {
		t.Run(kase.name, func(t *testing.T) {
			w := testRender(context.Background(), kase.p)
			for _, wantItem := range kase.wantList {
				assert.True(t, strings.Contains(w.Body.String(), wantItem), w.Body.String())
				assert.Equal(t, kase.wantCode, w.Code)
			}
		})
	}
}

// TestServerErrorConversion tests that we convert errors to ServerError problems and also log the
// stacktrace as unknown for non-rich errors
func TestServerErrorConversion(t *testing.T) {
	testCases := []struct {
		name          string
		err           error
		wantSubstring string
	}{
		{
			"non-rich errors",
			errors.New("broke"),
			"stack=unknown", // logs the stacktrace as unknown for non-rich errors
		}, {
			"rich errors",
			ge.New("broke"),
			"problem_test.go:",
		},
	}

	for _, kase := range testCases {
		t.Run(kase.name, func(t *testing.T) {
			ctx, buf := test.ContextWithLogBuffer()
			w := testRender(ctx, kase.err)
			logged := buf.String()

			assert.True(t, strings.Contains(w.Body.String(), "server_error"), w.Body.String())
			assert.Equal(t, 500, w.Code)

			// don't expose private error info in the response body
			assert.False(t, strings.Contains(w.Body.String(), "broke"), w.Body.String())

			// log additional information about the error
			assert.True(
				t,
				strings.Contains(logged, kase.wantSubstring),
				fmt.Sprintf("expecting substring: '%s' in '%s'", kase.wantSubstring, logged),
			)
		})
	}
}

// TestInflate test errors that come inflated from aurora
func TestInflate(t *testing.T) {
	testCase := struct {
		name string
		p    P
		want string
	}{
		"renders the type correctly",
		P{Type: "https://hcnet.org/aurora-errors/not_found"},
		"https://hcnet.org/aurora-errors/not_found",
	}

	t.Run(testCase.name, func(t *testing.T) {
		w := testRender(context.Background(), testCase.p)
		var payload P
		err := json.Unmarshal([]byte(w.Body.String()), &payload)
		if assert.NoError(t, err) {
			assert.Equal(t, testCase.want, payload.Type)
		}
	})
}

func testRender(ctx context.Context, err error) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	Render(ctx, w, err)
	return w
}
