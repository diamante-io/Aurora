package actions

import "github.com/diamnet/go/services/aurora/internal/corestate"

type CoreStateGetter interface {
	GetCoreState() corestate.State
}
