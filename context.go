package main

type ContextKey string

const (
	UserKey  ContextKey = "user"
	EventKey ContextKey = "listener"
)
