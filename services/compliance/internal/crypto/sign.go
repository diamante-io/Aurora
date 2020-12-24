package crypto

import (
	"encoding/base64"

	"github.com/diamnet/go/keypair"
)

// SignerVerifierInterface is the interface that helps mocking SignerVerifier
type SignerVerifierInterface interface {
	Sign(secretSeed string, message []byte) (string, error)
	Verify(publicKey string, message, signature []byte) error
}

// SignerVerifier implements methods to Sign and Verify signatures
type SignerVerifier struct{}

// Sign signs message using secretSeed. Returns base64-encoded signature.
func (s *SignerVerifier) Sign(secretSeed string, message []byte) (string, error) {
	kp, err := keypair.Parse(secretSeed)
	if err != nil {
		return "", err
	}

	signature, err := kp.Sign(message)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify verifies if signature is a valid signature of message signed by publicKey.
func (s *SignerVerifier) Verify(publicKey string, message, signature []byte) error {
	kp, err := keypair.Parse(publicKey)
	if err != nil {
		return err
	}

	err = kp.Verify(message, signature)
	if err != nil {
		return err
	}

	return nil
}
