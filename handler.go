package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/kitkitchen/mnemo"
)

type Service struct {
}

type Handler struct {
	mu            sync.Mutex
	id            string
	port          string
	handlers      map[string]http.HandlerFunc
	fncmpHandlers map[string]*FnComponent
}

func NewHandler(port string) *Handler {
	//TODO: Call an open API endpoint with package update information
	// Call for contributers etc.
	// TODO: Cache and make available on an endpoint all api usage statistics
	// User can opt out of this
	// If they opt in, they must log in to admin panel and change the default password.
	// Admin panel can also be registered with commands from mnemo
	h := &Handler{
		id:            uuid.New().String(),
		port:          port,
		handlers:      make(map[string]http.HandlerFunc),
		fncmpHandlers: make(map[string]*FnComponent),
	}
	RegisterDispatcher(h)
	return h
}

func (h *Handler) Dispatch(d *Dispatch) {

	caches, err := mnemo.UseCache[FnComponent](components_store, fnCacheKey(d.TargetID))
	if err != nil {
		if d.Action == "main" {
			d.send()
			return
		}
		handler, ok := h.fncmpHandlers[d.Action]
		if !ok {
			log.Println("error: handler not found")
			return
		}
		handler.Dispatch(d)
		return
	}
	cache, ok := caches.Get(d.TargetID)
	if !ok {
		log.Println("error: component not found in cache")
		return
	}
	fnComponent := cache.Data

	switch d.Function {
	case Render:
		fnComponent.Dispatch(d)
	case Redirect:
		fmt.Println("redirect")
		// make http request to HandlerFunc
		// change function to "render"
		// Replace the targetID with the body //TODO: give body a unique id
		// send the dispatch to the websocket
	case Event:
		handler, ok := evtHandlers.get(d.Event.ID)
		if !ok {
			log.Println("error: event handler not found")
			return
		}
		handler(d)
		// find event listener in global registry
		// call event listener function with dispatch
		// find function in global registry, pass a *copy* of dispatch to function and render the
		// returned component with the Dispatch renderer if the component is not nil.
		// then send the dispatch to the websocket
		// TODO: utility functions for dispatch cancellation, overwriting, etc.
	case Error:
		fmt.Println(fmt.Errorf(d.Message))
	default:
		// Call user custom function?
		fmt.Println("default")
	}
}

func (h *Handler) Handle(path string, handler http.Handler) *Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[path] = handler.ServeHTTP
	return h
}

func (h *Handler) HandleFunc(path string, handler http.HandlerFunc) *Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[path] = handler
	return h
}

func (h *Handler) HandleFnCmp(path string, fn FnComponent) *Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.fncmpHandlers[path] = &fn
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	cookie, err := r.Cookie(h.id)

	if err != nil {
		log.Println("error: ", err)
		cookie = NewCookie(h.id, uuid.NewString(), r.URL.Path)
		r.AddCookie(cookie)
	}
	http.SetCookie(w, cookie)

	paths := strings.Split(r.URL.Path, "/")

	switch paths[1] {
	case h.id:
		// upgrade to websocket
		conn, err := NewConn(w, r, cookie.Value)
		if err != nil {
			return
		}

		log.Print("new connection: ", conn.ID)

		main := Dispatch{
			Function: "_connect",
			TargetID: "root",
			ConnID:   conn.ID,
		}
		b, err := main.Marshall()
		if err != nil {
			panic(err)
		}

		go conn.listen()
		conn.Publish(b)

		_, ok := connPool.Get(cookie.Value)
		if !ok {
			log.Println("error: connection not found")
			panic(err)
		}

		d := Dispatch{
			Function: "render",
			TargetID: "body",
			ConnID:   cookie.Value,
			Action:   "/" + paths[3],
			Method:   r.Method,
			Message:  "New HTTP request",
		}

		h.Dispatch(&d)

	default:
		script := js(h.port, h.id, r.URL.Path)
		html := fmt.Sprintf("<html><head><script>%s</script><body id='body'><div id='root'></div></body></html>", script)
		w.Write([]byte(html))
	}
	// return

	_, ok := h.handlers[r.URL.Path]
	if ok {
		h.handlers[r.URL.Path](w, r)
	}
}

type Tester struct {
	io.Writer
	http.ResponseWriter
}

func (h *Handler) ListenAndServe() {
	fmt.Println("Serving fncmp handler on port: ", h.port)
	http.ListenAndServe(fmt.Sprintf(h.port), h)
}
