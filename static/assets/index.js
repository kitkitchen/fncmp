// let functions: DispatchFunctions;
let conn_id = undefined;
let base_url = undefined;
let verbose = false;
class Socket {
    constructor() {
        this.ws = null;
        this.addr = undefined;
        this.key = undefined;
        if (this.addr) {
            throw new Error("ws: already connected to server...");
        }
        // check local storage for key
        // if key exists, use it
        // if key does not exist, generate a new key and store in local storage
        let key = localStorage.getItem("fncmp_key");
        if (!key) {
            // generate uuid
            key = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
                var r = (Math.random() * 16) | 0, v = c == "x" ? r : (r & 0x3) | 0x8;
                return v.toString(16);
            });
            localStorage.setItem("fncmp_key", key);
        }
        this.key = key;
        // strip "/"" from end of window.location.pathname if it exists:
        let path = window.location.pathname.split("");
        console.log(path);
        let path_parsed = "";
        if (path[-1] == "/") {
            path.pop();
            path_parsed = path.join("");
        }
        else {
            path_parsed = path.join("");
        }
        if (path_parsed == "") {
            path_parsed = "/main";
        }
        console.log(path_parsed);
        //TODO: get this from local storage or generate a new id and store in local storage
        this.addr = "ws://" + window.location.host + path_parsed + "?fncmp_id=" + this.key;
        console.log(this.addr);
        this.connect();
    }
    connect() {
        // let key = "";
        try {
            this.ws = new WebSocket(this.addr);
        }
        catch (_a) {
            throw new Error("ws: failed to connect to server...");
        }
        this.ws.onopen = function () {
            document.getElementById('fncmp_script').remove();
        };
        this.ws.onclose = function () { };
        this.ws.onerror = function (e) { };
        this.ws.onmessage = function (event) {
            let d = JSON.parse(event.data);
            // if(d.key != key) {
            //     throw new Error("ws: invalid key...");
            // }
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
            initialize: (d) => {
                d = this.utils.parseEventListeners(document.body, d);
                this.Dispatch(this.utils.addEventListeners(d));
            },
            render: (d) => {
                console.log("RENDER:");
                console.log(d);
                let elem = null;
                const parsed = new DOMParser().parseFromString(d.render.html, "text/html").firstChild;
                const html = parsed.getElementsByTagName("body")[0].innerHTML;
                // Select element to render to
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
                // Render element
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
        };
        this.utils = {
            parseEventListeners: (element, d) => {
                const events = this.utils.getAttributes(element, "events");
                const listeners = events.map((e) => {
                    console.log(e);
                    const event = JSON.parse(e);
                    if (!event)
                        return;
                    return event;
                });
                const listeners_flat = listeners.flat();
                const listeners_filtered = listeners_flat.filter((e) => e != null);
                console.log("LISTENERS:");
                console.log(listeners_filtered);
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
            getElementByAttribute: (attribute) => document.querySelectorAll(`[${attribute}]`),
            trackTouch: (elem) => {
                elem.addEventListener("touchstart", (ev) => {
                    //TODO: event object comes back as touch specific
                    ev.preventDefault();
                    elem.classList.add("touch");
                    // send data to api
                });
                elem.addEventListener("touchend", (ev) => {
                    elem.classList.remove("touch");
                });
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
                    console.log("listener: " + listener);
                    let elem = document.getElementById(listener.target_id);
                    if (!elem) {
                        console.log("elem not found");
                        this.Error(d, "element not found");
                        return;
                    }
                    if (elem.firstChild) {
                        console.log("elem has children");
                        elem = elem.firstChild;
                    }
                    else {
                        console.log("elem has no children");
                    }
                    console.log("elem with listener: " + elem);
                    elem.addEventListener(listener.on, (ev) => {
                        ev.preventDefault();
                        console.log("EVENT LISTENER:");
                        console.log(listener.on);
                        console.log("EVENT:");
                        console.log(ev);
                        console.log("TARGET:");
                        console.log(ev.target);
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
                        console.log("DISPATCH:");
                        console.log(d);
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
        console.log("API: initialized...");
    }
    // Process is the entry point for all api calls via the websocket
    Process(ws, d) {
        if (!this.ws) {
            this.ws = ws;
        }
        switch (d.function) {
            case "initialize":
                this.Dispatch(this.funs.initialize(d));
                return;
            case "redirect":
                window.location.href = d.redirect.url;
                return;
            case "custom":
                this.Dispatch(window[d.custom.function](d.custom.data));
                return;
            case "render":
                this.Dispatch(this.funs.render(d));
            default:
                // this.Error(d, "invalid function: " + d.function);
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
// Parse events
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
    console.log(form);
    const formData = new FormData(form);
    console.log(JSON.stringify(formData));
    const data = Object.fromEntries(formData.entries());
    console.log(data);
    return data;
}
new Socket();
const api = new API();
