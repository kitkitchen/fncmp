package main

import "encoding/json"

type functionName string

const (
	render   functionName = "render"
	redirect functionName = "redirect"
	event    functionName = "event"
	custom   functionName = "custom"
	_error   functionName = "error"
)

type Tag string

const (
	HTMLTag Tag = "html"
	HeadTag Tag = "head"
	BodyTag Tag = "body"
	MainTag Tag = "main"
)

type (
	FnRender struct {
		TargetID       string          `json:"target_id"`
		Tag            Tag             `json:"tag"`
		Inner          bool            `json:"inner"`
		Outer          bool            `json:"outer"`
		Append         bool            `json:"append"`
		Prepend        bool            `json:"prepend"`
		HTML           string          `json:"html"`
		EventListeners []EventListener `json:"event_listeners"`
	}
	FnRedirect struct {
		URL string `json:"url"`
	}
	FnCustom struct {
		Function string `json:"function"`
		Data     any    `json:"data"`
	}
	FnError struct {
		Message string `json:"message"`
	}
)

func newDispatch(key string) *Dispatch {
	return &Dispatch{
		Key: key,
	}
}

// Dispatch contains necessary data for the web api.
//
// While this struct is exported, it is not intended to be used directly and is not exposed during runtime.
type Dispatch struct {
	buf        []byte        `json:"-"`
	conn       *conn         `json:"-"`
	ID         string        `json:"id"`
	Key        string        `json:"key"`
	ConnID     string        `json:"conn_id"`
	HandlerID  string        `json:"handler_id"`
	Action     string        `json:"action"`
	Label      string        `json:"label"`
	Function   functionName  `json:"function"`
	FnEvent    EventListener `json:"event"`
	FnRender   FnRender      `json:"render"`
	FnRedirect FnRedirect    `json:"redirect"`
	FnCustom   FnCustom      `json:"custom"`
	FnError    FnError       `json:"error"`
}

func (f *FnRender) listenerStrings() string {
	b, err := json.Marshal(f.EventListeners)
	if err != nil {
		return ""
	}
	return string(b)
}
