package hcnettoml

import "log"

// ExampleGetTOML gets the hcnet.toml file for coins.asia
func ExampleClient_GetHcNetToml() {
	_, err := DefaultClient.GetHcNetToml("coins.asia")
	if err != nil {
		log.Fatal(err)
	}
}
