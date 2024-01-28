package main

import (
	"log"
	"net/http"
)

var dispatcher Dispatcher

type (
	Opt[T any]      func(t *T)
	FunctionName    string
	DispatchHandler func(d Dispatch) Dispatch
	DispatchFunc    func(d *Dispatch)
)

const (
	Render   FunctionName = "render"
	Redirect FunctionName = "redirect"
	Event    FunctionName = "event"
	Error    FunctionName = "error"
)

func RegisterDispatcher(d Dispatcher) {
	dispatcher = d
}

type Dispatcher interface {
	Dispatch(d Dispatch)
	Serve()
	Handle(string, DispatchHandler)
	Error(Dispatch)
}

type DispatchManager struct {
	Dispatcher
	in       chan Dispatch
	out      chan Dispatch
	handlers map[string]DispatchHandler
}

func NewDispatchManager() *DispatchManager {
	return &DispatchManager{
		in:       make(chan Dispatch, 1028),
		out:      make(chan Dispatch, 1028),
		handlers: make(map[string]DispatchHandler),
	}
}

func (dm *DispatchManager) Handle(path string, handler DispatchHandler) {
	dm.handlers[path] = handler
}

func (dm *DispatchManager) Dispatch(d Dispatch) {
	handler, ok := dm.handlers[d.Action]
	if !ok {
		log.Printf("error: handler with action '%s' not found", d.Action)
		return
	}
	dm.in <- handler(d)
}

func (dm *DispatchManager) Serve() {
	for {
		select {
		case d := <-dm.in:
			dm.Dispatch(d)
		}
	}
}

func MiddleWareDispatchAdapter(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
	}
}
