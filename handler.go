package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// TODO: use mutex to lock handlers map
var handlers = make(map[string]Handler)

type Handler struct {
	http.Handler
	port        string
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
	handlers[handler.id] = handler
	return &handler
}

func (h Handler) ID() string {
	return h.id
}

func (h *Handler) HandleFn(path string, handler HandleFn) {
	h.handlesFn[path] = handler
}

func (h *Handler) HandleFunc(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	h.handlesHTTP[path] = handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get cookie and check if it exists
	cookie, err := r.Cookie(h.id)
	if err != nil {
		log.Println(err)
		// Set cookie
		cookie = NewCookie(h.id, uuid.NewString(), r.URL.Path)
		http.SetCookie(w, cookie)
		r.AddCookie(cookie)
	}

	// split path
	path := strings.Split(r.URL.Path, "/")

	if path[len(path)-1] == h.id {
		var p string
		if cookie.Path == "" {
			p = "/"
		} else {
			p = cookie.Path
		}
		h.HandleWS(w, r, cookie.Value, p)
		return
	} else {
		fmt.Println(path[len(path)-1])
	}

	writer := Writer{ResponseWriter: w}
	var socketPath string
	if r.URL.Path == "/" {
		socketPath = r.URL.Path + h.id
	} else {
		socketPath = r.URL.Path + "/" + h.id
	}
	script := js(h.port, socketPath)
	writer.Write([]byte("<script>" + script + "</script>"))
	// writer.Write([]byte("<body></body>"))
	if h.handlesHTTP[r.URL.Path] != nil {
		// FIXME: This is probably where bad url paths are getting stuck
		MiddleWareDispatch(h.handlesHTTP[r.URL.Path]).ServeHTTP(&writer, r)
	} else {
		h.Handler.ServeHTTP(w, r)
	}

	w.Write(writer.buf)
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request, id string, path string) {
	// Upgrade connection to websocket
	conn, err := NewConn(w, r, h.id, id)
	if err != nil {
		log.Println(err)
		return
	}
	conn.HandlerID = h.id

	go conn.listen()
	handler, ok := h.handlesFn[path]
	if !ok {
		log.Printf("error: no handler found for path '%s'", path)
		return
	}

	fn := handler(r.Context()).WithConnID(conn.ID).
		WithRender().SwapTagInner(Body)
	fn.dispatch.Conn = conn
	fn.dispatch.HandlerID = h.id
	h.out <- fn
	fmt.Println(fn)
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
			case Error:
				h.Error(*fn.dispatch)
			}
		}
	}(h)
}

// TODO: these should be Dispatch receiver methods
func (h Handler) Render(fn FnComponent) {
	var data Writer
	fn.Render(context.Background(), &data)
	fn.dispatch.FnRender.HTML = SanitizeHTML(string(data.buf))
	h.MarshalAndPublish(*fn.dispatch)
}

func (h Handler) Redirect(fn FnComponent) {
	fmt.Println(fn)
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
	d.Conn.Publish(b)
}

func (h Handler) Event(d Dispatch) {
	listener, ok := evtListeners.Get(d.FnEvent.ID)
	if !ok {
		d.FnError.Message = fmt.Sprintf("error: event listener with id '%s' not found", d.FnEvent.ID)
		h.Error(d)
		return
	}
	listener.Data = d.FnEvent.Data
	response := listener.Handler(context.WithValue(context.Background(), EventKey, listener))
	response.dispatch.Conn = d.Conn
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

func MiddleWareDispatch(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get route and pass to handler with information
		next(w, r)
	}
}

func MiddleWareFn(h http.HandlerFunc, hf HandleFn) http.HandlerFunc {
	//FIXME: This should be stored more gracefully
	// newConns := make(map[string]context.Context)
	handler := NewHandler()

	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: keep cookies for session management

		// srvAddr := r.Context().Value(http.LocalAddrContextKey).(net.Addr)

		//TODO: is this being used?
		handler.HandleFn(r.URL.Path, hf)

		// get id url param:
		id := r.URL.Query().Get("fncmp_id")
		if id == "" {
			writer := Writer{ResponseWriter: w}
			h(&writer, r)
			// hf(r.Context()).Render(r.Context(), &writer)

			//TODO: inject local storage script to store id for websocket

			w.Write(writer.buf)
		} else {
			newConnection, err := NewConn(w, r, handler.id, id)
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
			// Add event listeners only
			fn := hf(r.Context())
			fn.dispatch.Conn = newConnection
			fn.dispatch.ConnID = id
			fn.dispatch.HandlerID = handler.id
			// TODO: specify function to only process event listeners
			// Perhaps grab every event listener from document on first render
			handler.out <- fn
			handler.listen()
			newConnection.listen()
		}
	}
}

func ParsePort(addr string) string {
	port := strings.Split(addr, ":")
	return port[len(port)-1]
}

const TestKey ContextKey = "test"

func MiddleWareTest(next http.HandlerFunc) http.HandlerFunc {

	// ctx := context.WithValue(context.Background(), TestKey, "test")

	return func(w http.ResponseWriter, r *http.Request) {

		next(w, r)
	}
}
