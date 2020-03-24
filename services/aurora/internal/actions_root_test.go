package aurora

import (
	"encoding/json"
	"testing"

	"github.com/hcnet/go/protocols/aurora"
	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestRootAction(t *testing.T) {
	ht := StartHTTPTest(t, "base")
	defer ht.Finish()

	server := test.NewStaticMockServer(`{
			"info": {
				"network": "test",
				"build": "test-core",
				"ledger": {
					"version": 3
				},
				"protocol_version": 4
			}
		}`)
	defer server.Close()

	ht.App.auroraVersion = "test-aurora"
	ht.App.config.HcNetCoreURL = server.URL
	ht.App.config.NetworkPassphrase = "test"
	ht.App.UpdateHcNetCoreInfo()

	w := ht.Get("/")

	if ht.Assert.Equal(200, w.Code) {
		var actual aurora.Root
		err := json.Unmarshal(w.Body.Bytes(), &actual)
		ht.Require.NoError(err)
		ht.Assert.Equal("test-aurora", actual.AuroraVersion)
		ht.Assert.Equal("test-core", actual.HcNetCoreVersion)
		ht.Assert.Equal(int32(4), actual.CoreSupportedProtocolVersion)
		ht.Assert.Equal(int32(3), actual.CurrentProtocolVersion)
	}
}
