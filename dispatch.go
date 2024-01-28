package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
)

// For passing to render interface
type ContextWithDispatch struct {
	context.Context
	Dispatch
}

// TODO:
type Dispatch struct {
	context.Context `json:"-"`
	Function        FunctionName `json:"function"`
	// todo: mask connID
	ConnID         string          `json:"conn_id"`
	TargetID       string          `json:"target_id"`
	Inner          bool            `json:"inner"`
	Action         string          `json:"action"`
	Method         string          `json:"method"`
	Event          EventListener   `json:"event"`
	Data           string          `json:"data"`
	EventListeners []EventListener `json:"event_listeners"`
	Message        string          `json:"message"`
}

func (d *Dispatch) WithContext(ctx context.Context) *Dispatch {
	d.Context = ctx
	return d
}

func (d *Dispatch) WithFunction(f FunctionName) *Dispatch {
	d.Function = f
	return d
}

func (d *Dispatch) WithEventListeners(el ...EventListener) *Dispatch {
	d.EventListeners = append(d.EventListeners, el...)
	return d
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

func (d *Dispatch) AppendComponent(c Component) *Dispatch {
	buf := new(bytes.Buffer)
	c.Render(ContextWithDispatch{
		Context:  context.Background(),
		Dispatch: *d,
	}, buf)
	d.Data = buf.String()
	return d
}

func (d *Dispatch) AppendHTML(html string) *Dispatch {
	d.Data += html
	return d
}

func (d *Dispatch) SwapComponent(c Component) *Dispatch {
	buf := new(bytes.Buffer)
	c.Render(d.Context, buf)
	d.Data = buf.String()
	return d
}

func (d *Dispatch) InnerHTML() *Dispatch {
	d.Inner = true
	return d
}

func (d *Dispatch) Target(id string) *Dispatch {
	d.TargetID = id
	return d
}

func (d Dispatch) send() {
	conn, ok := connPool.Get(d.ConnID)
	if !ok {
		log.Println("fncmp: ERROR retrieving connection to client: " + d.ConnID)
	}
	b, err := json.Marshal(d)
	if err != nil {
		log.Println(err)
	}
	conn.Publish(b)
}

func (d Dispatch) Marshal() ([]byte, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return b, nil
}
