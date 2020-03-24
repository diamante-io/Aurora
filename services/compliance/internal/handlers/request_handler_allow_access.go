package handlers

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/hcnet/go/services/compliance/internal/db"
	"github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
)

// HandlerAllowAccess implements /allow_access endpoint
func (rh *RequestHandler) HandlerAllowAccess(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	domain := r.PostFormValue("domain")
	publicKey := r.PostFormValue("public_key")
	userID := r.PostFormValue("user_id")

	// TODO check params

	var err error

	if userID != "" {
		entity := &db.AllowedUser{
			FiName:      name,
			FiDomain:    domain,
			FiPublicKey: publicKey,
			UserID:      userID,
			AllowedAt:   time.Now(),
		}
		err = rh.Database.InsertAllowedUser(entity)
	} else {
		entity := &db.AllowedFI{
			Name:      name,
			Domain:    domain,
			PublicKey: publicKey,
			AllowedAt: time.Now(),
		}
		err = rh.Database.InsertAllowedFI(entity)
	}

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("Error persisting /allow entity")
		helpers.Write(w, helpers.InternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
