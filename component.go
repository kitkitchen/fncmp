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

func HandleLogin(ctx context.Context) FnComponent {
	event, ok := ctx.Value(EventKey).(EventListener)
	//TODO: Include original component with event so it can be returned or appended to
	if !ok {
		return NewFn(HTML(`
		<div>Event not found</div>
	`))
	}
	if !ok {
		err := fmt.Errorf("error: expected user, got %T", event.Data)
		return LoginPage(context.WithValue(ctx, ErrorKey, err))
	}

	testEvent, err := UnmarshalEventData[EventTarget](event)
	if err != nil {
		err := fmt.Errorf("error: expected touch event, got %T", event.Data)
		return LoginPage(context.WithValue(ctx, ErrorKey, err))
	}
	fmt.Println(testEvent)

	user, err := UnmarshalEventData[User](event)
	if err != nil {
		return LoginPage(context.WithValue(ctx, ErrorKey, err))
	}

	if user.Username != "Sean" || user.Password != "password" {
		err = fmt.Errorf("invalid username or password")
		return NewFn(ErrorMessage(err)).SwapTargetInner("error_message")
		// return LoginPage(context.WithValue(ctx, ErrorKey, err))
	}

	return Welcome(context.WithValue(ctx, UserKey, user))
}

func Welcome(ctx context.Context) FnComponent {

	return NewFn(HTML(`
	<div>Welcome</div>
	`))
}

func LoginPage(ctx context.Context) FnComponent {
	err, ok := ctx.Value(ErrorKey).(error)
	msg := ""
	if ok {
		msg = err.Error()
	}
	fmt.Println(msg)

	input := HTML(`<input type="text" id="username" name="username" placeholder="Username">`)
	// TODO: FnComponent render function needs to be taking account for event listeners and applying them appropriately
	// inner FnComponent event listeners are not being applied
	inputFn := NewFn(input).WithEvents(func(ctx context.Context) FnComponent {
		return RedirectPage("/portfolio")
	}, OnFocus)

	return NewFn(Modal(inputFn))
}

func HandleExperienceFn(ctx context.Context) FnComponent {
	return NewFn(Experience())
}
func HandleMainFn(ctx context.Context) FnComponent {
	return NewFn(HTML(""))
}

func HandleProjectsFn(ctx context.Context) FnComponent {
	return NewFn(InfoCard())
}

type Contact struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func HandleContactFn(ctx context.Context) FnComponent {
	form := NewFn(ContactForm()).WithEvents(func(ctx context.Context) FnComponent {
		event := ctx.Value(EventKey).(EventListener)
		info, err := UnmarshalEventData[Contact](event)
		if err != nil {
			return NewFn(nil).WithError(err)
		}
		fmt.Println(info)

		return RedirectPage("/contact")
	}, OnSubmit)

	// TODO: Append form to a page component
	return form
}

func HandleWelcome(ctx context.Context) FnComponent {

	user, ok := ctx.Value(UserKey).(User)
	if !ok {
		return NewFn(HTML(`
		<div><p>User not found</p></div>
		`))
	}

	return NewFn(HTML(fmt.Sprintf(`
	<div>Hello %s</div>
	`, user.Username)))
}

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

func (f FnComponent) WithContext(ctx context.Context) FnComponent {
	f.Context = ctx
	return f
}

func (f FnComponent) WithEvents(h HandleFn, e ...OnEvent) FnComponent {
	for _, v := range e {
		el := NewEventListener(v, f, h)
		f.dispatch.FnRender.EventListeners = append(f.dispatch.FnRender.EventListeners, el)
	}
	return f
}

func (f FnComponent) WithRender() FnComponent {
	f.dispatch.Function = Render
	return f
}

func (f FnComponent) WithRedirect(url string) FnComponent {
	f.dispatch.Function = Redirect
	f.dispatch.FnRedirect.URL = url
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
	c := HTML(html)
	err := c.Render(f.Context, f)
	if err != nil {
		f.dispatch.FnError.Message = err.Error()
	}
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
