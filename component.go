package fncmp

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

type FunComponent struct {
	// TOODO: THIS NEEDS TO BE IMPLEMENTED PROPERLY
	ID string
	Component
	// templ.Component
	w              http.ResponseWriter
	r              *http.Request
	scripts        []templ.ComponentScript
	EventListeners []EventListener
	elIds          []string
}

func NewFC(f func(w http.ResponseWriter, r *http.Request) Component, opts ...Opt[FunComponent]) FunComponent {
	fc := FunComponent{
		scripts:        []templ.ComponentScript{},
		EventListeners: []EventListener{},
	}
	if f == nil {
		fc.Component = Nil()
		return fc
	}

	for _, opt := range opts {
		opt(&fc)
	}
	for _, script := range fc.scripts {
		fmt.Println(script)
	}
	if fc.ID == "" {
		fc.ID = uuid.New().String()
	}
	for _, el := range fc.EventListeners {
		fc.elIds = append(fc.elIds, el.ID)
	}

	userComponent := f(fc.w, fc.r)
	if userComponent != nil {
		fc.Component = userComponent
	} else {
		fc.Component = Nil()
	}
	fc.Component = fC(fc)
	return fc
}

// WithID sets the ID of the component which default to a randomly generated UUID string
func WithID(id string) Opt[FunComponent] {
	return func(c *FunComponent) {
		c.ID = id
	}
}

// WithEventListeners sets variadic event listeners to a FunComponent
func WithEventListeners(e ...EventListener) Opt[FunComponent] {
	return func(c *FunComponent) {
		c.EventListeners = e
	}
}

func WithScripts(s ...templ.ComponentScript) Opt[FunComponent] {
	return func(c *FunComponent) {
		c.scripts = s
	}
}

func WithHandler(w http.ResponseWriter, r *http.Request) Opt[FunComponent] {
	return func(fc *FunComponent) {
		fc.r = r
		fc.w = w
	}
}

func (fc FunComponent) ListenerStrings() (s string) {
	for _, el := range fc.EventListeners {
		s += el.String()
	}
	return s
}
