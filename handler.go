package fncmp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"

	"github.com/google/uuid"
)

var handlers = handlerPool{
	pool: make(map[string]handler),
}

type handlerPool struct {
	mu   sync.Mutex
	pool map[string]handler
}

func (h *handlerPool) Get(id string) (handler, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	handler, ok := h.pool[id]
	return handler, ok
}

func (h *handlerPool) Set(id string, handler handler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pool[id] = handler
}

func (h *handlerPool) Delete(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.pool, id)
}

type HandleFn func(context.Context) FnComponent

type handler struct {
	http.Handler
	id        string
	in        chan Dispatch
	out       chan FnComponent
	handlesFn map[string]HandleFn
	logger    log.Logger
}

func NewHandler() *handler {
	handler := handler{
		id:        uuid.New().String(),
		in:        make(chan Dispatch, 1028),
		out:       make(chan FnComponent, 1028),
		handlesFn: make(map[string]HandleFn),
		logger: *log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.Kitchen,
			Prefix:          "FnCmp: ",
		}),
	}
	handlers.Set(handler.id, handler)
	return &handler
}

func (h handler) ID() string {
	return h.id
}

func (h *handler) listen() {
	go func(h *handler) {
		for d := range h.in {
			switch d.Function {
			case event:
				h.Event(d)
			case _error:
				h.Error(d)
			default:
				d.FnError.Message = fmt.Sprintf(
					"function '%s' found, expected event or error on 'in' channel", d.Function)
				h.Error(d)
				return
			}
		}
	}(h)
	go func(h *handler) {
		for fn := range h.out {
			switch fn.dispatch.Function {
			case render:
				h.Render(fn)
			case redirect:
				h.Redirect(fn)
			case custom:
				h.Custom(fn)
			case _error:
				h.Error(*fn.dispatch)
			default:
				fn.dispatch.FnError.Message = fmt.Sprintf(
					"function '%s' found, expected event or error on 'in' channel", fn.dispatch.Function)
				h.Error(*fn.dispatch)
				return
			}
		}
	}(h)
}

func (h handler) Render(fn FnComponent) {
	// If there is no HTML to render, cancel dispatch
	if len(fn.dispatch.buf) == 0 && fn.dispatch.FnRender.HTML == "" {
		return
	}
	var data Writer
	fn.Render(context.Background(), &data)
	fn.dispatch.FnRender.HTML = SanitizeHTML(string(data.buf))
	h.MarshalAndPublish(*fn.dispatch)
}

func (h handler) Redirect(fn FnComponent) {
	// If there is no URL to redirect to, cancel dispatch
	if fn.dispatch.FnRedirect.URL == "" {
		return
	}
	h.MarshalAndPublish(*fn.dispatch)
}

func (h handler) Custom(fn FnComponent) {
	if fn.dispatch.FnCustom.Function == "" {
		return
	}
	h.MarshalAndPublish(*fn.dispatch)
}

func (h handler) MarshalAndPublish(d Dispatch) {
	if d.conn == nil {
		d.FnError.Message = "connection not found"
		h.Error(d)
		return
	}
	b, err := json.Marshal(d)
	if err != nil {
		d.FnError.Message = err.Error()
		h.Error(d)
		return
	}
	d.conn.Publish(b)
}

func (h handler) Event(d Dispatch) {
	listener, ok := evtListeners.Get(d.FnEvent.ID, d.conn)
	if !ok {
		d.FnError.Message = fmt.Sprintf("event listener with id '%s' not found", d.FnEvent.ID)
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

func (h handler) Error(d Dispatch) {

	log.Error(d.FnError)
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
				log.Error("failed to create new connection")
				log.Error(err)
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
