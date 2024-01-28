package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	ID    string
	Label string
	FnRenderer
}

//TODO: Make a decoder that can parse custom html tags

type FnRenderer struct {
	FnID      string
	component func(d Dispatch) Dispatch
}

func (r FnRenderer) Render(ctx context.Context, w io.Writer) error {
	dctx, ok := ctx.(ContextWithDispatch)
	if !ok {
		panic("fncmp: error casting context to ContextWithDispatch")
	}
	dispatch := r.component(dctx.Dispatch)
	if dispatch.EventListeners != nil {
		b, err := json.Marshal(dispatch.EventListeners)
		if err != nil {
			return err
		}
		dispatch.Data = "<div id='" + r.FnID + "' fnc=" + string(b) + ">" + dispatch.Data + "</div>"
	} else {
		dispatch.Data = "<div id='" + r.FnID + "'>" + dispatch.Data + "</div>"
	}
	_, err := w.Write([]byte(dispatch.Data))
	return err
	// Read the component html into the FnRenderer
	// and wrap it in a div with the component ID
	//TODO: Does this need to be an array in string format?
	// TODO: Add event listeners to dispatch as it's being passed
	// then process all at once without using fc tags
	// els := ""
	// for k, el := range r.EventListeners {
	// 	els += el.String()
	// 	if k != len(r.EventListeners)-1 {
	// 		els += ","
	// 	}
	// }
}

func NewFnComponent(f func(d Dispatch) Dispatch) FnComponent {
	ID := uuid.New().String()
	fc := FnComponent{
		ID: ID,
		FnRenderer: FnRenderer{
			FnID:      ID,
			component: f,
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
		return state
	})
	fnCache.Cache(key, &fc)

	fmt.Println(fnCache)

	return fc
}

func (f *FnComponent) WithID(id string) *FnComponent {
	f.ID = id
	return f
}

func WithFnCache[T any](f *FnComponent) (*mnemo.Cache[T], error) {
	cache, err := mnemo.NewCache[T](components_store, fnCacheKey("component_"+f.ID+"_dev_cache"))
	return cache, err
}

func UseFnCache[T any](f *FnComponent) (*mnemo.Cache[T], error) {
	cache, err := mnemo.UseCache[T](components_store, fnCacheKey("component_"+f.ID+"_dev_cache"))
	return cache, err
}

func (fc FnComponent) Dispatch(d Dispatch) {
	rendered := d
	htmlBuf := new(bytes.Buffer)
	fc.Render(
		ContextWithDispatch{
			Context:  d.Context,
			Dispatch: d,
		}, htmlBuf)
	rendered.Data = htmlBuf.String()
	rendered.send()
}
