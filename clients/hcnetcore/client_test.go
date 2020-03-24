package hcnetcore

import (
	"context"
	"net/http"
	"testing"

	proto "github.com/hcnet/go/protocols/hcnetcore"
	"github.com/hcnet/go/support/http/httptest"
	"github.com/stretchr/testify/assert"
)

func TestSubmitTransaction(t *testing.T) {
	hmock := httptest.NewClient()
	c := &Client{HTTP: hmock, URL: "http://localhost:11626"}

	// happy path - new transaction
	hmock.On("GET", "http://localhost:11626/tx?blob=foo").
		ReturnJSON(http.StatusOK, proto.TXResponse{
			Status: proto.TXStatusPending,
		})

	resp, err := c.SubmitTransaction(context.Background(), "foo")

	if assert.NoError(t, err) {
		assert.Equal(t, proto.TXStatusPending, resp.Status)
	}
}
