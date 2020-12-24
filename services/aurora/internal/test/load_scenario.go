package test

import (
	"github.com/diamnet/go/services/aurora/internal/test/scenarios"
)

func loadScenario(scenarioName string, includeAurora bool) {
	diamnetCorePath := scenarioName + "-core.sql"
	auroraPath := scenarioName + "-aurora.sql"

	if !includeAurora {
		auroraPath = "blank-aurora.sql"
	}

	scenarios.Load(DiamNetCoreDatabaseURL(), diamnetCorePath)
	scenarios.Load(DatabaseURL(), auroraPath)
}
