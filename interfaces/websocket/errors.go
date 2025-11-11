// interfaces/websocket/errors.go
package websocket

type WSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	ErrNotAuthorized  = WSError{Code: "NOT_AUTHORIZED", Message: "Not authorized"}
	ErrInvalidMessage = WSError{Code: "INVALID_MESSAGE", Message: "Invalid message format"}
	ErrNotMember      = WSError{Code: "NOT_MEMBER", Message: "Not a member of conversation"}
)
