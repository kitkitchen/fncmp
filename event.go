package main

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

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
	ID       string         `json:"id"`
	TargetID string         `json:"target_id"`
	Handler  HandleFn       `json:"-"`
	On       OnEvent        `json:"on"`
	Action   string         `json:"action"`
	Method   string         `json:"method"`
	Data     map[string]any `json:"data"`
}

// TODO update this to regular dispatch func
// Creates a new EventListener with OnEvent for component with ID that triggers function f
func NewEventListener(on OnEvent, f FnComponent, h HandleFn) EventListener {
	id := uuid.New().String()
	el := EventListener{
		TargetID: f.id,
		Handler:  h,
		ID:       id,
		On:       on,
	}

	evtListeners.Add(id, el)
	return el
}

// Store and retrieve event listeners

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

func UnmarshalEventData[T any](e EventListener) (T, error) {
	var t T
	b, err := json.Marshal(e.Data)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(b, &t)
	return t, err
}
