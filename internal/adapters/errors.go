package adapters

import "fmt"

const (
	ErrorMissingCredential = "missing_credential"
	ErrorConnectionFailed  = "connection_failed"
	ErrorTimeout           = "timeout"
	ErrorUnauthorized      = "unauthorized"
	ErrorRateLimited       = "rate_limited"
	ErrorUpstreamStatus    = "upstream_status"
	ErrorInvalidResponse   = "invalid_response"
	ErrorEmptyResponse     = "empty_response"
)

type ProviderError struct {
	ProviderID string
	Code       string
	Message    string
	StatusCode int
	Err        error
}

func (e *ProviderError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("provider %s %s: %s (status %d)", e.ProviderID, e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("provider %s %s: %s", e.ProviderID, e.Code, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}
