package keystore

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/lib/pq"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/render/problem"
)

type encryptedKeys struct {
	KeysBlob      string     `json:"keysBlob"`
	Salt          string     `json:"salt"`
	EncrypterName string     `json:"encrypterName"`
	CreatedAt     time.Time  `json:"createdAt"`
	ModifiedAt    *time.Time `json:"modifiedAt,omitempty"`
}

type putKeysRequest struct {
	KeysBlob      string `json:"keysBlob"`
	Salt          string `json:"salt"`
	EncrypterName string `json:"encrypterName"`
}

func (s *Service) putKeys(ctx context.Context, in putKeysRequest) (*encryptedKeys, error) {
	userID := userID(ctx)
	if userID == "" {
		return nil, probNotAuthorized
	}

	if in.Salt == "" {
		return nil, problem.MakeInvalidFieldProblem("salt", errRequiredField)
	}
	if in.EncrypterName == "" {
		return nil, problem.MakeInvalidFieldProblem("encrypterName", errRequiredField)
	}
	if in.KeysBlob == "" {
		return nil, problem.MakeInvalidFieldProblem("keysBlob", errRequiredField)
	}

	keysData, err := base64.RawURLEncoding.DecodeString(in.KeysBlob)
	if err != nil {
		// TODO: we need to implement a helper function in the
		// support/error package for keeping the stack trace from err
		// and substitude the root error for the one we want for better
		// debugging experience.
		// Thowing away the original err makes it harder for debugging.
		return nil, probInvalidKeysBlob
	}

	q := `
		INSERT INTO encrypted_keys (user_id, encrypted_keys_data, salt, encrypter_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET encrypted_keys_data = excluded.encrypted_keys_data, salt = excluded.salt, encrypter_name = excluded.encrypter_name,  modified_at = NOW()
		RETURNING encrypted_keys_data, salt, encrypter_name, created_at, modified_at
	`
	var (
		keysBlob   []byte
		out        encryptedKeys
		modifiedAt pq.NullTime
	)
	err = s.db.QueryRowContext(ctx, q, userID, keysData, in.Salt, in.EncrypterName).Scan(&keysBlob, &out.Salt, &out.EncrypterName, &out.CreatedAt, &modifiedAt)
	if err != nil {
		return nil, errors.Wrap(err, "storing keys blob")
	}

	out.KeysBlob = base64.RawURLEncoding.EncodeToString(keysBlob)
	if modifiedAt.Valid {
		out.ModifiedAt = &modifiedAt.Time
	}
	return &out, nil
}

func (s *Service) getKeys(ctx context.Context) (*encryptedKeys, error) {
	userID := userID(ctx)
	if userID == "" {
		return nil, probNotAuthorized
	}

	q := `
		SELECT encrypted_keys_data, salt, encrypter_name, created_at, modified_at
		FROM encrypted_keys
		WHERE user_id = $1
	`
	var (
		keysBlob   []byte
		out        encryptedKeys
		modifiedAt pq.NullTime
	)
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&keysBlob, &out.Salt, &out.EncrypterName, &out.CreatedAt, &modifiedAt)
	if err != nil {
		return nil, errors.Wrap(err, "getting keys blob")
	}

	out.KeysBlob = base64.RawURLEncoding.EncodeToString(keysBlob)
	if modifiedAt.Valid {
		out.ModifiedAt = &modifiedAt.Time
	}
	return &out, nil
}

func (s *Service) deleteKeys(ctx context.Context) error {
	userID := userID(ctx)
	if userID == "" {
		return probNotAuthorized
	}

	q := `
		DELETE FROM encrypted_keys
		WHERE user_id = $1
	`
	_, err := s.db.ExecContext(ctx, q, userID)
	return errors.Wrap(err, "deleting keys blob")
}
