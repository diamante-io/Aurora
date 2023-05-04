package kycstatus

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/diamnet/go/services/regulated-assets-approval-server/internal/serve/httperror"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/http/httpdecode"
	"github.com/diamnet/go/support/log"
	"github.com/diamnet/go/support/render/httpjson"
)

type kycGetResponse struct {
	DiamnetAddress string     `json:"diamnet_address"`
	CallbackID     string     `json:"callback_id"`
	EmailAddress   string     `json:"email_address,omitempty"`
	CreatedAt      *time.Time `json:"created_at"`
	KYCSubmittedAt *time.Time `json:"kyc_submitted_at,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
	PendingAt      *time.Time `json:"pending_at,omitempty"`
}

func (k *kycGetResponse) Render(w http.ResponseWriter) {
	httpjson.Render(w, k, httpjson.JSON)
}

type GetDetailHandler struct {
	DB *sqlx.DB
}

func (h GetDetailHandler) validate() error {
	if h.DB == nil {
		return errors.New("database cannot be nil")
	}
	return nil
}

type getDetailRequest struct {
	DiamnetAddressOrCallbackID string `path:"diamnet_address_or_callback_id"`
}

func (h GetDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.validate()
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "validating kyc-status GetDetailHandler"))
		httperror.InternalServer.Render(w)
		return
	}

	in := getDetailRequest{}
	err = httpdecode.Decode(r, &in)
	if err != nil {
		log.Ctx(ctx).Error(errors.Wrap(err, "decoding kyc-status GET Request"))
		httperror.BadRequest.Render(w)
		return
	}

	kycGetResponse, err := h.handle(ctx, in)
	if err != nil {
		httpErr, ok := err.(*httperror.Error)
		if !ok {
			httpErr = httperror.InternalServer
		}
		httpErr.Render(w)
		return
	}

	kycGetResponse.Render(w)
}

func (h GetDetailHandler) handle(ctx context.Context, in getDetailRequest) (*kycGetResponse, error) {
	// Check if getDetailRequest DiamnetAddressOrCallbackID value is present.
	if in.DiamnetAddressOrCallbackID == "" {
		return nil, httperror.NewHTTPError(http.StatusBadRequest, "Missing diamnet address or callbackID.")
	}

	// Prepare SELECT query return values.
	var (
		diamnetAddress, callbackID                        string
		emailAddress                                      sql.NullString
		createdAt                                         time.Time
		kycSubmittedAt, approvedAt, rejectedAt, pendingAt sql.NullTime
	)
	const q = `
		SELECT diamnet_address, email_address, created_at, kyc_submitted_at, approved_at, rejected_at, pending_at, callback_id
		FROM accounts_kyc_status
		WHERE diamnet_address = $1 OR callback_id = $1
	`
	err := h.DB.QueryRowContext(ctx, q, in.DiamnetAddressOrCallbackID).Scan(&diamnetAddress, &emailAddress, &createdAt, &kycSubmittedAt, &approvedAt, &rejectedAt, &pendingAt, &callbackID)
	if err == sql.ErrNoRows {
		return nil, httperror.NewHTTPError(http.StatusNotFound, "Not found.")
	}
	if err != nil {
		return nil, errors.Wrap(err, "querying the database")
	}

	return &kycGetResponse{
		DiamnetAddress: diamnetAddress,
		CallbackID:     callbackID,
		EmailAddress:   emailAddress.String,
		CreatedAt:      &createdAt,
		KYCSubmittedAt: timePointerIfValid(kycSubmittedAt),
		ApprovedAt:     timePointerIfValid(approvedAt),
		RejectedAt:     timePointerIfValid(rejectedAt),
		PendingAt:      timePointerIfValid(pendingAt),
	}, nil
}

// timePointerIfValid returns a pointer to the date from the provided
// `sql.NullTime` if it's valid or `nil` if it's not.
func timePointerIfValid(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
