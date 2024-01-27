package main

import (
	"encoding/json"
	"net/http"
)

// Should this be DispatchType?

type FunctionName string

const (
	Render   FunctionName = "render"
	Redirect FunctionName = "redirect"
	Event    FunctionName = "event"
	Error    FunctionName = "error"
)

// Specialized requests

// TODO: Write a function to parse what type of request a handler is receiving
//
// RenderTargetRequest renders and html string parsed from the supplied action (handler) to a specified DOM element
// (target) with the specified ID.
// Data T will be passed to the handler in the request body
type RenderTargetRequest[T any] struct {
	ConnID   string `json:"conn_id"`
	TargetID string `json:"target_id"`
	Inner    bool   `json:"inner"`
	Action   string `json:"action"`
	Method   string `json:"method"`
	Data     T      `json:"data"`
}

func (ren *RenderTargetRequest[any]) Parse(r *http.Request) bool {
	if err := json.NewDecoder(r.Body).Decode(ren); err != nil {
		return false
	}
	return true
}

// EventsRequest is received from the client when an event is triggered
//
// Use Parse(*http.Request) to parse the request body with a given 'TargetData' type
type EventRequest[TargetData any] struct {
	ConnID   string        `json:"conn_id"`
	TargetID string        `json:"target_id"`
	Event    EventListener `json:"event"`
	Data     TargetData    `json:"data"`
}

func NewEventRequest[TargetData any]() *EventRequest[TargetData] {
	return &EventRequest[TargetData]{}
}

func (er *EventRequest[TargetData]) Parse(r *http.Request) (EventRequest[TargetData], bool) {
	err := json.NewDecoder(r.Body).Decode(er)
	if err != nil {
		return EventRequest[TargetData]{}, false
	}
	_, ok := evtListeners.Get(er.Event.ID)
	if !ok {
		// event was not assigned by the server
		return EventRequest[TargetData]{}, false
	}
	return *er, ok
}
