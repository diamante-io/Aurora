package aurora

import (
	"net/http"

	hProblem "github.com/diamnet/go/services/aurora/internal/render/problem"
	"github.com/diamnet/go/support/render/problem"
)

// RateLimitExceededAction renders a 429 response
type RateLimitExceededAction struct {
	Action
}

func (action RateLimitExceededAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ap := &action.Action
	ap.Prepare(w, r)
	problem.Render(action.R.Context(), action.W, hProblem.RateLimitExceeded)
}
