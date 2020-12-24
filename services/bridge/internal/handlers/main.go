package handlers

import (
	"github.com/diamnet/go/clients/federation"
	hc "github.com/diamnet/go/clients/auroraclient"
	"github.com/diamnet/go/clients/diamnettoml"
	"github.com/diamnet/go/services/bridge/internal/config"
	"github.com/diamnet/go/services/bridge/internal/db"
	"github.com/diamnet/go/services/bridge/internal/listener"
	"github.com/diamnet/go/services/bridge/internal/submitter"
	"github.com/diamnet/go/support/http"
)

// RequestHandler implements bridge server request handlers
type RequestHandler struct {
	Config               *config.Config                          `inject:""`
	Client               http.SimpleHTTPClientInterface          `inject:""`
	Aurora              hc.ClientInterface                      `inject:""`
	Database             db.Database                             `inject:""`
	DiamNetTomlResolver  diamnettoml.ClientInterface             `inject:""`
	FederationResolver   federation.ClientInterface              `inject:""`
	TransactionSubmitter submitter.TransactionSubmitterInterface `inject:""`
	PaymentListener      *listener.PaymentListener               `inject:""`
}
