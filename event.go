package fncmp

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

var evtHandlers = &eventHandlers{
	ef: make(map[string]http.HandlerFunc),
}

type eventHandlers struct {
	mu sync.Mutex
	ef map[string]http.HandlerFunc
}

func (e *eventHandlers) add(id string, h http.HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ef[id] = h
}

type OnEvent string

// Event types
const (
	OnAbort              OnEvent = "abort"
	OnAnimationEnd       OnEvent = "animationend"
	OnAnimationIteration OnEvent = "animationiteration"
	OnAnimationStart     OnEvent = "animationstart"
	OnBlur               OnEvent = "blur"
	OnCanPlay            OnEvent = "canplay"
	OnCanPlayThrough     OnEvent = "canplaythrough"
	OnChangeCapture      OnEvent = "changecapture"
	OnClick              OnEvent = "click"
	OnCompositionEnd     OnEvent = "compositionend"
	OnCompositionStart   OnEvent = "compositionstart"
	OnCompositionUpdate  OnEvent = "compositionupdate"
	OnContextMenuCapture OnEvent = "contextmenucapture"
	OnCopy               OnEvent = "copy"
	OnCut                OnEvent = "cut"
	OnDoubleClickCapture OnEvent = "doubleclickcapture"
	OnDrag               OnEvent = "drag"
	OnDragEnd            OnEvent = "dragend"
	OnDragEnter          OnEvent = "dragenter"
	OnDragExitCapture    OnEvent = "dragexitcapture"
	OnDragLeave          OnEvent = "dragleave"
	OnDragOver           OnEvent = "dragover"
	OnDragStart          OnEvent = "dragstart"
	OnDrop               OnEvent = "drop"
	OnDurationChange     OnEvent = "durationchange"
	OnEmptied            OnEvent = "emptied"
	OnEncrypted          OnEvent = "encrypted"
	OnEnded              OnEvent = "ended"
	OnError              OnEvent = "error"
	OnFocus              OnEvent = "focus"
	OnGotPointerCapture  OnEvent = "gotpointercapture"
	OnInput              OnEvent = "input"
	OnInvalid            OnEvent = "invalid"
	OnKeyDown            OnEvent = "keydown"
	OnKeyPress           OnEvent = "keypress"
	OnKeyUp              OnEvent = "keyup"
	OnLoad               OnEvent = "load"
	OnLoadEnd            OnEvent = "loadend"
	OnLoadStart          OnEvent = "loadstart"
	OnLoadedData         OnEvent = "loadeddata"
	OnLoadedMetadata     OnEvent = "loadedmetadata"
	OnLostPointerCapture OnEvent = "lostpointercapture"
	OnMouseDown          OnEvent = "mousedown"
	OnMouseEnter         OnEvent = "mouseenter"
	OnMouseLeave         OnEvent = "mouseleave"
	OnMouseMove          OnEvent = "mousemove"
	OnMouseOut           OnEvent = "mouseout"
	OnMouseOver          OnEvent = "mouseover"
	OnMouseUp            OnEvent = "mouseup"
	OnPause              OnEvent = "pause"
	OnPlay               OnEvent = "play"
	OnPlaying            OnEvent = "playing"
	OnPointerCancel      OnEvent = "pointercancel"
	OnPointerDown        OnEvent = "pointerdown"
	OnPointerEnter       OnEvent = "pointerenter"
	OnPointerLeave       OnEvent = "pointerleave"
	OnPointerMove        OnEvent = "pointermove"
	OnPointerOut         OnEvent = "pointerout"
	OnPointerOver        OnEvent = "pointerover"
	OnPointerUp          OnEvent = "pointerup"
	OnProgress           OnEvent = "progress"
	OnRateChange         OnEvent = "ratechange"
	OnResetCapture       OnEvent = "resetcapture"
	OnScroll             OnEvent = "scroll"
	OnSeeked             OnEvent = "seeked"
	OnSeeking            OnEvent = "seeking"
	OnSelectCapture      OnEvent = "selectcapture"
	OnStalled            OnEvent = "stalled"
	OnSubmit             OnEvent = "submit"
	OnSuspend            OnEvent = "suspend"
	OnTimeUpdate         OnEvent = "timeupdate"
	OnToggle             OnEvent = "toggle"
	OnTouchCancel        OnEvent = "touchcancel"
	OnTouchEnd           OnEvent = "touchend"
	OnTouchMove          OnEvent = "touchmove"
	OnTouchStart         OnEvent = "touchstart"
	OnTransitionEnd      OnEvent = "transitionend"
	OnVolumeChange       OnEvent = "volumechange"
	OnWaiting            OnEvent = "waiting"
	OnWheel              OnEvent = "wheel"
)

type EventListener struct {
	ID     string  `json:"id"`
	On     OnEvent `json:"on"`
	Action string  `json:"action"`
	Method string  `json:"method"`
}

func NewEventListener(on OnEvent, h http.HandlerFunc, o ...Opt[EventListener]) (action string, el EventListener) {
	id := uuid.New().String()
	el = EventListener{
		ID: id,
		On: on,
	}
	for _, opt := range o {
		opt(&el)
	}

	evtHandlers.add(id, h)
	evtListeners.Add(id, el)
	return el.Action, el
}

func WithAction(action string) Opt[EventListener] {
	return func(e *EventListener) {
		e.Action = action
	}
}

func WithMethod(method string) Opt[EventListener] {
	return func(e *EventListener) {
		e.Method = method
	}
}

func (e EventListener) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(b)
}

type eventListeners struct {
	mu sync.Mutex
	eh map[string]EventListener
}

var evtListeners = eventListeners{
	eh: make(map[string]EventListener),
}

func (e *eventListeners) Add(id string, el EventListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eh[id] = el
}

func (e *eventListeners) Remove(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.eh, id)
}

func (e *eventListeners) Get(id string) (EventListener, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	event, ok := e.eh[id]
	return event, ok
}

func (e *eventListeners) Every() (el []EventListener) {
	for _, e := range e.eh {
		el = append(el, e)
	}
	return
}
