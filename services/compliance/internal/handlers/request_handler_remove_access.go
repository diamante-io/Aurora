package handlers

import (
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/diamnet/go/services/internal/bridge-compliance-shared/http/helpers"
)

// HandlerRemoveAccess implements /remove_access endpoint
func (rh *RequestHandler) HandlerRemoveAccess(w http.ResponseWriter, r *http.Request) {
	domain := r.PostFormValue("domain")
	userID := r.PostFormValue("user_id")

	// TODO check params

	if userID != "" {
		err := rh.Database.DeleteAllowedUserByDomainAndUserID(domain, userID)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Warn("Error removing allowed user")
			helpers.Write(w, helpers.InternalServerError)
			return
		}
	} else {
		err := rh.Database.DeleteAllowedFIByDomain(domain)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Warn("Error removing allowed FI")
			helpers.Write(w, helpers.InternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
