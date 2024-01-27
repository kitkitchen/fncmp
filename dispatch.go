package main

import (
	"context"
	"encoding/json"
	"log"
)

// For passing to render interface
type ContextWithDispatch struct {
	context.Context
	*Dispatch
}

// TODO: Function component creates a new dispatch, it then reads the html from the component render function
// and wraps it in a div with the component ID. It then sends the dispatch to the client.

// TODO:
type Dispatch struct {
	Function FunctionName `json:"function"`
	// todo: mask connID
	ConnID   string `json:"conn_id"`
	TargetID string `json:"target_id"`
	Inner    bool   `json:"inner"`
	Action   string `json:"action"`
	Method   string `json:"method"`
	Event    struct {
		On     string `json:"on"`
		Action string `json:"action"`
		Method string `json:"method"`
		ID     string `json:"id"`
	} `json:"event"`
	Data    string `json:"data"`
	Message string `json:"message"`
}

func (d Dispatch) IsRender() bool {
	return d.Function == Render
}

func (d Dispatch) IsRedirect() bool {
	return d.Function == Redirect
}

func (d Dispatch) IsEvent() bool {
	return d.Function == Event
}

func (d Dispatch) IsError() bool {
	return d.Function == Error
}

func (d *Dispatch) send() {
	conn, ok := connPool.Get(d.ConnID)
	if !ok {
		panic("fncmp: error retrieving connection to DOM: " + d.ConnID)
	}
	b, err := json.Marshal(d)
	if err != nil {
		log.Println(err)
	}
	conn.Publish(b)
}

func (d *Dispatch) Marshall() ([]byte, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return b, nil
}
