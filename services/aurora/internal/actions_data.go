package aurora

import (
	"github.com/hcnet/go/services/aurora/internal/actions"
	"github.com/hcnet/go/services/aurora/internal/db2/core"
	"github.com/hcnet/go/services/aurora/internal/render/sse"
	"github.com/hcnet/go/support/render/hal"
)

// Interface verifications
var _ actions.JSONer = (*DataShowAction)(nil)
var _ actions.RawDataResponder = (*DataShowAction)(nil)
var _ actions.EventStreamer = (*DataShowAction)(nil)

// DataShowAction renders a account summary found by its address.
type DataShowAction struct {
	Action
	Address string
	Key     string
	Data    core.AccountData
}

// JSON is a method for actions.JSON
func (action *DataShowAction) JSON() error {
	action.Do(
		action.loadParams,
		action.loadRecord,
		func() { hal.Render(action.W, map[string]string{"value": action.Data.Value}) },
	)
	return action.Err
}

// Raw is a method for actions.Raw
func (action *DataShowAction) Raw() error {
	action.Do(
		action.loadParams,
		action.loadRecord,
		func() {
			var raw []byte
			raw, action.Err = action.Data.Raw()
			if action.Err != nil {
				return
			}

			action.W.Write(raw)
		},
	)
	return action.Err
}

// SSE is a method for actions.SSE
func (action *DataShowAction) SSE(stream *sse.Stream) error {
	action.Do(
		action.loadParams,
		action.loadRecord,
		func() {
			stream.Send(sse.Event{Data: action.Data.Value})
		},
	)
	return action.Err
}

func (action *DataShowAction) loadParams() {
	action.Address = action.GetAddress("account_id", actions.RequiredParam)
	action.Key = action.GetString("key")
}

func (action *DataShowAction) loadRecord() {
	action.Err = action.CoreQ().AccountDataByKey(&action.Data, action.Address, action.Key)
}
