package handlers

import (
	"strconv"
	"time"

	"github.com/hcnet/go/clients/federation"
	"github.com/hcnet/go/clients/hcnettoml"
	"github.com/hcnet/go/services/compliance/internal/config"
	"github.com/hcnet/go/services/compliance/internal/crypto"
	"github.com/hcnet/go/services/compliance/internal/db"
	"github.com/hcnet/go/support/http"
)

// RequestHandler implements compliance server request handlers
type RequestHandler struct {
	Config                  *config.Config                 `inject:""`
	Client                  http.SimpleHTTPClientInterface `inject:""`
	Database                db.Database                    `inject:""`
	SignatureSignerVerifier crypto.SignerVerifierInterface `inject:""`
	HcNetTomlResolver     hcnettoml.ClientInterface    `inject:""`
	FederationResolver      federation.ClientInterface     `inject:""`
	NonceGenerator          NonceGeneratorInterface        `inject:""`
}

type NonceGeneratorInterface interface {
	Generate() string
}

type NonceGenerator struct{}

func (n *NonceGenerator) Generate() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

type TestNonceGenerator struct{}

func (n *TestNonceGenerator) Generate() string {
	return "nonce"
}
