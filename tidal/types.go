package tidal

// loginResponse is the JSON object returned from a successful login request.
type loginResponse struct {
	SessionID   string `json:"sessionId"`
	CountryCode string `json:"countryCode"`
	UserID      int    `json:"userId"`
}
