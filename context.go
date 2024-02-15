package main

import (
	"context"
	"net/http"
)

type ContextKey string

const (
	UserKey  ContextKey = "user"
	EventKey ContextKey = "event"
	ErrorKey ContextKey = "error"
)

type ContextWithRequest struct {
	context.Context
	*http.Request
	dispatchDetails
}

type dispatchDetails struct {
	ConnID    string
	Conn      *Conn
	HandlerID string
}
