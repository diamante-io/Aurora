package diamnettoml

import "log"

// ExampleGetTOML gets the diamnet.toml file for coins.asia
func ExampleClient_GetDiamNetToml() {
	_, err := DefaultClient.GetDiamNetToml("coins.asia")
	if err != nil {
		log.Fatal(err)
	}
}
