package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
)

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

func RenderComponent(c ...Component) (html string) {
	w := Writer{}
	for _, v := range c {
		v.Render(context.Background(), &w)
	}
	html = string(w.buf)
	return html
}

type HandleFn func(context.Context) FnComponent

type FnComponent struct {
	context.Context
	dispatch *Dispatch
	id       string
}

func NewFn(c Component) FnComponent {
	id := "fncmp-" + uuid.New().String()
	f := FnComponent{
		Context:  context.Background(),
		id:       id,
		dispatch: newDispatch(id),
	}.SwapTagInner("main")
	if c != nil {
		c.Render(f.Context, f)
	}
	return f
}

// TODO: render custom attributes with a helper function in the api
func (f FnComponent) Render(ctx context.Context, w io.Writer) error {
	w.Write([]byte(fmt.Sprint("<div id='" + f.id + "' events=" + f.dispatch.FnRender.ListenerStrings() + ">")))
	HTML(f.dispatch.FnRender.HTML).Render(ctx, w)
	w.Write(f.dispatch.buf)
	w.Write([]byte("</div>"))
	return nil
}

func (f FnComponent) Write(p []byte) (n int, err error) {
	f.dispatch.buf = append(f.dispatch.buf, p...)
	return len(p), nil
}

// FIXME: Add context in a more idiomatic way
func (f FnComponent) WithContext(ctx context.Context) FnComponent {
	f.Context = ctx

	dd, ok := ctx.Value(dispatchKey).(dispatchDetails)
	if !ok {
		log.Println("warn: context does not contain dispatch details")
		return f
	}
	f.dispatch.ConnID = dd.ConnID
	f.dispatch.HandlerID = dd.HandlerID
	f.dispatch.Conn = dd.Conn
	return f
}

func (f FnComponent) WithEvents(h HandleFn, e ...OnEvent) FnComponent {
	// get connection from context
	for _, v := range e {
		el := NewEventListener(v, f, h)
		f.dispatch.FnRender.EventListeners = append(f.dispatch.FnRender.EventListeners, el)
	}
	return f
}

func (f FnComponent) WithRedirect(url string) FnComponent {
	//TODO: Give the option to render before redirect
	// ie loading feedback while redirecting
	f.dispatch.Function = Redirect
	f.dispatch.FnRedirect.URL = url
	return f
}

func (f FnComponent) WithError(err error) FnComponent {
	f.dispatch.Function = Error
	f.dispatch.FnError.Message = err.Error()
	return f
}

func (f FnComponent) JS(fn string, arg any) FnComponent {
	f.dispatch.Function = Custom
	f.dispatch.FnCustom.Function = fn
	f.dispatch.FnCustom.Data = arg
	return f
}

// WithLabel sets the label of the component
//
// The label may be used to identify the component in the client,
// especially during debugging.
func (f FnComponent) WithLabel(label string) FnComponent {
	f.dispatch.Label = label
	return f
}

func (f FnComponent) WithTargetID(id string) FnComponent {
	f.dispatch.FnRender.TargetID = id
	return f
}

func (f FnComponent) WithTag(tag Tag) FnComponent {
	f.dispatch.FnRender.Tag = tag
	return f
}

func (f FnComponent) AppendTag(t Tag) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = true
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) PrependTag(t Tag) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = true
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) SwapTagOuter(t Tag) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = true
	return f
}

func (f FnComponent) SwapTagInner(t Tag) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = true
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) AppendTarget(id string) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = true
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) PrependTarget(id string) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = true
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) SwapTargetOuter(id string) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = true
	return f
}

func (f FnComponent) SwapTargetInner(id string) FnComponent {
	f.dispatch.Function = Render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = true
	f.dispatch.FnRender.Outer = false
	return f
}

func RedirectPage(url string) FnComponent {
	return NewFn(nil).WithRedirect(url)
}

func JS(ctx context.Context, fn string, arg any) {
	NewFn(nil).JS(fn, arg).DispatchContext(ctx)
}

func (f FnComponent) Dispatch() {
	if f.dispatch.Conn == nil {
		log.Println("error: connection not found")
		return
	}
	h, ok := handlers.Get(f.dispatch.HandlerID)
	if !ok {
		log.Printf("error: handler '%s' not found", f.dispatch.HandlerID)
		return
	}
	h.out <- f
}

func (f FnComponent) DispatchContext(ctx context.Context) {
	f.WithContext(ctx).Dispatch()
}

// HTML implements the Component interface
type HTML string

func (h HTML) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(h))
	return err
}
