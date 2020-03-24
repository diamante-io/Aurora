package compliance

// FetchInfoRequest represents a request sent to fetch_info callback
type FetchInfoRequest struct {
	Address string `form:"address"`
}

// FetchInfoResponse represents a response returned by fetch_info callback
type FetchInfoResponse struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	DateOfBirth string `json:"date_of_birth"`
}
