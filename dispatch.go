package main

import (
	"context"
)

type ContextWithDispatch struct {
	context.Context
	Dispatch
}

const (
	Render   FunctionName = "render"
	Redirect FunctionName = "redirect"
	Event    FunctionName = "event"
	Error    FunctionName = "error"
	Custom   FunctionName = "custom"
)

const (
	HTMLTag Tag = "html"
	Head    Tag = "head"
	Body    Tag = "body"
)

type (
	FunctionName string
	Tag          string
	FnRender     struct {
		TargetID       string          `json:"target_id"`
		Tag            Tag             `json:"tag"`
		Inner          bool            `json:"inner"`
		Outer          bool            `json:"outer"`
		HTML           string          `json:"html"`
		EventListeners []EventListener `json:"event_listeners"`
	}
	FnCustom struct {
		Function string `json:"function"`
		Data     any    `json:"data"`
	}
	FnError struct {
		Message string `json:"message"`
	}
)

func (fn FnError) Error() string {
	return fn.Message
}

func newDispatch(key string) *Dispatch {
	return &Dispatch{
		Key: key,
	}
}

type Dispatch struct {
	buf       []byte        `json:"-"`
	Conn      *Conn         `json:"-"`
	ConnID    string        `json:"conn_id"`
	Key       string        `json:"key"`
	Function  FunctionName  `json:"function"`
	ID        string        `json:"id"`
	Action    string        `json:"action"`
	HandlerID string        `json:"handler_id"`
	Label     string        `json:"label"`
	FnEvent   EventListener `json:"event"`
	FnRender  FnRender      `json:"render"`
	FnError   FnError       `json:"error"`
	FnCustom  FnCustom      `json:"custom"`
}

func (d *Dispatch) Render() {

}
