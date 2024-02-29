package fncmp

type ContextKey string

const (
	EventKey    ContextKey = "event"
	RequestKey  ContextKey = "request"
	ErrorKey    ContextKey = "error"
	dispatchKey ContextKey = "__dispatch__"
)

type dispatchDetails struct {
	ConnID    string
	Conn      *conn
	HandlerID string
}
