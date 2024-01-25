package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Service struct {
}

type AppletManifest struct {
	Name  string   `json:"name"`
	ID    string   `json:"id"`
	Paths []string `json:"paths"`
	Port  int      `json:"port"`
}

type Applet struct {
	mu sync.Mutex
	AppletManifest
	conns    map[string]*Conn
	handlers map[string]http.HandlerFunc
	services map[string]Service
}

func NewApplet(manifest AppletManifest) *Applet {
	return &Applet{
		AppletManifest: manifest,
		conns:          make(map[string]*Conn),
		handlers:       make(map[string]http.HandlerFunc),
		services:       make(map[string]Service),
	}
}

func (applet *Applet) RegisterConn(conn_id string, conn *Conn) {
	applet.mu.Lock()
	defer applet.mu.Unlock()
	applet.conns[conn_id] = conn
}

func (applet *Applet) Handle(path string, handler http.Handler) *Applet {
	applet.mu.Lock()
	defer applet.mu.Unlock()
	applet.handlers[path] = handler.ServeHTTP
	return applet
}

func (applet *Applet) HandleFunc(path string, handler http.HandlerFunc) *Applet {
	applet.mu.Lock()
	defer applet.mu.Unlock()
	applet.handlers[path] = handler
	return applet
}

func (applet *Applet) RegisterService(name string, service Service) {
	applet.mu.Lock()
	defer applet.mu.Unlock()
	applet.services[name] = service
}

func (applet *Applet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	applet.mu.Lock()
	defer applet.mu.Unlock()
	fmt.Println(w)
	w.Header().Set("Content-Type", "*/*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	h, ok := applet.handlers[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	h(w, r)
}

func (applet *Applet) Run(port int) {
	applet.mu.Lock()
	applet.AppletManifest.Port = port
	applet.mu.Unlock()

	fmt.Println("Running applet on port: ", applet.AppletManifest.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", applet.AppletManifest.Port), applet)
}
