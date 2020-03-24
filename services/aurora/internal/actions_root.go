package aurora

import (
	"github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/actions"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/hcnet/go/services/aurora/internal/resourceadapter"
	"github.com/hcnet/go/support/render/hal"
)

// Interface verification
var _ actions.JSONer = (*RootAction)(nil)

// RootAction provides a summary of the aurora instance and links to various
// useful endpoints
type RootAction struct {
	Action
}

// JSON renders the json response for RootAction
func (action *RootAction) JSON() error {
	var res aurora.Root
	resourceadapter.PopulateRoot(
		action.R.Context(),
		&res,
		ledger.CurrentState(),
		action.App.auroraVersion,
		action.App.coreVersion,
		action.App.config.NetworkPassphrase,
		action.App.currentProtocolVersion,
		action.App.coreSupportedProtocolVersion,
		action.App.config.FriendbotURL,
	)

	hal.Render(action.W, res)
	return action.Err
}
