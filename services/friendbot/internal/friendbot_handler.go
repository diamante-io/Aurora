package internal

import (
	"net/http"
	"net/url"

	"github.com/diamnet/go/clients/aurora"
	"github.com/diamnet/go/strkey"
	"github.com/diamnet/go/support/render/hal"
	"github.com/diamnet/go/support/render/problem"
)

// FriendbotHandler causes an account at `Address` to be created.
type FriendbotHandler struct {
	Friendbot *Bot
}

// Handle is a method that implements http.HandlerFunc
func (handler *FriendbotHandler) Handle(w http.ResponseWriter, r *http.Request) {
	accountExistsProblem := problem.BadRequest
	accountExistsProblem.Detail = ErrAccountExists.Error()
	problem.RegisterError(ErrAccountExists, accountExistsProblem)

	result, err := handler.doHandle(r)
	if err != nil {
		problem.Render(r.Context(), w, err)
		return
	}

	hal.Render(w, *result)
}

// doHandle is just a convenience method that returns the object to be rendered
func (handler *FriendbotHandler) doHandle(r *http.Request) (*aurora.TransactionSuccess, error) {
	err := handler.checkEnabled()
	if err != nil {
		return nil, err
	}

	err = r.ParseForm()
	if err != nil {
		return nil, err
	}

	address, err := handler.loadAddress(r)
	if err != nil {
		return nil, problem.MakeInvalidFieldProblem("addr", err)
	}
	return handler.loadResult(address)
}

func (handler *FriendbotHandler) checkEnabled() error {
	if handler.Friendbot != nil {
		return nil
	}

	return &problem.P{
		Type:   "friendbot_disabled",
		Title:  "Friendbot is disabled",
		Status: http.StatusForbidden,
		Detail: "Friendbot is disabled on this network. Contact the server administrator if you believe this to be in error.",
	}
}

func (handler *FriendbotHandler) loadAddress(r *http.Request) (string, error) {
	address := r.Form.Get("addr")
	unescaped, err := url.QueryUnescape(address)
	if err != nil {
		return unescaped, err
	}

	_, err = strkey.Decode(strkey.VersionByteAccountID, unescaped)
	return unescaped, err
}

func (handler *FriendbotHandler) loadResult(address string) (*aurora.TransactionSuccess, error) {
	result, err := handler.Friendbot.Pay(address)
	switch e := err.(type) {
	case aurora.Error:
		return result, e.Problem.ToProblem()
	case *aurora.Error:
		return result, e.Problem.ToProblem()
	}
	return result, err
}
