package teta

type HTTPError struct {
	code    int
	message string
}
type HTTPErrorMessage struct {
	Message string `json:"message"`
}

func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		code:    code,
		message: message,
	}
}

func (e *HTTPError) Error() string {
	return e.message
}
