package handlers

import (
	"github.com/hcnet/go/clients/federation"
	hc "github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/clients/hcnettoml"
	"github.com/hcnet/go/services/bridge/internal/config"
	"github.com/hcnet/go/services/bridge/internal/db"
	"github.com/hcnet/go/services/bridge/internal/listener"
	"github.com/hcnet/go/services/bridge/internal/submitter"
	"github.com/hcnet/go/support/http"
)

// RequestHandler implements bridge server request handlers
type RequestHandler struct {
	Config               *config.Config                          `inject:""`
	Client               http.SimpleHTTPClientInterface          `inject:""`
	Aurora              hc.ClientInterface                      `inject:""`
	Database             db.Database                             `inject:""`
	HcNetTomlResolver  hcnettoml.ClientInterface             `inject:""`
	FederationResolver   federation.ClientInterface              `inject:""`
	TransactionSubmitter submitter.TransactionSubmitterInterface `inject:""`
	PaymentListener      *listener.PaymentListener               `inject:""`
}
