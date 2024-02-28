package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// TODO: Use a logger

var handlers = handlerPool{
	pool: make(map[string]Handler),
}

type handlerPool struct {
	mu   sync.Mutex
	pool map[string]Handler
}

func (h *handlerPool) Get(id string) (Handler, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	handler, ok := h.pool[id]
	return handler, ok
}

func (h *handlerPool) Set(id string, handler Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pool[id] = handler
}

func (h *handlerPool) Delete(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.pool, id)
}

type Handler struct {
	http.Handler
	id          string
	in          chan Dispatch
	out         chan FnComponent
	handlesFn   map[string]HandleFn
	handlesHTTP map[string]http.HandlerFunc
}

func NewHandler() *Handler {
	handler := Handler{
		id:          uuid.New().String(),
		in:          make(chan Dispatch, 1028),
		out:         make(chan FnComponent, 1028),
		handlesFn:   make(map[string]HandleFn),
		handlesHTTP: make(map[string]http.HandlerFunc),
	}
	handlers.Set(handler.id, handler)
	return &handler
}

func (h Handler) ID() string {
	return h.id
}

// Dispatch handles the routing of a Dispatch to the appropriate function.
// It is called by the Conn's listen method.
func (h *Handler) listen() {
	go func(h *Handler) {
		for d := range h.in {
			switch d.Function {
			case Event:
				h.Event(d)
			case Error:
				h.Error(d)
			default:
				d.FnError.Message = fmt.Sprintf(
					"function '%s' found, expected event or error on 'in' channel", d.Function)
				h.Error(d)
				return
			}
		}
	}(h)
	go func(h *Handler) {
		for fn := range h.out {
			switch fn.dispatch.Function {
			case Render:
				h.Render(fn)
			case Redirect:
				h.Redirect(fn)
			case Custom:
				h.Custom(fn)
			case Error:
				h.Error(*fn.dispatch)
			}
		}
	}(h)
}

// TODO: these should be Dispatch receiver methods
func (h Handler) Render(fn FnComponent) {
	// If there is no HTML to render, cancel dispatch
	if len(fn.dispatch.buf) == 0 && fn.dispatch.FnRender.HTML == "" {
		return
	}
	var data Writer
	fn.Render(context.Background(), &data)
	fn.dispatch.FnRender.HTML = SanitizeHTML(string(data.buf))
	h.MarshalAndPublish(*fn.dispatch)
}

func (h Handler) Redirect(fn FnComponent) {
	fmt.Println(fn)
	h.MarshalAndPublish(*fn.dispatch)
}

func (h Handler) Custom(fn FnComponent) {
	h.MarshalAndPublish(*fn.dispatch)
}

func (h Handler) MarshalAndPublish(d Dispatch) {
	b, err := json.Marshal(d)
	if err != nil {
		log.Println(err)
		d.FnError.Message = err.Error()
		h.Error(d)
		return
	}
	d.conn.Publish(b)
}

func (h Handler) Event(d Dispatch) {
	listener, ok := evtListeners.Get(d.FnEvent.ID, d.conn)
	if !ok {
		d.FnError.Message = fmt.Sprintf("error: event listener with id '%s' not found", d.FnEvent.ID)
		h.Error(d)
		return
	}
	listener.Data = d.FnEvent.Data

	ctx := context.WithValue(listener.Context, EventKey, listener)
	response := listener.Handler(ctx)
	response.dispatch.conn = d.conn
	response.dispatch.HandlerID = d.HandlerID
	h.out <- response
}

func (h Handler) Error(d Dispatch) {
	log.Println(d.FnError)
}

type Writer struct {
	http.ResponseWriter
	buf []byte
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func MiddleWareFn(h http.HandlerFunc, hf HandleFn) http.HandlerFunc {
	//FIXME: This should be stored more gracefully
	// newConns := make(map[string]context.Context)
	handler := NewHandler()

	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: keep cookies for session management

		// srvAddr := r.Context().Value(http.LocalAddrContextKey).(net.Addr)

		// get id url param:
		id := r.URL.Query().Get("fncmp_id")
		if id == "" {
			writer := Writer{ResponseWriter: w}
			h(&writer, r)
			// hf(r.Context()).Render(r.Context(), &writer)

			//TODO: inject local storage script to store id for websocket

			w.Write(writer.buf)
		} else {
			newConnection, err := newConn(w, r, handler.id, id)
			if err != nil {
				log.Println("error: failed to create new connection")
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error: failed to create new connection"))
				return
			}
			newConnection.HandlerID = handler.id
			//TODO: write cookie

			//TODO: get connection id from cookie and pass previous context

			ctx := context.WithValue(r.Context(), dispatchKey, dispatchDetails{
				ConnID:    id,
				Conn:      newConnection,
				HandlerID: handler.id,
			})
			ctx = context.WithValue(ctx, RequestKey, r)

			// fnContext := FnContext{
			// 	Context: r.Context(),
			// 	Request: r,
			// 	dispatchDetails: dispatchDetails{
			// 		ConnID:    id,
			// 		Conn:      newConnection,
			// 		HandlerID: handler.id,
			// 	},
			// }

			fn := hf(ctx)
			fn.dispatch.conn = newConnection
			fn.dispatch.ConnID = id
			fn.dispatch.HandlerID = handler.id
			handler.out <- fn
			handler.listen()
			newConnection.listen()
		}
	}
}
