package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/kitkitchen/mnemo"
)

const componentStore mnemo.StoreKey = "fn_component_cache_store"

func init() {
	_, err := mnemo.NewStore(componentStore)
	if err != nil {
		panic(err)
	}
}

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

type FnComponent[Cache any] struct {
	ID string
	FnRenderer
	EventListeners []EventListener
	elIds          []string
	cacheKey       mnemo.CacheKey
	cache          *mnemo.Cache[Cache]
}

type Listener struct {
	Action string `json:"action"`
}

//TODO: Make a decoder that can parse custom html tags

var test = func(w io.Writer, r io.ReadCloser) Component {

	return nil
}

type FnRenderer struct {
	html []byte
	context.Context
}

func (f FnRenderer) Render(ctx context.Context, w io.Writer) error {
	f.Context = ctx
	_, err := w.Write(f.html)
	return err
}
func (f FnRenderer) Write(p []byte) (n int, err error) {
	n, err = f.Read(p)
	return n, err
}

func (f FnRenderer) Read(p []byte) (n int, err error) {
	f.html = append(f.html, p...)
	return len(p), nil
}

func (f FnRenderer) Close() error {
	return nil
}

var x = func(w http.ResponseWriter, r *http.Request) {
	// f := templ.ComponentScript{}.Function
	// y := New(test).Render(context.Background(), w)
}

// TODO: !!! Event listener should be added to the component html during the render process and not here
func New[Cache any](f func(w io.Writer, r io.ReadCloser) Component) *FnComponent[Cache] {
	fc := FnComponent[Cache]{
		FnRenderer: FnRenderer{
			html:    []byte{},
			Context: context.Background(),
		},
		EventListeners: []EventListener{},
		elIds:          []string{},
	}

	// Create a cache for the component
	fc.cacheKey = mnemo.CacheKey(fc.ID)
	cache, _ := mnemo.NewCache[Cache](componentStore, fc.cacheKey)
	cache.SetReducer(cache.DefaultReducer)

	fmt.Println(cache)

	// Read the component html into the FnRenderer
	// and wrap it in a div with the component ID

	//TODO: !!! Event listeners should be added to the component html
	// during the render process and not here
	// Add it to the javascript file instead
	fc.Read([]byte(fmt.Sprint(
		`<div id="` + fc.ID + `">`,
	)))
	component := f(fc, fc)
	component.Render(fc.Context, fc)
	fc.Read([]byte("</div>"))

	return &fc
}

// WithEventListeners sets variadic event listeners to a FnComponent
func (f *FnComponent[any]) WithEventListeners(e ...EventListener) *FnComponent[any] {
	for _, el := range e {
		f.elIds = append(f.elIds, el.ID)
	}
	return f
}

func (fc FnComponent[any]) ListenerStrings() (s string) {
	for _, el := range fc.EventListeners {
		s += el.String()
	}
	return s
}

func (fc *FnComponent[any]) Cache() *mnemo.Cache[any /* type parameter */] {
	return fc.cache
}
