package fncmp

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

// RenderComponent is a helper function that takes a list of components and renders them to a string
//
// Example:
// 	html := RenderComponent(
// 		HTML("<h1>Hello, world!</h1>"),
// 		HTML("<p>This is a paragraph</p>"),
// 	)
// 	fmt.Println(html)
// 	// Output:
// 	// <h1>Hello, world!</h1><p>This is a paragraph</p>

func RenderComponent(c ...Component) (html string) {
	w := Writer{}
	for _, v := range c {
		v.Render(context.Background(), &w)
	}
	html = string(w.buf)
	return html
}

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
	}.SwapTagInner(MainTag)
	if c != nil {
		c.Render(f.Context, f)
	}
	return f
}

func (f FnComponent) Render(ctx context.Context, w io.Writer) error {
	w.Write([]byte(fmt.Sprint("<div id='" + f.id + "' label='" + f.dispatch.Label + "' events=" + f.dispatch.FnRender.ListenerStrings() + ">")))
	HTML(f.dispatch.FnRender.HTML).Render(ctx, w)
	w.Write(f.dispatch.buf)
	w.Write([]byte("</div>"))
	return nil
}

func (f FnComponent) Write(p []byte) (n int, err error) {
	f.dispatch.buf = append(f.dispatch.buf, p...)
	return len(p), nil
}

func (f FnComponent) WithContext(ctx context.Context) FnComponent {
	f.Context = ctx

	dd, ok := ctx.Value(dispatchKey).(dispatchDetails)
	if !ok {
		log.Warn("context does not contain dispatch details")
		return f
	}
	f.dispatch.ConnID = dd.ConnID
	f.dispatch.HandlerID = dd.HandlerID
	f.dispatch.conn = dd.Conn
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
	f.dispatch.Function = redirect
	f.dispatch.FnRedirect.URL = url
	return f
}

func (f FnComponent) WithError(err error) FnComponent {
	f.dispatch.Function = _error
	f.dispatch.FnError.Message = err.Error()
	return f
}

func (f FnComponent) JS(fn string, arg any) FnComponent {
	f.dispatch.Function = custom
	f.dispatch.FnCustom.Function = fn
	f.dispatch.FnCustom.Data = arg
	return f
}

// WithLabel sets the label of the component
//
// The label may be used to identify a component on the server and client,
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
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = true
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) PrependTag(t Tag) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = true
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) SwapTagOuter(t Tag) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = true
	return f
}

func (f FnComponent) SwapTagInner(t Tag) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = true
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) AppendTarget(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = true
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) PrependTarget(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = true
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) SwapTargetOuter(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = true
	return f
}

func (f FnComponent) SwapTargetInner(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = true
	f.dispatch.FnRender.Outer = false
	return f
}

func (f FnComponent) Dispatch() {
	if f.dispatch.conn == nil {
		log.Error("invalid dispatch", "conn", "nil")
		return
	}
	h, ok := handlers.Get(f.dispatch.HandlerID)
	if !ok {
		log.Printf("error: handler '%s' not found", f.dispatch.HandlerID)
		return
	}
	h.out <- f
}

func RedirectURL(url string) FnComponent {
	return NewFn(nil).WithRedirect(url)
}

func JS(ctx context.Context, fn string, arg any) {
	NewFn(nil).JS(fn, arg).DispatchContext(ctx)
}

func (f FnComponent) DispatchContext(ctx context.Context) {
	dd, ok := ctx.Value(dispatchKey).(dispatchDetails)
	if !ok {
		log.Error("context does not contain dispatch details")
	}
	if dd.Conn == nil || dd.HandlerID == "" {
		log.Error("invalid dispatch", "conn", dd.Conn, "handler", dd.HandlerID)
		return
	}
	f.WithContext(ctx).Dispatch()
}

// HTML implements the Component interface
type HTML string

func (h HTML) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(h))
	return err
}
