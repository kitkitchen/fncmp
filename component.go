package main

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}
type HandleFn func(context.Context) FnComponent

type FnComponent struct {
	context.Context
	dispatch *Dispatch
	id       string
}

func NewFnComponent(c Component) FnComponent {
	id := "fncmp-" + uuid.New().String()
	f := FnComponent{
		Context:  context.Background(),
		id:       id,
		dispatch: newDispatch(id),
	}
	c.Render(f.Context, f)
	return f
}

func HandleMain(ctx context.Context) FnComponent {

	return HandleWelcome(ctx)
}

func HandleWelcome(ctx context.Context) FnComponent {

	user, ok := ctx.Value(UserKey).(User)
	if !ok {
		return NewFnComponent(HTML(`
		<div>User not found</div>
		`))
	}

	return NewFnComponent(HTML(fmt.Sprintf(`
	<div>Hello %s</div>
	`, user.Username)))
}

func (f FnComponent) Render(ctx context.Context, w io.Writer) error {
	w.Write([]byte(fmt.Sprint("<div id='" + f.id + "'>")))
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
	return f
}

func (f FnComponent) WithListeners(el ...EventListener) FnComponent {
	f.dispatch.FnRender.EventListeners = append(f.dispatch.FnRender.EventListeners, el...)
	return f
}

func (f FnComponent) WithRender() FnComponent {
	f.dispatch.Function = Render
	return f
}

func (f FnComponent) WithRedirect(action string) FnComponent {
	f.dispatch.Function = Redirect
	f.dispatch.Action = action
	return f
}

func (f FnComponent) WithError(err error) FnComponent {
	f.dispatch.Function = Error
	f.dispatch.FnError.Message = err.Error()
	return f
}

func (f FnComponent) WithCustom(fn string, arg any) FnComponent {
	f.dispatch.Function = Custom
	f.dispatch.FnCustom.Function = fn
	f.dispatch.FnCustom.Data = arg
	return f
}

func (f FnComponent) WithID(id string) FnComponent {
	f.id = id
	return f
}

func (f FnComponent) WithConnID(id string) FnComponent {
	f.dispatch.ConnID = id
	return f
}

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

func (f FnComponent) AppendHTML(html string) FnComponent {
	f.dispatch.FnRender.HTML += html
	return f
}

func (f FnComponent) InnerHTML() FnComponent {
	f.dispatch.FnRender.Outer = false
	f.dispatch.FnRender.Inner = true
	return f
}

func (f FnComponent) OuterHTML() FnComponent {
	f.dispatch.FnRender.Inner = false
	f.dispatch.FnRender.Outer = true
	return f
}

// HTML implements the Component interface
type HTML string

func (h HTML) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(h))
	return err
}

func RenderHTML(c ...Component) (html string) {
	w := Writer{}
	for _, v := range c {
		v.Render(context.Background(), &w)
	}
	html = string(w.buf)
	return html
}
