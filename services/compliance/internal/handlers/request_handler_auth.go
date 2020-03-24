package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	baseAmount "github.com/hcnet/go/amount"
	"github.com/hcnet/go/protocols/compliance"
	"github.com/hcnet/go/services/compliance/internal/db"
	shared "github.com/hcnet/go/services/internal/bridge-compliance-shared"
	httpHelpers "github.com/hcnet/go/services/internal/bridge-compliance-shared/http/helpers"
	callback "github.com/hcnet/go/services/internal/bridge-compliance-shared/protocols/compliance"
	"github.com/hcnet/go/xdr"
)

// HandlerAuth implements authorize endpoint
func (rh *RequestHandler) HandlerAuth(w http.ResponseWriter, r *http.Request) {
	authreq := &compliance.AuthRequest{
		DataJSON:  r.PostFormValue("data"),
		Signature: r.PostFormValue("sig"),
	}

	log.WithFields(log.Fields{"data": authreq.DataJSON, "sig": authreq.Signature}).Info("HandlerAuth")

	err := authreq.Validate()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Info(err.Error())
		httpHelpers.Write(w, httpHelpers.NewInvalidParameterError("", err.Error()))
		return
	}

	authData, err := authreq.Data()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error(err.Error())
		httpHelpers.Write(w, httpHelpers.InternalServerError)
		return
	}

	senderHcNetToml, err := rh.HcNetTomlResolver.GetHcNetTomlByAddress(authData.Sender)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "sender": authData.Sender}).Warn("Cannot get hcnet.toml of sender")
		errorResponse := httpHelpers.NewInvalidParameterError("data.sender", "Cannot get hcnet.toml of sender")
		httpHelpers.Write(w, errorResponse)
		return
	}

	if !shared.IsValidAccountID(senderHcNetToml.SigningKey) {
		errorResponse := httpHelpers.NewInvalidParameterError("data.sender", "SIGNING_KEY in hcnet.toml of sender is invalid")
		// TODO
		// log.WithFields(errorResponse.LogData).Warn("SIGNING_KEY in hcnet.toml of sender is invalid")
		httpHelpers.Write(w, errorResponse)
		return
	}

	// Verify signature
	signatureBytes, err := base64.StdEncoding.DecodeString(authreq.Signature)
	if err != nil {
		errorResponse := httpHelpers.NewInvalidParameterError("sig", "Invalid base64 string.")
		// TODO
		// log.WithFields(errorResponse.LogData).Warn("Error decoding signature")
		httpHelpers.Write(w, errorResponse)
		return
	}
	err = rh.SignatureSignerVerifier.Verify(senderHcNetToml.SigningKey, []byte(authreq.DataJSON), signatureBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"signing_key": senderHcNetToml.SigningKey,
			"data":        authreq.Data,
			"sig":         authreq.Signature,
		}).Warn("Invalid signature")
		errorResponse := httpHelpers.NewInvalidParameterError("sig", "Invalid signature.")
		httpHelpers.Write(w, errorResponse)
		return
	}

	b64r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(authData.Tx))
	var tx xdr.Transaction
	_, err = xdr.Unmarshal(b64r, &tx)
	if err != nil {
		errorResponse := httpHelpers.NewInvalidParameterError("data.tx", "Error decoding Transaction XDR")
		log.WithFields(log.Fields{
			"err": err,
			"tx":  authData.Tx,
		}).Warn("Error decoding Transaction XDR")
		httpHelpers.Write(w, errorResponse)
		return
	}

	if tx.Memo.Hash == nil {
		errorResponse := httpHelpers.NewInvalidParameterError("data.tx", "Transaction does not contain Memo.Hash")
		log.WithFields(log.Fields{"tx": authData.Tx}).Warn("Transaction does not contain Memo.Hash")
		httpHelpers.Write(w, errorResponse)
		return
	}

	// Validate memo preimage hash
	memoPreimageHashBytes := sha256.Sum256([]byte(authData.AttachmentJSON))
	memoBytes := [32]byte(*tx.Memo.Hash)

	if memoPreimageHashBytes != memoBytes {
		h := xdr.Hash(memoPreimageHashBytes)
		tx.Memo.Hash = &h

		var txBytes bytes.Buffer
		_, err = xdr.Marshal(&txBytes, tx)
		if err != nil {
			log.Error("Error mashaling transaction")
			errorResponse := httpHelpers.NewInvalidParameterError("data.tx", "Error marshaling transaction")
			httpHelpers.Write(w, errorResponse)
			return
		}

		expectedTx := base64.StdEncoding.EncodeToString(txBytes.Bytes())

		log.WithFields(log.Fields{"tx": authData.Tx, "expected_tx": expectedTx}).Warn("Memo preimage hash does not equal tx Memo.Hash")
		errorResponse := httpHelpers.NewInvalidParameterError("data.tx", "Memo preimage hash does not equal tx Memo.Hash")
		httpHelpers.Write(w, errorResponse)
		return
	}

	attachment, err := authData.Attachment()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error getting attachment")
		httpHelpers.Write(w, httpHelpers.InternalServerError)
		return
	}

	transactionHash, err := shared.TransactionHash(&tx, rh.Config.NetworkPassphrase)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("Error calculating tx hash")
		httpHelpers.Write(w, httpHelpers.InternalServerError)
		return
	}

	response := compliance.AuthResponse{}

	// Sanctions check
	if rh.Config.Callbacks.Sanctions == "" {
		response.TxStatus = compliance.AuthStatusOk
	} else {
		var senderInfo []byte
		senderInfo, err = json.Marshal(attachment.Transaction.SenderInfo)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		var resp *http.Response
		resp, err = rh.Client.PostForm(
			rh.Config.Callbacks.Sanctions,
			url.Values{"sender": {string(senderInfo)}},
		)
		if err != nil {
			log.WithFields(log.Fields{
				"sanctions": rh.Config.Callbacks.Sanctions,
				"err":       err,
			}).Error("Error sending request to sanctions server")
			httpHelpers.Write(w, httpHelpers.InternalServerError)
			return
		}

		defer resp.Body.Close()
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error reading sanctions server response")
			httpHelpers.Write(w, httpHelpers.InternalServerError)
			return
		}

		switch resp.StatusCode {
		case http.StatusOK: // AuthStatusOk
			response.TxStatus = compliance.AuthStatusOk
		case http.StatusAccepted: // AuthStatusPending
			response.TxStatus = compliance.AuthStatusPending

			callbackResponse := callback.CallbackResponse{}
			err = json.Unmarshal(body, &callbackResponse)
			if err != nil {
				// Set default value
				response.Pending = 600
			} else {
				response.Pending = callbackResponse.Pending
			}
		case http.StatusBadRequest: // AuthStatusError
			response.TxStatus = compliance.AuthStatusError

			callbackResponse := callback.CallbackResponse{}
			err = json.Unmarshal(body, &callbackResponse)
			if err != nil {
				log.WithFields(log.Fields{
					"status": resp.StatusCode,
					"body":   string(body),
				}).Error("Error response from sanctions server")
			} else {
				response.Error = callbackResponse.Error
			}
		case http.StatusForbidden: // AuthStatusDenied
			response.TxStatus = compliance.AuthStatusDenied
		default:
			log.WithFields(log.Fields{
				"status": resp.StatusCode,
				"body":   string(body),
			}).Error("Error response from sanctions server")
			httpHelpers.Write(w, httpHelpers.InternalServerError)
			return
		}
	}

	// User info
	if authData.NeedInfo {
		if rh.Config.Callbacks.AskUser == "" {
			response.InfoStatus = compliance.AuthStatusDenied

			// Check AllowedFi
			tokens := strings.Split(authData.Sender, "*")
			if len(tokens) != 2 {
				log.WithFields(log.Fields{
					"sender": authData.Sender,
				}).Warn("Invalid hcnet address")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}

			allowedFi, err2 := rh.Database.GetAllowedFIByDomain(tokens[1])
			if err2 != nil {
				log.WithFields(log.Fields{"err": err2}).Error("Error getting AllowedFi from DB")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}

			if allowedFi == nil {
				// FI not found check AllowedUser
				allowedUser, err2 := rh.Database.GetAllowedUserByDomainAndUserID(tokens[1], tokens[0])
				if err2 != nil {
					log.WithFields(log.Fields{"err": err2}).Error("Error getting AllowedUser from DB")
					httpHelpers.Write(w, httpHelpers.InternalServerError)
					return
				}

				if allowedUser != nil {
					response.InfoStatus = compliance.AuthStatusOk
				}
			} else {
				response.InfoStatus = compliance.AuthStatusOk
			}
		} else {
			// Ask user
			var amount, assetType, assetCode, assetIssuer string

			if len(tx.Operations) > 0 {
				operationBody := tx.Operations[0].Body
				if operationBody.Type == xdr.OperationTypePayment {
					amount = baseAmount.String(operationBody.PaymentOp.Amount)
					operationBody.PaymentOp.Asset.Extract(&assetType, &assetCode, &assetIssuer)
				} else if operationBody.Type == xdr.OperationTypePathPayment {
					amount = baseAmount.String(operationBody.PathPaymentOp.DestAmount)
					operationBody.PathPaymentOp.DestAsset.Extract(&assetType, &assetCode, &assetIssuer)
				}
			}

			var senderInfo []byte
			senderInfo, err = json.Marshal(attachment.Transaction.SenderInfo)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}

			var resp *http.Response
			resp, err = rh.Client.PostForm(
				rh.Config.Callbacks.AskUser,
				url.Values{
					"amount":       {amount},
					"asset_code":   {assetCode},
					"asset_issuer": {assetIssuer},
					"sender":       {string(senderInfo)},
					"note":         {attachment.Transaction.Note},
				},
			)
			if err != nil {
				log.WithFields(log.Fields{
					"ask_user": rh.Config.Callbacks.AskUser,
					"err":      err,
				}).Error("Error sending request to ask_user server")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}

			defer resp.Body.Close()
			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error("Error reading ask_user server response")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}

			switch resp.StatusCode {
			case http.StatusOK: // AuthStatusOk
				response.InfoStatus = compliance.AuthStatusOk
			case http.StatusAccepted: // AuthStatusPending
				response.InfoStatus = compliance.AuthStatusPending

				callbackResponse := callback.CallbackResponse{}
				err = json.Unmarshal(body, &callbackResponse)
				if err != nil {
					// Set default value
					response.Pending = 600
				} else {
					response.Pending = callbackResponse.Pending
				}
			case http.StatusBadRequest: // AuthStatusError
				response.InfoStatus = compliance.AuthStatusError

				callbackResponse := callback.CallbackResponse{}
				err = json.Unmarshal(body, &callbackResponse)
				if err != nil {
					log.WithFields(log.Fields{
						"status": resp.StatusCode,
						"body":   string(body),
					}).Error("Error response from sanctions server")
				} else {
					response.Error = callbackResponse.Error
				}
			case http.StatusForbidden: // AuthStatusDenied
				response.InfoStatus = compliance.AuthStatusDenied
			default:
				log.WithFields(log.Fields{
					"status": resp.StatusCode,
					"body":   string(body),
				}).Error("Error response from ask_user server")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}
		}

		if response.InfoStatus == compliance.AuthStatusOk {
			// Fetch Info
			fetchInfoRequest := &callback.FetchInfoRequest{Address: string(attachment.Transaction.Route)}
			var resp *http.Response
			resp, err = rh.Client.PostForm(
				rh.Config.Callbacks.FetchInfo,
				httpHelpers.ToValues(fetchInfoRequest),
			)
			if err != nil {
				log.WithFields(log.Fields{
					"fetch_info": rh.Config.Callbacks.FetchInfo,
					"err":        err,
				}).Error("Error sending request to fetch_info server")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}

			defer resp.Body.Close()
			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.WithFields(log.Fields{
					"fetch_info": rh.Config.Callbacks.FetchInfo,
					"err":        err,
				}).Error("Error reading fetch_info server response")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}

			if resp.StatusCode != http.StatusOK {
				log.WithFields(log.Fields{
					"fetch_info": rh.Config.Callbacks.FetchInfo,
					"status":     resp.StatusCode,
					"body":       string(body),
				}).Error("Error response from fetch_info server")
				httpHelpers.Write(w, httpHelpers.InternalServerError)
				return
			}

			response.DestInfo = string(body)
		}
	} else {
		response.InfoStatus = compliance.AuthStatusOk
	}

	if response.TxStatus == compliance.AuthStatusOk && response.InfoStatus == compliance.AuthStatusOk {
		w.WriteHeader(http.StatusOK)
		authorizedTransaction := &db.AuthorizedTransaction{
			TransactionID:  hex.EncodeToString(transactionHash[:]),
			Memo:           base64.StdEncoding.EncodeToString(memoBytes[:]),
			TransactionXdr: authData.Tx,
			AuthorizedAt:   time.Now(),
			Data:           authreq.DataJSON,
		}
		err = rh.Database.InsertAuthorizedTransaction(authorizedTransaction)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Warn("Error persisting AuthorizedTransaction")
			httpHelpers.Write(w, httpHelpers.InternalServerError)
			return
		}
	} else if response.TxStatus == compliance.AuthStatusDenied || response.InfoStatus == compliance.AuthStatusDenied {
		w.WriteHeader(http.StatusForbidden)
	} else if response.TxStatus == compliance.AuthStatusError || response.InfoStatus == compliance.AuthStatusError {
		w.WriteHeader(http.StatusBadRequest)
	} else if response.TxStatus == compliance.AuthStatusPending || response.InfoStatus == compliance.AuthStatusPending {
		w.WriteHeader(http.StatusAccepted)
	}

	responseBody, err := response.Marshal()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(responseBody)
}
