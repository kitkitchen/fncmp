package main

import (
	"context"
	"net/http"
)

type ContextKey string

const (
	UserKey     ContextKey = "user"
	EventKey    ContextKey = "event"
	RequestKey  ContextKey = "request"
	ErrorKey    ContextKey = "error"
	dispatchKey ContextKey = "__dispatch__"
)

type FnContext struct {
	context.Context
	Event EventListener
	*http.Request
	dispatchDetails
}

type dispatchDetails struct {
	ConnID    string
	Conn      *Conn
	HandlerID string
}
