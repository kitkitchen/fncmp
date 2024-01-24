package fncmp

import (
	"log"
	"net/http"
)

type (
	Opt[T any] func(t *T)
)

// Websocket connection	handler for handling FunComponents
func HandleFuns(w http.ResponseWriter, r *http.Request) {
	// upgrade to websocket
	conn, err := NewConn(w, r)
	if err != nil {
		return
	}

	cfg := ConnectionInfo{}
	cfg.ConnID = conn.ID

	fr := NewFunRequest("_connect", cfg)
	fr.Send(conn)

	log.Print("new connection: ", conn.ID)

	render := NewFunRequest(Render, RenderTargetRequest[ConnectionInfo]{
		ConnID:   conn.ID,
		TargetID: "root",
		Action:   "/main",
		Method:   "POST",
		Data:     cfg,
	})
	render.Send(conn)
	conn.listen()
}
