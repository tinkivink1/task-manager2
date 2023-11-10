package apiserver

// StatusResponse represents a successful API response with a message.
// @Summary Successful API response with a message
// @Description Successful API response with a message.
// @Produce json
// @Success 200 {object} StatusResponse "Successful response with a message"
type StatusResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an API response with an error message.
// @Summary API response with an error message
// @Description API response with an error message.
// @Produce json
// @Param error body ErrorResponse true "Error message"
type ErrorResponse struct {
	Error string `json:"error"`
}

// TokenResponse represents an API response with an access token.
// @Summary API response with an access token
// @Description API response with an access token.
// @Produce json
// @Param token body TokenResponse true "Access token"
// @Success 200 {object} TokenResponse "Access token"
type TokenResponse struct {
	Token string `token:"token"`
}
