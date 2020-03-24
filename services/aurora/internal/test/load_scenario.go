package test

import (
	"github.com/hcnet/go/services/aurora/internal/test/scenarios"
)

func loadScenario(scenarioName string, includeAurora bool) {
	hcnetCorePath := scenarioName + "-core.sql"
	auroraPath := scenarioName + "-aurora.sql"

	if !includeAurora {
		auroraPath = "blank-aurora.sql"
	}

	scenarios.Load(HcNetCoreDatabaseURL(), hcnetCorePath)
	scenarios.Load(DatabaseURL(), auroraPath)
}
