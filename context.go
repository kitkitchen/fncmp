package main

import "context"

type ContextKey string

const (
	UserKey  ContextKey = "user"
	EventKey ContextKey = "listener"
)

type ContextWithDispatch struct {
	context.Context
	Dispatch
}
