package fncmp

type DispatchError string

func (e DispatchError) Error() string {
	return string(e)
}

const (
	ErrCtxMissingDispatch DispatchError = "context missing dispatch details"
	ErrNoClientConnection DispatchError = "no connection to client"
	ErrConnectionNotFound DispatchError = "connection not found"
	ErrConnectionFailed   DispatchError = "connection failed"
	ErrCtxMissingEvent    DispatchError = "context missing event details"
)
