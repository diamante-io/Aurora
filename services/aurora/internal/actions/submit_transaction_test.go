package actions

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/diamnet/go/network"
	"github.com/diamnet/go/services/aurora/internal/corestate"
	"github.com/diamnet/go/support/render/problem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDiamnetCoreMalformedTx(t *testing.T) {
	handler := SubmitTransactionHandler{}

	r := httptest.NewRequest("POST", "https://aurora.diamnet.org/transactions", nil)
	w := httptest.NewRecorder()
	_, err := handler.GetResource(w, r)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, err.(*problem.P).Status)
	assert.Equal(t, "Transaction Malformed", err.(*problem.P).Title)
}

type coreStateGetterMock struct {
	mock.Mock
}

func (m *coreStateGetterMock) GetCoreState() corestate.State {
	a := m.Called()
	return a.Get(0).(corestate.State)
}

func TestDiamnetCoreNotSynced(t *testing.T) {
	mock := &coreStateGetterMock{}
	mock.On("GetCoreState").Return(corestate.State{
		Synced: false,
	})

	handler := SubmitTransactionHandler{
		NetworkPassphrase: network.PublicNetworkPassphrase,
		CoreStateGetter:   mock,
	}

	form := url.Values{}
	form.Set("tx", "AAAAAAGUcmKO5465JxTSLQOQljwk2SfqAJmZSG6JH6wtqpwhAAABLAAAAAAAAAABAAAAAAAAAAEAAAALaGVsbG8gd29ybGQAAAAAAwAAAAAAAAAAAAAAABbxCy3mLg3hiTqX4VUEEp60pFOrJNxYM1JtxXTwXhY2AAAAAAvrwgAAAAAAAAAAAQAAAAAW8Qst5i4N4Yk6l+FVBBKetKRTqyTcWDNSbcV08F4WNgAAAAAN4Lazj4x61AAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABLaqcIQAAAEBKwqWy3TaOxoGnfm9eUjfTRBvPf34dvDA0Nf+B8z4zBob90UXtuCqmQqwMCyH+okOI3c05br3khkH0yP4kCwcE")

	request, err := http.NewRequest(
		"POST",
		"https://aurora.diamnet.org/transactions",
		strings.NewReader(form.Encode()),
	)
	require.NoError(t, err)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	_, err = handler.GetResource(w, request)
	assert.Error(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, err.(problem.P).Status)
	assert.Equal(t, "stale_history", err.(problem.P).Type)
	assert.Equal(t, "Historical DB Is Too Stale", err.(problem.P).Title)
}
