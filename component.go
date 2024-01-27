package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/kitkitchen/mnemo"
)

const components_store mnemo.StoreKey = "fn_component_cache_store"

var componentStore *mnemo.Store = nil

type fnCacheKey mnemo.CacheKey

func createStore() {
	store, err := mnemo.NewStore(components_store)
	if err != nil {
		panic(err)
	}
	componentStore = store
}

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

type FnComponent struct {
	ID string
	Renderer
	htmlCache *mnemo.Cache[string]
}

type FnComponentState struct {
	ID             string          `json:"id"`
	Html           string          `json:"html"`
	EventListeners []EventListener `json:"event_listeners"`
}

type Listener struct {
	Action string `json:"action"`
}

//TODO: Make a decoder that can parse custom html tags

type Renderer struct {
	component      func(d *Dispatch) Component
	html           string
	EventListeners []EventListener
	fcID           string
	elIds          []string
	context.Context
}

func (r *Renderer) Render(ctx context.Context, w io.Writer) error {
	dctx, ok := ctx.(ContextWithDispatch)
	if !ok {
		panic("fncmp: error casting context to ContextWithDispatch")
	}
	// Read the component html into the FnRenderer
	// and wrap it in a div with the component ID
	//TODO: Does this need to be an array in string format?
	els := ""
	for k, el := range r.EventListeners {
		els += el.String()
		if k != len(r.EventListeners)-1 {
			els += ","
		}
	}
	r.Read([]byte(fmt.Sprint(
		`<div id="` + r.fcID + `" ` + `fc=[` + els + `]>`,
	)))

	component := r.component(dctx.Dispatch)
	component.Render(dctx, r)
	r.Read([]byte("</div>"))
	_, err := w.Write([]byte(r.html))

	return err
}
func (r *Renderer) Write(p []byte) (n int, err error) {
	n, err = r.Read(p)
	return n, err
}

func (r *Renderer) Read(p []byte) (n int, err error) {
	r.html = r.html + string(p)
	return len(p), nil
}

func (r Renderer) Close() error {
	return nil
}

func NewFnComponent(f func(d *Dispatch) Component) FnComponent {
	ID := uuid.New().String()
	fc := FnComponent{
		ID: ID,
		Renderer: Renderer{
			component:      f,
			fcID:           ID,
			html:           "",
			EventListeners: []EventListener{},
			elIds:          []string{},
			Context:        context.Background(),
		},
	}

	if componentStore == nil {
		createStore()
	}
	// Create a cache for the component state
	key := fnCacheKey(fc.ID)
	fnCache, err := mnemo.NewCache[FnComponent](components_store, key)
	if err != nil {
		panic(err)
	}
	fnCache.SetReducer(func(state FnComponent) (mutation any) {
		return FnComponentState{
			ID:             state.ID,
			Html:           state.Renderer.html,
			EventListeners: state.Renderer.EventListeners,
		}
	})
	fnCache.Cache(key, &fc)

	fmt.Println(fnCache)

	return fc
}

func (f *FnComponent) WithContext(ctx context.Context) *FnComponent {
	f.Context = ctx
	return f
}

func (f *FnComponent) WithID(id string) *FnComponent {
	f.ID = id
	return f
}

// WithEventListeners sets variadic event listeners to a FnComponent
func (f *FnComponent) WithEventListeners(e ...EventListener) *FnComponent {
	f.EventListeners = append(f.EventListeners, e...)
	return f
}

func (fc FnComponent) ListenerStrings() (s string) {
	for _, el := range fc.EventListeners {
		s += el.String()
	}
	return s
}

func (fc *FnComponent) HtmlCache() *mnemo.Cache[string] {
	return fc.htmlCache
}

func WithFnCache[T any](f *FnComponent) (*mnemo.Cache[T], error) {
	cache, err := mnemo.NewCache[T](components_store, fnCacheKey("component_"+f.ID+"_dev_cache"))
	return cache, err
}

func UseFnCache[T any](f *FnComponent) (*mnemo.Cache[T], error) {
	cache, err := mnemo.UseCache[T](components_store, fnCacheKey("component_"+f.ID+"_dev_cache"))
	return cache, err
}

func (fc *FnComponent) Dispatch(d *Dispatch) {
	htmlBuf := new(bytes.Buffer)
	fc.Render(ContextWithDispatch{
		Context:  fc.Context,
		Dispatch: d,
	}, htmlBuf)
	d.Data = htmlBuf.String()
	d.send()
}
