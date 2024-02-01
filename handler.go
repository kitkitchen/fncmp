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

func NewHandler(port string, h http.Handler) *Handler {
	handler := Handler{
		id:          uuid.New().String(),
		port:        port,
		Handler:     h,
		in:          make(chan Dispatch, 1028),
		out:         make(chan FnComponent, 1028),
		handlesFn:   make(map[string]HandleFn),
		handlesHTTP: make(map[string]http.HandlerFunc),
	}
	handlers[handler.id] = handler
	handler.listen()
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
		WithRender().WithTag("body").InnerHTML()
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
			}
		}
	}(h)
}

func (h Handler) Render(fn FnComponent) {
	var data Writer
	fn.Render(context.Background(), &data)
	fn.dispatch.FnRender.HTML = string(data.buf)
	b, err := json.Marshal(fn.dispatch)
	if err != nil {
		log.Println(err)
		fn.dispatch.FnError.Message = err.Error()
		h.Error(*fn.dispatch)
		return
	}
	fn.dispatch.Conn.Publish(b)
}

func (h Handler) Redirect(fn FnComponent) {
	fmt.Println(fn)
	// Handler redirect
}

func (h Handler) Event(d Dispatch) {
	listener, ok := evtListeners.Get(d.FnEvent.ID)
	if !ok {
		d.FnError.Message = fmt.Sprintf("error: event listener with id '%s' not found", d.FnEvent.ID)
		h.Error(d)
		return
	}
	response := listener.Handler(context.WithValue(context.Background(), EventKey, d.FnEvent))
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
