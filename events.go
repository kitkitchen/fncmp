package main

import (
	"context"
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
	OnChange             OnEvent = "change"
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
	context.Context `json:"-"`
	ID              string   `json:"id"`
	TargetID        string   `json:"target_id"`
	Handler         HandleFn `json:"-"`
	On              OnEvent  `json:"on"`
	Action          string   `json:"action"`
	Method          string   `json:"method"`
	Data            any      `json:"data"`
}

// TODO update this to regular dispatch func
// Creates a new EventListener with OnEvent for component with ID that triggers function f
func NewEventListener(on OnEvent, f FnComponent, h HandleFn) EventListener {
	id := uuid.New().String()
	el := EventListener{
		Context:  f.Context,
		ID:       id,
		TargetID: f.id,
		Handler:  h,
		On:       on,
	}
	evtListeners.Add(f.dispatch.Conn, el)
	return el
}

// Store and retrieve event listeners

type eventListeners struct {
	mu sync.Mutex
	el map[string]map[string]EventListener
}

var evtListeners = eventListeners{
	el: make(map[string]map[string]EventListener),
}

func (e *eventListeners) Add(conn *Conn, el EventListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.el[conn.ID]; !ok {
		e.el[conn.ID] = make(map[string]EventListener)
	}
	e.el[conn.ID][el.ID] = el
}

func (e *eventListeners) Remove(conn *Conn) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.el, conn.ID)
}

func (e *eventListeners) Get(id string, conn *Conn) (EventListener, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	event, ok := e.el[conn.ID][id]
	return event, ok
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

// Event data types
type EventTarget struct {
	ID         string   `json:"id"`
	ClassList  []string `json:"classList"`
	TagName    string   `json:"tagName"`
	InnerHTML  string   `json:"innerHTML"`
	OuterHTML  string   `json:"outerHTML"`
	Value      string   `json:"value"`
	Checked    bool     `json:"checked"`
	Disabled   bool     `json:"disabled"`
	Hidden     bool     `json:"hidden"`
	Style      string   `json:"style"`
	Attributes []string `json:"attributes"`
	Dataset    []string `json:"dataset"`
}

type PointerEvent struct {
	IsTrusted        bool        `json:"isTrusted"`
	AltKey           bool        `json:"altKey"`
	Bubbles          bool        `json:"bubbles"`
	Button           int         `json:"button"`
	Buttons          int         `json:"buttons"`
	Cancelable       bool        `json:"cancelable"`
	ClientX          int         `json:"clientX"`
	ClientY          int         `json:"clientY"`
	Composed         bool        `json:"composed"`
	CtrlKey          bool        `json:"ctrlKey"`
	CurrentTarget    EventTarget `json:"currentTarget"`
	DefaultPrevented bool        `json:"defaultPrevented"`
	Detail           int         `json:"detail"`
	EventPhase       int         `json:"eventPhase"`
	Height           int         `json:"height"`
	IsPrimary        bool        `json:"isPrimary"`
	MetaKey          bool        `json:"metaKey"`
	MovementX        int         `json:"movementX"`
	MovementY        int         `json:"movementY"`
	OffsetX          int         `json:"offsetX"`
	OffsetY          int         `json:"offsetY"`
	PageX            int         `json:"pageX"`
	PageY            int         `json:"pageY"`
	PointerId        int         `json:"pointerId"`
	PointerType      string      `json:"pointerType"`
	Pressure         int         `json:"pressure"`
	RelatedTarget    EventTarget `json:"relatedTarget"`
}

type TouchEvent struct {
	ChangedTouches []Touch `json:"changedTouches"`
	TargetTouches  []Touch `json:"targetTouches"`
	Touches        []Touch `json:"touches"`
	LayerX         int     `json:"layerX"`
	LayerY         int     `json:"layerY"`
	PageX          int     `json:"pageX"`
	PageY          int     `json:"pageY"`
}

type Touch struct {
	ClientX       int         `json:"clientX"`
	ClientY       int         `json:"clientY"`
	Identifier    int         `json:"identifier"`
	PageX         int         `json:"pageX"`
	PageY         int         `json:"pageY"`
	RadiusX       float64     `json:"radiusX"`
	RadiusY       float64     `json:"radiusY"`
	RotationAngle int         `json:"rotationAngle"`
	ScreenX       int         `json:"screenX"`
	ScreenY       int         `json:"screenY"`
	Target        EventTarget `json:"target"`
}

type DragEvent struct {
	IsTrusted        bool        `json:"isTrusted"`
	AltKey           bool        `json:"altKey"`
	Bubbles          bool        `json:"bubbles"`
	Button           int         `json:"button"`
	Buttons          int         `json:"buttons"`
	Cancelable       bool        `json:"cancelable"`
	ClientX          int         `json:"clientX"`
	ClientY          int         `json:"clientY"`
	Composed         bool        `json:"composed"`
	CtrlKey          bool        `json:"ctrlKey"`
	CurrentTarget    EventTarget `json:"currentTarget"`
	DefaultPrevented bool        `json:"defaultPrevented"`
	Detail           int         `json:"detail"`
	EventPhase       int         `json:"eventPhase"`
	MetaKey          bool        `json:"metaKey"`
	MovementX        int         `json:"movementX"`
	MovementY        int         `json:"movementY"`
	OffsetX          int         `json:"offsetX"`
	OffsetY          int         `json:"offsetY"`
	PageX            int         `json:"pageX"`
	PageY            int         `json:"pageY"`
	RelatedTarget    EventTarget `json:"relatedTarget"`
}

type MouseEvent struct {
	IsTrusted        bool        `json:"isTrusted"`
	AltKey           bool        `json:"altKey"`
	Bubbles          bool        `json:"bubbles"`
	Button           int         `json:"button"`
	Buttons          int         `json:"buttons"`
	Cancelable       bool        `json:"cancelable"`
	ClientX          int         `json:"clientX"`
	ClientY          int         `json:"clientY"`
	Composed         bool        `json:"composed"`
	CtrlKey          bool        `json:"ctrlKey"`
	CurrentTarget    EventTarget `json:"currentTarget"`
	DefaultPrevented bool        `json:"defaultPrevented"`
	Detail           int         `json:"detail"`
	EventPhase       int         `json:"eventPhase"`
	MetaKey          bool        `json:"metaKey"`
	MovementX        int         `json:"movementX"`
	MovementY        int         `json:"movementY"`
	OffsetX          int         `json:"offsetX"`
	OffsetY          int         `json:"offsetY"`
	PageX            int         `json:"pageX"`
	PageY            int         `json:"pageY"`
	RelatedTarget    EventTarget `json:"relatedTarget"`
}

type KeyboardEvent struct {
	IsTrusted        bool        `json:"isTrusted"`
	AltKey           bool        `json:"altKey"`
	Bubbles          bool        `json:"bubbles"`
	Cancelable       bool        `json:"cancelable"`
	Code             string      `json:"code"`
	Composed         bool        `json:"composed"`
	CtrlKey          bool        `json:"ctrlKey"`
	CurrentTarget    EventTarget `json:"currentTarget"`
	DefaultPrevented bool        `json:"defaultPrevented"`
	Detail           int         `json:"detail"`
	EventPhase       int         `json:"eventPhase"`
	IsComposing      bool        `json:"isComposing"`
	Key              string      `json:"key"`
	Location         int         `json:"location"`
	MetaKey          bool        `json:"metaKey"`
	Repeat           bool        `json:"repeat"`
	ShiftKey         bool        `json:"shiftKey"`
}

type FormDataEvent struct {
	IsTrusted        bool           `json:"isTrusted"`
	Bubbles          bool           `json:"bubbles"`
	Cancelable       bool           `json:"cancelable"`
	Composed         bool           `json:"composed"`
	CurrentTarget    EventTarget    `json:"currentTarget"`
	DefaultPrevented bool           `json:"defaultPrevented"`
	EventPhase       int            `json:"eventPhase"`
	FormData         map[string]any `json:"formData"`
}
