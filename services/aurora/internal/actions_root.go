package aurora

import (
	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/actions"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
	"github.com/diamnet/go/support/render/hal"
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
