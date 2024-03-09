// let functions: DispatchFunctions;
let conn_id = undefined;
let base_url = undefined;
let verbose = false;
class Socket {
    constructor() {
        this.ws = null;
        this.addr = undefined;
        this.key = undefined;
        let key = localStorage.getItem("fncmp_key");
        if (!key) {
            key = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
                var r = (Math.random() * 16) | 0, v = c == "x" ? r : (r & 0x3) | 0x8;
                return v.toString(16);
            });
            localStorage.setItem("fncmp_key", key);
        }
        this.key = key;
        let path = window.location.pathname.split("");
        let path_parsed = "";
        if (path[-1] == "/" || (path.length == 1 && path[0] == "/")) {
            path.pop();
        }
        path_parsed = path.join("");
        if (path_parsed == "") {
            path_parsed = "/";
        }
        this.addr = "ws://" + window.location.host + path_parsed + "?fncmp_id=" + this.key;
        this.connect();
    }
    connect() {
        try {
            this.ws = new WebSocket(this.addr);
        }
        catch (_a) {
            throw new Error("ws: failed to connect to server...");
        }
        this.ws.onopen = function () { };
        this.ws.onclose = function () { };
        this.ws.onerror = function () { };
        this.ws.onmessage = function (event) {
            let d = JSON.parse(event.data);
            api.Process(this, d);
        };
    }
}
class API {
    constructor() {
        this.ws = null;
        this.Dispatch = (data) => {
            if (!data)
                return;
            if (!this.ws) {
                throw new Error("ws: not connected to server...");
            }
            this.ws.send(JSON.stringify(data));
        };
        this.funs = {
            render: (d) => {
                let elem = null;
                const parsed = new DOMParser().parseFromString(d.render.html, "text/html").firstChild;
                const html = parsed.getElementsByTagName("body")[0].innerHTML;
                if (d.render.tag != "") {
                    elem = document.getElementsByTagName(d.render.tag)[0];
                    if (!elem) {
                        return this.Error(d, "element with tag not found: " + d.render.tag);
                    }
                }
                else if (d.render.target_id != "") {
                    elem = document.getElementById(d.render.target_id);
                    if (!elem) {
                        return this.Error(d, "element with target_id not found: " +
                            d.render.target_id);
                    }
                }
                else {
                    return this.Error(d, "no target or tag specified");
                }
                if (d.render.inner) {
                    elem.innerHTML = html;
                }
                if (d.render.outer) {
                    elem.outerHTML = html;
                }
                if (d.render.append) {
                    elem.innerHTML += html;
                }
                if (d.render.prepend) {
                    elem.innerHTML = html + elem.innerHTML;
                }
                d = this.utils.parseEventListeners(elem, d);
                this.Dispatch(this.utils.addEventListeners(d));
                return;
            },
            class: (d) => {
                const elem = document.getElementById(d.class.target_id);
                if (!elem) {
                    return this.Error(d, "element not found");
                }
                if (d.class.remove) {
                    elem.classList.remove(...d.class.names);
                }
                else {
                    elem.classList.add(...d.class.names);
                }
                return;
            },
            custom: (d) => {
                d.custom.result = window[d.custom.function](d.custom.data);
                return d;
            },
        };
        this.utils = {
            parseEventListeners: (element, d) => {
                const events = this.utils.getAttributes(element, "events");
                const listeners = events.map((e) => {
                    const event = JSON.parse(e);
                    if (!event)
                        return;
                    return event;
                });
                const listeners_flat = listeners.flat();
                const listeners_filtered = listeners_flat.filter((e) => e != null);
                d.render.event_listeners = listeners_filtered;
                return d;
            },
            // Element selectors
            parseFormData: (ev, d) => {
                const form = ev.target;
                const formData = new FormData(form);
                d.event.data = Object.fromEntries(formData.entries());
                return d;
            },
            getAttributes: (elem, attribute) => {
                const elems = elem.querySelectorAll(`[${attribute}]`);
                return Array.from(elems).map((el) => el.getAttribute(attribute));
            },
            addEventListeners: (d) => {
                if (!d.render.event_listeners)
                    return;
                // Event listeners
                d.render.event_listeners.forEach((listener) => {
                    let elem = document.getElementById(listener.target_id);
                    if (!elem) {
                        this.Error(d, "element not found");
                        return;
                    }
                    if (elem.firstChild) {
                        elem = elem.firstChild;
                    }
                    elem.addEventListener(listener.on, (ev) => {
                        ev.preventDefault();
                        d.function = "event";
                        d.event = listener;
                        switch (listener.on) {
                            case "submit":
                                d = this.utils.parseFormData(ev, d);
                                break;
                            case "pointerdown" || "pointerup" || "pointermove" || "click" || "contextmenu" || "dblclick":
                                d.event.data = ParsePointerEvent(ev);
                                break;
                            case "drag" || "dragend" || "dragenter" || "dragexitcapture" || "dragleave" || "dragover" || "dragstart" || "drop":
                                d.event.data = ParseDragEvent(ev);
                                break;
                            case "mousedown" || "mouseup" || "mousemove":
                                d.event.data = ParseMouseEvent(ev);
                                break;
                            case "keydown" || "keyup" || "keypress":
                                d.event.data = ParseKeyboardEvent(ev);
                                break;
                            case "change" || "input" || "invalid" || "reset" || "search" || "select" || "focus" || "blur" || "copy" || "cut" || "paste":
                                d.event.data = ParseEventTarget(ev.target);
                                break;
                            case "touchstart" || "touchend" || "touchmove" || "touchcancel":
                                d.event.data = ParseTouchEvent(ev);
                                break;
                            default:
                                d.event.data = ParseEventTarget(ev.target);
                        }
                        this.Dispatch(d);
                    });
                });
            },
        };
        this.Error = (d, message) => {
            d.function = "error";
            d.error.message = message;
            this.Dispatch(d);
        };
    }
    Process(ws, d) {
        if (!this.ws) {
            this.ws = ws;
        }
        switch (d.function) {
            case "redirect":
                window.location.href = d.redirect.url;
                break;
            default:
                if (!this.funs[d.function]) {
                    this.Error(d, "function not found: " + d.function);
                    break;
                }
                const result = this.funs[d.function](d);
                this.Dispatch(result);
                break;
        }
    }
}
function ParseEventTarget(ev) {
    return {
        id: ev.id || "",
        name: ev.name || "",
        tagName: ev.tagName || "",
        innerHTML: ev.innerHTML || "",
        outerHTML: ev.outerHTML || "",
        value: ev.value || "",
    };
}
function ParsePointerEvent(ev) {
    return {
        isTrusted: ev.isTrusted,
        altKey: ev.altKey,
        bubbles: ev.bubbles,
        button: ev.button,
        buttons: ev.buttons,
        cancelable: ev.cancelable,
        clientX: ev.clientX,
        clientY: ev.clientY,
        composed: ev.composed,
        ctrlKey: ev.ctrlKey,
        currentTarget: ParseEventTarget(ev.currentTarget),
        defaultPrevented: ev.defaultPrevented,
        detail: ev.detail,
        eventPhase: ev.eventPhase,
        height: ev.height,
        isPrimary: ev.isPrimary,
        metaKey: ev.metaKey,
        movementX: ev.movementX,
        movementY: ev.movementY,
        offsetX: ev.offsetX,
        offsetY: ev.offsetY,
        pageX: ev.pageX,
        pageY: ev.pageY,
        pointerId: ev.pointerId,
        pointerType: ev.pointerType,
        pressure: ev.pressure,
        relatedTarget: ParseEventTarget(ev.relatedTarget),
    };
}
function ParseTouchEvent(ev) {
    return {
        changedTouches: Array.from(ev.changedTouches).map((t) => ParseTouch(t)),
        targetTouches: Array.from(ev.targetTouches).map((t) => ParseTouch(t)),
        touches: Array.from(ev.touches).map((t) => ParseTouch(t)),
        layerX: ev.layerX,
        layerY: ev.layerY,
        pageX: ev.pageX,
        pageY: ev.pageY,
    };
}
function ParseTouch(ev) {
    return {
        clientX: ev.clientX,
        clientY: ev.clientY,
        identifier: ev.identifier,
        pageX: ev.pageX,
        pageY: ev.pageY,
        radiusX: ev.radiusX,
        radiusY: ev.radiusY,
        rotationAngle: ev.rotationAngle,
        screenX: ev.screenX,
        screenY: ev.screenY,
        target: ParseEventTarget(ev.target),
    };
}
function ParseDragEvent(ev) {
    return {
        isTrusted: ev.isTrusted,
        altKey: ev.altKey,
        bubbles: ev.bubbles,
        button: ev.button,
        buttons: ev.buttons,
        cancelable: ev.cancelable,
        clientX: ev.clientX,
        clientY: ev.clientY,
        composed: ev.composed,
        ctrlKey: ev.ctrlKey,
        currentTarget: ParseEventTarget(ev.currentTarget),
        defaultPrevented: ev.defaultPrevented,
        detail: ev.detail,
        eventPhase: ev.eventPhase,
        metaKey: ev.metaKey,
        movementX: ev.movementX,
        movementY: ev.movementY,
        offsetX: ev.offsetX,
        offsetY: ev.offsetY,
        pageX: ev.pageX,
        pageY: ev.pageY,
        relatedTarget: ParseEventTarget(ev.relatedTarget),
    };
}
function ParseMouseEvent(ev) {
    return {
        isTrusted: ev.isTrusted,
        altKey: ev.altKey,
        bubbles: ev.bubbles,
        button: ev.button,
        buttons: ev.buttons,
        cancelable: ev.cancelable,
        clientX: ev.clientX,
        clientY: ev.clientY,
        composed: ev.composed,
        ctrlKey: ev.ctrlKey,
        currentTarget: ParseEventTarget(ev.currentTarget),
        defaultPrevented: ev.defaultPrevented,
        detail: ev.detail,
        eventPhase: ev.eventPhase,
        metaKey: ev.metaKey,
        movementX: ev.movementX,
        movementY: ev.movementY,
        offsetX: ev.offsetX,
        offsetY: ev.offsetY,
        pageX: ev.pageX,
        pageY: ev.pageY,
        relatedTarget: ParseEventTarget(ev.relatedTarget),
    };
}
function ParseKeyboardEvent(ev) {
    return {
        isTrusted: ev.isTrusted,
        altKey: ev.altKey,
        bubbles: ev.bubbles,
        cancelable: ev.cancelable,
        code: ev.code,
        composed: ev.composed,
        ctrlKey: ev.ctrlKey,
        currentTarget: ParseEventTarget(ev.currentTarget),
        defaultPrevented: ev.defaultPrevented,
        detail: ev.detail,
        eventPhase: ev.eventPhase,
        isComposing: ev.isComposing,
        key: ev.key,
        location: ev.location,
        metaKey: ev.metaKey,
        repeat: ev.repeat,
        shiftKey: ev.shiftKey,
    };
}
function ParseFormData(ev) {
    const form = ev.target;
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
    return data;
}
const api = new API();
new Socket();
