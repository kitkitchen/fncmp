package main

import "encoding/json"

type FunctionName string

const (
	Render   FunctionName = "render"
	Redirect FunctionName = "redirect"
	Event    FunctionName = "event"
	Custom   FunctionName = "custom"
	Error    FunctionName = "error"
)

type Tag string

const (
	HTMLTag Tag = "html"
	Head    Tag = "head"
	Body    Tag = "body"
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

func (fn FnError) Error() string {
	return fn.Message
}

func newDispatch(key string) *Dispatch {
	return &Dispatch{
		Key: key,
	}
}

type Dispatch struct {
	buf        []byte        `json:"-"`
	Conn       *Conn         `json:"-"`
	ID         string        `json:"id"`
	Key        string        `json:"key"`
	ConnID     string        `json:"conn_id"`
	HandlerID  string        `json:"handler_id"`
	Action     string        `json:"action"`
	Label      string        `json:"label"`
	Function   FunctionName  `json:"function"`
	FnEvent    EventListener `json:"event"`
	FnRender   FnRender      `json:"render"`
	FnRedirect FnRedirect    `json:"redirect"`
	FnCustom   FnCustom      `json:"custom"`
	FnError    FnError       `json:"error"`
}

func (f *FnRender) ListenerStrings() string {
	b, err := json.Marshal(f.EventListeners)
	if err != nil {
		return ""
	}
	return string(b)
}
