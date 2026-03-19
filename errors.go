package trading212

import "fmt"

// Error represents an error returned by the Trading 212 API. It is returned
// for any HTTP response with a status code >= 400.
//
// Use [errors.As] to inspect API errors:
//
//	var apiErr *trading212.Error
//	if errors.As(err, &apiErr) {
//	    fmt.Println(apiErr.StatusCode, apiErr.Message)
//	    if apiErr.StatusCode == 429 {
//	        fmt.Println("retry after", apiErr.RateLimit.RetryAfter)
//	    }
//	}
type Error struct {
	// StatusCode is the HTTP status code (e.g. 400, 401, 404, 429).
	StatusCode int

	// Code is the Trading 212 error code string included in the response body,
	// if any (e.g. "InsufficientFunds").
	Code string `json:"code"`

	// Message is the human-readable error message from the response body.
	// If the body is not JSON, this contains the raw body text.
	Message string `json:"message"`

	// RateLimit holds rate-limiting metadata from the response headers.
	// Most useful when StatusCode is 429: check RateLimit.RetryAfter for the
	// suggested wait duration.
	RateLimit RateLimit
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("trading212: status %d: %s (code: %s)", e.StatusCode, e.Message, e.Code)
	}
	return fmt.Sprintf("trading212: status %d: %s", e.StatusCode, e.Message)
}
