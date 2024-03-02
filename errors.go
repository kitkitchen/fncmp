package fncmp

type DispatchError string

func (e DispatchError) Error() string {
	return string(e)
}

const (
	ErrInvalidFunction    DispatchError = "invalid function"
	ErrInvalidAction      DispatchError = "invalid action"
	ErrInvalidMethod      DispatchError = "invalid method"
	ErrInvalidEvent       DispatchError = "invalid event"
	ErrInvalidID          DispatchError = "invalid id"
	ErrInvalidLabel       DispatchError = "invalid label"
	ErrInvalidConnID      DispatchError = "invalid conn_id"
	ErrInvalidTargetID    DispatchError = "invalid target_id"
	ErrInvalidInner       DispatchError = "invalid inner"
	ErrInvalidHTML        DispatchError = "invalid html"
	ErrInvalidFormData    DispatchError = "invalid form_data"
	ErrInvalidHeaders     DispatchError = "invalid headers"
	ErrInvalidBody        DispatchError = "invalid body"
	ErrInvalidMessage     DispatchError = "invalid message"
	ErrNoClientConnection DispatchError = "no connection to client"
)
