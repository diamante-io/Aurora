package keystore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/render/httpjson"
)

func TestPutKeysAPI(t *testing.T) {
	db := openKeystoreDB(t)
	defer db.Close() // drop test db

	conn := db.Open()
	defer conn.Close() // close db connection

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"userID":"test-user"}`)
	}))
	defer ts.Close()

	h := ServeMux(&Service{
		db: conn.DB,
		authenticator: &Authenticator{
			URL:     ts.URL,
			APIType: REST,
		},
	})

	blob := `[{
		"keyType": "plaintextKey",
		"publicKey": "hcnet-pubkey",
		"privateKey": "encrypted-hcnet-privatekey"
	}]`
	encodedBlob := base64.RawURLEncoding.EncodeToString([]byte(blob))
	encrypterName := "identity"
	salt := "random-salt"
	body, err := json.Marshal(putKeysRequest{
		KeysBlob:      encodedBlob,
		EncrypterName: encrypterName,
		Salt:          salt,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("PUT", "/keys", bytes.NewReader([]byte(body)))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("PUT %s responded with %s, want %s", req.URL, http.StatusText(rr.Code), http.StatusText(http.StatusOK))
	}
	got := &encryptedKeys{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got == nil {
		t.Error("Expected to receive an encryptedKeys response but did not")
	}

	if got.KeysBlob != encodedBlob {
		t.Errorf("got blob: %s, want: %s\n", got.KeysBlob, encodedBlob)
	}

	if got.EncrypterName != encrypterName {
		t.Errorf("got encrypter name: %s, want: %s\n", got.EncrypterName, encrypterName)
	}

	if got.Salt != salt {
		t.Errorf("got salt: %s, want: %s\n", got.Salt, salt)
	}

	if got.CreatedAt.Before(time.Now().Add(-time.Hour)) {
		t.Errorf("got CreatedAt=%s, want CreatedAt within the last hour", got.CreatedAt)
	}
}

func TestGetKeysAPI(t *testing.T) {
	db := openKeystoreDB(t)
	defer db.Close() // drop test db

	conn := db.Open()
	defer conn.Close() // close db connection

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"userID":"test-user"}`)
	}))
	defer ts.Close()

	ctx := withUserID(context.Background(), "test-user")
	s := &Service{
		db: conn.DB,
		authenticator: &Authenticator{
			URL:     ts.URL,
			APIType: REST,
		},
	}
	h := ServeMux(s)

	blob := `[{
		"keyType": "plaintextKey",
		"publicKey": "hcnet-pubkey",
		"privateKey": "encrypted-hcnet-privatekey"
	}]`
	encodedBlob := base64.RawURLEncoding.EncodeToString([]byte(blob))
	encrypterName := "identity"
	salt := "random-salt"

	_, err := s.putKeys(ctx, putKeysRequest{
		KeysBlob:      encodedBlob,
		EncrypterName: encrypterName,
		Salt:          salt,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/keys", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET %s responded with %s, want %s", req.URL, http.StatusText(rr.Code), http.StatusText(http.StatusOK))
	}
	got := &encryptedKeys{}
	json.Unmarshal(rr.Body.Bytes(), &got)
	if got == nil {
		t.Error("Expected to receive an encryptedKeys response but did not")
	}

	if got.KeysBlob != encodedBlob {
		t.Errorf("got blob: %s, want: %s\n", got.KeysBlob, encodedBlob)
	}

	if got.EncrypterName != encrypterName {
		t.Errorf("got encrypter name: %s, want: %s\n", got.EncrypterName, encrypterName)
	}

	if got.Salt != salt {
		t.Errorf("got salt: %s, want: %s\n", got.Salt, salt)
	}

	if got.CreatedAt.Before(time.Now().Add(-time.Hour)) {
		t.Errorf("got CreatedAt=%s, want CreatedAt within the last hour", got.CreatedAt)
	}

	err = s.deleteKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("GET %s responded with %s, want %s", req.URL, http.StatusText(rr.Code), http.StatusText(http.StatusNotFound))
	}
}

func TestDeleteKeysAPI(t *testing.T) {
	db := openKeystoreDB(t)
	defer db.Close() // drop test db

	conn := db.Open()
	defer conn.Close() // close db connection

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"userID":"test-user"}`)
	}))
	defer ts.Close()

	s := &Service{
		db: conn.DB,
		authenticator: &Authenticator{
			URL:     ts.URL,
			APIType: REST,
		},
	}
	h := ServeMux(s)

	blob := `[{
		"keyType": "plaintextKey",
		"publicKey": "hcnet-pubkey",
		"privateKey": "encrypted-hcnet-privatekey"
	}]`
	encodedBlob := base64.RawURLEncoding.EncodeToString([]byte(blob))
	encrypterName := "identity"
	salt := "random-salt"

	ctx := withUserID(context.Background(), "test-user")
	_, err := s.putKeys(ctx, putKeysRequest{
		KeysBlob:      encodedBlob,
		EncrypterName: encrypterName,
		Salt:          salt,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("DELETE", "/keys", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("DELETE %s responded with %s, want %s", req.URL, http.StatusText(rr.Code), http.StatusText(http.StatusOK))
	}

	got := rr.Body.Bytes()
	dr, _ := json.MarshalIndent(httpjson.DefaultResponse, "", "  ")
	if !bytes.Equal(got, dr) {
		t.Errorf("got: %s, expected: %s", got, dr)
	}

	_, err = s.getKeys(ctx)
	if errors.Cause(err) != sql.ErrNoRows {
		t.Errorf("expect the keys blob of the user %s to be deleted", userID(ctx))
	}
}
