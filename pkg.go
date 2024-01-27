package main

type (
	Opt[T any] func(t *T)
)

var dispatcher Dispatcher

type Dispatcher interface {
	Dispatch(d *Dispatch)
}

func RegisterDispatcher(d Dispatcher) {
	dispatcher = d
}

// // Websocket connection	handler for handling FunComponents
// func HandleFncmp(w http.ResponseWriter, r *http.Request) {
// 	// upgrade to websocket
// 	conn, err := NewConn(w, r)
// 	if err != nil {
// 		return
// 	}

// 	log.Print("new connection: ", conn.ID)

// 	main := Dispatch[string]{
// 		Function: "_connect",
// 		TargetID: "root",
// 		ConnID:   conn.ID,
// 	}
// 	b, err := main.Marshall()
// 	if err != nil {
// 		panic(err)
// 	}

// 	conn.Publish(b)
// 	conn.listen()
// }
