package fncmp

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// The Component interface is implemented by types that can be rendered
type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

// RenderComponent renders a component and returns the HTML string
func RenderComponent(c ...Component) (html string) {
	w := Writer{}
	for _, v := range c {
		v.Render(context.Background(), &w)
	}
	html = string(w.buf)
	return html
}

// FnComponent is a functional component that can be rendered and dispatched
// to the client
type FnComponent struct {
	context.Context
	dispatch *Dispatch
	id       string
}

// NewFn creates a new FnComponent from a Component
func NewFn(ctx context.Context, c Component) FnComponent {
	id := "fncmp-" + uuid.New().String()

	dispatch := newDispatch(id)
	dd, ok := ctx.Value(dispatchKey).(dispatchDetails)
	if !ok {
		config.Logger.Warn(ErrCtxMissingDispatch)
	} else {
		dispatch.conn = dd.Conn
		dispatch.ConnID = dd.ConnID
		dispatch.HandlerID = dd.HandlerID
	}

	f := FnComponent{
		Context:  ctx,
		id:       id,
		dispatch: dispatch,
	}.SwapTagInner(MainTag)
	if c != nil {
		c.Render(f.Context, f)
	}
	return f
}

// Render renders the FnComponent with necessary metadata for the client
func (f FnComponent) Render(ctx context.Context, w io.Writer) error {
	if f.dispatch.Label == "" {
		w.Write([]byte(fmt.Sprint("<div id='" + f.id + "' events=" + f.dispatch.FnRender.listenerStrings() + ">")))
	} else {
		w.Write([]byte(fmt.Sprint("<div id='" + f.id + "' label='" + f.dispatch.Label + "' events=" + f.dispatch.FnRender.listenerStrings() + ">")))
	}
	HTML(f.dispatch.FnRender.HTML).Render(ctx, w)
	w.Write(f.dispatch.buf)
	w.Write([]byte("</div>"))
	return nil
}

// Write writes to the FnComponent's buffer
func (f FnComponent) Write(p []byte) (n int, err error) {
	f.dispatch.buf = append(f.dispatch.buf, p...)
	return len(p), nil
}

// WithContext sets the context of the FnComponent
func (f FnComponent) WithContext(ctx context.Context) FnComponent {
	f.Context = ctx

	dd, ok := ctx.Value(dispatchKey).(dispatchDetails)
	if !ok {
		config.Logger.Warn(ErrCtxMissingDispatch)
		return f
	}
	f.dispatch.ConnID = dd.ConnID
	f.dispatch.HandlerID = dd.HandlerID
	f.dispatch.conn = dd.Conn
	return f
}

// WithEvents sets the event listeners of the FnComponent with variadic OnEvent
func (f FnComponent) WithEvents(h HandleFn, e ...OnEvent) FnComponent {
	// get connection from context
	for _, v := range e {
		el := newEventListener(v, f, h)
		f.dispatch.FnRender.EventListeners = append(f.dispatch.FnRender.EventListeners, el)
	}
	return f
}

// WithRedirect sets the FnComponent to redirect to a URL
func (f FnComponent) WithRedirect(url string) FnComponent {
	f.dispatch.Function = redirect
	f.dispatch.FnRedirect.URL = url
	return f
}

// WithError sets the FnComponent to render an error
func (f FnComponent) WithError(err error) FnComponent {
	f.dispatch.Function = _error
	f.dispatch.FnError.Message = err.Error()
	return f
}

// JS sets the FnComponent to run a custom JavaScript function
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

// AppendTag appends the rendered component to a tag in the DOM
func (f FnComponent) AppendTag(t Tag) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = true
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

// PrependTag prepends the rendered component to a tag in the DOM
func (f FnComponent) PrependTag(t Tag) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = true
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

// SwapTagOuter swaps the rendered component with a tag in the DOM
func (f FnComponent) SwapTagOuter(t Tag) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = true
	return f
}

// SwapTagInner swaps the inner HTML of a tag in the DOM with the rendered component
func (f FnComponent) SwapTagInner(t Tag) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = t
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = true
	f.dispatch.FnRender.Outer = false
	return f
}

// AppendTarget appends the rendered component to an element by ID in the DOM
func (f FnComponent) AppendElement(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = true
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

// PrependTarget prepends the rendered component to an element by ID in the DOM
func (f FnComponent) PrependElement(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = true
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = false
	return f
}

// SwapElementOuter swaps the rendered component with an element by ID in the DOM
func (f FnComponent) SwapElementOuter(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = true
	return f
}

// SwapElementInner swaps the inner HTML of an element by ID in the DOM with the rendered component
func (f FnComponent) SwapElementInner(id string) FnComponent {
	f.dispatch.Function = render
	f.dispatch.FnRender.Tag = ""
	f.dispatch.FnRender.TargetID = id
	f.dispatch.FnRender.Append = false
	f.dispatch.FnRender.Prepend = false
	f.dispatch.FnRender.Inner = true
	f.dispatch.FnRender.Outer = false
	return f
}

// Dispatch immediately sends the FnComponent to the client
func (f FnComponent) Dispatch() {
	if f.dispatch.conn == nil {
		config.Logger.Error(ErrConnectionNotFound)
		return
	}
	h, ok := handlers.Get(f.dispatch.HandlerID)
	if !ok {
		config.Logger.Error("handler not found", "HandlerID", f.dispatch.HandlerID)
		return
	}
	h.out <- f
}

// RedirectURL redirects the client to the given url when returned from a handler
func RedirectURL(ctx context.Context, url string) FnComponent {
	return NewFn(ctx, nil).WithRedirect(url)
}

// JS runs a custom JavaScript function on the client
func JS(ctx context.Context, fn string, arg any) {
	NewFn(ctx, nil).JS(fn, arg).Dispatch()
}

// HTML implements the Component interface for a string of HTML
type HTML string

func (h HTML) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(h))
	return err
}
