package compliance

// AuthStatus represents auth status returned by Auth Server
type AuthStatus string

const (
	// AuthStatusOk is returned when authentication was successful
	AuthStatusOk AuthStatus = "ok"
	// AuthStatusPending is returned when authentication is pending
	AuthStatusPending AuthStatus = "pending"
	// AuthStatusDenied is returned when authentication was denied
	AuthStatusDenied AuthStatus = "denied"
	// AuthStatusError is returned when there was an error
	AuthStatusError AuthStatus = "error"
)

// AuthRequest represents auth request sent to compliance server
type AuthRequest struct {
	// Marshalled AuthData JSON object (because of the attached signature, json can be marshalled to infinite number of valid JSON strings)
	DataJSON string `name:"data" valid:"required,json"`
	// Signature of sending FI
	Signature string `name:"sig" valid:"required,base64"`
}

// AuthData represents how AuthRequest.Data field looks like.
type AuthData struct {
	// The hcnet address of the customer that is initiating the send.
	Sender string `json:"sender" valid:"required,hcnet_address"`
	// If the caller needs the recipient's AML info in order to send the payment.
	NeedInfo bool `json:"need_info" valid:"-"`
	// The transaction that the sender would like to send in XDR format. This transaction is unsigned.
	Tx string `json:"tx" valid:"required,base64"`
	// The full text of the attachment the hash of this attachment is included in the transaction.
	AttachmentJSON string `json:"attachment" valid:"required,json"`
}

// AuthResponse represents response sent by auth server
type AuthResponse struct {
	// If this FI is willing to share AML information or not. {ok, denied, pending, error}
	InfoStatus AuthStatus `json:"info_status"`
	// If this FI is willing to accept this transaction. {ok, denied, pending, error}
	TxStatus AuthStatus `json:"tx_status"`
	// (only present if info_status is ok) JSON of the recipient's AML information. in the HcNet attachment convention
	DestInfo string `json:"dest_info,omitempty"`
	// (only present if info_status or tx_status is pending) Estimated number of seconds till the sender can check back for a change in status. The sender should just resubmit this request after the given number of seconds.
	Pending int `json:"pending,omitempty"`
	// (only present if info_status or tx_status is error)
	Error string `json:"error,omitempty"`
}

// Attachment represents preimage object of compliance protocol in
// HcNet attachment convention
type Attachment struct {
	Nonce       string `json:"nonce"`
	Transaction `json:"transaction"`
	Operations  []Operation `json:"operations"`
}

// Transaction represents transaction field in HcNet attachment
type Transaction struct {
	SenderInfo map[string]string `json:"sender_info"`
	Route      Route             `json:"route"`
	Note       string            `json:"note"`
	Extra      string            `json:"extra"`
}

// Operation represents a single operation object in HcNet attachment
type Operation struct {
	// Overriddes Transaction field for this operation
	SenderInfo map[string]string `json:"sender_info"`
	// Overriddes Transaction field for this operation
	Route Route `json:"route"`
	// Overriddes Transaction field for this operation
	Note string `json:"note"`
}

// Route allows unmarshalling both integer and string types into string
type Route string

// SenderInfo is a helper structure with standardized fields that contains
// information about the sender. Use Map() method to transform it to
// map[string]string used in Transaction and Operation structs.
type SenderInfo struct {
	FirstName   string `json:"first_name,omitempty"`
	MiddleName  string `json:"middle_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Province    string `json:"province,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Country     string `json:"country,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
}

// TransactionStatus is the status string returned be tx_status endpoint
type TransactionStatus string

const (
	// TransactionStatusUnknown is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// institution is not aware of the transaction
	TransactionStatusUnknown TransactionStatus = "unknown"

	// TransactionStatusApproved is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// payment was approved by the receiving FI but the HcNet
	// transaction hasn't been received yet
	TransactionStatusApproved TransactionStatus = "approved"

	// TransactionStatusNotApproved is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// HcNet transaction was found but it was never approved
	// by the receiving FI.
	TransactionStatusNotApproved TransactionStatus = "not_approved"

	// TransactionStatusPending is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// payment was received and being processed
	TransactionStatusPending TransactionStatus = "pending"

	// TransactionStatusFailed is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// payment was failed and could not be deposited
	TransactionStatusFailed TransactionStatus = "failed"

	// TransactionStatusRefunded is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// payment was sent back to sending FI
	TransactionStatusRefunded TransactionStatus = "refunded"

	// TransactionStatusClaimable is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// cash is ready to be picked up at specified locations.
	// Mostly used for cash pickup
	TransactionStatusClaimable TransactionStatus = "claimable"

	// TransactionStatusDelivered is a value of `status` field for the
	// tx_status endpoint response. It represents that the
	// payment has been delivered to the recepient
	TransactionStatusDelivered TransactionStatus = "delivered"
)

// TransactionStatusResponse represents a response from the tx_status endpoint
type TransactionStatusResponse struct {
	Status   TransactionStatus `json:"status"`
	RecvCode string            `json:"recv_code,omitempty"`
	RefundTx string            `json:"refund_tx,omitempty"`
	Msg      string            `json:"msg,omitempty"`
}
