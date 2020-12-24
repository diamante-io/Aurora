package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/diamnet/go/keypair"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
)

// KeyPair struct contains key pair public and private key
type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// CreateKeypair implements /create-keypair endpoint
func (rh *RequestHandler) CreateKeypair(w http.ResponseWriter, r *http.Request) {
	kp, err := keypair.Random()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error generating random keypair")
		helpers.Write(w, helpers.InternalServerError)
	}

	response, err := json.Marshal(KeyPair{kp.Address(), kp.Seed()})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error marshalling random keypair")
		helpers.Write(w, helpers.InternalServerError)
	}

	w.Write(response)
}
