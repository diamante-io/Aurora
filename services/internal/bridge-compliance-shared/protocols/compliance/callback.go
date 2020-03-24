package compliance

// CallbackResponse is a response from Sanctions and AskUser callbacks when they return 202 Accepted or 400 Bad Requests statuses
type CallbackResponse struct {
	// Estimated number of seconds utill the sender can check back for a change in status.
	Pending int    `json:"pending"`
	Error   string `json:"error"`
}
