package diamnettoml

import "log"

// ExampleGetTOML gets the diamnet.toml file for coins.asia
func ExampleClient_GetDiamnetToml() {
	_, err := DefaultClient.GetDiamnetToml("coins.asia")
	if err != nil {
		log.Fatal(err)
	}
}
