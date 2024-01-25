package main

import (
	"errors"
	"net/http"
)

// Utitility functions for rendering components

func RenderTarget[Data any](r *http.Request, targetID string, action string, opts ...Opt[RenderTargetRequest[Data]]) error {
	if r == nil {
		return errors.New("request is nil")
	}
	// get headers from request
	connId := r.Header.Get("Conn")
	if connId == "" {
		return errors.New("failed to parse request for connection")
	}
	conn, ok := UseConnPool().Get(connId)
	if !ok {
		return errors.New("failed to get connection from pool")
	}

	fr := NewFunRequest(Render, RenderTargetRequest[Data]{
		ConnID:   connId,
		TargetID: targetID,
		Action:   action,
		Method:   "POST",
	})
	for _, opt := range opts {
		opt(&fr.Data)
	}
	fr.Send(conn)
	return nil
}

// Append data to the request body
func WithData[Data any](data Data) Opt[RenderTargetRequest[Data]] {
	return func(r *RenderTargetRequest[Data]) {
		r.Data = data
	}
}

// Render the component inside the target element
func OutHTML[Data any]() Opt[RenderTargetRequest[Data]] {
	return func(r *RenderTargetRequest[Data]) {
		r.Inner = false
	}
}

// Render the component inside the target element
func InnerHTML[Data any]() Opt[RenderTargetRequest[Data]] {
	return func(r *RenderTargetRequest[Data]) {
		r.Inner = true
	}
}
