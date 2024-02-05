package main

import "context"

type ContextKey string

const (
	UserKey  ContextKey = "user"
	EventKey ContextKey = "event"
	ErrorKey ContextKey = "error"
)

type ContextWithDispatch struct {
	context.Context
	Dispatch
}
